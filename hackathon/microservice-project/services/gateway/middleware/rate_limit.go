package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/elotusteam/microservice-project/shared/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter holds the rate limiter for each IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for the given IP
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// CleanupOldEntries removes old entries from the limiters map
func (rl *RateLimiter) CleanupOldEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Clear all limiters periodically to prevent memory leaks
	// In a production environment, you might want to implement a more sophisticated cleanup
	if len(rl.limiters) > 1000 {
		rl.limiters = make(map[string]*rate.Limiter)
	}
}

var globalRateLimiter *RateLimiter
var once sync.Once

// RateLimit middleware implements rate limiting per IP address
func RateLimit(cfg *config.Config) gin.HandlerFunc {
	once.Do(func() {
		// Default rate limit: 100 requests per minute
		rateLimit := rate.Limit(100.0 / 60.0) // requests per second
		burst := 10

		if cfg.RateLimit.PerIP.Requests > 0 {
			// Convert requests per window to requests per second
			windowSeconds := cfg.RateLimit.PerIP.Window.Seconds()
			if windowSeconds > 0 {
				rateLimit = rate.Limit(float64(cfg.RateLimit.PerIP.Requests) / windowSeconds)
			}
		}
		if cfg.RateLimit.PerIP.Burst > 0 {
			burst = cfg.RateLimit.PerIP.Burst
		}

		globalRateLimiter = NewRateLimiter(rateLimit, burst)

		// Start cleanup goroutine
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				globalRateLimiter.CleanupOldEntries()
			}
		}()
	})

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get rate limiter for this IP
		limiter := globalRateLimiter.GetLimiter(clientIP)

		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
