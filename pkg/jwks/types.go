// Package jwks provides JWKS (JSON Web Key Set) fetching and caching functionality.
package jwks

// JWKSet represents a JSON Web Key Set as defined in RFC 7517.
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key as defined in RFC 7517.
type JWK struct {
	KeyType   string `json:"kty"`           // Key Type (e.g., "RSA")
	Use       string `json:"use,omitempty"` // Public Key Use (e.g., "sig" for signature)
	Algorithm string `json:"alg,omitempty"` // Algorithm (e.g., "RS256")
	KeyID     string `json:"kid,omitempty"` // Key ID
	Modulus   string `json:"n,omitempty"`   // RSA modulus (base64url encoded)
	Exponent  string `json:"e,omitempty"`   // RSA exponent (base64url encoded)
}
