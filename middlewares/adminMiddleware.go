package middlewares

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"ecommerce-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		var user models.User
		if err := config.DB.First(&user, userID).Error; err != nil {
			utils.RespondError(c, http.StatusUnauthorized, "User not found")
			c.Abort()
			return
		}

		if !user.IsAdmin {
			utils.RespondError(c, http.StatusForbidden, "Admin access only")
			c.Abort()
			return
		}

		c.Next()
	}
}
