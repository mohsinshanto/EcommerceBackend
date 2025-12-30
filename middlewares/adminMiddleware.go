package middlewares

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")

		var user models.User
		config.DB.First(&user, userID)

		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access only"})
			c.Abort()
			return
		}

		c.Next()
	}
}
