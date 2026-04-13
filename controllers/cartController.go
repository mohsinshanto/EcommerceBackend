package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"ecommerce-backend/utils"
	"net/http"
	"strconv"

	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AddToCart(c *gin.Context) {
	userID := c.GetUint("user_id")

	if userID == 0 {
		utils.RespondError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var body dto.AddToCartRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	var product models.Product
	if err := config.DB.First(&product, body.ProductID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "Product not found")
		return
	}

	if product.Stock < body.Quantity {
		utils.RespondError(c, http.StatusBadRequest, "Insufficient stock")
		return
	}

	var cart models.Cart
	err := config.DB.
		Where("user_id = ? AND product_id = ?", userID, body.ProductID).
		First(&cart).Error

	if err == nil {
		cart.Quantity += body.Quantity
		if err := config.DB.Save(&cart).Error; err != nil {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to update cart")
			return
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		cart = models.Cart{
			UserID:    userID,
			ProductID: body.ProductID,
			Quantity:  body.Quantity,
		}
		if err := config.DB.Create(&cart).Error; err != nil {
			utils.RespondError(c, http.StatusInternalServerError, "Failed to add item to cart")
			return
		}
	} else {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to load cart")
		return
	}

	if err := config.DB.Preload("Product").First(&cart, cart.ID).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to load updated cart item")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Cart updated", gin.H{
		"cart": gin.H{
			"id":       cart.ID,
			"quantity": cart.Quantity,
			"product": gin.H{
				"id":        cart.Product.ID,
				"name":      cart.Product.Name,
				"price":     cart.Product.Price,
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

		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch cart")
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

	utils.RespondSuccess(c, http.StatusOK, "Cart loaded", gin.H{
		"cart": response,
	})
}

func UpdateCartQuantity(c *gin.Context) {
	userID := c.GetUint("user_id")

	cartIDParam := c.Param("id")
	cartID, err := strconv.Atoi(cartIDParam)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid cart ID")
		return
	}

	var body struct {
		Quantity int `json:"quantity" binding:"required,gte=1"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	var cart models.Cart
	if err := config.DB.
		Where("id = ? AND user_id = ?", cartID, userID).
		First(&cart).Error; err != nil {

		utils.RespondError(c, http.StatusNotFound, "Cart item not found")
		return
	}

	// ✅ Check stock
	var product models.Product
	if err := config.DB.First(&product, cart.ProductID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "Product not found")
		return
	}

	if product.Stock < body.Quantity {
		utils.RespondError(c, http.StatusBadRequest, "Insufficient stock")
		return
	}

	cart.Quantity = body.Quantity
	if err := config.DB.Save(&cart).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update quantity")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Quantity updated", nil)
}

func RemoveFromCart(c *gin.Context) {
	userID := c.GetUint("user_id")
	cartIDParam := c.Param("id")
	cartID, err := strconv.Atoi(cartIDParam)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid cart ID")
		return
	}
	var cart models.Cart

	if err := config.DB.
		Where("id = ? AND user_id = ?", cartID, userID).
		First(&cart).Error; err != nil {

		utils.RespondError(c, http.StatusNotFound, "Cart item not found")
		return
	}

	if err := config.DB.Delete(&cart).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to remove item")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Item removed from cart", gin.H{
		"id": cart.ID,
	})
}
