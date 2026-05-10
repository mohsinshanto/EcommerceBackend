package services

import (
	"bytes"
	"context"
	"ecommerce-backend/config"
	"ecommerce-backend/dto"
	"ecommerce-backend/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type typesenseCollectionSchema struct {
	Name                string                 `json:"name"`
	Fields              []typesenseSchemaField `json:"fields"`
	DefaultSortingField string                 `json:"default_sorting_field,omitempty"`
}

type typesenseSchemaField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Facet    bool   `json:"facet,omitempty"`
	Optional bool   `json:"optional,omitempty"`
	Index    *bool  `json:"index,omitempty"`
	Sort     bool   `json:"sort,omitempty"`
}

type typesenseProductDocument struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url,omitempty"`
	CreatedAt   int64   `json:"created_at"`
}

type typesenseSearchResponse struct {
	Found int `json:"found"`
	Hits  []struct {
		Document typesenseProductDocument `json:"document"`
	} `json:"hits"`
}

func EnsureTypesenseProductsCollection(ctx context.Context) error {
	if config.Typesense == nil {
		return nil
	}

	path := "/collections/" + config.Typesense.Collection()
	_, status, err := config.Typesense.Request(ctx, http.MethodGet, path, nil, nil)
	if err == nil {
		return nil
	}
	if status != http.StatusNotFound {
		return err
	}

	falseValue := false
	schema := typesenseCollectionSchema{
		Name: config.Typesense.Collection(),
		Fields: []typesenseSchemaField{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "category", Type: "string", Facet: true},
			{Name: "price", Type: "float", Sort: true},
			{Name: "stock", Type: "int32", Sort: true},
			{Name: "created_at", Type: "int64", Sort: true},
			{Name: "description", Type: "string", Optional: true, Index: &falseValue},
			{Name: "image_url", Type: "string", Optional: true, Index: &falseValue},
		},
		DefaultSortingField: "created_at",
	}

	_, _, err = config.Typesense.Request(ctx, http.MethodPost, "/collections", nil, schema)
	return err
}

func UpsertProductInTypesense(ctx context.Context, product models.Product) error {
	if config.Typesense == nil {
		return nil
	}
	if err := EnsureTypesenseProductsCollection(ctx); err != nil {
		return err
	}

	document := typesenseProductDocumentFromModel(product)
	query := url.Values{}
	query.Set("action", "upsert")

	_, _, err := config.Typesense.Request(
		ctx,
		http.MethodPost,
		"/collections/"+config.Typesense.Collection()+"/documents",
		query,
		document,
	)
	return err
}

func DeleteProductFromTypesense(ctx context.Context, id uint) error {
	if config.Typesense == nil {
		return nil
	}

	_, status, err := config.Typesense.Request(
		ctx,
		http.MethodDelete,
		"/collections/"+config.Typesense.Collection()+"/documents/"+strconv.FormatUint(uint64(id), 10),
		nil,
		nil,
	)
	if status == http.StatusNotFound {
		return nil
	}
	return err
}

func SearchProductsWithTypesense(ctx context.Context, params GetProductsParams) ([]dto.ProductResponse, bool, error) {
	if config.Typesense == nil {
		return nil, false, fmt.Errorf("typesense is not enabled")
	}

	if err := EnsureTypesenseProductsCollection(ctx); err != nil {
		return nil, false, err
	}

	query := url.Values{}
	query.Set("q", strings.TrimSpace(params.Search))
	query.Set("query_by", "name")
	query.Set("prefix", "false")
	query.Set("page", strconv.Itoa(max(params.Page, 1)))
	query.Set("per_page", strconv.Itoa(max(params.Limit+1, 1)))

	filterBy := buildTypesenseFilter(params)
	if filterBy != "" {
		query.Set("filter_by", filterBy)
	}

	switch params.Sort {
	case "price_asc":
		query.Set("sort_by", "price:asc")
	case "price_desc":
		query.Set("sort_by", "price:desc")
	default:
		query.Set("sort_by", "_text_match:desc,created_at:desc")
	}

	body, _, err := config.Typesense.Request(
		ctx,
		http.MethodGet,
		"/collections/"+config.Typesense.Collection()+"/documents/search",
		query,
		nil,
	)
	if err != nil {
		return nil, false, err
	}

	var response typesenseSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, false, err
	}

	products := make([]dto.ProductResponse, 0, len(response.Hits))
	for _, hit := range response.Hits {
		products = append(products, dto.ProductResponse{
			ID:          parseUintDocumentID(hit.Document.ID),
			Name:        hit.Document.Name,
			Description: hit.Document.Description,
			Price:       hit.Document.Price,
			Stock:       hit.Document.Stock,
			ImageURL:    hit.Document.ImageURL,
			Category:    hit.Document.Category,
		})
	}

	hasNext := len(products) > params.Limit
	if hasNext {
		products = products[:params.Limit]
	}

	return products, hasNext, nil
}

func typesenseProductDocumentFromModel(product models.Product) typesenseProductDocument {
	return typesenseProductDocument{
		ID:          strconv.FormatUint(uint64(product.ID), 10),
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		CreatedAt:   product.CreatedAt.Unix(),
	}
}

func buildTypesenseFilter(params GetProductsParams) string {
	filters := make([]string, 0, 3)

	if params.Category != "" {
		filters = append(filters, fmt.Sprintf("category:=%s", escapeTypesenseFilterValue(params.Category)))
	}

	if params.MinPrice != "" {
		if min, err := strconv.ParseFloat(params.MinPrice, 64); err == nil {
			filters = append(filters, fmt.Sprintf("price:>=%s", strconv.FormatFloat(min, 'f', -1, 64)))
		}
	}

	if params.MaxPrice != "" {
		if maxValue, err := strconv.ParseFloat(params.MaxPrice, 64); err == nil {
			filters = append(filters, fmt.Sprintf("price:<=%s", strconv.FormatFloat(maxValue, 'f', -1, 64)))
		}
	}

	return strings.Join(filters, " && ")
}

func escapeTypesenseFilterValue(value string) string {
	value = strings.TrimSpace(value)
	if strings.ContainsAny(value, ",` ") {
		return "`" + strings.ReplaceAll(value, "`", "\\`") + "`"
	}
	return value
}

func parseUintDocumentID(value string) uint {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return uint(parsed)
}

func SyncAllProductsToTypesense(ctx context.Context, batchSize int) error {
	if config.Typesense == nil {
		return fmt.Errorf("typesense is not enabled")
	}
	if batchSize <= 0 {
		batchSize = 500
	}
	if err := EnsureTypesenseProductsCollection(ctx); err != nil {
		return err
	}

	var lastID uint
	for {
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		var products []models.Product
		err := config.DB.WithContext(batchCtx).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&products).Error
		if err != nil {
			cancel()
			return err
		}
		if len(products) == 0 {
			cancel()
			return nil
		}

		if err := BulkUpsertProductsToTypesense(batchCtx, products); err != nil {
			cancel()
			return err
		}

		lastID = products[len(products)-1].ID
		cancel()
	}
}

func BulkUpsertProductsToTypesense(ctx context.Context, products []models.Product) error {
	if config.Typesense == nil || len(products) == 0 {
		return nil
	}

	if err := EnsureTypesenseProductsCollection(ctx); err != nil {
		return err
	}

	var payload bytes.Buffer
	encoder := json.NewEncoder(&payload)
	for _, product := range products {
		if err := encoder.Encode(typesenseProductDocumentFromModel(product)); err != nil {
			return err
		}
	}

	query := url.Values{}
	query.Set("action", "upsert")

	_, _, err := config.Typesense.Request(
		ctx,
		http.MethodPost,
		"/collections/"+config.Typesense.Collection()+"/documents/import",
		query,
		payload.Bytes(),
	)
	return err
}
