// Package random provides functions for generating random numbers, strings, etc.
package random

import (
	"crypto/rand"
	"math/big"
)

// Int generates a secure random integer between 0 and maxNum (exclusive).
// Returns 0 if maxNum <= 0 or if crypto/rand fails.
func Int(maxNum int64) int {
	if maxNum <= 0 {
		return 0
	}

	n, err := rand.Int(rand.Reader, big.NewInt(maxNum))
	if err != nil {
		return 0
	}

	return int(n.Int64())
}

// IntRange returns a random int64 between minNum and maxNum.
func IntRange(minNum, maxNum int64) int64 {
	if maxNum <= minNum {
		return minNum
	}

	n, err := rand.Int(rand.Reader, big.NewInt(maxNum-minNum))
	if err != nil {
		return minNum
	}

	return minNum + n.Int64()
}
