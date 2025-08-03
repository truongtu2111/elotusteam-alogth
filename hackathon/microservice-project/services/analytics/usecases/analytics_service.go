package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/elotusteam/microservice-project/services/analytics/domain"
)

// analyticsService implements the AnalyticsService interface
type analyticsService struct {
	repoManager domain.RepositoryManager
}

// NewAnalyticsService creates a new analytics service instance
func NewAnalyticsService(repoManager domain.RepositoryManager) AnalyticsService {
	return &analyticsService{
		repoManager: repoManager,
	}
}

// Event Service Methods

func (s *analyticsService) TrackEvent(ctx context.Context, req *TrackEventRequest) error {
	event := &domain.Event{
		ID:        uuid.New(),
		UserID:    &req.UserID,
		SessionID: req.SessionID,
		Type:      req.EventType,
		Action:    req.Action,
		Resource:  req.Resource,
		Metadata:  req.Metadata,
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}

	if req.Timestamp != nil {
		event.Timestamp = *req.Timestamp
	}

	return s.repoManager.Event().Create(ctx, event)
}

func (s *analyticsService) TrackBatchEvents(ctx context.Context, req *TrackBatchEventsRequest) error {
	events := make([]*domain.Event, len(req.Events))
	for i, eventReq := range req.Events {
		event := &domain.Event{
			ID:        uuid.New(),
			UserID:    &eventReq.UserID,
			SessionID: eventReq.SessionID,
			Type:      eventReq.EventType,
			Action:    eventReq.Action,
			Resource:  eventReq.Resource,
			Metadata:  eventReq.Metadata,
			Timestamp: time.Now(),
			CreatedAt: time.Now(),
		}

		if eventReq.Timestamp != nil {
			event.Timestamp = *eventReq.Timestamp
		}

		events[i] = event
	}

	return s.repoManager.Event().CreateBatch(ctx, events)
}

func (s *analyticsService) GetEvents(ctx context.Context, req *GetEventsRequest) (*GetEventsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var events []*domain.Event
	var err error

	if req.UserID != nil && req.StartDate != nil && req.EndDate != nil {
		events, err = s.repoManager.Event().GetByUserAndDateRange(ctx, *req.UserID, *req.StartDate, *req.EndDate, limit, offset)
	} else if req.UserID != nil {
		events, err = s.repoManager.Event().GetByUserID(ctx, *req.UserID, limit, offset)
	} else if req.EventType != nil {
		events, err = s.repoManager.Event().GetByType(ctx, *req.EventType, limit, offset)
	} else if req.StartDate != nil && req.EndDate != nil {
		events, err = s.repoManager.Event().GetByDateRange(ctx, *req.StartDate, *req.EndDate, limit, offset)
	} else {
		// Default to recent events
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -7) // Last 7 days
		events, err = s.repoManager.Event().GetByDateRange(ctx, startDate, endDate, limit, offset)
	}

	if err != nil {
		return nil, err
	}

	return &GetEventsResponse{
		Events:  events,
		Total:   int64(len(events)),
		Limit:   limit,
		Offset:  offset,
		HasMore: len(events) == limit,
	}, nil
}

func (s *analyticsService) GetEventsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*GetEventsResponse, error) {
	events, err := s.repoManager.Event().GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &GetEventsResponse{
		Events:  events,
		Total:   int64(len(events)),
		Limit:   limit,
		Offset:  offset,
		HasMore: len(events) == limit,
	}, nil
}

func (s *analyticsService) GetEventsByType(ctx context.Context, eventType domain.EventType, limit, offset int) (*GetEventsResponse, error) {
	events, err := s.repoManager.Event().GetByType(ctx, eventType, limit, offset)
	if err != nil {
		return nil, err
	}

	return &GetEventsResponse{
		Events:  events,
		Total:   int64(len(events)),
		Limit:   limit,
		Offset:  offset,
		HasMore: len(events) == limit,
	}, nil
}

func (s *analyticsService) GetEventStats(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count events by type
	eventTypes := []domain.EventType{
		domain.EventTypeUserLogin,
		domain.EventTypeFileUpload,
		domain.EventTypeFileDownload,
		domain.EventTypeAPICall,
		domain.EventTypeError,
	}

	for _, eventType := range eventTypes {
		count, err := s.repoManager.Event().CountByType(ctx, eventType, startDate, endDate)
		if err != nil {
			return nil, err
		}
		stats[string(eventType)] = count
	}

	return stats, nil
}

// User Activity Service Methods

func (s *analyticsService) GetUserActivity(ctx context.Context, req *GetUserActivityRequest) (*GetUserActivityResponse, error) {
	var activities []*domain.UserActivity
	var err error

	if req.StartDate != nil && req.EndDate != nil {
		activities, err = s.repoManager.UserActivity().GetByUser(ctx, req.UserID, *req.StartDate, *req.EndDate)
	} else {
		// Default to last 30 days
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		activities, err = s.repoManager.UserActivity().GetByUser(ctx, req.UserID, startDate, endDate)
	}

	if err != nil {
		return nil, err
	}

	return &GetUserActivityResponse{
		Activities: activities,
		Total:      int64(len(activities)),
	}, nil
}

func (s *analyticsService) GetTopActiveUsers(ctx context.Context, req *GetTopUsersRequest) (*GetTopUsersResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	users, err := s.repoManager.UserActivity().GetTopActiveUsers(ctx, req.StartDate, req.EndDate, limit)
	if err != nil {
		return nil, err
	}

	return &GetTopUsersResponse{
		Users: users,
		Total: int64(len(users)),
	}, nil
}

func (s *analyticsService) UpdateUserActivity(ctx context.Context, userID uuid.UUID, action string) error {
	today := time.Now().Truncate(24 * time.Hour)
	activity, err := s.repoManager.UserActivity().GetByUserAndDate(ctx, userID, today)
	if err != nil {
		// Create new activity record
		activity = &domain.UserActivity{
			ID:              uuid.New(),
			UserID:          userID,
			Date:            today,
			TotalEvents:     0,
			FileUploads:     0,
			FileDownloads:   0,
			FileViews:       0,
			FileShares:      0,
			APICallsCount:   0,
			ErrorsCount:     0,
			SessionDuration: 0,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
	}

	// Update activity based on action
	switch action {
	case "file_upload":
		activity.FileUploads++
	case "file_download":
		activity.FileDownloads++
	case "file_view":
		activity.FileViews++
	case "file_share":
		activity.FileShares++
	case "api_call":
		activity.APICallsCount++
	case "error":
		activity.ErrorsCount++
	default:
		activity.TotalEvents++
	}

	activity.UpdatedAt = time.Now()

	if activity.CreatedAt.IsZero() {
		return s.repoManager.UserActivity().Create(ctx, activity)
	}
	return s.repoManager.UserActivity().Update(ctx, activity)
}

func (s *analyticsService) GetUserStats(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*domain.UserActivity, error) {
	return s.repoManager.UserActivity().GetAggregatedByDateRange(ctx, startDate, endDate)
}

// System Metrics Service Methods

func (s *analyticsService) GetSystemMetrics(ctx context.Context, req *GetSystemMetricsRequest) (*GetSystemMetricsResponse, error) {
	var metrics []*domain.SystemMetrics
	var err error

	if req.StartDate != nil && req.EndDate != nil {
		metrics, err = s.repoManager.SystemMetrics().GetByDateRange(ctx, *req.StartDate, *req.EndDate)
	} else {
		// Default to last 24 hours
		endDate := time.Now()
		startDate := endDate.Add(-24 * time.Hour)
		metrics, err = s.repoManager.SystemMetrics().GetByDateRange(ctx, startDate, endDate)
	}

	if err != nil {
		return nil, err
	}

	return &GetSystemMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

func (s *analyticsService) GetLatestSystemMetrics(ctx context.Context) (*domain.SystemMetrics, error) {
	return s.repoManager.SystemMetrics().GetLatest(ctx)
}

func (s *analyticsService) UpdateSystemMetrics(ctx context.Context, metrics *domain.SystemMetrics) error {
	metrics.ID = uuid.New()
	metrics.Date = time.Now()
	metrics.CreatedAt = time.Now()
	metrics.UpdatedAt = time.Now()
	return s.repoManager.SystemMetrics().Create(ctx, metrics)
}

func (s *analyticsService) GetSystemHealth(ctx context.Context) (map[string]interface{}, error) {
	latest, err := s.repoManager.SystemMetrics().GetLatest(ctx)
	if err != nil {
		return nil, err
	}

	health := map[string]interface{}{
		"status":           "healthy",
		"total_users":      latest.TotalUsers,
		"active_users":     latest.ActiveUsers,
		"new_users":        latest.NewUsers,
		"total_files":      latest.TotalFiles,
		"total_events":     latest.TotalEvents,
		"api_calls":        latest.APICallsCount,
		"error_rate":       latest.ErrorRate,
		"response_time":    latest.AverageResponseTime,
		"last_updated":     latest.Date,
	}

	// Determine health status based on metrics
	if latest.ErrorRate > 2 {
		health["status"] = "warning"
	}
	if latest.ErrorRate > 5 {
		health["status"] = "critical"
	}

	return health, nil
}

// File Metrics Service Methods

func (s *analyticsService) GetFileMetrics(ctx context.Context, req *GetFileMetricsRequest) (*GetFileMetricsResponse, error) {
	var metrics []*domain.FileMetrics
	var err error

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	if req.FileID != nil {
		metric, err := s.repoManager.FileMetrics().GetByFileID(ctx, *req.FileID)
		if err != nil {
			return nil, err
		}
		metrics = []*domain.FileMetrics{metric}
	} else if req.OwnerID != nil {
		metrics, err = s.repoManager.FileMetrics().GetByOwner(ctx, *req.OwnerID, limit, offset)
	} else if req.Metric != "" {
		metrics, err = s.repoManager.FileMetrics().GetTopFiles(ctx, req.Metric, limit)
	} else {
		return nil, fmt.Errorf("invalid request: must specify file_id, owner_id, or metric")
	}

	if err != nil {
		return nil, err
	}

	return &GetFileMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

func (s *analyticsService) UpdateFileMetrics(ctx context.Context, req *UpdateFileMetricsRequest) error {
	switch req.MetricType {
	case "view":
		return s.repoManager.FileMetrics().IncrementViewCount(ctx, req.FileID)
	case "download":
		return s.repoManager.FileMetrics().IncrementDownloadCount(ctx, req.FileID)
	case "share":
		return s.repoManager.FileMetrics().IncrementShareCount(ctx, req.FileID)
	default:
		return fmt.Errorf("invalid metric type: %s", req.MetricType)
	}
}

func (s *analyticsService) GetTopFiles(ctx context.Context, metric string, limit int) (*GetFileMetricsResponse, error) {
	metrics, err := s.repoManager.FileMetrics().GetTopFiles(ctx, metric, limit)
	if err != nil {
		return nil, err
	}

	return &GetFileMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

func (s *analyticsService) GetFileStats(ctx context.Context, fileID uuid.UUID) (*domain.FileMetrics, error) {
	return s.repoManager.FileMetrics().GetByFileID(ctx, fileID)
}

// API Metrics Service Methods

func (s *analyticsService) GetAPIMetrics(ctx context.Context, req *GetAPIMetricsRequest) (*GetAPIMetricsResponse, error) {
	var metrics []*domain.APIMetrics
	var err error

	if req.StartDate != nil && req.EndDate != nil {
		metrics, err = s.repoManager.APIMetrics().GetByDateRange(ctx, *req.StartDate, *req.EndDate)
	} else {
		// Default to last 24 hours
		endDate := time.Now()
		startDate := endDate.Add(-24 * time.Hour)
		metrics, err = s.repoManager.APIMetrics().GetByDateRange(ctx, startDate, endDate)
	}

	if err != nil {
		return nil, err
	}

	return &GetAPIMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

func (s *analyticsService) TrackAPICall(ctx context.Context, endpoint, method string, responseTime time.Duration, statusCode int) error {
	today := time.Now().Truncate(24 * time.Hour)
	metrics, err := s.repoManager.APIMetrics().GetByEndpointAndDate(ctx, endpoint, method, today)
	if err != nil {
		// Create new metrics record
		metrics = &domain.APIMetrics{
			ID:             uuid.New(),
			Endpoint:       endpoint,
			Method:         method,
			Date:           today,
			RequestCount:   0,
			SuccessCount:   0,
			ErrorCount:     0,
			AverageLatency: 0,
			MinLatency:     float64(responseTime.Milliseconds()),
			MaxLatency:     float64(responseTime.Milliseconds()),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
	}

	// Update metrics
	metrics.RequestCount++
	if statusCode >= 200 && statusCode < 400 {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	// Update response time statistics
	latencyMs := float64(responseTime.Milliseconds())
	if metrics.RequestCount == 1 {
		metrics.AverageLatency = latencyMs
		metrics.MinLatency = latencyMs
		metrics.MaxLatency = latencyMs
	} else {
		// Calculate new average
		totalLatency := metrics.AverageLatency*float64(metrics.RequestCount-1) + latencyMs
		metrics.AverageLatency = totalLatency / float64(metrics.RequestCount)
		
		if latencyMs < metrics.MinLatency {
			metrics.MinLatency = latencyMs
		}
		if latencyMs > metrics.MaxLatency {
			metrics.MaxLatency = latencyMs
		}
	}

	metrics.UpdatedAt = time.Now()

	if metrics.CreatedAt.IsZero() {
		return s.repoManager.APIMetrics().Create(ctx, metrics)
	}
	return s.repoManager.APIMetrics().Update(ctx, metrics)
}

func (s *analyticsService) GetTopEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) (*GetAPIMetricsResponse, error) {
	metrics, err := s.repoManager.APIMetrics().GetTopEndpoints(ctx, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}

	return &GetAPIMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

func (s *analyticsService) GetSlowestEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) (*GetAPIMetricsResponse, error) {
	metrics, err := s.repoManager.APIMetrics().GetSlowestEndpoints(ctx, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}

	return &GetAPIMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

// Error Metrics Service Methods

func (s *analyticsService) GetErrorMetrics(ctx context.Context, req *GetErrorMetricsRequest) (*GetErrorMetricsResponse, error) {
	var metrics []*domain.ErrorMetrics
	var err error

	if req.StartDate != nil && req.EndDate != nil {
		metrics, err = s.repoManager.ErrorMetrics().GetByDateRange(ctx, *req.StartDate, *req.EndDate)
	} else {
		// Default to last 24 hours
		endDate := time.Now()
		startDate := endDate.Add(-24 * time.Hour)
		metrics, err = s.repoManager.ErrorMetrics().GetByDateRange(ctx, startDate, endDate)
	}

	if err != nil {
		return nil, err
	}

	return &GetErrorMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

func (s *analyticsService) TrackError(ctx context.Context, req *TrackErrorRequest) error {
	return s.repoManager.ErrorMetrics().IncrementErrorCount(ctx, req.ErrorType, req.ErrorMessage, req.Service, req.Endpoint)
}

func (s *analyticsService) GetTopErrors(ctx context.Context, startDate, endDate time.Time, limit int) (*GetErrorMetricsResponse, error) {
	metrics, err := s.repoManager.ErrorMetrics().GetTopErrors(ctx, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}

	return &GetErrorMetricsResponse{
		Metrics: metrics,
		Total:   int64(len(metrics)),
	}, nil
}

func (s *analyticsService) GetErrorStats(ctx context.Context, startDate, endDate time.Time) (map[string]int64, error) {
	metrics, err := s.repoManager.ErrorMetrics().GetByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, metric := range metrics {
		stats[metric.ErrorType] += metric.Count
	}

	return stats, nil
}

// Report Service Methods

func (s *analyticsService) GenerateReport(ctx context.Context, req *GenerateReportRequest, userID uuid.UUID) (*domain.Report, error) {
	report := &domain.Report{
		ID:          uuid.New(),
		GeneratedBy: userID,
		Type:        req.ReportType,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Filters:     req.Filters,
		Status:      domain.ReportStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Generate report data based on type
	var data interface{}
	var err error

	switch req.ReportType {
	case domain.ReportTypeUserActivity:
		data, err = s.generateUserActivityReport(ctx, req.StartDate, req.EndDate, req.Filters)
	case domain.ReportTypeSystemMetrics:
		data, err = s.generateSystemMetricsReport(ctx, req.StartDate, req.EndDate, req.Filters)
	case domain.ReportTypeFileMetrics:
		data, err = s.generateFileMetricsReport(ctx, req.StartDate, req.EndDate, req.Filters)
	case domain.ReportTypeAPIMetrics:
		data, err = s.generateAPIMetricsReport(ctx, req.StartDate, req.EndDate, req.Filters)
	case domain.ReportTypeErrorMetrics:
		data, err = s.generateErrorMetricsReport(ctx, req.StartDate, req.EndDate, req.Filters)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", req.ReportType)
	}

	if err != nil {
		report.Status = domain.ReportStatusFailed
		report.UpdatedAt = time.Now()
		s.repoManager.Report().Create(ctx, report)
		return nil, err
	}

	// Convert data to JSON
	dataBytes, err := json.Marshal(data)
	if err != nil {
		report.Status = domain.ReportStatusFailed
		report.UpdatedAt = time.Now()
		s.repoManager.Report().Create(ctx, report)
		return nil, err
	}

	report.Data = dataBytes
	report.Status = domain.ReportStatusCompleted
	now := time.Now()
	report.CompletedAt = &now
	report.UpdatedAt = time.Now()

	err = s.repoManager.Report().Create(ctx, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s *analyticsService) GetReports(ctx context.Context, req *GetReportsRequest) (*GetReportsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var reports []*domain.Report
	var err error

	if req.UserID != nil {
		reports, err = s.repoManager.Report().GetByUser(ctx, *req.UserID, limit, offset)
	} else if req.ReportType != nil {
		reports, err = s.repoManager.Report().GetByType(ctx, *req.ReportType, limit, offset)
	} else {
		return nil, fmt.Errorf("invalid request: must specify user_id or report_type")
	}

	if err != nil {
		return nil, err
	}

	return &GetReportsResponse{
		Reports: reports,
		Total:   int64(len(reports)),
	}, nil
}

func (s *analyticsService) GetReportByID(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	return s.repoManager.Report().GetByID(ctx, id)
}

func (s *analyticsService) DeleteReport(ctx context.Context, id uuid.UUID) error {
	return s.repoManager.Report().Delete(ctx, id)
}

func (s *analyticsService) ScheduleReport(ctx context.Context, req *GenerateReportRequest, userID uuid.UUID, schedule string) error {
	// This would typically integrate with a job scheduler
	// For now, just create a pending report
	report := &domain.Report{
		ID:          uuid.New(),
		GeneratedBy: userID,
		Type:        req.ReportType,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Filters:     req.Filters,
		Status:      domain.ReportStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.repoManager.Report().Create(ctx, report)
}

// Dashboard Service Methods

func (s *analyticsService) GetDashboardData(ctx context.Context, startDate, endDate time.Time) (*DashboardData, error) {
	// Get system health
	systemHealth, _ := s.GetLatestSystemMetrics(ctx)

	// Get top files
	topFiles, _ := s.GetTopFiles(ctx, "views", 5)

	// Get top endpoints
	topEndpoints, _ := s.GetTopEndpoints(ctx, startDate, endDate, 5)

	// Get recent errors
	recentErrors, _ := s.GetTopErrors(ctx, startDate, endDate, 5)

	// Get event distribution
	eventStats, _ := s.GetEventStats(ctx, startDate, endDate)

	dashboard := &DashboardData{
		SystemHealth:      systemHealth,
		EventDistribution: eventStats,
	}

	if topFiles != nil {
		dashboard.TopFiles = topFiles.Metrics
	}
	if topEndpoints != nil {
		dashboard.TopEndpoints = topEndpoints.Metrics
	}
	if recentErrors != nil {
		dashboard.RecentErrors = recentErrors.Metrics
	}

	return dashboard, nil
}

func (s *analyticsService) GetUserDashboard(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*DashboardData, error) {
	// Get user activity
	userActivity, _ := s.GetUserActivity(ctx, &GetUserActivityRequest{
		UserID:    userID,
		StartDate: &startDate,
		EndDate:   &endDate,
	})

	// Get user events
	userEvents, _ := s.GetEventsByUser(ctx, userID, 10, 0)

	dashboard := &DashboardData{}

	if userActivity != nil {
		dashboard.UserActivity = userActivity.Activities
	}
	if userEvents != nil {
		dashboard.TotalEvents = userEvents.Total
	}

	return dashboard, nil
}

func (s *analyticsService) GetRealTimeMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get latest system metrics
	systemMetrics, err := s.GetLatestSystemMetrics(ctx)
	if err == nil {
		metrics["system"] = systemMetrics
	}

	// Get recent events (last hour)
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)
	recentEvents, err := s.GetEvents(ctx, &GetEventsRequest{
		StartDate: &startTime,
		EndDate:   &endTime,
		Limit:     100,
	})
	if err == nil {
		metrics["recent_events"] = len(recentEvents.Events)
	}

	return metrics, nil
}

// Helper methods for report generation

func (s *analyticsService) generateUserActivityReport(ctx context.Context, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	activities, err := s.repoManager.UserActivity().GetAggregatedByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return activities, nil
}

func (s *analyticsService) generateSystemMetricsReport(ctx context.Context, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	metrics, err := s.repoManager.SystemMetrics().GetByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (s *analyticsService) generateFileMetricsReport(ctx context.Context, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	topFiles, err := s.repoManager.FileMetrics().GetTopFiles(ctx, "views", 100)
	if err != nil {
		return nil, err
	}
	return topFiles, nil
}

func (s *analyticsService) generateAPIMetricsReport(ctx context.Context, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	metrics, err := s.repoManager.APIMetrics().GetByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (s *analyticsService) generateErrorMetricsReport(ctx context.Context, startDate, endDate time.Time, filters map[string]interface{}) (interface{}, error) {
	metrics, err := s.repoManager.ErrorMetrics().GetByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}