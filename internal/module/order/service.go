package order

import (
	"context"
	"ecommerce-backend/config"
	"ecommerce-backend/internal/module/cart"
	"ecommerce-backend/internal/module/product"
	"ecommerce-backend/internal/module/user"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// OrderService handles order business logic
type OrderService struct {
	orderRepo     OrderRepository
	orderItemRepo OrderItemRepository
	productRepo   product.ProductRepository
	userRepo      user.UserRepository
}

// NewOrderService creates a new order service instance
func NewOrderService(
	orderRepo OrderRepository,
	orderItemRepo OrderItemRepository,
	productRepo product.ProductRepository,
	userRepo user.UserRepository,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		productRepo:   productRepo,
		userRepo:      userRepo,
	}
}

// CreateCashOnDeliveryOrder creates an order with COD payment

func (s *OrderService) CreateCashOnDeliveryOrder(parent context.Context, userID uint, req CreateOrderRequest) (*Order, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	// Start transaction
	tx := config.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}

	order := Order{
		UserID:        userID,
		CustomerName:  req.CustomerName,
		Phone:         req.Phone,
		AddressLine:   req.AddressLine,
		City:          req.City,
		Area:          req.Area,
		PostalCode:    req.PostalCode,
		Notes:         req.Notes,
		PaymentMethod: "Cash on Delivery",
		PaymentStatus: "Pending",
		Status:        "Pending",
		Currency:      "BDT",
	}

	// Get cart items
	var cartItems []cart.Cart
	if err := tx.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to load cart")
	}

	if len(cartItems) == 0 {
		tx.Rollback()
		return nil, errors.New("cart is empty")
	}

	// Calculate total and create order items
	totalPrice := 0.0
	orderItems := make([]OrderItem, 0, len(cartItems))

	for _, cartItem := range cartItems {
		// Check stock
		if cartItem.Product.Stock < cartItem.Quantity {
			tx.Rollback()
			return nil, errors.New("insufficient stock for " + cartItem.Product.Name)
		}

		price := cartItem.Product.Price * float64(cartItem.Quantity)
		totalPrice += price

		orderItems = append(orderItems, OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Product.Price,
		})

		// Update product stock
		if err := tx.Model(&cartItem.Product).
			Update("stock", gorm.Expr("stock - ?", cartItem.Quantity)).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to update product stock")
		}
	}

	order.TotalPrice = totalPrice

	// Create order
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create order")
	}

	// Create order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to create order items")
		}
	}

	// Clear cart
	if err := tx.Where("user_id = ?", userID).Delete(&cart.Cart{}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to clear cart")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, errors.New("transaction failed")
	}

	return &order, nil
}

// CreateSSLCommerzOrder creates an order with SSLCommerz payment
func (s *OrderService) CreateSSLCommerzOrder(parent context.Context, userID uint, req CreateOrderRequest) (map[string]interface{}, error) {
	if !SSLCommerzEnabled() {
		return nil, errors.New("sslcommerz is not configured")
	}

	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Start transaction
	tx := config.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}

	order := Order{
		UserID:        userID,
		CustomerName:  req.CustomerName,
		Phone:         req.Phone,
		AddressLine:   req.AddressLine,
		City:          req.City,
		Area:          req.Area,
		PostalCode:    req.PostalCode,
		Notes:         req.Notes,
		PaymentMethod: "SSLCommerz",
		PaymentStatus: "Pending",
		Status:        "Pending",
		Currency:      "BDT",
	}

	// Get cart items
	var cartItems []cart.Cart
	if err := tx.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to load cart")
	}

	if len(cartItems) == 0 {
		tx.Rollback()
		return nil, errors.New("cart is empty")
	}

	// Calculate total
	totalPrice := 0.0
	for _, item := range cartItems {
		totalPrice += item.Product.Price * float64(item.Quantity)
	}

	order.TotalPrice = totalPrice

	// Create order
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create order")
	}

	// Create order items
	for _, cartItem := range cartItems {
		item := OrderItem{
			OrderID:   order.ID,
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Product.Price,
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to create order item")
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, errors.New("transaction failed")
	}

	transactionID := "TRX-" + time.Now().Format("20060102150405")
	session, err := CreateSSLCommerzSession(ctx, SSLCommerzSessionRequest{
		TransactionID: transactionID,
		Amount:        order.TotalPrice,
		Currency:      order.Currency,
		ProductName:   "E-Commerce Products",
		ProductNames:  []string{"E-Commerce Products"},
		Category:      "general",
		CustomerName:  user.Name,
		CustomerEmail: user.Email,
		Phone:         req.Phone,
		AddressLine:   req.AddressLine,
		City:          req.City,
		Area:          req.Area,
		PostalCode:    req.PostalCode,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"order_id":       order.ID,
		"transaction_id": transactionID,
		"total_amount":   order.TotalPrice,
		"payment_url":    session.GatewayPageURL,
		"payment_method": "sslcommerz",
	}, nil
}

// GetUserOrders retrieves all orders for a user
func (s *OrderService) GetUserOrders(parent context.Context, userID uint) ([]Order, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to load orders")
	}

	return orders, nil
}

// GetAllOrders retrieves all orders (admin only)
func (s *OrderService) GetAllOrders(parent context.Context) ([]Order, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	orders, err := s.orderRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.New("failed to load orders")
	}

	return orders, nil
}

// ArchiveOrder archives an order
func (s *OrderService) ArchiveOrder(parent context.Context, userID uint, orderID uint) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.UserID != userID {
		return errors.New("unauthorized")
	}

	order.Archived = true
	return s.orderRepo.Update(ctx, order)
}

// NormalizeOrderRequest normalizes order request data
func (s *OrderService) NormalizeOrderRequest(req *CreateOrderRequest) {
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.AddressLine = strings.TrimSpace(req.AddressLine)
	req.City = strings.TrimSpace(req.City)
	req.Area = strings.TrimSpace(req.Area)
	req.PostalCode = strings.TrimSpace(req.PostalCode)
	req.Notes = strings.TrimSpace(req.Notes)
}

// ValidateOrderRequest validates order request data
func (s *OrderService) ValidateOrderRequest(req CreateOrderRequest) error {
	if req.CustomerName == "" {
		return errors.New("customer name is required")
	}
	if req.Phone == "" {
		return errors.New("phone is required")
	}
	if req.AddressLine == "" {
		return errors.New("address is required")
	}
	if req.City == "" {
		return errors.New("city is required")
	}
	if req.PostalCode == "" {
		return errors.New("postal code is required")
	}
	return nil
}
