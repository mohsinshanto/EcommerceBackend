package user

import (
	appError "ecommerce-backend/errors"
	"ecommerce-backend/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service UserService
}

func NewUserController(service UserService) *UserController {
	return &UserController{service: service}
}

func (c *UserController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.Register(ctx.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, appError.ErrEmailAlreadyRegistered):
			utils.RespondError(ctx, http.StatusConflict, err.Error())

		case errors.Is(err, appError.ErrInvalidCredentials):
			utils.RespondError(ctx, http.StatusUnauthorized, err.Error())

		default:
			utils.RespondError(ctx, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "User registered successfully", result)
}

func (c *UserController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.Login(ctx.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, appError.ErrInvalidCredentials):
			utils.RespondError(ctx, http.StatusUnauthorized, err.Error())
		default:
			utils.RespondError(ctx, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	utils.RespondSuccess(ctx, http.StatusOK, "Login successful", result)
}
