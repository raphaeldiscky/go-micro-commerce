package entities

// ValidatedSeller represents a seller that has been validated.
type ValidatedSeller struct {
	Seller
	isValidated bool
}

// IsValid checks if the ValidatedSeller is valid.
func (vp *ValidatedSeller) IsValid() bool {
	return vp.isValidated
}

// NewValidatedSeller creates a new ValidatedSeller from a Seller and validates it.
func NewValidatedSeller(seller *Seller) (*ValidatedSeller, error) {
	if err := seller.validate(); err != nil {
		return nil, err
	}

	return &ValidatedSeller{
		Seller:      *seller,
		isValidated: true,
	}, nil
}
