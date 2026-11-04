package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"errors"
    "strconv"
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
    "github.com/go-chi/chi/v5"
    "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/TnebebZewdu/teter-product-inventory/internal/database"
)

// ProductHandler handles product-related HTTP requests.
type ProductHandler struct {
	queries *database.Queries
}

// NewProductHandler creates a new product handler.
func NewProductHandler(queries *database.Queries) *ProductHandler {
	return &ProductHandler{
		queries: queries,
	}
}

// ProductRequest represents the data sent when creating or updating a product.
type ProductRequest struct {
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Description   *string `json:"description"`
	Price         string  `json:"price"`
	StockQuantity int32   `json:"stock_quantity"`
	Status        string  `json:"status"`
}

// writeJSON sends a JSON response with the given HTTP status code.
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

// writeError sends a consistent JSON error response.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

// productResponse represents a product returned by the API.
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

// toProductResponse converts a database product into an API response.
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

// productListResponse represents a paginated list of products.
type productListResponse struct {
	Data       []productResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int32             `json:"page"`
	Limit      int32             `json:"limit"`
	TotalPages int64             `json:"total_pages"`
}

// GetProductByID handles requests to retrieve a product by its ID.
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	// Read the product ID from the URL.
	idParam := chi.URLParam(r, "id")

	// Convert the ID from string to int64.
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Retrieve the product from the database.
	product, err := h.queries.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Product not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "Unable to retrieve product")
		return
	}

	// Convert the database product into an API response.
	writeJSON(w, http.StatusOK, toProductResponse(product))
}

// parsePrice validates a decimal price and converts it to pgtype.Numeric.
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

// isValidStatus checks whether a product status is allowed.
func isValidStatus(status string) bool {
	return status == "draft" ||
		status == "active" ||
		status == "archived"
}

// CreateProduct handles requests to create a new product.
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var request ProductRequest

	// Decode the incoming JSON.
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate required fields.
	if strings.TrimSpace(request.Name) == "" {
		writeError(w, http.StatusBadRequest, "Product name is required")
		return
	}

	if strings.TrimSpace(request.SKU) == "" {
		writeError(w, http.StatusBadRequest, "Product SKU is required")
		return
	}

	// Validate price.
	price, err := parsePrice(request.Price)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate stock quantity.
	if request.StockQuantity < 0 {
		writeError(w, http.StatusBadRequest, "Stock quantity cannot be negative")
		return
	}

	// Validate status.
	if !isValidStatus(request.Status) {
		writeError(w, http.StatusBadRequest, "Status must be draft, active, or archived")
		return
	}

	// Prepare the nullable description.
	description := pgtype.Text{
		String: "",
		Valid:  false,
	}

	if request.Description != nil {
		description = pgtype.Text{
			String: *request.Description,
			Valid:  true,
		}
	}

	// Save the product in PostgreSQL.
	product, err := h.queries.CreateProduct(r.Context(), database.CreateProductParams{
		Name:          strings.TrimSpace(request.Name),
		Sku:           strings.TrimSpace(request.SKU),
		Description:   description,
		Price:         price,
		StockQuantity: request.StockQuantity,
		Status:        request.Status,
	})

	// Handle database errors.
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

	// Return the created product.
	writeJSON(w, http.StatusCreated, toProductResponse(product))
}

// ListProducts handles requests to retrieve products with pagination.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Read pagination values from the URL.
	pageParam := r.URL.Query().Get("page")
	limitParam := r.URL.Query().Get("limit")

	// Set default values.
	page := int32(1)
	limit := int32(10)

	// Validate page number.
	if pageParam != "" {
		value, err := strconv.ParseInt(pageParam, 10, 32)
		if err != nil || value < 1 {
			writeError(w, http.StatusBadRequest, "Page must be a positive number")
			return
		}
		page = int32(value)
	}

	// Validate page size.
	if limitParam != "" {
		value, err := strconv.ParseInt(limitParam, 10, 32)
		if err != nil || value < 1 || value > 100 {
			writeError(w, http.StatusBadRequest, "Limit must be between 1 and 100")
			return
		}
		limit = int32(value)
	}

	// Calculate how many products to skip.
	offset := (page - 1) * limit

	// Retrieve products from the database.
	products, err := h.queries.ListProducts(r.Context(), database.ListProductsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Unable to retrieve products")
		return
	}

	// Count all products.
	total, err := h.queries.CountProducts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Unable to count products")
		return
	}

	// Convert database products into API responses.
	data := make([]productResponse, 0, len(products))
	for _, product := range products {
		data = append(data, toProductResponse(product))
	}

	// Calculate total pages.
	totalPages := (total + int64(limit) - 1) / int64(limit)

	// Return the paginated response.
	writeJSON(w, http.StatusOK, productListResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// UpdateProduct handles requests to update an existing product.
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Get and validate the product ID from the URL.
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Read the JSON request body.
	var request ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	// Validate the product information.
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

	// Convert the optional description.
	description := pgtype.Text{String: "", Valid: false}
	if request.Description != nil {
		description = pgtype.Text{
			String: *request.Description,
			Valid:  true,
		}
	}

	// Update the product in the database.
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

// DeleteProduct handles requests to delete an existing product.
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Get and validate the product ID.
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Check whether the product exists.
	_, err = h.queries.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Product not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "Unable to retrieve product")
		return
	}

	// Delete the product.
	err = h.queries.DeleteProduct(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Unable to delete product")
		return
	}

	// Return a successful response.
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}