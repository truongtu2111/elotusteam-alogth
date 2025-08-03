package communication

import (
	"context"
	"time"
)

// Message represents a generic message for inter-service communication
type Message struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Payload       interface{}       `json:"payload"`
	Headers       map[string]string `json:"headers"`
	Timestamp     time.Time         `json:"timestamp"`
	CorrelationID string            `json:"correlation_id"`
	ReplyTo       string            `json:"reply_to,omitempty"`
	Expiration    *time.Time        `json:"expiration,omitempty"`
	RetryCount    int               `json:"retry_count"`
	MaxRetries    int               `json:"max_retries"`
}

// Response represents a response message
type Response struct {
	ID            string      `json:"id"`
	CorrelationID string      `json:"correlation_id"`
	Payload       interface{} `json:"payload"`
	Error         *Error      `json:"error,omitempty"`
	Timestamp     time.Time   `json:"timestamp"`
}

// Error represents an error in communication
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// MessageHandler defines the interface for handling incoming messages
type MessageHandler interface {
	Handle(ctx context.Context, msg *Message) (*Response, error)
}

// MessageHandlerFunc is an adapter to allow the use of ordinary functions as MessageHandlers
type MessageHandlerFunc func(ctx context.Context, msg *Message) (*Response, error)

// Handle calls f(ctx, msg)
func (f MessageHandlerFunc) Handle(ctx context.Context, msg *Message) (*Response, error) {
	return f(ctx, msg)
}

// SyncCommunicator defines the interface for synchronous communication
type SyncCommunicator interface {
	// Send sends a message and waits for a response
	Send(ctx context.Context, destination string, msg *Message) (*Response, error)

	// SendWithTimeout sends a message with a specific timeout
	SendWithTimeout(ctx context.Context, destination string, msg *Message, timeout time.Duration) (*Response, error)

	// RegisterHandler registers a handler for a specific message type
	RegisterHandler(messageType string, handler MessageHandler) error

	// Start starts the communicator
	Start(ctx context.Context) error

	// Stop stops the communicator
	Stop(ctx context.Context) error

	// Health checks the health of the communicator
	Health(ctx context.Context) error
}

// AsyncCommunicator defines the interface for asynchronous communication
type AsyncCommunicator interface {
	// Publish publishes a message to a topic/queue
	Publish(ctx context.Context, topic string, msg *Message) error

	// Subscribe subscribes to a topic/queue with a handler
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error

	// Unsubscribe unsubscribes from a topic/queue
	Unsubscribe(ctx context.Context, topic string) error

	// Start starts the communicator
	Start(ctx context.Context) error

	// Stop stops the communicator
	Stop(ctx context.Context) error

	// Health checks the health of the communicator
	Health(ctx context.Context) error
}

// EventBus defines the interface for event-driven communication
type EventBus interface {
	// PublishEvent publishes an event
	PublishEvent(ctx context.Context, event *Event) error

	// SubscribeToEvent subscribes to an event type
	SubscribeToEvent(ctx context.Context, eventType string, handler EventHandler) error

	// UnsubscribeFromEvent unsubscribes from an event type
	UnsubscribeFromEvent(ctx context.Context, eventType string) error

	// Start starts the event bus
	Start(ctx context.Context) error

	// Stop stops the event bus
	Stop(ctx context.Context) error
}

// Event represents a domain event
type Event struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Version       int                    `json:"version"`
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]string      `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id"`
	CausationID   string                 `json:"causation_id"`
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
}

// EventHandlerFunc is an adapter to allow the use of ordinary functions as EventHandlers
type EventHandlerFunc func(ctx context.Context, event *Event) error

// Handle calls f(ctx, event)
func (f EventHandlerFunc) Handle(ctx context.Context, event *Event) error {
	return f(ctx, event)
}

// CircuitBreaker defines the interface for circuit breaker pattern
type CircuitBreaker interface {
	// Execute executes a function with circuit breaker protection
	Execute(ctx context.Context, fn func() error) error

	// State returns the current state of the circuit breaker
	State() CircuitBreakerState

	// Reset resets the circuit breaker to closed state
	Reset()
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "closed"
	case CircuitBreakerOpen:
		return "open"
	case CircuitBreakerHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// RetryPolicy defines the retry policy for failed operations
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// DefaultRetryPolicy returns a default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"connection_error",
			"timeout_error",
			"temporary_error",
		},
	}
}

// CommunicationConfig holds configuration for communication
type CommunicationConfig struct {
	Type               string                 `json:"type"` // "grpc", "http", "rabbitmq", "kafka", etc.
	Endpoints          []string               `json:"endpoints"`
	Timeout            time.Duration          `json:"timeout"`
	RetryPolicy        *RetryPolicy           `json:"retry_policy"`
	CircuitBreaker     *CircuitBreakerConfig  `json:"circuit_breaker"`
	TLS                *TLSConfig             `json:"tls"`
	Authentication     *AuthConfig            `json:"authentication"`
	Compression        bool                   `json:"compression"`
	MaxMessageSize     int64                  `json:"max_message_size"`
	ConnectionPoolSize int                    `json:"connection_pool_size"`
	AdditionalSettings map[string]interface{} `json:"additional_settings"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`
	RecoveryTimeout    time.Duration `json:"recovery_timeout"`
	SuccessThreshold   int           `json:"success_threshold"`
	MonitoringInterval time.Duration `json:"monitoring_interval"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	CertFile           string `json:"cert_file"`
	KeyFile            string `json:"key_file"`
	CAFile             string `json:"ca_file"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type     string            `json:"type"` // "jwt", "oauth", "basic", "apikey"
	Settings map[string]string `json:"settings"`
}

// CommunicationFactory creates communicators based on configuration
type CommunicationFactory interface {
	// CreateSyncCommunicator creates a synchronous communicator
	CreateSyncCommunicator(config *CommunicationConfig) (SyncCommunicator, error)

	// CreateAsyncCommunicator creates an asynchronous communicator
	CreateAsyncCommunicator(config *CommunicationConfig) (AsyncCommunicator, error)

	// CreateEventBus creates an event bus
	CreateEventBus(config *CommunicationConfig) (EventBus, error)

	// CreateCircuitBreaker creates a circuit breaker
	CreateCircuitBreaker(config *CircuitBreakerConfig) (CircuitBreaker, error)
}

// Metrics defines the interface for communication metrics
type Metrics interface {
	// IncrementMessagesSent increments the messages sent counter
	IncrementMessagesSent(destination string, messageType string)

	// IncrementMessagesReceived increments the messages received counter
	IncrementMessagesReceived(source string, messageType string)

	// IncrementErrors increments the error counter
	IncrementErrors(operation string, errorType string)

	// RecordLatency records the latency of an operation
	RecordLatency(operation string, duration time.Duration)

	// RecordMessageSize records the size of a message
	RecordMessageSize(messageType string, size int64)
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
