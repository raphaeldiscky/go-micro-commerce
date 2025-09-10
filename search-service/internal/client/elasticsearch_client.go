// Package client provides a client for interacting with Elasticsearch.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/healthstatus"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/config"
)

// ElasticsearchClient wraps the Elasticsearch client with additional functionality.
type ElasticsearchClient struct {
	client *elasticsearch.TypedClient
	config *config.ElasticsearchConfig
	logger logger.Logger
}

// ElasticsearchClientInterface defines the interface for Elasticsearch operations.
type ElasticsearchClientInterface interface {
	GetClient() *elasticsearch.TypedClient
	Ping(ctx context.Context) error
	HealthCheck(ctx context.Context) (bool, error)
	CreateIndex(ctx context.Context, indexName string, mapping map[string]interface{}) error
	DeleteIndex(ctx context.Context, indexName string) error
	IndexExists(ctx context.Context, indexName string) (bool, error)
}

// NewElasticsearchClient creates a new Elasticsearch client instance.
func NewElasticsearchClient(
	cfg *config.ElasticsearchConfig,
	appLogger logger.Logger,
) (ElasticsearchClientInterface, error) {
	// Configure HTTP transport for ES v9
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("failed to get http transport")
	}

	transport = transport.Clone()
	transport.MaxIdleConns = cfg.MaxIdleConns
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConns
	transport.IdleConnTimeout = time.Duration(cfg.MaxIdleTime) * time.Second

	// Configure TLS if SSL is disabled (for development)
	if !cfg.EnableSSL {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// ES v9 client configuration
	esConfig := elasticsearch.Config{
		Addresses: []string{cfg.GetElasticsearchURL()},
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
	esConfig.DiscoverNodesInterval = 60 * time.Second

	// Create the typed client
	typedClient, err := elasticsearch.NewTypedClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch TypedClient: %w", err)
	}

	esClient := &ElasticsearchClient{
		client: typedClient,
		config: cfg,
		logger: appLogger,
	}

	// Test connection
	if err := esClient.Ping(context.Background()); err != nil {
		appLogger.Warnf("Elasticsearch connection test failed: %v", err)
	} else {
		appLogger.Info("Elasticsearch client connected successfully")
	}

	return esClient, nil
}

// GetClient returns the underlying Elasticsearch client.
func (c *ElasticsearchClient) GetClient() *elasticsearch.TypedClient {
	return c.client
}

// Ping tests the connection to Elasticsearch.
func (c *ElasticsearchClient) Ping(ctx context.Context) error {
	_, err := c.client.Ping().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}

	return nil
}

// HealthCheck checks the health of the Elasticsearch cluster.
func (c *ElasticsearchClient) HealthCheck(ctx context.Context) (bool, error) {
	resp, err := c.client.Cluster.Health().
		WaitForStatus(healthstatus.Yellow).
		Timeout("10s").
		Do(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check Elasticsearch health: %w", err)
	}

	// Check if cluster status is at least yellow
	if resp.Status == healthstatus.Green || resp.Status == healthstatus.Yellow {
		return true, nil
	}

	c.logger.Errorf("Elasticsearch cluster status is: %v", resp.Status)
	return false, nil
}

// CreateIndex creates a new index with the specified mapping.
func (c *ElasticsearchClient) CreateIndex(
	ctx context.Context,
	indexName string,
	mapping map[string]interface{},
) error {
	exists, err := c.IndexExists(ctx, indexName)
	if err != nil {
		return err
	}

	if exists {
		c.logger.Infof("Index %s already exists", indexName)

		return nil
	}

	// Convert map to typed mappings if needed
	createReq := &create.Request{
		Mappings: convertToTypedMappings(mapping),
	}
	
	_, err = c.client.Indices.Create(indexName).
		Request(createReq).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", indexName, err)
	}

	c.logger.Infof("Successfully created index: %s", indexName)

	return nil
}

// DeleteIndex deletes an index.
func (c *ElasticsearchClient) DeleteIndex(ctx context.Context, indexName string) error {
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

// IndexExists checks if an index exists.
func (c *ElasticsearchClient) IndexExists(ctx context.Context, indexName string) (bool, error) {
	_, err := c.client.Indices.Exists(indexName).Do(ctx)
	if err != nil {
		// If the error is "index not found", return false, not an error
		if strings.Contains(err.Error(), "index_not_found") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if index %s exists: %w", indexName, err)
	}

	return true, nil
}

// convertToTypedMappings is a helper to convert from map to TypedAPI mapping
// For now, we'll use a simple approach - in production you'd want proper conversion
func convertToTypedMappings(mapping map[string]interface{}) *types.TypeMapping {
	// This is a simplified conversion - you might want to implement proper mapping conversion
	// based on your specific mapping structure
	if mappings, ok := mapping["mappings"].(map[string]interface{}); ok {
		return &types.TypeMapping{
			Properties: convertToProperties(mappings),
		}
	}
	return &types.TypeMapping{
		Properties: convertToProperties(mapping),
	}
}

// convertToProperties converts map properties to TypedAPI properties
func convertToProperties(props map[string]interface{}) map[string]types.Property {
	if props == nil {
		return nil
	}
	
	result := make(map[string]types.Property)
	if properties, ok := props["properties"].(map[string]interface{}); ok {
		for fieldName, fieldDef := range properties {
			if fieldMap, ok := fieldDef.(map[string]interface{}); ok {
				result[fieldName] = convertToProperty(fieldMap)
			}
		}
	}
	return result
}

// convertToProperty converts individual field definitions
func convertToProperty(fieldDef map[string]interface{}) types.Property {
	fieldType, ok := fieldDef["type"].(string)
	if !ok {
		return types.TextProperty{} // default fallback
	}
	
	switch fieldType {
	case "text":
		prop := types.TextProperty{}
		if analyzer, ok := fieldDef["analyzer"].(string); ok {
			prop.Analyzer = &analyzer
		}
		return prop
	case "keyword":
		return types.KeywordProperty{}
	case "long":
		return types.LongNumberProperty{}
	case "double":
		return types.DoubleNumberProperty{}
	case "float":
		return types.FloatNumberProperty{}
	case "integer":
		return types.IntegerNumberProperty{}
	case "boolean":
		return types.BooleanProperty{}
	case "date":
		return types.DateProperty{}
	case "completion":
		prop := types.CompletionProperty{}
		if analyzer, ok := fieldDef["analyzer"].(string); ok {
			prop.Analyzer = &analyzer
		}
		if maxInputLength, ok := fieldDef["max_input_length"].(float64); ok {
			maxLen := int(maxInputLength)
			prop.MaxInputLength = &maxLen
		}
		return prop
	default:
		return types.TextProperty{} // fallback
	}
}
