package services

import (
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"strconv"
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
	var products []models.Product

	offset := (params.Page - 1) * params.Limit

	query := config.DB.Model(&models.Product{})

	// search
	if params.Search != "" {
		query = query.Where(
			"name LIKE ? OR description LIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%",
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
	if err := config.DB.Model(&models.Product{}).Count(&count).Error; err != nil {
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
