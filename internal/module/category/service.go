package category

import "context"

type CategoryService interface {
	GetCategoryTree(ctx context.Context) ([]Category, error)
	CreateCategory(ctx context.Context, category *Category) error
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

	return buildCategoryTree(categories), nil
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

	// Step 1: Store every category in a map.
	for i := range categories {
		categories[i].Children = nil
		categoryMap[categories[i].ID] = &categories[i]
	}

	var roots []Category

	// Step 2: Connect children with their parents.
	for i := range categories {
		category := &categories[i]

		if category.ParentID == nil {
			roots = append(roots, *category)
			continue
		}

		parent, exists := categoryMap[*category.ParentID]
		if exists {
			parent.Children = append(parent.Children, *category)
		}
	}

	return roots
}
