package handlers

import (
	"net/http"

	"claw-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DailyStat struct {
	Date    string  `json:"date"`
	Expense float64 `json:"expense"`
	Income  float64 `json:"income"`
}

type CategoryStat struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	CategoryIcon string  `json:"category_icon"`
	Amount       float64 `json:"amount"`
	Count        int64   `json:"count"`
}

func MonthlyStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		year := c.Query("year")
		month := c.Query("month")
		if year == "" || month == "" {
			Fail(c, http.StatusBadRequest, "year and month required")
			return
		}

		baseQuery := db.Model(&models.Bill{}).Where("user_id = ? AND deleted_at IS NULL AND YEAR(bill_date) = ? AND MONTH(bill_date) = ?", userID, year, month)

		var totalExpense, totalIncome float64
		baseQuery.Select("COALESCE(SUM(CASE WHEN type=0 THEN amount ELSE 0 END),0)").Scan(&totalExpense)
		baseQuery.Select("COALESCE(SUM(CASE WHEN type=1 THEN amount ELSE 0 END),0)").Scan(&totalIncome)

		var dailyStats []DailyStat
		db.Model(&models.Bill{}).
			Select("bill_date as date, COALESCE(SUM(CASE WHEN type=0 THEN amount ELSE 0 END),0) as expense, COALESCE(SUM(CASE WHEN type=1 THEN amount ELSE 0 END),0) as income").
			Where("user_id = ? AND deleted_at IS NULL AND YEAR(bill_date) = ? AND MONTH(bill_date) = ?", userID, year, month).
			Group("bill_date").
			Order("bill_date ASC").
			Find(&dailyStats)

		var categoryStats []CategoryStat
		db.Model(&models.Bill{}).
			Select("bills.category_id, COALESCE(bc.name,'') as category_name, COALESCE(bc.icon,'') as category_icon, SUM(bills.amount) as amount, COUNT(*) as count").
			Joins("LEFT JOIN bill_categories bc ON bills.category_id = bc.id").
			Where("bills.user_id = ? AND bills.deleted_at IS NULL AND bills.type = 0 AND YEAR(bills.bill_date) = ? AND MONTH(bills.bill_date) = ?", userID, year, month).
			Group("bills.category_id").
			Order("amount DESC").
			Find(&categoryStats)

		Success(c, gin.H{
			"total_expense":  totalExpense,
			"total_income":   totalIncome,
			"balance":        totalIncome - totalExpense,
			"daily_stats":    dailyStats,
			"category_stats": categoryStats,
		}, "ok")
	}
}
