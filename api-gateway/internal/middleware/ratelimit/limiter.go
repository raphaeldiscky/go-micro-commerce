// Package ratelimit implements rate limiting middleware for Echo framework.
package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	"github.com/raphaeldiscky/go-micro-commerce/api-gateway/internal/config"
)

// RateLimiter manages rate limiting for different clients.
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(cfg config.RateLimitConfig) *RateLimiter {
	requests := float64(cfg.Requests)
	window := cfg.Window.Seconds()
	rateLimit := rate.Limit(requests / window)

	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rateLimit,
		burst:    cfg.BurstLimit,
	}
}

// GetLimiter returns a rate limiter for the given key.
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	rl.mutex.RLock()
	limiter, exists := rl.limiters[key]
	rl.mutex.RUnlock()

	if exists {
		return limiter
	}

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if limiter, exists = rl.limiters[key]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[key] = limiter

	return limiter
}

// Allow checks if a request is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	limiter := rl.GetLimiter(key)

	return limiter.Allow()
}

// CleanupExpired removes expired limiters (call periodically).
func (rl *RateLimiter) CleanupExpired() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Simple cleanup - in production, you might want more sophisticated cleanup
	// based on last access time
	for key := range rl.limiters {
		delete(rl.limiters, key)
	}
}

// Middleware creates rate limiting middleware.
func Middleware(cfg config.RateLimitConfig) echo.MiddlewareFunc {
	if !cfg.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	rateLimiter := NewRateLimiter(cfg)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			rateLimiter.CleanupExpired()
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Use IP address as the key (you might want to use user ID for authenticated requests)
			key := c.RealIP()

			if !rateLimiter.Allow(key) {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}

			return next(c)
		}
	}
}

// UserBasedMiddleware creates rate limiting middleware based on user ID.
func UserBasedMiddleware(cfg config.RateLimitConfig) echo.MiddlewareFunc {
	if !cfg.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	rateLimiter := NewRateLimiter(cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Use user ID if available, fallback to IP
			key := c.RealIP()

			if userID := c.Get("user_id"); userID != nil {
				if id, ok := userID.(string); ok {
					key = "user:" + id
				}
			}

			if !rateLimiter.Allow(key) {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}

			return next(c)
		}
	}
}

// ServiceBasedMiddleware creates rate limiting middleware based on service.
func ServiceBasedMiddleware(cfg config.RateLimitConfig, serviceName string) echo.MiddlewareFunc {
	if !cfg.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	rateLimiter := NewRateLimiter(cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := serviceName + ":" + c.RealIP()

			if !rateLimiter.Allow(key) {
				return echo.NewHTTPError(http.StatusTooManyRequests,
					map[string]string{
						"error":   "rate limit exceeded",
						"service": serviceName,
					})
			}

			return next(c)
		}
	}
}
