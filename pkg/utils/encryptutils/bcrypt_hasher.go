package encryptutils

import "golang.org/x/crypto/bcrypt"

// BcryptHasher provides methods for hashing and checking passwords.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new instance of BcryptHasher.
func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{
		cost: cost,
	}
}

// Hash hashes the password using bcrypt.
func (h *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// Check compares the provided password with the hashed password.
func (h *BcryptHasher) Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
