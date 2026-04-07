package models

import (
	"time"

	"gorm.io/gorm"
)

type ChatMessage struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	SessionID int64          `json:"session_id" gorm:"not null;index:idx_chat_messages_session_created,priority:1;index:idx_chat_messages_user_session,priority:2;uniqueIndex:uk_chat_messages_session_seq,priority:1"`
	UserID    int64          `json:"user_id" gorm:"not null;index:idx_chat_messages_user_session,priority:1"`
	Role      string         `json:"role" gorm:"type:varchar(20);not null"`
	Content   string         `json:"content" gorm:"type:longtext;not null"`
	Seq       int            `json:"seq" gorm:"not null;uniqueIndex:uk_chat_messages_session_seq,priority:2"`
	Status    int8           `json:"status" gorm:"not null;default:1"`
	CreatedAt time.Time      `json:"created_at" gorm:"index:idx_chat_messages_session_created,priority:2"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

