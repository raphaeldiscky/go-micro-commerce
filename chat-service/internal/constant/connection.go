package constant

import "time"

const (
	// ConnTicketExpiration is the default expiration time for connection tickets.
	ConnTicketExpiration = 10 * time.Minute
	// ConnMaxConnections is the default maximum number of connections.
	ConnMaxConnections = 1000
)
