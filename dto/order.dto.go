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

type AdminOrderItemResponse struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type AdminOrderResponse struct {
	ID            uint                     `json:"id"`
	UserID        uint                     `json:"user_id"`
	UserEmail     string                   `json:"user_email"`
	CustomerName  string                   `json:"customer_name"`
	Phone         string                   `json:"phone"`
	AddressLine   string                   `json:"address_line"`
	City          string                   `json:"city"`
	Area          string                   `json:"area"`
	PostalCode    string                   `json:"postal_code"`
	Notes         string                   `json:"notes"`
	PaymentMethod string                   `json:"payment_method"`
	Status        string                   `json:"status"`
	TotalPrice    float64                  `json:"total_price"`
	Archived      bool                     `json:"archived"`
	CreatedAt     time.Time                `json:"created_at"`
	Items         []AdminOrderItemResponse `json:"items"`
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

func ToAdminOrderResponse(order models.Order) AdminOrderResponse {
	items := make([]AdminOrderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, AdminOrderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.Product.Name,
			Quantity:    item.Quantity,
			Price:       item.Price,
		})
	}

	return AdminOrderResponse{
		ID:            order.ID,
		UserID:        order.UserID,
		UserEmail:     order.User.Email,
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
		Archived:      order.Archived,
		CreatedAt:     order.CreatedAt,
		Items:         items,
	}
}
