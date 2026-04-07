package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;size:50;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Email     string    `json:"email" gorm:"size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{}, &ChatSession{}, &ChatMessage{})
}
