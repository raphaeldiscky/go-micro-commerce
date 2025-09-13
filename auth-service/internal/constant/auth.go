package constant

import "time"

const (
	// AuthVerificationTokenExpiration is the expiration time for user verify account.
	AuthVerificationTokenExpiration time.Duration = 10 * time.Minute
)
