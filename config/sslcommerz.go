package config

import (
	"log"
	"os"
	"strings"
)

type SSLCommerzConfig struct {
	StoreID        string
	StorePassword  string
	BaseURL        string
	FrontendURL    string
	BackendBaseURL string
}

var SSLCommerz *SSLCommerzConfig

func InitSSLCommerz() {
	storeID := strings.TrimSpace(os.Getenv("SSLCOMMERZ_STORE_ID"))
	storePassword := strings.TrimSpace(os.Getenv("SSLCOMMERZ_STORE_PASSWORD"))
	if storeID == "" || storePassword == "" {
		storeID = "testbox"
		storePassword = "qwerty"
		log.Println("WARNING: SSLCOMMERZ_STORE_ID or SSLCOMMERZ_STORE_PASSWORD is not set. Falling back to sandbox credentials.")
	}

	baseURL := strings.TrimSpace(os.Getenv("SSLCOMMERZ_BASE_URL"))
	if baseURL == "" {
		baseURL = "https://sandbox.sslcommerz.com"
	}

	frontendURL := strings.TrimSpace(os.Getenv("FRONTEND_APP_URL"))
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	backendBaseURL := strings.TrimSpace(os.Getenv("BACKEND_PUBLIC_URL"))
	if backendBaseURL == "" {
		backendBaseURL = "http://localhost:8080"
	}

	SSLCommerz = &SSLCommerzConfig{
		StoreID:        storeID,
		StorePassword:  storePassword,
		BaseURL:        strings.TrimRight(baseURL, "/"),
		FrontendURL:    strings.TrimRight(frontendURL, "/"),
		BackendBaseURL: strings.TrimRight(backendBaseURL, "/"),
	}

	log.Printf("SSLCOMMERZ enabled with base URL %s", SSLCommerz.BaseURL)
}
