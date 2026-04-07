package handlers

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(200, APIResponse{
		Code:    0,
		Data:    data,
		Message: message,
	})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Code:    status,
		Data:    nil,
		Message: message,
	})
}
