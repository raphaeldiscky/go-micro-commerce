package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
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
			order := sortorder.Asc
			if sort.Order == "desc" {
				order = sortorder.Desc
			}

			sorts = append(sorts, types.SortOptions{
				SortOptions: map[string]types.FieldSort{
					sort.Field: {Order: &order},
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

	for i := range resp.Hits.Hits {
		hit := &resp.Hits.Hits[i]

		var source map[string]interface{}

		if err := json.Unmarshal(hit.Source_, &source); err != nil {
			r.logger.Warnf("Failed to decode hit source: %v", err)

			continue
		}

		var id string
		if hit.Id_ != nil {
			id = *hit.Id_
		}

		var score float64
		if hit.Score_ != nil {
			score = float64(*hit.Score_)
		}

		searchResult := entity.SearchResult{
			ID:     id,
			Score:  score,
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

// BulkIndex performs bulk indexing using individual index operations.
func (r *SearchRepository) BulkIndex(ctx context.Context, documents []entity.SearchDocument) error {
	if len(documents) == 0 {
		return nil
	}

	// For now, perform individual operations since bulk API structure is complex
	for _, doc := range documents {
		_, err := r.client.GetClient().Index(doc.GetIndexName()).
			Id(doc.GetID()).
			Document(doc).
			Refresh(refresh.True).
			Do(ctx)
		if err != nil {
			r.logger.Warnf("Failed to index document %s: %v", doc.GetID(), err)

			return fmt.Errorf("failed to index document %s: %w", doc.GetID(), err)
		}
	}

	r.logger.Infof("Successfully bulk indexed %d documents", len(documents))

	return nil
}

// CreateIndices creates indices with proper typed mappings.
func (r *SearchRepository) CreateIndices(ctx context.Context) error {
	// Product mapping with typed fields
	standard := "standard"
	english := "english"
	simple := "simple"
	preserveSeparators := true
	preservePositionIncrements := true
	maxInputLength := 50

	productMapping := &types.TypeMapping{
		Properties: map[string]types.Property{
			"id":               types.KeywordProperty{},
			"name":             types.TextProperty{Analyzer: &standard},
			"description":      types.TextProperty{Analyzer: &english},
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
				Analyzer:                   &simple,
				PreserveSeparators:         &preserveSeparators,
				PreservePositionIncrements: &preservePositionIncrements,
				MaxInputLength:             &maxInputLength,
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

// BulkUpdate performs bulk updates using individual update operations.
func (r *SearchRepository) BulkUpdate(
	ctx context.Context,
	documents []entity.SearchDocument,
) error {
	if len(documents) == 0 {
		return nil
	}

	// For now, perform individual operations since bulk API structure is complex
	for _, doc := range documents {
		_, err := r.client.GetClient().Update(
			doc.GetIndexName(),
			doc.GetID(),
		).Doc(doc).Refresh(refresh.True).Do(ctx)
		if err != nil {
			r.logger.Warnf("Failed to update document %s: %v", doc.GetID(), err)

			return fmt.Errorf("failed to update document %s: %w", doc.GetID(), err)
		}
	}

	r.logger.Infof("Successfully bulk updated %d documents", len(documents))

	return nil
}

// BulkDelete performs bulk deletion using individual delete operations.
func (r *SearchRepository) BulkDelete(
	ctx context.Context,
	documentIDs []string,
	indexName string,
) error {
	if len(documentIDs) == 0 {
		return nil
	}

	// For now, perform individual operations since bulk API structure is complex
	for _, docID := range documentIDs {
		_, err := r.client.GetClient().Delete(indexName, docID).
			Refresh(refresh.True).
			Do(ctx)
		if err != nil {
			r.logger.Warnf("Failed to delete document %s: %v", docID, err)

			return fmt.Errorf("failed to delete document %s: %w", docID, err)
		}
	}

	r.logger.Infof("Successfully bulk deleted %d documents", len(documentIDs))

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
func (r *SearchRepository) AutoComplete(
	ctx context.Context,
	query string,
	documentType string,
) ([]string, error) {
	indexName := r.getIndexNameByType(documentType)
	if indexName == "" {
		return nil, fmt.Errorf("unknown document type: %s", documentType)
	}

	// Use prefix query for autocomplete
	size := 10
	searchReq := &search.Request{
		Query: &types.Query{
			Prefix: map[string]types.PrefixQuery{
				"name": {
					Value: query,
				},
			},
		},
		Size:    &size,
		Source_: types.SourceConfigParam([]string{"name"}),
	}

	resp, err := r.client.GetClient().Search().
		Index(indexName).
		Request(searchReq).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute autocomplete search: %w", err)
	}

	var suggestions []string

	for i := range resp.Hits.Hits {
		hit := &resp.Hits.Hits[i]

		var source map[string]interface{}

		if err := json.Unmarshal(hit.Source_, &source); err != nil {
			continue
		}

		if name, ok := source["name"].(string); ok {
			suggestions = append(suggestions, name)
		}
	}

	return suggestions, nil
}

// GetSuggestions provides enhanced suggestions using TypedAPI.
func (r *SearchRepository) GetSuggestions(
	ctx context.Context,
	query string,
	documentType string,
) ([]entity.SuggestionResult, error) {
	indexName := r.getIndexNameByType(documentType)
	if indexName == "" {
		return nil, fmt.Errorf("unknown document type: %s", documentType)
	}

	// Use fuzzy query for suggestions
	size := 10
	searchReq := &search.Request{
		Query: &types.Query{
			Fuzzy: map[string]types.FuzzyQuery{
				"name": {
					Value:     query,
					Fuzziness: types.Fuzziness("AUTO"),
				},
			},
		},
		Size:    &size,
		Source_: types.SourceConfigParam([]string{"name"}),
	}

	resp, err := r.client.GetClient().Search().
		Index(indexName).
		Request(searchReq).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute suggestions search: %w", err)
	}

	var suggestions []entity.SuggestionResult

	for i := range resp.Hits.Hits {
		hit := &resp.Hits.Hits[i]

		var source map[string]interface{}

		if err := json.Unmarshal(hit.Source_, &source); err != nil {
			continue
		}

		if name, ok := source["name"].(string); ok {
			score := 0
			if hit.Score_ != nil {
				score = int(*hit.Score_)
			}

			suggestions = append(suggestions, entity.SuggestionResult{
				Text:   name,
				Score:  score,
				Weight: 0,
			})
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
