package featureflags

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"strconv"
	"sync"
	"time"
)

// featureFlagManager implements the FeatureFlagManager interface
type featureFlagManager struct {
	config     *FeatureFlagConfig
	repository FeatureFlagRepository
	cache      FeatureFlagCache
	analytics  FeatureFlagAnalytics
	evaluator  FeatureFlagEvaluator
	storage    *flagStorage
	mu         sync.RWMutex
	running    bool
	stopCh     chan struct{}
}

// NewFeatureFlagManager creates a new feature flag manager
func NewFeatureFlagManager(config *FeatureFlagConfig, repository FeatureFlagRepository, cache FeatureFlagCache, analytics FeatureFlagAnalytics) FeatureFlagManager {
	return &featureFlagManager{
		config:     config,
		repository: repository,
		cache:      cache,
		analytics:  analytics,
		evaluator:  NewDefaultEvaluator(),
		storage:    newFlagStorage(),
		stopCh:     make(chan struct{}),
	}
}

// Start initializes and starts the feature flag manager
func (m *featureFlagManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("feature flag manager is already running")
	}

	// Load initial flags
	if err := m.loadFlags(ctx); err != nil {
		return fmt.Errorf("failed to load initial flags: %w", err)
	}

	m.running = true

	// Start background refresh if enabled
	if m.config.RefreshInterval > 0 {
		go m.refreshLoop()
	}

	log.Printf("Feature flag manager started for service: %s, environment: %s", m.config.Service, m.config.Environment)
	return nil
}

// Stop gracefully stops the feature flag manager
func (m *featureFlagManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.running = false
	close(m.stopCh)

	log.Printf("Feature flag manager stopped")
	return nil
}

// HealthCheck performs a health check on the feature flag manager
func (m *featureFlagManager) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()

	if !running {
		return fmt.Errorf("feature flag manager is not running")
	}

	// Test repository connection
	if m.repository != nil {
		_, err := m.repository.GetFlags(ctx, m.config.Service, m.config.Environment)
		if err != nil {
			return fmt.Errorf("repository health check failed: %w", err)
		}
	}

	// Test cache connection
	if m.cache != nil && m.config.CacheEnabled {
		testKey := "health_check_" + strconv.FormatInt(time.Now().Unix(), 10)
		testFlag := &FeatureFlag{ID: testKey, Name: "health_check", Enabled: true}
		if err := m.cache.Set(ctx, testKey, testFlag, time.Second); err != nil {
			return fmt.Errorf("cache health check failed: %w", err)
		}
		m.cache.Delete(ctx, testKey)
	}

	return nil
}

// IsEnabled checks if a feature flag is enabled for the given user context
func (m *featureFlagManager) IsEnabled(ctx context.Context, flagID string, userContext *UserContext) (bool, error) {
	result, err := m.EvaluateFlag(ctx, flagID, userContext)
	if err != nil {
		return false, err
	}
	return result.Enabled, nil
}

// GetVariant returns the variant for a feature flag
func (m *featureFlagManager) GetVariant(ctx context.Context, flagID string, userContext *UserContext) (string, error) {
	result, err := m.EvaluateFlag(ctx, flagID, userContext)
	if err != nil {
		return m.config.DefaultVariant, err
	}
	if result.Variant != "" {
		return result.Variant, nil
	}
	return m.config.DefaultVariant, nil
}

// GetValue returns the value for a feature flag
func (m *featureFlagManager) GetValue(ctx context.Context, flagID string, userContext *UserContext, defaultValue interface{}) (interface{}, error) {
	result, err := m.EvaluateFlag(ctx, flagID, userContext)
	if err != nil {
		return defaultValue, err
	}
	if result.Value != nil {
		return result.Value, nil
	}
	return defaultValue, nil
}

// EvaluateFlag evaluates a single feature flag
func (m *featureFlagManager) EvaluateFlag(ctx context.Context, flagID string, userContext *UserContext) (*EvaluationResult, error) {
	if err := ValidateFlagID(flagID); err != nil {
		return nil, err
	}

	// Get flag from storage/cache
	flag, err := m.getFlag(ctx, flagID)
	if err != nil {
		return &EvaluationResult{
			FlagID:    flagID,
			Enabled:   false,
			Reason:    fmt.Sprintf("flag not found: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	// Check if flag is expired
	if flag.ExpiresAt != nil && time.Now().After(*flag.ExpiresAt) {
		return &EvaluationResult{
			FlagID:    flagID,
			Enabled:   false,
			Reason:    "flag expired",
			Timestamp: time.Now(),
		}, nil
	}

	// Evaluate using the evaluator
	result, err := m.evaluator.Evaluate(ctx, flag, userContext)
	if err != nil {
		return &EvaluationResult{
			FlagID:    flagID,
			Enabled:   false,
			Reason:    fmt.Sprintf("evaluation failed: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	// Track evaluation event
	if m.analytics != nil && m.config.AnalyticsEnabled {
		event := &FeatureFlagEvent{
			ID:        GenerateEventID(),
			FlagID:    flagID,
			UserID:    userContext.UserID,
			Service:   m.config.Service,
			EventType: "evaluation",
			Result:    result.Enabled,
			Variant:   result.Variant,
			Timestamp: time.Now(),
		}
		go m.analytics.TrackEvaluation(ctx, event)
	}

	return result, nil
}

// EvaluateAllFlags evaluates all flags for the given user context
func (m *featureFlagManager) EvaluateAllFlags(ctx context.Context, userContext *UserContext) (map[string]*EvaluationResult, error) {
	flags := m.storage.getAll()
	results := make(map[string]*EvaluationResult)

	for flagID := range flags {
		result, err := m.EvaluateFlag(ctx, flagID, userContext)
		if err != nil {
			// Log error but continue with other flags
			log.Printf("Error evaluating flag %s: %v", flagID, err)
			continue
		}
		results[flagID] = result
	}

	return results, nil
}

// CreateFlag creates a new feature flag
func (m *featureFlagManager) CreateFlag(ctx context.Context, flag *FeatureFlag) error {
	if err := ValidateFlag(flag); err != nil {
		return err
	}

	// Set timestamps
	now := time.Now()
	flag.CreatedAt = now
	flag.UpdatedAt = now

	// Set service and environment if not provided
	if flag.Service == "" {
		flag.Service = m.config.Service
	}
	if flag.Environment == "" {
		flag.Environment = m.config.Environment
	}

	// Save to repository
	if m.repository != nil {
		if err := m.repository.CreateFlag(ctx, flag); err != nil {
			return err
		}
	}

	// Update local storage
	m.storage.set(flag.ID, flag)

	// Update cache
	if m.cache != nil && m.config.CacheEnabled {
		m.cache.Set(ctx, m.getCacheKey(flag.ID), flag, m.config.CacheTTL)
	}

	return nil
}

// UpdateFlag updates an existing feature flag
func (m *featureFlagManager) UpdateFlag(ctx context.Context, flag *FeatureFlag) error {
	if err := ValidateFlag(flag); err != nil {
		return err
	}

	// Set update timestamp
	flag.UpdatedAt = time.Now()

	// Save to repository
	if m.repository != nil {
		if err := m.repository.UpdateFlag(ctx, flag); err != nil {
			return err
		}
	}

	// Update local storage
	m.storage.set(flag.ID, flag)

	// Update cache
	if m.cache != nil && m.config.CacheEnabled {
		m.cache.Set(ctx, m.getCacheKey(flag.ID), flag, m.config.CacheTTL)
	}

	return nil
}

// DeleteFlag deletes a feature flag
func (m *featureFlagManager) DeleteFlag(ctx context.Context, flagID string) error {
	if err := ValidateFlagID(flagID); err != nil {
		return err
	}

	// Delete from repository
	if m.repository != nil {
		if err := m.repository.DeleteFlag(ctx, flagID); err != nil {
			return err
		}
	}

	// Delete from local storage
	m.storage.delete(flagID)

	// Delete from cache
	if m.cache != nil && m.config.CacheEnabled {
		m.cache.Delete(ctx, m.getCacheKey(flagID))
	}

	return nil
}

// GetFlag retrieves a single feature flag
func (m *featureFlagManager) GetFlag(ctx context.Context, flagID string) (*FeatureFlag, error) {
	return m.getFlag(ctx, flagID)
}

// GetAllFlags retrieves all feature flags
func (m *featureFlagManager) GetAllFlags(ctx context.Context) ([]*FeatureFlag, error) {
	flags := m.storage.getAll()
	result := make([]*FeatureFlag, 0, len(flags))
	for _, flag := range flags {
		result = append(result, flag)
	}
	return result, nil
}

// RefreshCache refreshes the flag cache from the repository
func (m *featureFlagManager) RefreshCache(ctx context.Context) error {
	return m.loadFlags(ctx)
}

// ClearCache clears the flag cache
func (m *featureFlagManager) ClearCache(ctx context.Context) error {
	m.storage.clear()
	if m.cache != nil && m.config.CacheEnabled {
		return m.cache.Clear(ctx)
	}
	return nil
}

// TrackEvent tracks a feature flag event
func (m *featureFlagManager) TrackEvent(ctx context.Context, event *FeatureFlagEvent) error {
	if m.analytics == nil || !m.config.AnalyticsEnabled {
		return nil
	}

	switch event.EventType {
	case "evaluation":
		return m.analytics.TrackEvaluation(ctx, event)
	case "exposure":
		return m.analytics.TrackExposure(ctx, event)
	case "conversion":
		return m.analytics.TrackConversion(ctx, event)
	default:
		return m.analytics.TrackEvaluation(ctx, event)
	}
}

// GetMetrics retrieves metrics for a feature flag
func (m *featureFlagManager) GetMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	if m.analytics == nil {
		return nil, fmt.Errorf("analytics not configured")
	}
	return m.analytics.GetFlagMetrics(ctx, flagID, startDate, endDate)
}

// Private helper methods

func (m *featureFlagManager) getFlag(ctx context.Context, flagID string) (*FeatureFlag, error) {
	// Try local storage first
	if flag, exists := m.storage.get(flagID); exists {
		return flag, nil
	}

	// Try cache if enabled
	if m.cache != nil && m.config.CacheEnabled {
		if flag, err := m.cache.Get(ctx, m.getCacheKey(flagID)); err == nil {
			m.storage.set(flagID, flag)
			return flag, nil
		}
	}

	// Try repository
	if m.repository != nil {
		if flag, err := m.repository.GetFlag(ctx, flagID); err == nil {
			m.storage.set(flagID, flag)
			if m.cache != nil && m.config.CacheEnabled {
				m.cache.Set(ctx, m.getCacheKey(flagID), flag, m.config.CacheTTL)
			}
			return flag, nil
		}
	}

	return nil, ErrFlagNotFound
}

func (m *featureFlagManager) loadFlags(ctx context.Context) error {
	if m.repository == nil {
		return nil
	}

	flags, err := m.repository.GetFlags(ctx, m.config.Service, m.config.Environment)
	if err != nil {
		return err
	}

	// Clear existing flags
	m.storage.clear()

	// Load new flags
	for _, flag := range flags {
		m.storage.set(flag.ID, flag)
		// Update cache if enabled
		if m.cache != nil && m.config.CacheEnabled {
			m.cache.Set(ctx, m.getCacheKey(flag.ID), flag, m.config.CacheTTL)
		}
	}

	log.Printf("Loaded %d feature flags for service: %s, environment: %s", len(flags), m.config.Service, m.config.Environment)
	return nil
}

func (m *featureFlagManager) refreshLoop() {
	ticker := time.NewTicker(m.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := m.loadFlags(ctx); err != nil {
				log.Printf("Failed to refresh flags: %v", err)
			}
			cancel()
		case <-m.stopCh:
			return
		}
	}
}

func (m *featureFlagManager) getCacheKey(flagID string) string {
	return fmt.Sprintf("ff:%s:%s:%s", m.config.Service, m.config.Environment, flagID)
}

// DefaultEvaluator implements basic feature flag evaluation logic
type defaultEvaluator struct{}

// NewDefaultEvaluator creates a new default evaluator
func NewDefaultEvaluator() FeatureFlagEvaluator {
	return &defaultEvaluator{}
}

// Evaluate evaluates a feature flag against user context
func (e *defaultEvaluator) Evaluate(ctx context.Context, flag *FeatureFlag, userContext *UserContext) (*EvaluationResult, error) {
	result := &EvaluationResult{
		FlagID:    flag.ID,
		Enabled:   false,
		Variant:   DefaultVariant,
		Timestamp: time.Now(),
	}

	// Check if flag is globally disabled
	if !flag.Enabled {
		result.Reason = "flag disabled"
		return result, nil
	}

	// Evaluate conditions if present
	if len(flag.Conditions) > 0 {
		conditionsMet, err := e.EvaluateConditions(ctx, flag.Conditions, userContext)
		if err != nil {
			result.Reason = fmt.Sprintf("condition evaluation failed: %v", err)
			return result, nil
		}
		if !conditionsMet {
			result.Reason = "conditions not met"
			return result, nil
		}
	}

	// Evaluate rollout percentage
	rolloutPassed, err := e.EvaluateRollout(ctx, flag.Rollout, userContext)
	if err != nil {
		result.Reason = fmt.Sprintf("rollout evaluation failed: %v", err)
		return result, nil
	}

	if !rolloutPassed {
		result.Reason = "not in rollout"
		return result, nil
	}

	// Flag is enabled
	result.Enabled = true
	result.Reason = "enabled"

	// Set variant if specified in metadata
	if variant, ok := flag.Metadata["variant"].(string); ok {
		result.Variant = variant
	}

	// Set value if specified in metadata
	if value, ok := flag.Metadata["value"]; ok {
		result.Value = value
	}

	return result, nil
}

// EvaluateConditions evaluates flag conditions against user context
func (e *defaultEvaluator) EvaluateConditions(ctx context.Context, conditions map[string]interface{}, userContext *UserContext) (bool, error) {
	for key, value := range conditions {
		switch key {
		case "user_id":
			if !e.evaluateStringCondition(userContext.UserID, value) {
				return false, nil
			}
		case "email":
			if !e.evaluateStringCondition(userContext.Email, value) {
				return false, nil
			}
		case "role":
			if !e.evaluateStringCondition(userContext.Role, value) {
				return false, nil
			}
		case "groups":
			if !e.evaluateArrayCondition(userContext.Groups, value) {
				return false, nil
			}
		case "country":
			if !e.evaluateStringCondition(userContext.Country, value) {
				return false, nil
			}
		case "region":
			if !e.evaluateStringCondition(userContext.Region, value) {
				return false, nil
			}
		default:
			// Check custom attributes
			if userContext.Attributes != nil {
				if attrValue, exists := userContext.Attributes[key]; exists {
					if !e.evaluateGenericCondition(attrValue, value) {
						return false, nil
					}
				} else {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
	}
	return true, nil
}

// EvaluateRollout evaluates rollout percentage using consistent hashing
func (e *defaultEvaluator) EvaluateRollout(ctx context.Context, rollout float64, userContext *UserContext) (bool, error) {
	if rollout >= 1.0 {
		return true, nil
	}
	if rollout <= 0.0 {
		return false, nil
	}

	// Use consistent hashing based on user ID
	hash := e.hashUserID(userContext.UserID)
	threshold := uint32(rollout * math.MaxUint32)
	return hash <= threshold, nil
}

// Helper methods for condition evaluation

func (e *defaultEvaluator) evaluateStringCondition(userValue string, condition interface{}) bool {
	switch v := condition.(type) {
	case string:
		return userValue == v
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok && userValue == str {
				return true
			}
		}
		return false
	case map[string]interface{}:
		// Support for operators like {"$in": ["value1", "value2"]}
		if inValues, ok := v["$in"].([]interface{}); ok {
			for _, item := range inValues {
				if str, ok := item.(string); ok && userValue == str {
					return true
				}
			}
		}
		if notInValues, ok := v["$nin"].([]interface{}); ok {
			for _, item := range notInValues {
				if str, ok := item.(string); ok && userValue == str {
					return false
				}
			}
			return true
		}
		return false
	default:
		return false
	}
}

func (e *defaultEvaluator) evaluateArrayCondition(userValues []string, condition interface{}) bool {
	switch v := condition.(type) {
	case string:
		for _, userValue := range userValues {
			if userValue == v {
				return true
			}
		}
		return false
	case []interface{}:
		for _, userValue := range userValues {
			for _, item := range v {
				if str, ok := item.(string); ok && userValue == str {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

func (e *defaultEvaluator) evaluateGenericCondition(userValue interface{}, condition interface{}) bool {
	// Simple equality check for now
	return userValue == condition
}

func (e *defaultEvaluator) hashUserID(userID string) uint32 {
	if userID == "" {
		return 0
	}

	// Use FNV-1a hash for consistent distribution
	h := fnv.New32a()
	h.Write([]byte(userID))
	return h.Sum32()
}

// MD5 hash alternative (for compatibility)
func (e *defaultEvaluator) md5Hash(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}
