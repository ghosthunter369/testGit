package handlers

import (
	"strings"
	"time"

	"claw-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateSessionInput struct {
	Title string `json:"title"`
}

type RenameSessionInput struct {
	SessionID int64  `json:"session_id" binding:"required"`
	Title     string `json:"title" binding:"required"`
}

type DeleteSessionInput struct {
	SessionID int64 `json:"session_id" binding:"required"`
}

func ListChatSessions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var sessions []models.ChatSession
		if err := db.Where("user_id = ?", userID).
			Order("last_message_at DESC").
			Order("updated_at DESC").
			Find(&sessions).Error; err != nil {
			Fail(c, 500, "failed to list chat sessions: "+err.Error())
			return
		}

		Success(c, gin.H{"sessions": sessions}, "ok")
	}
}

func CreateChatSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req CreateSessionInput
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, 400, err.Error())
			return
		}

		title := strings.TrimSpace(req.Title)
		if title == "" {
			title = "新对话"
		}

		now := time.Now()
		session := models.ChatSession{
			UserID:        userID,
			Title:         title,
			LastMessageAt: &now,
		}
		if err := db.Create(&session).Error; err != nil {
			Fail(c, 500, "failed to create chat session: "+err.Error())
			return
		}

		Success(c, gin.H{"session": session}, "created")
	}
}

func RenameChatSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req RenameSessionInput
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, 400, err.Error())
			return
		}

		title := strings.TrimSpace(req.Title)
		if title == "" {
			Fail(c, 400, "title is required")
			return
		}

		var session models.ChatSession
		if err := db.Where("id = ? AND user_id = ?", req.SessionID, userID).First(&session).Error; err != nil {
			Fail(c, 404, "chat session not found")
			return
		}

		session.Title = title
		if err := db.Save(&session).Error; err != nil {
			Fail(c, 500, "failed to rename chat session: "+err.Error())
			return
		}

		Success(c, gin.H{"session": session}, "renamed")
	}
}

func DeleteChatSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req DeleteSessionInput
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, 400, err.Error())
			return
		}

		var session models.ChatSession
		if err := db.Where("id = ? AND user_id = ?", req.SessionID, userID).First(&session).Error; err != nil {
			Fail(c, 404, "chat session not found")
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("session_id = ? AND user_id = ?", req.SessionID, userID).Delete(&models.ChatMessage{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&session).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			Fail(c, 500, "failed to delete chat session: "+err.Error())
			return
		}

		Success(c, nil, "deleted")
	}
}

