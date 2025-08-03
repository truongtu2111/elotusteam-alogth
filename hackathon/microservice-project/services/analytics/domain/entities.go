package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of analytics event
type EventType string

const (
	EventTypeFileUpload   EventType = "file_upload"
	EventTypeFileDownload EventType = "file_download"
	EventTypeFileView     EventType = "file_view"
	EventTypeFileShare    EventType = "file_share"
	EventTypeFileDelete   EventType = "file_delete"
	EventTypeUserLogin    EventType = "user_login"
	EventTypeUserLogout   EventType = "user_logout"
	EventTypeUserRegister EventType = "user_register"
	EventTypeAPICall      EventType = "api_call"
	EventTypeError        EventType = "error"
)

// Event represents an analytics event
type Event struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	UserID      *uuid.UUID             `json:"user_id" db:"user_id"`
	SessionID   *uuid.UUID             `json:"session_id" db:"session_id"`
	Type        EventType              `json:"type" db:"type"`
	Action      string                 `json:"action" db:"action"`
	Resource    string                 `json:"resource" db:"resource"`
	ResourceID  *uuid.UUID             `json:"resource_id" db:"resource_id"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	IPAddress   string                 `json:"ip_address" db:"ip_address"`
	UserAgent   string                 `json:"user_agent" db:"user_agent"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// UserActivity represents aggregated user activity
type UserActivity struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	Date            time.Time `json:"date" db:"date"`
	TotalEvents     int64     `json:"total_events" db:"total_events"`
	FileUploads     int64     `json:"file_uploads" db:"file_uploads"`
	FileDownloads   int64     `json:"file_downloads" db:"file_downloads"`
	FileViews       int64     `json:"file_views" db:"file_views"`
	FileShares      int64     `json:"file_shares" db:"file_shares"`
	APICallsCount   int64     `json:"api_calls_count" db:"api_calls_count"`
	ErrorsCount     int64     `json:"errors_count" db:"errors_count"`
	SessionDuration int64     `json:"session_duration" db:"session_duration"` // in seconds
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Date              time.Time `json:"date" db:"date"`
	TotalUsers        int64     `json:"total_users" db:"total_users"`
	ActiveUsers       int64     `json:"active_users" db:"active_users"`
	NewUsers          int64     `json:"new_users" db:"new_users"`
	TotalFiles        int64     `json:"total_files" db:"total_files"`
	TotalFileSize     int64     `json:"total_file_size" db:"total_file_size"`
	TotalEvents       int64     `json:"total_events" db:"total_events"`
	APICallsCount     int64     `json:"api_calls_count" db:"api_calls_count"`
	ErrorRate         float64   `json:"error_rate" db:"error_rate"`
	AverageResponseTime float64 `json:"average_response_time" db:"average_response_time"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// FileMetrics represents file-related metrics
type FileMetrics struct {
	ID            uuid.UUID `json:"id" db:"id"`
	FileID        uuid.UUID `json:"file_id" db:"file_id"`
	FileName      string    `json:"file_name" db:"file_name"`
	FileType      string    `json:"file_type" db:"file_type"`
	FileSize      int64     `json:"file_size" db:"file_size"`
	OwnerID       uuid.UUID `json:"owner_id" db:"owner_id"`
	ViewCount     int64     `json:"view_count" db:"view_count"`
	DownloadCount int64     `json:"download_count" db:"download_count"`
	ShareCount    int64     `json:"share_count" db:"share_count"`
	LastAccessed  *time.Time `json:"last_accessed" db:"last_accessed"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// APIMetrics represents API endpoint metrics
type APIMetrics struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Endpoint        string    `json:"endpoint" db:"endpoint"`
	Method          string    `json:"method" db:"method"`
	Date            time.Time `json:"date" db:"date"`
	RequestCount    int64     `json:"request_count" db:"request_count"`
	SuccessCount    int64     `json:"success_count" db:"success_count"`
	ErrorCount      int64     `json:"error_count" db:"error_count"`
	AverageLatency  float64   `json:"average_latency" db:"average_latency"`
	MinLatency      float64   `json:"min_latency" db:"min_latency"`
	MaxLatency      float64   `json:"max_latency" db:"max_latency"`
	P95Latency      float64   `json:"p95_latency" db:"p95_latency"`
	P99Latency      float64   `json:"p99_latency" db:"p99_latency"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ErrorMetrics represents error tracking metrics
type ErrorMetrics struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ErrorType    string    `json:"error_type" db:"error_type"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	Service      string    `json:"service" db:"service"`
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	Date         time.Time `json:"date" db:"date"`
	Count        int64     `json:"count" db:"count"`
	FirstSeen    time.Time `json:"first_seen" db:"first_seen"`
	LastSeen     time.Time `json:"last_seen" db:"last_seen"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ReportType represents the type of analytics report
type ReportType string

const (
	ReportTypeDaily        ReportType = "daily"
	ReportTypeWeekly       ReportType = "weekly"
	ReportTypeMonthly      ReportType = "monthly"
	ReportTypeCustom       ReportType = "custom"
	ReportTypeUserActivity ReportType = "user_activity"
	ReportTypeSystemMetrics ReportType = "system_metrics"
	ReportTypeFileMetrics  ReportType = "file_metrics"
	ReportTypeAPIMetrics   ReportType = "api_metrics"
	ReportTypeErrorMetrics ReportType = "error_metrics"
)

// ReportStatus represents the status of a report
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusCompleted ReportStatus = "completed"
	ReportStatusFailed    ReportStatus = "failed"
)

// Report represents an analytics report
type Report struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Type        ReportType             `json:"type" db:"type"`
	Description string                 `json:"description" db:"description"`
	Filters     map[string]interface{} `json:"filters" db:"filters"`
	Data        []byte                 `json:"data" db:"data"`
	GeneratedBy uuid.UUID              `json:"generated_by" db:"generated_by"`
	StartDate   time.Time              `json:"start_date" db:"start_date"`
	EndDate     time.Time              `json:"end_date" db:"end_date"`
	Status      ReportStatus           `json:"status" db:"status"`
	CompletedAt *time.Time             `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}