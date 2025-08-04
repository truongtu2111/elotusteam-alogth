package featureflags

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// PostgreSQLRepository implements FeatureFlagRepository for PostgreSQL
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a new PostgreSQL repository
func NewPostgreSQLRepository(db *sql.DB) FeatureFlagRepository {
	return &PostgreSQLRepository{db: db}
}

// GetFlag retrieves a single feature flag by ID
func (r *PostgreSQLRepository) GetFlag(ctx context.Context, flagID string) (*FeatureFlag, error) {
	query := `
		SELECT id, name, description, enabled, rollout, conditions, environment, 
		       service, created_by, created_at, updated_at, expires_at, tags, metadata
		FROM feature_flags 
		WHERE id = $1
	`

	var flag FeatureFlag
	var conditionsJSON, tagsJSON, metadataJSON []byte
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, flagID).Scan(
		&flag.ID, &flag.Name, &flag.Description, &flag.Enabled, &flag.Rollout,
		&conditionsJSON, &flag.Environment, &flag.Service, &flag.CreatedBy,
		&flag.CreatedAt, &flag.UpdatedAt, &expiresAt, &tagsJSON, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFlagNotFound
		}
		return nil, fmt.Errorf("failed to get flag: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal(conditionsJSON, &flag.Conditions); err != nil {
		flag.Conditions = make(map[string]interface{})
	}
	if err := json.Unmarshal(tagsJSON, &flag.Tags); err != nil {
		flag.Tags = []string{}
	}
	if err := json.Unmarshal(metadataJSON, &flag.Metadata); err != nil {
		flag.Metadata = make(map[string]interface{})
	}

	if expiresAt.Valid {
		flag.ExpiresAt = &expiresAt.Time
	}

	return &flag, nil
}

// GetFlags retrieves all feature flags for a service and environment
func (r *PostgreSQLRepository) GetFlags(ctx context.Context, service string, environment string) ([]*FeatureFlag, error) {
	query := `
		SELECT id, name, description, enabled, rollout, conditions, environment, 
		       service, created_by, created_at, updated_at, expires_at, tags, metadata
		FROM feature_flags 
		WHERE service = $1 AND environment = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, service, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to query flags: %w", err)
	}
	defer rows.Close()

	var flags []*FeatureFlag
	for rows.Next() {
		var flag FeatureFlag
		var conditionsJSON, tagsJSON, metadataJSON []byte
		var expiresAt sql.NullTime

		err := rows.Scan(
			&flag.ID, &flag.Name, &flag.Description, &flag.Enabled, &flag.Rollout,
			&conditionsJSON, &flag.Environment, &flag.Service, &flag.CreatedBy,
			&flag.CreatedAt, &flag.UpdatedAt, &expiresAt, &tagsJSON, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal(conditionsJSON, &flag.Conditions); err != nil {
			flag.Conditions = make(map[string]interface{})
		}
		if err := json.Unmarshal(tagsJSON, &flag.Tags); err != nil {
			flag.Tags = []string{}
		}
		if err := json.Unmarshal(metadataJSON, &flag.Metadata); err != nil {
			flag.Metadata = make(map[string]interface{})
		}

		if expiresAt.Valid {
			flag.ExpiresAt = &expiresAt.Time
		}

		flags = append(flags, &flag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flags: %w", err)
	}

	return flags, nil
}

// CreateFlag creates a new feature flag
func (r *PostgreSQLRepository) CreateFlag(ctx context.Context, flag *FeatureFlag) error {
	query := `
		INSERT INTO feature_flags (
			id, name, description, enabled, rollout, conditions, environment, 
			service, created_by, created_at, updated_at, expires_at, tags, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	// Marshal JSON fields
	conditionsJSON, _ := json.Marshal(flag.Conditions)
	tagsJSON, _ := json.Marshal(flag.Tags)
	metadataJSON, _ := json.Marshal(flag.Metadata)

	var expiresAt interface{}
	if flag.ExpiresAt != nil {
		expiresAt = *flag.ExpiresAt
	}

	_, err := r.db.ExecContext(ctx, query,
		flag.ID, flag.Name, flag.Description, flag.Enabled, flag.Rollout,
		conditionsJSON, flag.Environment, flag.Service, flag.CreatedBy,
		flag.CreatedAt, flag.UpdatedAt, expiresAt, tagsJSON, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to create flag: %w", err)
	}

	return nil
}

// UpdateFlag updates an existing feature flag
func (r *PostgreSQLRepository) UpdateFlag(ctx context.Context, flag *FeatureFlag) error {
	query := `
		UPDATE feature_flags SET 
			name = $2, description = $3, enabled = $4, rollout = $5, 
			conditions = $6, environment = $7, service = $8, 
			updated_at = $9, expires_at = $10, tags = $11, metadata = $12
		WHERE id = $1
	`

	// Marshal JSON fields
	conditionsJSON, _ := json.Marshal(flag.Conditions)
	tagsJSON, _ := json.Marshal(flag.Tags)
	metadataJSON, _ := json.Marshal(flag.Metadata)

	var expiresAt interface{}
	if flag.ExpiresAt != nil {
		expiresAt = *flag.ExpiresAt
	}

	result, err := r.db.ExecContext(ctx, query,
		flag.ID, flag.Name, flag.Description, flag.Enabled, flag.Rollout,
		conditionsJSON, flag.Environment, flag.Service,
		flag.UpdatedAt, expiresAt, tagsJSON, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrFlagNotFound
	}

	return nil
}

// DeleteFlag deletes a feature flag
func (r *PostgreSQLRepository) DeleteFlag(ctx context.Context, flagID string) error {
	query := `DELETE FROM feature_flags WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, flagID)
	if err != nil {
		return fmt.Errorf("failed to delete flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrFlagNotFound
	}

	return nil
}

// GetFlagsByTags retrieves feature flags by tags
func (r *PostgreSQLRepository) GetFlagsByTags(ctx context.Context, tags []string) ([]*FeatureFlag, error) {
	query := `
		SELECT id, name, description, enabled, rollout, conditions, environment, 
		       service, created_by, created_at, updated_at, expires_at, tags, metadata
		FROM feature_flags 
		WHERE tags && $1
		ORDER BY created_at DESC
	`

	// Convert tags to PostgreSQL array format
	tagsStr := "{"
	for i, tag := range tags {
		if i > 0 {
			tagsStr += ","
		}
		tagsStr += tag
	}
	tagsStr += "}"
	rows, err := r.db.QueryContext(ctx, query, tagsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query flags by tags: %w", err)
	}
	defer rows.Close()

	var flags []*FeatureFlag
	for rows.Next() {
		var flag FeatureFlag
		var conditionsJSON, tagsJSON, metadataJSON []byte
		var expiresAt sql.NullTime

		err := rows.Scan(
			&flag.ID, &flag.Name, &flag.Description, &flag.Enabled, &flag.Rollout,
			&conditionsJSON, &flag.Environment, &flag.Service, &flag.CreatedBy,
			&flag.CreatedAt, &flag.UpdatedAt, &expiresAt, &tagsJSON, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal(conditionsJSON, &flag.Conditions); err != nil {
			flag.Conditions = make(map[string]interface{})
		}
		if err := json.Unmarshal(tagsJSON, &flag.Tags); err != nil {
			flag.Tags = []string{}
		}
		if err := json.Unmarshal(metadataJSON, &flag.Metadata); err != nil {
			flag.Metadata = make(map[string]interface{})
		}

		if expiresAt.Valid {
			flag.ExpiresAt = &expiresAt.Time
		}

		flags = append(flags, &flag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flags: %w", err)
	}

	return flags, nil
}

// GetExpiredFlags retrieves expired feature flags
func (r *PostgreSQLRepository) GetExpiredFlags(ctx context.Context) ([]*FeatureFlag, error) {
	query := `
		SELECT id, name, description, enabled, rollout, conditions, environment, 
		       service, created_by, created_at, updated_at, expires_at, tags, metadata
		FROM feature_flags 
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired flags: %w", err)
	}
	defer rows.Close()

	var flags []*FeatureFlag
	for rows.Next() {
		var flag FeatureFlag
		var conditionsJSON, tagsJSON, metadataJSON []byte
		var expiresAt sql.NullTime

		err := rows.Scan(
			&flag.ID, &flag.Name, &flag.Description, &flag.Enabled, &flag.Rollout,
			&conditionsJSON, &flag.Environment, &flag.Service, &flag.CreatedBy,
			&flag.CreatedAt, &flag.UpdatedAt, &expiresAt, &tagsJSON, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal(conditionsJSON, &flag.Conditions); err != nil {
			flag.Conditions = make(map[string]interface{})
		}
		if err := json.Unmarshal(tagsJSON, &flag.Tags); err != nil {
			flag.Tags = []string{}
		}
		if err := json.Unmarshal(metadataJSON, &flag.Metadata); err != nil {
			flag.Metadata = make(map[string]interface{})
		}

		if expiresAt.Valid {
			flag.ExpiresAt = &expiresAt.Time
		}

		flags = append(flags, &flag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flags: %w", err)
	}

	return flags, nil
}

// CreateTables creates the necessary database tables for feature flags
func (r *PostgreSQLRepository) CreateTables(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS feature_flags (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			enabled BOOLEAN NOT NULL DEFAULT false,
			rollout DECIMAL(3,2) NOT NULL DEFAULT 0.0 CHECK (rollout >= 0.0 AND rollout <= 1.0),
			conditions JSONB DEFAULT '{}',
			environment VARCHAR(100) NOT NULL,
			service VARCHAR(100) NOT NULL,
			created_by VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE,
			tags TEXT[] DEFAULT '{}',
			metadata JSONB DEFAULT '{}'
		);

		CREATE INDEX IF NOT EXISTS idx_feature_flags_service_env ON feature_flags(service, environment);
		CREATE INDEX IF NOT EXISTS idx_feature_flags_tags ON feature_flags USING GIN(tags);
		CREATE INDEX IF NOT EXISTS idx_feature_flags_expires_at ON feature_flags(expires_at) WHERE expires_at IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled ON feature_flags(enabled);

		-- Create feature flag events table for analytics
		CREATE TABLE IF NOT EXISTS feature_flag_events (
			id VARCHAR(255) PRIMARY KEY,
			flag_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255),
			service VARCHAR(100) NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			result BOOLEAN NOT NULL,
			variant VARCHAR(100),
			metadata JSONB DEFAULT '{}',
			timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			duration INTERVAL
		);

		CREATE INDEX IF NOT EXISTS idx_feature_flag_events_flag_id ON feature_flag_events(flag_id);
		CREATE INDEX IF NOT EXISTS idx_feature_flag_events_user_id ON feature_flag_events(user_id);
		CREATE INDEX IF NOT EXISTS idx_feature_flag_events_timestamp ON feature_flag_events(timestamp);
		CREATE INDEX IF NOT EXISTS idx_feature_flag_events_service ON feature_flag_events(service);
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// InMemoryRepository implements FeatureFlagRepository for in-memory storage
type InMemoryRepository struct {
	flags map[string]*FeatureFlag
	mu    sync.RWMutex
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() FeatureFlagRepository {
	return &InMemoryRepository{
		flags: make(map[string]*FeatureFlag),
	}
}

// GetFlag retrieves a single feature flag by ID
func (r *InMemoryRepository) GetFlag(ctx context.Context, flagID string) (*FeatureFlag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	flag, exists := r.flags[flagID]
	if !exists {
		return nil, ErrFlagNotFound
	}

	// Return a copy to prevent external modifications
	flagCopy := *flag
	return &flagCopy, nil
}

// GetFlags retrieves all feature flags for a service and environment
func (r *InMemoryRepository) GetFlags(ctx context.Context, service string, environment string) ([]*FeatureFlag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var flags []*FeatureFlag
	for _, flag := range r.flags {
		if flag.Service == service && flag.Environment == environment {
			// Return a copy to prevent external modifications
			flagCopy := *flag
			flags = append(flags, &flagCopy)
		}
	}

	return flags, nil
}

// CreateFlag creates a new feature flag
func (r *InMemoryRepository) CreateFlag(ctx context.Context, flag *FeatureFlag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.flags[flag.ID]; exists {
		return fmt.Errorf("flag with ID %s already exists", flag.ID)
	}

	// Store a copy to prevent external modifications
	flagCopy := *flag
	r.flags[flag.ID] = &flagCopy

	return nil
}

// UpdateFlag updates an existing feature flag
func (r *InMemoryRepository) UpdateFlag(ctx context.Context, flag *FeatureFlag) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.flags[flag.ID]; !exists {
		return ErrFlagNotFound
	}

	// Store a copy to prevent external modifications
	flagCopy := *flag
	r.flags[flag.ID] = &flagCopy

	return nil
}

// DeleteFlag deletes a feature flag
func (r *InMemoryRepository) DeleteFlag(ctx context.Context, flagID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.flags[flagID]; !exists {
		return ErrFlagNotFound
	}

	delete(r.flags, flagID)
	return nil
}

// GetFlagsByTags retrieves feature flags by tags
func (r *InMemoryRepository) GetFlagsByTags(ctx context.Context, tags []string) ([]*FeatureFlag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var flags []*FeatureFlag
	for _, flag := range r.flags {
		if r.hasAnyTag(flag.Tags, tags) {
			// Return a copy to prevent external modifications
			flagCopy := *flag
			flags = append(flags, &flagCopy)
		}
	}

	return flags, nil
}

// GetExpiredFlags retrieves expired feature flags
func (r *InMemoryRepository) GetExpiredFlags(ctx context.Context) ([]*FeatureFlag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var flags []*FeatureFlag
	for _, flag := range r.flags {
		if flag.ExpiresAt != nil && now.After(*flag.ExpiresAt) {
			// Return a copy to prevent external modifications
			flagCopy := *flag
			flags = append(flags, &flagCopy)
		}
	}

	return flags, nil
}

// Helper method to check if flag has any of the specified tags
func (r *InMemoryRepository) hasAnyTag(flagTags []string, searchTags []string) bool {
	for _, flagTag := range flagTags {
		for _, searchTag := range searchTags {
			if flagTag == searchTag {
				return true
			}
		}
	}
	return false
}
