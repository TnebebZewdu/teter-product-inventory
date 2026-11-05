
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/TnebebZewdu/teter-product-inventory/internal/database"
)

// ProductHandler handles product-related HTTP requests.
type ProductHandler struct {
    queries database.Querier
}

func NewProductHandler(queries database.Querier) *ProductHandler {
    return &ProductHandler{queries: queries}
}
type ProductRequest struct {
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Description   *string `json:"description"`
	Price         string  `json:"price"`
	StockQuantity int32   `json:"stock_quantity"`
	Status        string  `json:"status"`
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
    code := "INTERNAL_SERVER_ERROR"

    switch statusCode {
    case http.StatusBadRequest:
        code = "BAD_REQUEST"
    case http.StatusNotFound:
        code = "NOT_FOUND"
    case http.StatusConflict:
        code = "CONFLICT"
    case http.StatusInternalServerError:
        code = "INTERNAL_SERVER_ERROR"
    }

    writeJSON(w, statusCode, map[string]interface{}{
        "error": map[string]string{
            "code":    code,
            "message": message,
        },
    })
}

type productResponse struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Description   *string `json:"description"`
	Price         string  `json:"price"`
	StockQuantity int32   `json:"stock_quantity"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func toProductResponse(product database.Product) productResponse {
	var description *string

	if product.Description.Valid {
		description = &product.Description.String
	}

	price, err := product.Price.Value()
	if err != nil {
		price = "0"
	}

	return productResponse{
		ID:            product.ID,
		Name:          product.Name,
		SKU:           product.Sku,
		Description:   description,
		Price:         fmt.Sprint(price),
		StockQuantity: product.StockQuantity,
		Status:        product.Status,
		CreatedAt:     product.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:     product.UpdatedAt.Time.Format(time.RFC3339),
	}
}

type paginationResponse struct {
	TotalItems int64 `json:"total_items"`
	Page       int32 `json:"page"`
	PageSize   int32 `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
}

type productListResponse struct {
	Data       []productResponse  `json:"data"`
	Pagination paginationResponse `json:"pagination"`
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	product, err := h.queries.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Product not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Unable to retrieve product")
		return
	}

	writeJSON(w, http.StatusOK, toProductResponse(product))
}

func parsePrice(value string) (pgtype.Numeric, error) {
	var price pgtype.Numeric

	if err := price.Scan(value); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid price")
	}

	if !price.Valid || price.NaN || price.InfinityModifier != pgtype.Finite {
		return pgtype.Numeric{}, fmt.Errorf("price must be a finite number")
	}

	if price.Int == nil || price.Int.Sign() < 0 {
		return pgtype.Numeric{}, fmt.Errorf("price cannot be negative")
	}

	decimal := new(big.Rat)
	if _, ok := decimal.SetString(value); !ok {
		return pgtype.Numeric{}, fmt.Errorf("invalid price")
	}

	maximum := new(big.Rat).SetFrac(big.NewInt(9999999999), big.NewInt(100))
	if decimal.Cmp(maximum) > 0 {
		return pgtype.Numeric{}, fmt.Errorf("price exceeds the maximum allowed value")
	}

	return price, nil
}

func isValidStatus(status string) bool {
	return status == "draft" || status == "active" || status == "archived"
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var request ProductRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if strings.TrimSpace(request.Name) == "" {
		writeError(w, http.StatusBadRequest, "Product name is required")
		return
	}

	if strings.TrimSpace(request.SKU) == "" {
		writeError(w, http.StatusBadRequest, "Product SKU is required")
		return
	}

	price, err := parsePrice(request.Price)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.StockQuantity < 0 {
		writeError(w, http.StatusBadRequest, "Stock quantity cannot be negative")
		return
	}

	if !isValidStatus(request.Status) {
		writeError(w, http.StatusBadRequest, "Status must be draft, active, or archived")
		return
	}

	description := pgtype.Text{String: "", Valid: false}
	if request.Description != nil {
		description = pgtype.Text{String: *request.Description, Valid: true}
	}

	product, err := h.queries.CreateProduct(r.Context(), database.CreateProductParams{
		Name:          strings.TrimSpace(request.Name),
		Sku:           strings.TrimSpace(request.SKU),
		Description:   description,
		Price:         price,
		StockQuantity: request.StockQuantity,
		Status:        request.Status,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "A product with this SKU already exists")
			return
		}

		fmt.Printf("Create product error: %v\n", err)
		writeError(w, http.StatusInternalServerError, "Unable to create product")
		return
	}

	writeJSON(w, http.StatusCreated, toProductResponse(product))
}

// ListProducts supports pagination, status filtering, name search,
// and low-stock filtering.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page := int32(1)
	pageSize := int32(10)

	if value := query.Get("page"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 1 {
			writeError(w, http.StatusBadRequest, "Page must be a positive number")
			return
		}
		page = int32(parsed)
	}

	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "Limit must be between 1 and 100")
			return
		}
		pageSize = int32(parsed)
	}

	status := strings.TrimSpace(query.Get("status"))
	if status != "" && !isValidStatus(status) {
		writeError(w, http.StatusBadRequest, "Status must be draft, active, or archived")
		return
	}

	search := strings.TrimSpace(query.Get("search"))

	lowStock := false
	if value := query.Get("low_stock"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "low_stock must be true or false")
			return
		}
		lowStock = parsed
	}

	offset := (page - 1) * pageSize

	params := database.ListProductsParams{
		Status:     status,
		Search:     search,
		LowStock:   lowStock,
		PageSize:   pageSize,
		PageOffset: offset,
	}

	products, err := h.queries.ListProducts(r.Context(), params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Unable to retrieve products")
		return
	}

	total, err := h.queries.CountProducts(r.Context(), database.CountProductsParams{
		Status:   status,
		Search:   search,
		LowStock: lowStock,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Unable to count products")
		return
	}

	data := make([]productResponse, 0, len(products))
	for _, product := range products {
		data = append(data, toProductResponse(product))
	}

	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	writeJSON(w, http.StatusOK, productListResponse{
		Data: data,
		Pagination: paginationResponse{
			TotalItems: total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var request ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if strings.TrimSpace(request.Name) == "" {
		writeError(w, http.StatusBadRequest, "Product name is required")
		return
	}

	if strings.TrimSpace(request.SKU) == "" {
		writeError(w, http.StatusBadRequest, "Product SKU is required")
		return
	}

	price, err := parsePrice(request.Price)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.StockQuantity < 0 {
		writeError(w, http.StatusBadRequest, "Stock quantity cannot be negative")
		return
	}

	if !isValidStatus(request.Status) {
		writeError(w, http.StatusBadRequest, "Status must be draft, active, or archived")
		return
	}

	description := pgtype.Text{String: "", Valid: false}
	if request.Description != nil {
		description = pgtype.Text{String: *request.Description, Valid: true}
	}

	product, err := h.queries.UpdateProduct(r.Context(), database.UpdateProductParams{
		Name:          strings.TrimSpace(request.Name),
		Sku:           strings.TrimSpace(request.SKU),
		Description:   description,
		Price:         price,
		StockQuantity: request.StockQuantity,
		Status:        request.Status,
		ID:            id,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Product not found")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "A product with this SKU already exists")
			return
		}

		writeError(w, http.StatusInternalServerError, "Unable to update product")
		return
	}

	writeJSON(w, http.StatusOK, toProductResponse(product))
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	_, err = h.queries.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Product not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "Unable to retrieve product")
		return
	}

	if err := h.queries.DeleteProduct(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "Unable to delete product")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}
