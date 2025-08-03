package chaos

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	Closed CircuitBreakerState = iota
	Open
	HalfOpen
)

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	maxFailures  int
	timeout      time.Duration
	failureCount int64
	lastFailTime time.Time
	state        CircuitBreakerState
	mutex        sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       Closed,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == Open {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.state = HalfOpen
			cb.failureCount = 0
		} else {
			return errors.New("circuit breaker is open")
		}
	}

	err := fn()
	if err != nil {
		cb.failureCount++
		cb.lastFailTime = time.Now()
		if cb.failureCount >= int64(cb.maxFailures) {
			cb.state = Open
		}
		return err
	}

	if cb.state == HalfOpen {
		cb.state = Closed
	}
	cb.failureCount = 0
	return nil
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// RetryConfig defines retry mechanism configuration
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// RetryWithBackoff implements exponential backoff retry
func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error
	delay := config.BaseDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt == config.MaxAttempts {
			break
		}

		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		// Calculate next delay
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// TestCircuitBreakerResilience tests circuit breaker functionality
func TestCircuitBreakerResilience(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	failureCount := int64(0)

	// Simulate a failing service
	failingService := func() error {
		atomic.AddInt64(&failureCount, 1)
		return errors.New("service failure")
	}

	// Test circuit breaker opening
	t.Run("CircuitBreakerOpens", func(t *testing.T) {
		// First 3 calls should fail and open the circuit
		for i := 0; i < 3; i++ {
			err := cb.Call(failingService)
			assert.Error(t, err, "Service should fail")
			assert.NotEqual(t, "circuit breaker is open", err.Error(), "Circuit should not be open yet")
		}

		// Circuit should now be open
		assert.Equal(t, Open, cb.GetState(), "Circuit breaker should be open")

		// Next call should fail immediately due to open circuit
		err := cb.Call(failingService)
		assert.Error(t, err)
		assert.Equal(t, "circuit breaker is open", err.Error())
	})

	// Test circuit breaker recovery
	t.Run("CircuitBreakerRecovery", func(t *testing.T) {
		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// Create a working service
		workingService := func() error {
			return nil
		}

		// Circuit should transition to half-open and then closed
		err := cb.Call(workingService)
		assert.NoError(t, err, "Working service should succeed")
		assert.Equal(t, Closed, cb.GetState(), "Circuit breaker should be closed")
	})

	t.Logf("Circuit breaker test completed. Total failures: %d", atomic.LoadInt64(&failureCount))
}

// TestRetryMechanism tests retry with exponential backoff
func TestRetryMechanism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	config := RetryConfig{
		MaxAttempts: 5,
		BaseDelay:   10 * time.Millisecond,
		MaxDelay:    100 * time.Millisecond,
		Multiplier:  2.0,
	}

	t.Run("EventualSuccess", func(t *testing.T) {
		attempts := int64(0)
		ctx := context.Background()

		// Service that fails first 3 times, then succeeds
		flakeyService := func() error {
			current := atomic.AddInt64(&attempts, 1)
			if current <= 3 {
				return fmt.Errorf("attempt %d failed", current)
			}
			return nil
		}

		start := time.Now()
		err := RetryWithBackoff(ctx, config, flakeyService)
		duration := time.Since(start)

		assert.NoError(t, err, "Retry should eventually succeed")
		assert.Equal(t, int64(4), atomic.LoadInt64(&attempts), "Should take 4 attempts")
		assert.True(t, duration >= 70*time.Millisecond, "Should respect backoff delays")
		t.Logf("Retry succeeded after %d attempts in %v", atomic.LoadInt64(&attempts), duration)
	})

	t.Run("MaxAttemptsExceeded", func(t *testing.T) {
		attempts := int64(0)
		ctx := context.Background()

		// Service that always fails
		alwaysFailingService := func() error {
			atomic.AddInt64(&attempts, 1)
			return errors.New("persistent failure")
		}

		err := RetryWithBackoff(ctx, config, alwaysFailingService)
		assert.Error(t, err, "Retry should fail after max attempts")
		assert.Equal(t, int64(5), atomic.LoadInt64(&attempts), "Should attempt exactly 5 times")
		assert.Contains(t, err.Error(), "retry failed after 5 attempts")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		attempts := int64(0)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Service that always fails
		slowFailingService := func() error {
			atomic.AddInt64(&attempts, 1)
			time.Sleep(20 * time.Millisecond)
			return errors.New("slow failure")
		}

		err := RetryWithBackoff(ctx, config, slowFailingService)
		assert.Error(t, err, "Retry should fail due to context cancellation")
		assert.True(t, errors.Is(err, context.DeadlineExceeded), "Should be context deadline exceeded")
		t.Logf("Context cancelled after %d attempts", atomic.LoadInt64(&attempts))
	})
}

// TestNetworkChaos simulates network failures and partitions
func TestNetworkChaos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Simulate network conditions
	networkConditions := []struct {
		name        string
		latency     time.Duration
		packetLoss  float64
		timeout     time.Duration
		expectError bool
	}{
		{"Normal", 10 * time.Millisecond, 0.0, 100 * time.Millisecond, false},
		{"HighLatency", 200 * time.Millisecond, 0.0, 100 * time.Millisecond, true},
		{"PacketLoss", 50 * time.Millisecond, 0.5, 200 * time.Millisecond, true},
		{"NetworkPartition", 0, 1.0, 100 * time.Millisecond, true},
	}

	for _, condition := range networkConditions {
		t.Run(condition.name, func(t *testing.T) {
			successCount := 0
			totalRequests := 100

			for i := 0; i < totalRequests; i++ {
				err := simulateNetworkRequest(condition.latency, condition.packetLoss, condition.timeout)
				if err == nil {
					successCount++
				}
			}

			successRate := float64(successCount) / float64(totalRequests)
			t.Logf("%s: Success rate: %.2f%% (%d/%d)", condition.name, successRate*100, successCount, totalRequests)

			if condition.expectError {
				assert.True(t, successRate < 0.9, "Success rate should be low under adverse conditions")
			} else {
				assert.True(t, successRate > 0.9, "Success rate should be high under normal conditions")
			}
		})
	}
}

// TestInfrastructureChaos simulates infrastructure failures
func TestInfrastructureChaos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	t.Run("DatabaseFailure", func(t *testing.T) {
		// Simulate database connection failures
		db := &MockDatabase{failureRate: 0.3}
		cb := NewCircuitBreaker(5, 1*time.Second)

		successCount := 0
		totalOperations := 50

		for i := 0; i < totalOperations; i++ {
			err := cb.Call(func() error {
				return db.Query("SELECT * FROM users")
			})
			if err == nil {
				successCount++
			}
		}

		successRate := float64(successCount) / float64(totalOperations)
		t.Logf("Database operations: Success rate: %.2f%% (%d/%d)", successRate*100, successCount, totalOperations)

		// With circuit breaker, we should have some protection
		assert.True(t, successRate > 0.3, "Circuit breaker should provide some protection")
	})

	t.Run("MemoryPressure", func(t *testing.T) {
		// Simulate memory pressure
		memoryIntensiveOperations := 10
		successCount := 0

		for i := 0; i < memoryIntensiveOperations; i++ {
			err := simulateMemoryIntensiveOperation()
			if err == nil {
				successCount++
			}
		}

		successRate := float64(successCount) / float64(memoryIntensiveOperations)
		t.Logf("Memory intensive operations: Success rate: %.2f%% (%d/%d)", successRate*100, successCount, memoryIntensiveOperations)

		// Most operations should succeed unless system is under extreme pressure
		assert.True(t, successRate > 0.5, "Most memory operations should succeed")
	})

	t.Run("CPUStarvation", func(t *testing.T) {
		// Simulate CPU-intensive operations under load
		var wg sync.WaitGroup
		successCount := int64(0)
		totalOperations := int64(20)

		// Start multiple CPU-intensive goroutines
		for i := 0; i < int(totalOperations); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := simulateCPUIntensiveOperation()
				if err == nil {
					atomic.AddInt64(&successCount, 1)
				}
			}()
		}

		wg.Wait()

		successRate := float64(atomic.LoadInt64(&successCount)) / float64(totalOperations)
		t.Logf("CPU intensive operations: Success rate: %.2f%% (%d/%d)", successRate*100, atomic.LoadInt64(&successCount), totalOperations)

		// Operations should complete even under CPU pressure
		assert.True(t, successRate > 0.7, "Most CPU operations should succeed")
	})
}

// TestFaultTolerance tests system behavior under various fault conditions
func TestFaultTolerance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	t.Run("CascadingFailures", func(t *testing.T) {
		// Simulate cascading failures across services
		services := []*MockService{
			{name: "auth", failureRate: 0.05},
			{name: "user", failureRate: 0.08, dependencies: []string{"auth"}},
			{name: "file", failureRate: 0.07, dependencies: []string{"auth", "user"}},
		}

		totalRequests := 100
		successfulChains := 0

		for i := 0; i < totalRequests; i++ {
			if simulateServiceChain(services) {
				successfulChains++
			}
		}

		successRate := float64(successfulChains) / float64(totalRequests)
		t.Logf("Service chain success rate: %.2f%% (%d/%d)", successRate*100, successfulChains, totalRequests)

		// Even with failures, some requests should succeed
		assert.True(t, successRate > 0.4, "Service chain should have reasonable success rate")
	})

	t.Run("BulkheadIsolation", func(t *testing.T) {
		// Test that failures in one area don't affect others
		bulkheads := map[string]*MockBulkhead{
			"critical":   {capacity: 10, failureRate: 0.0},
			"normal":     {capacity: 20, failureRate: 0.3},
			"background": {capacity: 5, failureRate: 0.8},
		}

		results := make(map[string]int)
		var wg sync.WaitGroup
		var mu sync.Mutex

		for category, bulkhead := range bulkheads {
			wg.Add(1)
			go func(cat string, bh *MockBulkhead) {
				defer wg.Done()
				successCount := 0
				for i := 0; i < 50; i++ {
					if bh.Execute(func() error {
						if rand.Float64() < bh.failureRate {
							return errors.New("operation failed")
						}
						return nil
					}) == nil {
						successCount++
					}
				}
				mu.Lock()
				results[cat] = successCount
				mu.Unlock()
			}(category, bulkhead)
		}

		wg.Wait()

		t.Logf("Bulkhead results: %+v", results)

		// Critical operations should have high success rate
		assert.True(t, float64(results["critical"])/50 > 0.9, "Critical operations should succeed")
		// Background operations can fail more
		assert.True(t, float64(results["background"])/50 < 0.5, "Background operations should fail more")
	})
}

// Helper functions and mock implementations

func simulateNetworkRequest(latency time.Duration, packetLoss float64, timeout time.Duration) error {
	// Simulate packet loss
	if rand.Float64() < packetLoss {
		return errors.New("packet lost")
	}

	// Simulate network latency
	if latency > timeout {
		return errors.New("request timeout")
	}

	time.Sleep(latency)
	return nil
}

type MockDatabase struct {
	failureRate float64
}

func (db *MockDatabase) Query(query string) error {
	if rand.Float64() < db.failureRate {
		return errors.New("database connection failed")
	}
	return nil
}

func simulateMemoryIntensiveOperation() error {
	// Allocate and use memory
	data := make([][]byte, 1000)
	for i := range data {
		data[i] = make([]byte, 1024)
		// Simulate some work
		for j := range data[i] {
			data[i][j] = byte(i + j)
		}
	}
	// Force usage to prevent optimization
	_ = data[0][0]
	return nil
}

func simulateCPUIntensiveOperation() error {
	// Simulate CPU-intensive work
	result := 0
	for i := 0; i < 1000000; i++ {
		result += i * i
	}
	// Prevent optimization
	_ = result
	return nil
}

type MockService struct {
	name         string
	failureRate  float64
	dependencies []string
}

func simulateServiceChain(services []*MockService) bool {
	for _, service := range services {
		// Check if service fails
		if rand.Float64() < service.failureRate {
			return false
		}

		// Simulate dependency failures affecting this service
		for range service.dependencies {
			if rand.Float64() < 0.1 { // 10% chance dependency affects this service
				return false
			}
		}
	}
	return true
}

type MockBulkhead struct {
	capacity    int
	failureRate float64
	current     int64
	mutex       sync.Mutex
}

func (b *MockBulkhead) Execute(fn func() error) error {
	b.mutex.Lock()
	if int(atomic.LoadInt64(&b.current)) >= b.capacity {
		b.mutex.Unlock()
		return errors.New("bulkhead capacity exceeded")
	}
	atomic.AddInt64(&b.current, 1)
	b.mutex.Unlock()

	defer atomic.AddInt64(&b.current, -1)

	return fn()
}
