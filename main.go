package main

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"ecommerce-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {

	config.ConnectDB()
	config.DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Cart{},
		&models.Order{},
		&models.OrderItem{},
	)
	r := gin.Default()
	r.Static("/products", "./public/products")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	routes.SetupRoutes(r)

	r.Run(":8080")
}
