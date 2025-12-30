package routes

import (
	"ecommerce-backend/controllers"
	"ecommerce-backend/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{

		// Public
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)

		// User authenticated
		api.Use(middlewares.AuthMiddleware())
		{
			// Products (user can view)
			api.GET("/products", controllers.GetProducts)

			// Cart
			api.POST("/cart", controllers.AddToCart)
			api.GET("/cart", controllers.GetCart)
			api.DELETE("/cart/:id", controllers.RemoveFromCart)

			// Orders
			api.POST("/order", controllers.CreateOrder)
			api.GET("/orders", controllers.GetMyOrders)
		}

		// Admin-only
		api.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
		{
			api.POST("/products", controllers.CreateProduct)
			api.GET("/admin/orders", controllers.GetAllOrders)
		}
	}
}
