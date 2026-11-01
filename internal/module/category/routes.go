package category

import (
	"ecommerce-backend/internal/module/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Controller *CategoryController
}

func NewModule(db *gorm.DB) *Module {
	repo := NewCategoryRepository(db)
	service := NewCategoryService(repo)
	controller := NewCategoryController(service)

	return &Module{
		Controller: controller,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")

	// Public route
	api.GET("/categories", m.Controller.GetCategoryTree)

	// Admin-only route
	admin := api.Group("/")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())

	admin.POST("/categories", m.Controller.CreateCategory)
}
