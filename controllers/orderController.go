package controllers

import (
	"context"
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"ecommerce-backend/services"
	"ecommerce-backend/utils"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req dto.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	normalizeOrderRequest(&req)
	if err := validateOrderRequest(req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	switch req.PaymentMethod {
	case "cod":
		createCashOnDeliveryOrder(c, userID, req)
	case "sslcommerz":
		createSSLCommerzOrder(c, userID, req)
	default:
		utils.RespondError(c, http.StatusBadRequest, "Unsupported payment method")
	}
}

func createCashOnDeliveryOrder(c *gin.Context, userID uint, req dto.CreateOrderRequest) {
	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	cart, order, err := createOrderFromCart(tx, userID, req, models.Order{
		PaymentMethod: "Cash on Delivery",
		PaymentStatus: "Pending",
		Status:        "Pending",
		Currency:      "BDT",
	})
	if err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := finalizeReservedOrder(tx, userID, cart, &order); err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Transaction failed")
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Order placed successfully", gin.H{
		"order_id":       order.ID,
		"total_paid":     order.TotalPrice,
		"status":         order.Status,
		"payment_method": req.PaymentMethod,
	})
}

func createSSLCommerzOrder(c *gin.Context, userID uint, req dto.CreateOrderRequest) {
	if !services.SSLCommerzEnabled() {
		utils.RespondError(c, http.StatusServiceUnavailable, "SSLCommerz is not configured on the server")
		return
	}

	var user models.User
	if err := config.DB.Select("id", "email").First(&user, userID).Error; err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "User not found")
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	transactionID := buildTransactionID(userID)
	cart, order, err := createOrderFromCart(tx, userID, req, models.Order{
		PaymentMethod: "SSLCommerz",
		PaymentStatus: "Initiated",
		Status:        "Pending Payment",
		Currency:      "BDT",
		TransactionID: transactionID,
	})
	if err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := finalizeReservedOrder(tx, userID, cart, &order); err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Transaction failed")
		return
	}

	productNames := make([]string, 0, len(cart))
	for _, item := range cart {
		productNames = append(productNames, item.Product.Name)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	session, err := services.CreateSSLCommerzSession(ctx, services.SSLCommerzSessionRequest{
		TransactionID: transactionID,
		Amount:        order.TotalPrice,
		Currency:      "BDT",
		ProductName:   buildProductSummary(productNames),
		ProductNames:  productNames,
		Category:      "general",
		CustomerName:  req.CustomerName,
		CustomerEmail: user.Email,
		Phone:         req.Phone,
		AddressLine:   req.AddressLine,
		City:          req.City,
		Area:          req.Area,
		PostalCode:    req.PostalCode,
	})
	if err != nil {
		restoreOnlineOrder(order.ID, userID, "Failed", "Gateway Error")
		utils.RespondError(c, http.StatusBadGateway, "Failed to initialize SSLCommerz payment session")
		return
	}

	if err := config.DB.Model(&models.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
		"session_key": session.SessionKey,
	}).Error; err != nil {
		restoreOnlineOrder(order.ID, userID, "Failed", "Gateway Error")
		utils.RespondError(c, http.StatusInternalServerError, "Failed to persist payment session")
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "SSLCommerz session created", gin.H{
		"order_id":       order.ID,
		"transaction_id": transactionID,
		"payment_method": req.PaymentMethod,
		"status":         order.Status,
		"gateway_url":    session.GatewayPageURL,
	})
}

func GetMyOrders(c *gin.Context) {
	userID := c.GetUint("user_id")

	var orders []models.Order
	if err := config.DB.Where("user_id = ? AND archived = ?", userID, false).Order("created_at DESC").Find(&orders).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	response := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, dto.ToOrderResponse(order))
	}

	utils.RespondSuccess(c, http.StatusOK, "Orders loaded", gin.H{
		"orders": response,
	})
}

func ArchiveOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var order models.Order
	if err := config.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "Order not found")
		return
	}

	if order.Archived {
		utils.RespondSuccess(c, http.StatusOK, "Order already archived", nil)
		return
	}

	if err := config.DB.Model(&order).Update("archived", true).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to archive order")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Order archived", gin.H{
		"id": order.ID,
	})
}

func GetAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 8
	}

	statusFilter := strings.ToLower(strings.TrimSpace(c.DefaultQuery("archived", "active")))
	offset := (page - 1) * limit

	var orders []models.Order
	query := config.DB.Model(&models.Order{})

	switch statusFilter {
	case "archived":
		query = query.Where("archived = ?", true)
	case "all":
	default:
		query = query.Where("archived = ?", false)
		statusFilter = "active"
	}

	if err := query.
		Preload("User").
		Preload("Items.Product").
		Order("created_at DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&orders).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	hasNext := len(orders) > limit
	if hasNext {
		orders = orders[:limit]
	}

	response := make([]dto.AdminOrderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, dto.ToAdminOrderResponse(order))
	}

	utils.RespondSuccess(c, http.StatusOK, "Orders loaded", gin.H{
		"orders":   response,
		"page":     page,
		"limit":    limit,
		"has_next": hasNext,
		"has_prev": page > 1,
		"archived": statusFilter,
	})
}

func SSLCommerzSuccess(c *gin.Context) {
	redirectSSLCommerzResult(c, "success")
}

func SSLCommerzFail(c *gin.Context) {
	redirectSSLCommerzResult(c, "fail")
}

func SSLCommerzCancel(c *gin.Context) {
	redirectSSLCommerzResult(c, "cancel")
}

func SSLCommerzIPN(c *gin.Context) {
	tranID := readGatewayField(c, "tran_id")
	valID := readGatewayField(c, "val_id")
	if tranID == "" || valID == "" {
		c.String(http.StatusBadRequest, "missing payment payload")
		return
	}

	if _, err := confirmSSLCommerzPayment(tranID, valID); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.String(http.StatusOK, "payment validated")
}

func redirectSSLCommerzResult(c *gin.Context, result string) {
	tranID := readGatewayField(c, "tran_id")
	valID := readGatewayField(c, "val_id")

	switch result {
	case "success":
		if tranID == "" || valID == "" {
			redirectToPaymentResult(c, "failed", 0, "", "Missing payment confirmation details.")
			return
		}

		orderID, err := confirmSSLCommerzPayment(tranID, valID)
		if err != nil {
			redirectToPaymentResult(c, "failed", 0, tranID, err.Error())
			return
		}

		redirectToPaymentResult(c, "success", orderID, tranID, "")
	case "fail":
		orderID := failSSLCommerzPayment(tranID, "Failed", "Payment Failed")
		redirectToPaymentResult(c, "failed", orderID, tranID, "Your payment was not completed.")
	case "cancel":
		orderID := failSSLCommerzPayment(tranID, "Cancelled", "Payment Cancelled")
		redirectToPaymentResult(c, "cancelled", orderID, tranID, "You cancelled the payment.")
	default:
		redirectToPaymentResult(c, "failed", 0, tranID, "Unknown payment state.")
	}
}

func confirmSSLCommerzPayment(tranID string, valID string) (uint, error) {
	var order models.Order
	if err := config.DB.Where("transaction_id = ?", tranID).First(&order).Error; err != nil {
		return 0, fmt.Errorf("order not found for transaction")
	}

	if order.PaymentStatus == "Paid" {
		return order.ID, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	validation, err := services.ValidateSSLCommerzPayment(ctx, valID)
	if err != nil {
		return 0, fmt.Errorf("payment validation failed")
	}

	status := strings.ToUpper(strings.TrimSpace(validation.Status))
	if status != "VALID" && status != "VALIDATED" {
		return 0, fmt.Errorf("payment status is not valid")
	}

	if validation.TransactionID != order.TransactionID {
		return 0, fmt.Errorf("transaction mismatch detected")
	}

	amount, err := strconv.ParseFloat(validation.Amount, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid gateway amount")
	}

	if math.Abs(amount-order.TotalPrice) > 0.01 {
		return 0, fmt.Errorf("payment amount mismatch detected")
	}

	currency := strings.ToUpper(strings.TrimSpace(validation.Currency))
	if currency == "" {
		currency = "BDT"
	}
	if order.Currency != "" && !strings.EqualFold(order.Currency, currency) {
		return 0, fmt.Errorf("payment currency mismatch detected")
	}

	storeAmount, _ := strconv.ParseFloat(validation.StoreAmount, 64)
	updates := map[string]any{
		"payment_status": "Paid",
		"status":         "Confirmed",
		"validation_id":  validation.ValidationID,
		"bank_tran_id":   validation.BankTranID,
		"currency":       currency,
		"gateway_amount": storeAmount,
		"card_type":      validation.CardType,
	}

	if validation.SessionKey != "" {
		updates["session_key"] = validation.SessionKey
	}

	if err := config.DB.Model(&models.Order{}).Where("id = ?", order.ID).Updates(updates).Error; err != nil {
		return 0, fmt.Errorf("failed to update paid order")
	}

	return order.ID, nil
}

func failSSLCommerzPayment(tranID string, paymentStatus string, orderStatus string) uint {
	if tranID == "" {
		return 0
	}

	var order models.Order
	if err := config.DB.Where("transaction_id = ?", tranID).First(&order).Error; err != nil {
		return 0
	}

	if order.PaymentStatus == "Paid" {
		return order.ID
	}

	restoreOnlineOrder(order.ID, order.UserID, paymentStatus, orderStatus)
	return order.ID
}

func createOrderFromCart(tx *gorm.DB, userID uint, req dto.CreateOrderRequest, seed models.Order) ([]models.Cart, models.Order, error) {
	var cart []models.Cart
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Product").
		Where("user_id = ?", userID).
		Find(&cart).Error; err != nil {
		return nil, models.Order{}, fmt.Errorf("failed to load cart")
	}
	if len(cart) == 0 {
		return nil, models.Order{}, fmt.Errorf("cart is empty")
	}

	order := seed
	order.UserID = userID
	order.CustomerName = req.CustomerName
	order.Phone = req.Phone
	order.AddressLine = req.AddressLine
	order.City = req.City
	order.Area = req.Area
	order.PostalCode = req.PostalCode
	order.Notes = req.Notes

	if err := tx.Create(&order).Error; err != nil {
		return nil, models.Order{}, fmt.Errorf("failed to create order")
	}

	return cart, order, nil
}

func finalizeReservedOrder(tx *gorm.DB, userID uint, cart []models.Cart, order *models.Order) error {
	total := 0.0

	for _, item := range cart {
		var product models.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, item.ProductID).Error; err != nil {
			return fmt.Errorf("product not found")
		}

		if product.Stock < item.Quantity {
			return fmt.Errorf("not enough stock for %s", product.Name)
		}

		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Price:     float64(item.Quantity) * product.Price,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			return fmt.Errorf("failed to create order item")
		}

		if err := tx.Model(&product).Update("stock", product.Stock-item.Quantity).Error; err != nil {
			return fmt.Errorf("failed to update product stock")
		}

		total += orderItem.Price
	}

	if err := tx.Model(order).Updates(map[string]any{
		"total_price": total,
	}).Error; err != nil {
		return fmt.Errorf("failed to update order total")
	}

	if err := tx.Where("user_id = ?", userID).Delete(&models.Cart{}).Error; err != nil {
		return fmt.Errorf("failed to clear cart")
	}

	order.TotalPrice = total
	return nil
}

func restoreOnlineOrder(orderID uint, userID uint, paymentStatus string, orderStatus string) {
	tx := config.DB.Begin()
	if tx.Error != nil {
		return
	}

	var order models.Order
	if err := tx.Preload("Items").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		return
	}

	if order.PaymentStatus == "Paid" {
		tx.Rollback()
		return
	}

	for _, item := range order.Items {
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			return
		}

		var cart models.Cart
		err := tx.Where("user_id = ? AND product_id = ?", userID, item.ProductID).First(&cart).Error
		switch {
		case err == nil:
			if err := tx.Model(&cart).Update("quantity", cart.Quantity+item.Quantity).Error; err != nil {
				tx.Rollback()
				return
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			cart = models.Cart{
				UserID:    userID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
			if err := tx.Create(&cart).Error; err != nil {
				tx.Rollback()
				return
			}
		default:
			tx.Rollback()
			return
		}
	}

	if err := tx.Model(&order).Updates(map[string]any{
		"payment_status": paymentStatus,
		"status":         orderStatus,
	}).Error; err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()
}

func redirectToPaymentResult(c *gin.Context, status string, orderID uint, tranID string, message string) {
	params := url.Values{}
	params.Set("status", status)
	if orderID > 0 {
		params.Set("order_id", strconv.FormatUint(uint64(orderID), 10))
	}
	if tranID != "" {
		params.Set("transaction_id", tranID)
	}
	if message != "" {
		params.Set("message", message)
	}

	target := config.SSLCommerz.FrontendURL + "/payment/result?" + params.Encode()
	c.Redirect(http.StatusSeeOther, target)
}

func readGatewayField(c *gin.Context, key string) string {
	value := strings.TrimSpace(c.PostForm(key))
	if value != "" {
		return value
	}
	return strings.TrimSpace(c.Query(key))
}

func normalizeOrderRequest(req *dto.CreateOrderRequest) {
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.AddressLine = strings.TrimSpace(req.AddressLine)
	req.City = strings.TrimSpace(req.City)
	req.Area = strings.TrimSpace(req.Area)
	req.PostalCode = strings.TrimSpace(req.PostalCode)
	req.Notes = strings.TrimSpace(req.Notes)
	req.PaymentMethod = strings.ToLower(strings.TrimSpace(req.PaymentMethod))
}

func validateOrderRequest(req dto.CreateOrderRequest) error {
	if req.CustomerName == "" || req.Phone == "" || req.AddressLine == "" || req.City == "" || req.Area == "" {
		return fmt.Errorf("please complete the delivery information")
	}
	return nil
}

func buildTransactionID(userID uint) string {
	return fmt.Sprintf("ECOM-%d-%d", userID, time.Now().UnixNano())
}

func buildProductSummary(productNames []string) string {
	if len(productNames) == 0 {
		return "Order Payment"
	}
	if len(productNames) == 1 {
		return productNames[0]
	}
	return fmt.Sprintf("%s and %d more", productNames[0], len(productNames)-1)
}
