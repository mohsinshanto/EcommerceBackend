package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"ecommerce-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	var products []models.Product

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
			utils.RespondError(c, http.StatusBadRequest, "Invalid min_price")
			return
		}
		query = query.Where("price >= ?", min)
	}

	if maxPrice != "" {
		max, err := strconv.ParseFloat(maxPrice, 64)
		if err != nil {
			utils.RespondError(c, http.StatusBadRequest, "Invalid max_price")
			return
		}
		query = query.Where("price <= ?", max)
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
	if err := query.Limit(limit + 1).Offset(offset).Find(&products).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	hasNext := len(products) > limit
	if hasNext {
		products = products[:limit]
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
	utils.RespondSuccess(c, http.StatusOK, "Products loaded", gin.H{
		"products": response,
		"page":     page,
		"limit":    limit,
		"has_next": hasNext,
		"has_prev": page > 1,
	})
}

func CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
	}

	if err := config.DB.Create(&product).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create product")
		return
	}

	response := dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}

	utils.RespondSuccess(c, http.StatusCreated, "Product created successfully", response)
}

func DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product models.Product

	if err := config.DB.First(&product, id).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "Product not found")
		return
	}

	if err := config.DB.Delete(&product).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Product deleted successfully", nil)
}

func GetProductByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product models.Product

	if err := config.DB.First(&product, id).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "Product not found")
		return
	}

	response := dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		Category:    product.Category,
	}

	utils.RespondSuccess(c, http.StatusOK, "Product loaded", response)
}
