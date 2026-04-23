package main

import (
	"claw-backend/config"
	"claw-backend/handlers"
	"claw-backend/middleware"
	"claw-backend/models"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env not found, fallback to system environment variables")
	}

	db := config.InitDB()
	models.AutoMigrate(db)
	handlers.InitCategories(db)

	rdb := config.InitRedis()

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20
	r.Use(middleware.MethodOnlyGetPost())
	r.Static("/uploads", "./uploads")

	r.POST("/api/register", handlers.Register(db))
	r.POST("/api/login", handlers.Login(db, rdb))

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(rdb))
	{
		protected.GET("/user", handlers.GetUserInfo(db))
		protected.POST("/logout", handlers.Logout(rdb))

		// AI 对话
		protected.GET("/chat/sessions", handlers.ListChatSessions(db))
		protected.POST("/chat/session/create", handlers.CreateChatSession(db))
		protected.POST("/chat/session/rename", handlers.RenameChatSession(db))
		protected.POST("/chat/session/delete", handlers.DeleteChatSession(db))
		protected.GET("/chat/messages", handlers.ListChatMessages(db))
		protected.POST("/chat/stream", handlers.ChatStream(db))

		// 记账
		protected.POST("/bill/upload", handlers.UploadBillImage(db))
		protected.POST("/bill/recognize", handlers.RecognizeBill(db))
		protected.POST("/bill/create", handlers.CreateBill(db))
		protected.GET("/bill/list", handlers.ListBills(db))
		protected.GET("/bill/detail/:id", handlers.GetBillDetail(db))
		protected.PUT("/bill/update/:id", handlers.UpdateBill(db))
		protected.DELETE("/bill/delete/:id", handlers.DeleteBill(db))
		protected.GET("/bill/stats/monthly", handlers.MonthlyStats(db))
		protected.GET("/bill/categories", handlers.ListCategories(db))
		protected.POST("/bill/category/create", handlers.CreateCategory(db))
		protected.DELETE("/bill/category/delete/:id", handlers.DeleteCategory(db))
	}

	r.Run(":8080")
}
