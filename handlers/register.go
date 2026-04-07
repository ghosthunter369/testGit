package handlers

import (
	"claw-backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

func Register(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			Fail(c, 400, err.Error())
			return
		}

		var user models.User
		if err := db.Where("username = ?", input.Username).First(&user).Error; err == nil {
			Fail(c, 400, "username already exists")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			Fail(c, 500, "failed to hash password")
			return
		}

		newUser := models.User{
			Username: input.Username,
			Password: string(hash),
			Email:    input.Email,
		}

		if err := db.Create(&newUser).Error; err != nil {
			Fail(c, 500, "failed to create user")
			return
		}

		Success(c, gin.H{"user_id": newUser.ID}, "register success")
	}
}

