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

	rdb := config.InitRedis()

	r := gin.Default()
	r.Use(middleware.MethodOnlyGetPost())

	r.POST("/api/register", handlers.Register(db))
	r.POST("/api/login", handlers.Login(db, rdb))

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(rdb))
	{
		protected.GET("/user", handlers.GetUserInfo(db))
		protected.POST("/logout", handlers.Logout(rdb))
		protected.GET("/chat/sessions", handlers.ListChatSessions(db))
		protected.POST("/chat/session/create", handlers.CreateChatSession(db))
		protected.POST("/chat/session/rename", handlers.RenameChatSession(db))
		protected.POST("/chat/session/delete", handlers.DeleteChatSession(db))
		protected.GET("/chat/messages", handlers.ListChatMessages(db))
		protected.POST("/chat/stream", handlers.ChatStream(db))
	}

	r.Run(":8080")
}
