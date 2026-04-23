package models

import (
	"time"

	"gorm.io/gorm"
)

type BillCategory struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	UserID    int64     `json:"user_id" gorm:"not null;index:idx_bill_categories_user"`
	Name      string    `json:"name" gorm:"size:30;not null"`
	Icon      string    `json:"icon" gorm:"size:30;not null;default:''"`
	Type      int8      `json:"type" gorm:"not null;default:0;comment:0=支出 1=收入"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Bill struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	UserID      int64          `json:"user_id" gorm:"not null;index:idx_bills_user_date,priority:1;index:idx_bills_user_category,priority:1"`
	CategoryID  int64          `json:"category_id" gorm:"not null;index:idx_bills_user_category,priority:2"`
	Amount      float64        `json:"amount" gorm:"type:decimal(12,2);not null"`
	Type        int8           `json:"type" gorm:"not null;default:0;comment:0=支出 1=收入"`
	Merchant    string         `json:"merchant" gorm:"size:100;not null;default:''"`
	Description string         `json:"description" gorm:"size:255;not null;default:''"`
	BillDate    string         `json:"bill_date" gorm:"type:date;not null"`
	ImageURL    string         `json:"image_url" gorm:"size:500;not null;default:''"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index:idx_bills_deleted"`
}
