package category

type CreateCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parent_id"`
}
type UpdateCategoryRequest struct {
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parent_id"`
}
type CategoryResponse struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	ParentID *uint              `json:"parent_id"`
	Children []CategoryResponse `json:"children"`
}
