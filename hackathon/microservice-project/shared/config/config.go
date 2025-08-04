package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`
	Debug       bool   `json:"debug"`

	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Cache configuration
	Cache CacheConfig `json:"cache"`

	// Storage configuration
	Storage StorageConfig `json:"storage"`

	// Message queue configuration
	MessageQueue MessageQueueConfig `json:"message_queue"`

	// Search configuration
	Search SearchConfig `json:"search"`

	// Monitoring configuration
	Monitoring MonitoringConfig `json:"monitoring"`

	// Security configuration
	Security SecurityConfig `json:"security"`

	// Rate limiting configuration
	RateLimit RateLimitConfig `json:"rate_limit"`

	// File upload configuration
	FileUpload FileUploadConfig `json:"file_upload"`

	// Image processing configuration
	ImageProcessing ImageProcessingConfig `json:"image_processing"`

	// Notification configuration
	Notification NotificationConfig `json:"notification"`

	// External services configuration
	ExternalServices ExternalServicesConfig `json:"external_services"`

	// Microservices configuration
	Services ServicesConfig `json:"services"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	TLSEnabled      bool          `json:"tls_enabled"`
	TLSCertFile     string        `json:"tls_cert_file"`
	TLSKeyFile      string        `json:"tls_key_file"`
	CORSEnabled     bool          `json:"cors_enabled"`
	CORSOrigins     []string      `json:"cors_origins"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver             string        `json:"driver"`
	Host               string        `json:"host"`
	Port               int           `json:"port"`
	Database           string        `json:"database"`
	Username           string        `json:"username"`
	Password           string        `json:"-"` // Hidden from JSON
	SSLMode            string        `json:"ssl_mode"`
	MaxOpenConnections int           `json:"max_open_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
	ConnectionLifetime time.Duration `json:"connection_lifetime"`
	ConnectionTimeout  time.Duration `json:"connection_timeout"`
	QueryTimeout       time.Duration `json:"query_timeout"`
	MigrationsPath     string        `json:"migrations_path"`
	AutoMigrate        bool          `json:"auto_migrate"`

	// Read replicas configuration
	ReadReplicas []DatabaseReplicaConfig `json:"read_replicas"`
}

// DatabaseReplicaConfig holds database replica configuration
type DatabaseReplicaConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"-"`      // Hidden from JSON
	Weight   int    `json:"weight"` // Load balancing weight
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Driver       string        `json:"driver"` // redis, memcached, in-memory
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"-"` // Hidden from JSON
	Database     int           `json:"database"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	DefaultTTL   time.Duration `json:"default_ttl"`

	// Cluster configuration for Redis Cluster
	Cluster CacheClusterConfig `json:"cluster"`
}

// CacheClusterConfig holds cache cluster configuration
type CacheClusterConfig struct {
	Enabled   bool     `json:"enabled"`
	Addresses []string `json:"addresses"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Driver    string `json:"driver"` // s3, gcs, azure, local
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"-"` // Hidden from JSON
	UseSSL    bool   `json:"use_ssl"`

	// Local storage configuration
	LocalPath string `json:"local_path"`

	// CDN configuration
	CDN CDNConfig `json:"cdn"`
}

// CDNConfig holds CDN configuration
type CDNConfig struct {
	Enabled    bool          `json:"enabled"`
	BaseURL    string        `json:"base_url"`
	SigningKey string        `json:"-"` // Hidden from JSON
	TTL        time.Duration `json:"ttl"`
}

// MessageQueueConfig holds message queue configuration
type MessageQueueConfig struct {
	Driver   string `json:"driver"` // rabbitmq, kafka, redis, sqs
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"-"` // Hidden from JSON
	VHost    string `json:"vhost"`

	// Kafka specific configuration
	Kafka KafkaConfig `json:"kafka"`

	// RabbitMQ specific configuration
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`

	// Redis specific configuration
	Redis RedisConfig `json:"redis"`

	// AWS SQS specific configuration
	SQS SQSConfig `json:"sqs"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers         []string      `json:"brokers"`
	GroupID         string        `json:"group_id"`
	RetryMax        int           `json:"retry_max"`
	RetryBackoff    time.Duration `json:"retry_backoff"`
	FlushTimeout    time.Duration `json:"flush_timeout"`
	BatchSize       int           `json:"batch_size"`
	CompressionType string        `json:"compression_type"`
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	Exchange      string `json:"exchange"`
	ExchangeType  string `json:"exchange_type"`
	Durable       bool   `json:"durable"`
	AutoDelete    bool   `json:"auto_delete"`
	PrefetchCount int    `json:"prefetch_count"`
}

// RedisConfig holds Redis configuration for message queue
type RedisConfig struct {
	StreamName    string        `json:"stream_name"`
	ConsumerGroup string        `json:"consumer_group"`
	BlockTime     time.Duration `json:"block_time"`
	MaxLen        int64         `json:"max_len"`
}

// SQSConfig holds AWS SQS configuration
type SQSConfig struct {
	Region            string `json:"region"`
	AccessKeyID       string `json:"access_key_id"`
	SecretAccessKey   string `json:"-"` // Hidden from JSON
	QueueURL          string `json:"queue_url"`
	VisibilityTimeout int    `json:"visibility_timeout"`
	WaitTimeSeconds   int    `json:"wait_time_seconds"`
}

// SearchConfig holds search engine configuration
type SearchConfig struct {
	Driver   string   `json:"driver"` // elasticsearch, opensearch, solr
	Hosts    []string `json:"hosts"`
	Username string   `json:"username"`
	Password string   `json:"-"` // Hidden from JSON
	Index    string   `json:"index"`
	Shards   int      `json:"shards"`
	Replicas int      `json:"replicas"`

	// Elasticsearch specific configuration
	Elasticsearch ElasticsearchConfig `json:"elasticsearch"`
}

// ElasticsearchConfig holds Elasticsearch configuration
type ElasticsearchConfig struct {
	Version       string        `json:"version"`
	Sniff         bool          `json:"sniff"`
	Healthcheck   bool          `json:"healthcheck"`
	RetryOnStatus []int         `json:"retry_on_status"`
	MaxRetries    int           `json:"max_retries"`
	RetryBackoff  time.Duration `json:"retry_backoff"`
	Timeout       time.Duration `json:"timeout"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Enabled bool `json:"enabled"`

	// Metrics configuration
	Metrics MetricsConfig `json:"metrics"`

	// Tracing configuration
	Tracing TracingConfig `json:"tracing"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`

	// Health check configuration
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled   bool          `json:"enabled"`
	Provider  string        `json:"provider"` // prometheus, datadog, newrelic
	Endpoint  string        `json:"endpoint"`
	Namespace string        `json:"namespace"`
	Interval  time.Duration `json:"interval"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled     bool    `json:"enabled"`
	Provider    string  `json:"provider"` // jaeger, zipkin, datadog
	Endpoint    string  `json:"endpoint"`
	ServiceName string  `json:"service_name"`
	SampleRate  float64 `json:"sample_rate"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"` // json, text
	Output     string `json:"output"` // stdout, file, syslog
	FilePath   string `json:"file_path"`
	MaxSize    int    `json:"max_size"` // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"` // days
	Compress   bool   `json:"compress"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Enabled  bool          `json:"enabled"`
	Endpoint string        `json:"endpoint"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	// JWT configuration
	JWT JWTConfig `json:"jwt"`

	// Password configuration
	Password PasswordConfig `json:"password"`

	// Encryption configuration
	Encryption EncryptionConfig `json:"encryption"`

	// API key configuration
	APIKey APIKeyConfig `json:"api_key"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string        `json:"-"` // Hidden from JSON
	Issuer          string        `json:"issuer"`
	Audience        string        `json:"audience"`
	AccessTokenTTL  time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL time.Duration `json:"refresh_token_ttl"`
	Algorithm       string        `json:"algorithm"`
	PublicKeyPath   string        `json:"public_key_path"`
	PrivateKeyPath  string        `json:"private_key_path"`
}

// PasswordConfig holds password configuration
type PasswordConfig struct {
	MinLength           int           `json:"min_length"`
	RequireUppercase    bool          `json:"require_uppercase"`
	RequireLowercase    bool          `json:"require_lowercase"`
	RequireNumbers      bool          `json:"require_numbers"`
	RequireSymbols      bool          `json:"require_symbols"`
	RequireSpecialChars bool          `json:"require_special_chars"`
	BcryptCost          int           `json:"bcrypt_cost"`
	ResetTokenTTL       time.Duration `json:"reset_token_ttl"`
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	Key       string `json:"-"` // Hidden from JSON
	Algorithm string `json:"algorithm"`
	KeySize   int    `json:"key_size"`
}

// APIKeyConfig holds API key configuration
type APIKeyConfig struct {
	Enabled    bool          `json:"enabled"`
	HeaderName string        `json:"header_name"`
	Prefix     string        `json:"prefix"`
	Length     int           `json:"length"`
	TTL        time.Duration `json:"ttl"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled bool `json:"enabled"`

	// Global rate limits
	Global RateLimitRule `json:"global"`

	// Per-user rate limits
	PerUser RateLimitRule `json:"per_user"`

	// Per-IP rate limits
	PerIP RateLimitRule `json:"per_ip"`

	// API-specific rate limits
	API map[string]RateLimitRule `json:"api"`
}

// RateLimitRule holds rate limit rule configuration
type RateLimitRule struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
	Burst    int           `json:"burst"`
}

// FileUploadConfig holds file upload configuration
type FileUploadConfig struct {
	MaxFileSize       int64         `json:"max_file_size"`  // bytes
	MaxTotalSize      int64         `json:"max_total_size"` // bytes per user
	AllowedMimeTypes  []string      `json:"allowed_mime_types"`
	AllowedExtensions []string      `json:"allowed_extensions"`
	ChunkSize         int64         `json:"chunk_size"` // bytes
	UploadTimeout     time.Duration `json:"upload_timeout"`
	TempDir           string        `json:"temp_dir"`
	VirusScanEnabled  bool          `json:"virus_scan_enabled"`
}

// ImageProcessingConfig holds image processing configuration
type ImageProcessingConfig struct {
	Enabled    bool              `json:"enabled"`
	MaxWidth   int               `json:"max_width"`
	MaxHeight  int               `json:"max_height"`
	Quality    int               `json:"quality"`
	Formats    []string          `json:"formats"`
	Thumbnails []ThumbnailConfig `json:"thumbnails"`
	Watermark  WatermarkConfig   `json:"watermark"`
	Workers    int               `json:"workers"`
	QueueSize  int               `json:"queue_size"`
}

// ThumbnailConfig holds thumbnail configuration
type ThumbnailConfig struct {
	Name    string `json:"name"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Quality int    `json:"quality"`
	Format  string `json:"format"`
}

// WatermarkConfig holds watermark configuration
type WatermarkConfig struct {
	Enabled   bool    `json:"enabled"`
	ImagePath string  `json:"image_path"`
	Position  string  `json:"position"` // top-left, top-right, bottom-left, bottom-right, center
	Opacity   float64 `json:"opacity"`
	Scale     float64 `json:"scale"`
}

// NotificationConfig holds notification configuration
type NotificationConfig struct {
	Enabled bool `json:"enabled"`

	// Email configuration
	Email EmailConfig `json:"email"`

	// SMS configuration
	SMS SMSConfig `json:"sms"`

	// Push notification configuration
	Push PushConfig `json:"push"`

	// WebSocket configuration
	WebSocket WebSocketConfig `json:"websocket"`
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"` // smtp, sendgrid, ses, mailgun
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"-"` // Hidden from JSON
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
	TLS       bool   `json:"tls"`
	APIKey    string `json:"-"` // Hidden from JSON
}

// SMSConfig holds SMS configuration
type SMSConfig struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"` // twilio, aws-sns, nexmo
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"-"` // Hidden from JSON
	FromNumber string `json:"from_number"`
}

// PushConfig holds push notification configuration
type PushConfig struct {
	Enabled bool `json:"enabled"`

	// Firebase configuration
	Firebase FirebaseConfig `json:"firebase"`

	// Apple Push Notification configuration
	APNS APNSConfig `json:"apns"`
}

// FirebaseConfig holds Firebase configuration
type FirebaseConfig struct {
	ProjectID       string `json:"project_id"`
	CredentialsPath string `json:"credentials_path"`
}

// APNSConfig holds Apple Push Notification configuration
type APNSConfig struct {
	KeyID      string `json:"key_id"`
	TeamID     string `json:"team_id"`
	BundleID   string `json:"bundle_id"`
	KeyPath    string `json:"key_path"`
	Production bool   `json:"production"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	Enabled    bool          `json:"enabled"`
	Path       string        `json:"path"`
	Origins    []string      `json:"origins"`
	PingPeriod time.Duration `json:"ping_period"`
	PongWait   time.Duration `json:"pong_wait"`
	WriteWait  time.Duration `json:"write_wait"`
	BufferSize int           `json:"buffer_size"`
}

// ExternalServicesConfig holds external services configuration
type ExternalServicesConfig struct {
	// Virus scanning service
	VirusScanner VirusScannerConfig `json:"virus_scanner"`

	// Geolocation service
	Geolocation GeolocationConfig `json:"geolocation"`

	// Analytics service
	Analytics AnalyticsConfig `json:"analytics"`
}

// VirusScannerConfig holds virus scanner configuration
type VirusScannerConfig struct {
	Enabled  bool          `json:"enabled"`
	Provider string        `json:"provider"` // clamav, virustotal
	Endpoint string        `json:"endpoint"`
	APIKey   string        `json:"-"` // Hidden from JSON
	Timeout  time.Duration `json:"timeout"`
}

// GeolocationConfig holds geolocation configuration
type GeolocationConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"` // maxmind, ipapi
	APIKey   string `json:"-"`        // Hidden from JSON
	Database string `json:"database"`
}

// AnalyticsConfig holds analytics configuration
type AnalyticsConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"` // google-analytics, mixpanel, amplitude
	APIKey   string `json:"-"`        // Hidden from JSON
	Endpoint string `json:"endpoint"`
}

// ServicesConfig holds microservices configuration
type ServicesConfig struct {
	User         ServiceConfig `json:"user"`
	File         ServiceConfig `json:"file"`
	Notification ServiceConfig `json:"notification"`
	Analytics    ServiceConfig `json:"analytics"`
	Search       ServiceConfig `json:"search"`
	Auth         ServiceConfig `json:"auth"`
}

// ServiceConfig holds individual service configuration
type ServiceConfig struct {
	BaseURL string        `json:"base_url"`
	Timeout time.Duration `json:"timeout"`
	Retries int           `json:"retries"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Debug:       getEnvBool("DEBUG", false),
	}

	// Load server configuration
	config.Server = ServerConfig{
		Host:            getEnv("SERVER_HOST", "0.0.0.0"),
		Port:            getEnvInt("SERVER_PORT", 8080),
		ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		TLSEnabled:      getEnvBool("SERVER_TLS_ENABLED", false),
		TLSCertFile:     getEnv("SERVER_TLS_CERT_FILE", ""),
		TLSKeyFile:      getEnv("SERVER_TLS_KEY_FILE", ""),
		CORSEnabled:     getEnvBool("SERVER_CORS_ENABLED", true),
		CORSOrigins:     getEnvSlice("SERVER_CORS_ORIGINS", []string{"*"}),
	}

	// Load database configuration
	config.Database = DatabaseConfig{
		Driver:             getEnv("DB_DRIVER", "postgres"),
		Host:               getEnv("DB_HOST", "localhost"),
		Port:               getEnvInt("DB_PORT", 5432),
		Database:           getEnv("DB_NAME", "fileserver"),
		Username:           getEnv("DB_USER", "postgres"),
		Password:           getEnv("DB_PASSWORD", ""),
		SSLMode:            getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConnections: getEnvInt("DB_MAX_OPEN_CONNECTIONS", 25),
		MaxIdleConnections: getEnvInt("DB_MAX_IDLE_CONNECTIONS", 5),
		ConnectionLifetime: getEnvDuration("DB_CONNECTION_LIFETIME", 5*time.Minute),
		ConnectionTimeout:  getEnvDuration("DB_CONNECTION_TIMEOUT", 30*time.Second),
		QueryTimeout:       getEnvDuration("DB_QUERY_TIMEOUT", 30*time.Second),
		MigrationsPath:     getEnv("DB_MIGRATIONS_PATH", "./migrations"),
		AutoMigrate:        getEnvBool("DB_AUTO_MIGRATE", true),
	}

	// Load cache configuration
	config.Cache = CacheConfig{
		Driver:       getEnv("CACHE_DRIVER", "redis"),
		Host:         getEnv("CACHE_HOST", "localhost"),
		Port:         getEnvInt("CACHE_PORT", 6379),
		Password:     getEnv("CACHE_PASSWORD", ""),
		Database:     getEnvInt("CACHE_DATABASE", 0),
		PoolSize:     getEnvInt("CACHE_POOL_SIZE", 10),
		MinIdleConns: getEnvInt("CACHE_MIN_IDLE_CONNS", 2),
		DialTimeout:  getEnvDuration("CACHE_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  getEnvDuration("CACHE_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: getEnvDuration("CACHE_WRITE_TIMEOUT", 3*time.Second),
		IdleTimeout:  getEnvDuration("CACHE_IDLE_TIMEOUT", 5*time.Minute),
		DefaultTTL:   getEnvDuration("CACHE_DEFAULT_TTL", 1*time.Hour),
	}

	// Load storage configuration
	config.Storage = StorageConfig{
		Driver:    getEnv("STORAGE_DRIVER", "local"),
		Bucket:    getEnv("STORAGE_BUCKET", ""),
		Region:    getEnv("STORAGE_REGION", ""),
		Endpoint:  getEnv("STORAGE_ENDPOINT", ""),
		AccessKey: getEnv("STORAGE_ACCESS_KEY", ""),
		SecretKey: getEnv("STORAGE_SECRET_KEY", ""),
		UseSSL:    getEnvBool("STORAGE_USE_SSL", true),
		LocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
	}

	// Load JWT configuration
	config.Security.JWT = JWTConfig{
		SecretKey:       getEnv("JWT_SECRET_KEY", "your-secret-key"),
		Issuer:          getEnv("JWT_ISSUER", "fileserver"),
		Audience:        getEnv("JWT_AUDIENCE", "fileserver-users"),
		AccessTokenTTL:  getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getEnvDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
		Algorithm:       getEnv("JWT_ALGORITHM", "HS256"),
		PublicKeyPath:   getEnv("JWT_PUBLIC_KEY_PATH", ""),
		PrivateKeyPath:  getEnv("JWT_PRIVATE_KEY_PATH", ""),
	}

	// Load file upload configuration
	config.FileUpload = FileUploadConfig{
		MaxFileSize:       getEnvInt64("FILE_UPLOAD_MAX_FILE_SIZE", 100*1024*1024),   // 100MB
		MaxTotalSize:      getEnvInt64("FILE_UPLOAD_MAX_TOTAL_SIZE", 1024*1024*1024), // 1GB
		AllowedMimeTypes:  getEnvSlice("FILE_UPLOAD_ALLOWED_MIME_TYPES", []string{"image/*", "application/pdf", "text/*"}),
		AllowedExtensions: getEnvSlice("FILE_UPLOAD_ALLOWED_EXTENSIONS", []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".doc", ".docx"}),
		ChunkSize:         getEnvInt64("FILE_UPLOAD_CHUNK_SIZE", 5*1024*1024), // 5MB
		UploadTimeout:     getEnvDuration("FILE_UPLOAD_TIMEOUT", 10*time.Minute),
		TempDir:           getEnv("FILE_UPLOAD_TEMP_DIR", "/tmp"),
		VirusScanEnabled:  getEnvBool("FILE_UPLOAD_VIRUS_SCAN_ENABLED", false),
	}

	// Load services configuration
	config.Services = ServicesConfig{
		User: ServiceConfig{
			BaseURL: getEnv("USER_SERVICE_URL", "http://localhost:8083"),
			Timeout: getEnvDuration("USER_SERVICE_TIMEOUT", 30*time.Second),
			Retries: getEnvInt("USER_SERVICE_RETRIES", 3),
		},
		File: ServiceConfig{
			BaseURL: getEnv("FILE_SERVICE_URL", "http://localhost:8082"),
			Timeout: getEnvDuration("FILE_SERVICE_TIMEOUT", 30*time.Second),
			Retries: getEnvInt("FILE_SERVICE_RETRIES", 3),
		},
		Notification: ServiceConfig{
			BaseURL: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8084"),
			Timeout: getEnvDuration("NOTIFICATION_SERVICE_TIMEOUT", 30*time.Second),
			Retries: getEnvInt("NOTIFICATION_SERVICE_RETRIES", 3),
		},
		Analytics: ServiceConfig{
			BaseURL: getEnv("ANALYTICS_SERVICE_URL", "http://localhost:8085"),
			Timeout: getEnvDuration("ANALYTICS_SERVICE_TIMEOUT", 30*time.Second),
			Retries: getEnvInt("ANALYTICS_SERVICE_RETRIES", 3),
		},
		Search: ServiceConfig{
			BaseURL: getEnv("SEARCH_SERVICE_URL", "http://localhost:8086"),
			Timeout: getEnvDuration("SEARCH_SERVICE_TIMEOUT", 30*time.Second),
			Retries: getEnvInt("SEARCH_SERVICE_RETRIES", 3),
		},
		Auth: ServiceConfig{
			BaseURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
			Timeout: getEnvDuration("AUTH_SERVICE_TIMEOUT", 30*time.Second),
			Retries: getEnvInt("AUTH_SERVICE_RETRIES", 3),
		},
	}

	return config, nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// GetDatabaseConnectionString returns the database connection string
func (c *DatabaseConfig) GetConnectionString() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// GetCacheConnectionString returns the cache connection string
func (c *CacheConfig) GetConnectionString() string {
	switch c.Driver {
	case "redis":
		if c.Password != "" {
			return fmt.Sprintf("redis://:%s@%s:%d/%d", c.Password, c.Host, c.Port, c.Database)
		}
		return fmt.Sprintf("redis://%s:%d/%d", c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Security.JWT.SecretKey == "" || c.Security.JWT.SecretKey == "your-secret-key" {
		return fmt.Errorf("JWT secret key must be set and not use default value")
	}

	if c.Database.Password == "" && c.Environment == "production" {
		return fmt.Errorf("database password must be set in production")
	}

	if c.FileUpload.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be greater than 0")
	}

	return nil
}
