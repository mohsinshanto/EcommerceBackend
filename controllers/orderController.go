package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Start DB transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// 1. Load user's cart with product info
	var cart []models.Cart
	if err := tx.Preload("Product").Where("user_id = ?", userID).Find(&cart).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
		return
	}
	fmt.Println(cart)
	if len(cart) == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart empty"})
		return
	}

	// 2. Create initial order (total price = 0 for now)
	order := models.Order{UserID: userID}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	total := 0.0

	// 3. Loop through each cart item
	for _, item := range cart {
		product := item.Product

		// Stock validation
		if product.Stock < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Not enough stock for product %s", product.Name),
			})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order item"})
			return
		}

		// Reduce stock
		newStock := product.Stock - item.Quantity
		if err := tx.Model(&product).Update("stock", newStock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}

		total += oi.Price
	}

	// 4. Update order total
	if err := tx.Model(&order).Update("total_price", total).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order total"})
		return
	}

	// 5. Clear cart
	if err := tx.Where("user_id = ?", userID).Delete(&models.Cart{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	// 6. Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Order placed successfully",
		"order_id":  order.ID,
		"totalPaid": total,
	})
}

func GetMyOrders(c *gin.Context) {
	userID := c.GetUint("user_id")

	var orders []models.Order
	config.DB.Where("user_id = ?", userID).Find(&orders)

	c.JSON(200, orders)
}

// Admin-only: Get all orders
func GetAllOrders(c *gin.Context) {
	var orders []models.Order

	if err := config.DB.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch orders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
	})
}
