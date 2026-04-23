package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"claw-backend/models"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type RecognizeResult struct {
	Amount         float64 `json:"amount"`
	Merchant       string  `json:"merchant"`
	Category       string  `json:"category"`
	Date           string  `json:"date"`
	Description    string  `json:"description"`
	RecognizeError string  `json:"error,omitempty"`
}

func UploadBillImage(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := getUserID(c)
		if !ok {
			return
		}

		file, header, err := c.Request.FormFile("image")
		if err != nil {
			Fail(c, http.StatusBadRequest, "请上传图片")
			return
		}
		defer file.Close()

		ext := filepath.Ext(header.Filename)
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			Fail(c, http.StatusBadRequest, "仅支持 jpg/png/webp 格式")
			return
		}

		userID, _ := c.Get("user_id")
		uid := userID.(int64)

		dir := filepath.Join("uploads", "bills", fmt.Sprintf("%d", uid))
		os.MkdirAll(dir, 0755)

		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		dst := filepath.Join(dir, filename)

		if err := c.SaveUploadedFile(header, dst); err != nil {
			Fail(c, http.StatusInternalServerError, "文件保存失败")
			return
		}

		Success(c, gin.H{"image_url": dst}, "ok")
	}
}

func RecognizeBill(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := getUserID(c)
		if !ok {
			return
		}

		var req struct {
			ImageURL string `json:"image_url" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}

		imageData, err := os.ReadFile(req.ImageURL)
		if err != nil {
			Fail(c, http.StatusBadRequest, "图片文件不存在")
			return
		}

		b64 := base64.StdEncoding.EncodeToString(imageData)
		ext := filepath.Ext(req.ImageURL)
		mimeType := "image/jpeg"
		if ext == ".png" {
			mimeType = "image/png"
		} else if ext == ".webp" {
			mimeType = "image/webp"
		}
		dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, b64)

		ctx := context.Background()
		timeout := 30 * time.Second
		apiKey := strings.TrimSpace(os.Getenv("ARKAPIKEY"))
		if apiKey == "" {
			apiKey = strings.TrimSpace(os.Getenv("ARK_API_KEY"))
		}
		if apiKey == "" {
			Fail(c, http.StatusInternalServerError, "missing ARKAPIKEY")
			return
		}

		visionModel := strings.TrimSpace(os.Getenv("VISION_MODEL"))
		if visionModel == "" {
			Fail(c, http.StatusInternalServerError, "missing VISION_MODEL")
			return
		}

		model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:  apiKey,
			Model:   visionModel,
			Timeout: &timeout,
		})
		if err != nil {
			Fail(c, http.StatusInternalServerError, "初始化模型失败: "+err.Error())
			return
		}

		systemPrompt := `你是一个专业的账单识别助手。用户会上传小票、账单、支付截图等图片。
请从图片中提取以下信息，严格以 JSON 格式返回，不要添加任何其他文字：
{
  "amount": 金额(数字),
  "merchant": "商户名称",
  "category": "分类(只能从以下选择: 餐饮美食, 交通出行, 购物消费, 生活缴费, 休闲娱乐, 医疗健康, 教育学习, 住房物业, 人情往来, 职场办公, 工资收入, 兼职副业, 理财收益, 其他)",
  "date": "日期(YYYY-MM-DD，如果图片中没有明确日期则返回今天的日期)",
  "description": "简要描述这笔消费"
}
如果图片模糊或无法识别，返回：
{"error": "无法识别该图片，请重新上传或手动录入"}`

		today := time.Now().Format("2006-01-02")
		userMsg := fmt.Sprintf("请识别这张图片中的账单信息，今天的日期是 %s。只返回JSON，不要任何其他文字。", today)

		messages := []*schema.Message{
			schema.SystemMessage(systemPrompt),
			schema.UserMessage(
				schema.NewUserMessage(
					[]schema.ChatMessagePart{
						{Type: schema.ChatMessagePartTypeText, Text: userMsg},
						{Type: schema.ChatMessagePartTypeImage, ImageURL: &schema.ImageURL{URL: dataURL}},
					},
				),
			),
		}

		resp, err := model.Generate(ctx, messages)
		if err != nil {
			Fail(c, http.StatusInternalServerError, "AI识别失败: "+err.Error())
			return
		}

		resultText := strings.TrimSpace(resp.Content)
		resultText = strings.TrimPrefix(resultText, "```json")
		resultText = strings.TrimSuffix(resultText, "```")
		resultText = strings.TrimSpace(resultText)

		var result RecognizeResult
		if err := json.Unmarshal([]byte(resultText), &result); err != nil {
			Fail(c, http.StatusInternalServerError, "AI返回格式异常，请重试或手动录入")
			return
		}

		if result.RecognizeError != "" {
			Fail(c, http.StatusBadRequest, result.RecognizeError)
			return
		}

		Success(c, result, "ok")
	}
}

type CreateBillRequest struct {
	CategoryID  int64   `json:"category_id"`
	Amount      float64 `json:"amount" binding:"required"`
	Type        int8    `json:"type"`
	Merchant    string  `json:"merchant"`
	Description string  `json:"description"`
	BillDate    string  `json:"bill_date" binding:"required"`
	ImageURL    string  `json:"image_url"`
}

func CreateBill(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		var req CreateBillRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}

		bill := models.Bill{
			UserID:      userID,
			CategoryID:  req.CategoryID,
			Amount:      req.Amount,
			Type:        req.Type,
			Merchant:    req.Merchant,
			Description: req.Description,
			BillDate:    req.BillDate,
			ImageURL:    req.ImageURL,
		}
		if err := db.Create(&bill).Error; err != nil {
			Fail(c, http.StatusInternalServerError, err.Error())
			return
		}
		Success(c, bill, "ok")
	}
}

func ListBills(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		year := c.Query("year")
		month := c.Query("month")
		categoryID := c.Query("category_id")
		billType := c.Query("type")

		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		query := db.Where("user_id = ? AND deleted_at IS NULL", userID)
		if year != "" {
			query = query.Where("YEAR(bill_date) = ?", year)
		}
		if month != "" {
			query = query.Where("MONTH(bill_date) = ?", month)
		}
		if categoryID != "" && categoryID != "0" {
			query = query.Where("category_id = ?", categoryID)
		}
		if billType != "" && billType != "-1" {
			query = query.Where("type = ?", billType)
		}

		var total int64
		query.Model(&models.Bill{}).Count(&total)

		var bills []models.Bill
		query.Order("bill_date DESC, id DESC").
			Offset((page - 1) * size).Limit(size).
			Find(&bills)

		type BillVO struct {
			ID           int64   `json:"id"`
			CategoryID   int64   `json:"category_id"`
			CategoryName string  `json:"category_name"`
			CategoryIcon string  `json:"category_icon"`
			Amount       float64 `json:"amount"`
			Type         int8    `json:"type"`
			Merchant     string  `json:"merchant"`
			Description  string  `json:"description"`
			BillDate     string  `json:"bill_date"`
			ImageURL     string  `json:"image_url"`
			CreatedAt    string  `json:"created_at"`
		}

		voList := make([]BillVO, 0, len(bills))
		catMap := make(map[int64]models.BillCategory)

		for _, b := range bills {
			vo := BillVO{
				ID:          b.ID,
				CategoryID:  b.CategoryID,
				Amount:      b.Amount,
				Type:        b.Type,
				Merchant:    b.Merchant,
				Description: b.Description,
				BillDate:    b.BillDate,
				ImageURL:    b.ImageURL,
				CreatedAt:   b.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			if cat, ok := catMap[b.CategoryID]; ok {
				vo.CategoryName = cat.Name
				vo.CategoryIcon = cat.Icon
			} else {
				var cat models.BillCategory
				db.First(&cat, b.CategoryID)
				catMap[b.CategoryID] = cat
				vo.CategoryName = cat.Name
				vo.CategoryIcon = cat.Icon
			}
			voList = append(voList, vo)
		}

		var summary struct {
			Expense float64 `json:"expense"`
			Income  float64 `json:"income"`
		}
		summaryQuery := db.Model(&models.Bill{}).Where("user_id = ? AND deleted_at IS NULL", userID)
		if year != "" {
			summaryQuery = summaryQuery.Where("YEAR(bill_date) = ?", year)
		}
		if month != "" {
			summaryQuery = summaryQuery.Where("MONTH(bill_date) = ?", month)
		}
		summaryQuery.Select("COALESCE(SUM(CASE WHEN type=0 THEN amount ELSE 0 END),0) as expense, COALESCE(SUM(CASE WHEN type=1 THEN amount ELSE 0 END),0) as income").Scan(&summary)

		Success(c, gin.H{
			"list":    voList,
			"total":   total,
			"page":    page,
			"size":    size,
			"summary": summary,
		}, "ok")
	}
}

func GetBillDetail(db *gorm.DB) gin.HandlerFunc {
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

		var bill models.Bill
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&bill).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				Fail(c, http.StatusNotFound, "账单不存在")
				return
			}
			Fail(c, http.StatusInternalServerError, err.Error())
			return
		}

		var cat models.BillCategory
		db.First(&cat, bill.CategoryID)

		Success(c, gin.H{
			"id":            bill.ID,
			"category_id":   bill.CategoryID,
			"category_name": cat.Name,
			"category_icon": cat.Icon,
			"amount":        bill.Amount,
			"type":          bill.Type,
			"merchant":      bill.Merchant,
			"description":   bill.Description,
			"bill_date":     bill.BillDate,
			"image_url":     bill.ImageURL,
		}, "ok")
	}
}

func UpdateBill(db *gorm.DB) gin.HandlerFunc {
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

		var bill models.Bill
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&bill).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				Fail(c, http.StatusNotFound, "账单不存在")
				return
			}
			Fail(c, http.StatusInternalServerError, err.Error())
			return
		}

		var req CreateBillRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Fail(c, http.StatusBadRequest, err.Error())
			return
		}

		updates := map[string]interface{}{
			"amount":      req.Amount,
			"type":        req.Type,
			"merchant":    req.Merchant,
			"description": req.Description,
			"bill_date":   req.BillDate,
			"category_id": req.CategoryID,
			"image_url":   req.ImageURL,
		}
		db.Model(&bill).Updates(updates)
		Success(c, nil, "ok")
	}
}

func DeleteBill(db *gorm.DB) gin.HandlerFunc {
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

		result := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Bill{})
		if result.RowsAffected == 0 {
			Fail(c, http.StatusNotFound, "账单不存在")
			return
		}
		Success(c, nil, "ok")
	}
}
