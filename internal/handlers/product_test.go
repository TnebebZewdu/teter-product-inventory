
package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/TnebebZewdu/teter-product-inventory/internal/database"
)

func TestListProducts(t *testing.T) {
	mockDB := &MockDatabase{
		Products: []database.Product{},
		Total:    0,
	}

	handler := NewProductHandler(mockDB)

	router := chi.NewRouter()
	router.Get("/products", handler.ListProducts)

	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response productListResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Pagination.TotalItems != 0 {
		t.Errorf(
			"expected 0 total items, got %d",
			response.Pagination.TotalItems,
		)
	}

	if response.Pagination.Page != 1 {
		t.Errorf(
			"expected page 1, got %d",
			response.Pagination.Page,
		)
	}

	if response.Pagination.PageSize != 10 {
		t.Errorf(
			"expected page size 10, got %d",
			response.Pagination.PageSize,
		)
	}
}

func TestListProductsInvalidPage(t *testing.T) {
	mockDB := &MockDatabase{}
	handler := NewProductHandler(mockDB)

	router := chi.NewRouter()
	router.Get("/products", handler.ListProducts)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products?page=0",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestListProductsInvalidStatus(t *testing.T) {
	mockDB := &MockDatabase{}
	handler := NewProductHandler(mockDB)

	router := chi.NewRouter()
	router.Get("/products", handler.ListProducts)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products?status=unknown",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestListProductsInvalidLowStock(t *testing.T) {
	mockDB := &MockDatabase{}
	handler := NewProductHandler(mockDB)

	router := chi.NewRouter()
	router.Get("/products", handler.ListProducts)

	request := httptest.NewRequest(
		http.MethodGet,
		"/products?low_stock=maybe",
		nil,
	)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
func TestListProductsInvalidPageSize(t *testing.T) {
    mockDB := &MockDatabase{}
    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Get("/products", handler.ListProducts)

    request := httptest.NewRequest(
        http.MethodGet,
        "/products?limit=101",
        nil,
    )
    recorder := httptest.NewRecorder()

    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusBadRequest {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusBadRequest,
            recorder.Code,
        )
    }
}
func TestListProductsDatabaseError(t *testing.T) {
    mockDB := &MockDatabase{
        Err: errors.New("database connection failed"),
    }

    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Get("/products", handler.ListProducts)

    request := httptest.NewRequest(
        http.MethodGet,
        "/products",
        nil,
    )
    recorder := httptest.NewRecorder()

    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusInternalServerError {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusInternalServerError,
            recorder.Code,
        )
    }
}
func TestGetProductByIDNotFound(t *testing.T) {
    mockDB := &MockDatabase{
        Err: pgx.ErrNoRows,
    }

    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Get("/products/{id}", handler.GetProductByID)

    request := httptest.NewRequest(
        http.MethodGet,
        "/products/999",
        nil,
    )
    recorder := httptest.NewRecorder()

    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusNotFound {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusNotFound,
            recorder.Code,
        )
    }

	var response map[string]map[string]string

if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
    t.Fatalf("failed to decode error response: %v", err)
}

if response["error"]["code"] != "NOT_FOUND" {
    t.Errorf(
        "expected error code NOT_FOUND, got %s",
        response["error"]["code"],
    )
}

if response["error"]["message"] != "Product not found" {
    t.Errorf(
        "expected error message 'Product not found', got %s",
        response["error"]["message"],
    )
}
}
func TestCreateProductDuplicateSKU(t *testing.T) {
    mockDB := &MockDatabase{
        CreateProductErr: &pgconn.PgError{
            Code: "23505",
        },
    }

    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Post("/products", handler.CreateProduct)

    body := `{
        "name": "Wireless Mouse",
        "sku": "MOUSE-001",
        "description": "Wireless computer mouse",
        "price": "25.50",
        "stock_quantity": 10,
        "status": "active"
    }`

    request := httptest.NewRequest(
        http.MethodPost,
        "/products",
        strings.NewReader(body),
    )
    request.Header.Set("Content-Type", "application/json")

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusConflict {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusConflict,
            recorder.Code,
        )
    }
}
func TestCreateProductInvalidName(t *testing.T) {
    mockDB := &MockDatabase{}
    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Post("/products", handler.CreateProduct)

    body := `{
        "name": "",
        "sku": "MOUSE-002",
        "description": "Wireless computer mouse",
        "price": "25.50",
        "stock_quantity": 10,
        "status": "active"
    }`

    request := httptest.NewRequest(
        http.MethodPost,
        "/products",
        strings.NewReader(body),
    )
    request.Header.Set("Content-Type", "application/json")

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusBadRequest {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusBadRequest,
            recorder.Code,
        )
    }
}
func TestCreateProductInvalidJSON(t *testing.T) {
    mockDB := &MockDatabase{}
    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Post("/products", handler.CreateProduct)

    request := httptest.NewRequest(
        http.MethodPost,
        "/products",
        strings.NewReader(`{"name": "Mouse",`),
    )
    request.Header.Set("Content-Type", "application/json")

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusBadRequest {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusBadRequest,
            recorder.Code,
        )
    }
}
func TestCreateProductSuccess(t *testing.T) {
    mockDB := &MockDatabase{
        CreatedProduct: database.Product{
            ID:     1,
            Name:   "Wireless Mouse",
            Sku:    "MOUSE-002",
            Status: "active",
        },
    }

    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Post("/products", handler.CreateProduct)

    body := `{
        "name": "Wireless Mouse",
        "sku": "MOUSE-002",
        "description": "Wireless computer mouse",
        "price": "25.50",
        "stock_quantity": 10,
        "status": "active"
    }`

    request := httptest.NewRequest(
        http.MethodPost,
        "/products",
        strings.NewReader(body),
    )
    request.Header.Set("Content-Type", "application/json")

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusCreated {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusCreated,
            recorder.Code,
        )
    }

    if !strings.Contains(recorder.Body.String(), "Wireless Mouse") {
        t.Errorf("expected response to contain the created product")
    }
}
func TestGetProductByIDSuccess(t *testing.T) {
    mockDB := &MockDatabase{
        CreatedProduct: database.Product{
            ID:     1,
            Name:   "Wireless Mouse",
            Sku:    "MOUSE-001",
            Status: "active",
        },
    }

    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Get("/products/{id}", handler.GetProductByID)

    request := httptest.NewRequest(
        http.MethodGet,
        "/products/1",
        nil,
    )

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusOK {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusOK,
            recorder.Code,
        )
    }

    if !strings.Contains(recorder.Body.String(), "Wireless Mouse") {
        t.Errorf("expected response to contain Wireless Mouse")
    }
}
func TestDeleteProductSuccess(t *testing.T) {
    mockDB := &MockDatabase{}
    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Delete("/products/{id}", handler.DeleteProduct)

    request := httptest.NewRequest(
        http.MethodDelete,
        "/products/1",
        nil,
    )

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)
if recorder.Code != http.StatusOK {
    t.Errorf(
        "expected status %d, got %d",
        http.StatusOK,
        recorder.Code,
    )
}

if !strings.Contains(recorder.Body.String(), "Product deleted successfully") {
    t.Errorf("expected successful deletion message")
}
}
func TestUpdateProductSuccess(t *testing.T) {
    mockDB := &MockDatabase{
        CreatedProduct: database.Product{
            ID:     1,
            Name:   "Updated Wireless Mouse",
            Sku:    "MOUSE-001",
            Status: "active",
        },
    }

    handler := NewProductHandler(mockDB)

    router := chi.NewRouter()
    router.Put("/products/{id}", handler.UpdateProduct)

    body := `{
        "name": "Updated Wireless Mouse",
        "sku": "MOUSE-001",
        "description": "Updated wireless mouse",
        "price": "30.00",
        "stock_quantity": 15,
        "status": "active"
    }`

    request := httptest.NewRequest(
        http.MethodPut,
        "/products/1",
        strings.NewReader(body),
    )
    request.Header.Set("Content-Type", "application/json")

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, request)

    if recorder.Code != http.StatusOK {
        t.Errorf(
            "expected status %d, got %d",
            http.StatusOK,
            recorder.Code,
        )
    }

    if !strings.Contains(recorder.Body.String(), "Updated Wireless Mouse") {
        t.Errorf("expected response to contain the updated product name")
    }
}