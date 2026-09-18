package order

import (
	"context"

	"gorm.io/gorm"
)

// OrderRepositoryImpl implements the order repository for the order module.
type OrderRepositoryImpl struct {
	db *gorm.DB
}

type OrderRepository interface {
	GetByID(ctx context.Context, id uint) (*Order, error)
	GetByUserID(ctx context.Context, userID uint) ([]Order, error)
	GetAll(ctx context.Context) ([]Order, error)
	Create(ctx context.Context, order *Order) error
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id uint) error
}

type OrderItemRepository interface {
	GetByOrderID(ctx context.Context, orderID uint) ([]OrderItem, error)
	Create(ctx context.Context, item *OrderItem) error
	Update(ctx context.Context, item *OrderItem) error
	Delete(ctx context.Context, id uint) error
}

// NewOrderRepository creates an order repository implementation scoped to the order module.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &OrderRepositoryImpl{db: db}
}

func (r *OrderRepositoryImpl) GetByID(ctx context.Context, id uint) (*Order, error) {
	var order Order
	if err := r.db.WithContext(ctx).Preload("Items").Preload("User").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepositoryImpl) GetByUserID(ctx context.Context, userID uint) ([]Order, error) {
	var orders []Order
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Items").Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepositoryImpl) GetAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	if err := r.db.WithContext(ctx).Preload("Items").Preload("User").Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepositoryImpl) Create(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepositoryImpl) Update(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *OrderRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Order{}, id).Error
}

// OrderItemRepositoryImpl implements the order item repository for the order module.
type OrderItemRepositoryImpl struct {
	db *gorm.DB
}

// NewOrderItemRepository creates an order item repository implementation scoped to the order module.
func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &OrderItemRepositoryImpl{db: db}
}

func (r *OrderItemRepositoryImpl) GetByOrderID(ctx context.Context, orderID uint) ([]OrderItem, error) {
	var items []OrderItem
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Preload("Product").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *OrderItemRepositoryImpl) Create(ctx context.Context, item *OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *OrderItemRepositoryImpl) Update(ctx context.Context, item *OrderItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *OrderItemRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&OrderItem{}, id).Error
}
