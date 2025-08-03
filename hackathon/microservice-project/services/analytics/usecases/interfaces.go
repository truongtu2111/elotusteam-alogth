package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/elotusteam/microservice-project/services/analytics/domain"
)

// Event tracking requests and responses
type TrackEventRequest struct {
	UserID     uuid.UUID                `json:"user_id" binding:"required"`
	SessionID  *uuid.UUID               `json:"session_id,omitempty"`
	EventType  domain.EventType         `json:"event_type" binding:"required"`
	Action     string                   `json:"action" binding:"required"`
	Resource   string                   `json:"resource,omitempty"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
	Timestamp  *time.Time               `json:"timestamp,omitempty"`
}

type TrackBatchEventsRequest struct {
	Events []TrackEventRequest `json:"events" binding:"required,dive"`
}

type GetEventsRequest struct {
	UserID    *uuid.UUID        `json:"user_id,omitempty"`
	EventType *domain.EventType `json:"event_type,omitempty"`
	StartDate *time.Time        `json:"start_date,omitempty"`
	EndDate   *time.Time        `json:"end_date,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
}

type GetEventsResponse struct {
	Events     []*domain.Event `json:"events"`
	Total      int64           `json:"total"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	HasMore    bool            `json:"has_more"`
}

// User activity requests and responses
type GetUserActivityRequest struct {
	UserID    uuid.UUID  `json:"user_id" binding:"required"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

type GetUserActivityResponse struct {
	Activities []*domain.UserActivity `json:"activities"`
	Total      int64                  `json:"total"`
}

type GetTopUsersRequest struct {
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
	Limit     int       `json:"limit,omitempty"`
}

type GetTopUsersResponse struct {
	Users []*domain.UserActivity `json:"users"`
	Total int64                  `json:"total"`
}

// System metrics requests and responses
type GetSystemMetricsRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

type GetSystemMetricsResponse struct {
	Metrics []*domain.SystemMetrics `json:"metrics"`
	Total   int64                   `json:"total"`
}

// File metrics requests and responses
type GetFileMetricsRequest struct {
	FileID   *uuid.UUID `json:"file_id,omitempty"`
	OwnerID  *uuid.UUID `json:"owner_id,omitempty"`
	Metric   string     `json:"metric,omitempty"` // "views", "downloads", "shares"
	Limit    int        `json:"limit,omitempty"`
	Offset   int        `json:"offset,omitempty"`
}

type GetFileMetricsResponse struct {
	Metrics []*domain.FileMetrics `json:"metrics"`
	Total   int64                 `json:"total"`
}

type UpdateFileMetricsRequest struct {
	FileID     uuid.UUID `json:"file_id" binding:"required"`
	MetricType string    `json:"metric_type" binding:"required"` // "view", "download", "share"
}

// API metrics requests and responses
type GetAPIMetricsRequest struct {
	Endpoint  string     `json:"endpoint,omitempty"`
	Method    string     `json:"method,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

type GetAPIMetricsResponse struct {
	Metrics []*domain.APIMetrics `json:"metrics"`
	Total   int64                `json:"total"`
}

// Error metrics requests and responses
type GetErrorMetricsRequest struct {
	ErrorType string     `json:"error_type,omitempty"`
	Service   string     `json:"service,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

type GetErrorMetricsResponse struct {
	Metrics []*domain.ErrorMetrics `json:"metrics"`
	Total   int64                  `json:"total"`
}

type TrackErrorRequest struct {
	ErrorType    string `json:"error_type" binding:"required"`
	ErrorMessage string `json:"error_message" binding:"required"`
	Service      string `json:"service" binding:"required"`
	Endpoint     string `json:"endpoint,omitempty"`
}

// Report requests and responses
type GenerateReportRequest struct {
	ReportType domain.ReportType      `json:"report_type" binding:"required"`
	StartDate  time.Time              `json:"start_date" binding:"required"`
	EndDate    time.Time              `json:"end_date" binding:"required"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	Format     string                 `json:"format,omitempty"` // "json", "csv", "pdf"
}

type GetReportsRequest struct {
	UserID     *uuid.UUID         `json:"user_id,omitempty"`
	ReportType *domain.ReportType `json:"report_type,omitempty"`
	Limit      int                `json:"limit,omitempty"`
	Offset     int                `json:"offset,omitempty"`
}

type GetReportsResponse struct {
	Reports []*domain.Report `json:"reports"`
	Total   int64            `json:"total"`
}

// Dashboard data response
type DashboardData struct {
	TotalUsers       int64                   `json:"total_users"`
	ActiveUsers      int64                   `json:"active_users"`
	TotalFiles       int64                   `json:"total_files"`
	TotalEvents      int64                   `json:"total_events"`
	SystemHealth     *domain.SystemMetrics   `json:"system_health"`
	TopFiles         []*domain.FileMetrics   `json:"top_files"`
	TopEndpoints     []*domain.APIMetrics    `json:"top_endpoints"`
	RecentErrors     []*domain.ErrorMetrics  `json:"recent_errors"`
	UserActivity     []*domain.UserActivity  `json:"user_activity"`
	EventDistribution map[string]int64       `json:"event_distribution"`
}

// Service interfaces

// EventService defines the interface for event tracking operations
type EventService interface {
	// TrackEvent tracks a single event
	TrackEvent(ctx context.Context, req *TrackEventRequest) error
	
	// TrackBatchEvents tracks multiple events in a batch
	TrackBatchEvents(ctx context.Context, req *TrackBatchEventsRequest) error
	
	// GetEvents retrieves events based on filters
	GetEvents(ctx context.Context, req *GetEventsRequest) (*GetEventsResponse, error)
	
	// GetEventsByUser retrieves events for a specific user
	GetEventsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*GetEventsResponse, error)
	
	// GetEventsByType retrieves events by type
	GetEventsByType(ctx context.Context, eventType domain.EventType, limit, offset int) (*GetEventsResponse, error)
	
	// GetEventStats retrieves event statistics
	GetEventStats(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error)
}

// UserActivityService defines the interface for user activity operations
type UserActivityService interface {
	// GetUserActivity retrieves user activity data
	GetUserActivity(ctx context.Context, req *GetUserActivityRequest) (*GetUserActivityResponse, error)
	
	// GetTopActiveUsers retrieves the most active users
	GetTopActiveUsers(ctx context.Context, req *GetTopUsersRequest) (*GetTopUsersResponse, error)
	
	// UpdateUserActivity updates user activity metrics
	UpdateUserActivity(ctx context.Context, userID uuid.UUID, action string) error
	
	// GetUserStats retrieves user statistics
	GetUserStats(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domain.UserActivity, error)
}

// SystemMetricsService defines the interface for system metrics operations
type SystemMetricsService interface {
	// GetSystemMetrics retrieves system metrics
	GetSystemMetrics(ctx context.Context, req *GetSystemMetricsRequest) (*GetSystemMetricsResponse, error)
	
	// GetLatestSystemMetrics retrieves the latest system metrics
	GetLatestSystemMetrics(ctx context.Context) (*domain.SystemMetrics, error)
	
	// UpdateSystemMetrics updates system metrics
	UpdateSystemMetrics(ctx context.Context, metrics *domain.SystemMetrics) error
	
	// GetSystemHealth retrieves system health status
	GetSystemHealth(ctx context.Context) (map[string]interface{}, error)
}

// FileMetricsService defines the interface for file metrics operations
type FileMetricsService interface {
	// GetFileMetrics retrieves file metrics
	GetFileMetrics(ctx context.Context, req *GetFileMetricsRequest) (*GetFileMetricsResponse, error)
	
	// UpdateFileMetrics updates file metrics
	UpdateFileMetrics(ctx context.Context, req *UpdateFileMetricsRequest) error
	
	// GetTopFiles retrieves the most accessed files
	GetTopFiles(ctx context.Context, metric string, limit int) (*GetFileMetricsResponse, error)
	
	// GetFileStats retrieves file statistics
	GetFileStats(ctx context.Context, fileID uuid.UUID) (*domain.FileMetrics, error)
}

// APIMetricsService defines the interface for API metrics operations
type APIMetricsService interface {
	// GetAPIMetrics retrieves API metrics
	GetAPIMetrics(ctx context.Context, req *GetAPIMetricsRequest) (*GetAPIMetricsResponse, error)
	
	// TrackAPICall tracks an API call
	TrackAPICall(ctx context.Context, endpoint, method string, responseTime time.Duration, statusCode int) error
	
	// GetTopEndpoints retrieves the most called endpoints
	GetTopEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) (*GetAPIMetricsResponse, error)
	
	// GetSlowestEndpoints retrieves the slowest endpoints
	GetSlowestEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) (*GetAPIMetricsResponse, error)
}

// ErrorMetricsService defines the interface for error metrics operations
type ErrorMetricsService interface {
	// GetErrorMetrics retrieves error metrics
	GetErrorMetrics(ctx context.Context, req *GetErrorMetricsRequest) (*GetErrorMetricsResponse, error)
	
	// TrackError tracks an error occurrence
	TrackError(ctx context.Context, req *TrackErrorRequest) error
	
	// GetTopErrors retrieves the most frequent errors
	GetTopErrors(ctx context.Context, startDate, endDate time.Time, limit int) (*GetErrorMetricsResponse, error)
	
	// GetErrorStats retrieves error statistics
	GetErrorStats(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error)
}

// ReportService defines the interface for report operations
type ReportService interface {
	// GenerateReport generates a new report
	GenerateReport(ctx context.Context, req *GenerateReportRequest, userID uuid.UUID) (*domain.Report, error)
	
	// GetReports retrieves reports
	GetReports(ctx context.Context, req *GetReportsRequest) (*GetReportsResponse, error)
	
	// GetReportByID retrieves a report by ID
	GetReportByID(ctx context.Context, id uuid.UUID) (*domain.Report, error)
	
	// DeleteReport deletes a report
	DeleteReport(ctx context.Context, id uuid.UUID) error
	
	// ScheduleReport schedules a recurring report
	ScheduleReport(ctx context.Context, req *GenerateReportRequest, userID uuid.UUID, schedule string) error
}

// DashboardService defines the interface for dashboard operations
type DashboardService interface {
	// GetDashboardData retrieves dashboard data
	GetDashboardData(ctx context.Context, startDate, endDate time.Time) (*DashboardData, error)
	
	// GetUserDashboard retrieves user-specific dashboard data
	GetUserDashboard(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*DashboardData, error)
	
	// GetRealTimeMetrics retrieves real-time metrics
	GetRealTimeMetrics(ctx context.Context) (map[string]interface{}, error)
}

// AnalyticsService defines the main analytics service interface
type AnalyticsService interface {
	// Event service
	EventService
	
	// User activity service
	UserActivityService
	
	// System metrics service
	SystemMetricsService
	
	// File metrics service
	FileMetricsService
	
	// API metrics service
	APIMetricsService
	
	// Error metrics service
	ErrorMetricsService
	
	// Report service
	ReportService
	
	// Dashboard service
	DashboardService
}