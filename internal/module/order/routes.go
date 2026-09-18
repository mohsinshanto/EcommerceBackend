package order

import (
	"ecommerce-backend/internal/module/middleware"
	"ecommerce-backend/internal/module/product"
	"ecommerce-backend/internal/module/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Controller *OrderController
}

func NewModule(db *gorm.DB) *Module {
	orderRepo := NewOrderRepository(db)
	orderItemRepo := NewOrderItemRepository(db)
	productRepo := product.NewProductRepository(db)
	userRepo := user.NewUserRepository(db)
	service := NewOrderService(orderRepo, orderItemRepo, productRepo, userRepo)
	controller := NewOrderController(service)

	return &Module{
		Controller: controller,
	}
}
func (m *Module) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	api.POST("/payments/sslcommerz/success", m.Controller.SSLCommerzSuccess)
	api.POST("/payments/sslcommerz/fail", m.Controller.SSLCommerzFail)
	api.POST("/payments/sslcommerz/cancel", m.Controller.SSLCommerzCancel)
	api.POST("/payments/sslcommerz/ipn", m.Controller.SSLCommerzIPN)
	api.GET("/payments/sslcommerz/success", m.Controller.SSLCommerzSuccess)
	api.GET("/payments/sslcommerz/fail", m.Controller.SSLCommerzFail)
	api.GET("/payments/sslcommerz/cancel", m.Controller.SSLCommerzCancel)

	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())
	auth.POST("/order", m.Controller.CreateOrder)
	auth.GET("/orders", m.Controller.GetMyOrders)
	auth.PATCH("/orders/:id/archive", m.Controller.ArchiveOrder)

	admin := api.Group("/")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
	admin.GET("/admin/orders", m.Controller.GetAllOrders)
}
