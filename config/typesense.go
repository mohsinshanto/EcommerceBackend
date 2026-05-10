package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type TypesenseClient struct {
	baseURL    string
	apiKey     string
	collection string
	httpClient *http.Client
}

var Typesense *TypesenseClient

func InitTypesense() {
	apiKey := strings.TrimSpace(os.Getenv("TYPESENSE_API_KEY"))
	if apiKey == "" {
		log.Println("Typesense disabled: TYPESENSE_API_KEY is not set")
		return
	}

	protocol := strings.TrimSpace(os.Getenv("TYPESENSE_PROTOCOL"))
	if protocol == "" {
		protocol = "http"
	}

	host := strings.TrimSpace(os.Getenv("TYPESENSE_HOST"))
	if host == "" {
		host = "localhost"
	}

	port := strings.TrimSpace(os.Getenv("TYPESENSE_PORT"))
	if port == "" {
		port = "8108"
	}

	collection := strings.TrimSpace(os.Getenv("TYPESENSE_PRODUCTS_COLLECTION"))
	if collection == "" {
		collection = "products"
	}

	Typesense = &TypesenseClient{
		baseURL:    fmt.Sprintf("%s://%s:%s", protocol, host, port),
		apiKey:     apiKey,
		collection: collection,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	log.Printf("Typesense enabled at %s for collection %s", Typesense.baseURL, Typesense.collection)
}

func (c *TypesenseClient) Collection() string {
	return c.collection
}

func (c *TypesenseClient) Request(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	body any,
) ([]byte, int, error) {
	if c == nil {
		return nil, 0, fmt.Errorf("typesense client is not initialized")
	}

	fullURL := c.baseURL + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var payload io.Reader
	if body != nil {
		switch typed := body.(type) {
		case []byte:
			payload = bytes.NewReader(typed)
		case string:
			payload = strings.NewReader(typed)
		default:
			data, err := json.Marshal(body)
			if err != nil {
				return nil, 0, err
			}
			payload = bytes.NewReader(data)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, payload)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode >= 400 {
		return responseBody, resp.StatusCode, fmt.Errorf("typesense returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return responseBody, resp.StatusCode, nil
}
