package testcontainers

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = l.Close()
		if err != nil {
			panic(err)
		}
	}()

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("failed to get TCP address")
	}

	return tcpAddr.Port, nil
}

// GenerateTestJWT generates a JWT token for testing.
func GenerateTestJWT(userID uuid.UUID, roles []string) (string, error) {
	// Use auth-service private key for signing test tokens
	privateKeyPath := "../../../auth-service/keys/private.pem"

	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create token with claims matching auth-service format
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       userID.String(),
		"user_id":   userID.String(),
		"email":     fmt.Sprintf("test-%s@example.com", userID.String()[:8]),
		"roles":     roles,
		"is_active": true,
		"iat":       now.Unix(),
		"exp":       now.Add(1 * time.Hour).Unix(),
		"nbf":       now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
