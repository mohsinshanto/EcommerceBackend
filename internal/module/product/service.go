package product

import (
	"context"
	"ecommerce-backend/config"
	"log"
	"strconv"
	"strings"
	"time"
)

type ProductService interface {
	CreateProduct(parent context.Context, req CreateProductRequest) (ProductResponse, error)
	GetProducts(parent context.Context, params GetProductsParams) ([]ProductResponse, int64, *int64, bool, error)
	GetCategoryCounts(parent context.Context, params GetProductsParams) map[string]int
	DeleteProduct(parent context.Context, id int) error
	GetProductCount(parent context.Context) (int64, error)
	GetProductByID(parent context.Context, id int) (ProductResponse, error)
}
type productService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) GetProducts(parent context.Context, params GetProductsParams) ([]ProductResponse, int64, *int64, bool, error) {
	queryCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	offset := max((params.Page-1)*params.Limit, 0)

	searchQuery := s.buildProductSearchQuery(params.Search)
	if searchQuery != "" && config.Typesense != nil {
		products, found, hasNext, err := SearchProductsWithTypesense(queryCtx, params)
		if err == nil {
			totalCount, countErr := s.GetProductCount(queryCtx)
			if countErr != nil {
				return nil, 0, nil, false, countErr
			}
			return products, totalCount, &found, hasNext, nil
		}
	}

	// Build filters for repository
	minPrice := 0.0
	maxPrice := 0.0
	hasPrice := false

	if params.MinPrice != "" {
		val, err := strconv.ParseFloat(params.MinPrice, 64)
		if err != nil {
			return nil, 0, nil, false, err
		}
		minPrice = val
		hasPrice = true
	}

	if params.MaxPrice != "" {
		val, err := strconv.ParseFloat(params.MaxPrice, 64)
		if err != nil {
			return nil, 0, nil, false, err
		}
		maxPrice = val
		hasPrice = true
	}

	filters := ProductFilters{
		Search:   searchQuery,
		Category: params.Category,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Sort:     params.Sort,
		Limit:    params.Limit + 1,
		Offset:   offset,
		HasPrice: hasPrice,
	}
	totalProduct, err := s.GetProductCount(queryCtx)
	if err != nil {
		return nil, 0, nil, false, err
	}
	products, err := s.repo.GetWithFilters(queryCtx, filters)
	if err != nil {
		return nil, 0, nil, false, err
	}

	hasNext := len(products) > params.Limit
	if hasNext {
		products = products[:params.Limit]
	}

	// Map to response DTOs
	response := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		response = append(response, ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			ImageURL:    p.ImageURL,
			Category:    p.Category,
		})
	}

	return response, totalProduct, nil, hasNext, nil
}

func (s *productService) GetCategoryCounts(parent context.Context, params GetProductsParams) map[string]int {
	defaultCounts := map[string]int{
		"mobile":      0,
		"laptop":      0,
		"audio":       0,
		"accessories": 0,
	}

	if config.Typesense == nil {
		return defaultCounts
	}

	queryCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	counts, err := GetCategoryCountsWithTypesense(queryCtx, params)
	if err != nil {
		return defaultCounts
	}

	for key, value := range counts {
		defaultCounts[key] = value
	}

	return defaultCounts
}
func (s *productService) CreateProduct(parent context.Context, req CreateProductRequest) (ProductResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	product := Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	}

	err := s.repo.Create(ctx, &product)
	if err != nil {
		return ProductResponse{}, err
	}

	if config.RedisClient != nil {
		config.RedisClient.Incr(ctx, productCountCacheKey)
	}

	if err := UpsertProductInTypesense(ctx, product); err != nil {
		log.Printf("WARNING: failed to sync product %d to Typesense: %v", product.ID, err)
	}

	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}, nil
}
func (s *productService) DeleteProduct(parent context.Context, id int) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	product, err := s.repo.GetByID(ctx, uint(id))
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, product.ID)
	if err != nil {
		return err
	}

	if config.RedisClient != nil {
		config.RedisClient.Decr(ctx, productCountCacheKey)
	}

	if err := DeleteProductFromTypesense(ctx, product.ID); err != nil {
		log.Printf("WARNING: failed to delete product %d from Typesense: %v", product.ID, err)
	}

	return nil
}

func (s *productService) GetProductCount(parent context.Context) (int64, error) {
	queryCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	if config.RedisClient != nil {
		cachedCount, err := config.RedisClient.Get(queryCtx, productCountCacheKey).Int64()
		if err == nil {
			return cachedCount, nil
		}
	}

	count, err := s.repo.Count(queryCtx)
	if err != nil {
		return 0, err
	}

	if config.RedisClient != nil {
		config.RedisClient.Set(queryCtx, productCountCacheKey, count, 0)
	}

	return count, nil
}

func (s *productService) GetProductByID(parent context.Context, id int) (ProductResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	product, err := s.repo.GetByID(ctx, uint(id))
	if err != nil {
		return ProductResponse{}, err
	}

	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}, nil
}

func (s *productService) buildProductSearchQuery(search string) string {
	terms := strings.Fields(search)
	if len(terms) == 0 {
		return ""
	}

	processed := make([]string, 0, len(terms))
	for _, term := range terms {
		cleaned := strings.Trim(term, `+-<>~*"()@`)
		if cleaned == "" {
			continue
		}

		processed = append(processed, cleaned)
	}

	return strings.Join(processed, " ")
}
