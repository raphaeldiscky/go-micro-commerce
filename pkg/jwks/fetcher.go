package jwks

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/bytedance/sonic"

	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
)

const (
	// defaultHTTPTimeout is the default timeout for HTTP requests.
	defaultHTTPTimeout = 10 * time.Second
)

var (
	// ErrNoKeys is returned when no keys are found in the JWKS.
	ErrNoKeys = errors.New("no keys found in JWKS")
	// ErrInvalidKeyType is returned when the key type is not RSA.
	ErrInvalidKeyType = errors.New("invalid key type, expected RSA")
)

// Fetcher fetches and caches JWKS from a remote endpoint.
type Fetcher struct {
	url             string
	cacheTTL        time.Duration
	refreshInterval time.Duration
	httpClient      *http.Client

	mu        sync.RWMutex
	publicKey *rsa.PublicKey
	lastFetch time.Time
	ctx       context.Context
	cancel    context.CancelFunc

	logger logger.Logger
}

// NewFetcher creates a new JWKS fetcher with caching and auto-refresh.
func NewFetcher(
	url string,
	cacheTTL, refreshInterval time.Duration,
	logger logger.Logger,
) *Fetcher {
	ctx, cancel := context.WithCancel(context.Background())

	f := &Fetcher{
		url:             url,
		cacheTTL:        cacheTTL,
		refreshInterval: refreshInterval,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}

	// Initial fetch - ignore error to allow fallback to file-based keys
	if err := f.fetch(); err != nil {
		logger.Warnf("failed to fetch JWKS: %v", err)
	}

	// Start background refresh
	go f.refreshLoop()

	return f
}

// GetPublicKey returns the cached public key.
func (f *Fetcher) GetPublicKey() (*rsa.PublicKey, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.publicKey == nil {
		return nil, errors.New("no public key available")
	}

	// Check if cache expired
	if time.Since(f.lastFetch) > f.cacheTTL {
		return nil, errors.New("cached key expired")
	}

	return f.publicKey, nil
}

// fetch fetches JWKS from the remote endpoint and updates the cache.
func (f *Fetcher) fetch() error {
	req, err := http.NewRequestWithContext(f.ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create JWKS request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			f.logger.Errorf("failed to close JWKS response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKSet
	if err = sonic.ConfigFastest.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	if len(jwks.Keys) == 0 {
		return ErrNoKeys
	}

	// Use the first key (for key rotation, you'd match by kid)
	jwk := jwks.Keys[0]

	publicKey, err := jwkToRSAPublicKey(&jwk)
	if err != nil {
		return fmt.Errorf("failed to convert JWK to RSA public key: %w", err)
	}

	f.mu.Lock()
	f.publicKey = publicKey
	f.lastFetch = time.Now()
	f.mu.Unlock()

	return nil
}

// refreshLoop periodically refreshes the JWKS cache.
func (f *Fetcher) refreshLoop() {
	ticker := time.NewTicker(f.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := f.fetch() // Ignore errors, keep using cached key
			if err != nil {
				f.logger.Warnf("failed to refresh JWKS: %v", err)
			}
		case <-f.ctx.Done():
			return
		}
	}
}

// Stop stops the background refresh loop.
func (f *Fetcher) Stop() {
	f.cancel()
}

// jwkToRSAPublicKey converts a JWK to an RSA public key.
func jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	if jwk.KeyType != "RSA" {
		return nil, ErrInvalidKeyType
	}

	// Decode modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.Modulus)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.Exponent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big.Int
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}
