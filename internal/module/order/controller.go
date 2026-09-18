package order

import (
	"ecommerce-backend/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// OrderController handles HTTP requests for the order module.
type OrderController struct {
	service *OrderService
}

func NewOrderController(service *OrderService) *OrderController {
	return &OrderController{service: service}
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	var req CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	c.service.NormalizeOrderRequest(&req)
	if err := c.service.ValidateOrderRequest(req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	switch strings.ToLower(req.PaymentMethod) {
	case "cod", "cash":
		createdOrder, err := c.service.CreateCashOnDeliveryOrder(ctx.Request.Context(), userID, req)
		if err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondSuccess(ctx, http.StatusCreated, "Order placed successfully", gin.H{"order_id": createdOrder.ID, "total_paid": createdOrder.TotalPrice, "status": createdOrder.Status, "payment_method": req.PaymentMethod})
	case "sslcommerz":
		result, err := c.service.CreateSSLCommerzOrder(ctx.Request.Context(), userID, req)
		if err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())
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

func (c *OrderController) SSLCommerzSuccess(ctx *gin.Context) { SSLCommerzSuccess(ctx) }
func (c *OrderController) SSLCommerzFail(ctx *gin.Context)    { SSLCommerzFail(ctx) }
func (c *OrderController) SSLCommerzCancel(ctx *gin.Context)  { SSLCommerzCancel(ctx) }
func (c *OrderController) SSLCommerzIPN(ctx *gin.Context)     { SSLCommerzIPN(ctx) }
