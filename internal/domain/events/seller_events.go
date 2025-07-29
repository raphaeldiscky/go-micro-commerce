package events

import "github.com/google/uuid"

// SellerCreated event is published when a new seller is created.
type SellerCreated struct {
	BaseDomainEvent
	SellerID   uuid.UUID `json:"seller_id"`
	SellerName string    `json:"seller_name"`
}

// NewSellerCreatedEvent creates a new SellerCreated event.
func NewSellerCreatedEvent(sellerID uuid.UUID, sellerName string) *SellerCreated {
	data := struct {
		SellerID   uuid.UUID `json:"seller_id"`
		SellerName string    `json:"seller_name"`
	}{
		SellerID:   sellerID,
		SellerName: sellerName,
	}

	return &SellerCreated{
		BaseDomainEvent: NewBaseDomainEvent(
			"SellerCreated",
			sellerID.String(),
			"Seller",
			1,
			data,
		),
		SellerID:   sellerID,
		SellerName: sellerName,
	}
}

// SellerUpdated event is published when a seller is updated.
type SellerUpdated struct {
	BaseDomainEvent
	SellerID   uuid.UUID `json:"seller_id"`
	SellerName string    `json:"seller_name"`
}

// NewSellerUpdatedEvent creates a new SellerUpdated event.
func NewSellerUpdatedEvent(sellerID uuid.UUID, sellerName string) *SellerUpdated {
	data := struct {
		SellerID   uuid.UUID `json:"seller_id"`
		SellerName string    `json:"seller_name"`
	}{
		SellerID:   sellerID,
		SellerName: sellerName,
	}

	return &SellerUpdated{
		BaseDomainEvent: NewBaseDomainEvent(
			"SellerUpdated",
			sellerID.String(),
			"Seller",
			1,
			data,
		),
		SellerID:   sellerID,
		SellerName: sellerName,
	}
}

// SellerDeleted event is published when a seller is deleted.
type SellerDeleted struct {
	BaseDomainEvent
	SellerID uuid.UUID `json:"seller_id"`
}

// NewSellerDeletedEvent creates a new SellerDeleted event.
func NewSellerDeletedEvent(sellerID uuid.UUID) *SellerDeleted {
	data := struct {
		SellerID uuid.UUID `json:"seller_id"`
	}{
		SellerID: sellerID,
	}

	return &SellerDeleted{
		BaseDomainEvent: NewBaseDomainEvent(
			"SellerDeleted",
			sellerID.String(),
			"Seller",
			1,
			data,
		),
		SellerID: sellerID,
	}
}
