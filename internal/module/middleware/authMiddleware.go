package middleware

import (
	"ecommerce-backend/config"
	"ecommerce-backend/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			utils.RespondError(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			utils.RespondError(c, http.StatusUnauthorized, "Invalid authorization header")
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return config.JwtSecret, nil
		})

		if err != nil || token == nil || !token.Valid {
			utils.RespondError(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondError(c, http.StatusUnauthorized, "Invalid token claims")
			c.Abort()
			return
		}

		userIDValue, ok := claims["user_id"].(float64)
		if !ok || userIDValue <= 0 || userIDValue != float64(uint(userIDValue)) {
			utils.RespondError(c, http.StatusUnauthorized, "Invalid user ID in token")
			c.Abort()
			return
		}

		c.Set("user_id", uint(userIDValue))
		c.Next()
	}
}
