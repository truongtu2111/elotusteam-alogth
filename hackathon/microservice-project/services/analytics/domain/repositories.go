package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventRepository defines the interface for event data operations
type EventRepository interface {
	// Create creates a new event
	Create(ctx context.Context, event *Event) error

	// CreateBatch creates multiple events in a batch
	CreateBatch(ctx context.Context, events []*Event) error

	// GetByID retrieves an event by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Event, error)

	// GetByUserID retrieves events for a user with pagination
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Event, error)

	// GetByType retrieves events by type with pagination
	GetByType(ctx context.Context, eventType EventType, limit, offset int) ([]*Event, error)

	// GetByDateRange retrieves events within a date range
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*Event, error)

	// GetByUserAndDateRange retrieves events for a user within a date range
	GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*Event, error)

	// CountByType counts events by type within a date range
	CountByType(ctx context.Context, eventType EventType, startDate, endDate time.Time) (int64, error)

	// CountByUser counts events for a user within a date range
	CountByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (int64, error)

	// Delete deletes events older than the specified date
	DeleteOlderThan(ctx context.Context, date time.Time) error
}

// UserActivityRepository defines the interface for user activity data operations
type UserActivityRepository interface {
	// Create creates a new user activity record
	Create(ctx context.Context, activity *UserActivity) error

	// Update updates an existing user activity record
	Update(ctx context.Context, activity *UserActivity) error

	// GetByUserAndDate retrieves user activity for a specific user and date
	GetByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*UserActivity, error)

	// GetByUser retrieves user activity for a user within a date range
	GetByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*UserActivity, error)

	// GetTopActiveUsers retrieves the most active users within a date range
	GetTopActiveUsers(ctx context.Context, startDate, endDate time.Time, limit int) ([]*UserActivity, error)

	// GetAggregatedByDateRange retrieves aggregated activity within a date range
	GetAggregatedByDateRange(ctx context.Context, startDate, endDate time.Time) (*UserActivity, error)
}

// SystemMetricsRepository defines the interface for system metrics data operations
type SystemMetricsRepository interface {
	// Create creates a new system metrics record
	Create(ctx context.Context, metrics *SystemMetrics) error

	// Update updates an existing system metrics record
	Update(ctx context.Context, metrics *SystemMetrics) error

	// GetByDate retrieves system metrics for a specific date
	GetByDate(ctx context.Context, date time.Time) (*SystemMetrics, error)

	// GetByDateRange retrieves system metrics within a date range
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*SystemMetrics, error)

	// GetLatest retrieves the latest system metrics
	GetLatest(ctx context.Context) (*SystemMetrics, error)
}

// FileMetricsRepository defines the interface for file metrics data operations
type FileMetricsRepository interface {
	// Create creates a new file metrics record
	Create(ctx context.Context, metrics *FileMetrics) error

	// Update updates an existing file metrics record
	Update(ctx context.Context, metrics *FileMetrics) error

	// GetByFileID retrieves file metrics for a specific file
	GetByFileID(ctx context.Context, fileID uuid.UUID) (*FileMetrics, error)

	// GetByOwner retrieves file metrics for files owned by a user
	GetByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*FileMetrics, error)

	// GetTopFiles retrieves the most accessed files
	GetTopFiles(ctx context.Context, metric string, limit int) ([]*FileMetrics, error)

	// IncrementViewCount increments the view count for a file
	IncrementViewCount(ctx context.Context, fileID uuid.UUID) error

	// IncrementDownloadCount increments the download count for a file
	IncrementDownloadCount(ctx context.Context, fileID uuid.UUID) error

	// IncrementShareCount increments the share count for a file
	IncrementShareCount(ctx context.Context, fileID uuid.UUID) error

	// UpdateLastAccessed updates the last accessed time for a file
	UpdateLastAccessed(ctx context.Context, fileID uuid.UUID, accessTime time.Time) error
}

// APIMetricsRepository defines the interface for API metrics data operations
type APIMetricsRepository interface {
	// Create creates a new API metrics record
	Create(ctx context.Context, metrics *APIMetrics) error

	// Update updates an existing API metrics record
	Update(ctx context.Context, metrics *APIMetrics) error

	// GetByEndpointAndDate retrieves API metrics for a specific endpoint and date
	GetByEndpointAndDate(ctx context.Context, endpoint, method string, date time.Time) (*APIMetrics, error)

	// GetByDateRange retrieves API metrics within a date range
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*APIMetrics, error)

	// GetTopEndpoints retrieves the most called endpoints
	GetTopEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) ([]*APIMetrics, error)

	// GetSlowestEndpoints retrieves the slowest endpoints
	GetSlowestEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) ([]*APIMetrics, error)
}

// ErrorMetricsRepository defines the interface for error metrics data operations
type ErrorMetricsRepository interface {
	// Create creates a new error metrics record
	Create(ctx context.Context, metrics *ErrorMetrics) error

	// Update updates an existing error metrics record
	Update(ctx context.Context, metrics *ErrorMetrics) error

	// GetByTypeAndDate retrieves error metrics for a specific type and date
	GetByTypeAndDate(ctx context.Context, errorType string, date time.Time) (*ErrorMetrics, error)

	// GetByDateRange retrieves error metrics within a date range
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*ErrorMetrics, error)

	// GetTopErrors retrieves the most frequent errors
	GetTopErrors(ctx context.Context, startDate, endDate time.Time, limit int) ([]*ErrorMetrics, error)

	// IncrementErrorCount increments the error count for a specific error type
	IncrementErrorCount(ctx context.Context, errorType, errorMessage, service, endpoint string) error
}

// ReportRepository defines the interface for report data operations
type ReportRepository interface {
	// Create creates a new report
	Create(ctx context.Context, report *Report) error

	// GetByID retrieves a report by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Report, error)

	// GetByUser retrieves reports generated by a user
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Report, error)

	// GetByType retrieves reports by type
	GetByType(ctx context.Context, reportType ReportType, limit, offset int) ([]*Report, error)

	// Update updates an existing report
	Update(ctx context.Context, report *Report) error

	// Delete deletes a report
	Delete(ctx context.Context, id uuid.UUID) error
}

// RepositoryManager defines the interface for managing all repositories
type RepositoryManager interface {
	// Event returns the event repository
	Event() EventRepository

	// UserActivity returns the user activity repository
	UserActivity() UserActivityRepository

	// SystemMetrics returns the system metrics repository
	SystemMetrics() SystemMetricsRepository

	// FileMetrics returns the file metrics repository
	FileMetrics() FileMetricsRepository

	// APIMetrics returns the API metrics repository
	APIMetrics() APIMetricsRepository

	// ErrorMetrics returns the error metrics repository
	ErrorMetrics() ErrorMetricsRepository

	// Report returns the report repository
	Report() ReportRepository

	// BeginTx begins a database transaction
	BeginTx(ctx context.Context) (RepositoryManager, error)

	// Commit commits the current transaction
	Commit() error

	// Rollback rolls back the current transaction
	Rollback() error
}
