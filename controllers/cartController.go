package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)
func AddToCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var body dto.AddToCartRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var product models.Product
	if err := config.DB.First(&product, body.ProductID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
		return
	}

	if product.Stock < body.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
		return
	}

	var cart models.Cart
	err := config.DB.
		Where("user_id = ? AND product_id = ?", userID, body.ProductID).
		First(&cart).Error

	if err == nil {
		cart.Quantity += body.Quantity
		config.DB.Save(&cart)
	} else {
		cart = models.Cart{
			UserID:    userID,
			ProductID: body.ProductID,
			Quantity:  body.Quantity,
		}
		config.DB.Create(&cart)
	}

	config.DB.Preload("Product").First(&cart)

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart updated",
		"cart": gin.H{
			"id":       cart.ID,
			"quantity": cart.Quantity,
			"product": gin.H{
				"name":  cart.Product.Name,
				"price": cart.Product.Price,
			},
		},
	})
}


func GetCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	var cartItems []models.Cart
	if err := config.DB.Preload("Product").
		Where("user_id = ?", userID).
		Find(&cartItems).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// --- Build clean response ---
	response := make([]gin.H, 0) // ✅ IMPORTANT

	for _, item := range cartItems {
		response = append(response, gin.H{
			"id":       item.ID,
			"quantity": item.Quantity,
			"product": gin.H{
				"id":    item.ProductID,
				"name":  item.Product.Name,
				"price": item.Product.Price,
			},
		})
	}

	c.JSON(200, gin.H{
		"cart": response, // always []
	})
}


func RemoveFromCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	cartID := c.Param("id")

	var cart models.Cart

	// Check if the item exists and belongs to the user
	if err := config.DB.
		Where("id = ? AND user_id = ?", cartID, userID).
		First(&cart).Error; err != nil {

		c.JSON(404, gin.H{"error": "Cart item not found"})
		return
	}

	// Delete the item
	if err := config.DB.Delete(&cart).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to remove item"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Item removed from cart",
		"id":      cart.ID,
	})
}
