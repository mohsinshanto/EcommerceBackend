package user

import (
	"context"
	"ecommerce-backend/config"
	appError "ecommerce-backend/errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(parent context.Context, req RegisterRequest) (*RegisterResponse, error)
	Login(parent context.Context, req LoginRequest) (*LoginResponse, error)
}

// UserService handles user business logic
type userService struct {
	repo UserRepository
}

// NewUserService creates a new user service instance
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(parent context.Context, req RegisterRequest) (*RegisterResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	// Validation
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	// Check if email already exists
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, appError.ErrEmailAlreadyRegistered
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsAdmin:  false,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		Message: "User registered successfully",
		User: UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

func (s *userService) Login(parent context.Context, req LoginRequest) (*LoginResponse, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, appError.ErrInvalidCredentials
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(config.JwtSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: tokenString}, nil
}
