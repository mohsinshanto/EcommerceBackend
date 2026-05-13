package models

import "gorm.io/gorm"

type OrderItem struct {
	gorm.Model
	OrderID   uint `gorm:"index"`
	ProductID uint `gorm:"index"`
	Quantity  int
	Price     float64
	Product   Product `gorm:"foreignKey:ProductID"`
}
