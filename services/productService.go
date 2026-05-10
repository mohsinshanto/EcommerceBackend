package services

import (
	"context"
	"database/sql"
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const productCountCacheKey = "product:count"

type GetProductsParams struct {
	Search   string
	Category string
	Sort     string
	MinPrice string
	MaxPrice string
	Page     int
	Limit    int
}

func GetProducts(params GetProductsParams) ([]dto.ProductResponse, int64, bool, error) {
	queryCtx, cancel := context.WithTimeout(config.Ctx, 5*time.Second)
	defer cancel()

	offset := max((params.Page-1)*params.Limit, 0)

	searchQuery := buildProductSearchQuery(params.Search)
	if searchQuery != "" && config.Typesense != nil {
		products, hasNext, err := SearchProductsWithTypesense(queryCtx, params)
		if err == nil {
			totalCount, countErr := GetProductCount()
			if countErr != nil {
				return nil, 0, false, countErr
			}
			return products, totalCount, hasNext, nil
		}
	}

	var products []models.Product
	query := config.DB.WithContext(queryCtx).Model(&models.Product{})

	if searchQuery != "" {
		query = query.Where(
			"MATCH(name) AGAINST(? IN BOOLEAN MODE)",
			searchQuery,
		)
	}

	// category
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}

	// price
	if params.MinPrice != "" {
		min, err := strconv.ParseFloat(params.MinPrice, 64)
		if err != nil {
			return nil, 0, false, err
		}
		query = query.Where("price >= ?", min)
	}

	if params.MaxPrice != "" {
		max, err := strconv.ParseFloat(params.MaxPrice, 64)
		if err != nil {
			return nil, 0, false, err
		}
		query = query.Where("price <= ?", max)
	}

	// sort
	switch params.Sort {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// execute
	if err := query.Limit(params.Limit + 1).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, false, err
	}

	hasNext := len(products) > params.Limit
	if hasNext {
		products = products[:params.Limit]
	}

	// map response
	var response []dto.ProductResponse
	for _, p := range products {
		response = append(response, dto.ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			ImageURL:    p.ImageURL,
			Category:    p.Category,
		})
	}

	totalCount, err := GetProductCount()
	if err != nil {
		return nil, 0, false, err
	}

	return response, totalCount, hasNext, nil
}

func getProductsWithRankedSearch(
	queryCtx context.Context,
	params GetProductsParams,
	searchQuery string,
	offset int,
) ([]dto.ProductResponse, int64, bool, error) {
	type rankedProductRow struct {
		ID          uint
		Name        string
		Description string
		Price       float64
		Stock       int
		ImageURL    string
		Category    string
		CreatedAt   time.Time
		Relevance   sql.NullFloat64
	}

	limit := params.Limit + 1
	args := []interface{}{searchQuery, searchQuery}
	filters := []string{
		"MATCH(name) AGAINST(? IN BOOLEAN MODE)",
		"deleted_at IS NULL",
	}

	if params.Category != "" {
		filters = append(filters, "category = ?")
		args = append(args, params.Category)
	}

	if params.MinPrice != "" {
		min, err := strconv.ParseFloat(params.MinPrice, 64)
		if err != nil {
			return nil, 0, false, err
		}
		filters = append(filters, "price >= ?")
		args = append(args, min)
	}

	if params.MaxPrice != "" {
		max, err := strconv.ParseFloat(params.MaxPrice, 64)
		if err != nil {
			return nil, 0, false, err
		}
		filters = append(filters, "price <= ?")
		args = append(args, max)
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.price, p.stock, p.image_url, p.category, p.created_at, ranked.relevance
		FROM (
			SELECT id, MATCH(name) AGAINST(? IN BOOLEAN MODE) AS relevance
			FROM products
			WHERE %s
			ORDER BY relevance DESC
			LIMIT ?, ?
		) AS ranked
		JOIN products p ON p.id = ranked.id
		ORDER BY ranked.relevance DESC, p.created_at DESC
	`, strings.Join(filters, " AND "))

	args = append(args, offset, limit)
	rows := make([]rankedProductRow, 0, limit)
	if err := config.DB.WithContext(queryCtx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, 0, false, err
	}

	hasNext := len(rows) > params.Limit
	if hasNext {
		rows = rows[:params.Limit]
	}

	response := make([]dto.ProductResponse, 0, len(rows))
	for _, row := range rows {
		response = append(response, dto.ProductResponse{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			Price:       row.Price,
			Stock:       row.Stock,
			ImageURL:    row.ImageURL,
			Category:    row.Category,
		})
	}

	totalCount, err := GetProductCount()
	if err != nil {
		return nil, 0, false, err
	}

	return response, totalCount, hasNext, nil
}
func CreateProduct(req dto.CreateProductRequest) (dto.ProductResponse, error) {
	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	}

	err := config.DB.Create(&product).Error
	if err != nil {
		return dto.ProductResponse{}, err
	}

	if config.RedisClient != nil {
		config.RedisClient.Incr(config.Ctx, productCountCacheKey)
	}

	if err := UpsertProductInTypesense(config.Ctx, product); err != nil {
		log.Printf("WARNING: failed to sync product %d to Typesense: %v", product.ID, err)
	}

	return dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}, nil
}
func DeleteProduct(id int) error {
	var product models.Product

	if err := config.DB.First(&product, id).Error; err != nil {
		return err
	}

	if err := config.DB.Delete(&product).Error; err != nil {
		return err
	}

	if config.RedisClient != nil {
		config.RedisClient.Decr(config.Ctx, productCountCacheKey)
	}

	if err := DeleteProductFromTypesense(config.Ctx, product.ID); err != nil {
		log.Printf("WARNING: failed to delete product %d from Typesense: %v", product.ID, err)
	}

	return nil
}

func GetProductCount() (int64, error) {
	if config.RedisClient != nil {
		cachedCount, err := config.RedisClient.Get(config.Ctx, productCountCacheKey).Int64()
		if err == nil {
			return cachedCount, nil
		}
	}

	var count int64
	queryCtx, cancel := context.WithTimeout(config.Ctx, 5*time.Second)
	defer cancel()

	if err := config.DB.WithContext(queryCtx).Model(&models.Product{}).Count(&count).Error; err != nil {
		return 0, err
	}

	if config.RedisClient != nil {
		config.RedisClient.Set(config.Ctx, productCountCacheKey, count, 0)
	}

	return count, nil
}

func GetProductByID(id int) (dto.ProductResponse, error) {
	var product models.Product

	if err := config.DB.First(&product, id).Error; err != nil {
		return dto.ProductResponse{}, err
	}

	return dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}, nil
}

func buildProductSearchQuery(search string) string {
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

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
