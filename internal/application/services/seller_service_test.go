package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/raphaeldiscky/go-ddd-template/internal/application/command"
	"github.com/raphaeldiscky/go-ddd-template/internal/domain/entities"
)

// MockSellerRepository is a mock implementation of the SellerRepository interface.
type MockSellerRepository struct {
	sellers []*entities.ValidatedSeller
}

// Create adds a new seller to the repository.
func (m *MockSellerRepository) Create(seller *entities.ValidatedSeller) (*entities.Seller, error) {
	m.sellers = append(m.sellers, seller)

	return &seller.Seller, nil
}

// FindAll retrieves all sellers.
func (m *MockSellerRepository) FindAll() ([]*entities.Seller, error) {
	var sellers []*entities.Seller
	for _, s := range m.sellers {
		sellers = append(sellers, &s.Seller)
	}

	return sellers, nil
}

// FindByID retrieves a seller by its ID.
func (m *MockSellerRepository) FindByID(id uuid.UUID) (*entities.Seller, error) {
	for _, s := range m.sellers {
		if s.Id == id {
			return &s.Seller, nil
		}

		fmt.Printf("Id: %s - %s\n", s.Id, id)
	}

	return nil, errors.New("seller not found")
}

// Delete removes a seller from the repository.
func (m *MockSellerRepository) Delete(id uuid.UUID) error {
	for index, s := range m.sellers {
		if s.Id == id {
			m.sellers = append(m.sellers[:index], m.sellers[index+1:]...)

			return nil
		}
	}

	return errors.New("seller not found for deletion")
}

// Update modifies an existing seller in the repository.
func (m *MockSellerRepository) Update(seller *entities.ValidatedSeller) (*entities.Seller, error) {
	for index, s := range m.sellers {
		if s.Id == seller.Id {
			m.sellers[index] = seller

			return &seller.Seller, nil
		}
	}

	return nil, errors.New("seller not found for update")
}

// TestSellerService_CreateSeller tests the CreateSeller method of SellerService.
func TestSellerService_CreateSeller(t *testing.T) {
	repo := &MockSellerRepository{}
	service := NewSellerService(repo, nil) // nil for eventPublisher in tests

	_, err := service.CreateSeller(getCreateSellerCommand("John Doe"))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if len(repo.sellers) != 1 {
		t.Errorf("Expected 1 seller in productRepository, but got %d", len(repo.sellers))
	}
}

func TestSellerService_GetAllSellers(t *testing.T) {
	repo := &MockSellerRepository{}
	service := NewSellerService(repo, nil) // nil for eventPublisher in tests

	// Add two sellers
	_, err := service.CreateSeller(getCreateSellerCommand("John Doe"))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	_, err = service.CreateSeller(getCreateSellerCommand("Jane Doe"))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	sellers, err := service.FindAllSellers()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if len(sellers.Result) != 2 {
		t.Errorf("Expected 2 sellers, but got %d", len(sellers.Result))
	}
}

func TestSellerService_GetSellerById(t *testing.T) {
	repo := &MockSellerRepository{}
	service := NewSellerService(repo, nil) // nil for eventPublisher in tests

	createdSellerResult, err := service.CreateSeller(getCreateSellerCommand("John Doe"))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	sellerID := createdSellerResult.Result.Id

	foundSeller, err := service.FindSellerByID(sellerID)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if foundSeller.Result.Name != "John Doe" {
		t.Errorf("Expected seller name 'John Doe', but got %s", foundSeller.Result.Name)
	}

	_, err = service.FindSellerByID(uuid.New()) // some non-existent Id
	if err == nil {
		t.Error("Expected error for non-existent seller, but got none")
	}
}

// TestSellerService_UpdateSeller tests the UpdateSeller method of SellerService.
func TestSellerService_UpdateSeller(t *testing.T) {
	repo := &MockSellerRepository{}
	service := NewSellerService(repo, nil) // nil for eventPublisher in tests

	createdSellerResult, err := service.CreateSeller(getCreateSellerCommand("John Doe"))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	sellerID := createdSellerResult.Result.Id

	updatableSeller := entities.Seller{
		Id:   sellerID,
		Name: "Doe Johnny",
	}

	_, err = service.UpdateSeller(&command.UpdateSellerCommand{
		Id:   sellerID,
		Name: updatableSeller.Name,
	})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	updatedSeller, err := service.FindSellerByID(sellerID)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if updatedSeller.Result.Name != "Doe Johnny" {
		t.Errorf("Expected seller name 'Johnny Doe', but got %s", updatedSeller.Result.Name)
	}
}

// TestSellerService_DeleteSeller tests the DeleteSeller method of SellerService.
func getCreateSellerCommand(name string) *command.CreateSellerCommand {
	return &command.CreateSellerCommand{
		Name:  name,
		Email: name + "@example.com",
	}
}
