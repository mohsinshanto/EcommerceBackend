package cart

import (
	"ecommerce-backend/internal/module/middleware"
	"ecommerce-backend/internal/module/product"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Controller *CartController
}

func NewModule(db *gorm.DB) *Module {
	cartRepo := NewCartRepository(db)
	productRepo := product.NewProductRepository(db)
	service := NewCartService(cartRepo, productRepo)
	controller := NewCartController(service)

	return &Module{
		Controller: controller,
	}
}
func (m *Module) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())
	auth.POST("/cart", m.Controller.AddToCart)
	auth.GET("/cart", m.Controller.GetCart)
	auth.PUT("/cart/:id", m.Controller.UpdateCartQuantity)
	auth.DELETE("/cart/:id", m.Controller.RemoveFromCart)
}
