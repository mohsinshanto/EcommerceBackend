package dto

import (
	"ecommerce-backend/models"
	"time"
)

type OrderResponse struct {
	ID         uint      `json:"id"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
}

// Convert Order model to OrderResponse DTO
func ToOrderResponse(order models.Order) OrderResponse {
	return OrderResponse{
		ID:         order.ID,
		TotalPrice: order.TotalPrice,
		CreatedAt:  order.CreatedAt,
	}
}
