// Package client provides a client for interacting with Elasticsearch.
package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/config"
)

// ElasticSearchClient defines the interface for Elasticsearch operations.
type ElasticSearchClient interface {
	GetClient() *elasticsearch.TypedClient
	DeleteIndex(ctx context.Context, indexName string) error
}

// elasticSearchClient wraps the Elasticsearch client with additional functionality.
type elasticSearchClient struct {
	client *elasticsearch.TypedClient
	logger logger.Logger
}

// NewElasticSearchClient creates a new Elasticsearch client instance.
func NewElasticSearchClient(
	cfg *config.ESConfig,
	appLogger logger.Logger,
) (ElasticSearchClient, error) {
	// Configure HTTP transport for ES v9
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, errors.New("failed to get http transport")
	}

	transport = transport.Clone()
	transport.MaxIdleConns = cfg.MaxIdleConns
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConns
	transport.IdleConnTimeout = cfg.MaxIdleTime

	// Configure TLS settings
	if cfg.EnableSSL {
		// Enable TLS with configurable verification
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: cfg.SkipTLSVerify, //nolint:gosec // false positive
			MinVersion:         tls.VersionTLS12,
		}
	}

	// ES v9 client configuration
	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.GetESURL()},
		Transport: transport,
	}

	// Add authentication if security is enabled
	if cfg.EnableSecurity {
		esConfig.Username = cfg.Username
		esConfig.Password = cfg.Password
	}

	// ES v9 specific configurations
	esConfig.RetryOnStatus = []int{502, 503, 504, 429}
	esConfig.MaxRetries = cfg.MaxRetries
	esConfig.RetryBackoff = func(i int) time.Duration {
		return time.Duration(i) * 100 * time.Millisecond
	}

	// Enable sniffing if configured (usually disabled in containerized environments)
	esConfig.DiscoverNodesOnStart = cfg.SnifferEnabled
	esConfig.DiscoverNodesInterval = cfg.DiscoverNodesInterval

	// Create the typed client
	typedClient, err := elasticsearch.NewTypedClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch TypedClient: %w", err)
	}

	esClient := &elasticSearchClient{
		client: typedClient,
		logger: appLogger,
	}

	// Test connection with direct ping since Ping method will be removed
	if _, err = typedClient.Ping().Do(context.Background()); err != nil {
		appLogger.Warnf("Elasticsearch connection test failed: %v", err)
	} else {
		appLogger.Info("Elasticsearch client connected successfully")
	}

	return esClient, nil
}

// GetClient returns the underlying Elasticsearch client.
func (c *elasticSearchClient) GetClient() *elasticsearch.TypedClient {
	return c.client
}

// DeleteIndex deletes an index.
func (c *elasticSearchClient) DeleteIndex(ctx context.Context, indexName string) error {
	_, err := c.client.Indices.Delete(indexName).Do(ctx)
	if err != nil {
		// Ignore "index not found" errors
		if strings.Contains(err.Error(), "index_not_found") {
			c.logger.Infof("Index %s does not exist, skipping deletion", indexName)

			return nil
		}

		return fmt.Errorf("failed to delete index %s: %w", indexName, err)
	}

	c.logger.Infof("Successfully deleted index: %s", indexName)

	return nil
}
