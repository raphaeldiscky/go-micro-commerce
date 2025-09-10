package provider

import (
	"github.com/raphaeldiscky/go-micro-commerce/pkg/logger"

	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/config"
	"github.com/raphaeldiscky/go-micro-commerce/search-service/internal/worker"
)

// SetupInboxProcessor initializes the inbox processor service for search operations.
func SetupInboxProcessor(
	cfg *config.Config,
	appLogger logger.Logger,
	providers *Providers,
) *worker.InboxProcessor {
	// Create inbox processor with search service
	inboxProcessor := worker.NewInboxProcessor(
		providers.DataStore,
		appLogger,
		providers.SearchService,
		*cfg.InboxProcessor,
	)

	return inboxProcessor
}
