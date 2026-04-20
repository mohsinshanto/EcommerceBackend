package dto

import (
	"ecommerce-backend/models"
	"time"
)

type OrderResponse struct {
	ID            uint      `json:"id"`
	CustomerName  string    `json:"customer_name"`
	Phone         string    `json:"phone"`
	AddressLine   string    `json:"address_line"`
	City          string    `json:"city"`
	Area          string    `json:"area"`
	PostalCode    string    `json:"postal_code"`
	Notes         string    `json:"notes"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	TotalPrice    float64   `json:"total_price"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateOrderRequest struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	Phone         string `json:"phone" binding:"required,min=8"`
	AddressLine   string `json:"address_line" binding:"required"`
	City          string `json:"city" binding:"required"`
	Area          string `json:"area" binding:"required"`
	PostalCode    string `json:"postal_code"`
	Notes         string `json:"notes"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=cod"`
}

// Convert Order model to OrderResponse DTO
func ToOrderResponse(order models.Order) OrderResponse {
	return OrderResponse{
		ID:            order.ID,
		CustomerName:  order.CustomerName,
		Phone:         order.Phone,
		AddressLine:   order.AddressLine,
		City:          order.City,
		Area:          order.Area,
		PostalCode:    order.PostalCode,
		Notes:         order.Notes,
		PaymentMethod: order.PaymentMethod,
		Status:        order.Status,
		TotalPrice:    order.TotalPrice,
		CreatedAt:     order.CreatedAt,
	}
}
