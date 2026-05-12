package controllers

import (
	"ecommerce-backend/dto"
	"ecommerce-backend/services"
	"ecommerce-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))

	params := services.GetProductsParams{
		Search:   c.Query("search"),
		Category: c.Query("category"),
		Sort:     c.Query("sort"),
		MinPrice: c.Query("min_price"),
		MaxPrice: c.Query("max_price"),
		Page:     page,
		Limit:    limit,
	}

	products, count, hasNext, err := services.GetProducts(params)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	categoryCounts := services.GetCategoryCounts(params)

	utils.RespondSuccess(c, http.StatusOK, "Products loaded", gin.H{
		"products":        products,
		"count":           count,
		"category_counts": categoryCounts,
		"page":            page,
		"limit":           limit,
		"has_next":        hasNext,
		"has_prev":        page > 1,
	})
}

func GetProductCount(c *gin.Context) {
	count, err := services.GetProductCount()
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to load product count")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Product count loaded", gin.H{
		"count": count,
	})
}

func CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	product, err := services.CreateProduct(req)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create product")
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, "Product created successfully", product)
}
func DeleteProduct(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	err = services.DeleteProduct(id)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Product deleted successfully", nil)
}
func GetProductByID(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product, err := services.GetProductByID(id)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, "Product not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Product loaded", product)
}
