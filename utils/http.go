package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func RespondSuccess(c *gin.Context, status int, message string, data interface{}) {
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

func RespondBindError(c *gin.Context, err error) {
	RespondError(c, http.StatusBadRequest, HumanizeValidationError(err))
}

func HumanizeValidationError(err error) string {
	if err == nil {
		return "Invalid request data"
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		messages := make([]string, 0, len(validationErrs))
		for _, validationErr := range validationErrs {
			field := strings.ToLower(validationErr.Field())
			switch validationErr.Tag() {
			case "required":
				messages = append(messages, fmt.Sprintf("%s is required", field))
			case "email":
				messages = append(messages, fmt.Sprintf("%s must be a valid email address", field))
			case "min":
				messages = append(messages, fmt.Sprintf("%s is too short", field))
			case "gte":
				messages = append(messages, fmt.Sprintf("%s must be at least %s", field, validationErr.Param()))
			default:
				messages = append(messages, fmt.Sprintf("%s is invalid", field))
			}
		}

		if len(messages) > 0 {
			return strings.Join(messages, ", ")
		}
	}

	return "Invalid request data"
}
