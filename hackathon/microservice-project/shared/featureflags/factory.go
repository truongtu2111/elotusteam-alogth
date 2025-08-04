package featureflags

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Factory provides a centralized way to create and configure feature flag components
type Factory struct {
	config *FeatureFlagConfig
	db     *sql.DB
}

// NewFactory creates a new feature flag factory
func NewFactory(config *FeatureFlagConfig, db *sql.DB) *Factory {
	return &Factory{
		config: config,
		db:     db,
	}
}

// CreateManager creates a fully configured feature flag manager
func (f *Factory) CreateManager() (FeatureFlagManager, error) {
	// Create repository
	repository, err := f.CreateRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// Create cache
	cache, err := f.CreateCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	// Create analytics
	analytics, err := f.CreateAnalytics()
	if err != nil {
		return nil, fmt.Errorf("failed to create analytics: %w", err)
	}

	// Create manager
	manager := NewFeatureFlagManager(f.config, repository, cache, analytics)

	// Start the manager
	if err := manager.Start(nil); err != nil {
		return nil, fmt.Errorf("failed to start manager: %w", err)
	}

	return manager, nil
}

// CreateRepository creates a repository based on configuration
func (f *Factory) CreateRepository() (FeatureFlagRepository, error) {
	switch f.config.StorageType {
	case "database":
		if f.db == nil {
			return nil, fmt.Errorf("database connection required for database storage")
		}
		return NewPostgreSQLRepository(f.db), nil
	case "memory", "":
		return NewInMemoryRepository(), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", f.config.StorageType)
	}
}

// CreateCache creates a cache based on configuration
func (f *Factory) CreateCache() (FeatureFlagCache, error) {
	factory := NewCacheFactory()
	return factory.CreateCache(f.config), nil
}

// CreateAnalytics creates an analytics instance based on configuration
func (f *Factory) CreateAnalytics() (FeatureFlagAnalytics, error) {
	factory := NewAnalyticsFactory()
	analytics := factory.CreateAnalytics(f.config, f.db)

	// For now, return the analytics directly
	// Async analytics can be added later if needed

	return analytics, nil
}

// CreateMiddleware creates a Gin middleware with the manager
func (f *Factory) CreateMiddleware(manager FeatureFlagManager, config *MiddlewareConfig) *GinMiddleware {
	return NewGinMiddleware(manager, config)
}

// CreateHandler creates an HTTP handler for feature flag management
func (f *Factory) CreateHandler(manager FeatureFlagManager) *FeatureFlagHandler {
	return NewFeatureFlagHandler(manager)
}

// SetupDatabase creates the necessary database tables
func (f *Factory) SetupDatabase() error {
	if f.db == nil {
		return fmt.Errorf("database connection required")
	}

	// For now, we'll assume tables are created externally
	// In a real implementation, you would run the schema.sql file
	return nil
}

// DefaultConfig returns a default feature flag configuration
func DefaultConfig() *FeatureFlagConfig {
	return &FeatureFlagConfig{
		Enabled:          true,
		StorageType:      "memory",
		CacheEnabled:     true,
		CacheTTL:         5 * time.Minute,
		RefreshInterval:  30 * time.Second,
		AnalyticsEnabled: true,
		Environment:      "development",
		Service:          "unknown",
		MetricsEnabled:   true,
		DebugMode:        false,
		DefaultVariant:   "default",
	}
}

// ProductionConfig returns a production-ready feature flag configuration
func ProductionConfig() *FeatureFlagConfig {
	config := DefaultConfig()
	config.StorageType = "database"
	config.Environment = "production"
	config.RefreshInterval = 1 * time.Minute
	config.CacheTTL = 10 * time.Minute
	config.DebugMode = false
	return config
}

// DevelopmentConfig returns a development-friendly feature flag configuration
func DevelopmentConfig() *FeatureFlagConfig {
	config := DefaultConfig()
	config.Environment = "development"
	config.RefreshInterval = 10 * time.Second
	config.CacheTTL = 1 * time.Minute
	config.DebugMode = true
	return config
}

// TestConfig returns a configuration suitable for testing
func TestConfig() *FeatureFlagConfig {
	config := DefaultConfig()
	config.StorageType = "memory"
	config.Environment = "test"
	config.RefreshInterval = 1 * time.Second
	config.CacheTTL = 10 * time.Second
	config.AnalyticsEnabled = false
	config.MetricsEnabled = false
	return config
}

// ConfigFromEnvironment creates a configuration from environment variables
func ConfigFromEnvironment() *FeatureFlagConfig {
	config := DefaultConfig()

	// This would typically read from environment variables
	// For now, we'll return the default config
	// In a real implementation, you would use os.Getenv() to read values

	return config
}

// QuickStart provides a simple way to get started with feature flags
type QuickStart struct {
	Manager    FeatureFlagManager
	Middleware *GinMiddleware
	Handler    *FeatureFlagHandler
	factory    *Factory
}

// NewQuickStart creates a feature flag system with sensible defaults
func NewQuickStart(db *sql.DB, environment string) (*QuickStart, error) {
	var config *FeatureFlagConfig

	switch environment {
	case "production":
		config = ProductionConfig()
	case "development":
		config = DevelopmentConfig()
	case "test":
		config = TestConfig()
	default:
		config = DefaultConfig()
		config.Environment = environment
	}

	factory := NewFactory(config, db)

	// Setup database if using database storage
	if config.StorageType == "database" && db != nil {
		if err := factory.SetupDatabase(); err != nil {
			log.Printf("Warning: Failed to setup database: %v", err)
		}
	}

	// Create manager
	manager, err := factory.CreateManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Create middleware with default config
	middleware := factory.CreateMiddleware(manager, DefaultMiddlewareConfig())

	// Create handler
	handler := factory.CreateHandler(manager)

	return &QuickStart{
		Manager:    manager,
		Middleware: middleware,
		Handler:    handler,
		factory:    factory,
	}, nil
}

// Stop gracefully stops all components
func (qs *QuickStart) Stop() error {
	return qs.Manager.Stop(nil)
}

// CreateSampleFlags creates some sample feature flags for testing
func (qs *QuickStart) CreateSampleFlags() error {
	sampleFlags := []*FeatureFlag{
		{
			ID:          "new-ui",
			Name:        "New UI",
			Description: "Enable the new user interface",
			Enabled:     true,
			Environment: qs.factory.config.Environment,
			Service:     "frontend",
			Tags:        []string{"ui", "frontend"},
			CreatedBy:   "system",
		},
		{
			ID:          "advanced-analytics",
			Name:        "Advanced Analytics",
			Description: "Enable advanced analytics features",
			Enabled:     false,
			Environment: qs.factory.config.Environment,
			Service:     "analytics",
			Tags:        []string{"analytics", "premium"},
			CreatedBy:   "system",
		},
		{
			ID:          "beta-features",
			Name:        "Beta Features",
			Description: "Enable beta features for testing",
			Enabled:     true,
			Environment: qs.factory.config.Environment,
			Service:     "all",
			Tags:        []string{"beta", "experimental"},
			CreatedBy:   "system",
			Conditions: map[string]interface{}{
				"user_attributes": map[string]interface{}{
					"beta_user": true,
				},
			},
		},
		{
			ID:          "performance-mode",
			Name:        "Performance Mode",
			Description: "Enable performance optimizations",
			Enabled:     true,
			Environment: qs.factory.config.Environment,
			Service:     "backend",
			Tags:        []string{"performance", "optimization"},
			CreatedBy:   "system",
		},
	}

	for _, flag := range sampleFlags {
		if err := qs.Manager.CreateFlag(nil, flag); err != nil {
			log.Printf("Warning: Failed to create sample flag %s: %v", flag.ID, err)
		}
	}

	return nil
}

// Helper functions for common use cases

// SimpleFeatureFlag creates a simple on/off feature flag
func SimpleFeatureFlag(id, name, description string, enabled bool, rollout float64, environment, service string) *FeatureFlag {
	return &FeatureFlag{
		ID:          id,
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Rollout:     rollout,
		Environment: environment,
		Service:     service,
		CreatedBy:   "system",
	}
}

// ABTestFlag creates a feature flag for A/B testing with metadata
func ABTestFlag(id, name, description string, enabled bool, variants map[string]interface{}, environment, service string) *FeatureFlag {
	return &FeatureFlag{
		ID:          id,
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Rollout:     1.0, // A/B tests typically target all users
		Metadata:    map[string]interface{}{"variants": variants},
		Environment: environment,
		Service:     service,
		CreatedBy:   "system",
	}
}

// ConfigFlag creates a feature flag that returns configuration values
func ConfigFlag(id, name, description string, enabled bool, value interface{}, environment, service string) *FeatureFlag {
	return &FeatureFlag{
		ID:          id,
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Rollout:     1.0,
		Metadata:    map[string]interface{}{"value": value},
		Environment: environment,
		Service:     service,
		CreatedBy:   "system",
	}
}

// ConditionalFlag creates a feature flag with user targeting conditions
func ConditionalFlag(id, name, description string, enabled bool, rollout float64, conditions map[string]interface{}, environment, service string) *FeatureFlag {
	return &FeatureFlag{
		ID:          id,
		Name:        name,
		Description: description,
		Enabled:     enabled,
		Rollout:     rollout,
		Conditions:  conditions,
		Environment: environment,
		Service:     service,
		CreatedBy:   "system",
	}
}
