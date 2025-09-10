package handler

import (
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/echoutils"
	"github.com/raphaeldiscky/go-micro-commerce/pkg/utils/pageutils"

	pkgconstant "github.com/raphaeldiscky/go-micro-commerce/pkg/constant"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/dto"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/entity"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/service"
)

// SearchHandler handles HTTP requests for search operations.
type SearchHandler struct {
	searchService service.SearchService
	logger        logger.Logger
}

// NewSearchHandler creates a new search handler.
func NewSearchHandler(searchService service.SearchService, appLogger logger.Logger) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        appLogger,
	}
}

// SearchProducts handles product search requests.
func (h *SearchHandler) SearchProducts(c echo.Context) error {
	searchQuery := h.parseSearchQuery(c)

	results, paging, err := h.searchService.SearchProducts(c.Request().Context(), searchQuery)
	if err != nil {
		h.logger.Errorf("Failed to search products: %v", err)

		return err
	}

	// Add pagination links
	paging.Links = pageutils.NewLinks(
		c.Request(),
		paging.Page,
		paging.Size,
		paging.TotalPage,
	)

	return echoutils.ResponseOKPagination(c, results, paging)
}

// AutoComplete handles autocomplete requests.
func (h *SearchHandler) AutoComplete(c echo.Context) error {
	query := c.QueryParam("q")
	docType := c.QueryParam("type")

	if query == "" {
		return echo.NewHTTPError(400, "Query parameter 'q' is required")
	}

	if docType == "" {
		return echo.NewHTTPError(400, "Query parameter 'type' is required")
	}

	if !isValidDocumentType(docType) {
		return echo.NewHTTPError(400, "Invalid document type. Valid types: product")
	}

	suggestions, err := h.searchService.AutoComplete(c.Request().Context(), query, docType)
	if err != nil {
		h.logger.Errorf("Failed to get autocomplete suggestions: %v", err)

		return err
	}

	return echoutils.ResponseOK(c, suggestions)
}

// GetSuggestions handles enhanced suggestion requests.
func (h *SearchHandler) GetSuggestions(c echo.Context) error {
	query := c.QueryParam("q")
	docType := c.QueryParam("type")

	if query == "" {
		return echo.NewHTTPError(400, "Query parameter 'q' is required")
	}

	if docType == "" {
		return echo.NewHTTPError(400, "Query parameter 'type' is required")
	}

	if !isValidDocumentType(docType) {
		return echo.NewHTTPError(400, "Invalid document type. Valid types: product")
	}

	suggestions, err := h.searchService.GetSuggestions(c.Request().Context(), query, docType)
	if err != nil {
		h.logger.Errorf("Failed to get suggestions: %v", err)

		return err
	}

	return echoutils.ResponseOK(c, suggestions)
}

// IndexProduct handles product indexing requests.
func (h *SearchHandler) IndexProduct(c echo.Context) error {
	var req dto.ProductIndexRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Failed to bind product index request: %v", err)

		return err
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("Product index request validation failed: %v", err)

		return err
	}

	productDoc := req.ToEntity()
	if err := h.searchService.IndexProduct(c.Request().Context(), productDoc); err != nil {
		h.logger.Errorf("Failed to index product: %v", err)

		return err
	}

	responseData := map[string]string{
		"message": "Product indexed successfully",
		"id":      productDoc.ID.String(),
	}

	return echoutils.ResponseCreated(c, responseData)
}

// UpdateProduct handles product update requests.
func (h *SearchHandler) UpdateProduct(c echo.Context) error {
	var req dto.ProductIndexRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Errorf("Failed to bind product update request: %v", err)

		return err
	}

	if err := req.Validate(); err != nil {
		h.logger.Errorf("Product update request validation failed: %v", err)

		return err
	}

	productDoc := req.ToEntity()
	if err := h.searchService.UpdateProduct(c.Request().Context(), productDoc); err != nil {
		h.logger.Errorf("Failed to update product: %v", err)

		return err
	}

	responseData := map[string]string{
		"message": "Product updated successfully",
		"id":      productDoc.ID.String(),
	}

	return echoutils.ResponseOK(c, responseData)
}

// DeleteProduct handles product deletion requests.
func (h *SearchHandler) DeleteProduct(c echo.Context) error {
	productID := c.Param("id")
	if productID == "" {
		return echo.NewHTTPError(400, "Product ID is required")
	}

	if err := h.searchService.DeleteProduct(c.Request().Context(), productID); err != nil {
		h.logger.Errorf("Failed to delete product: %v", err)

		return err
	}

	responseData := map[string]string{
		"message": "Product deleted successfully",
		"id":      productID,
	}

	return echoutils.ResponseOK(c, responseData)
}

// GetProduct handles product retrieval requests.
func (h *SearchHandler) GetProduct(c echo.Context) error {
	productID := c.Param("id")
	if productID == "" {
		return echo.NewHTTPError(400, "Product ID is required")
	}

	product, err := h.searchService.GetProduct(c.Request().Context(), productID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(404, "Product not found")
		}

		h.logger.Errorf("Failed to get product: %v", err)

		return err
	}

	return echoutils.ResponseOK(c, product)
}

// InitializeIndices handles index initialization requests.
func (h *SearchHandler) InitializeIndices(c echo.Context) error {
	if err := h.searchService.InitializeIndices(c.Request().Context()); err != nil {
		h.logger.Errorf("Failed to initialize indices: %v", err)

		return err
	}

	responseData := map[string]string{"message": "Indices initialized successfully"}

	return echoutils.ResponseOK(c, responseData)
}

// RefreshIndices handles index refresh requests.
func (h *SearchHandler) RefreshIndices(c echo.Context) error {
	if err := h.searchService.RefreshIndices(c.Request().Context()); err != nil {
		h.logger.Errorf("Failed to refresh indices: %v", err)

		return err
	}

	responseData := map[string]string{"message": "Indices refreshed successfully"}

	return echoutils.ResponseOK(c, responseData)
}

// Helper functions

// parseSearchQuery parses query parameters into SearchQuery entity.
func (h *SearchHandler) parseSearchQuery(c echo.Context) *entity.SearchQuery {
	// Parse pagination using pageutils
	limit := pageutils.ParseQueryInt64(
		c,
		"limit",
		pkgconstant.DefaultLimit,
		1,
		100,
	) // min=1, max=100
	page := pageutils.ParseQueryInt64(
		c,
		"page",
		pkgconstant.DefaultPage,
		1,
		0,
	) // min=1, max=0 (no max)

	query := &entity.SearchQuery{
		Query:   c.QueryParam("q"),
		Filters: make(map[string]interface{}),
		Sort:    []entity.SortField{},
		From:    int((page - 1) * limit),
		Size:    int(limit),
	}

	// Parse filters based on query parameters
	h.parseFilters(c, query)

	// Parse sorting
	if sortField := c.QueryParam("sort"); sortField != "" {
		order := c.QueryParam("order")
		if order == "" {
			order = "asc"
		}

		if order != "asc" && order != "desc" {
			order = "asc"
		}

		query.Sort = append(query.Sort, entity.SortField{
			Field: sortField,
			Order: order,
		})
	}

	// Add default sorting by relevance score if no sort specified and there's a query
	if len(query.Sort) == 0 && query.Query != "" {
		query.Sort = append(query.Sort, entity.SortField{
			Field: "_score",
			Order: "desc",
		})
	}

	return query
}

// parseFilters parses filter parameters based on query params.
func (h *SearchHandler) parseFilters(c echo.Context, query *entity.SearchQuery) {
	// Common filters
	if category := c.QueryParam("category"); category != "" {
		query.Filters["category"] = category
	}

	if brand := c.QueryParam("brand"); brand != "" {
		query.Filters["brand"] = brand
	}

	if status := c.QueryParam("status"); status != "" {
		query.Filters["status"] = status
	}

	if customerID := c.QueryParam("customer_id"); customerID != "" {
		query.Filters["customer_id"] = customerID
	}

	// Boolean filters
	if inStockStr := c.QueryParam("in_stock"); inStockStr != "" {
		if inStock, err := strconv.ParseBool(inStockStr); err == nil {
			query.Filters["in_stock"] = inStock
		}
	}

	// Range filters (price, amount, etc.)
	h.parseRangeFilters(c, query)
}

// parseRangeFilters parses range filter parameters.
func (h *SearchHandler) parseRangeFilters(c echo.Context, query *entity.SearchQuery) {
	// Price range
	h.parseRangeParam(c, query, "min_price", "price", "gte")
	h.parseRangeParam(c, query, "max_price", "price", "lte")

	// Amount range
	h.parseRangeParam(c, query, "min_amount", "total_amount", "gte")
	h.parseRangeParam(c, query, "max_amount", "total_amount", "lte")
}

// parseRangeParam parses a single range parameter.
func (h *SearchHandler) parseRangeParam(
	c echo.Context,
	query *entity.SearchQuery,
	paramName, filterName, operator string,
) {
	paramValue := c.QueryParam(paramName)
	if paramValue == "" {
		return
	}

	value, err := strconv.ParseFloat(paramValue, 64)
	if err != nil {
		return
	}

	rangeFilter, ok := query.Filters[filterName].(map[string]interface{})
	if !ok {
		rangeFilter = make(map[string]interface{})
		query.Filters[filterName] = rangeFilter
	}

	rangeFilter[operator] = value
}

// isValidDocumentType checks if the document type is valid.
func isValidDocumentType(docType string) bool {
	validTypes := map[string]bool{
		"product": true,
		// "order" and "customer": removed for now, only handling products
	}

	return validTypes[docType]
}
