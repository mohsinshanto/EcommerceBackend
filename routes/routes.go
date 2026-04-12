package routes

import (
	"ecommerce-backend/controllers"
	"ecommerce-backend/middlewares"
	"github.com/gin-gonic/gin"
)

// func SetupRoutes(r *gin.Engine) {
// 	api := r.Group("/api")
// 	{

// 		// Public
// 		api.POST("/register", controllers.Register)
// 		api.POST("/login", controllers.Login)
       
// 		// User authenticated
// 		api.Use(middlewares.AuthMiddleware())
// 		{
// 			// Products (user can view)
// 			api.GET("/products", controllers.GetProducts)
// 			api.GET("/products/:id", controllers.GetProductByID)

// 			// Cart
// 			api.POST("/cart", controllers.AddToCart)
// 			api.GET("/cart", controllers.GetCart)
// 			api.PUT("/cart/:id", controllers.UpdateCartQuantity)
// 			api.DELETE("/cart/:id", controllers.RemoveFromCart)

// 			// Orders
// 			api.POST("/order", controllers.CreateOrder)
// 			api.GET("/orders", controllers.GetMyOrders)
			
// 		}

// 		// Admin-only
// 		api.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
// 		{
// 			api.POST("/products", controllers.CreateProduct)
// 			api.DELETE("/products/:id", controllers.DeleteProduct)
// 			api.GET("/admin/orders", controllers.GetAllOrders)
// 		}
// 	}
// }
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
		auth.GET("/products/:id", controllers.GetProductByID)

		// Cart
		auth.POST("/cart", controllers.AddToCart)
		auth.GET("/cart", controllers.GetCart)
		auth.PUT("/cart/:id", controllers.UpdateCartQuantity) // ✅ FIXED
		auth.DELETE("/cart/:id", controllers.RemoveFromCart)

		// Orders
		auth.POST("/order", controllers.CreateOrder)
		auth.GET("/orders", controllers.GetMyOrders)
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
