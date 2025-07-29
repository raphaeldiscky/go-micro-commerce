// Package entity defines the ValidatedProduct entity.
package entity

// ValidatedProduct represents a product that has been validated for business rules.
type ValidatedProduct struct {
	Product
	isValidated bool
}

// IsValid checks if the ValidatedProduct is valid.
func (vp *ValidatedProduct) IsValid() bool {
	return vp.isValidated
}

// NewValidatedProduct creates a new ValidatedProduct from a Product after validation.
func NewValidatedProduct(product *Product) (*ValidatedProduct, error) {
	if err := product.validate(); err != nil {
		return nil, err
	}

	return &ValidatedProduct{
		Product:     *product,
		isValidated: true,
	}, nil
}
