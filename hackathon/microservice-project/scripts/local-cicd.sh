#!/bin/bash

# Local CI/CD Pipeline Script
# This script runs the same CI/CD pipeline locally that would run in GitHub Actions
# It can be triggered manually or automatically via Git hooks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$PROJECT_ROOT/.local-cicd-logs"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="$LOG_DIR/pipeline_$TIMESTAMP.log"

# Create log directory
mkdir -p "$LOG_DIR"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_stage() {
    echo -e "${PURPLE}[STAGE]${NC} $1" | tee -a "$LOG_FILE"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Run command with logging
run_with_log() {
    local cmd="$1"
    local description="$2"
    
    log_info "Running: $description"
    echo "Command: $cmd" >> "$LOG_FILE"
    
    if eval "$cmd" >> "$LOG_FILE" 2>&1; then
        log_success "$description completed"
        return 0
    else
        log_error "$description failed"
        return 1
    fi
}

# Stage 1: Code Quality & Security
stage_code_quality() {
    log_stage "Stage 1: Code Quality & Security"
    
    # Go formatting check
    log_info "Checking Go formatting..."
    if ! gofmt -l . | grep -q .; then
        log_success "Go formatting is correct"
    else
        log_warning "Go formatting issues found. Running gofmt..."
        run_with_log "gofmt -w ." "Go formatting fix"
    fi
    
    # Go vet
    run_with_log "go vet ./..." "Go vet analysis"
    
    # Install and run golangci-lint if available
    if command_exists golangci-lint; then
        run_with_log "golangci-lint run" "Linting with golangci-lint"
    else
        log_warning "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
    
    # Security scanning with gosec if available
    if command_exists gosec; then
        run_with_log "gosec ./..." "Security scanning with gosec"
    else
        log_warning "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
    fi
    
    # Dependency vulnerability check with nancy if available
    if command_exists nancy; then
        run_with_log "go list -json -deps ./... | nancy sleuth" "Vulnerability scanning with nancy"
    else
        log_warning "nancy not found. Install with: go install github.com/sonatypecommunity/nancy@latest"
    fi
}

# Stage 2: Unit Tests
stage_unit_tests() {
    log_stage "Stage 2: Unit Tests"
    
    # Run unit tests with coverage
    run_with_log "go test -v -race -coverprofile=coverage.out ./tests/unit/..." "Unit tests with race detection"
    
    # Generate coverage report
    if [ -f "coverage.out" ]; then
        run_with_log "go tool cover -html=coverage.out -o coverage.html" "Coverage report generation"
        
        # Show coverage percentage
        local coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        log_info "Total test coverage: $coverage"
        
        # Check coverage threshold (80%)
        local coverage_num=$(echo $coverage | sed 's/%//')
        if (( $(echo "$coverage_num >= 80" | bc -l) )); then
            log_success "Coverage meets threshold (≥80%)"
        else
            log_warning "Coverage below threshold: $coverage (target: ≥80%)"
        fi
    fi
}

# Stage 3: Integration Tests
stage_integration_tests() {
    log_stage "Stage 3: Integration Tests"
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker to run integration tests."
        return 1
    fi
    
    # Start test services
    log_info "Starting test services..."
    run_with_log "docker-compose -f docker-compose.yml up -d postgres redis" "Starting PostgreSQL and Redis"
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 10
    
    # Run integration tests
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_USER=postgres
    export DB_PASSWORD=password
    export DB_NAME=test_db
    export REDIS_HOST=localhost
    export REDIS_PORT=6379
    
    run_with_log "go test -v -tags=integration ./tests/integration/..." "Integration tests"
    
    # Cleanup
    run_with_log "docker-compose down" "Stopping test services"
}

# Stage 4: Security Tests
stage_security_tests() {
    log_stage "Stage 4: Security Tests"
    
    # Run security tests
    run_with_log "go test -v ./tests/security/..." "Security tests"
}

# Stage 5: Build Docker Images
stage_build_images() {
    log_stage "Stage 5: Build Docker Images"
    
    # Build all service images
    local services=("auth" "file" "user" "analytics")
    
    for service in "${services[@]}"; do
        local image_name="microservice-${service}:local-$(git rev-parse --short HEAD)"
        run_with_log "docker build -t $image_name -f services/$service/Dockerfile ." "Building $service image"
    done
    
    # Build with docker-compose for consistency
    run_with_log "docker-compose build" "Building all services with docker-compose"
}

# Stage 6: End-to-End Tests
stage_e2e_tests() {
    log_stage "Stage 6: End-to-End Tests"
    
    # Start all services
    log_info "Starting all services for E2E tests..."
    run_with_log "docker-compose up -d" "Starting all services"
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 30
    
    # Check service health
    local services=("auth-service:8001" "file-service:8002" "user-service:8003" "analytics-service:8004")
    for service_port in "${services[@]}"; do
        local service=$(echo $service_port | cut -d: -f1)
        local port=$(echo $service_port | cut -d: -f2)
        
        if curl -f "http://localhost:$port/health" >/dev/null 2>&1; then
            log_success "$service is healthy"
        else
            log_warning "$service health check failed"
        fi
    done
    
    # Run E2E tests
    run_with_log "go test -v ./tests/e2e/..." "End-to-end tests"
    
    # Cleanup
    run_with_log "docker-compose down" "Stopping all services"
}

# Stage 7: Performance Tests with Profiling (optional)
stage_performance_tests() {
    log_stage "Stage 7: Performance Tests with Profiling"
    
    # Create profiling directory
    local profile_dir="$PROJECT_ROOT/profiles"
    mkdir -p "$profile_dir"
    
    # Run performance tests with CPU profiling
    log_info "Running performance tests with CPU profiling..."
    run_with_log "go test -v -short -cpuprofile=$profile_dir/cpu.prof -memprofile=$profile_dir/mem.prof -benchmem -bench=. ./tests/performance/..." "Performance tests with profiling"
    
    # Generate profiling reports if profiles exist
    if [ -f "$profile_dir/cpu.prof" ]; then
        log_info "Generating CPU profile report..."
        run_with_log "go tool pprof -text $profile_dir/cpu.prof > $profile_dir/cpu_profile.txt" "CPU profile text report"
        run_with_log "go tool pprof -svg $profile_dir/cpu.prof > $profile_dir/cpu_profile.svg" "CPU profile SVG report"
    fi
    
    if [ -f "$profile_dir/mem.prof" ]; then
        log_info "Generating memory profile report..."
        run_with_log "go tool pprof -text $profile_dir/mem.prof > $profile_dir/mem_profile.txt" "Memory profile text report"
        run_with_log "go tool pprof -svg $profile_dir/mem.prof > $profile_dir/mem_profile.svg" "Memory profile SVG report"
    fi
    
    # Run service-specific profiling if services are running
    if docker ps | grep -q "microservice"; then
        log_info "Running service profiling..."
        stage_service_profiling
    fi
    
    # Analyze performance metrics
    analyze_performance_metrics
}

# Service-specific profiling for running services
stage_service_profiling() {
    local profile_dir="$PROJECT_ROOT/profiles"
    local services=("auth-service:8001" "file-service:8002" "user-service:8003" "analytics-service:8004")
    
    for service_port in "${services[@]}"; do
        local service=$(echo $service_port | cut -d: -f1)
        local port=$(echo $service_port | cut -d: -f2)
        
        # Check if service is running and has debug endpoint
        if curl -f "http://localhost:$port/health" >/dev/null 2>&1; then
            log_info "Profiling $service..."
            
            # CPU profile (30 seconds)
            if curl -f "http://localhost:$port/debug/pprof/profile?seconds=30" -o "$profile_dir/${service}_cpu.prof" >/dev/null 2>&1; then
                log_success "CPU profile collected for $service"
                run_with_log "go tool pprof -text $profile_dir/${service}_cpu.prof > $profile_dir/${service}_cpu_profile.txt" "$service CPU profile analysis"
            else
                log_warning "Could not collect CPU profile for $service (debug endpoint may not be enabled)"
            fi
            
            # Memory profile
            if curl -f "http://localhost:$port/debug/pprof/heap" -o "$profile_dir/${service}_heap.prof" >/dev/null 2>&1; then
                log_success "Memory profile collected for $service"
                run_with_log "go tool pprof -text $profile_dir/${service}_heap.prof > $profile_dir/${service}_heap_profile.txt" "$service memory profile analysis"
            else
                log_warning "Could not collect memory profile for $service"
            fi
            
            # Goroutine profile
            if curl -f "http://localhost:$port/debug/pprof/goroutine" -o "$profile_dir/${service}_goroutine.prof" >/dev/null 2>&1; then
                log_success "Goroutine profile collected for $service"
                run_with_log "go tool pprof -text $profile_dir/${service}_goroutine.prof > $profile_dir/${service}_goroutine_profile.txt" "$service goroutine profile analysis"
            fi
        else
            log_warning "$service is not running or not accessible"
        fi
    done
}

# Analyze performance metrics and generate recommendations
analyze_performance_metrics() {
    local profile_dir="$PROJECT_ROOT/profiles"
    local analysis_file="$profile_dir/performance_analysis.md"
    
    log_info "Analyzing performance metrics..."
    
    cat > "$analysis_file" << 'EOF'
# Performance Analysis Report

## Overview
This report contains performance analysis results from the CI/CD pipeline.

## Test Results
### Benchmark Tests
- CPU profiling data available in `cpu.prof`
- Memory profiling data available in `mem.prof`
- Detailed text reports generated for analysis

### Service Profiling
EOF

    # Add service-specific analysis
    local services=("auth-service" "file-service" "user-service" "analytics-service")
    for service in "${services[@]}"; do
        if [ -f "$profile_dir/${service}_cpu_profile.txt" ]; then
            echo "#### $service" >> "$analysis_file"
            echo "- CPU profile: \`${service}_cpu_profile.txt\`" >> "$analysis_file"
            echo "- Memory profile: \`${service}_heap_profile.txt\`" >> "$analysis_file"
            echo "- Goroutine profile: \`${service}_goroutine_profile.txt\`" >> "$analysis_file"
            echo "" >> "$analysis_file"
        fi
    done
    
    cat >> "$analysis_file" << 'EOF'
## Performance Recommendations

### CPU Optimization
1. Review CPU profile for hot spots in `cpu_profile.txt`
2. Look for functions consuming >10% CPU time
3. Consider optimizing algorithms in high-usage functions

### Memory Optimization
1. Check memory profile for allocation patterns in `mem_profile.txt`
2. Look for memory leaks or excessive allocations
3. Consider object pooling for frequently allocated objects

### Concurrency Analysis
1. Review goroutine profiles for potential deadlocks
2. Check for goroutine leaks
3. Optimize channel usage and synchronization

## Action Items
- [ ] Review profiles for performance bottlenecks
- [ ] Implement optimizations for identified issues
- [ ] Set up continuous performance monitoring
- [ ] Establish performance regression alerts

EOF

    log_success "Performance analysis report generated: $analysis_file"
    
    # Check for common performance issues
    check_performance_issues
}

# Check for common performance issues
check_performance_issues() {
    local profile_dir="$PROJECT_ROOT/profiles"
    local issues_found=false
    
    log_info "Checking for common performance issues..."
    
    # Check CPU profile for hot spots
    if [ -f "$profile_dir/cpu_profile.txt" ]; then
        if grep -q "[1-9][0-9]\+%" "$profile_dir/cpu_profile.txt"; then
            log_warning "High CPU usage detected in some functions (>10%)"
            issues_found=true
        fi
    fi
    
    # Check memory profile for large allocations
    if [ -f "$profile_dir/mem_profile.txt" ]; then
        if grep -q "[1-9][0-9]\+MB" "$profile_dir/mem_profile.txt"; then
            log_warning "Large memory allocations detected (>10MB)"
            issues_found=true
        fi
    fi
    
    # Check for goroutine leaks in service profiles
    for service in auth-service file-service user-service analytics-service; do
        if [ -f "$profile_dir/${service}_goroutine_profile.txt" ]; then
            local goroutine_count=$(head -1 "$profile_dir/${service}_goroutine_profile.txt" | grep -o '[0-9]\+' | head -1)
            if [ "$goroutine_count" -gt 1000 ]; then
                log_warning "High goroutine count detected in $service: $goroutine_count"
                issues_found=true
            fi
        fi
    done
    
    if [ "$issues_found" = false ]; then
        log_success "No major performance issues detected"
    else
        log_warning "Performance issues detected. Review the analysis report for details."
    fi
}

# Stage 8: Chaos Tests (optional)
stage_chaos_tests() {
    log_stage "Stage 8: Chaos Engineering Tests"
    
    # Run chaos tests in short mode
    run_with_log "go test -v -short ./tests/chaos/..." "Chaos engineering tests (short mode)"
}

# Generate pipeline report
generate_report() {
    log_stage "Generating Pipeline Report"
    
    local report_file="$LOG_DIR/report_$TIMESTAMP.html"
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Local CI/CD Pipeline Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .stage { margin: 20px 0; padding: 15px; border-left: 4px solid #007cba; }
        .success { border-left-color: #28a745; }
        .warning { border-left-color: #ffc107; }
        .error { border-left-color: #dc3545; }
        .timestamp { color: #666; font-size: 0.9em; }
        pre { background-color: #f8f9fa; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Local CI/CD Pipeline Report</h1>
        <p class="timestamp">Generated: $(date)</p>
        <p>Git Commit: $(git rev-parse HEAD)</p>
        <p>Branch: $(git branch --show-current)</p>
    </div>
    
    <div class="stage">
        <h2>Pipeline Summary</h2>
        <p>Log file: <code>$LOG_FILE</code></p>
        <p>Coverage report: <code>coverage.html</code> (if generated)</p>
    </div>
    
    <div class="stage">
        <h2>Recent Log Output</h2>
        <pre>$(tail -50 "$LOG_FILE")</pre>
    </div>
</body>
</html>
EOF
    
    log_success "Pipeline report generated: $report_file"
    
    # Open report in browser if available
    if command_exists open; then
        open "$report_file"
    elif command_exists xdg-open; then
        xdg-open "$report_file"
    fi
}

# Main pipeline function
run_pipeline() {
    local stages="$1"
    
    log_info "Starting Local CI/CD Pipeline"
    log_info "Project: $(basename "$PROJECT_ROOT")"
    log_info "Git Commit: $(git rev-parse --short HEAD)"
    log_info "Branch: $(git branch --show-current)"
    log_info "Timestamp: $(date)"
    log_info "Log file: $LOG_FILE"
    
    cd "$PROJECT_ROOT"
    
    # Default stages if none specified
    if [ -z "$stages" ]; then
        stages="quality,unit,integration,security,build,e2e"
    fi
    
    local failed_stages=()
    
    # Run stages based on input
    IFS=',' read -ra STAGE_ARRAY <<< "$stages"
    for stage in "${STAGE_ARRAY[@]}"; do
        case "$stage" in
            quality)
                if ! stage_code_quality; then
                    failed_stages+=("Code Quality")
                fi
                ;;
            unit)
                if ! stage_unit_tests; then
                    failed_stages+=("Unit Tests")
                fi
                ;;
            integration)
                if ! stage_integration_tests; then
                    failed_stages+=("Integration Tests")
                fi
                ;;
            security)
                if ! stage_security_tests; then
                    failed_stages+=("Security Tests")
                fi
                ;;
            build)
                if ! stage_build_images; then
                    failed_stages+=("Build Images")
                fi
                ;;
            e2e)
                if ! stage_e2e_tests; then
                    failed_stages+=("E2E Tests")
                fi
                ;;
            performance)
                if ! stage_performance_tests; then
                    failed_stages+=("Performance Tests")
                fi
                ;;
            chaos)
                if ! stage_chaos_tests; then
                    failed_stages+=("Chaos Tests")
                fi
                ;;
            *)
                log_warning "Unknown stage: $stage"
                ;;
        esac
    done
    
    # Generate report
    generate_report
    
    # Summary
    echo
    log_stage "Pipeline Summary"
    
    if [ ${#failed_stages[@]} -eq 0 ]; then
        log_success "All stages completed successfully! ✅"
        echo
        log_info "Next steps:"
        echo "  • Review coverage report: coverage.html"
        echo "  • Check pipeline report: $LOG_DIR/report_$TIMESTAMP.html"
        echo "  • Push changes to trigger remote CI/CD"
        return 0
    else
        log_error "Pipeline failed in stages: ${failed_stages[*]} ❌"
        echo
        log_info "Check the log file for details: $LOG_FILE"
        return 1
    fi
}

# Setup Git hooks for automatic pipeline execution
setup_git_hooks() {
    log_info "Setting up Git hooks for automatic CI/CD..."
    
    local hooks_dir="$PROJECT_ROOT/.git/hooks"
    
    # Pre-push hook
    cat > "$hooks_dir/pre-push" << 'EOF'
#!/bin/bash

# Pre-push hook to run local CI/CD pipeline
echo "🚀 Running local CI/CD pipeline before push..."

# Get the directory of this script
HOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$HOOKS_DIR/../.." && pwd)"

# Run quick pipeline (quality + unit tests)
if "$PROJECT_ROOT/scripts/local-cicd.sh" quick; then
    echo "✅ Local CI/CD pipeline passed. Proceeding with push."
    exit 0
else
    echo "❌ Local CI/CD pipeline failed. Push aborted."
    echo "Run 'scripts/local-cicd.sh' to see detailed results."
    exit 1
fi
EOF
    
    chmod +x "$hooks_dir/pre-push"
    log_success "Pre-push hook installed"
    
    # Post-commit hook for full pipeline (optional)
    cat > "$hooks_dir/post-commit" << 'EOF'
#!/bin/bash

# Post-commit hook to run full CI/CD pipeline (optional)
# Uncomment the line below to enable automatic full pipeline after each commit

# echo "🔄 Running full CI/CD pipeline after commit..."
# "$(git rev-parse --show-toplevel)/scripts/local-cicd.sh" full &
EOF
    
    chmod +x "$hooks_dir/post-commit"
    log_success "Post-commit hook installed (disabled by default)"
}

# Show help
show_help() {
    echo "Local CI/CD Pipeline Script"
    echo
    echo "Usage: $0 [command] [options]"
    echo
    echo "Commands:"
    echo "  full                    Run full pipeline (all stages)"
    echo "  quick                   Run quick pipeline (quality + unit tests)"
    echo "  custom <stages>         Run custom stages (comma-separated)"
    echo "  setup-hooks            Setup Git hooks for automatic execution"
    echo "  install-tools          Install required CI/CD tools"
    echo "  clean                  Clean up logs and artifacts"
    echo "  help                   Show this help message"
    echo
    echo "Available stages:"
    echo "  quality                Code quality and security checks"
    echo "  unit                   Unit tests with coverage"
    echo "  integration            Integration tests"
    echo "  security               Security tests"
    echo "  build                  Build Docker images"
    echo "  e2e                    End-to-end tests"
    echo "  performance            Performance tests"
    echo "  chaos                  Chaos engineering tests"
    echo
    echo "Examples:"
    echo "  $0 full                # Run all stages"
    echo "  $0 quick               # Run quality + unit tests"
    echo "  $0 custom quality,unit,build  # Run specific stages"
    echo "  $0 setup-hooks         # Setup Git hooks"
    echo
}

# Install required tools
install_tools() {
    log_info "Installing CI/CD tools..."
    
    # Go tools
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    go install github.com/sonatypecommunity/nancy@latest
    
    log_success "CI/CD tools installed"
}

# Clean up logs and artifacts
clean_artifacts() {
    log_info "Cleaning up CI/CD artifacts..."
    
    rm -rf "$LOG_DIR"
    rm -f coverage.out coverage.html
    docker system prune -f
    
    log_success "Artifacts cleaned"
}

# Main script
main() {
    case "${1:-help}" in
        full)
            run_pipeline "quality,unit,integration,security,build,e2e,performance,chaos"
            ;;
        quick)
            run_pipeline "quality,unit"
            ;;
        custom)
            if [ -z "$2" ]; then
                log_error "Please specify stages for custom pipeline"
                show_help
                exit 1
            fi
            run_pipeline "$2"
            ;;
        setup-hooks)
            setup_git_hooks
            ;;
        install-tools)
            install_tools
            ;;
        clean)
            clean_artifacts
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"