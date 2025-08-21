package encryptutils

// HasherInterface is an interface for password hashing and checking.
type HasherInterface interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}
