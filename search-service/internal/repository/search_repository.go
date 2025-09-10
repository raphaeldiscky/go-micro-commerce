package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/textquerytype"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/client"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/entity"
)

// SearchRepositoryInterface defines the interface for search operations.
type SearchRepositoryInterface interface {
	// Product operations
	IndexProduct(ctx context.Context, product *entity.ProductDocument) error
	UpdateProduct(ctx context.Context, product *entity.ProductDocument) error
	DeleteProduct(ctx context.Context, productID string) error
	SearchProducts(ctx context.Context, query *entity.SearchQuery) (*entity.SearchResponse, error)
	GetProduct(ctx context.Context, productID string) (*entity.ProductDocument, error)

	// Bulk operations
	BulkIndex(ctx context.Context, documents []entity.SearchDocument) error
	BulkUpdate(ctx context.Context, documents []entity.SearchDocument) error
	BulkDelete(ctx context.Context, documentIDs []string, indexName string) error

	// Index operations
	CreateIndices(ctx context.Context) error
	DeleteIndices(ctx context.Context) error
	RefreshIndices(ctx context.Context) error

	// Autocomplete and suggestions
	AutoComplete(ctx context.Context, query string, documentType string) ([]string, error)
	GetSuggestions(
		ctx context.Context,
		query string,
		documentType string,
	) ([]entity.SuggestionResult, error)
}

// SearchRepository implements SearchRepository using Elasticsearch.
type SearchRepository struct {
	client client.ElasticsearchClientInterface
	logger logger.Logger
}

// NewSearchRepository creates a new Elasticsearch repository.
func NewSearchRepository(
	clt client.ElasticsearchClientInterface,
	appLogger logger.Logger,
) SearchRepositoryInterface {
	return &SearchRepository{
		client: clt,
		logger: appLogger,
	}
}

// Product operations
// IndexProduct indexes a product document using TypedAPI.
func (r *SearchRepository) IndexProduct(
	ctx context.Context,
	product *entity.ProductDocument,
) error {
	_, err := r.client.GetClient().Index(product.GetIndexName()).
		Id(product.GetID()).
		Document(product).
		Refresh(refresh.True).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to index product: %w", err)
	}

	r.logger.Infof("Successfully indexed product: %s", product.GetID())
	return nil
}

// UpdateProduct updates a product document using TypedAPI.
func (r *SearchRepository) UpdateProduct(
	ctx context.Context,
	product *entity.ProductDocument,
) error {
	_, err := r.client.GetClient().Update(
		product.GetIndexName(),
		product.GetID(),
	).Doc(product).Refresh(refresh.True).Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	r.logger.Infof("Successfully updated product: %s", product.GetID())
	return nil
}

// GetProduct retrieves a product document using TypedAPI.
func (r *SearchRepository) GetProduct(
	ctx context.Context,
	productID string,
) (*entity.ProductDocument, error) {
	resp, err := r.client.GetClient().Get("products", productID).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if !resp.Found {
		return nil, fmt.Errorf("product not found: %s", productID)
	}

	var product entity.ProductDocument
	if err := json.Unmarshal(resp.Source_, &product); err != nil {
		return nil, fmt.Errorf("failed to decode product: %w", err)
	}

	return &product, nil
}

// DeleteProduct deletes a product document using TypedAPI.
func (r *SearchRepository) DeleteProduct(ctx context.Context, productID string) error {
	_, err := r.client.GetClient().Delete("products", productID).
		Refresh(refresh.True).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	r.logger.Infof("Successfully deleted product: %s", productID)
	return nil
}

// SearchProducts searches for products using TypedAPI.
func (r *SearchRepository) SearchProducts(
	ctx context.Context,
	query *entity.SearchQuery,
) (*entity.SearchResponse, error) {
	// Build the search request using typed queries
	from := query.From
	size := query.Size
	searchRequest := &search.Request{
		Query: &types.Query{
			Bool: &types.BoolQuery{
				Must: []types.Query{
					{
						MultiMatch: &types.MultiMatchQuery{
							Query:     query.Query,
							Fields:    []string{"name^2", "description"},
							Type:      &textquerytype.Bestfields,
							Fuzziness: types.Fuzziness("AUTO"),
						},
					},
				},
			},
		},
		From: &from,
		Size: &size,
	}

	// Add sorting if specified
	if len(query.Sort) > 0 {
		var sorts []types.SortCombinations
		for _, sort := range query.Sort {
			sorts = append(sorts, types.SortOptions{
				SortOptions: map[string]types.FieldSort{
					sort.Field: {Order: types.SortOrder(sort.Order)},
				},
			})
		}
		searchRequest.Sort = sorts
	}

	resp, err := r.client.GetClient().Search().
		Index("products").
		Request(searchRequest).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}

	return r.parseTypedSearchResponse(resp, query)
}

// parseTypedSearchResponse parses the typed search response.
func (r *SearchRepository) parseTypedSearchResponse(
	resp *search.Response,
	query *entity.SearchQuery,
) (*entity.SearchResponse, error) {
	results := make([]entity.SearchResult, 0, len(resp.Hits.Hits))

	for _, hit := range resp.Hits.Hits {
		var source map[string]interface{}
		if err := hit.Source_.Decode(&source); err != nil {
			r.logger.Warnf("Failed to decode hit source: %v", err)
			continue
		}

		searchResult := entity.SearchResult{
			ID:     hit.Id_,
			Score:  float64(*hit.Score),
			Source: source,
		}

		// Handle highlights if available
		if hit.Highlight != nil {
			searchResult.Highlight = make(map[string][]string)
			for field, fragments := range hit.Highlight {
				searchResult.Highlight[field] = fragments
			}
		}

		results = append(results, searchResult)
	}

	perPage := query.Size
	if perPage == 0 {
		perPage = 10
	}

	page := (query.From / perPage) + 1
	totalPages := int((resp.Hits.Total.Value + int64(perPage) - 1) / int64(perPage))

	response := &entity.SearchResponse{
		Results:    results,
		Total:      resp.Hits.Total.Value,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		Took:       int(resp.Took),
	}

	return response, nil
}

// BulkIndex performs bulk indexing using TypedAPI.
func (r *SearchRepository) BulkIndex(ctx context.Context, documents []entity.SearchDocument) error {
	if len(documents) == 0 {
		return nil
	}

	var operations []types.Operation
	for _, doc := range documents {
		operations = append(operations,
			types.Operation{
				Index: &types.IndexOperation{
					Id_:    doc.GetID(),
					Index_: doc.GetIndexName(),
				},
			},
			types.Operation{
				Document: doc,
			},
		)
	}

	resp, err := r.client.GetClient().Bulk().
		Operations(operations...).
		Refresh("true").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to execute bulk index: %w", err)
	}

	if resp.Errors {
		r.logger.Warnf("Bulk index completed with errors")
		// You can iterate through resp.Items to handle specific errors
	} else {
		r.logger.Infof("Successfully bulk indexed %d documents", len(documents))
	}

	return nil
}

// CreateIndices creates indices with proper typed mappings.
func (r *SearchRepository) CreateIndices(ctx context.Context) error {
	// Product mapping with typed fields
	productMapping := types.TypeMapping{
		Properties: map[string]types.Property{
			"id":               types.KeywordProperty{},
			"name":             types.TextProperty{Analyzer: "standard"},
			"description":      types.TextProperty{Analyzer: "english"},
			"price":            types.FloatNumberProperty{},
			"category":         types.KeywordProperty{},
			"brand":            types.KeywordProperty{},
			"in_stock":         types.BooleanProperty{},
			"tags":             types.KeywordProperty{},
			"attributes.color": types.KeywordProperty{},
			"attributes.size":  types.KeywordProperty{},
			"rating":           types.HalfFloatNumberProperty{},
			"review_count":     types.IntegerNumberProperty{},
			"created_at":       types.DateProperty{},
			"updated_at":       types.DateProperty{},
			"suggest": types.CompletionProperty{
				Analyzer:                   "simple",
				PreserveSeparators:         true,
				PreservePositionIncrements: true,
				MaxInputLength:             50,
			},
		},
	}

	_, err := r.client.GetClient().Indices.Create("products").
		Mappings(productMapping).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to create products index: %w", err)
	}

	r.logger.Info("Successfully created products index")
	return nil
}

// BulkUpdate performs bulk updates using TypedAPI.
func (r *SearchRepository) BulkUpdate(ctx context.Context, documents []entity.SearchDocument) error {
	if len(documents) == 0 {
		return nil
	}

	var operations []types.OperationContainer
	for _, doc := range documents {
		operations = append(operations, types.OperationContainer{
			Update: &types.UpdateOperation{
				Index_: &doc.GetIndexName(),
				Id_:    &doc.GetID(),
			},
		})
		operations = append(operations, types.OperationContainer{
			Doc: doc,
		})
	}

	resp, err := r.client.GetClient().Bulk().
		Operations(operations...).
		Refresh(refresh.True).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to execute bulk update: %w", err)
	}

	if resp.Errors {
		r.logger.Warnf("Bulk update completed with errors")
	} else {
		r.logger.Infof("Successfully bulk updated %d documents", len(documents))
	}

	return nil
}

// BulkDelete performs bulk deletion using TypedAPI.
func (r *SearchRepository) BulkDelete(ctx context.Context, documentIDs []string, indexName string) error {
	if len(documentIDs) == 0 {
		return nil
	}

	var operations []types.OperationContainer
	for _, docID := range documentIDs {
		operations = append(operations, types.OperationContainer{
			Delete: &types.DeleteOperation{
				Index_: &indexName,
				Id_:    &docID,
			},
		})
	}

	resp, err := r.client.GetClient().Bulk().
		Operations(operations...).
		Refresh(refresh.True).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to execute bulk delete: %w", err)
	}

	if resp.Errors {
		r.logger.Warnf("Bulk delete completed with errors")
	} else {
		r.logger.Infof("Successfully bulk deleted %d documents", len(documentIDs))
	}

	return nil
}

// DeleteIndices deletes indices using TypedAPI.
func (r *SearchRepository) DeleteIndices(ctx context.Context) error {
	indices := []string{"products"}

	for _, indexName := range indices {
		_, err := r.client.GetClient().Indices.Delete(indexName).Do(ctx)
		if err != nil {
			r.logger.Warnf("Failed to delete index %s: %v", indexName, err)
		} else {
			r.logger.Infof("Successfully deleted index: %s", indexName)
		}
	}

	return nil
}

// RefreshIndices refreshes indices using TypedAPI.
func (r *SearchRepository) RefreshIndices(ctx context.Context) error {
	_, err := r.client.GetClient().Indices.Refresh().
		Index("products").
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to refresh indices: %w", err)
	}

	return nil
}

// AutoComplete provides autocomplete functionality using TypedAPI.
func (r *SearchRepository) AutoComplete(ctx context.Context, query string, documentType string) ([]string, error) {
	indexName := r.getIndexNameByType(documentType)
	if indexName == "" {
		return nil, fmt.Errorf("unknown document type: %s", documentType)
	}

	searchReq := &search.Request{
		Suggest: map[string]types.Suggester{
			"autocomplete": types.Suggester{
				Completion: &types.CompletionSuggester{
					Field: "suggest",
					Size:  types.Uint(10),
				},
				Prefix: &query,
			},
		},
	}

	resp, err := r.client.GetClient().Search().
		Index(indexName).
		Request(searchReq).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to execute autocomplete search: %w", err)
	}

	var suggestions []string
	if resp.Suggest != nil {
		if autocomplete, ok := resp.Suggest["autocomplete"]; ok {
			for _, item := range autocomplete {
				for _, option := range item.Options {
					suggestions = append(suggestions, option.Text)
				}
			}
		}
	}

	return suggestions, nil
}

// GetSuggestions provides enhanced suggestions using TypedAPI.
func (r *SearchRepository) GetSuggestions(ctx context.Context, query string, documentType string) ([]entity.SuggestionResult, error) {
	indexName := r.getIndexNameByType(documentType)
	if indexName == "" {
		return nil, fmt.Errorf("unknown document type: %s", documentType)
	}

	searchReq := &search.Request{
		Suggest: map[string]types.Suggester{
			"suggestions": types.Suggester{
				Completion: &types.CompletionSuggester{
					Field: "suggest",
					Size:  types.Uint(10),
				},
				Prefix: &query,
			},
		},
	}

	resp, err := r.client.GetClient().Search().
		Index(indexName).
		Request(searchReq).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to execute suggestions search: %w", err)
	}

	var suggestions []entity.SuggestionResult
	if resp.Suggest != nil {
		if suggestionResult, ok := resp.Suggest["suggestions"]; ok {
			for _, item := range suggestionResult {
				for _, option := range item.Options {
					suggestions = append(suggestions, entity.SuggestionResult{
						Text:   option.Text,
						Score:  int(*option.Score),
						Weight: 0, // Weight might not be available in completion suggester
					})
				}
			}
		}
	}

	return suggestions, nil
}

// getIndexNameByType returns index name for document type.
func (r *SearchRepository) getIndexNameByType(documentType string) string {
	switch documentType {
	case "product":
		return "products"
	default:
		return ""
	}
}
