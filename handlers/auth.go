package handlers

import (
	"context"
	"strings"
	"time"

	"claw-backend/models"
	"claw-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const authCookieName = "auth_token"

type LoginInput struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

func redisTokenKey(token string) string {
	return "token:session:" + token
}

func tokenFromRequest(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" && strings.TrimSpace(parts[1]) != "" {
			return strings.TrimSpace(parts[1])
		}
	}
	cookieToken, _ := c.Cookie(authCookieName)
	return strings.TrimSpace(cookieToken)
}

func Login(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			Fail(c, 400, err.Error())
			return
		}

		var user models.User
		if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
			Fail(c, 401, "invalid username or password")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			Fail(c, 401, "invalid username or password")
			return
		}

		ttl := 24 * time.Hour
		cookieMaxAge := 0
		if input.RememberMe {
			ttl = 30 * 24 * time.Hour
			cookieMaxAge = int(ttl.Seconds())
		}

		token, err := utils.GenerateTokenWithTTL(user.ID, user.Username, ttl)
		if err != nil {
			Fail(c, 500, "failed to generate token")
			return
		}

		ctx := context.Background()
		if err := rdb.Set(ctx, redisTokenKey(token), user.ID, ttl).Err(); err != nil {
			Fail(c, 500, "failed to persist login session")
			return
		}

		c.SetCookie(authCookieName, token, cookieMaxAge, "/", "", false, true)

		Success(c, gin.H{
			"token": token,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
			"remember_me": input.RememberMe,
		}, "login success")
	}
}

func GetUserInfo(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			Fail(c, 404, "user not found")
			return
		}

		Success(c, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		}, "ok")
	}
}

func Logout(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromRequest(c)
		if token != "" {
			_ = rdb.Del(context.Background(), redisTokenKey(token)).Err()
		}
		c.SetCookie(authCookieName, "", -1, "/", "", false, true)
		Success(c, nil, "logout success")
	}
}

