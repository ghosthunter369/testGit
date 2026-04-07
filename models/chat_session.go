package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatSession struct {
	ID            int64          `json:"id" gorm:"primaryKey"`
	UserID        int64          `json:"user_id" gorm:"not null;index:idx_chat_sessions_user_updated,priority:1;index:idx_chat_sessions_user_deleted_last,priority:1"`
	Title         string         `json:"title" gorm:"size:120;not null"`
	LastMessageAt *time.Time     `json:"last_message_at" gorm:"index:idx_chat_sessions_user_deleted_last,priority:3"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"index:idx_chat_sessions_user_updated,sort:desc,priority:2"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index:idx_chat_sessions_user_deleted_last,priority:2"`
}

