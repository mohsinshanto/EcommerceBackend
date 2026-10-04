package order

import (
	"context"

	"gorm.io/gorm"
)

type OrderRepository interface {
	GetByID(ctx context.Context, id uint) (*Order, error)
	GetByUserID(ctx context.Context, userID uint) ([]Order, error)
	GetAll(ctx context.Context) ([]Order, error)
	Create(ctx context.Context, tx *gorm.DB, order *Order) error
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id uint) error
	TransactionDB(ctx context.Context) *gorm.DB
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}
func (r *orderRepository) TransactionDB(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Begin()
}
func (r *orderRepository) GetByID(ctx context.Context, id uint) (*Order, error) {
	var order Order
	if err := r.db.WithContext(ctx).Preload("Items").Preload("User").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByUserID(ctx context.Context, userID uint) ([]Order, error) {
	var orders []Order
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Items").Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) GetAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	if err := r.db.WithContext(ctx).Preload("Items").Preload("User").Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	order *Order,
) error {
	return tx.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) Update(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Order{}, id).Error
}

type OrderItemRepository interface {
	GetByOrderID(ctx context.Context, orderID uint) ([]OrderItem, error)
	Create(ctx context.Context, tx *gorm.DB, orderItem *OrderItem) error
	Update(ctx context.Context, item *OrderItem) error
	Delete(ctx context.Context, id uint) error
	CreateMany(ctx context.Context, tx *gorm.DB, orderItems []OrderItem) error
}

type orderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID uint) ([]OrderItem, error) {
	var items []OrderItem
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Preload("Product").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *orderItemRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	orderItem *OrderItem,
) error {
	return tx.WithContext(ctx).Create(orderItem).Error
}
func (r *orderItemRepository) CreateMany(
	ctx context.Context,
	tx *gorm.DB,
	orderItems []OrderItem,
) error {
	return tx.WithContext(ctx).Create(&orderItems).Error
}
func (r *orderItemRepository) Update(ctx context.Context, item *OrderItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *orderItemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&OrderItem{}, id).Error
}
