package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name        string  `gorm:"index;size:191" json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url"`
	Category    string  `gorm:"index;size:100" json:"category"`
}
