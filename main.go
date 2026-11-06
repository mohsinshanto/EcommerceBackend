package main

import (
	"context"
	"ecommerce-backend/config"
	"ecommerce-backend/internal/module/cart"
	"ecommerce-backend/internal/module/category"
	"ecommerce-backend/internal/module/order"
	"ecommerce-backend/internal/module/product"
	"ecommerce-backend/internal/module/user"
	"ecommerce-backend/utils"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	config.LoadJWTSecret()
	config.InitSSLCommerz()
	config.InitRedis()
	config.ConnectDB()
	config.InitTypesense()
	config.DB.AutoMigrate(
		&user.User{},
		&category.Category{},
		&product.Product{},
		&cart.Cart{},
		&order.Order{},
		&order.OrderItem{},
	)

	if err := product.EnsureTypesenseProductsCollection(context.Background()); err != nil {
		log.Printf("WARNING: failed to ensure Typesense products collection: %v", err)
	}
	if os.Getenv("SYNC_TYPESENSE") == "true" {
		log.Println("Starting Typesense product re-sync...")

		if err := product.SyncAllProductsToTypesense(
			context.Background(),
			500,
		); err != nil {
			log.Fatalf("Typesense sync failed: %v", err)
		}

		log.Println("Typesense product re-sync completed successfully")
		return
	}
	frontendOrigins := []string{"http://localhost:3000"}
	if configuredOrigins := os.Getenv("FRONTEND_ORIGINS"); configuredOrigins != "" {
		frontendOrigins = splitAndTrim(configuredOrigins)
	}
	userModule := user.NewModule(config.DB)
	categoryModule := category.NewModule(config.DB)
	categoryRepo := category.NewCategoryRepository(config.DB)
	categoryService := category.NewCategoryService(categoryRepo)
	productModule := product.NewModule(config.DB, categoryService)
	cartModule := cart.NewModule(config.DB)
	orderModule := order.NewModule(config.DB)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     frontendOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			//log.Printf("CORS origin: %q", origin)

			return origin == "http://localhost:3000" ||
				origin == "http://192.168.0.109:3000" ||
				origin == "null"
		},
	}))

	r.NoRoute(func(c *gin.Context) {
		utils.RespondError(c, http.StatusNotFound, "Route not found")
	})

	r.NoMethod(func(c *gin.Context) {
		utils.RespondError(c, http.StatusMethodNotAllowed, "Method not allowed")
	})
	userModule.RegisterRoutes(r)
	productModule.RegisterRoutes(r)
	categoryModule.RegisterRoutes(r)
	cartModule.RegisterRoutes(r)
	orderModule.RegisterRoutes(r)
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}

	return origins
}
