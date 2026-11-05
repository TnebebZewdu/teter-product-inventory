package handlers

import (
	"context"

	"github.com/TnebebZewdu/teter-product-inventory/internal/database"
)

// MockDatabase provides fake database methods for handler tests.
type MockDatabase struct {
	Products            []database.Product
	CreatedProduct      database.Product
	CreateProductParams database.CreateProductParams
	ProductByID         database.Product
	Total               int64
	Err                 error
	CreateProductErr    error
}

func (m *MockDatabase) CountProducts(
	ctx context.Context,
	arg database.CountProductsParams,
) (int64, error) {
	if m.Err != nil {
		return 0, m.Err
	}
	return m.Total, nil
}

func (m *MockDatabase) CreateProduct(
	ctx context.Context,
	arg database.CreateProductParams,
) (database.Product, error) {
	m.CreateProductParams = arg

	if m.CreateProductErr != nil {
		return database.Product{}, m.CreateProductErr
	}
	if m.Err != nil {
		return database.Product{}, m.Err
	}
	return m.CreatedProduct, nil
}

func (m *MockDatabase) DeleteProduct(
	ctx context.Context,
	id int64,
) error {
	return m.Err
}

func (m *MockDatabase) GetProductByID(
	ctx context.Context,
	id int64,
) (database.Product, error) {
	if m.Err != nil {
		return database.Product{}, m.Err
	}
	return m.CreatedProduct, nil
}

func (m *MockDatabase) GetProductBySKU(
	ctx context.Context,
	sku string,
) (database.Product, error) {
	if m.Err != nil {
		return database.Product{}, m.Err
	}
	return database.Product{}, nil
}

func (m *MockDatabase) ListProducts(
	ctx context.Context,
	arg database.ListProductsParams,
) ([]database.Product, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Products, nil
}

func (m *MockDatabase) UpdateProduct(
	ctx context.Context,
	arg database.UpdateProductParams,
) (database.Product, error) {
	if m.Err != nil {
		return database.Product{}, m.Err
	}
	return m.CreatedProduct, nil
}

// Verify that MockDatabase implements the required interface.
var _ database.Querier = (*MockDatabase)(nil)
