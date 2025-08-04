package featureflags

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// InMemoryCache implements FeatureFlagCache using in-memory storage
type InMemoryCache struct {
	data map[string]*cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	flag      *FeatureFlag
	expiresAt time.Time
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() FeatureFlagCache {
	cache := &InMemoryCache{
		data: make(map[string]*cacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a feature flag from cache
func (c *InMemoryCache) Get(ctx context.Context, key string) (*FeatureFlag, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, ErrCacheUnavailable
	}

	// Check if entry has expired
	if time.Now().After(entry.expiresAt) {
		return nil, ErrCacheUnavailable
	}

	// Return a copy to prevent external modifications
	flagCopy := *entry.flag
	return &flagCopy, nil
}

// Set stores a feature flag in cache
func (c *InMemoryCache) Set(ctx context.Context, key string, flag *FeatureFlag, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store a copy to prevent external modifications
	flagCopy := *flag
	c.data[key] = &cacheEntry{
		flag:      &flagCopy,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a feature flag from cache
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Clear removes all feature flags from cache
func (c *InMemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheEntry)
	return nil
}

// GetMultiple retrieves multiple feature flags from cache
func (c *InMemoryCache) GetMultiple(ctx context.Context, keys []string) (map[string]*FeatureFlag, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*FeatureFlag)
	now := time.Now()

	for _, key := range keys {
		entry, exists := c.data[key]
		if exists && now.Before(entry.expiresAt) {
			// Return a copy to prevent external modifications
			flagCopy := *entry.flag
			result[key] = &flagCopy
		}
	}

	return result, nil
}

// SetMultiple stores multiple feature flags in cache
func (c *InMemoryCache) SetMultiple(ctx context.Context, flags map[string]*FeatureFlag, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Now().Add(ttl)
	for key, flag := range flags {
		// Store a copy to prevent external modifications
		flagCopy := *flag
		c.data[key] = &cacheEntry{
			flag:      &flagCopy,
			expiresAt: expiresAt,
		}
	}

	return nil
}

// cleanup removes expired entries from cache
func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// RedisCache implements FeatureFlagCache using Redis
// Note: This is a placeholder implementation. In a real scenario,
// you would use a Redis client library like go-redis
type RedisCache struct {
	// redisClient redis.Client // Would be actual Redis client
	fallback FeatureFlagCache // Fallback to in-memory cache for this example
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(redisURL string) FeatureFlagCache {
	// In a real implementation, you would initialize Redis client here
	// For now, we'll use in-memory cache as fallback
	return &RedisCache{
		fallback: NewInMemoryCache(),
	}
}

// Get retrieves a feature flag from Redis cache
func (c *RedisCache) Get(ctx context.Context, key string) (*FeatureFlag, error) {
	// In a real implementation, you would:
	// 1. Get JSON string from Redis
	// 2. Unmarshal to FeatureFlag
	// For now, fallback to in-memory cache
	return c.fallback.Get(ctx, key)
}

// Set stores a feature flag in Redis cache
func (c *RedisCache) Set(ctx context.Context, key string, flag *FeatureFlag, ttl time.Duration) error {
	// In a real implementation, you would:
	// 1. Marshal FeatureFlag to JSON
	// 2. Set in Redis with TTL
	// For now, fallback to in-memory cache
	return c.fallback.Set(ctx, key, flag, ttl)
}

// Delete removes a feature flag from Redis cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	// In a real implementation, you would delete from Redis
	// For now, fallback to in-memory cache
	return c.fallback.Delete(ctx, key)
}

// Clear removes all feature flags from Redis cache
func (c *RedisCache) Clear(ctx context.Context) error {
	// In a real implementation, you would clear Redis keys with pattern
	// For now, fallback to in-memory cache
	return c.fallback.Clear(ctx)
}

// GetMultiple retrieves multiple feature flags from Redis cache
func (c *RedisCache) GetMultiple(ctx context.Context, keys []string) (map[string]*FeatureFlag, error) {
	// In a real implementation, you would use Redis MGET
	// For now, fallback to in-memory cache
	return c.fallback.GetMultiple(ctx, keys)
}

// SetMultiple stores multiple feature flags in Redis cache
func (c *RedisCache) SetMultiple(ctx context.Context, flags map[string]*FeatureFlag, ttl time.Duration) error {
	// In a real implementation, you would use Redis pipeline
	// For now, fallback to in-memory cache
	return c.fallback.SetMultiple(ctx, flags, ttl)
}

// NoOpCache implements FeatureFlagCache with no-op operations
type NoOpCache struct{}

// NewNoOpCache creates a new no-op cache
func NewNoOpCache() FeatureFlagCache {
	return &NoOpCache{}
}

// Get always returns cache unavailable error
func (c *NoOpCache) Get(ctx context.Context, key string) (*FeatureFlag, error) {
	return nil, ErrCacheUnavailable
}

// Set does nothing
func (c *NoOpCache) Set(ctx context.Context, key string, flag *FeatureFlag, ttl time.Duration) error {
	return nil
}

// Delete does nothing
func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Clear does nothing
func (c *NoOpCache) Clear(ctx context.Context) error {
	return nil
}

// GetMultiple returns empty map
func (c *NoOpCache) GetMultiple(ctx context.Context, keys []string) (map[string]*FeatureFlag, error) {
	return make(map[string]*FeatureFlag), nil
}

// SetMultiple does nothing
func (c *NoOpCache) SetMultiple(ctx context.Context, flags map[string]*FeatureFlag, ttl time.Duration) error {
	return nil
}

// CacheFactory creates cache instances based on configuration
type CacheFactory struct{}

// NewCacheFactory creates a new cache factory
func NewCacheFactory() *CacheFactory {
	return &CacheFactory{}
}

// CreateCache creates a cache instance based on the provided configuration
func (f *CacheFactory) CreateCache(config *FeatureFlagConfig) FeatureFlagCache {
	if !config.CacheEnabled {
		return NewNoOpCache()
	}

	switch config.StorageType {
	case "redis":
		return NewRedisCache(config.RedisURL)
	case "memory", "":
		return NewInMemoryCache()
	default:
		return NewInMemoryCache()
	}
}

// Helper functions for cache key generation

// GenerateCacheKey generates a cache key for a feature flag
func GenerateCacheKey(service, environment, flagID string) string {
	return fmt.Sprintf("ff:%s:%s:%s", service, environment, flagID)
}

// GenerateServiceCacheKey generates a cache key for all flags of a service
func GenerateServiceCacheKey(service, environment string) string {
	return fmt.Sprintf("ff:%s:%s:*", service, environment)
}

// ParseCacheKey parses a cache key to extract service, environment, and flag ID
func ParseCacheKey(key string) (service, environment, flagID string, err error) {
	parts := make([]string, 0)
	current := ""
	for _, char := range key {
		if char == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	parts = append(parts, current)

	if len(parts) != 4 || parts[0] != "ff" {
		return "", "", "", fmt.Errorf("invalid cache key format")
	}

	return parts[1], parts[2], parts[3], nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Size        int       `json:"size"`
	LastUpdated time.Time `json:"last_updated"`
}

// GetStats returns cache statistics (for in-memory cache)
func (c *InMemoryCache) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &CacheStats{
		Size:        len(c.data),
		LastUpdated: time.Now(),
	}
}

// CacheMetrics represents cache metrics for monitoring
type CacheMetrics struct {
	HitRate     float64 `json:"hit_rate"`
	MissRate    float64 `json:"miss_rate"`
	Size        int     `json:"size"`
	MemoryUsage int64   `json:"memory_usage_bytes"`
}

// GetMetrics returns cache metrics
func (c *InMemoryCache) GetMetrics() *CacheMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Calculate approximate memory usage
	memoryUsage := int64(0)
	for _, entry := range c.data {
		// Rough estimation of memory usage
		flagJSON, _ := json.Marshal(entry.flag)
		memoryUsage += int64(len(flagJSON))
	}

	return &CacheMetrics{
		Size:        len(c.data),
		MemoryUsage: memoryUsage,
	}
}
