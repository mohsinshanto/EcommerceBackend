package main

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"ecommerce-backend/routes"
	"ecommerce-backend/utils"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	config.LoadJWTSecret()
	config.InitRedis()
	config.ConnectDB()
	config.DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Cart{},
		&models.Order{},
		&models.OrderItem{},
	)

	frontendOrigins := []string{"http://localhost:3000"}
	if configuredOrigins := os.Getenv("FRONTEND_ORIGINS"); configuredOrigins != "" {
		frontendOrigins = splitAndTrim(configuredOrigins)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     frontendOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.NoRoute(func(c *gin.Context) {
		utils.RespondError(c, http.StatusNotFound, "Route not found")
	})

	r.NoMethod(func(c *gin.Context) {
		utils.RespondError(c, http.StatusMethodNotAllowed, "Method not allowed")
	})

	routes.SetupRoutes(r)

	r.Run(":8080")
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}

	return origins
}
