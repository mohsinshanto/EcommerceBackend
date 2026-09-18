package product

import (
	"ecommerce-backend/internal/module/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Controller *ProductController
}

func NewModule(db *gorm.DB) *Module {
	repo := NewProductRepository(db)
	service := NewProductService(repo)
	controller := NewProductController(service)
	return &Module{
		Controller: controller,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	api.GET("/products", m.Controller.GetProducts)
	api.GET("/products/count", m.Controller.GetProductCount)
	api.GET("/products/:id", m.Controller.GetProductByID)

	admin := api.Group("/")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
	admin.POST("/products", m.Controller.CreateProduct)
	admin.DELETE("/products/:id", m.Controller.DeleteProduct)
}
