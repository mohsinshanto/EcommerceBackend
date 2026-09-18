package product

import (
	"context"

	"gorm.io/gorm"
)

type ProductRepository interface {
	GetByID(ctx context.Context, id uint) (*Product, error)
	GetAll(ctx context.Context) ([]Product, error)
	GetWithFilters(ctx context.Context, filters ProductFilters) ([]Product, error)
	Create(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
	GetByIDWithDeleted(ctx context.Context, id uint) (*Product, error)
}
type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetByID(ctx context.Context, id uint) (*Product, error) {
	var product Product
	if err := r.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetAll(ctx context.Context) ([]Product, error) {
	var products []Product
	if err := r.db.WithContext(ctx).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *productRepository) GetWithFilters(ctx context.Context, filters ProductFilters) ([]Product, error) {
	query := r.db.WithContext(ctx).Model(&Product{})

	if filters.Search != "" {
		query = query.Where("MATCH(name) AGAINST(? IN BOOLEAN MODE)", filters.Search)
	}
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.HasPrice {
		if filters.MinPrice > 0 {
			query = query.Where("price >= ?", filters.MinPrice)
		}
		if filters.MaxPrice > 0 {
			query = query.Where("price <= ?", filters.MaxPrice)
		}
	}

	switch filters.Sort {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	default:
		query = query.Order("created_at DESC")
	}

	var products []Product
	if err := query.Limit(filters.Limit).Offset(filters.Offset).Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepository) Create(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) Update(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Product{}, id).Error
}

func (r *productRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Product{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *productRepository) GetByIDWithDeleted(ctx context.Context, id uint) (*Product, error) {
	var product Product
	if err := r.db.WithContext(ctx).Unscoped().First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
