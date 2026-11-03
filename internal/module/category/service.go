package category

import (
	"context"
	appError "ecommerce-backend/errors"
)

type CategoryService interface {
	GetCategoryTree(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, category *Category) error
	UpdateCategory(ctx context.Context, id uint, req UpdateCategoryRequest) error
	DeleteCategory(ctx context.Context, id uint) error
}

type categoryService struct {
	categoryRepo CategoryRepository
}

func NewCategoryService(repo CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: repo,
	}
}

func (s *categoryService) GetCategoryTree(ctx context.Context) ([]Category, error) {
	categories, err := s.categoryRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	tree := buildCategoryTree(categories)
	return tree, nil
}
func (s *categoryService) CreateCategory(ctx context.Context, category *Category) error {
	if category.ParentID != nil {
		_, err := s.categoryRepo.GetByID(ctx, *category.ParentID)
		if err != nil {
			return err
		}
	}

	return s.categoryRepo.Create(ctx, category)
}

func buildCategoryTree(categories []Category) []Category {
	categoryMap := make(map[uint]*Category)

	for i := range categories {
		categories[i].Children = nil
		categoryMap[categories[i].ID] = &categories[i]
	}

	var roots []*Category

	for i := range categories {
		category := &categories[i]

		if category.ParentID == nil {
			roots = append(roots, category)
			continue
		}

		parent, exists := categoryMap[*category.ParentID]
		if exists {
			parent.Children = append(parent.Children, category)
		}
	}

	result := make([]Category, 0, len(roots))

	for _, root := range roots {
		result = append(result, *root)
	}

	return result
}
func ToCategoryResponse(category Category) CategoryResponse {
	response := CategoryResponse{
		ID:       category.ID,
		Name:     category.Name,
		ParentID: category.ParentID,
		Children: make([]CategoryResponse, 0, len(category.Children)),
	}

	for _, child := range category.Children {
		response.Children = append(
			response.Children,
			ToCategoryResponse(*child),
		)
	}

	return response
}
func (s *categoryService) UpdateCategory(ctx context.Context, id uint, req UpdateCategoryRequest) error {

	// 1. Find the category being updated.
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Validate the new parent.
	if req.ParentID != nil {
		parentID := *req.ParentID

		visited := make(map[uint]bool)

		for {
			// A category cannot be its own parent.
			if parentID == id {
				return appError.ErrCircularCategoryReference
			}

			// Prevent infinite loops if existing data is already corrupted.
			if visited[parentID] {
				return appError.ErrExistingCircularReference
			}
			visited[parentID] = true

			parent, err := s.categoryRepo.GetByID(ctx, parentID)
			if err != nil {
				return err
			}

			// Reached the root category.
			if parent.ParentID == nil {
				break
			}

			parentID = *parent.ParentID
		}
	}

	// 3. Update category fields.
	category.Name = req.Name
	category.ParentID = req.ParentID

	return s.categoryRepo.Update(ctx, category)
}
func (s *categoryService) DeleteCategory(ctx context.Context, id uint) error {

	// 1. Check whether the category exists.
	_, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Check whether the category has children.
	hasChildren, err := s.categoryRepo.HasChildren(ctx, id)
	if err != nil {
		return err
	}

	if hasChildren {
		return appError.ErrCategoryHasChildren
	}

	// 3. Delete the category.
	return s.categoryRepo.Delete(ctx, id)
}
