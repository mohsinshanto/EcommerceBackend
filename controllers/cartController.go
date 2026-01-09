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

	var body dto.AddToCartRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cart models.Cart
	err := config.DB.
		Where("user_id = ? AND product_id = ?", userID, body.ProductID).
		First(&cart).Error

	// If exists → update
	if err == nil {
		cart.Quantity += body.Quantity
		if err := config.DB.Save(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
			return
		}
	} else {
		// Create new
		cart = models.Cart{
			UserID:    userID,
			ProductID: body.ProductID,
			Quantity:  body.Quantity,
		}

		if err := config.DB.Create(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to cart"})
			return
		}
	}

	// Load product
	config.DB.Preload("Product").First(&cart)

	resp := dto.CartItemResponse{
		Quantity: cart.Quantity,
		Product: dto.CartProductResponse{
			Name:  cart.Product.Name,
			Price: cart.Product.Price,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart updated",
		"cart":    resp,
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
	var response []gin.H

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
		"cart": response,
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
