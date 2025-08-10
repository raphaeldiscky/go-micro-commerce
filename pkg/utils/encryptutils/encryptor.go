package encryptutils

// Encryptor is an interface for data encryption and decryption.
type Encryptor interface {
	Encrypt(data string) (string, error)
	Decrypt(data string) (string, error)
}
