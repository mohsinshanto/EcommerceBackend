package cart

import (
	"context"
	appError "ecommerce-backend/errors"
	"ecommerce-backend/internal/module/product"
	"errors"
	"time"

	"gorm.io/gorm"
)

type CartService interface {
	AddToCart(parent context.Context, userID uint, req AddToCartRequest) (*CartItemResponse, error)
	GetUserCart(parent context.Context, userID uint) (*CartResponse, error)
	UpdateCartQuantity(parent context.Context, userID uint, cartID uint, quantity int) error
	RemoveFromCart(parent context.Context, userID uint, cartID uint) error
	ClearCart(parent context.Context, userID uint) error
}

// CartService handles cart business logic
type cartService struct {
	cartRepo    CartRepository
	productRepo product.ProductRepository
}

// NewCartService creates a new cart service instance
func NewCartService(cartRepo CartRepository, productRepo product.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) AddToCart(parent context.Context, userID uint, req AddToCartRequest) (*CartItemResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, appError.ErrProductNotFound
	}

	if product.Stock < req.Quantity {
		return nil, appError.ErrInsufficientStock
	}
	cartItem, err := s.cartRepo.GetByProductID(ctx, userID, req.ProductID)

	if err == nil {
		// Item exists, update quantity
		newQuantity := cartItem.Quantity + req.Quantity
		if product.Stock < newQuantity {
			return nil, appError.ErrInsufficientStock
		}
		cartItem.Quantity = newQuantity
		if err := s.cartRepo.Update(ctx, cartItem); err != nil {
			return nil, appError.ErrUpdateCart
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// New item, create it
		cartItem = &Cart{
			UserID:    userID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := s.cartRepo.Create(ctx, cartItem); err != nil {
			return nil, appError.ErrFailedToAddCart
		}
	} else {
		return nil, appError.ErrFailedToLoadCart
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

// get the user cart
func (s *cartService) GetUserCart(parent context.Context, userID uint) (*CartResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cartItems, err := s.cartRepo.GetUserCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate total
	total := 0.0
	items := make([]CartItemResponse, 0, len(cartItems))
	for _, item := range cartItems {
		total += float64(item.Quantity) * item.Product.Price
		items = append(items, CartItemResponse{
			ID:       item.ID,
			Quantity: item.Quantity,
			Product: CartProductResponse{
				ID:       item.Product.ID,
				Name:     item.Product.Name,
				Price:    item.Product.Price,
				ImageURL: item.Product.ImageURL,
			},
		})
	}
	response := &CartResponse{
		CartItems: items,
		Total:     total,
	}
	return response, nil

}

// UpdateCartQuantity updates the quantity of an item in cart
func (s *cartService) UpdateCartQuantity(parent context.Context, userID uint, cartID uint, quantity int) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cartItem, err := s.cartRepo.FindCartByCartID(ctx, cartID)
	if err != nil {
		return appError.ErrCartNotFound
	}
	if cartItem.UserID != userID {
		return appError.ErrForbidden
	}

	if quantity <= 0 {
		return appError.ErrQuantityValidation
	}

	// Check stock
	product, err := s.productRepo.GetByID(ctx, cartItem.ProductID)
	if err != nil {
		return appError.ErrProductNotFound
	}

	if product.Stock < quantity {
		return appError.ErrInsufficientStock
	}

	cartItem.Quantity = quantity
	if err := s.cartRepo.Update(ctx, cartItem); err != nil {
		return appError.ErrUpdateCart
	}

	return nil
}

// RemoveFromCart removes an item from the user's cart
func (s *cartService) RemoveFromCart(parent context.Context, userID uint, cartID uint) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cartItem, err := s.cartRepo.FindCartByCartID(ctx, cartID)
	if err != nil {
		return appError.ErrCartNotFound
	}
	if cartItem.UserID != userID {
		return appError.ErrForbidden
	}

	if err := s.cartRepo.Delete(ctx, cartID); err != nil {
		return appError.ErrRemoveCartItem
	}

	return nil
}

// ClearCart removes all items from user's cart
func (s *cartService) ClearCart(parent context.Context, userID uint) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	return s.cartRepo.DeleteByUserID(ctx, userID)
}
