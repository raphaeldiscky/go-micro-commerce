package constant

import "time"

const (
	// ElasticPort is the port for the ElasticSearch server.
	ElasticPort = 9200
	// ElasticMaxRetries is the maximum number of retries for ElasticSearch requests.
	ElasticMaxRetries = 3
	// ElasticMaxIdleConns is the maximum number of idle connections to ElasticSearch.
	ElasticMaxIdleConns = 10 * 1024
	// ElasticMaxIdleTime is the maximum lifetime of an idle connection to ElasticSearch.
	ElasticMaxIdleTime = 30 * time.Second
	// ElasticRequestTimeout is the timeout for ElasticSearch requests.
	ElasticRequestTimeout = 30 * time.Second
	// ElasticDiscoverNodesInterval is the interval at which ElasticSearch nodes are discovered.
	ElasticDiscoverNodesInterval time.Duration = 60 * time.Second
	// ElasticTimeout is the timeout for ElasticSearch requests.
	ElasticTimeout time.Duration = 60 * time.Second
)

const (
	// SortOrderAsc for ascending order.
	SortOrderAsc = "asc"
	// SortOrderDesc for descending order.
	SortOrderDesc = "desc"
)
