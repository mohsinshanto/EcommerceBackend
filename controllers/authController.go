package controllers

import (
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"ecommerce-backend/utils"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" {
		utils.RespondError(c, http.StatusBadRequest, "name is required")
		return
	}

	var existingUser models.User
	err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		utils.RespondError(c, http.StatusBadRequest, "Email already registered")
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to validate email address")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to secure password")
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsAdmin:  false, // force safety
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	res := dto.RegisterResponse{
		Message: "User registered successfully",
		User: dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}

	utils.RespondSuccess(c, http.StatusOK, "User registered successfully", res)
}

func Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBindError(c, err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	err := config.DB.Where("email = ?", req.Email).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.RespondError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to process login")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // 3 days expiry
	})

	tokenString, err := token.SignedString(config.JwtSecret)

	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, "Login successful", gin.H{
		"token": tokenString,
	})
}
