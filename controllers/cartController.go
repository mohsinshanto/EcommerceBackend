package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"net/http"
	"strconv"
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
				"id":    cart.Product.ID,
	            "name":  cart.Product.Name,
	            "price": cart.Product.Price,
	            "image_url": cart.Product.ImageURL,
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

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dto.CartItemResponse

	for _, item := range cartItems {
		response = append(response, dto.CartItemResponse{
			ID:       item.ID,
			Quantity: item.Quantity,
			Product: dto.CartProductResponse{
				ID:       item.Product.ID,
				Name:     item.Product.Name,
				Price:    item.Product.Price,
				ImageURL: item.Product.ImageURL, // ✅ important
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"cart": response,
	})
}
func UpdateCartQuantity(c *gin.Context) {
	userID := c.GetUint("user_id")

	cartIDParam := c.Param("id")
	cartID, err := strconv.Atoi(cartIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
		return
	}

	var body struct {
		Quantity int `json:"quantity" binding:"required,gte=1"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cart models.Cart
	if err := config.DB.
		Where("id = ? AND user_id = ?", cartID, userID).
		First(&cart).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	// ✅ Check stock
	var product models.Product
	if err := config.DB.First(&product, cart.ProductID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
		return
	}

	if product.Stock < body.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
		return
	}

	// ✅ Update quantity
	cart.Quantity = body.Quantity
	config.DB.Save(&cart)

	c.JSON(http.StatusOK, gin.H{
		"message": "Quantity updated",
	})
}
func RemoveFromCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	cartIDParam := c.Param("id")
    cartID, err := strconv.Atoi(cartIDParam)
    if err != nil {
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
	return
   }
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
