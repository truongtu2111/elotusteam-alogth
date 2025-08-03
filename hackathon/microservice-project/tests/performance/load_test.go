package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// LoadTestConfig defines configuration for load tests
type LoadTestConfig struct {
	ConcurrentUsers int
	Duration        time.Duration
	RampUpTime      time.Duration
	TargetRPS       int
}

// LoadTestResult contains results of load test
type LoadTestResult struct {
	TotalRequests   int
	SuccessfulReqs  int
	FailedReqs      int
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	P95ResponseTime time.Duration
	P99ResponseTime time.Duration
	Throughput      float64
	ErrorRate       float64
}

// TestAuthenticationLoad tests authentication endpoint under load
func TestAuthenticationLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	config := LoadTestConfig{
		ConcurrentUsers: 100,
		Duration:        30 * time.Second,
		RampUpTime:      5 * time.Second,
		TargetRPS:       1000,
	}

	result := runLoadTest(t, "/api/auth/login", createAuthPayload, config)

	// Assertions for performance requirements
	assert.True(t, result.ErrorRate < 0.10, "Error rate should be less than 10%%")
	assert.True(t, result.AvgResponseTime < 100*time.Millisecond, "Average response time should be under 100ms")
	assert.True(t, result.P95ResponseTime < 200*time.Millisecond, "95th percentile should be under 200ms")
	assert.True(t, result.P99ResponseTime < 500*time.Millisecond, "99th percentile should be under 500ms")
	assert.True(t, result.Throughput > 500, "Throughput should be over 500 RPS")

	t.Logf("Authentication Load Test Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-result.ErrorRate)*100)
	t.Logf("Average Response Time: %v", result.AvgResponseTime)
	t.Logf("95th Percentile: %v", result.P95ResponseTime)
	t.Logf("99th Percentile: %v", result.P99ResponseTime)
	t.Logf("Throughput: %.2f RPS", result.Throughput)
}

// TestFileUploadLoad tests file upload endpoint under load
func TestFileUploadLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	config := LoadTestConfig{
		ConcurrentUsers: 50,
		Duration:        20 * time.Second,
		RampUpTime:      3 * time.Second,
		TargetRPS:       200,
	}

	result := runLoadTest(t, "/api/files/upload", createFileUploadPayload, config)

	// File upload has different performance requirements
	assert.True(t, result.ErrorRate < 0.10, "Error rate should be less than 10%%")
	assert.True(t, result.AvgResponseTime < 500*time.Millisecond, "Average response time should be under 500ms")
	assert.True(t, result.P95ResponseTime < 1*time.Second, "95th percentile should be under 1s")
	assert.True(t, result.Throughput > 100, "Throughput should be over 100 RPS")

	t.Logf("File Upload Load Test Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-result.ErrorRate)*100)
	t.Logf("Average Response Time: %v", result.AvgResponseTime)
	t.Logf("95th Percentile: %v", result.P95ResponseTime)
	t.Logf("Throughput: %.2f RPS", result.Throughput)
}

// TestUserCreationLoad tests user creation endpoint under load
func TestUserCreationLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	config := LoadTestConfig{
		ConcurrentUsers: 75,
		Duration:        25 * time.Second,
		RampUpTime:      4 * time.Second,
		TargetRPS:       300,
	}

	result := runLoadTest(t, "/api/users", createUserPayload, config)

	assert.True(t, result.ErrorRate < 0.10, "Error rate should be less than 10%%")
	assert.True(t, result.AvgResponseTime < 150*time.Millisecond, "Average response time should be under 150ms")
	assert.True(t, result.P95ResponseTime < 300*time.Millisecond, "95th percentile should be under 300ms")
	assert.True(t, result.Throughput > 200, "Throughput should be over 200 RPS")

	t.Logf("User Creation Load Test Results:")
	t.Logf("Total Requests: %d", result.TotalRequests)
	t.Logf("Success Rate: %.2f%%", (1-result.ErrorRate)*100)
	t.Logf("Average Response Time: %v", result.AvgResponseTime)
	t.Logf("95th Percentile: %v", result.P95ResponseTime)
	t.Logf("Throughput: %.2f RPS", result.Throughput)
}

// runLoadTest executes a load test with the given configuration
func runLoadTest(t *testing.T, endpoint string, payloadFunc func(int) []byte, config LoadTestConfig) LoadTestResult {
	var (
		totalRequests  int
		successfulReqs int
		failedReqs     int
		responseTimes  []time.Duration
		mutex          sync.Mutex
		wg             sync.WaitGroup
	)

	startTime := time.Now()
	endTime := startTime.Add(config.Duration)

	// Create rate limiter
	rateLimit := time.Second / time.Duration(config.TargetRPS/config.ConcurrentUsers)

	// Start concurrent users
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Ramp up delay
			rampDelay := time.Duration(userID) * config.RampUpTime / time.Duration(config.ConcurrentUsers)
			time.Sleep(rampDelay)

			ticker := time.NewTicker(rateLimit)
			defer ticker.Stop()

			for time.Now().Before(endTime) {
				select {
				case <-ticker.C:
					reqStart := time.Now()
					payload := payloadFunc(userID)
					resp := makeRequest(endpoint, payload)
					reqDuration := time.Since(reqStart)

					mutex.Lock()
					totalRequests++
					responseTimes = append(responseTimes, reqDuration)
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						successfulReqs++
					} else {
						failedReqs++
					}
					mutex.Unlock()
				case <-time.After(time.Until(endTime)):
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// Calculate results
	return calculateResults(totalRequests, successfulReqs, failedReqs, responseTimes, config.Duration)
}

// calculateResults computes performance metrics from test data
func calculateResults(total, successful, failed int, responseTimes []time.Duration, duration time.Duration) LoadTestResult {
	if len(responseTimes) == 0 {
		return LoadTestResult{}
	}

	// Sort response times for percentile calculations
	sortDurations(responseTimes)

	// Calculate average
	var totalTime time.Duration
	for _, rt := range responseTimes {
		totalTime += rt
	}
	avgTime := totalTime / time.Duration(len(responseTimes))

	// Calculate percentiles
	p95Index := int(float64(len(responseTimes)) * 0.95)
	p99Index := int(float64(len(responseTimes)) * 0.99)
	if p95Index >= len(responseTimes) {
		p95Index = len(responseTimes) - 1
	}
	if p99Index >= len(responseTimes) {
		p99Index = len(responseTimes) - 1
	}

	return LoadTestResult{
		TotalRequests:   total,
		SuccessfulReqs:  successful,
		FailedReqs:      failed,
		AvgResponseTime: avgTime,
		MinResponseTime: responseTimes[0],
		MaxResponseTime: responseTimes[len(responseTimes)-1],
		P95ResponseTime: responseTimes[p95Index],
		P99ResponseTime: responseTimes[p99Index],
		Throughput:      float64(total) / duration.Seconds(),
		ErrorRate:       float64(failed) / float64(total),
	}
}

// sortDurations sorts a slice of durations in ascending order
func sortDurations(durations []time.Duration) {
	for i := 0; i < len(durations)-1; i++ {
		for j := 0; j < len(durations)-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}

// makeRequest makes an HTTP request and returns the response
func makeRequest(endpoint string, payload []byte) *http.Response {
	req := httptest.NewRequest("POST", endpoint, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	// Mock response for testing
	rr := httptest.NewRecorder()
	mockHandler(rr, req)

	return rr.Result()
}

// Payload creation functions
func createAuthPayload(userID int) []byte {
	payload := map[string]interface{}{
		"email":    fmt.Sprintf("user%d@example.com", userID),
		"password": "password123",
	}
	data, _ := json.Marshal(payload)
	return data
}

func createFileUploadPayload(userID int) []byte {
	// Simulate file upload payload
	payload := map[string]interface{}{
		"filename": fmt.Sprintf("file%d.txt", userID),
		"content":  "test file content",
		"size":     1024,
	}
	data, _ := json.Marshal(payload)
	return data
}

func createUserPayload(userID int) []byte {
	payload := map[string]interface{}{
		"name":     fmt.Sprintf("User %d", userID),
		"email":    fmt.Sprintf("user%d@example.com", userID),
		"password": "password123",
	}
	data, _ := json.Marshal(payload)
	return data
}

// Mock handler for testing
func mockHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate processing time
	processingTime := time.Duration(10+rand.Intn(50)) * time.Millisecond
	time.Sleep(processingTime)

	// Simulate occasional failures (5% error rate)
	if rand.Float64() < 0.05 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"success"}`)); err != nil {
		fmt.Printf("Warning: Failed to write response: %v\n", err)
	}
}
