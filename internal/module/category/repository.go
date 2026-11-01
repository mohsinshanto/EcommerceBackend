package category

import (
	"context"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uint) (*Category, error)
	GetAll(ctx context.Context) ([]Category, error)
	GetAllWithChildren(ctx context.Context) ([]Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id uint) (*Category, error) {
	var category Category

	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) GetAll(ctx context.Context) ([]Category, error) {
	var categories []Category

	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}
func (r *categoryRepository) GetAllWithChildren(ctx context.Context) ([]Category, error) {
	var categories []Category

	err := r.db.WithContext(ctx).
		Preload("Children").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Category{}, id).Error
}
