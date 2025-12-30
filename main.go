package main

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"ecommerce-backend/routes"

	"github.com/gin-gonic/gin"
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
	routes.SetupRoutes(r)

	r.Run(":8080")
}
