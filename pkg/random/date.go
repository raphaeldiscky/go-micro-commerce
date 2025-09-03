package random

import (
	"crypto/rand"
	"math/big"
	"time"
)

// Duration returns a random time.Duration between 0 and num.
func Duration(num time.Duration) time.Duration {
	if num <= 0 {
		return 0
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(num)))
	if err != nil {
		return 0
	}

	return time.Duration(n.Int64())
}
