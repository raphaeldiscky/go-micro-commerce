package events

import "github.com/google/uuid"

// ProductCreated event is published when a new product is created.
type ProductCreated struct {
	BaseDomainEvent
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Price       float64   `json:"price"`
	SellerID    uuid.UUID `json:"seller_id"`
	SellerName  string    `json:"seller_name"`
}

// NewProductCreatedEvent creates a new ProductCreated event.
func NewProductCreatedEvent(
	productID uuid.UUID,
	productName string,
	price float64,
	sellerID uuid.UUID,
	sellerName string,
) *ProductCreated {
	data := struct {
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Price       float64   `json:"price"`
		SellerID    uuid.UUID `json:"seller_id"`
		SellerName  string    `json:"seller_name"`
	}{
		ProductID:   productID,
		ProductName: productName,
		Price:       price,
		SellerID:    sellerID,
		SellerName:  sellerName,
	}

	return &ProductCreated{
		BaseDomainEvent: NewBaseDomainEvent(
			"ProductCreated",
			productID.String(),
			"Product",
			1,
			data,
		),
		ProductID:   productID,
		ProductName: productName,
		Price:       price,
		SellerID:    sellerID,
		SellerName:  sellerName,
	}
}

// ProductUpdated event is published when a product is updated.
type ProductUpdated struct {
	BaseDomainEvent
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Price       float64   `json:"price"`
	SellerID    uuid.UUID `json:"seller_id"`
}

// NewProductUpdatedEvent creates a new ProductUpdated event.
func NewProductUpdatedEvent(
	productID uuid.UUID,
	productName string,
	price float64,
	sellerID uuid.UUID,
) *ProductUpdated {
	data := struct {
		ProductID   uuid.UUID `json:"product_id"`
		ProductName string    `json:"product_name"`
		Price       float64   `json:"price"`
		SellerID    uuid.UUID `json:"seller_id"`
	}{
		ProductID:   productID,
		ProductName: productName,
		Price:       price,
		SellerID:    sellerID,
	}

	return &ProductUpdated{
		BaseDomainEvent: NewBaseDomainEvent(
			"ProductUpdated",
			productID.String(),
			"Product",
			1,
			data,
		),
		ProductID:   productID,
		ProductName: productName,
		Price:       price,
		SellerID:    sellerID,
	}
}

// ProductDeleted event is published when a product is deleted.
type ProductDeleted struct {
	BaseDomainEvent
	ProductID uuid.UUID `json:"product_id"`
}

// NewProductDeletedEvent creates a new ProductDeleted event.
func NewProductDeletedEvent(productID uuid.UUID) *ProductDeleted {
	data := struct {
		ProductID uuid.UUID `json:"product_id"`
	}{
		ProductID: productID,
	}

	return &ProductDeleted{
		BaseDomainEvent: NewBaseDomainEvent(
			"ProductDeleted",
			productID.String(),
			"Product",
			1,
			data,
		),
		ProductID: productID,
	}
}
