package handlers

import (
	"os"
	"strconv"
	"strings"
	"time"

	"claw-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getUserID(c *gin.Context) (int64, bool) {
	raw, ok := c.Get("user_id")
	if !ok {
		Fail(c, 401, "missing user context")
		return 0, false
	}
	userID, ok := raw.(int64)
	if !ok || userID <= 0 {
		Fail(c, 401, "invalid user context")
		return 0, false
	}
	return userID, true
}

func defaultSessionTitle(message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return "新对话"
	}
	runes := []rune(trimmed)
	if len(runes) > 24 {
		return string(runes[:24])
	}
	return trimmed
}

func loadOrCreateSession(db *gorm.DB, userID int64, sessionID int64, titleHint string) (models.ChatSession, error) {
	var session models.ChatSession
	if sessionID > 0 {
		err := db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
		return session, err
	}

	title := defaultSessionTitle(titleHint)
	now := time.Now()
	session = models.ChatSession{
		UserID:        userID,
		Title:         title,
		LastMessageAt: &now,
	}
	if err := db.Create(&session).Error; err != nil {
		return session, err
	}
	return session, nil
}

func nextMessageSeq(db *gorm.DB, sessionID int64) (int, error) {
	var maxSeq int
	err := db.Model(&models.ChatMessage{}).
		Select("COALESCE(MAX(seq), 0)").
		Where("session_id = ?", sessionID).
		Scan(&maxSeq).Error
	if err != nil {
		return 0, err
	}
	return maxSeq + 1, nil
}

func contextMaxMessages() int {
	raw := strings.TrimSpace(os.Getenv("CHAT_CONTEXT_MAX_MESSAGES"))
	if raw == "" {
		return 20
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 20
	}
	if n > 100 {
		return 100
	}
	return n
}

