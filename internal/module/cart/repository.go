package cart

import (
	"context"

	"gorm.io/gorm"
)

// CartRepositoryImpl implements the cart repository for the cart module.
type CartRepositoryImpl struct {
	db *gorm.DB
}

type CartRepository interface {
	GetByUserID(ctx context.Context, userID uint) (*Cart, error)
	Create(ctx context.Context, cart *Cart) error
	Update(ctx context.Context, cart *Cart) error
	Delete(ctx context.Context, id uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
}

// NewCartRepository creates a cart repository implementation scoped to the cart module.
func NewCartRepository(db *gorm.DB) CartRepository {
	return &CartRepositoryImpl{db: db}
}

func (r *CartRepositoryImpl) GetByUserID(ctx context.Context, userID uint) (*Cart, error) {
	var cart Cart
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Product").First(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepositoryImpl) Create(ctx context.Context, cart *Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *CartRepositoryImpl) Update(ctx context.Context, cart *Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

func (r *CartRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Cart{}, id).Error
}

func (r *CartRepositoryImpl) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&Cart{}).Error
}
