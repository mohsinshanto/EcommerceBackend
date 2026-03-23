package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"ecommerce-backend/dto"
	"net/http"
	"strconv" 
	"math"
	"github.com/gin-gonic/gin"
)


func GetProducts(c *gin.Context) {
	var products []models.Product
	var total int64

	// --- Query params ---
	search := c.Query("search")
	category := c.Query("category")
	sort := c.Query("sort")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "8"))
	if err != nil || limit < 1 {
		limit = 8
	}

	offset := (page - 1) * limit

	// --- Base query ---
	query := config.DB.Model(&models.Product{})

	// --- Search filter ---
	if search != "" {
		query = query.Where(
			"name LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%",
		)
	}

	// --- Category filter ---
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// --- Price filter ---
	if minPrice != "" {
		min, err := strconv.ParseFloat(minPrice, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_price"})
			return
		}
		query = query.Where("price >= ?", min)
	}

	if maxPrice != "" {
		max, err := strconv.ParseFloat(maxPrice, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_price"})
			return
		}
		query = query.Where("price <= ?", max)
	}

	// --- Count total (before pagination) ---
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
		return
	}

	// --- Sorting ---
	switch sort {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// --- Pagination & Execution ---
	if err := query.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	// --- Response mapping ---
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

	// --- Final JSON response ---
	c.JSON(http.StatusOK, gin.H{
		"products":   response,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"last_page":  int(math.Ceil(float64(total) / float64(limit))),
	})
}

func CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest

	// Bind incoming JSON to DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Map DTO → Model
	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Category:    req.Category, 
	}

	// Save to database
	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Map Model → Response DTO
	response := dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}

	c.JSON(http.StatusCreated, response)
}
func DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")

	// Convert ID to uint
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var product models.Product

	// Explicit primary key query
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Delete and check error
	if err := config.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
