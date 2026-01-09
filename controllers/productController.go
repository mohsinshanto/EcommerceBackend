package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"net/http"

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
	var response []models.ProductResponse
	for _, p := range products {
		response = append(response, models.ProductResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
		})
	}
	c.JSON(http.StatusOK, gin.H{"products": response})
}

func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&product)
	c.JSON(http.StatusOK, product)
}
func DeleteProduct(c *gin.Context) {
	id := c.Param("id") // get product id from URL

	var product models.Product
	if err := config.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	config.DB.Delete(&product)
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
