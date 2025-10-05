package handler

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/dto"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/jwtutils"
)

// JWKSHandler handles JWKS endpoint requests.
type JWKSHandler struct {
	jwtUtils jwtutils.JWT
}

// NewJWKSHandler creates a new instance of JWKSHandler.
func NewJWKSHandler(jwtUtils jwtutils.JWT) *JWKSHandler {
	return &JWKSHandler{
		jwtUtils: jwtUtils,
	}
}

// jwk represents a JSON Web Key.
type jwk struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// jwkSet represents a set of JSON Web Keys.
type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// GetJWKS returns the public key as JWKS.
func (h *JWKSHandler) GetJWKS(e echo.Context) error {
	publicKey := h.jwtUtils.GetPublicKey()
	if publicKey == nil {
		return e.JSON(http.StatusServiceUnavailable, dto.WebResponse[any, any]{
			Message: "JWKS not available",
		})
	}

	jwkKey := rsaPublicKeyToJWK(publicKey)

	return e.JSON(http.StatusOK, jwkSet{
		Keys: []jwk{jwkKey},
	})
}

// rsaPublicKeyToJWK converts an RSA public key to JWK format.
func rsaPublicKeyToJWK(publicKey *rsa.PublicKey) jwk {
	return jwk{
		KeyType:   "RSA",
		Use:       "sig",
		Algorithm: "RS256",
		KeyID:     "auth-service-key-1",
		Modulus:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		Exponent:  base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
	}
}
