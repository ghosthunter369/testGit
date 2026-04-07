package middleware

import (
	"context"
	"net/http"
	"strings"

	"claw-backend/handlers"
	"claw-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const authCookieName = "auth_token"

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

func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := tokenFromRequest(c)
		if token == "" {
			handlers.Fail(c, http.StatusUnauthorized, "missing authorization")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			handlers.Fail(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		ctx := context.Background()
		exists, err := rdb.Exists(ctx, redisTokenKey(token)).Result()
		if err != nil {
			handlers.Fail(c, http.StatusInternalServerError, "failed to validate session")
			c.Abort()
			return
		}
		if exists == 0 {
			handlers.Fail(c, http.StatusUnauthorized, "session expired, please login again")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token", token)
		c.Next()
	}
}

