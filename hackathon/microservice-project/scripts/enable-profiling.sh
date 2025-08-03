#!/bin/bash

# Enable Profiling Script
# This script helps enable profiling endpoints in Go services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../" && pwd)"
SERVICES_DIR="$PROJECT_ROOT/services"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if pprof is already enabled in a service
check_pprof_enabled() {
    local service_dir="$1"
    local main_file="$service_dir/main.go"
    
    if [ -f "$main_file" ]; then
        if grep -q "net/http/pprof" "$main_file"; then
            return 0  # Already enabled
        fi
    fi
    return 1  # Not enabled
}

# Add pprof import and endpoint to a service
enable_pprof_in_service() {
    local service_name="$1"
    local service_dir="$SERVICES_DIR/$service_name"
    local main_file="$service_dir/main.go"
    
    if [ ! -f "$main_file" ]; then
        log_error "Main file not found for service: $service_name"
        return 1
    fi
    
    if check_pprof_enabled "$service_dir"; then
        log_info "Profiling already enabled in $service_name"
        return 0
    fi
    
    log_info "Enabling profiling in $service_name..."
    
    # Create backup
    cp "$main_file" "$main_file.backup"
    
    # Add pprof import
    if ! grep -q "_ \"net/http/pprof\"" "$main_file"; then
        # Find the import block and add pprof
        sed -i '' '/^import (/a\
	_ "net/http/pprof"
' "$main_file"
    fi
    
    # Add debug endpoint if not exists
    if ! grep -q "/debug/pprof" "$main_file"; then
        # Add debug routes before the main server start
        cat >> "$main_file.tmp" << 'EOF'

	// Enable profiling endpoints
	go func() {
		log.Println("Starting debug server on :6060")
		log.Println(http.ListenAndServe(":6060", nil))
	}()
EOF
        
        # Insert before the main server start
        sed -i '' '/log.Fatal(http.ListenAndServe/i\
	// Enable profiling endpoints\
	go func() {\
		log.Println("Starting debug server on :6060")\
		log.Println(http.ListenAndServe(":6060", nil))\
	}()\
' "$main_file"
    fi
    
    log_success "Profiling enabled in $service_name"
}

# Create profiling middleware for HTTP handlers
create_profiling_middleware() {
    local middleware_file="$PROJECT_ROOT/shared/middleware/profiling.go"
    
    mkdir -p "$(dirname "$middleware_file")"
    
    cat > "$middleware_file" << 'EOF'
package middleware

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// ProfilingMiddleware adds performance monitoring to HTTP handlers
func ProfilingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Capture initial memory stats
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		// Process request
		c.Next()
		
		// Calculate metrics
		duration := time.Since(start)
		
		// Capture final memory stats
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		// Add performance headers
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Memory-Alloc", fmt.Sprintf("%d", m2.Alloc-m1.Alloc))
		c.Header("X-Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))
		
		// Log slow requests (>100ms)
		if duration > 100*time.Millisecond {
			log.Printf("SLOW REQUEST: %s %s took %v", c.Request.Method, c.Request.URL.Path, duration)
		}
	}
}

// HealthCheckWithMetrics provides detailed health check with performance metrics
func HealthCheckWithMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		health := map[string]interface{}{
			"status":     "healthy",
			"timestamp":  time.Now().Unix(),
			"memory": map[string]interface{}{
				"alloc":       m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"num_gc":      m.NumGC,
			},
			"goroutines": runtime.NumGoroutine(),
			"version":    runtime.Version(),
		}
		
		c.JSON(http.StatusOK, health)
	}
}
EOF

    log_success "Profiling middleware created at $middleware_file"
}

# Create performance monitoring configuration
create_performance_config() {
    local config_file="$PROJECT_ROOT/configs/performance.yaml"
    
    mkdir -p "$(dirname "$config_file")"
    
    cat > "$config_file" << 'EOF'
# Performance Monitoring Configuration
performance:
  profiling:
    enabled: true
    cpu_profile_duration: 30s
    memory_profile_interval: 60s
    debug_port: 6060
    
  monitoring:
    slow_request_threshold: 100ms
    memory_threshold: 100MB
    goroutine_threshold: 1000
    
  alerts:
    cpu_usage_threshold: 80
    memory_usage_threshold: 85
    response_time_threshold: 500ms
    error_rate_threshold: 5
    
  metrics:
    collect_interval: 10s
    retention_period: 24h
    export_prometheus: true
    export_file: true
EOF

    log_success "Performance configuration created at $config_file"
}

# Create performance test runner script
create_performance_test_runner() {
    local runner_file="$PROJECT_ROOT/scripts/run-performance-tests.sh"
    
    cat > "$runner_file" << 'EOF'
#!/bin/bash

# Performance Test Runner
# Runs comprehensive performance tests with profiling

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../" && pwd)"
PROFILE_DIR="$PROJECT_ROOT/profiles"
REPORT_DIR="$PROJECT_ROOT/performance-reports"

# Create directories
mkdir -p "$PROFILE_DIR" "$REPORT_DIR"

echo "🚀 Starting Performance Test Suite..."

# Run benchmark tests with profiling
echo "📊 Running benchmark tests..."
go test -v -bench=. -benchmem -cpuprofile="$PROFILE_DIR/bench_cpu.prof" -memprofile="$PROFILE_DIR/bench_mem.prof" ./tests/performance/...

# Run load tests
echo "🔥 Running load tests..."
go test -v -run="TestLoad" ./tests/performance/...

# Run profiling tests
echo "🔍 Running profiling tests..."
go test -v -run="TestProfiling" ./tests/performance/...

# Generate reports
echo "📈 Generating performance reports..."
if [ -f "$PROFILE_DIR/bench_cpu.prof" ]; then
    go tool pprof -text "$PROFILE_DIR/bench_cpu.prof" > "$REPORT_DIR/cpu_analysis.txt"
    go tool pprof -svg "$PROFILE_DIR/bench_cpu.prof" > "$REPORT_DIR/cpu_analysis.svg"
fi

if [ -f "$PROFILE_DIR/bench_mem.prof" ]; then
    go tool pprof -text "$PROFILE_DIR/bench_mem.prof" > "$REPORT_DIR/memory_analysis.txt"
    go tool pprof -svg "$PROFILE_DIR/bench_mem.prof" > "$REPORT_DIR/memory_analysis.svg"
fi

echo "✅ Performance tests completed!"
echo "📁 Reports available in: $REPORT_DIR"
echo "🔬 Profiles available in: $PROFILE_DIR"
EOF

    chmod +x "$runner_file"
    log_success "Performance test runner created at $runner_file"
}

# Main function
main() {
    local command="${1:-help}"
    
    case "$command" in
        "enable")
            log_info "Enabling profiling in all services..."
            
            # Enable profiling in each service
            for service in auth file user analytics; do
                if [ -d "$SERVICES_DIR/$service" ]; then
                    enable_pprof_in_service "$service"
                else
                    log_warning "Service directory not found: $service"
                fi
            done
            
            # Create supporting files
            create_profiling_middleware
            create_performance_config
            create_performance_test_runner
            
            log_success "Profiling setup completed!"
            echo
            log_info "Next steps:"
            echo "  1. Rebuild your services: docker-compose build"
            echo "  2. Start services: docker-compose up -d"
            echo "  3. Run performance tests: ./scripts/run-performance-tests.sh"
            echo "  4. Access profiling data at: http://localhost:6060/debug/pprof/"
            ;;
        "disable")
            log_info "Disabling profiling in all services..."
            
            for service in auth file user analytics; do
                local main_file="$SERVICES_DIR/$service/main.go"
                local backup_file="$main_file.backup"
                
                if [ -f "$backup_file" ]; then
                    mv "$backup_file" "$main_file"
                    log_success "Profiling disabled in $service"
                else
                    log_warning "No backup found for $service"
                fi
            done
            ;;
        "status")
            log_info "Checking profiling status..."
            
            for service in auth file user analytics; do
                if check_pprof_enabled "$SERVICES_DIR/$service"; then
                    log_success "Profiling enabled in $service"
                else
                    log_warning "Profiling disabled in $service"
                fi
            done
            ;;
        "help")
            echo "Usage: $0 {enable|disable|status|help}"
            echo
            echo "Commands:"
            echo "  enable   - Enable profiling in all services"
            echo "  disable  - Disable profiling and restore backups"
            echo "  status   - Check profiling status in services"
            echo "  help     - Show this help message"
            echo
            echo "After enabling profiling:"
            echo "  - Rebuild services: docker-compose build"
            echo "  - Access profiling: http://localhost:6060/debug/pprof/"
            echo "  - Run tests: ./scripts/run-performance-tests.sh"
            ;;
        *)
            log_error "Unknown command: $command"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"