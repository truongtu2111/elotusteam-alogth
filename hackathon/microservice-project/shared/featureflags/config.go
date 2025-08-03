package featureflags

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FeatureFlag represents a feature flag configuration
type FeatureFlag struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	Rollout     float64                `json:"rollout" db:"rollout"` // 0.0 to 1.0 (percentage)
	Conditions  map[string]interface{} `json:"conditions" db:"conditions"`
	Environment string                 `json:"environment" db:"environment"`
	Service     string                 `json:"service" db:"service"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
	Tags        []string               `json:"tags" db:"tags"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// UserContext represents user context for feature flag evaluation
type UserContext struct {
	UserID     string                 `json:"user_id"`
	Email      string                 `json:"email"`
	Role       string                 `json:"role"`
	Groups     []string               `json:"groups"`
	Attributes map[string]interface{} `json:"attributes"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Country    string                 `json:"country"`
	Region     string                 `json:"region"`
}

// EvaluationResult represents the result of feature flag evaluation
type EvaluationResult struct {
	FlagID    string      `json:"flag_id"`
	Enabled   bool        `json:"enabled"`
	Variant   string      `json:"variant,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	Reason    string      `json:"reason"`
	Timestamp time.Time   `json:"timestamp"`
}

// FeatureFlagEvent represents an event for feature flag usage tracking
type FeatureFlagEvent struct {
	ID        string        `json:"id"`
	FlagID    string        `json:"flag_id"`
	UserID    string        `json:"user_id"`
	Service   string        `json:"service"`
	EventType string        `json:"event_type"` // "evaluation", "exposure", "conversion"
	Result    bool          `json:"result"`
	Variant   string        `json:"variant,omitempty"`
	Metadata  interface{}   `json:"metadata,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration,omitempty"`
}

// FeatureFlagConfig represents the configuration for the feature flag system
type FeatureFlagConfig struct {
	Enabled           bool          `json:"enabled"`
	RefreshInterval   time.Duration `json:"refresh_interval"`
	CacheEnabled      bool          `json:"cache_enabled"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	AnalyticsEnabled  bool          `json:"analytics_enabled"`
	DefaultVariant    string        `json:"default_variant"`
	Environment       string        `json:"environment"`
	Service           string        `json:"service"`
	StorageType       string        `json:"storage_type"` // "database", "redis", "file", "remote"
	RemoteURL         string        `json:"remote_url,omitempty"`
	RemoteAPIKey      string        `json:"remote_api_key,omitempty"`
	DatabaseURL       string        `json:"database_url,omitempty"`
	RedisURL          string        `json:"redis_url,omitempty"`
	FilePath          string        `json:"file_path,omitempty"`
	MetricsEnabled    bool          `json:"metrics_enabled"`
	DebugMode         bool          `json:"debug_mode"`
}

// FeatureFlagRepository defines the interface for feature flag storage
type FeatureFlagRepository interface {
	GetFlag(ctx context.Context, flagID string) (*FeatureFlag, error)
	GetFlags(ctx context.Context, service string, environment string) ([]*FeatureFlag, error)
	CreateFlag(ctx context.Context, flag *FeatureFlag) error
	UpdateFlag(ctx context.Context, flag *FeatureFlag) error
	DeleteFlag(ctx context.Context, flagID string) error
	GetFlagsByTags(ctx context.Context, tags []string) ([]*FeatureFlag, error)
	GetExpiredFlags(ctx context.Context) ([]*FeatureFlag, error)
}

// FeatureFlagCache defines the interface for feature flag caching
type FeatureFlagCache interface {
	Get(ctx context.Context, key string) (*FeatureFlag, error)
	Set(ctx context.Context, key string, flag *FeatureFlag, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetMultiple(ctx context.Context, keys []string) (map[string]*FeatureFlag, error)
	SetMultiple(ctx context.Context, flags map[string]*FeatureFlag, ttl time.Duration) error
}

// FeatureFlagAnalytics defines the interface for feature flag analytics
type FeatureFlagAnalytics interface {
	TrackEvaluation(ctx context.Context, event *FeatureFlagEvent) error
	TrackExposure(ctx context.Context, event *FeatureFlagEvent) error
	TrackConversion(ctx context.Context, event *FeatureFlagEvent) error
	GetFlagMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetFlagUsage(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error)
}

// FeatureFlagManager defines the main interface for feature flag management
type FeatureFlagManager interface {
	// Flag evaluation
	IsEnabled(ctx context.Context, flagID string, userContext *UserContext) (bool, error)
	GetVariant(ctx context.Context, flagID string, userContext *UserContext) (string, error)
	GetValue(ctx context.Context, flagID string, userContext *UserContext, defaultValue interface{}) (interface{}, error)
	EvaluateFlag(ctx context.Context, flagID string, userContext *UserContext) (*EvaluationResult, error)
	EvaluateAllFlags(ctx context.Context, userContext *UserContext) (map[string]*EvaluationResult, error)

	// Flag management
	CreateFlag(ctx context.Context, flag *FeatureFlag) error
	UpdateFlag(ctx context.Context, flag *FeatureFlag) error
	DeleteFlag(ctx context.Context, flagID string) error
	GetFlag(ctx context.Context, flagID string) (*FeatureFlag, error)
	GetAllFlags(ctx context.Context) ([]*FeatureFlag, error)

	// Cache management
	RefreshCache(ctx context.Context) error
	ClearCache(ctx context.Context) error

	// Analytics
	TrackEvent(ctx context.Context, event *FeatureFlagEvent) error
	GetMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error)

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) error
}

// FeatureFlagEvaluator defines the interface for flag evaluation logic
type FeatureFlagEvaluator interface {
	Evaluate(ctx context.Context, flag *FeatureFlag, userContext *UserContext) (*EvaluationResult, error)
	EvaluateConditions(ctx context.Context, conditions map[string]interface{}, userContext *UserContext) (bool, error)
	EvaluateRollout(ctx context.Context, rollout float64, userContext *UserContext) (bool, error)
}

// FeatureFlagMiddleware defines middleware for HTTP requests
type FeatureFlagMiddleware interface {
	HTTPMiddleware() func(next http.Handler) http.Handler
	GinMiddleware() gin.HandlerFunc
	ExtractUserContext(r *http.Request) (*UserContext, error)
}

// Thread-safe flag storage
type flagStorage struct {
	mu    sync.RWMutex
	flags map[string]*FeatureFlag
}

func newFlagStorage() *flagStorage {
	return &flagStorage{
		flags: make(map[string]*FeatureFlag),
	}
}

func (fs *flagStorage) get(flagID string) (*FeatureFlag, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	flag, exists := fs.flags[flagID]
	return flag, exists
}

func (fs *flagStorage) set(flagID string, flag *FeatureFlag) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.flags[flagID] = flag
}

func (fs *flagStorage) delete(flagID string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	delete(fs.flags, flagID)
}

func (fs *flagStorage) getAll() map[string]*FeatureFlag {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	result := make(map[string]*FeatureFlag)
	for k, v := range fs.flags {
		result[k] = v
	}
	return result
}

func (fs *flagStorage) clear() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.flags = make(map[string]*FeatureFlag)
}

// Helper functions
func GenerateFlagID() string {
	return uuid.New().String()
}

func GenerateEventID() string {
	return uuid.New().String()
}

func ValidateFlagID(flagID string) error {
	if flagID == "" {
		return fmt.Errorf("flag ID cannot be empty")
	}
	if len(flagID) > 255 {
		return fmt.Errorf("flag ID cannot exceed 255 characters")
	}
	return nil
}

func ValidateRollout(rollout float64) error {
	if rollout < 0.0 || rollout > 1.0 {
		return fmt.Errorf("rollout must be between 0.0 and 1.0")
	}
	return nil
}

func ValidateFlag(flag *FeatureFlag) error {
	if err := ValidateFlagID(flag.ID); err != nil {
		return err
	}
	if flag.Name == "" {
		return fmt.Errorf("flag name cannot be empty")
	}
	if err := ValidateRollout(flag.Rollout); err != nil {
		return err
	}
	return nil
}

// Default implementations
var (
	DefaultRefreshInterval = 30 * time.Second
	DefaultCacheTTL        = 5 * time.Minute
	DefaultVariant         = "default"
)

// Error definitions
var (
	ErrFlagNotFound     = fmt.Errorf("feature flag not found")
	ErrInvalidFlagID    = fmt.Errorf("invalid flag ID")
	ErrInvalidRollout   = fmt.Errorf("invalid rollout percentage")
	ErrFlagExpired      = fmt.Errorf("feature flag has expired")
	ErrEvaluationFailed = fmt.Errorf("flag evaluation failed")
	ErrCacheUnavailable = fmt.Errorf("cache is unavailable")
	ErrStorageUnavailable = fmt.Errorf("storage is unavailable")
)