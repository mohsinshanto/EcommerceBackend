package order

import (
	"ecommerce-backend/config"
	appError "ecommerce-backend/errors"
	"ecommerce-backend/utils"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	service OrderService
}

func NewOrderController(service OrderService) *OrderController {
	return &OrderController{service: service}
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	var req CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	NormalizeOrderRequest(&req)
	if err := ValidateOrderRequest(req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	switch strings.ToLower(req.PaymentMethod) {
	case "cod":
		response, err := c.service.CreateCashOnDeliveryOrder(ctx.Request.Context(), userID, req)
		if err != nil {
			switch {
			case errors.Is(err, appError.ErrCartEmpty):
				utils.RespondError(ctx, http.StatusBadRequest, err.Error())

			case errors.Is(err, appError.ErrInsufficientStock):
				utils.RespondError(ctx, http.StatusBadRequest, err.Error())

			case errors.Is(err, appError.ErrFailedToCreateOrder),
				errors.Is(err, appError.ErrFailedToCreateOrderItems),
				errors.Is(err, appError.ErrFailedToUpdateStock),
				errors.Is(err, appError.ErrFailedToClearCart),
				errors.Is(err, appError.ErrTransactionFailed):
				utils.RespondError(ctx, http.StatusInternalServerError, err.Error())

			default:
				utils.RespondError(ctx, http.StatusInternalServerError, "Failed to place order")
			}

			return
		}
		utils.RespondSuccess(ctx, http.StatusCreated, "Order placed successfully", response)
	case "sslcommerz":
		result, err := c.service.CreateSSLCommerzOrder(ctx.Request.Context(), userID, req)
		if err != nil {
			switch {
			case errors.Is(err, appError.ErrSSLCommerzNotConfigured):
				utils.RespondError(ctx, http.StatusServiceUnavailable, err.Error())
			case errors.Is(err, appError.ErrCartEmpty):
				utils.RespondError(ctx, http.StatusBadRequest, err.Error())
			case errors.Is(err, appError.ErrInsufficientStock):
				utils.RespondError(ctx, http.StatusConflict, err.Error())
			case errors.Is(err, appError.ErrPaymentSessionUnknown):
				utils.RespondError(ctx, http.StatusBadGateway, err.Error())
			case errors.Is(err, appError.ErrPaymentSessionSaveFailed):
				utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
			case errors.Is(err, appError.ErrFailedToStartTransaction),
				errors.Is(err, appError.ErrFailedToLoadCart),
				errors.Is(err, appError.ErrFailedToCreateOrder),
				errors.Is(err, appError.ErrFailedToCreateOrderItems),
				errors.Is(err, appError.ErrFailedToUpdateStock),
				errors.Is(err, appError.ErrFailedToClearCart),
				errors.Is(err, appError.ErrTransactionFailed):
				utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
			default:
				utils.RespondError(ctx, http.StatusInternalServerError, "Failed to create SSLCommerz order")
			}
			return
		}
		utils.RespondSuccess(ctx, http.StatusCreated, "Order created, proceed to payment", result)
	default:
		utils.RespondError(ctx, http.StatusBadRequest, "Unsupported payment method")
	}
}

func (c *OrderController) GetMyOrders(ctx *gin.Context) {
	orders, err := c.service.GetUserOrders(ctx.Request.Context(), ctx.GetUint("user_id"))
	if err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]gin.H, 0, len(orders))
	for _, item := range orders {
		response = append(response, gin.H{"id": item.ID, "total_price": item.TotalPrice, "payment_method": item.PaymentMethod, "status": item.Status, "created_at": item.CreatedAt})
	}
	utils.RespondSuccess(ctx, http.StatusOK, "Orders loaded", gin.H{"orders": response})
}

func (c *OrderController) GetAllOrders(ctx *gin.Context) {
	orders, err := c.service.GetAllOrders(ctx.Request.Context())
	if err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]gin.H, 0, len(orders))
	for _, item := range orders {
		response = append(response, gin.H{"id": item.ID, "user_id": item.UserID, "customer_name": item.CustomerName, "total_price": item.TotalPrice, "payment_method": item.PaymentMethod, "status": item.Status, "created_at": item.CreatedAt})
	}
	utils.RespondSuccess(ctx, http.StatusOK, "All orders", gin.H{"orders": response})
}

func (c *OrderController) ArchiveOrder(ctx *gin.Context) {
	orderID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid order ID")
		return
	}

	if err := c.service.ArchiveOrder(ctx.Request.Context(), ctx.GetUint("user_id"), uint(orderID)); err != nil {
		if err.Error() == "unauthorized" {
			utils.RespondError(ctx, http.StatusForbidden, "Cannot archive this order")
		} else {
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		}
		return
	}
	utils.RespondSuccess(ctx, http.StatusOK, "Order archived", nil)
}
func (c *OrderController) SSLCommerzSuccess(ctx *gin.Context) {
	// Parse callback data from either POST form or GET query.
	if err := ctx.Request.ParseForm(); err != nil {
		utils.RespondError(
			ctx,
			http.StatusBadRequest,
			"Failed to parse payment callback",
		)
		return
	}

	transactionID := strings.TrimSpace(ctx.Request.FormValue("tran_id"))
	validationID := strings.TrimSpace(ctx.Request.FormValue("val_id"))

	if transactionID == "" || validationID == "" {
		utils.RespondError(
			ctx,
			http.StatusBadRequest,
			appError.ErrMissingPaymentInformation.Error(),
		)
		return
	}

	err := c.service.CompleteSSLCommerzPayment(
		ctx.Request.Context(),
		transactionID,
		validationID,
	)
	if err != nil {
		switch {
		case errors.Is(err, appError.ErrMissingPaymentInformation),
			errors.Is(err, appError.ErrPaymentValidationFailed),
			errors.Is(err, appError.ErrPaymentTransactionMismatch),
			errors.Is(err, appError.ErrPaymentValidationIDMismatch),
			errors.Is(err, appError.ErrPaymentAmountMismatch),
			errors.Is(err, appError.ErrPaymentCurrencyMismatch),
			errors.Is(err, appError.ErrInvalidPaymentMethod):

			utils.RespondError(
				ctx,
				http.StatusBadRequest,
				"Payment could not be verified",
			)

		case errors.Is(err, appError.ErrOrderNotFound):
			utils.RespondError(
				ctx,
				http.StatusNotFound,
				appError.ErrOrderNotFound.Error(),
			)

		default:
			utils.RespondError(
				ctx,
				http.StatusInternalServerError,
				"Failed to complete payment",
			)
		}
		return
	}

	ctx.Redirect(
		http.StatusSeeOther,
		config.SSLCommerz.FrontendURL+"/payment/success",
	)
}

func (c *OrderController) SSLCommerzFail(ctx *gin.Context)   { SSLCommerzFail(ctx) }
func (c *OrderController) SSLCommerzCancel(ctx *gin.Context) { SSLCommerzCancel(ctx) }
func (c *OrderController) SSLCommerzIPN(ctx *gin.Context)    { SSLCommerzIPN(ctx) }
