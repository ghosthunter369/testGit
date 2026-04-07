package handlers

import (
	"strconv"

	"claw-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListChatMessages(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		sessionID, err := strconv.ParseInt(c.Query("session_id"), 10, 64)
		if err != nil || sessionID <= 0 {
			Fail(c, 400, "invalid session_id")
			return
		}

		var session models.ChatSession
		if err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
			Fail(c, 404, "chat session not found")
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		if page <= 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
		if pageSize <= 0 {
			pageSize = 100
		}
		if pageSize > 200 {
			pageSize = 200
		}

		var messages []models.ChatMessage
		query := db.Where("user_id = ? AND session_id = ?", userID, sessionID).
			Order("seq ASC").
			Offset((page - 1) * pageSize).
			Limit(pageSize)
		if err := query.Find(&messages).Error; err != nil {
			Fail(c, 500, "failed to list chat messages: "+err.Error())
			return
		}

		Success(c, gin.H{
			"messages":   messages,
			"session_id": sessionID,
			"page":       page,
			"page_size":  pageSize,
		}, "ok")
	}
}

