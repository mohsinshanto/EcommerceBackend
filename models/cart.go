package models

import "gorm.io/gorm"

type Cart struct {
	gorm.Model
	UserID    uint `gorm:"index:idx_cart_user_product,priority:1;index"`
	ProductID uint `gorm:"index:idx_cart_user_product,priority:2;index"`
	Quantity  int
	Product   Product `gorm:"foreignKey:ProductID"`
}
