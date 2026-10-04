package cart

import (
	appError "ecommerce-backend/errors"
	"ecommerce-backend/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	service CartService
}

func NewCartController(service CartService) *CartController {
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

	utils.RespondSuccess(ctx, http.StatusOK, "Item added to cart", cartItem)
}

func (c *CartController) GetCart(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondError(ctx, http.StatusUnauthorized, "Unauthorized")
		return
	}

	response, err := c.service.GetUserCart(ctx.Request.Context(), userID)
	if err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Cart loaded", response)
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
	err = c.service.UpdateCartQuantity(ctx.Request.Context(), userID, uint(cartID), req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, appError.ErrCartNotFound),
			errors.Is(err, appError.ErrQuantityValidation):
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		case errors.Is(err, appError.ErrInsufficientStock):
			utils.RespondError(ctx, http.StatusConflict, err.Error())
		case errors.Is(err, appError.ErrForbidden):
			utils.RespondError(ctx, http.StatusForbidden, err.Error())
		case errors.Is(err, appError.ErrProductNotFound):
			utils.RespondError(ctx, http.StatusNotFound, err.Error())
		default:
			utils.RespondError(ctx, http.StatusInternalServerError, err.Error())

		}
		return
	}

	// here to use switch to improve
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

	err = c.service.RemoveFromCart(ctx.Request.Context(), userID, uint(cartID))
	if err != nil {
		switch {
		case errors.Is(err, appError.ErrCartNotFound):
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		case errors.Is(err, appError.ErrForbidden):
			utils.RespondError(ctx, http.StatusForbidden, err.Error())
		case errors.Is(err, appError.ErrRemoveCartItem):
			utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		default:
			utils.RespondError(ctx, http.StatusInternalServerError, "internal server error")

		}
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Item removed from cart", nil)
}
