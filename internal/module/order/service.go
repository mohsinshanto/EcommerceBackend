package order

import (
	"context"
	"ecommerce-backend/config"
	appError "ecommerce-backend/errors"
	"ecommerce-backend/internal/module/cart"
	"ecommerce-backend/internal/module/product"
	"ecommerce-backend/internal/module/user"
	"errors"
	"fmt"
	"strings"
	"time"
)

type OrderService interface {
	CreateCashOnDeliveryOrder(parent context.Context, userID uint, req CreateOrderRequest) (*Order, error)
	CreateSSLCommerzOrder(parent context.Context, userID uint, req CreateOrderRequest) (map[string]interface{}, error)
	GetUserOrders(parent context.Context, userID uint) ([]Order, error)
	GetAllOrders(parent context.Context) ([]Order, error)
	ArchiveOrder(parent context.Context, userID uint, orderID uint) error
}
type orderService struct {
	orderRepo     OrderRepository
	orderItemRepo OrderItemRepository
	productRepo   product.ProductRepository
	userRepo      user.UserRepository
	cartRepo      cart.CartRepository
}

func NewOrderService(
	orderRepo OrderRepository,
	orderItemRepo OrderItemRepository,
	productRepo product.ProductRepository,
	userRepo user.UserRepository,
	cartRepo cart.CartRepository,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		productRepo:   productRepo,
		userRepo:      userRepo,
		cartRepo:      cartRepo,
	}
}

// CreateCashOnDeliveryOrder creates an order with COD payment

func (s *orderService) CreateCashOnDeliveryOrder(parent context.Context, userID uint, req CreateOrderRequest) (*Order, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	tx := s.orderRepo.TransactionDB(ctx)
	if tx.Error != nil {
		return nil, appError.ErrFailedToStartTransaction
	}

	// Get cart items
	cartItems, err := s.cartRepo.GetUserCarts(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, appError.ErrFailedToLoadCart
	}

	if len(cartItems) == 0 {
		tx.Rollback()
		return nil, appError.ErrCartEmpty
	}

	// Calculate total and create order items
	totalPrice := 0.0
	orderItems := make([]OrderItem, 0, len(cartItems))

	for _, cartItem := range cartItems {
		// Check stock
		if cartItem.Product.Stock < cartItem.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("%w for %s", appError.ErrInsufficientStock, cartItem.Product.Name)
		}

		price := cartItem.Product.Price * float64(cartItem.Quantity)
		totalPrice += price

		orderItems = append(orderItems, OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Product.Price,
		})

		// Update product stock
		if err := s.productRepo.DecreaseStock(
			ctx,
			tx,
			cartItem.ProductID,
			cartItem.Quantity,
		); err != nil {
			tx.Rollback()
			return nil, appError.ErrFailedToUpdateStock
		}
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
		PaymentMethod: "cod",
		PaymentStatus: "Pending",
		Status:        "Pending",
		Currency:      "BDT",
		TotalPrice:    totalPrice,
	}

	// Create order
	if err := s.orderRepo.Create(ctx, tx, &order); err != nil {
		tx.Rollback()
		return nil, appError.ErrFailedToCreateOrder
	}
	// Create order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}
	if err := s.orderItemRepo.CreateMany(
		ctx,
		tx,
		orderItems,
	); err != nil {
		tx.Rollback()

		return nil, appError.ErrFailedToCreateOrderItems
	}

	// Clear cart
	if err := s.cartRepo.ClearCart(
		ctx,
		tx,
		userID,
	); err != nil {

		tx.Rollback()

		return nil, appError.ErrFailedToClearCart
	}

	if err := tx.Commit().Error; err != nil {
		return nil, appError.ErrTransactionFailed
	}

	return &order, nil
}

// CreateSSLCommerzOrder creates an order with SSLCommerz payment
func (s *orderService) CreateSSLCommerzOrder(parent context.Context, userID uint, req CreateOrderRequest) (map[string]interface{}, error) {
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
func (s *orderService) GetUserOrders(parent context.Context, userID uint) ([]Order, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to load orders")
	}

	return orders, nil
}

// GetAllOrders retrieves all orders (admin only)
func (s *orderService) GetAllOrders(parent context.Context) ([]Order, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	orders, err := s.orderRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.New("failed to load orders")
	}

	return orders, nil
}

// ArchiveOrder archives an order
func (s *orderService) ArchiveOrder(parent context.Context, userID uint, orderID uint) error {
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

func NormalizeOrderRequest(req *CreateOrderRequest) {
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.AddressLine = strings.TrimSpace(req.AddressLine)
	req.City = strings.TrimSpace(req.City)
	req.Area = strings.TrimSpace(req.Area)
	req.PostalCode = strings.TrimSpace(req.PostalCode)
	req.Notes = strings.TrimSpace(req.Notes)
}

func ValidateOrderRequest(req CreateOrderRequest) error {
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
	if req.Area == "" {
		return errors.New("area is required")
	}
	return nil
}
