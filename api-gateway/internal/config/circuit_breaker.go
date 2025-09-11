package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultCircuitBreakerMaxRequests = 3
	defaultCircuitBreakerInterval    = 10 * time.Second
	defaultCircuitBreakerTimeout     = 30 * time.Second
)

// CircuitBreakerConfig holds CircuitBreaker configuration values.
type CircuitBreakerConfig struct {
	MaxRequests uint32        `mapstructure:"CIRCUIT_BREAKER_MAX_REQUESTS"`
	Interval    time.Duration `mapstructure:"CIRCUIT_BREAKER_INTERVAL"`
	Timeout     time.Duration `mapstructure:"CIRCUIT_BREAKER_TIMEOUT"`
}

// initCircuitBreakerConfig initializes the CircuitBreaker configuration.
func initCircuitBreakerConfig() *CircuitBreakerConfig {
	viper.SetDefault("CIRCUIT_BREAKER_MAX_REQUESTS", defaultCircuitBreakerMaxRequests)
	viper.SetDefault("CIRCUIT_BREAKER_INTERVAL", defaultCircuitBreakerInterval)
	viper.SetDefault("CIRCUIT_BREAKER_TIMEOUT", defaultCircuitBreakerTimeout)

	circuitBreakerConfig := &CircuitBreakerConfig{}
	if err := viper.Unmarshal(circuitBreakerConfig); err != nil {
		panic(err)
	}

	return circuitBreakerConfig
}
