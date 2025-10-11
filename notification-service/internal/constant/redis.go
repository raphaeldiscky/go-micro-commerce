package constant

import "time"

const (
	// RedisDialTimeout is the timeout for Redis client connections.
	RedisDialTimeout = 5 * time.Second
	// RedisReadTimeout is the read timeout for Redis client connections.
	RedisReadTimeout = 3 * time.Second
	// RedisWriteTimeout is the write timeout for Redis client connections.
	RedisWriteTimeout = 3 * time.Second
	// RedisMinIdleConn is the minimum number of idle connections in the connection pool.
	RedisMinIdleConn = 8
	// RedisMaxIdleConn is the maximum number of idle connections in the connection pool.
	RedisMaxIdleConn = 12
	// RedisMaxActiveConn is the maximum number of open connections to the database.
	RedisMaxActiveConn = 32
	// RedisMaxConnLifetime is the maximum lifetime of a connection in the connection pool.
	RedisMaxConnLifetime = 1 * time.Minute
)
