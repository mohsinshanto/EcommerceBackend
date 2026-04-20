package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"ecommerce-backend/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req dto.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	req.CustomerName = strings.TrimSpace(req.CustomerName)
	req.Phone = strings.TrimSpace(req.Phone)
	req.AddressLine = strings.TrimSpace(req.AddressLine)
	req.City = strings.TrimSpace(req.City)
	req.Area = strings.TrimSpace(req.Area)
	req.PostalCode = strings.TrimSpace(req.PostalCode)
	req.Notes = strings.TrimSpace(req.Notes)

	if req.CustomerName == "" || req.Phone == "" || req.AddressLine == "" || req.City == "" || req.Area == "" {
		utils.RespondError(c, http.StatusBadRequest, "Please complete the delivery information")
		return
	}

	paymentMethod := "Cash on Delivery"

	// Start DB transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	// 1. Load user's cart with product info
	var cart []models.Cart
	if err := tx.Preload("Product").Where("user_id = ?", userID).Find(&cart).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Failed to load cart")
		return
	}
	if len(cart) == 0 {
		tx.Rollback()
		utils.RespondError(c, http.StatusBadRequest, "Cart is empty")
		return
	}

	// 2. Create initial order (total price = 0 for now)
	order := models.Order{
		UserID:        userID,
		CustomerName:  req.CustomerName,
		Phone:         req.Phone,
		AddressLine:   req.AddressLine,
		City:          req.City,
		Area:          req.Area,
		PostalCode:    req.PostalCode,
		Notes:         req.Notes,
		PaymentMethod: paymentMethod,
		Status:        "Pending",
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create order")
		return
	}

	total := 0.0

	// 3. Loop through each cart item
	for _, item := range cart {
		product := item.Product

		// Stock validation
		if product.Stock < item.Quantity {
			tx.Rollback()
			utils.RespondError(c, http.StatusBadRequest, "Not enough stock for "+product.Name)
			return
		}

		// Create order item
		oi := models.OrderItem{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Price:     float64(item.Quantity) * product.Price,
		}

		if err := tx.Create(&oi).Error; err != nil {
			tx.Rollback()
			utils.RespondError(c, http.StatusInternalServerError, "Failed to create order item")
			return
		}

		// Reduce stock
		newStock := product.Stock - item.Quantity
		if err := tx.Model(&product).Update("stock", newStock).Error; err != nil {
			tx.Rollback()
			utils.RespondError(c, http.StatusInternalServerError, "Failed to update product stock")
			return
		}

		total += oi.Price
	}

	// 4. Update order total
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"total_price": total,
	}).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update order total")
		return
	}

	// 5. Clear cart
	if err := tx.Where("user_id = ?", userID).Delete(&models.Cart{}).Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Failed to clear cart")
		return
	}

	// 6. Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.RespondError(c, http.StatusInternalServerError, "Transaction failed")
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Order placed successfully", gin.H{
		"order_id":  order.ID,
		"totalPaid": total,
		"status":    order.Status,
	})
}

func GetMyOrders(c *gin.Context) {
	userID := c.GetUint("user_id")

	var orders []models.Order
	if err := config.DB.Where("user_id = ? AND archived = ?", userID, false).Order("created_at DESC").Find(&orders).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	var response []dto.OrderResponse
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

// Admin-only: Get all orders
func GetAllOrders(c *gin.Context) {
	var orders []models.Order

	if err := config.DB.Find(&orders).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Orders loaded", gin.H{
		"orders": orders,
	})
}
