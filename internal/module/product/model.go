package product

import (
	"ecommerce-backend/internal/module/category"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `gorm:"index:idx_products_deleted_created,priority:2;index:idx_products_category_deleted_created,priority:3" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_products_deleted_created,priority:1;index:idx_products_category_deleted_created,priority:2" json:"-"`

	Name        string  `gorm:"index;size:191" json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url"`

	// Existing category column
	Category string `gorm:"column:category;index;index:idx_products_category_deleted_created,priority:1;size:100" json:"category"`

	// New category relationship
	CategoryID   *uint              `gorm:"index:idx_products_category_id_deleted_created,priority:1" json:"category_id"`
	CategoryInfo *category.Category `gorm:"foreignKey:CategoryID;references:ID" json:"category_info"`
}
