package handlers

import (
	"net/http"
	"strconv"

	"claw-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var defaultCategories = []models.BillCategory{
	{UserID: 0, Name: "餐饮美食", Icon: "🍜", Type: 0, SortOrder: 1},
	{UserID: 0, Name: "交通出行", Icon: "🚗", Type: 0, SortOrder: 2},
	{UserID: 0, Name: "购物消费", Icon: "🛒", Type: 0, SortOrder: 3},
	{UserID: 0, Name: "生活缴费", Icon: "💡", Type: 0, SortOrder: 4},
	{UserID: 0, Name: "休闲娱乐", Icon: "🎮", Type: 0, SortOrder: 5},
	{UserID: 0, Name: "医疗健康", Icon: "🏥", Type: 0, SortOrder: 6},
	{UserID: 0, Name: "教育学习", Icon: "📚", Type: 0, SortOrder: 7},
	{UserID: 0, Name: "住房物业", Icon: "🏠", Type: 0, SortOrder: 8},
	{UserID: 0, Name: "人情往来", Icon: "🎁", Type: 0, SortOrder: 9},
	{UserID: 0, Name: "职场办公", Icon: "💼", Type: 0, SortOrder: 10},
	{UserID: 0, Name: "工资收入", Icon: "💰", Type: 1, SortOrder: 1},
	{UserID: 0, Name: "兼职副业", Icon: "💸", Type: 1, SortOrder: 2},
	{UserID: 0, Name: "理财收益", Icon: "📈", Type: 1, SortOrder: 3},
	{UserID: 0, Name: "其他", Icon: "📌", Type: 0, SortOrder: 99},
}

// InitCategories 初始化系统预置分类，首次启动时调用
func InitCategories(db *gorm.DB) {
	var count int64
	db.Model(&models.BillCategory{}).Where("user_id = 0").Count(&count)
	if count > 0 {
		return
	}
	for i := range defaultCategories {
		db.FirstOrCreate(&defaultCategories[i], models.BillCategory{
			UserID: defaultCategories[i].UserID,
			Name:   defaultCategories[i].Name,
		})
	}
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
	Icon string `json:"icon" binding:"required"`
	Type int8   `json:"type"`
}

func ListCategories(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		typeParam := c.Query("type")
		query := db.Where("user_id = ? OR user_id = 0", userID)
		if typeParam != "" {
			t, err := strconv.Atoi(typeParam)
			if err == nil {
				query = query.Where("type = ?", t)
			}
		}

		var categories []models.BillCategory
		query.Order("sort_order ASC, id ASC").Find(&categories)

		Success(c, categories, "ok")
	}
}

func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req CreateCategoryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}

		cat := models.BillCategory{
			UserID:    userID,
			Name:      req.Name,
			Icon:      req.Icon,
			Type:      req.Type,
			SortOrder: 100,
		}
		if err := db.Create(&cat).Error; err != nil {
			Fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		Success(c, cat, "ok")
	}
}

func DeleteCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			Fail(c, http.StatusBadRequest, "invalid id")
			return
		}

		var cat models.BillCategory
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&cat).Error; err != nil {
			Fail(c, http.StatusNotFound, "分类不存在或不可删除")
			return
		}

		if cat.UserID == 0 {
			Fail(c, http.StatusForbidden, "系统预置分类不可删除")
			return
		}

		db.Delete(&cat)
		Success(c, nil, "ok")
	}
}
