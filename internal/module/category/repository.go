package category

import (
	"context"
	"errors"

	appError "ecommerce-backend/errors"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id uint) (*Category, error)
	GetAll(ctx context.Context) ([]Category, error)
	GetAllWithChildren(ctx context.Context) ([]Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id uint) error
	HasChildren(ctx context.Context, id uint) (bool, error)
	GetByParentID(ctx context.Context, parentID uint) ([]Category, error)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appError.ErrCategoryNotFound
		}
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
func (r *categoryRepository) HasChildren(ctx context.Context, id uint) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&Category{}).
		Where("parent_id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
func (r *categoryRepository) GetByParentID(
	ctx context.Context,
	parentID uint,
) ([]Category, error) {
	var categories []Category

	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}
