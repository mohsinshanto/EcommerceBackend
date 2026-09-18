package cart

import (
	"ecommerce-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	service *CartService
}

func NewCartController(service *CartService) *CartController {
	return &CartController{service: service}
}

func (c *CartController) AddToCart(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req AddToCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	cartItem, err := c.service.AddToCart(ctx.Request.Context(), userID, req)
	if err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Item added to cart", gin.H{"cart": cartItem})
}

func (c *CartController) GetCart(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cartItems, total, err := c.service.GetUserCart(ctx.Request.Context(), userID)
	if err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]gin.H, 0, len(cartItems))
	for _, item := range cartItems {
		response = append(response, gin.H{
			"id": item.ID, "quantity": item.Quantity,
			"product": gin.H{"id": item.Product.ID, "name": item.Product.Name, "price": item.Product.Price, "image_url": item.Product.ImageURL},
		})
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Cart loaded", gin.H{"items": response, "total": total})
}

func (c *CartController) UpdateCartQuantity(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cartID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid cart item ID")
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"gte=1"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.service.UpdateCartQuantity(ctx.Request.Context(), userID, uint(cartID), req.Quantity); err != nil {
		if err.Error() == "unauthorized" {
			utils.RespondError(ctx, http.StatusForbidden, "Cannot update this cart item")
		} else {
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		}
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Cart updated", nil)
}

func (c *CartController) RemoveFromCart(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cartID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid cart item ID")
		return
	}

	if err := c.service.RemoveFromCart(ctx.Request.Context(), userID, uint(cartID)); err != nil {
		if err.Error() == "unauthorized" {
			utils.RespondError(ctx, http.StatusForbidden, "Cannot remove this cart item")
		} else {
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		}
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Item removed from cart", nil)
}
