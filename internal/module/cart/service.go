package cart

import (
	"context"
	"ecommerce-backend/config"
	"ecommerce-backend/internal/module/product"
	"errors"
	"time"

	"gorm.io/gorm"
)

// CartService handles cart business logic
type CartService struct {
	cartRepo    CartRepository
	productRepo product.ProductRepository
}

// NewCartService creates a new cart service instance
func NewCartService(cartRepo CartRepository, productRepo product.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// AddToCart adds an item to the user's cart or updates quantity if exists
func (s *CartService) AddToCart(parent context.Context, userID uint, req AddToCartRequest) (*CartItemResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	// Validate product exists and has stock
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	// Check if item already in cart
	// We need to modify the cart repository to support getting by user and product
	// For now, we'll work with the database directly
	var cartItem Cart
	err = config.DB.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, req.ProductID).
		First(&cartItem).Error

	if err == nil {
		// Item exists, update quantity
		newQuantity := cartItem.Quantity + req.Quantity
		if product.Stock < newQuantity {
			return nil, errors.New("insufficient stock")
		}
		cartItem.Quantity = newQuantity
		if err := s.cartRepo.Update(ctx, &cartItem); err != nil {
			return nil, errors.New("failed to update cart")
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// New item, create it
		cartItem = Cart{
			UserID:    userID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := s.cartRepo.Create(ctx, &cartItem); err != nil {
			return nil, errors.New("failed to add item to cart")
		}
	} else {
		return nil, errors.New("failed to load cart")
	}

	return &CartItemResponse{
		ID:       cartItem.ID,
		Quantity: cartItem.Quantity,
		Product: CartProductResponse{
			ID:       product.ID,
			Name:     product.Name,
			Price:    product.Price,
			ImageURL: product.ImageURL,
		},
	}, nil
}

// GetUserCart retrieves all cart items for a user
func (s *CartService) GetUserCart(parent context.Context, userID uint) ([]Cart, float64, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	// Get all cart items for user
	var cartItems []Cart
	if err := config.DB.WithContext(ctx).Preload("Product").
		Where("user_id = ?", userID).
		Find(&cartItems).Error; err != nil {
		return nil, 0, errors.New("failed to load cart")
	}

	// Calculate total
	total := 0.0
	for _, item := range cartItems {
		total += float64(item.Quantity) * item.Product.Price
	}

	return cartItems, total, nil
}

// UpdateCartQuantity updates the quantity of an item in cart
func (s *CartService) UpdateCartQuantity(parent context.Context, userID uint, cartID uint, quantity int) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	var cartItem Cart
	if err := config.DB.WithContext(ctx).First(&cartItem, cartID).Error; err != nil {
		return errors.New("cart item not found")
	}

	if cartItem.UserID != userID {
		return errors.New("unauthorized")
	}

	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	// Check stock
	product, err := s.productRepo.GetByID(ctx, cartItem.ProductID)
	if err != nil {
		return errors.New("product not found")
	}

	if product.Stock < quantity {
		return errors.New("insufficient stock")
	}

	cartItem.Quantity = quantity
	if err := s.cartRepo.Update(ctx, &cartItem); err != nil {
		return errors.New("failed to update cart")
	}

	return nil
}

// RemoveFromCart removes an item from the user's cart
func (s *CartService) RemoveFromCart(parent context.Context, userID uint, cartID uint) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	var cartItem Cart
	if err := config.DB.WithContext(ctx).First(&cartItem, cartID).Error; err != nil {
		return errors.New("cart item not found")
	}

	if cartItem.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.cartRepo.Delete(ctx, cartID); err != nil {
		return errors.New("failed to remove item from cart")
	}

	return nil
}

// ClearCart removes all items from user's cart
func (s *CartService) ClearCart(parent context.Context, userID uint) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	return s.cartRepo.DeleteByUserID(ctx, userID)
}
