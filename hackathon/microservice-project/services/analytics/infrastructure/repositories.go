package infrastructure

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/elotusteam/microservice-project/services/analytics/domain"
)

// MockEventRepository implements domain.EventRepository for testing/demo purposes
type MockEventRepository struct {
	events map[uuid.UUID]*domain.Event
}

func NewMockEventRepository() domain.EventRepository {
	return &MockEventRepository{
		events: make(map[uuid.UUID]*domain.Event),
	}
}

func (r *MockEventRepository) Create(ctx context.Context, event *domain.Event) error {
	r.events[event.ID] = event
	return nil
}

func (r *MockEventRepository) CreateBatch(ctx context.Context, events []*domain.Event) error {
	for _, event := range events {
		r.events[event.ID] = event
	}
	return nil
}

func (r *MockEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	if event, exists := r.events[id]; exists {
		return event, nil
	}
	return nil, nil
}

func (r *MockEventRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Event, error) {
	var result []*domain.Event
	for _, event := range r.events {
		if event.UserID != nil && *event.UserID == userID {
			result = append(result, event)
		}
	}
	return result, nil
}

func (r *MockEventRepository) GetByType(ctx context.Context, eventType domain.EventType, limit, offset int) ([]*domain.Event, error) {
	var result []*domain.Event
	for _, event := range r.events {
		if event.Type == eventType {
			result = append(result, event)
		}
	}
	return result, nil
}

func (r *MockEventRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*domain.Event, error) {
	var result []*domain.Event
	for _, event := range r.events {
		if event.CreatedAt.After(startDate) && event.CreatedAt.Before(endDate) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (r *MockEventRepository) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*domain.Event, error) {
	var result []*domain.Event
	for _, event := range r.events {
		if event.UserID != nil && *event.UserID == userID && event.CreatedAt.After(startDate) && event.CreatedAt.Before(endDate) {
			result = append(result, event)
		}
	}
	return result, nil
}

func (r *MockEventRepository) CountByType(ctx context.Context, eventType domain.EventType, startDate, endDate time.Time) (int64, error) {
	var count int64
	for _, event := range r.events {
		if event.Type == eventType && event.CreatedAt.After(startDate) && event.CreatedAt.Before(endDate) {
			count++
		}
	}
	return count, nil
}

func (r *MockEventRepository) CountByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (int64, error) {
	var count int64
	for _, event := range r.events {
		if event.UserID != nil && *event.UserID == userID && event.CreatedAt.After(startDate) && event.CreatedAt.Before(endDate) {
			count++
		}
	}
	return count, nil
}

func (r *MockEventRepository) DeleteOlderThan(ctx context.Context, date time.Time) error {
	for id, event := range r.events {
		if event.CreatedAt.Before(date) {
			delete(r.events, id)
		}
	}
	return nil
}

// MockUserActivityRepository implements domain.UserActivityRepository
type MockUserActivityRepository struct {
	activities map[uuid.UUID]*domain.UserActivity
}

func NewMockUserActivityRepository() domain.UserActivityRepository {
	return &MockUserActivityRepository{
		activities: make(map[uuid.UUID]*domain.UserActivity),
	}
}

func (r *MockUserActivityRepository) Create(ctx context.Context, activity *domain.UserActivity) error {
	r.activities[activity.ID] = activity
	return nil
}

func (r *MockUserActivityRepository) Update(ctx context.Context, activity *domain.UserActivity) error {
	r.activities[activity.ID] = activity
	return nil
}

func (r *MockUserActivityRepository) GetByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*domain.UserActivity, error) {
	for _, activity := range r.activities {
		if activity.UserID == userID && activity.Date.Equal(date) {
			return activity, nil
		}
	}
	return nil, nil
}

func (r *MockUserActivityRepository) GetByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*domain.UserActivity, error) {
	var result []*domain.UserActivity
	for _, activity := range r.activities {
		if activity.UserID == userID && activity.Date.After(startDate) && activity.Date.Before(endDate) {
			result = append(result, activity)
		}
	}
	return result, nil
}

func (r *MockUserActivityRepository) GetTopActiveUsers(ctx context.Context, startDate, endDate time.Time, limit int) ([]*domain.UserActivity, error) {
	var result []*domain.UserActivity
	for _, activity := range r.activities {
		if activity.Date.After(startDate) && activity.Date.Before(endDate) {
			result = append(result, activity)
		}
	}
	return result, nil
}

func (r *MockUserActivityRepository) GetAggregatedByDateRange(ctx context.Context, startDate, endDate time.Time) (*domain.UserActivity, error) {
	aggregated := &domain.UserActivity{
		ID:   uuid.New(),
		Date: time.Now(),
	}
	return aggregated, nil
}

// MockSystemMetricsRepository implements domain.SystemMetricsRepository
type MockSystemMetricsRepository struct {
	metrics map[uuid.UUID]*domain.SystemMetrics
}

func NewMockSystemMetricsRepository() domain.SystemMetricsRepository {
	return &MockSystemMetricsRepository{
		metrics: make(map[uuid.UUID]*domain.SystemMetrics),
	}
}

func (r *MockSystemMetricsRepository) Create(ctx context.Context, metrics *domain.SystemMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockSystemMetricsRepository) Update(ctx context.Context, metrics *domain.SystemMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockSystemMetricsRepository) GetByDate(ctx context.Context, date time.Time) (*domain.SystemMetrics, error) {
	for _, metric := range r.metrics {
		if metric.Date.Equal(date) {
			return metric, nil
		}
	}
	return nil, nil
}

func (r *MockSystemMetricsRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.SystemMetrics, error) {
	var result []*domain.SystemMetrics
	for _, metric := range r.metrics {
		if metric.Date.After(startDate) && metric.Date.Before(endDate) {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (r *MockSystemMetricsRepository) GetLatest(ctx context.Context) (*domain.SystemMetrics, error) {
	var latest *domain.SystemMetrics
	for _, metric := range r.metrics {
		if latest == nil || metric.CreatedAt.After(latest.CreatedAt) {
			latest = metric
		}
	}
	return latest, nil
}

// MockFileMetricsRepository implements domain.FileMetricsRepository
type MockFileMetricsRepository struct {
	metrics map[uuid.UUID]*domain.FileMetrics
}

func NewMockFileMetricsRepository() domain.FileMetricsRepository {
	return &MockFileMetricsRepository{
		metrics: make(map[uuid.UUID]*domain.FileMetrics),
	}
}

func (r *MockFileMetricsRepository) Create(ctx context.Context, metrics *domain.FileMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockFileMetricsRepository) Update(ctx context.Context, metrics *domain.FileMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockFileMetricsRepository) GetByFileID(ctx context.Context, fileID uuid.UUID) (*domain.FileMetrics, error) {
	for _, metric := range r.metrics {
		if metric.FileID == fileID {
			return metric, nil
		}
	}
	return nil, nil
}

func (r *MockFileMetricsRepository) GetByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.FileMetrics, error) {
	var result []*domain.FileMetrics
	for _, metric := range r.metrics {
		if metric.OwnerID == ownerID {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (r *MockFileMetricsRepository) GetTopFiles(ctx context.Context, metric string, limit int) ([]*domain.FileMetrics, error) {
	var result []*domain.FileMetrics
	for _, m := range r.metrics {
		result = append(result, m)
	}
	return result, nil
}

func (r *MockFileMetricsRepository) IncrementViewCount(ctx context.Context, fileID uuid.UUID) error {
	if metric, exists := r.metrics[fileID]; exists {
		metric.ViewCount++
	}
	return nil
}

func (r *MockFileMetricsRepository) IncrementDownloadCount(ctx context.Context, fileID uuid.UUID) error {
	if metric, exists := r.metrics[fileID]; exists {
		metric.DownloadCount++
	}
	return nil
}

func (r *MockFileMetricsRepository) IncrementShareCount(ctx context.Context, fileID uuid.UUID) error {
	if metric, exists := r.metrics[fileID]; exists {
		metric.ShareCount++
	}
	return nil
}

func (r *MockFileMetricsRepository) UpdateLastAccessed(ctx context.Context, fileID uuid.UUID, accessTime time.Time) error {
	if metric, exists := r.metrics[fileID]; exists {
		metric.LastAccessed = &accessTime
	}
	return nil
}

// MockAPIMetricsRepository implements domain.APIMetricsRepository
type MockAPIMetricsRepository struct {
	metrics map[uuid.UUID]*domain.APIMetrics
}

func NewMockAPIMetricsRepository() domain.APIMetricsRepository {
	return &MockAPIMetricsRepository{
		metrics: make(map[uuid.UUID]*domain.APIMetrics),
	}
}

func (r *MockAPIMetricsRepository) Create(ctx context.Context, metrics *domain.APIMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockAPIMetricsRepository) Update(ctx context.Context, metrics *domain.APIMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockAPIMetricsRepository) GetByEndpointAndDate(ctx context.Context, endpoint, method string, date time.Time) (*domain.APIMetrics, error) {
	for _, metric := range r.metrics {
		if metric.Endpoint == endpoint && metric.Method == method && metric.Date.Equal(date) {
			return metric, nil
		}
	}
	return nil, nil
}

func (r *MockAPIMetricsRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.APIMetrics, error) {
	var result []*domain.APIMetrics
	for _, metric := range r.metrics {
		if metric.Date.After(startDate) && metric.Date.Before(endDate) {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (r *MockAPIMetricsRepository) GetTopEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) ([]*domain.APIMetrics, error) {
	var result []*domain.APIMetrics
	for _, metric := range r.metrics {
		if metric.Date.After(startDate) && metric.Date.Before(endDate) {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (r *MockAPIMetricsRepository) GetSlowestEndpoints(ctx context.Context, startDate, endDate time.Time, limit int) ([]*domain.APIMetrics, error) {
	var result []*domain.APIMetrics
	for _, metric := range r.metrics {
		if metric.Date.After(startDate) && metric.Date.Before(endDate) {
			result = append(result, metric)
		}
	}
	return result, nil
}

// MockErrorMetricsRepository implements domain.ErrorMetricsRepository
type MockErrorMetricsRepository struct {
	metrics map[uuid.UUID]*domain.ErrorMetrics
}

func NewMockErrorMetricsRepository() domain.ErrorMetricsRepository {
	return &MockErrorMetricsRepository{
		metrics: make(map[uuid.UUID]*domain.ErrorMetrics),
	}
}

func (r *MockErrorMetricsRepository) Create(ctx context.Context, metrics *domain.ErrorMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockErrorMetricsRepository) Update(ctx context.Context, metrics *domain.ErrorMetrics) error {
	r.metrics[metrics.ID] = metrics
	return nil
}

func (r *MockErrorMetricsRepository) GetByTypeAndDate(ctx context.Context, errorType string, date time.Time) (*domain.ErrorMetrics, error) {
	for _, metric := range r.metrics {
		if metric.ErrorType == errorType && metric.Date.Equal(date) {
			return metric, nil
		}
	}
	return nil, nil
}

func (r *MockErrorMetricsRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*domain.ErrorMetrics, error) {
	var result []*domain.ErrorMetrics
	for _, metric := range r.metrics {
		if metric.Date.After(startDate) && metric.Date.Before(endDate) {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (r *MockErrorMetricsRepository) GetTopErrors(ctx context.Context, startDate, endDate time.Time, limit int) ([]*domain.ErrorMetrics, error) {
	var result []*domain.ErrorMetrics
	for _, metric := range r.metrics {
		if metric.Date.After(startDate) && metric.Date.Before(endDate) {
			result = append(result, metric)
		}
	}
	return result, nil
}

func (r *MockErrorMetricsRepository) IncrementErrorCount(ctx context.Context, errorType, errorMessage, service, endpoint string) error {
	// Find existing metric or create new one
	for _, metric := range r.metrics {
		if metric.ErrorType == errorType && metric.ErrorMessage == errorMessage {
			metric.Count++
			metric.LastSeen = time.Now()
			return nil
		}
	}
	// Create new metric
	newMetric := &domain.ErrorMetrics{
		ID:           uuid.New(),
		ErrorType:    errorType,
		ErrorMessage: errorMessage,
		Service:      service,
		Endpoint:     endpoint,
		Date:         time.Now(),
		Count:        1,
		FirstSeen:    time.Now(),
		LastSeen:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	r.metrics[newMetric.ID] = newMetric
	return nil
}

// MockReportRepository implements domain.ReportRepository
type MockReportRepository struct {
	reports map[uuid.UUID]*domain.Report
}

func NewMockReportRepository() domain.ReportRepository {
	return &MockReportRepository{
		reports: make(map[uuid.UUID]*domain.Report),
	}
}

func (r *MockReportRepository) Create(ctx context.Context, report *domain.Report) error {
	r.reports[report.ID] = report
	return nil
}

func (r *MockReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Report, error) {
	if report, exists := r.reports[id]; exists {
		return report, nil
	}
	return nil, nil
}

func (r *MockReportRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Report, error) {
	var result []*domain.Report
	for _, report := range r.reports {
		if report.GeneratedBy == userID {
			result = append(result, report)
		}
	}
	return result, nil
}

func (r *MockReportRepository) GetByType(ctx context.Context, reportType domain.ReportType, limit, offset int) ([]*domain.Report, error) {
	var result []*domain.Report
	for _, report := range r.reports {
		if report.Type == reportType {
			result = append(result, report)
		}
	}
	return result, nil
}

func (r *MockReportRepository) Update(ctx context.Context, report *domain.Report) error {
	r.reports[report.ID] = report
	return nil
}

func (r *MockReportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.reports, id)
	return nil
}

// MockRepositoryManager implements domain.RepositoryManager
type MockRepositoryManager struct {
	eventRepo        domain.EventRepository
	userActivityRepo domain.UserActivityRepository
	systemMetricsRepo domain.SystemMetricsRepository
	fileMetricsRepo  domain.FileMetricsRepository
	apiMetricsRepo   domain.APIMetricsRepository
	errorMetricsRepo domain.ErrorMetricsRepository
	reportRepo       domain.ReportRepository
}

func NewMockRepositoryManager() domain.RepositoryManager {
	return &MockRepositoryManager{
		eventRepo:        NewMockEventRepository(),
		userActivityRepo: NewMockUserActivityRepository(),
		systemMetricsRepo: NewMockSystemMetricsRepository(),
		fileMetricsRepo:  NewMockFileMetricsRepository(),
		apiMetricsRepo:   NewMockAPIMetricsRepository(),
		errorMetricsRepo: NewMockErrorMetricsRepository(),
		reportRepo:       NewMockReportRepository(),
	}
}

func (rm *MockRepositoryManager) Event() domain.EventRepository {
	return rm.eventRepo
}

func (rm *MockRepositoryManager) UserActivity() domain.UserActivityRepository {
	return rm.userActivityRepo
}

func (rm *MockRepositoryManager) SystemMetrics() domain.SystemMetricsRepository {
	return rm.systemMetricsRepo
}

func (rm *MockRepositoryManager) FileMetrics() domain.FileMetricsRepository {
	return rm.fileMetricsRepo
}

func (rm *MockRepositoryManager) APIMetrics() domain.APIMetricsRepository {
	return rm.apiMetricsRepo
}

func (rm *MockRepositoryManager) ErrorMetrics() domain.ErrorMetricsRepository {
	return rm.errorMetricsRepo
}

func (rm *MockRepositoryManager) Report() domain.ReportRepository {
	return rm.reportRepo
}

func (rm *MockRepositoryManager) BeginTx(ctx context.Context) (domain.RepositoryManager, error) {
	// For mock implementation, return the same instance
	return rm, nil
}

func (rm *MockRepositoryManager) Commit() error {
	// Mock implementation - no-op
	return nil
}

func (rm *MockRepositoryManager) Rollback() error {
	// Mock implementation - no-op
	return nil
}