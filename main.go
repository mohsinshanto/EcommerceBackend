package main

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"ecommerce-backend/routes"
	"ecommerce-backend/utils"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	config.LoadJWTSecret()
	config.ConnectDB()
	config.DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Cart{},
		&models.Order{},
		&models.OrderItem{},
	)
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
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
