package featureflags

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// PostgreSQLAnalytics implements FeatureFlagAnalytics using PostgreSQL
type PostgreSQLAnalytics struct {
	db *sql.DB
}

// NewPostgreSQLAnalytics creates a new PostgreSQL analytics instance
func NewPostgreSQLAnalytics(db *sql.DB) FeatureFlagAnalytics {
	return &PostgreSQLAnalytics{db: db}
}

// TrackEvaluation tracks a feature flag evaluation event
func (a *PostgreSQLAnalytics) TrackEvaluation(ctx context.Context, event *FeatureFlagEvent) error {
	return a.trackEvent(ctx, event)
}

// TrackExposure tracks a feature flag exposure event
func (a *PostgreSQLAnalytics) TrackExposure(ctx context.Context, event *FeatureFlagEvent) error {
	return a.trackEvent(ctx, event)
}

// TrackConversion tracks a feature flag conversion event
func (a *PostgreSQLAnalytics) TrackConversion(ctx context.Context, event *FeatureFlagEvent) error {
	return a.trackEvent(ctx, event)
}

// trackEvent stores an event in the database
func (a *PostgreSQLAnalytics) trackEvent(ctx context.Context, event *FeatureFlagEvent) error {
	query := `
		INSERT INTO feature_flag_events (
			id, flag_id, user_id, service, event_type, result, variant, metadata, timestamp, duration
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	metadataJSON, _ := json.Marshal(event.Metadata)

	var duration interface{}
	if event.Duration > 0 {
		duration = event.Duration
	}

	_, err := a.db.ExecContext(ctx, query,
		event.ID, event.FlagID, event.UserID, event.Service, event.EventType,
		event.Result, event.Variant, metadataJSON, event.Timestamp, duration,
	)

	if err != nil {
		return fmt.Errorf("failed to track event: %w", err)
	}

	return nil
}

// GetFlagMetrics retrieves metrics for a specific feature flag
func (a *PostgreSQLAnalytics) GetFlagMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Total evaluations
	totalQuery := `
		SELECT COUNT(*) as total_evaluations,
		       COUNT(CASE WHEN result = true THEN 1 END) as enabled_count,
		       COUNT(CASE WHEN result = false THEN 1 END) as disabled_count
		FROM feature_flag_events 
		WHERE flag_id = $1 AND event_type = 'evaluation' 
		      AND timestamp BETWEEN $2 AND $3
	`

	var totalEvaluations, enabledCount, disabledCount int64
	err := a.db.QueryRowContext(ctx, totalQuery, flagID, startDate, endDate).Scan(
		&totalEvaluations, &enabledCount, &disabledCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get total metrics: %w", err)
	}

	metrics["total_evaluations"] = totalEvaluations
	metrics["enabled_count"] = enabledCount
	metrics["disabled_count"] = disabledCount

	if totalEvaluations > 0 {
		metrics["enabled_rate"] = float64(enabledCount) / float64(totalEvaluations)
	} else {
		metrics["enabled_rate"] = 0.0
	}

	// Unique users
	uniqueUsersQuery := `
		SELECT COUNT(DISTINCT user_id) as unique_users
		FROM feature_flag_events 
		WHERE flag_id = $1 AND timestamp BETWEEN $2 AND $3
	`

	var uniqueUsers int64
	err = a.db.QueryRowContext(ctx, uniqueUsersQuery, flagID, startDate, endDate).Scan(&uniqueUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique users: %w", err)
	}
	metrics["unique_users"] = uniqueUsers

	// Variant distribution
	variantQuery := `
		SELECT variant, COUNT(*) as count
		FROM feature_flag_events 
		WHERE flag_id = $1 AND timestamp BETWEEN $2 AND $3 AND variant IS NOT NULL
		GROUP BY variant
	`

	rows, err := a.db.QueryContext(ctx, variantQuery, flagID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get variant distribution: %w", err)
	}
	defer rows.Close()

	variantDistribution := make(map[string]int64)
	for rows.Next() {
		var variant string
		var count int64
		if err := rows.Scan(&variant, &count); err != nil {
			return nil, fmt.Errorf("failed to scan variant: %w", err)
		}
		variantDistribution[variant] = count
	}
	metrics["variant_distribution"] = variantDistribution

	// Daily breakdown
	dailyQuery := `
		SELECT DATE(timestamp) as date, 
		       COUNT(*) as evaluations,
		       COUNT(CASE WHEN result = true THEN 1 END) as enabled
		FROM feature_flag_events 
		WHERE flag_id = $1 AND event_type = 'evaluation' 
		      AND timestamp BETWEEN $2 AND $3
		GROUP BY DATE(timestamp)
		ORDER BY date
	`

	dailyRows, err := a.db.QueryContext(ctx, dailyQuery, flagID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily breakdown: %w", err)
	}
	defer dailyRows.Close()

	dailyBreakdown := make([]map[string]interface{}, 0)
	for dailyRows.Next() {
		var date time.Time
		var evaluations, enabled int64
		if err := dailyRows.Scan(&date, &evaluations, &enabled); err != nil {
			return nil, fmt.Errorf("failed to scan daily data: %w", err)
		}
		dailyBreakdown = append(dailyBreakdown, map[string]interface{}{
			"date":        date.Format("2006-01-02"),
			"evaluations": evaluations,
			"enabled":     enabled,
		})
	}
	metrics["daily_breakdown"] = dailyBreakdown

	return metrics, nil
}

// GetFlagUsage retrieves usage statistics for a specific feature flag
func (a *PostgreSQLAnalytics) GetFlagUsage(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	usage := make(map[string]interface{})

	// Service breakdown
	serviceQuery := `
		SELECT service, COUNT(*) as count
		FROM feature_flag_events 
		WHERE flag_id = $1 AND timestamp BETWEEN $2 AND $3
		GROUP BY service
		ORDER BY count DESC
	`

	rows, err := a.db.QueryContext(ctx, serviceQuery, flagID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get service breakdown: %w", err)
	}
	defer rows.Close()

	serviceBreakdown := make(map[string]int64)
	for rows.Next() {
		var service string
		var count int64
		if err := rows.Scan(&service, &count); err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		serviceBreakdown[service] = count
	}
	usage["service_breakdown"] = serviceBreakdown

	// Event type breakdown
	eventTypeQuery := `
		SELECT event_type, COUNT(*) as count
		FROM feature_flag_events 
		WHERE flag_id = $1 AND timestamp BETWEEN $2 AND $3
		GROUP BY event_type
	`

	eventRows, err := a.db.QueryContext(ctx, eventTypeQuery, flagID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get event type breakdown: %w", err)
	}
	defer eventRows.Close()

	eventTypeBreakdown := make(map[string]int64)
	for eventRows.Next() {
		var eventType string
		var count int64
		if err := eventRows.Scan(&eventType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan event type: %w", err)
		}
		eventTypeBreakdown[eventType] = count
	}
	usage["event_type_breakdown"] = eventTypeBreakdown

	// Top users
	topUsersQuery := `
		SELECT user_id, COUNT(*) as count
		FROM feature_flag_events 
		WHERE flag_id = $1 AND timestamp BETWEEN $2 AND $3 AND user_id IS NOT NULL
		GROUP BY user_id
		ORDER BY count DESC
		LIMIT 10
	`

	topUserRows, err := a.db.QueryContext(ctx, topUsersQuery, flagID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top users: %w", err)
	}
	defer topUserRows.Close()

	topUsers := make([]map[string]interface{}, 0)
	for topUserRows.Next() {
		var userID string
		var count int64
		if err := topUserRows.Scan(&userID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan top user: %w", err)
		}
		topUsers = append(topUsers, map[string]interface{}{
			"user_id": userID,
			"count":   count,
		})
	}
	usage["top_users"] = topUsers

	return usage, nil
}

// InMemoryAnalytics implements FeatureFlagAnalytics using in-memory storage
type InMemoryAnalytics struct {
	events []FeatureFlagEvent
	mu     sync.RWMutex
}

// NewInMemoryAnalytics creates a new in-memory analytics instance
func NewInMemoryAnalytics() FeatureFlagAnalytics {
	return &InMemoryAnalytics{
		events: make([]FeatureFlagEvent, 0),
	}
}

// TrackEvaluation tracks a feature flag evaluation event
func (a *InMemoryAnalytics) TrackEvaluation(ctx context.Context, event *FeatureFlagEvent) error {
	return a.trackEvent(ctx, event)
}

// TrackExposure tracks a feature flag exposure event
func (a *InMemoryAnalytics) TrackExposure(ctx context.Context, event *FeatureFlagEvent) error {
	return a.trackEvent(ctx, event)
}

// TrackConversion tracks a feature flag conversion event
func (a *InMemoryAnalytics) TrackConversion(ctx context.Context, event *FeatureFlagEvent) error {
	return a.trackEvent(ctx, event)
}

// trackEvent stores an event in memory
func (a *InMemoryAnalytics) trackEvent(ctx context.Context, event *FeatureFlagEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Store a copy to prevent external modifications
	eventCopy := *event
	a.events = append(a.events, eventCopy)

	// Keep only recent events to prevent memory issues
	if len(a.events) > 10000 {
		a.events = a.events[1000:] // Remove oldest 1000 events
	}

	return nil
}

// GetFlagMetrics retrieves metrics for a specific feature flag
func (a *InMemoryAnalytics) GetFlagMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	metrics := make(map[string]interface{})
	totalEvaluations := int64(0)
	enabledCount := int64(0)
	disabledCount := int64(0)
	uniqueUsers := make(map[string]bool)
	variantDistribution := make(map[string]int64)
	dailyBreakdown := make(map[string]map[string]int64)

	for _, event := range a.events {
		if event.FlagID == flagID && 
		   event.Timestamp.After(startDate) && 
		   event.Timestamp.Before(endDate) &&
		   event.EventType == "evaluation" {
			
			totalEvaluations++
			if event.Result {
				enabledCount++
			} else {
				disabledCount++
			}

			if event.UserID != "" {
				uniqueUsers[event.UserID] = true
			}

			if event.Variant != "" {
				variantDistribution[event.Variant]++
			}

			dateKey := event.Timestamp.Format("2006-01-02")
			if dailyBreakdown[dateKey] == nil {
				dailyBreakdown[dateKey] = make(map[string]int64)
			}
			dailyBreakdown[dateKey]["evaluations"]++
			if event.Result {
				dailyBreakdown[dateKey]["enabled"]++
			}
		}
	}

	metrics["total_evaluations"] = totalEvaluations
	metrics["enabled_count"] = enabledCount
	metrics["disabled_count"] = disabledCount
	metrics["unique_users"] = int64(len(uniqueUsers))
	metrics["variant_distribution"] = variantDistribution

	if totalEvaluations > 0 {
		metrics["enabled_rate"] = float64(enabledCount) / float64(totalEvaluations)
	} else {
		metrics["enabled_rate"] = 0.0
	}

	// Convert daily breakdown to slice format
	dailySlice := make([]map[string]interface{}, 0)
	for date, data := range dailyBreakdown {
		dailySlice = append(dailySlice, map[string]interface{}{
			"date":        date,
			"evaluations": data["evaluations"],
			"enabled":     data["enabled"],
		})
	}
	metrics["daily_breakdown"] = dailySlice

	return metrics, nil
}

// GetFlagUsage retrieves usage statistics for a specific feature flag
func (a *InMemoryAnalytics) GetFlagUsage(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	usage := make(map[string]interface{})
	serviceBreakdown := make(map[string]int64)
	eventTypeBreakdown := make(map[string]int64)
	userCounts := make(map[string]int64)

	for _, event := range a.events {
		if event.FlagID == flagID && 
		   event.Timestamp.After(startDate) && 
		   event.Timestamp.Before(endDate) {
			
			serviceBreakdown[event.Service]++
			eventTypeBreakdown[event.EventType]++
			if event.UserID != "" {
				userCounts[event.UserID]++
			}
		}
	}

	usage["service_breakdown"] = serviceBreakdown
	usage["event_type_breakdown"] = eventTypeBreakdown

	// Get top 10 users
	topUsers := make([]map[string]interface{}, 0)
	for userID, count := range userCounts {
		topUsers = append(topUsers, map[string]interface{}{
			"user_id": userID,
			"count":   count,
		})
		if len(topUsers) >= 10 {
			break
		}
	}
	usage["top_users"] = topUsers

	return usage, nil
}

// NoOpAnalytics implements FeatureFlagAnalytics with no-op operations
type NoOpAnalytics struct{}

// NewNoOpAnalytics creates a new no-op analytics instance
func NewNoOpAnalytics() FeatureFlagAnalytics {
	return &NoOpAnalytics{}
}

// TrackEvaluation does nothing
func (a *NoOpAnalytics) TrackEvaluation(ctx context.Context, event *FeatureFlagEvent) error {
	return nil
}

// TrackExposure does nothing
func (a *NoOpAnalytics) TrackExposure(ctx context.Context, event *FeatureFlagEvent) error {
	return nil
}

// TrackConversion does nothing
func (a *NoOpAnalytics) TrackConversion(ctx context.Context, event *FeatureFlagEvent) error {
	return nil
}

// GetFlagMetrics returns empty metrics
func (a *NoOpAnalytics) GetFlagMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

// GetFlagUsage returns empty usage
func (a *NoOpAnalytics) GetFlagUsage(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

// AnalyticsFactory creates analytics instances based on configuration
type AnalyticsFactory struct{}

// NewAnalyticsFactory creates a new analytics factory
func NewAnalyticsFactory() *AnalyticsFactory {
	return &AnalyticsFactory{}
}

// CreateAnalytics creates an analytics instance based on the provided configuration
func (f *AnalyticsFactory) CreateAnalytics(config *FeatureFlagConfig, db *sql.DB) FeatureFlagAnalytics {
	if !config.AnalyticsEnabled {
		return NewNoOpAnalytics()
	}

	switch config.StorageType {
	case "database":
		if db != nil {
			return NewPostgreSQLAnalytics(db)
		}
		fallthrough
	case "memory", "":
		return NewInMemoryAnalytics()
	default:
		return NewInMemoryAnalytics()
	}
}

// AsyncAnalytics wraps an analytics implementation with async processing
type AsyncAnalytics struct {
	analytics FeatureFlagAnalytics
	eventCh   chan *FeatureFlagEvent
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// NewAsyncAnalytics creates a new async analytics wrapper
func NewAsyncAnalytics(analytics FeatureFlagAnalytics, bufferSize int) *AsyncAnalytics {
	a := &AsyncAnalytics{
		analytics: analytics,
		eventCh:   make(chan *FeatureFlagEvent, bufferSize),
		stopCh:    make(chan struct{}),
	}

	// Start worker goroutine
	a.wg.Add(1)
	go a.worker()

	return a
}

// TrackEvaluation tracks a feature flag evaluation event asynchronously
func (a *AsyncAnalytics) TrackEvaluation(ctx context.Context, event *FeatureFlagEvent) error {
	select {
	case a.eventCh <- event:
		return nil
	default:
		// Channel is full, log warning and drop event
		log.Printf("Analytics event channel is full, dropping event for flag: %s", event.FlagID)
		return nil
	}
}

// TrackExposure tracks a feature flag exposure event asynchronously
func (a *AsyncAnalytics) TrackExposure(ctx context.Context, event *FeatureFlagEvent) error {
	return a.TrackEvaluation(ctx, event)
}

// TrackConversion tracks a feature flag conversion event asynchronously
func (a *AsyncAnalytics) TrackConversion(ctx context.Context, event *FeatureFlagEvent) error {
	return a.TrackEvaluation(ctx, event)
}

// GetFlagMetrics delegates to the underlying analytics implementation
func (a *AsyncAnalytics) GetFlagMetrics(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return a.analytics.GetFlagMetrics(ctx, flagID, startDate, endDate)
}

// GetFlagUsage delegates to the underlying analytics implementation
func (a *AsyncAnalytics) GetFlagUsage(ctx context.Context, flagID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return a.analytics.GetFlagUsage(ctx, flagID, startDate, endDate)
}

// Stop stops the async analytics processing
func (a *AsyncAnalytics) Stop() {
	close(a.stopCh)
	a.wg.Wait()
}

// worker processes events asynchronously
func (a *AsyncAnalytics) worker() {
	defer a.wg.Done()

	for {
		select {
		case event := <-a.eventCh:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			switch event.EventType {
			case "evaluation":
				a.analytics.TrackEvaluation(ctx, event)
			case "exposure":
				a.analytics.TrackExposure(ctx, event)
			case "conversion":
				a.analytics.TrackConversion(ctx, event)
			default:
				a.analytics.TrackEvaluation(ctx, event)
			}
			cancel()
		case <-a.stopCh:
			// Process remaining events
			for {
				select {
				case event := <-a.eventCh:
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					a.analytics.TrackEvaluation(ctx, event)
					cancel()
				default:
					return
				}
			}
		}
	}
}