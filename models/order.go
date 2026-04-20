package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID        uint
	CustomerName  string
	Phone         string
	AddressLine   string
	City          string
	Area          string
	PostalCode    string
	Notes         string `gorm:"type:text"`
	PaymentMethod string
	Status        string
	TotalPrice    float64
	Archived      bool `gorm:"default:false"`
}
