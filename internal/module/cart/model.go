package cart

import (
	"ecommerce-backend/internal/module/product"

	"gorm.io/gorm"
)

type Cart struct {
	gorm.Model
	UserID    uint `gorm:"index:idx_cart_user_product,priority:1;index"`
	ProductID uint `gorm:"index:idx_cart_user_product,priority:2;index"`
	Quantity  int
	Product   product.Product `gorm:"foreignKey:ProductID"`
}
