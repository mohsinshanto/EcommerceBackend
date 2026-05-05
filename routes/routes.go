package routes

import (
	"ecommerce-backend/controllers"
	"ecommerce-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	// Base API group
	api := r.Group("/api")

	// ---------------- PUBLIC ----------------
	api.POST("/register", controllers.Register)
	api.POST("/login", controllers.Login)

	// ---------------- USER AUTH ----------------
	auth := api.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		// Products
		auth.GET("/products", controllers.GetProducts)
		auth.GET("/products/count", controllers.GetProductCount)
		auth.GET("/products/:id", controllers.GetProductByID)

		// Cart
		auth.POST("/cart", controllers.AddToCart)
		auth.GET("/cart", controllers.GetCart)
		auth.PUT("/cart/:id", controllers.UpdateCartQuantity) // ✅ FIXED
		auth.DELETE("/cart/:id", controllers.RemoveFromCart)

		// Orders
		auth.POST("/order", controllers.CreateOrder)
		auth.GET("/orders", controllers.GetMyOrders)
		auth.PATCH("/orders/:id/archive", controllers.ArchiveOrder)
	}

	// ---------------- ADMIN ----------------
	admin := api.Group("/")
	admin.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
	{
		admin.POST("/products", controllers.CreateProduct)
		admin.DELETE("/products/:id", controllers.DeleteProduct)
		admin.GET("/admin/orders", controllers.GetAllOrders)
	}
}
