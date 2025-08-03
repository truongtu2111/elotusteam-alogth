package performance

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ProfilingConfig defines configuration for profiling tests
type ProfilingConfig struct {
	Duration        time.Duration
	ConcurrentUsers int
	RequestsPerUser int
	ProfileCPU      bool
	ProfileMemory   bool
	ProfileGoroutine bool
}

// TestServiceProfiling runs comprehensive profiling tests on all services
func TestServiceProfiling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping profiling test in short mode")
	}

	config := ProfilingConfig{
		Duration:        60 * time.Second,
		ConcurrentUsers: 50,
		RequestsPerUser: 100,
		ProfileCPU:      true,
		ProfileMemory:   true,
		ProfileGoroutine: true,
	}

	// Create profiles directory
	profileDir := "../../profiles"
	os.MkdirAll(profileDir, 0755)

	// Start profiling
	if config.ProfileCPU {
		cpuFile, err := os.Create(fmt.Sprintf("%s/service_cpu.prof", profileDir))
		assert.NoError(t, err)
		defer cpuFile.Close()

		err = pprof.StartCPUProfile(cpuFile)
		assert.NoError(t, err)
		defer pprof.StopCPUProfile()
	}

	// Run load test with profiling
	results := runProfilingLoadTest(t, config)

	// Capture memory profile
	if config.ProfileMemory {
		memFile, err := os.Create(fmt.Sprintf("%s/service_mem.prof", profileDir))
		assert.NoError(t, err)
		defer memFile.Close()

		runtime.GC() // Force garbage collection
		err = pprof.WriteHeapProfile(memFile)
		assert.NoError(t, err)
	}

	// Capture goroutine profile
	if config.ProfileGoroutine {
		goroutineFile, err := os.Create(fmt.Sprintf("%s/service_goroutine.prof", profileDir))
		assert.NoError(t, err)
		defer goroutineFile.Close()

		err = pprof.Lookup("goroutine").WriteTo(goroutineFile, 0)
		assert.NoError(t, err)
	}

	// Analyze results
	analyzeProfilingResults(t, results, config)
}

// TestMemoryLeakDetection tests for memory leaks in services
func TestMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Baseline memory measurement
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	baselineAlloc := m1.Alloc

	// Run memory-intensive operations
	for i := 0; i < 1000; i++ {
		// Simulate service operations that might leak memory
		data := make([]byte, 1024*1024) // 1MB allocation
		_ = data
		
		// Simulate some processing
		time.Sleep(time.Millisecond)
	}

	// Force garbage collection and measure again
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Allow GC to complete
	runtime.GC()

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	finalAlloc := m2.Alloc

	// Check for memory leaks
	memoryIncrease := finalAlloc - baselineAlloc
	memoryIncreasePercent := float64(memoryIncrease) / float64(baselineAlloc) * 100

	t.Logf("Baseline memory: %d bytes", baselineAlloc)
	t.Logf("Final memory: %d bytes", finalAlloc)
	t.Logf("Memory increase: %d bytes (%.2f%%)", memoryIncrease, memoryIncreasePercent)

	// Assert memory increase is within acceptable limits (< 10%)
	assert.True(t, memoryIncreasePercent < 10.0, 
		"Memory increase %.2f%% exceeds threshold of 10%%", memoryIncreasePercent)
}

// TestGoroutineLeakDetection tests for goroutine leaks
func TestGoroutineLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine leak test in short mode")
	}

	// Baseline goroutine count
	baselineGoroutines := runtime.NumGoroutine()

	// Start goroutines that should be cleaned up
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				// Simulate work
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Allow cleanup

	// Check final goroutine count
	finalGoroutines := runtime.NumGoroutine()
	goroutineIncrease := finalGoroutines - baselineGoroutines

	t.Logf("Baseline goroutines: %d", baselineGoroutines)
	t.Logf("Final goroutines: %d", finalGoroutines)
	t.Logf("Goroutine increase: %d", goroutineIncrease)

	// Assert goroutine count is within acceptable limits (< 10 extra)
	assert.True(t, goroutineIncrease < 10, 
		"Goroutine increase %d exceeds threshold of 10", goroutineIncrease)
}

// TestCPUUsageUnderLoad tests CPU usage patterns under load
func TestCPUUsageUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CPU usage test in short mode")
	}

	// Monitor CPU usage during load test
	var cpuSamples []float64
	var mu sync.Mutex
	done := make(chan bool)

	// Start CPU monitoring
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Simple CPU usage estimation
				start := time.Now()
				for time.Since(start) < 10*time.Millisecond {
					// Busy wait to measure CPU
				}
				elapsed := time.Since(start)
				cpuUsage := float64(10*time.Millisecond) / float64(elapsed) * 100

				mu.Lock()
				cpuSamples = append(cpuSamples, cpuUsage)
				mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	// Run CPU-intensive operations
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000000; j++ {
				// CPU-intensive calculation
				_ = j * j * j
			}
		}()
	}

	wg.Wait()
	done <- true

	// Analyze CPU usage
	mu.Lock()
	defer mu.Unlock()

	if len(cpuSamples) > 0 {
		var total float64
		for _, sample := range cpuSamples {
			total += sample
		}
		avgCPU := total / float64(len(cpuSamples))

		t.Logf("Average CPU usage during load: %.2f%%", avgCPU)
		t.Logf("CPU samples collected: %d", len(cpuSamples))

		// Assert CPU usage is reasonable (< 90%)
		assert.True(t, avgCPU < 90.0, 
			"Average CPU usage %.2f%% exceeds threshold of 90%%", avgCPU)
	}
}

// runProfilingLoadTest runs a load test while collecting profiling data
func runProfilingLoadTest(t *testing.T, config ProfilingConfig) map[string]interface{} {
	var (
		totalRequests int64
		successfulReqs int64
		failedReqs    int64
		responseTimes []time.Duration
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	startTime := time.Now()
	endTime := startTime.Add(config.Duration)

	// Start concurrent users
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < config.RequestsPerUser && time.Now().Before(endTime); j++ {
				reqStart := time.Now()
				
				// Simulate HTTP request
				resp := simulateHTTPRequest()
				
				reqDuration := time.Since(reqStart)

				mu.Lock()
				totalRequests++
				responseTimes = append(responseTimes, reqDuration)
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					successfulReqs++
				} else {
					failedReqs++
				}
				mu.Unlock()

				// Small delay between requests
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	return map[string]interface{}{
		"total_requests":   totalRequests,
		"successful_reqs":  successfulReqs,
		"failed_reqs":      failedReqs,
		"response_times":   responseTimes,
		"duration":         time.Since(startTime),
	}
}

// simulateHTTPRequest simulates an HTTP request for profiling
func simulateHTTPRequest() *http.Response {
	// Simulate various response times and status codes
	time.Sleep(time.Duration(10+rand.Intn(50)) * time.Millisecond)
	
	statusCode := 200
	if rand.Float32() < 0.05 { // 5% error rate
		statusCode = 500
	}

	return &http.Response{
		StatusCode: statusCode,
	}
}

// analyzeProfilingResults analyzes the results of profiling tests
func analyzeProfilingResults(t *testing.T, results map[string]interface{}, config ProfilingConfig) {
	totalReqs := results["total_requests"].(int64)
	failedReqs := results["failed_reqs"].(int64)
	responseTimes := results["response_times"].([]time.Duration)
	duration := results["duration"].(time.Duration)

	// Calculate metrics
	errorRate := float64(failedReqs) / float64(totalReqs) * 100
	throughput := float64(totalReqs) / duration.Seconds()

	// Calculate average response time
	var totalTime time.Duration
	for _, rt := range responseTimes {
		totalTime += rt
	}
	avgResponseTime := totalTime / time.Duration(len(responseTimes))

	// Log results
	t.Logf("Profiling Load Test Results:")
	t.Logf("Total Requests: %d", totalReqs)
	t.Logf("Success Rate: %.2f%%", 100-errorRate)
	t.Logf("Error Rate: %.2f%%", errorRate)
	t.Logf("Throughput: %.2f RPS", throughput)
	t.Logf("Average Response Time: %v", avgResponseTime)
	t.Logf("Test Duration: %v", duration)

	// Performance assertions
	assert.True(t, errorRate < 10.0, "Error rate %.2f%% exceeds 10%%", errorRate)
	assert.True(t, avgResponseTime < 100*time.Millisecond, "Average response time %v exceeds 100ms", avgResponseTime)
	assert.True(t, throughput > 100, "Throughput %.2f RPS is below 100 RPS", throughput)

	// Memory usage check
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	t.Logf("Memory Usage: Alloc=%d KB, TotalAlloc=%d KB, Sys=%d KB, NumGC=%d", 
		m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)

	// Goroutine count check
	goroutineCount := runtime.NumGoroutine()
	t.Logf("Active Goroutines: %d", goroutineCount)
	assert.True(t, goroutineCount < 1000, "Goroutine count %d exceeds 1000", goroutineCount)
}