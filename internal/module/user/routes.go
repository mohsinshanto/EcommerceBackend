package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	Controller *UserController
}

func NewModule(db *gorm.DB) *Module {
	repo := NewUserRepository(db)
	service := NewUserService(repo)
	controller := NewUserController(service)
	return &Module{
		Controller: controller,
	}
}
func (m *Module) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	api.POST("/register", m.Controller.Register)
	api.POST("/login", m.Controller.Login)
}
