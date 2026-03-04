package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"ecommerce-backend/dto"
	"net/http"
	"strconv" 

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	var products []models.Product

	if err := config.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	if len(products) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No products found"})
		return
	}

	var response []dto.ProductResponse

	for _, p := range products {
		response = append(response, dto.ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			ImageURL:    p.ImageURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{"products": response})
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
