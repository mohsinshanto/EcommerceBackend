package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID        uint `gorm:"index:idx_order_user_archived,priority:1;index"`
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
	Archived      bool `gorm:"default:false;index:idx_order_user_archived,priority:2"`
}
