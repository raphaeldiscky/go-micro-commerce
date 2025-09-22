package redis

const (
	// DefaultRetryAttempts is the default number of retry attempts for failed operations.
	DefaultRetryAttempts = 3

	// DefaultRetryDelayMs is the default initial delay between retries in milliseconds.
	DefaultRetryDelayMs = 100

	// DefaultMaxRetryDelaySec is the default maximum delay between retries in seconds.
	DefaultMaxRetryDelaySec = 5

	// DefaultChannelBufferSize is the default buffer size for subscription channels.
	DefaultChannelBufferSize = 100

	// DefaultPoolSize is the default maximum number of connections in the pool.
	DefaultPoolSize = 10

	// DefaultMinIdleConns is the default minimum number of idle connections.
	DefaultMinIdleConns = 2

	// DefaultMaxIdleConns is the default maximum number of idle connections.
	DefaultMaxIdleConns = 5

	// DefaultConnMaxIdleTimeMin is the default maximum amount of time a connection may be idle in minutes.
	DefaultConnMaxIdleTimeMin = 30

	// DefaultConnMaxLifetimeHour is the default maximum amount of time a connection may be reused in hours.
	DefaultConnMaxLifetimeHour = 1

	// DefaultDialTimeoutSec is the default timeout for establishing new connections in seconds.
	DefaultDialTimeoutSec = 5

	// DefaultReadTimeoutSec is the default timeout for socket reads in seconds.
	DefaultReadTimeoutSec = 3

	// DefaultWriteTimeoutSec is the default timeout for socket writes in seconds.
	DefaultWriteTimeoutSec = 3
)
