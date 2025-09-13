package constant

import "time"

const (
	// PostgresPort is the port for the PostgreSQL database.
	PostgresPort = 15432
	// PostgresMaxIdleConns is the maximum number of idle connections in the connection pool.
	PostgresMaxIdleConns = 10
	// PostgresMaxOpenConns is the maximum number of open connections to the database.
	PostgresMaxOpenConns = 32
	// PostgresConnMaxLifetime is the maximum lifetime of a connection in the connection pool.
	PostgresConnMaxLifetime = 60 * time.Second
)
