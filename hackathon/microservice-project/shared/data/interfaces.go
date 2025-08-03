package data

import (
	"context"
	"time"
)

// Repository defines the base interface for all repositories
type Repository interface {
	// Health checks the health of the repository
	Health(ctx context.Context) error
	
	// Close closes the repository connection
	Close() error
}

// TransactionManager defines the interface for managing database transactions
type TransactionManager interface {
	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context) (Transaction, error)
	
	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(tx Transaction) error) error
}

// Transaction defines the interface for database transactions
type Transaction interface {
	// Commit commits the transaction
	Commit() error
	
	// Rollback rolls back the transaction
	Rollback() error
	
	// Context returns the transaction context
	Context() context.Context
}

// QueryBuilder defines the interface for building database queries
type QueryBuilder interface {
	// Select adds SELECT clause
	Select(columns ...string) QueryBuilder
	
	// From adds FROM clause
	From(table string) QueryBuilder
	
	// Where adds WHERE clause
	Where(condition string, args ...interface{}) QueryBuilder
	
	// Join adds JOIN clause
	Join(table string, condition string) QueryBuilder
	
	// LeftJoin adds LEFT JOIN clause
	LeftJoin(table string, condition string) QueryBuilder
	
	// OrderBy adds ORDER BY clause
	OrderBy(column string, direction string) QueryBuilder
	
	// Limit adds LIMIT clause
	Limit(limit int) QueryBuilder
	
	// Offset adds OFFSET clause
	Offset(offset int) QueryBuilder
	
	// GroupBy adds GROUP BY clause
	GroupBy(columns ...string) QueryBuilder
	
	// Having adds HAVING clause
	Having(condition string, args ...interface{}) QueryBuilder
	
	// Build builds the final query
	Build() (string, []interface{}, error)
}

// DatabaseConnection defines the interface for database connections
type DatabaseConnection interface {
	// Query executes a query and returns rows
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	
	// QueryRow executes a query and returns a single row
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
	
	// Exec executes a query without returning rows
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	
	// Prepare prepares a statement
	Prepare(ctx context.Context, query string) (Statement, error)
	
	// Begin starts a transaction
	Begin(ctx context.Context) (Transaction, error)
	
	// Ping pings the database
	Ping(ctx context.Context) error
	
	// Close closes the connection
	Close() error
}

// Rows defines the interface for query result rows
type Rows interface {
	// Next advances to the next row
	Next() bool
	
	// Scan scans the current row into variables
	Scan(dest ...interface{}) error
	
	// Close closes the rows
	Close() error
	
	// Err returns any error encountered during iteration
	Err() error
	
	// Columns returns the column names
	Columns() ([]string, error)
}

// Row defines the interface for a single query result row
type Row interface {
	// Scan scans the row into variables
	Scan(dest ...interface{}) error
}

// Result defines the interface for query execution results
type Result interface {
	// LastInsertId returns the last inserted ID
	LastInsertId() (int64, error)
	
	// RowsAffected returns the number of affected rows
	RowsAffected() (int64, error)
}

// Statement defines the interface for prepared statements
type Statement interface {
	// Query executes the statement with query
	Query(ctx context.Context, args ...interface{}) (Rows, error)
	
	// QueryRow executes the statement and returns a single row
	QueryRow(ctx context.Context, args ...interface{}) Row
	
	// Exec executes the statement
	Exec(ctx context.Context, args ...interface{}) (Result, error)
	
	// Close closes the statement
	Close() error
}

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	Repository
	
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) ([]byte, error)
	
	// Set stores a value in cache
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	
	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error
	
	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)
	
	// Expire sets expiration for a key
	Expire(ctx context.Context, key string, expiration time.Duration) error
	
	// Keys returns all keys matching a pattern
	Keys(ctx context.Context, pattern string) ([]string, error)
	
	// FlushAll removes all keys from cache
	FlushAll(ctx context.Context) error
	
	// Increment increments a numeric value
	Increment(ctx context.Context, key string) (int64, error)
	
	// Decrement decrements a numeric value
	Decrement(ctx context.Context, key string) (int64, error)
	
	// SetNX sets a value only if key doesn't exist
	SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error)
}

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	// Upload uploads a file
	Upload(ctx context.Context, path string, data []byte, metadata map[string]string) error
	
	// UploadStream uploads a file from a stream
	UploadStream(ctx context.Context, path string, reader interface{}, size int64, metadata map[string]string) error
	
	// Download downloads a file
	Download(ctx context.Context, path string) ([]byte, error)
	
	// DownloadStream downloads a file as a stream
	DownloadStream(ctx context.Context, path string) (interface{}, error)
	
	// Delete deletes a file
	Delete(ctx context.Context, path string) error
	
	// Exists checks if a file exists
	Exists(ctx context.Context, path string) (bool, error)
	
	// GetMetadata retrieves file metadata
	GetMetadata(ctx context.Context, path string) (map[string]string, error)
	
	// SetMetadata sets file metadata
	SetMetadata(ctx context.Context, path string, metadata map[string]string) error
	
	// List lists files in a directory
	List(ctx context.Context, prefix string) ([]FileInfo, error)
	
	// Copy copies a file
	Copy(ctx context.Context, srcPath, dstPath string) error
	
	// Move moves a file
	Move(ctx context.Context, srcPath, dstPath string) error
	
	// GetSignedURL generates a signed URL for file access
	GetSignedURL(ctx context.Context, path string, expiration time.Duration, method string) (string, error)
}

// FileInfo represents file information
type FileInfo struct {
	Path         string            `json:"path"`
	Size         int64             `json:"size"`
	LastModified time.Time         `json:"last_modified"`
	ContentType  string            `json:"content_type"`
	Metadata     map[string]string `json:"metadata"`
	ETag         string            `json:"etag"`
}

// SearchRepository defines the interface for search operations
type SearchRepository interface {
	Repository
	
	// Index indexes a document
	Index(ctx context.Context, index string, id string, document interface{}) error
	
	// Search searches for documents
	Search(ctx context.Context, index string, query SearchQuery) (*SearchResult, error)
	
	// Get retrieves a document by ID
	Get(ctx context.Context, index string, id string) (interface{}, error)
	
	// Update updates a document
	Update(ctx context.Context, index string, id string, document interface{}) error
	
	// Delete deletes a document
	Delete(ctx context.Context, index string, id string) error
	
	// BulkIndex indexes multiple documents
	BulkIndex(ctx context.Context, index string, documents []IndexDocument) error
	
	// CreateIndex creates an index
	CreateIndex(ctx context.Context, index string, mapping interface{}) error
	
	// DeleteIndex deletes an index
	DeleteIndex(ctx context.Context, index string) error
}

// SearchQuery represents a search query
type SearchQuery struct {
	Query      interface{}       `json:"query"`
	Filters    []Filter          `json:"filters,omitempty"`
	Sort       []SortField       `json:"sort,omitempty"`
	From       int               `json:"from,omitempty"`
	Size       int               `json:"size,omitempty"`
	Highlight  *Highlight        `json:"highlight,omitempty"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
}

// Filter represents a search filter
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "eq", "ne", "gt", "gte", "lt", "lte", "in", "nin", "exists"
	Value    interface{} `json:"value"`
}

// SortField represents a sort field
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc", "desc"
}

// Highlight represents search highlighting
type Highlight struct {
	Fields map[string]interface{} `json:"fields"`
}

// SearchResult represents search results
type SearchResult struct {
	Total    int64                  `json:"total"`
	Hits     []SearchHit            `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Took     int64                  `json:"took"`
}

// SearchHit represents a search hit
type SearchHit struct {
	ID        string                 `json:"id"`
	Score     float64                `json:"score"`
	Source    interface{}            `json:"source"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
}

// IndexDocument represents a document to be indexed
type IndexDocument struct {
	ID       string      `json:"id"`
	Document interface{} `json:"document"`
}

// TimeSeriesRepository defines the interface for time series data operations
type TimeSeriesRepository interface {
	Repository
	
	// WritePoint writes a data point
	WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error
	
	// WritePoints writes multiple data points
	WritePoints(ctx context.Context, points []TimeSeriesPoint) error
	
	// Query queries time series data
	Query(ctx context.Context, query string) (*TimeSeriesResult, error)
	
	// QueryRange queries time series data within a time range
	QueryRange(ctx context.Context, measurement string, start, end time.Time, filters map[string]string) (*TimeSeriesResult, error)
}

// TimeSeriesPoint represents a time series data point
type TimeSeriesPoint struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Timestamp   time.Time              `json:"timestamp"`
}

// TimeSeriesResult represents time series query results
type TimeSeriesResult struct {
	Series []TimeSeries `json:"series"`
}

// TimeSeries represents a time series
type TimeSeries struct {
	Name    string          `json:"name"`
	Tags    map[string]string `json:"tags"`
	Columns []string        `json:"columns"`
	Values  [][]interface{} `json:"values"`
}

// DataSourceConfig holds configuration for data sources
type DataSourceConfig struct {
	Type               string                 `json:"type"` // "postgres", "mysql", "redis", "s3", "elasticsearch", etc.
	ConnectionString   string                 `json:"connection_string"`
	Host               string                 `json:"host"`
	Port               int                    `json:"port"`
	Database           string                 `json:"database"`
	Username           string                 `json:"username"`
	Password           string                 `json:"password"`
	SSLMode            string                 `json:"ssl_mode"`
	MaxConnections     int                    `json:"max_connections"`
	MaxIdleConnections int                    `json:"max_idle_connections"`
	ConnectionTimeout  time.Duration          `json:"connection_timeout"`
	QueryTimeout       time.Duration          `json:"query_timeout"`
	RetryPolicy        *RetryPolicy           `json:"retry_policy"`
	AdditionalSettings map[string]interface{} `json:"additional_settings"`
}

// RetryPolicy defines the retry policy for failed operations
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DataSourceFactory creates data sources based on configuration
type DataSourceFactory interface {
	// CreateDatabaseConnection creates a database connection
	CreateDatabaseConnection(config *DataSourceConfig) (DatabaseConnection, error)
	
	// CreateCacheRepository creates a cache repository
	CreateCacheRepository(config *DataSourceConfig) (CacheRepository, error)
	
	// CreateFileStorage creates a file storage
	CreateFileStorage(config *DataSourceConfig) (FileStorage, error)
	
	// CreateSearchRepository creates a search repository
	CreateSearchRepository(config *DataSourceConfig) (SearchRepository, error)
	
	// CreateTimeSeriesRepository creates a time series repository
	CreateTimeSeriesRepository(config *DataSourceConfig) (TimeSeriesRepository, error)
	
	// CreateTransactionManager creates a transaction manager
	CreateTransactionManager(config *DataSourceConfig) (TransactionManager, error)
	
	// CreateQueryBuilder creates a query builder
	CreateQueryBuilder(dialectType string) QueryBuilder
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
	Limit    int `json:"limit"`
}

// PaginatedResult represents paginated results
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// Metrics defines the interface for data access metrics
type Metrics interface {
	// IncrementQueries increments the query counter
	IncrementQueries(operation string, table string)
	
	// IncrementErrors increments the error counter
	IncrementErrors(operation string, errorType string)
	
	// RecordQueryDuration records query execution time
	RecordQueryDuration(operation string, duration time.Duration)
	
	// RecordConnectionPoolStats records connection pool statistics
	RecordConnectionPoolStats(active, idle, total int)
	
	// RecordCacheHitRatio records cache hit ratio
	RecordCacheHitRatio(operation string, hitRatio float64)
}

// Logger defines the interface for logging
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields ...interface{})
	
	// Info logs an info message
	Info(msg string, fields ...interface{})
	
	// Warn logs a warning message
	Warn(msg string, fields ...interface{})
	
	// Error logs an error message
	Error(msg string, fields ...interface{})
	
	// WithFields returns a logger with additional fields
	WithFields(fields map[string]interface{}) Logger
}