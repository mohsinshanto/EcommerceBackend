package category

import (
	"ecommerce-backend/utils"
	"net/http"

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
