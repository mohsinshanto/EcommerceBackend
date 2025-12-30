package models

import "gorm.io/gorm"

type Cart struct {
	gorm.Model
	UserID    uint
	ProductID uint
	Quantity  int
	Product   Product `gorm:"foreignKey:ProductID"`
}
