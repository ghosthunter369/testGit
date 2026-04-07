package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"claw-backend/models"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type ChatRequest struct {
	SessionID int64  `json:"session_id"`
	Message   string `json:"message" binding:"required"`
}

type ChatStreamData struct {
	Type      string `json:"type"`
	SessionID int64  `json:"session_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

func ChatStream(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		req.Message = strings.TrimSpace(req.Message)
		if req.Message == "" {
			Fail(c, http.StatusBadRequest, "message is required")
			return
		}

		session, err := loadOrCreateSession(db, userID, req.SessionID, req.Message)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				Fail(c, http.StatusNotFound, "chat session not found")
				return
			}
			Fail(c, http.StatusInternalServerError, "failed to load chat session: "+err.Error())
			return
		}

		now := time.Now()
		userSeq, err := nextMessageSeq(db, session.ID)
		if err != nil {
			Fail(c, http.StatusInternalServerError, "failed to allocate message sequence: "+err.Error())
			return
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			msg := models.ChatMessage{
				SessionID: session.ID,
				UserID:    userID,
				Role:      "user",
				Content:   req.Message,
				Seq:       userSeq,
				Status:    1,
			}
			if err := tx.Create(&msg).Error; err != nil {
				return err
			}
			return tx.Model(&models.ChatSession{}).
				Where("id = ? AND user_id = ?", session.ID, userID).
				Updates(map[string]interface{}{
					"last_message_at": now,
					"updated_at":      now,
				}).Error
		}); err != nil {
			Fail(c, http.StatusInternalServerError, "failed to persist user message: "+err.Error())
			return
		}

		ctx := context.Background()
		timeout := 30 * time.Second
		apiKey := strings.TrimSpace(os.Getenv("ARKAPIKEY"))
		if apiKey == "" {
			apiKey = strings.TrimSpace(os.Getenv("ARK_API_KEY"))
		}
		if apiKey == "" {
			Fail(c, http.StatusInternalServerError, "missing model api key: set ARKAPIKEY or ARK_API_KEY")
			return
		}

		modelName := strings.TrimSpace(os.Getenv("MODEL"))
		if modelName == "" {
			Fail(c, http.StatusInternalServerError, "missing model name: set MODEL")
			return
		}

		model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:  apiKey,
			Model:   modelName,
			Timeout: &timeout,
		})
		if err != nil {
			Fail(c, http.StatusInternalServerError, "failed to initialize model: "+err.Error())
			return
		}

		var history []models.ChatMessage
		if err := db.Where("user_id = ? AND session_id = ?", userID, session.ID).
			Order("seq DESC").
			Limit(contextMaxMessages()).
			Find(&history).Error; err != nil {
			Fail(c, http.StatusInternalServerError, "failed to query chat history: "+err.Error())
			return
		}

		messages := make([]*schema.Message, 0, len(history)+1)
		messages = append(messages, schema.SystemMessage("You are a helpful AI assistant."))
		for i := len(history) - 1; i >= 0; i-- {
			h := history[i]
			if h.Content == "" {
				continue
			}
			switch h.Role {
			case "assistant":
				messages = append(messages, schema.AssistantMessage(h.Content, nil))
			case "user":
				messages = append(messages, schema.UserMessage(h.Content))
			}
		}

		reader, err := model.Stream(ctx, messages)
		if err != nil {
			Fail(c, http.StatusInternalServerError, "failed to call model stream: "+err.Error())
			return
		}
		defer reader.Close()

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			Fail(c, http.StatusInternalServerError, "streaming unsupported by server")
			return
		}

		writeEvent := func(code int, data interface{}, message string) bool {
			payload := APIResponse{
				Code:    code,
				Data:    data,
				Message: message,
			}
			b, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				return false
			}
			if _, writeErr := fmt.Fprintf(c.Writer, "data: %s\n\n", b); writeErr != nil {
				return false
			}
			flusher.Flush()
			return true
		}

		if !writeEvent(0, ChatStreamData{Type: "start", SessionID: session.ID}, "stream started") {
			return
		}

		var assistantBuilder strings.Builder
		for {
			chunk, recvErr := reader.Recv()
			if recvErr != nil {
				if errors.Is(recvErr, io.EOF) {
					break
				}
				writeEvent(http.StatusInternalServerError, ChatStreamData{Type: "error", SessionID: session.ID}, recvErr.Error())
				return
			}
			if chunk.Content == "" {
				continue
			}
			assistantBuilder.WriteString(chunk.Content)
			if !writeEvent(0, ChatStreamData{Type: "chunk", SessionID: session.ID, Content: chunk.Content}, "ok") {
				return
			}
		}

		assistantText := strings.TrimSpace(assistantBuilder.String())
		if assistantText != "" {
			assistantSeq, seqErr := nextMessageSeq(db, session.ID)
			if seqErr != nil {
				writeEvent(http.StatusInternalServerError, ChatStreamData{Type: "error", SessionID: session.ID}, "failed to save assistant message: "+seqErr.Error())
				return
			}

			now = time.Now()
			if err := db.Transaction(func(tx *gorm.DB) error {
				msg := models.ChatMessage{
					SessionID: session.ID,
					UserID:    userID,
					Role:      "assistant",
					Content:   assistantText,
					Seq:       assistantSeq,
					Status:    1,
				}
				if err := tx.Create(&msg).Error; err != nil {
					return err
				}
				return tx.Model(&models.ChatSession{}).
					Where("id = ? AND user_id = ?", session.ID, userID).
					Updates(map[string]interface{}{
						"last_message_at": now,
						"updated_at":      now,
					}).Error
			}); err != nil {
				writeEvent(http.StatusInternalServerError, ChatStreamData{Type: "error", SessionID: session.ID}, "failed to save assistant message: "+err.Error())
				return
			}
		}

		writeEvent(0, ChatStreamData{Type: "done", SessionID: session.ID}, "stream done")
	}
}

