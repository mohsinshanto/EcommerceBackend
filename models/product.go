package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `gorm:"index:idx_products_deleted_created,priority:2;index:idx_products_category_deleted_created,priority:3" json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index:idx_products_deleted_created,priority:1;index:idx_products_category_deleted_created,priority:2" json:"-"`
	Name        string         `gorm:"index;size:191" json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Stock       int            `json:"stock"`
	ImageURL    string         `json:"image_url"`
	Category    string         `gorm:"index;index:idx_products_category_deleted_created,priority:1;size:100" json:"category"`
}
