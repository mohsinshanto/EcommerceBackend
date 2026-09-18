package middleware

import (
	"ecommerce-backend/config"
	"ecommerce-backend/internal/module/user"
	"ecommerce-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		var account user.User
		if err := config.DB.WithContext(c.Request.Context()).First(&account, userID).Error; err != nil {
			utils.RespondError(c, http.StatusUnauthorized, "User not found")
			c.Abort()
			return
		}

		if !account.IsAdmin {
			utils.RespondError(c, http.StatusForbidden, "Admin access only")
			c.Abort()
			return
		}

		c.Next()
	}
}
