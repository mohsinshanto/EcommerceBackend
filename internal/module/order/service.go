package order

import (
	"context"
	appError "ecommerce-backend/errors"
	"ecommerce-backend/internal/module/cart"
	"ecommerce-backend/internal/module/product"
	"ecommerce-backend/internal/module/user"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateCashOnDeliveryOrder(parent context.Context, userID uint, req CreateOrderRequest) (*OrderResponse, error)
	CreateSSLCommerzOrder(parent context.Context, userID uint, req CreateOrderRequest) (*OrderResponse, error)
	CompleteSSLCommerzPayment(parent context.Context, transactionID string, validationID string) (*PaymentResult, error)
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

func (s *orderService) CreateCashOnDeliveryOrder(parent context.Context, userID uint, req CreateOrderRequest) (*OrderResponse, error) {
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
	response := OrderResponse{
		ID:            order.ID,
		TotalPrice:    order.TotalPrice,
		PaymentMethod: order.PaymentMethod,
		Status:        order.Status,
	}

	return &response, nil

}

// CreateSSLCommerzOrder creates an order with SSLCommerz payment
func (s *orderService) CreateSSLCommerzOrder(parent context.Context, userID uint, req CreateOrderRequest) (*OrderResponse, error) {
	if !SSLCommerzEnabled() {
		return nil, appError.ErrSSLCommerzNotConfigured
	}

	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Start transaction
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

	// Calculate total
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
	transactionID := "TRX-" + uuid.NewString()
	order := Order{
		UserID:        userID,
		TransactionID: transactionID,
		CustomerName:  req.CustomerName,
		Phone:         req.Phone,
		AddressLine:   req.AddressLine,
		City:          req.City,
		Area:          req.Area,
		PostalCode:    req.PostalCode,
		Notes:         req.Notes,
		PaymentMethod: "sslcommerz",
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
	// Use a separate timeout for saving the recovery status.
	// The original ctx may have expired during the gateway request.
	recoveryCtx, recoveryCancel := context.WithTimeout(
		context.WithoutCancel(parent),
		3*time.Second,
	)
	defer recoveryCancel()

	if err != nil {
		// We cannot assume the gateway failed to create the session.
		// A timeout may mean that the result is unknown.
		updateErr := s.orderRepo.UpdatePaymentSession(
			recoveryCtx,
			order.ID,
			"unknown",
			"",
		)

		if updateErr != nil {
			return nil, fmt.Errorf(
				"%w payment session result unknown: %v; failed to save recovery status: %w",
				appError.ErrPaymentSessionUnknown,
				err,
				updateErr,
			)
		}

		return nil, fmt.Errorf(
			"%w payment session result unknown for order %d: %w",
			appError.ErrPaymentSessionUnknown,
			order.ID,
			err,
		)
	}

	// Gateway returned a session URL.
	// Save it before returning the URL to the frontend.
	if err := s.orderRepo.UpdatePaymentSession(
		recoveryCtx,
		order.ID,
		"ready",
		session.GatewayPageURL,
	); err != nil {
		return nil, fmt.Errorf(
			"%w payment session created but failed to save session details for order %d: %w",
			appError.ErrPaymentSessionSaveFailed,
			order.ID,
			err,
		)
	}

	return &OrderResponse{
		ID:            order.ID,
		TransactionID: transactionID,
		TotalPrice:    order.TotalPrice,
		PaymentURL:    session.GatewayPageURL,
		PaymentMethod: "sslcommerz",
	}, nil
}

// confirming payment
func (s *orderService) CompleteSSLCommerzPayment(
	parent context.Context,
	transactionID string,
	validationID string,
) (*PaymentResult, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	transactionID = strings.TrimSpace(transactionID)
	validationID = strings.TrimSpace(validationID)

	if transactionID == "" || validationID == "" {
		return nil, appError.ErrMissingPaymentInformation
	}

	// Step 1: Validate payment with SSLCommerz.
	validation, err := ValidateSSLCommerzPayment(ctx, validationID)
	if err != nil {
		return nil, fmt.Errorf("validate SSLCommerz payment: %w", err)
	}

	if validation == nil {
		return nil, appError.ErrPaymentValidationFailed
	}

	// Step 2: Check validation status.
	status := strings.ToUpper(strings.TrimSpace(validation.Status))
	if status != "VALID" && status != "VALIDATED" {
		return nil, appError.ErrPaymentValidationFailed
	}

	// Step 3: Verify transaction IDs.
	if strings.TrimSpace(validation.TransactionID) != transactionID {
		return nil, appError.ErrPaymentTransactionMismatch
	}

	returnedValidationID := strings.TrimSpace(validation.ValidationID)
	if returnedValidationID == "" ||
		returnedValidationID != validationID {
		return nil, appError.ErrPaymentValidationIDMismatch
	}

	// Step 4: Find the order.
	order, err := s.orderRepo.GetByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("find order by transaction ID: %w", err)
	}

	if order == nil {
		return nil, appError.ErrOrderNotFound
	}

	if !strings.EqualFold(
		strings.TrimSpace(order.PaymentMethod),
		"sslcommerz",
	) {
		return nil, appError.ErrInvalidPaymentMethod
	}

	// Step 5: Parse and verify the payment amount.
	amount, err := strconv.ParseFloat(validation.Amount, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %v",
			appError.ErrInvalidPaymentAmount,
			err,
		)
	}

	if math.IsNaN(amount) ||
		math.IsInf(amount, 0) ||
		amount < 0 {
		return nil, appError.ErrInvalidPaymentAmount
	}

	if math.Abs(amount-order.TotalPrice) > 0.009 {
		return nil, appError.ErrPaymentAmountMismatch
	}

	// Step 6: Verify currency.
	if !strings.EqualFold(
		strings.TrimSpace(validation.Currency),
		strings.TrimSpace(order.Currency),
	) {
		return nil, appError.ErrPaymentCurrencyMismatch
	}

	// Step 7: Handle repeated callbacks.
	if strings.EqualFold(
		strings.TrimSpace(order.PaymentStatus),
		"Paid",
	) {
		return &PaymentResult{
			OrderID:       order.ID,
			TransactionID: order.TransactionID,
		}, nil
	}

	// Step 8: Prepare payment information.
	order.PaymentStatus = "Paid"
	order.ValidationID = returnedValidationID
	order.GatewayAmount = amount
	order.Currency = validation.Currency
	order.BankTranID = validation.BankTranID
	order.CardType = validation.CardType
	order.SessionKey = validation.SessionKey

	// Step 9: Atomically update payment if it is not already paid.
	updated, err := s.orderRepo.UpdatePaymentIfUnpaid(ctx, order)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %w",
			appError.ErrFailedToUpdatePayment,
			err,
		)
	}

	// Step 10: If another callback updated the order first,
	// reload it and verify that the payment is now paid.
	if !updated {
		latestOrder, err := s.orderRepo.GetByTransactionID(
			ctx,
			transactionID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"reload order after payment update: %w",
				err,
			)
		}

		if latestOrder == nil {
			return nil, appError.ErrOrderNotFound
		}

		if !strings.EqualFold(
			strings.TrimSpace(latestOrder.PaymentStatus),
			"Paid",
		) {
			return nil, appError.ErrFailedToUpdatePayment
		}

		order = latestOrder
	}

	return &PaymentResult{
		OrderID:       order.ID,
		TransactionID: order.TransactionID,
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
