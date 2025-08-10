// Package encryptutils provides encryption utilities.
package encryptutils

import "encoding/base64"

// Base64Encryptor provides methods for Base64 encoding and decoding.
type Base64Encryptor struct{}

// NewBase64Encryptor creates a new instance of Base64Encryptor.
func NewBase64Encryptor() *Base64Encryptor {
	return &Base64Encryptor{}
}

// Encrypt encodes the input data using Base64.
func (e *Base64Encryptor) Encrypt(data string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

// Decrypt decodes the input data using Base64.
func (e *Base64Encryptor) Decrypt(data string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	return string(decodedBytes), nil
}
