package category

import (
	appError "ecommerce-backend/errors"
	"ecommerce-backend/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	service CategoryService
}

func NewCategoryController(service CategoryService) *CategoryController {
	return &CategoryController{
		service: service,
	}
}
func (c *CategoryController) GetCategoryTree(ctx *gin.Context) {
	categories, err := c.service.GetCategoryTree(ctx.Request.Context())
	if err != nil {
		utils.RespondError(
			ctx,
			http.StatusInternalServerError,
			"Failed to load categories",
		)
		return
	}

	utils.RespondSuccess(
		ctx,
		http.StatusOK,
		"Categories loaded successfully",
		categories,
	)
}
func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	var req CreateCategoryRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	category := &Category{
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	if err := c.service.CreateCategory(ctx.Request.Context(), category); err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(
		ctx,
		http.StatusCreated,
		"Category created successfully",
		category,
	)
}
func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	// 1. Get category ID from URL.
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// 2. Bind request body.
	var req UpdateCategoryRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Call service.
	if err := c.service.UpdateCategory(ctx.Request.Context(), uint(id), req); err != nil {

		switch {
		case errors.Is(err, appError.ErrCategoryNotFound):
			utils.RespondError(ctx, http.StatusNotFound, err.Error())

		case errors.Is(err, appError.ErrCircularCategoryReference):
			utils.RespondError(ctx, http.StatusBadRequest, err.Error())

		case errors.Is(err, appError.ErrExistingCircularReference):
			utils.RespondError(ctx, http.StatusInternalServerError, err.Error())

		default:
			utils.RespondError(ctx, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// 4. Return success response.
	utils.RespondSuccess(ctx, http.StatusOK, "Category updated successfully", nil)
}
