package encryptutils

// Hasher is an interface for password hashing and checking.
type Hasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}
