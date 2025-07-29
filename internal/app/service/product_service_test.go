package services

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/app/command"
	entity "github.com/raphaeldiscky/go-ddd-template/internal/domain/entity"
)

// MockProductRepository is a mock implementation of the ProductRepository interface.
type MockProductRepository struct {
	products []*entity.ValidatedProduct
}

func (m *MockProductRepository) Create(
	product *entity.ValidatedProduct,
) (*entity.Product, error) {
	m.products = append(m.products, product)

	return &product.Product, nil
}

func (m *MockProductRepository) FindAll() ([]*entity.Product, error) {
	var products []*entity.Product
	for _, p := range m.products {
		products = append(products, &p.Product)
	}

	return products, nil
}

func (m *MockProductRepository) Update(
	product *entity.ValidatedProduct,
) (*entity.Product, error) {
	for index, p := range m.products {
		if p.ID == product.ID {
			m.products[index] = product

			return &product.Product, nil
		}
	}

	return nil, errors.New("product not found for update")
}

func (m *MockProductRepository) Delete(id uuid.UUID) error {
	for index, p := range m.products {
		if p.ID == id {
			m.products = append(m.products[:index], m.products[index+1:]...)

			return nil
		}
	}

	return errors.New("product not found for delete")
}

func (m *MockProductRepository) FindByID(id uuid.UUID) (*entity.Product, error) {
	for _, p := range m.products {
		if p.ID == id {
			return &p.Product, nil
		}
	}

	return nil, errors.New("product not found")
}

func TestProductService_CreateProduct(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	service := NewProductService(productRepo, sellerRepo, nil) // nil for eventPublisher in tests

	// Create seller
	seller := createPersistedSeller(t, sellerRepo)

	// Create product
	product := entity.NewProduct("Example", 100.0, seller)
	productCommand := getCreateProductCommand(product)

	_, err := service.CreateProduct(productCommand)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if len(productRepo.products) != 1 {
		t.Errorf("Expected 1 product in productRepository, but got %d", len(productRepo.products))
	}
}

func TestProductService_GetAllProducts(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	service := NewProductService(productRepo, sellerRepo, nil) // nil for eventPublisher in tests

	// Create seller
	seller := createPersistedSeller(t, sellerRepo)

	// Add two products
	_, err := service.CreateProduct(
		getCreateProductCommand(entity.NewProduct("Example1", 100.0, seller)),
	)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	_, err = service.CreateProduct(
		getCreateProductCommand(entity.NewProduct("Example2", 200.0, seller)),
	)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	products, err := service.FindAllProducts()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if len(products.Result) != 2 {
		t.Errorf("Expected 2 products, but got %d", len(products.Result))
	}
}

func TestProductService_FindProductById(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	service := NewProductService(productRepo, sellerRepo, nil) // nil for eventPublisher in tests

	// Create seller
	seller := createPersistedSeller(t, sellerRepo)

	product := entity.NewProduct("Example", 100.0, seller)

	result, err := service.CreateProduct(getCreateProductCommand(product))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	foundProduct, err := service.FindProductByID(result.Result.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if foundProduct.Result.Name != "Example" {
		t.Errorf("Expected product name 'Example', but got %s", foundProduct.Result.Name)
	}

	_, err = service.FindProductByID(uuid.New()) // some non-existent Id
	if err == nil {
		t.Error("Expected error for non-existent product, but got none")
	}
}

func getCreateProductCommand(product *entity.Product) *command.CreateProductCommand {
	return &command.CreateProductCommand{
		Name:     product.Name,
		Price:    product.Price,
		SellerID: product.Seller.ID,
	}
}

// MockSellerRepository is a mock implementation of the SellerRepository interface.
func createPersistedSeller(
	t *testing.T,
	sellerRepo *MockSellerRepository,
) *entity.ValidatedSeller {
	t.Helper()

	seller := entity.NewSeller("John Doe", "john@example.com")

	validatedSeller, err := entity.NewValidatedSeller(seller)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	_, err = sellerRepo.Create(validatedSeller)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	return validatedSeller
}
