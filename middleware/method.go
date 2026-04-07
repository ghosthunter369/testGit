package middleware

import (
	"net/http"

	"claw-backend/handlers"

	"github.com/gin-gonic/gin"
)

func MethodOnlyGetPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodPost {
			handlers.Fail(c, http.StatusMethodNotAllowed, "only GET and POST are allowed")
			c.Abort()
			return
		}
		c.Next()
	}
}

