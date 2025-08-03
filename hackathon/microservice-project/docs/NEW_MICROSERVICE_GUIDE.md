# Complete Guide: Creating a New Microservice

This comprehensive guide walks you through creating a new microservice in this project, including database setup, inter-service communication, migrations, and feature flag integration.

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Step 1: Create Service Structure](#step-1-create-service-structure)
3. [Step 2: Database Setup & Migration](#step-2-database-setup--migration)
4. [Step 3: Inter-Service Communication](#step-3-inter-service-communication)
5. [Step 4: Feature Flag Integration](#step-4-feature-flag-integration)
6. [Step 5: Configuration & Environment](#step-5-configuration--environment)
7. [Step 6: Testing Setup](#step-6-testing-setup)
8. [Step 7: Monitoring & Observability](#step-7-monitoring--observability)
9. [Step 8: Deployment Configuration](#step-8-deployment-configuration)
10. [Best Practices](#best-practices)

## Prerequisites

- Go 1.19+ installed
- Docker and Docker Compose
- PostgreSQL client tools
- Access to the project repository
- Understanding of Clean Architecture principles

## Step 1: Create Service Structure

### 1.1 Create Service Directory

```bash
# Navigate to services directory
cd services/

# Create new service directory (replace 'newservice' with your service name)
mkdir newservice
cd newservice

# Create Clean Architecture structure
mkdir -p {\
  config,\
  domain/{entities,repositories,usecases},\
  infrastructure/{database,http,external},\
  presentation/{handlers,middleware,routes},\
  internal/{errors,utils}\
}
```

### 1.2 Create Main Application File

Create `services/newservice/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	// Add service-specific metrics here
	newserviceActionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "newservice_actions_total",
			Help: "Total number of newservice actions",
		},
		[]string{"action"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(newserviceActionsTotal)
}

// Prometheus middleware
func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

func main() {
	// Load configuration from environment
	host := getEnv("SERVER_HOST", "localhost")
	port := getEnvAsInt("SERVER_PORT", 8086) // Use unique port

	// Set Gin mode
	gin.SetMode(gin.DebugMode)

	// Setup router
	router := gin.Default()
	
	// Add Prometheus middleware
	router.Use(prometheusMiddleware())

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "newservice",
			"time":    time.Now().UTC(),
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Add your service routes here
		newservice := api.Group("/newservice")
		{
			newservice.GET("/", func(c *gin.Context) {
				newserviceActionsTotal.WithLabelValues("list").Inc()
				c.JSON(http.StatusOK, gin.H{"message": "NewService endpoint - implementation pending"})
			})
			newservice.POST("/", func(c *gin.Context) {
				newserviceActionsTotal.WithLabelValues("create").Inc()
				c.JSON(http.StatusOK, gin.H{"message": "Create endpoint - implementation pending"})
			})
		}
	}

	// Start server
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           router,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("NewService starting on %s:%d", host, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
```

### 1.3 Create Dockerfile

Create `services/newservice/Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./services/newservice

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8086

# Run the binary
CMD ["./main"]
```

## Step 2: Database Setup & Migration

### 2.1 Create Database Migration

```bash
# Navigate to migrations directory
cd migrations/

# Create migration for your new service
./tools/create-migration.sh \
  --type SCHEMA_CHANGE \
  --description "Create newservice database schema" \
  --risk LOW \
  --duration "2m" \
  --author "$(whoami)" \
  --ticket "PROJ-XXX"
```

### 2.2 Define Database Schema

Edit the generated migration file in `migrations/scripts/`:

```sql
-- Migration: Create newservice database schema
-- Type: SCHEMA_CHANGE
-- Risk: LOW
-- Duration: 2m
-- Author: your-name
-- Ticket: PROJ-XXX

-- Create database for newservice
CREATE DATABASE newservice_db;

-- Connect to the new database
\c newservice_db;

-- Create main table for your service
CREATE TABLE IF NOT EXISTS newservice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

-- Create indexes
CREATE INDEX idx_newservice_items_status ON newservice_items(status);
CREATE INDEX idx_newservice_items_created_at ON newservice_items(created_at);
CREATE INDEX idx_newservice_items_name ON newservice_items(name);

-- Create audit table
CREATE TABLE IF NOT EXISTS newservice_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(255) NOT NULL,
    operation VARCHAR(10) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changed_by UUID,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create trigger function for audit
CREATE OR REPLACE FUNCTION newservice_audit_trigger()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO newservice_audit (table_name, operation, new_values, changed_by)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(NEW), NEW.created_by);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO newservice_audit (table_name, operation, old_values, new_values, changed_by)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD), row_to_json(NEW), NEW.updated_by);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO newservice_audit (table_name, operation, old_values, changed_by)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD), OLD.updated_by);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers
CREATE TRIGGER newservice_items_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON newservice_items
    FOR EACH ROW EXECUTE FUNCTION newservice_audit_trigger();

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE newservice_db TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;
```

### 2.3 Create Rollback Migration

Create corresponding rollback file:

```sql
-- Rollback: Create newservice database schema
-- This rollback script removes the newservice database and all related objects

-- Drop triggers
DROP TRIGGER IF EXISTS newservice_items_audit_trigger ON newservice_items;

-- Drop functions
DROP FUNCTION IF EXISTS newservice_audit_trigger();

-- Drop tables
DROP TABLE IF EXISTS newservice_audit;
DROP TABLE IF EXISTS newservice_items;

-- Drop database
DROP DATABASE IF EXISTS newservice_db;
```

### 2.4 Apply Migration

```bash
# Apply migration to development environment
./tools/migrate.sh migrate development

# Check migration status
./tools/migrate.sh status development
```

## Step 3: Inter-Service Communication

### 3.1 Synchronous Communication (HTTP/gRPC)

Create `services/newservice/infrastructure/http/client.go`:

```go
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/elotusteam/microservice-project/shared/communication"
)

// HTTPClient implements synchronous communication
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
	}
}

// Send implements SyncCommunicator interface
func (c *HTTPClient) Send(ctx context.Context, destination string, msg *communication.Message) (*communication.Response, error) {
	return c.SendWithTimeout(ctx, destination, msg, 30*time.Second)
}

// SendWithTimeout sends a message with timeout
func (c *HTTPClient) SendWithTimeout(ctx context.Context, destination string, msg *communication.Message, timeout time.Duration) (*communication.Response, error) {
	// Create request context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Marshal message to JSON
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s", c.baseURL, destination)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", msg.CorrelationID)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var response communication.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// Example: Call User Service
func (c *HTTPClient) GetUser(ctx context.Context, userID string) (*User, error) {
	msg := &communication.Message{
		ID:            generateID(),
		Type:          "GET_USER",
		Payload:       map[string]string{"user_id": userID},
		Timestamp:     time.Now(),
		CorrelationID: generateCorrelationID(),
	}

	resp, err := c.Send(ctx, "/api/v1/users/"+userID, msg)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("user service error: %s", resp.Error.Message)
	}

	// Parse user data
	var user User
	userData, _ := json.Marshal(resp.Payload)
	if err := json.Unmarshal(userData, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &user, nil
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func generateID() string {
	// Implement ID generation
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateCorrelationID() string {
	// Implement correlation ID generation
	return fmt.Sprintf("corr-%d", time.Now().UnixNano())
}
```

### 3.2 Asynchronous Communication (Event-Driven)

Create `services/newservice/infrastructure/events/publisher.go`:

```go
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/elotusteam/microservice-project/shared/communication"
)

// EventPublisher handles event publishing
type EventPublisher struct {
	// Add your message broker client (RabbitMQ, Kafka, etc.)
	// For now, we'll use a simple in-memory implementation
	subscribers map[string][]communication.EventHandler
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		subscribers: make(map[string][]communication.EventHandler),
	}
}

// PublishEvent publishes an event
func (p *EventPublisher) PublishEvent(ctx context.Context, event *communication.Event) error {
	// In a real implementation, this would publish to your message broker
	// For now, we'll simulate by calling local handlers
	
	handlers, exists := p.subscribers[event.Type]
	if !exists {
		return nil // No subscribers
	}

	for _, handler := range handlers {
		go func(h communication.EventHandler) {
			if err := h.Handle(ctx, event); err != nil {
				// Log error or send to dead letter queue
				fmt.Printf("Error handling event %s: %v\n", event.Type, err)
			}
		}(handler)
	}

	return nil
}

// SubscribeToEvent subscribes to an event type
func (p *EventPublisher) SubscribeToEvent(ctx context.Context, eventType string, handler communication.EventHandler) error {
	p.subscribers[eventType] = append(p.subscribers[eventType], handler)
	return nil
}

// Example: Publish NewService Created Event
func (p *EventPublisher) PublishNewServiceCreated(ctx context.Context, itemID, userID string, item interface{}) error {
	event := &communication.Event{
		ID:            generateEventID(),
		Type:          "newservice.item.created",
		AggregateID:   itemID,
		AggregateType: "newservice_item",
		Version:       1,
		Payload: map[string]interface{}{
			"item_id": itemID,
			"user_id": userID,
			"item":    item,
		},
		Metadata: map[string]string{
			"service": "newservice",
			"version": "v1",
		},
		Timestamp:     time.Now(),
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

func generateEventID() string {
	return fmt.Sprintf("evt-%d", time.Now().UnixNano())
}
```

### 3.3 Service Registration

Add your service to `shared/config/config.go`:

```go
// Update ServicesConfig struct
type ServicesConfig struct {
	User         ServiceConfig `json:"user"`
	File         ServiceConfig `json:"file"`
	Notification ServiceConfig `json:"notification"`
	Analytics    ServiceConfig `json:"analytics"`
	Search       ServiceConfig `json:"search"`
	Auth         ServiceConfig `json:"auth"`
	NewService   ServiceConfig `json:"newservice"` // Add your service
}
```

## Step 4: Feature Flag Integration

### 4.1 Add Feature Flag Middleware

Create `services/newservice/presentation/middleware/featureflags.go`:

```go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/elotusteam/microservice-project/shared/featureflags"
)

// FeatureFlagMiddleware adds feature flag support
func FeatureFlagMiddleware(manager *featureflags.Manager) gin.HandlerFunc {
	middleware := featureflags.NewGinMiddleware(manager, featureflags.DefaultMiddlewareConfig())
	return middleware.Handler()
}

// RequireFeatureFlag middleware that requires a specific feature flag
func RequireFeatureFlag(manager *featureflags.Manager, flagID string) gin.HandlerFunc {
	middleware := featureflags.NewGinMiddleware(manager, featureflags.DefaultMiddlewareConfig())
	return middleware.RequireFlag(flagID)
}
```

### 4.2 Integrate Feature Flags in Service

Update your `main.go` to include feature flags:

```go
// Add to imports
import (
	"github.com/elotusteam/microservice-project/shared/featureflags"
	"github.com/elotusteam/microservice-project/services/newservice/presentation/middleware"
)

// Add to main function after router setup
func main() {
	// ... existing code ...

	// Initialize feature flag manager
	ffManager, err := featureflags.NewManager(featureflags.Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/newservice_db?sslmode=disable"),
		CacheEnabled: true,
		CacheTTL: 5 * time.Minute,
	})
	if err != nil {
		log.Fatalf("Failed to initialize feature flag manager: %v", err)
	}

	// Add feature flag middleware
	router.Use(middleware.FeatureFlagMiddleware(ffManager))

	// Example: Protected route with feature flag
	api := router.Group("/api/v1")
	{
		newservice := api.Group("/newservice")
		{
			// Regular endpoint
			newservice.GET("/", func(c *gin.Context) {
				newserviceActionsTotal.WithLabelValues("list").Inc()
				c.JSON(http.StatusOK, gin.H{"message": "NewService endpoint"})
			})

			// Feature flag protected endpoint
			newservice.POST("/advanced", 
				middleware.RequireFeatureFlag(ffManager, "newservice_advanced_features"),
				func(c *gin.Context) {
					newserviceActionsTotal.WithLabelValues("advanced_create").Inc()
					c.JSON(http.StatusOK, gin.H{"message": "Advanced feature enabled"})
				},
			)

			// Conditional feature implementation
			newservice.GET("/features", func(c *gin.Context) {
				userContext := featureflags.UserContext{
					UserID: c.GetHeader("X-User-ID"),
					Email:  c.GetHeader("X-User-Email"),
				}

				// Check feature flag programmatically
				if ffManager.IsEnabled("newservice_beta_ui", userContext) {
					c.JSON(http.StatusOK, gin.H{
						"ui_version": "beta",
						"features": []string{"advanced_search", "real_time_updates"},
					})
				} else {
					c.JSON(http.StatusOK, gin.H{
						"ui_version": "stable",
						"features": []string{"basic_search"},
					})
				}
			})
		}
	}

	// ... rest of the code ...
}
```

### 4.3 Create Feature Flags for Your Service

Create initial feature flags:

```bash
# Using the feature flag factory
go run -c '
import (
	"github.com/elotusteam/microservice-project/shared/featureflags"
)

func main() {
	manager, _ := featureflags.NewManager(featureflags.Config{
		DatabaseURL: "postgres://postgres:password@localhost:5432/newservice_db?sslmode=disable",
	})

	// Create feature flags for your service
	flags := []featureflags.FeatureFlag{
		featureflags.CreateBooleanFlag(
			"newservice_advanced_features",
			"Advanced NewService Features",
			"Enable advanced features for newservice",
			false, // Initially disabled
			"development",
			"newservice",
		),
		featureflags.CreateRolloutFlag(
			"newservice_beta_ui",
			"Beta UI for NewService",
			"Gradual rollout of new UI",
			0.1, // 10% rollout
			"development",
			"newservice",
		),
	}

	for _, flag := range flags {
		manager.CreateFlag(flag)
	}
}'
```

## Step 5: Configuration & Environment

### 5.1 Update Docker Compose

Add your service to `docker-compose.yml`:

```yaml
# Add to services section
newservice-service:
  build:
    context: .
    dockerfile: services/newservice/Dockerfile
  ports:
    - "8086:8086"
  environment:
    - SERVER_HOST=0.0.0.0
    - SERVER_PORT=8086
    - DB_HOST=postgres
    - DB_PORT=5432
    - DB_NAME=newservice_db
    - DB_USER=postgres
    - DB_PASSWORD=password
    - DATABASE_URL=postgres://postgres:password@postgres:5432/newservice_db?sslmode=disable
  depends_on:
    - postgres
  networks:
    - microservice-network
```

### 5.2 Update Nginx Configuration

Add routing to `nginx.conf`:

```nginx
# Add to the upstream and location blocks
upstream newservice {
    server newservice-service:8086;
}

# Add location block
location /api/v1/newservice {
    proxy_pass http://newservice;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### 5.3 Update Makefile

Add build and run targets to `Makefile`:

```makefile
# Add to build target
build:
	@echo "Building all services..."
	go build -o bin/auth ./services/auth
	go build -o bin/file ./services/file
	go build -o bin/user ./services/user
	go build -o bin/analytics ./services/analytics
	go build -o bin/newservice ./services/newservice  # Add this line
	@echo "Build complete!"

# Add run target
run-newservice:
	@echo "Starting newservice on port 8086..."
	SERVER_PORT=8086 go run ./services/newservice
```

## Step 6: Testing Setup

### 6.1 Create Unit Tests

Create `services/newservice/domain/usecases/newservice_test.go`:

```go
package usecases

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockNewServiceRepository struct {
	mock.Mock
}

func (m *MockNewServiceRepository) Create(ctx context.Context, item *NewServiceItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockNewServiceRepository) GetByID(ctx context.Context, id string) (*NewServiceItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*NewServiceItem), args.Error(1)
}

// Test cases
func TestNewServiceUseCase_CreateItem(t *testing.T) {
	// Setup
	mockRepo := new(MockNewServiceRepository)
	useCase := NewNewServiceUseCase(mockRepo)

	item := &NewServiceItem{
		Name:        "Test Item",
		Description: "Test Description",
	}

	mockRepo.On("Create", mock.Anything, item).Return(nil)

	// Execute
	err := useCase.CreateItem(context.Background(), item)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
```

### 6.2 Create Integration Tests

Create `tests/integration/newservice_test.go`:

```go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceAPI(t *testing.T) {
	// Setup test router
	router := gin.New()
	
	// Add your routes here
	api := router.Group("/api/v1")
	{
		newservice := api.Group("/newservice")
		{
			newservice.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			newservice.POST("/", func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"message": "created"})
			})
		}
	}

	t.Run("GET /api/v1/newservice", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/v1/newservice", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})

	t.Run("POST /api/v1/newservice", func(t *testing.T) {
		payload := map[string]string{"name": "test"}
		jsonPayload, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/newservice", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
```

## Step 7: Monitoring & Observability

### 7.1 Update Prometheus Configuration

Add your service to `monitoring/prometheus/prometheus.yml`:

```yaml
scrape_configs:
  # ... existing configs ...
  
  - job_name: 'newservice'
    static_configs:
      - targets: ['newservice-service:8086']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

### 7.2 Create Grafana Dashboard

Create `monitoring/grafana/dashboards/application/newservice-dashboard.json`:

```json
{
  "dashboard": {
    "title": "NewService Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"newservice\"}[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"newservice\"}[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "NewService Actions",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(newservice_actions_total[5m])",
            "legendFormat": "{{action}}"
          }
        ]
      }
    ]
  }
}
```

## Step 8: Deployment Configuration

### 8.1 Build and Deploy

```bash
# Build your service
make build

# Or build specific service
go build -o bin/newservice ./services/newservice

# Run locally
make run-newservice

# Or with Docker Compose
docker-compose up newservice-service

# Run all services
docker-compose up
```

### 8.2 Health Check Verification

```bash
# Check service health
curl http://localhost:8086/health

# Check metrics
curl http://localhost:8086/metrics

# Test API endpoint
curl http://localhost:8086/api/v1/newservice
```

## Best Practices

### 1. Code Organization
- Follow Clean Architecture principles
- Separate concerns (domain, infrastructure, presentation)
- Use dependency injection
- Implement proper error handling

### 2. Database Management
- Always create rollback migrations
- Use transactions for complex operations
- Implement proper indexing
- Add audit trails for important tables

### 3. Inter-Service Communication
- Use correlation IDs for tracing
- Implement circuit breakers for resilience
- Handle timeouts gracefully
- Use async communication for non-critical operations

### 4. Feature Flags
- Start with flags disabled in production
- Use gradual rollouts for new features
- Monitor flag performance impact
- Clean up unused flags regularly

### 5. Monitoring
- Add custom metrics for business logic
- Set up alerts for critical errors
- Monitor resource usage
- Track feature flag usage

### 6. Security
- Validate all inputs
- Use environment variables for secrets
- Implement proper authentication/authorization
- Log security events

### 7. Testing
- Write unit tests for business logic
- Create integration tests for APIs
- Test feature flag scenarios
- Perform load testing

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check database connectivity
   psql -h localhost -p 5432 -U postgres -d newservice_db
   ```

2. **Service Discovery Issues**
   ```bash
   # Check if service is registered
   docker-compose ps
   
   # Check network connectivity
   docker network ls
   ```

3. **Feature Flag Issues**
   ```bash
   # Check feature flag database
   psql -h localhost -p 5432 -U postgres -d newservice_db -c "SELECT * FROM feature_flags;"
   ```

4. **Monitoring Issues**
   ```bash
   # Check Prometheus targets
   curl http://localhost:9090/api/v1/targets
   
   # Check metrics endpoint
   curl http://localhost:8086/metrics
   ```

## Next Steps

After creating your microservice, follow these detailed implementation steps:

### 1. Implement Business Logic in Domain Layer

#### 1.1 Create Domain Entities

Create `services/newservice/domain/entities/newservice_item.go`:

```go
package entities

import (
	"time"
	"github.com/google/uuid"
)

type NewServiceItem struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description" db:"description"`
	Status      ItemStatus             `json:"status" db:"status"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
}

type ItemStatus string

const (
	ItemStatusActive   ItemStatus = "active"
	ItemStatusInactive ItemStatus = "inactive"
	ItemStatusDeleted  ItemStatus = "deleted"
	ItemStatusPending  ItemStatus = "pending"
)

// Business logic methods
func (item *NewServiceItem) IsActive() bool {
	return item.Status == ItemStatusActive
}

func (item *NewServiceItem) CanBeModified() bool {
	return item.Status == ItemStatusActive || item.Status == ItemStatusPending
}

func (item *NewServiceItem) Activate() error {
	if item.Status == ItemStatusDeleted {
		return errors.New("cannot activate deleted item")
	}
	item.Status = ItemStatusActive
	item.UpdatedAt = time.Now()
	return nil
}

func (item *NewServiceItem) SoftDelete() {
	item.Status = ItemStatusDeleted
	item.UpdatedAt = time.Now()
}
```

#### 1.2 Create Repository Interface

Create `services/newservice/domain/repositories/newservice_repository.go`:

```go
package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/elotusteam/microservice-project/services/newservice/domain/entities"
)

type NewServiceRepository interface {
	Create(ctx context.Context, item *entities.NewServiceItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.NewServiceItem, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.NewServiceItem, error)
	Update(ctx context.Context, item *entities.NewServiceItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, query string, filters map[string]interface{}) ([]*entities.NewServiceItem, error)
	GetStats(ctx context.Context, userID uuid.UUID) (*ItemStats, error)
}

type ItemStats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
	Pending  int64 `json:"pending"`
}
```

#### 1.3 Create Use Cases

Create `services/newservice/domain/usecases/newservice_usecase.go`:

```go
package usecases

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/elotusteam/microservice-project/services/newservice/domain/entities"
	"github.com/elotusteam/microservice-project/services/newservice/domain/repositories"
	"github.com/elotusteam/microservice-project/shared/featureflags"
)

type NewServiceUseCase struct {
	repo        repositories.NewServiceRepository
	ffManager   *featureflags.Manager
	eventBus    EventPublisher
}

func NewNewServiceUseCase(repo repositories.NewServiceRepository, ffManager *featureflags.Manager, eventBus EventPublisher) *NewServiceUseCase {
	return &NewServiceUseCase{
		repo:      repo,
		ffManager: ffManager,
		eventBus:  eventBus,
	}
}

func (uc *NewServiceUseCase) CreateItem(ctx context.Context, req *CreateItemRequest) (*entities.NewServiceItem, error) {
	// Validate input
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check feature flags
	userContext := featureflags.UserContext{UserID: req.UserID.String()}
	if req.IsAdvanced && !uc.ffManager.IsEnabled("newservice_advanced_features", userContext) {
		return nil, fmt.Errorf("advanced features not enabled for user")
	}

	// Create entity
	item := &entities.NewServiceItem{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Status:      entities.ItemStatusPending,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   req.UserID,
		UpdatedBy:   req.UserID,
	}

	// Save to repository
	if err := uc.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Publish event
	if err := uc.eventBus.PublishItemCreated(ctx, item); err != nil {
		// Log error but don't fail the operation
		log.Printf("Failed to publish item created event: %v", err)
	}

	return item, nil
}

type CreateItemRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=255"`
	Description string                 `json:"description" validate:"max=1000"`
	Metadata    map[string]interface{} `json:"metadata"`
	UserID      uuid.UUID              `json:"user_id" validate:"required"`
	IsAdvanced  bool                   `json:"is_advanced"`
}

func (r *CreateItemRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) > 255 {
		return errors.New("name too long")
	}
	if r.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	return nil
}
```

### 2. Add Comprehensive Error Handling

#### 2.1 Create Custom Error Types

Create `services/newservice/internal/errors/errors.go`:

```go
package errors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeFeatureDisabled ErrorCode = "FEATURE_DISABLED"
)

type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

func NewFeatureDisabledError(feature string) *AppError {
	return &AppError{
		Code:       ErrCodeFeatureDisabled,
		Message:    fmt.Sprintf("Feature '%s' is not enabled", feature),
		HTTPStatus: http.StatusForbidden,
	}
}
```

#### 2.2 Create Error Handling Middleware

Create `services/newservice/presentation/middleware/error_handler.go`:

```go
package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/elotusteam/microservice-project/services/newservice/internal/errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			
			if appErr, ok := err.(*errors.AppError); ok {
				c.JSON(appErr.HTTPStatus, gin.H{
					"error": appErr,
					"timestamp": time.Now().UTC(),
					"path": c.Request.URL.Path,
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code": "INTERNAL_ERROR",
						"message": "Internal server error",
					},
					"timestamp": time.Now().UTC(),
					"path": c.Request.URL.Path,
				})
			}
		}
	}
}
```

### 3. Set up CI/CD Pipeline

#### 3.1 Create GitHub Actions Workflow

Create `.github/workflows/newservice-ci.yml`:

```yaml
name: NewService CI/CD

on:
  push:
    paths:
      - 'services/newservice/**'
      - '.github/workflows/newservice-ci.yml'
  pull_request:
    paths:
      - 'services/newservice/**'

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: newservice_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.19
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: |
        cd services/newservice
        go test -v -race -coverprofile=coverage.out ./...
      env:
        DATABASE_URL: postgres://postgres:password@localhost:5432/newservice_test?sslmode=disable
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./services/newservice/coverage.out
        flags: newservice
    
    - name: Build binary
      run: |
        cd services/newservice
        go build -o newservice .
    
    - name: Run integration tests
      run: |
        cd tests/integration
        go test -v ./newservice_test.go
      env:
        DATABASE_URL: postgres://postgres:password@localhost:5432/newservice_test?sslmode=disable

  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v3
      with:
        context: .
        file: ./services/newservice/Dockerfile
        push: true
        tags: |
          ghcr.io/${{ github.repository }}/newservice:latest
          ghcr.io/${{ github.repository }}/newservice:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
```

### 4. Configure Production Environment

#### 4.1 Create Production Configuration

Create `services/newservice/config/production.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8086
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: "${DB_HOST}"
  port: 5432
  name: "${DB_NAME}"
  user: "${DB_USER}"
  password: "${DB_PASSWORD}"
  ssl_mode: "require"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s

redis:
  host: "${REDIS_HOST}"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0
  pool_size: 10

feature_flags:
  cache_enabled: true
  cache_ttl: 300s
  refresh_interval: 60s

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  path: "/metrics"

tracing:
  enabled: true
  jaeger_endpoint: "${JAEGER_ENDPOINT}"
  service_name: "newservice"
  sample_rate: 0.1

security:
  cors:
    allowed_origins: ["https://yourdomain.com"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["*"]
  rate_limiting:
    enabled: true
    requests_per_minute: 100
```

### 5. Add Performance Optimization

#### 5.1 Implement Caching Layer

Create `services/newservice/infrastructure/cache/redis_cache.go`:

```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	return json.Unmarshal([]byte(data), dest)
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	return nil
}

var ErrCacheMiss = fmt.Errorf("cache miss")
```

#### 5.2 Add Database Connection Pooling

Create `services/newservice/infrastructure/database/connection.go`:

```go
package database

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/lib/pq"
)

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func NewConnection(config Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
```

### 6. Implement Caching Strategies

#### 6.1 Repository with Cache

Create `services/newservice/infrastructure/repositories/cached_newservice_repository.go`:

```go
package repositories

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/elotusteam/microservice-project/services/newservice/domain/entities"
	"github.com/elotusteam/microservice-project/services/newservice/domain/repositories"
	"github.com/elotusteam/microservice-project/services/newservice/infrastructure/cache"
)

type CachedNewServiceRepository struct {
	baseRepo repositories.NewServiceRepository
	cache    *cache.RedisCache
}

func NewCachedNewServiceRepository(baseRepo repositories.NewServiceRepository, cache *cache.RedisCache) *CachedNewServiceRepository {
	return &CachedNewServiceRepository{
		baseRepo: baseRepo,
		cache:    cache,
	}
}

func (r *CachedNewServiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.NewServiceItem, error) {
	cacheKey := fmt.Sprintf("newservice:item:%s", id.String())
	
	// Try cache first
	var item entities.NewServiceItem
	if err := r.cache.Get(ctx, cacheKey, &item); err == nil {
		return &item, nil
	}

	// Cache miss, get from database
	item, err := r.baseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, item); err != nil {
		// Log error but don't fail the operation
		log.Printf("Failed to cache item %s: %v", id, err)
	}

	return item, nil
}

func (r *CachedNewServiceRepository) Update(ctx context.Context, item *entities.NewServiceItem) error {
	if err := r.baseRepo.Update(ctx, item); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("newservice:item:%s", item.ID.String())
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("Failed to invalidate cache for item %s: %v", item.ID, err)
	}

	// Invalidate user-specific caches
	userCachePattern := fmt.Sprintf("newservice:user:%s:*", item.CreatedBy.String())
	if err := r.cache.InvalidatePattern(ctx, userCachePattern); err != nil {
		log.Printf("Failed to invalidate user cache pattern %s: %v", userCachePattern, err)
	}

	return nil
}
```

### 7. Set up Log Aggregation

#### 7.1 Structured Logging

Create `services/newservice/internal/logger/logger.go`:

```go
package logger

import (
	"context"
	"os"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(level string, format string) *Logger {
	logger := logrus.New()
	
	// Set log level
	if lvl, err := logrus.ParseLevel(level); err == nil {
		logger.SetLevel(lvl)
	}

	// Set format
	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logger.SetOutput(os.Stdout)

	return &Logger{Logger: logger}
}

func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.Logger.WithContext(ctx)
	
	// Add correlation ID if present
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		entry = entry.WithField("correlation_id", correlationID)
	}
	
	// Add user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		entry = entry.WithField("user_id", userID)
	}
	
	// Add service name
	entry = entry.WithField("service", "newservice")
	
	return entry
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(fields)
}
```

#### 7.2 Log Aggregation with ELK Stack

Create `docker-compose.logging.yml`:

```yaml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.15.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    networks:
      - microservice-network

  logstash:
    image: docker.elastic.co/logstash/logstash:7.15.0
    volumes:
      - ./monitoring/logstash/pipeline:/usr/share/logstash/pipeline
      - ./monitoring/logstash/config:/usr/share/logstash/config
    ports:
      - "5044:5044"
      - "9600:9600"
    environment:
      LS_JAVA_OPTS: "-Xmx256m -Xms256m"
    depends_on:
      - elasticsearch
    networks:
      - microservice-network

  kibana:
    image: docker.elastic.co/kibana/kibana:7.15.0
    ports:
      - "5601:5601"
    environment:
      ELASTICSEARCH_HOSTS: http://elasticsearch:9200
    depends_on:
      - elasticsearch
    networks:
      - microservice-network

  filebeat:
    image: docker.elastic.co/beats/filebeat:7.15.0
    user: root
    volumes:
      - ./monitoring/filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    depends_on:
      - logstash
    networks:
      - microservice-network

volumes:
  elasticsearch_data:

networks:
  microservice-network:
    external: true
```

### 8. Configure Backup Strategies

#### 8.1 Database Backup Script

Create `scripts/backup-newservice-db.sh`:

```bash
#!/bin/bash

set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-newservice_db}
DB_USER=${DB_USER:-postgres}
BACKUP_DIR=${BACKUP_DIR:-/backups/newservice}
RETENTION_DAYS=${RETENTION_DAYS:-7}

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/newservice_backup_$TIMESTAMP.sql"

echo "Starting backup of $DB_NAME database..."

# Create database backup
pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  --verbose --clean --no-owner --no-privileges \
  --file="$BACKUP_FILE"

# Compress backup
gzip "$BACKUP_FILE"
BACKUP_FILE="$BACKUP_FILE.gz"

echo "Backup completed: $BACKUP_FILE"

# Calculate backup size
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
echo "Backup size: $BACKUP_SIZE"

# Clean up old backups
echo "Cleaning up backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -name "newservice_backup_*.sql.gz" -mtime +"$RETENTION_DAYS" -delete

# Upload to cloud storage (optional)
if [ -n "$AWS_S3_BUCKET" ]; then
  echo "Uploading backup to S3..."
  aws s3 cp "$BACKUP_FILE" "s3://$AWS_S3_BUCKET/newservice/$(basename $BACKUP_FILE)"
  echo "Backup uploaded to S3"
fi

echo "Backup process completed successfully"
```

#### 8.2 Automated Backup with Cron

Create `scripts/setup-backup-cron.sh`:

```bash
#!/bin/bash

# Add backup job to crontab
# Run daily at 2 AM
echo "0 2 * * * /path/to/scripts/backup-newservice-db.sh >> /var/log/newservice-backup.log 2>&1" | crontab -

echo "Backup cron job installed"
echo "Backups will run daily at 2:00 AM"
echo "Logs will be written to /var/log/newservice-backup.log"
```

### 9. Additional Production Considerations

#### 9.1 Health Checks and Readiness Probes

Enhance your health check endpoint:

```go
func (h *HealthHandler) DetailedHealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "newservice",
		"version":   os.Getenv("SERVICE_VERSION"),
		"timestamp": time.Now().UTC(),
		"checks":    make(map[string]interface{}),
	}

	// Database health check
	if err := h.db.Ping(); err != nil {
		health["status"] = "unhealthy"
		health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
			"status": "down",
			"error":  err.Error(),
		}
	} else {
		health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
			"status": "up",
		}
	}

	// Redis health check
	if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
		health["status"] = "unhealthy"
		health["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
			"status": "down",
			"error":  err.Error(),
		}
	} else {
		health["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
			"status": "up",
		}
	}

	if health["status"] == "healthy" {
		c.JSON(http.StatusOK, health)
	} else {
		c.JSON(http.StatusServiceUnavailable, health)
	}
}
```

#### 9.2 Graceful Shutdown Enhancement

```go
func gracefulShutdown(server *http.Server, db *sql.DB, redis *redis.Client) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connections
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	// Close Redis connections
	if err := redis.Close(); err != nil {
		log.Printf("Error closing Redis: %v", err)
	}

	log.Println("Server shutdown complete")
}
```

For more detailed information, refer to:
- [Feature Flags Integration Guide](./FEATURE_FLAGS_INTEGRATION.md)
- [Database Migration Guide](../migrations/README.md)
- [Monitoring Setup Guide](../MONITORING_SETUP_GUIDE.md)
- [Testing Guide](./FEATURE_FLAGS_TESTING.md)
- [Performance Optimization Guide](./PERFORMANCE_OPTIMIZATION.md)
- [Security Best Practices](./SECURITY_BEST_PRACTICES.md)