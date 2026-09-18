package utils

import (
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   any    `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func RespondSuccess(c *gin.Context, status int, message string, data any) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Error:   message,
	})
}
