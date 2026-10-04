package cart

import (
	"context"

	"gorm.io/gorm"
)

type cartRepository struct {
	db *gorm.DB
}

type CartRepository interface {
	GetByUserID(ctx context.Context, userID uint) (*Cart, error)
	Create(ctx context.Context, cart *Cart) error
	Update(ctx context.Context, cart *Cart) error
	Delete(ctx context.Context, id uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
	GetByProductID(ctx context.Context, userID uint, productID uint) (*Cart, error)
	GetUserCart(ctx context.Context, userID uint) ([]Cart, error)
	GetUserCarts(ctx context.Context, tx *gorm.DB, userID uint) ([]Cart, error)
	FindCartByCartID(ctx context.Context, cartID uint) (*Cart, error)
	ClearCart(ctx context.Context, tx *gorm.DB, userID uint) error
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}
func (r *cartRepository) FindCartByCartID(ctx context.Context, cartID uint) (*Cart, error) {
	var cartItem Cart
	if err := r.db.WithContext(ctx).First(&cartItem, cartID).Error; err != nil {
		return nil, err
	}
	return &cartItem, nil
}
func (r *cartRepository) GetUserCart(ctx context.Context, userID uint) ([]Cart, error) {
	var cartItems []Cart
	err := r.db.WithContext(ctx).Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error
	if err != nil {
		return nil, err
	}
	return cartItems, nil
}
func (r *cartRepository) GetUserCarts(ctx context.Context, tx *gorm.DB, userID uint) ([]Cart, error) {
	var cartItems []Cart
	err := tx.WithContext(ctx).Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error
	if err != nil {
		return nil, err
	}
	return cartItems, nil
}
func (r *cartRepository) GetByUserID(ctx context.Context, userID uint) (*Cart, error) {
	var cart Cart
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Product").First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) Create(ctx context.Context, cart *Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *cartRepository) Update(ctx context.Context, cart *Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

func (r *cartRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Cart{}, id).Error
}

func (r *cartRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&Cart{}).Error
}
func (r *cartRepository) GetByProductID(ctx context.Context, userID uint, productID uint) (*Cart, error) {
	var cartItem Cart
	err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).
		First(&cartItem).Error
	return &cartItem, err
}
func (r *cartRepository) ClearCart(
	ctx context.Context,
	tx *gorm.DB,
	userID uint,
) error {
	return tx.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&Cart{}).Error
}
