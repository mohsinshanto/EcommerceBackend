package product

const productCountCacheKey = "product:count"

type ProductResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock" binding:"required"`
	ImageURL    string  `json:"image_url" binding:"required"`
	Category    string  `json:"category" binding:"required"`
}

// For service
type ProductFilters struct {
	Search   string
	Category string
	MinPrice float64
	MaxPrice float64
	Sort     string
	Limit    int
	Offset   int
	HasPrice bool
}

// For controller
type GetProductsParams struct {
	Search   string
	Category string
	Sort     string
	MinPrice string
	MaxPrice string
	Page     int
	Limit    int
}
