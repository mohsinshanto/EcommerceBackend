package product

import (
	"ecommerce-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	service ProductService
}

func NewProductController(service ProductService) *ProductController {
	return &ProductController{service: service}
}

func (c *ProductController) GetProducts(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "8"))

	params := GetProductsParams{
		Search:   ctx.Query("search"),
		Category: ctx.Query("category"),
		Sort:     ctx.Query("sort"),
		MinPrice: ctx.Query("min_price"),
		MaxPrice: ctx.Query("max_price"),
		Page:     page,
		Limit:    limit,
	}

	products, count, searchCount, hasNext, err := c.service.GetProducts(ctx.Request.Context(), params)
	if err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	categoryCounts := c.service.GetCategoryCounts(ctx.Request.Context(), params)
	utils.RespondSuccess(ctx, http.StatusOK, "Products loaded", gin.H{
		"products":        products,
		"count":           count,
		"search_count":    searchCount,
		"category_counts": categoryCounts,
		"page":            page,
		"limit":           limit,
		"has_next":        hasNext,
		"has_prev":        page > 1,
	})
}

func (c *ProductController) GetProductCount(ctx *gin.Context) {
	count, err := c.service.GetProductCount(ctx.Request.Context())
	if err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, "Failed to load product count")
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Product count loaded", gin.H{"count": count})
}

func (c *ProductController) GetProductByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product, err := c.service.GetProductByID(ctx.Request.Context(), id)
	if err != nil {
		utils.RespondError(ctx, http.StatusNotFound, "Product not found")
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Product loaded", product)
}

func (c *ProductController) CreateProduct(ctx *gin.Context) {
	var req CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	product, err := c.service.CreateProduct(ctx.Request.Context(), req)
	if err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, "Failed to create product")
		return
	}

	utils.RespondSuccess(ctx, http.StatusCreated, "Product created successfully", product)
}

func (c *ProductController) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if err := c.service.DeleteProduct(ctx.Request.Context(), id); err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Product deleted successfully", nil)
}
