package encryptutils

// EncryptorInterface is an interface for data encryption and decryption.
type EncryptorInterface interface {
	Encrypt(data string) (string, error)
	Decrypt(data string) (string, error)
}
