package services

import (
	"context"
	"ecommerce-backend/config"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SSLCommerzSessionRequest struct {
	TransactionID string
	Amount        float64
	Currency      string
	ProductName   string
	ProductNames  []string
	Category      string
	CustomerName  string
	CustomerEmail string
	Phone         string
	AddressLine   string
	City          string
	Area          string
	PostalCode    string
}

type SSLCommerzSessionResponse struct {
	Status         string `json:"status"`
	FailedReason   string `json:"failedreason"`
	SessionKey     string `json:"sessionkey"`
	GatewayPageURL string `json:"GatewayPageURL"`
	StoreLogo      string `json:"storeLogo"`
}

type SSLCommerzValidationResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"tran_id"`
	ValidationID  string `json:"val_id"`
	Amount        string `json:"amount"`
	StoreAmount   string `json:"store_amount"`
	Currency      string `json:"currency"`
	BankTranID    string `json:"bank_tran_id"`
	CardType      string `json:"card_type"`
	SessionKey    string `json:"sessionkey"`
	RiskLevel     string `json:"risk_level"`
	RiskTitle     string `json:"risk_title"`
}

func SSLCommerzEnabled() bool {
	return config.SSLCommerz != nil
}

func CreateSSLCommerzSession(ctx context.Context, req SSLCommerzSessionRequest) (SSLCommerzSessionResponse, error) {
	if config.SSLCommerz == nil {
		return SSLCommerzSessionResponse{}, fmt.Errorf("sslcommerz is not configured")
	}

	form := url.Values{}
	form.Set("store_id", config.SSLCommerz.StoreID)
	form.Set("store_passwd", config.SSLCommerz.StorePassword)
	form.Set("total_amount", strconv.FormatFloat(req.Amount, 'f', 2, 64))
	form.Set("currency", nonEmpty(req.Currency, "BDT"))
	form.Set("tran_id", req.TransactionID)
	form.Set("success_url", config.SSLCommerz.BackendBaseURL+"/api/payments/sslcommerz/success")
	form.Set("fail_url", config.SSLCommerz.BackendBaseURL+"/api/payments/sslcommerz/fail")
	form.Set("cancel_url", config.SSLCommerz.BackendBaseURL+"/api/payments/sslcommerz/cancel")
	form.Set("ipn_url", config.SSLCommerz.BackendBaseURL+"/api/payments/sslcommerz/ipn")
	form.Set("shipping_method", "NO")
	form.Set("product_name", nonEmpty(req.ProductName, "Order Payment"))
	form.Set("product_category", nonEmpty(req.Category, "general"))
	form.Set("product_profile", "physical-goods")
	form.Set("cus_name", req.CustomerName)
	form.Set("cus_email", nonEmpty(req.CustomerEmail, "customer@example.com"))
	form.Set("cus_add1", req.AddressLine)
	form.Set("cus_city", req.City)
	form.Set("cus_state", req.Area)
	form.Set("cus_postcode", nonEmpty(req.PostalCode, "1200"))
	form.Set("cus_country", "Bangladesh")
	form.Set("cus_phone", req.Phone)
	form.Set("value_a", req.TransactionID)
	form.Set("value_b", strings.Join(req.ProductNames, ", "))

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		config.SSLCommerz.BaseURL+"/gwprocess/v4/api.php",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return SSLCommerzSessionResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return SSLCommerzSessionResponse{}, err
	}
	defer resp.Body.Close()

	var response SSLCommerzSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return SSLCommerzSessionResponse{}, err
	}

	if response.GatewayPageURL == "" {
		reason := strings.TrimSpace(response.FailedReason)
		if reason == "" {
			reason = "gateway session could not be created"
		}
		return SSLCommerzSessionResponse{}, fmt.Errorf(reason)
	}

	return response, nil
}

func ValidateSSLCommerzPayment(ctx context.Context, validationID string) (SSLCommerzValidationResponse, error) {
	if config.SSLCommerz == nil {
		return SSLCommerzValidationResponse{}, fmt.Errorf("sslcommerz is not configured")
	}

	query := url.Values{}
	query.Set("val_id", validationID)
	query.Set("store_id", config.SSLCommerz.StoreID)
	query.Set("store_passwd", config.SSLCommerz.StorePassword)
	query.Set("format", "json")

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		config.SSLCommerz.BaseURL+"/validator/api/validationserverAPI.php?"+query.Encode(),
		nil,
	)
	if err != nil {
		return SSLCommerzValidationResponse{}, err
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return SSLCommerzValidationResponse{}, err
	}
	defer resp.Body.Close()

	var response SSLCommerzValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return SSLCommerzValidationResponse{}, err
	}

	return response, nil
}

func nonEmpty(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
