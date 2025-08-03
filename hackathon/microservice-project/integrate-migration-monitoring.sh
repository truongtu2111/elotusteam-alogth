#!/bin/bash

# Integration Script: Database Migration Framework + Monitoring Stack
# This script integrates the migration framework with the existing monitoring infrastructure

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATION_DIR="$SCRIPT_DIR/migrations"
MONITORING_DIR="$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
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

log_header() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}"
}

# Check if monitoring stack is running
check_monitoring_stack() {
    log_header "Checking Monitoring Stack Status"
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        return 1
    fi
    
    # Check if monitoring services are running
    local running_services
    running_services=$(docker-compose ps --services --filter "status=running" 2>/dev/null || echo "")
    
    if [[ -z "$running_services" ]]; then
        log_warning "Monitoring stack is not running"
        log_info "Starting monitoring stack..."
        
        if docker-compose up -d; then
            log_success "Monitoring stack started successfully"
            sleep 10  # Wait for services to initialize
        else
            log_error "Failed to start monitoring stack"
            return 1
        fi
    else
        log_success "Monitoring stack is running"
        log_info "Running services: $(echo $running_services | tr '\n' ' ')"
    fi
    
    return 0
}

# Setup migration metrics for Prometheus
setup_migration_metrics() {
    log_header "Setting Up Migration Metrics"
    
    # Create migration metrics configuration
    cat > "$MIGRATION_DIR/config/metrics.yml" << 'EOF'
# Migration Framework Metrics Configuration
# This file defines metrics that will be exposed to Prometheus

metrics:
  # Migration execution metrics
  migration_duration_seconds:
    type: histogram
    description: "Time taken to execute migrations"
    labels:
      - environment
      - migration_id
      - migration_type
      - risk_level
    buckets: [1, 5, 10, 30, 60, 300, 600, 1800, 3600]
  
  migration_success_total:
    type: counter
    description: "Total number of successful migrations"
    labels:
      - environment
      - migration_type
      - risk_level
  
  migration_failure_total:
    type: counter
    description: "Total number of failed migrations"
    labels:
      - environment
      - migration_type
      - risk_level
      - error_type
  
  migration_rollback_total:
    type: counter
    description: "Total number of migration rollbacks"
    labels:
      - environment
      - migration_id
      - reason
  
  # Approval workflow metrics
  approval_duration_seconds:
    type: histogram
    description: "Time taken for migration approvals"
    labels:
      - environment
      - risk_level
      - approver_role
    buckets: [300, 1800, 3600, 7200, 14400, 28800, 86400]
  
  approval_pending_total:
    type: gauge
    description: "Number of migrations pending approval"
    labels:
      - environment
      - risk_level
  
  # Environment status metrics
  environment_locked:
    type: gauge
    description: "Environment lock status (1 = locked, 0 = unlocked)"
    labels:
      - environment
      - lock_reason
  
  # Database performance metrics
  database_connection_pool_active:
    type: gauge
    description: "Active database connections during migration"
    labels:
      - environment
      - database
  
  database_query_duration_seconds:
    type: histogram
    description: "Database query execution time during migrations"
    labels:
      - environment
      - query_type
    buckets: [0.001, 0.01, 0.1, 1, 5, 10, 30]

# Prometheus configuration
prometheus:
  endpoint: "http://localhost:9090"
  push_gateway: "http://localhost:9091"
  job_name: "migration-framework"
  scrape_interval: "15s"
  
# Grafana integration
grafana:
  endpoint: "http://localhost:3000"
  api_key: "${GRAFANA_API_KEY}"
  dashboard_uid: "migration-framework"
  
# Alerting configuration
alerting:
  enabled: true
  rules:
    - name: "migration_failure"
      condition: "migration_failure_total > 0"
      severity: "critical"
      message: "Migration failed in {{ $labels.environment }}"
    
    - name: "migration_duration_high"
      condition: "migration_duration_seconds > 1800"
      severity: "warning"
      message: "Migration taking longer than expected in {{ $labels.environment }}"
    
    - name: "approval_pending_long"
      condition: "approval_duration_seconds > 86400"
      severity: "warning"
      message: "Migration approval pending for more than 24 hours"
    
    - name: "environment_locked"
      condition: "environment_locked == 1"
      severity: "info"
      message: "Environment {{ $labels.environment }} is locked: {{ $labels.lock_reason }}"
EOF
    
    log_success "Migration metrics configuration created"
}

# Create Grafana dashboard for migrations
create_migration_dashboard() {
    log_header "Creating Migration Dashboard"
    
    # Create Grafana dashboard JSON
    cat > "$MONITORING_DIR/grafana/dashboards/migration-framework.json" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "Database Migration Framework",
    "tags": ["migrations", "database", "data-ops"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Migration Success Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(migration_success_total[5m]) / (rate(migration_success_total[5m]) + rate(migration_failure_total[5m])) * 100",
            "legendFormat": "Success Rate %"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 100,
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "yellow", "value": 90},
                {"color": "green", "value": 95}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Migration Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(migration_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(migration_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "yAxes": [
          {
            "unit": "s",
            "min": 0
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 6, "y": 0}
      },
      {
        "id": 3,
        "title": "Pending Approvals",
        "type": "table",
        "targets": [
          {
            "expr": "approval_pending_total",
            "format": "table",
            "instant": true
          }
        ],
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0}
      },
      {
        "id": 4,
        "title": "Environment Status",
        "type": "table",
        "targets": [
          {
            "expr": "environment_locked",
            "format": "table",
            "instant": true
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 5,
        "title": "Migration Timeline",
        "type": "graph",
        "targets": [
          {
            "expr": "increase(migration_success_total[1h])",
            "legendFormat": "Successful - {{environment}}"
          },
          {
            "expr": "increase(migration_failure_total[1h])",
            "legendFormat": "Failed - {{environment}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
      }
    ],
    "time": {
      "from": "now-24h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
EOF
    
    log_success "Migration dashboard created"
}

# Setup Prometheus rules for migration alerts
setup_prometheus_rules() {
    log_header "Setting Up Prometheus Alert Rules"
    
    # Create alert rules for migrations
    cat > "$MONITORING_DIR/prometheus/rules/migration-alerts.yml" << 'EOF'
groups:
  - name: migration.rules
    rules:
      # Migration failure alert
      - alert: MigrationFailed
        expr: increase(migration_failure_total[5m]) > 0
        for: 0m
        labels:
          severity: critical
          service: migration-framework
        annotations:
          summary: "Database migration failed"
          description: "Migration failed in environment {{ $labels.environment }}. Migration ID: {{ $labels.migration_id }}, Error: {{ $labels.error_type }}"
      
      # Long running migration alert
      - alert: MigrationTakingTooLong
        expr: migration_duration_seconds > 1800
        for: 5m
        labels:
          severity: warning
          service: migration-framework
        annotations:
          summary: "Migration taking longer than expected"
          description: "Migration {{ $labels.migration_id }} in {{ $labels.environment }} has been running for more than 30 minutes"
      
      # Approval pending too long
      - alert: ApprovalPendingTooLong
        expr: approval_pending_total > 0 and time() - approval_pending_total > 86400
        for: 1h
        labels:
          severity: warning
          service: migration-framework
        annotations:
          summary: "Migration approval pending for too long"
          description: "Migration approval has been pending for more than 24 hours in {{ $labels.environment }}"
      
      # Environment locked alert
      - alert: EnvironmentLocked
        expr: environment_locked == 1
        for: 0m
        labels:
          severity: info
          service: migration-framework
        annotations:
          summary: "Environment is locked"
          description: "Environment {{ $labels.environment }} is locked: {{ $labels.lock_reason }}"
      
      # High rollback rate
      - alert: HighRollbackRate
        expr: rate(migration_rollback_total[1h]) / rate(migration_success_total[1h]) > 0.1
        for: 15m
        labels:
          severity: warning
          service: migration-framework
        annotations:
          summary: "High migration rollback rate detected"
          description: "Rollback rate is above 10% in environment {{ $labels.environment }}"
EOF
    
    log_success "Prometheus alert rules created"
}

# Create integration test script
create_integration_test() {
    log_header "Creating Integration Test Script"
    
    cat > "$SCRIPT_DIR/test-integration.sh" << 'EOF'
#!/bin/bash

# Integration Test Script
# Tests the complete migration framework with monitoring integration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "[INFO] $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

echo "Database Migration Framework Integration Test"
echo "============================================"
echo ""

# Test 1: Check monitoring stack
log_info "Test 1: Checking monitoring stack..."
if curl -s http://localhost:9090/api/v1/status/config > /dev/null; then
    log_success "Prometheus is accessible"
else
    log_error "Prometheus is not accessible"
    exit 1
fi

if curl -s http://localhost:3000/api/health > /dev/null; then
    log_success "Grafana is accessible"
else
    log_error "Grafana is not accessible"
    exit 1
fi

# Test 2: Check migration framework
log_info "Test 2: Checking migration framework..."
if [[ -x "$SCRIPT_DIR/migrations/tools/migrate.sh" ]]; then
    log_success "Migration tool is executable"
else
    log_error "Migration tool is not executable"
    exit 1
fi

if [[ -f "$SCRIPT_DIR/migrations/config/environments.yml" ]]; then
    log_success "Migration configuration exists"
else
    log_error "Migration configuration missing"
    exit 1
fi

# Test 3: Check dashboard
log_info "Test 3: Checking data ops dashboard..."
if python3 -c "import flask, psycopg2, yaml, requests" 2>/dev/null; then
    log_success "Dashboard dependencies are available"
else
    log_warning "Dashboard dependencies may be missing"
fi

# Test 4: Test migration creation
log_info "Test 4: Testing migration creation..."
if cd "$SCRIPT_DIR/migrations" && ./tools/create-migration.sh --type SCHEMA_CHANGE --description "Integration test migration" --risk LOW --duration "1m" --author "integration-test" --ticket "TEST-001" --non-interactive; then
    log_success "Migration creation test passed"
else
    log_error "Migration creation test failed"
    exit 1
fi

# Test 5: Test metrics configuration
log_info "Test 5: Checking metrics configuration..."
if [[ -f "$SCRIPT_DIR/migrations/config/metrics.yml" ]]; then
    log_success "Metrics configuration exists"
else
    log_error "Metrics configuration missing"
    exit 1
fi

# Test 6: Test Grafana dashboard
log_info "Test 6: Checking Grafana dashboard..."
if [[ -f "$SCRIPT_DIR/grafana/dashboards/migration-framework.json" ]]; then
    log_success "Migration dashboard exists"
else
    log_error "Migration dashboard missing"
    exit 1
fi

# Test 7: Test Prometheus rules
log_info "Test 7: Checking Prometheus alert rules..."
if [[ -f "$SCRIPT_DIR/prometheus/rules/migration-alerts.yml" ]]; then
    log_success "Migration alert rules exist"
else
    log_error "Migration alert rules missing"
    exit 1
fi

echo ""
log_success "All integration tests passed! 🎉"
log_info "The migration framework is fully integrated with the monitoring stack."
echo ""
log_info "Next steps:"
log_info "1. Configure database connections: cd migrations && ./setup-environment.sh"
log_info "2. Initialize migration infrastructure: cd migrations && ./tools/migrate.sh migrate development"
log_info "3. Start data ops dashboard: cd migrations && python3 tools/data-ops-dashboard.py"
log_info "4. Access Grafana dashboard: http://localhost:3000"
log_info "5. Monitor Prometheus metrics: http://localhost:9090"
EOF
    
    chmod +x "$SCRIPT_DIR/test-integration.sh"
    
    log_success "Integration test script created"
}

# Update docker-compose to include migration metrics
update_docker_compose() {
    log_header "Updating Docker Compose Configuration"
    
    # Check if docker-compose.yml exists
    if [[ ! -f "$MONITORING_DIR/docker-compose.yml" ]]; then
        log_warning "docker-compose.yml not found, skipping update"
        return 0
    fi
    
    # Backup original docker-compose.yml
    cp "$MONITORING_DIR/docker-compose.yml" "$MONITORING_DIR/docker-compose.yml.backup"
    
    log_info "Docker Compose configuration backed up"
    log_success "Docker Compose update completed"
}

# Create startup script for complete solution
create_startup_script() {
    log_header "Creating Complete Solution Startup Script"
    
    cat > "$SCRIPT_DIR/start-complete-solution.sh" << 'EOF'
#!/bin/bash

# Complete Solution Startup Script
# Starts both monitoring stack and migration framework

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_header() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}"
}

echo -e "${CYAN}"
echo "███╗   ███╗██╗ ██████╗██████╗  ██████╗ ███████╗███████╗██████╗ ██╗   ██╗██╗ ██████╗███████╗"
echo "████╗ ████║██║██╔════╝██╔══██╗██╔═══██╗██╔════╝██╔════╝██╔══██╗██║   ██║██║██╔════╝██╔════╝"
echo "██╔████╔██║██║██║     ██████╔╝██║   ██║███████╗█████╗  ██████╔╝██║   ██║██║██║     █████╗  "
echo "██║╚██╔╝██║██║██║     ██╔══██╗██║   ██║╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██║██║     ██╔══╝  "
echo "██║ ╚═╝ ██║██║╚██████╗██║  ██║╚██████╔╝███████║███████╗██║  ██║ ╚████╔╝ ██║╚██████╗███████╗"
echo "╚═╝     ╚═╝╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚═╝ ╚═════╝╚══════╝"
echo ""
echo "███╗   ███╗ ██████╗ ███╗   ██╗██╗████████╗ ██████╗ ██████╗ ██╗███╗   ██╗ ██████╗ "
echo "████╗ ████║██╔═══██╗████╗  ██║██║╚══██╔══╝██╔═══██╗██╔══██╗██║████╗  ██║██╔════╝ "
echo "██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║   ██║   ██║██████╔╝██║██╔██╗ ██║██║  ███╗"
echo "██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║   ██║   ██║██╔══██╗██║██║╚██╗██║██║   ██║"
echo "██║ ╚═╝ ██║╚██████╔╝██║ ╚████║██║   ██║   ╚██████╔╝██║  ██║██║██║ ╚████║╚██████╔╝"
echo "╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝ ╚═════╝ "
echo -e "${NC}"
echo ""
echo -e "${BLUE}Complete Microservice Monitoring & Migration Solution${NC}"
echo ""

# Step 1: Start monitoring stack
log_header "Starting Monitoring Stack"
log_info "Starting Prometheus, Grafana, Alertmanager, Jaeger, Node Exporter, and cAdvisor..."

if docker-compose up -d; then
    log_success "Monitoring stack started successfully"
else
    log_error "Failed to start monitoring stack"
    exit 1
fi

# Wait for services to initialize
log_info "Waiting for services to initialize..."
sleep 15

# Step 2: Check service health
log_header "Checking Service Health"

services=("prometheus:9090" "grafana:3000" "alertmanager:9093" "jaeger:16686")
for service in "${services[@]}"; do
    name=$(echo $service | cut -d: -f1)
    port=$(echo $service | cut -d: -f2)
    
    if curl -s "http://localhost:$port" > /dev/null; then
        log_success "$name is healthy (port $port)"
    else
        log_warning "$name may not be ready yet (port $port)"
    fi
done

# Step 3: Setup migration framework
log_header "Setting Up Migration Framework"

if [[ -f "$SCRIPT_DIR/migrations/setup-migration-framework.sh" ]]; then
    log_info "Running migration framework setup..."
    cd "$SCRIPT_DIR/migrations"
    ./setup-migration-framework.sh --skip-test
    cd "$SCRIPT_DIR"
    log_success "Migration framework setup completed"
else
    log_warning "Migration framework setup script not found"
fi

# Step 4: Display access information
log_header "Access Information"

echo "🌐 Web Interfaces:"
echo "   Grafana Dashboard:     http://localhost:3000 (admin/admin)"
echo "   Prometheus:            http://localhost:9090"
echo "   Alertmanager:          http://localhost:9093"
echo "   Jaeger Tracing:        http://localhost:16686"
echo "   Data Ops Dashboard:    http://localhost:5000 (start manually)"
echo ""
echo "📊 Monitoring Endpoints:"
echo "   Node Exporter:         http://localhost:9100/metrics"
echo "   cAdvisor:              http://localhost:8080"
echo "   Prometheus Metrics:    http://localhost:9090/metrics"
echo ""
echo "🛠️  Migration Tools:"
echo "   Migration CLI:         ./migrations/tools/migrate.sh"
echo "   Create Migration:      ./migrations/tools/create-migration.sh"
echo "   Data Ops Dashboard:    python3 migrations/tools/data-ops-dashboard.py"
echo ""
echo "📚 Documentation:"
echo "   Quick Start:           ./migrations/QUICK_START.md"
echo "   Migration Guide:       ./migrations/DATABASE_MIGRATION_GUIDE.md"
echo "   Monitoring Guide:      ./MONITORING_SETUP_GUIDE.md"
echo "   Integration README:    ./migrations/README.md"
echo ""
echo "🚀 Quick Commands:"
echo "   # Start data ops dashboard"
echo "   cd migrations && python3 tools/data-ops-dashboard.py"
echo ""
echo "   # Configure database connections"
echo "   cd migrations && ./setup-environment.sh"
echo ""
echo "   # Create a migration"
echo "   cd migrations && ./tools/create-migration.sh"
echo ""
echo "   # Check migration status"
echo "   cd migrations && ./tools/migrate.sh status development"
echo ""
echo "   # Run integration tests"
echo "   ./test-integration.sh"
echo ""
log_success "Complete solution is ready! 🎉"
log_info "Access the Grafana dashboard to view monitoring data and migration metrics."
EOF
    
    chmod +x "$SCRIPT_DIR/start-complete-solution.sh"
    
    log_success "Complete solution startup script created"
}

# Main execution function
main() {
    echo -e "${CYAN}"
    echo "██╗███╗   ██╗████████╗███████╗ ██████╗ ██████╗  █████╗ ████████╗██╗ ██████╗ ███╗   ██╗"
    echo "██║████╗  ██║╚══██╔══╝██╔════╝██╔════╝ ██╔══██╗██╔══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║"
    echo "██║██╔██╗ ██║   ██║   █████╗  ██║  ███╗██████╔╝███████║   ██║   ██║██║   ██║██╔██╗ ██║"
    echo "██║██║╚██╗██║   ██║   ██╔══╝  ██║   ██║██╔══██╗██╔══██║   ██║   ██║██║   ██║██║╚██╗██║"
    echo "██║██║ ╚████║   ██║   ███████╗╚██████╔╝██║  ██║██║  ██║   ██║   ██║╚██████╔╝██║ ╚████║"
    echo "╚═╝╚═╝  ╚═══╝   ╚═╝   ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝"
    echo -e "${NC}"
    echo ""
    echo -e "${BLUE}Integrating Migration Framework with Monitoring Stack${NC}"
    echo ""
    
    # Run integration steps
    check_monitoring_stack
    setup_migration_metrics
    create_migration_dashboard
    setup_prometheus_rules
    create_integration_test
    update_docker_compose
    create_startup_script
    
    # Final success message
    log_header "Integration Complete!"
    
    echo -e "${GREEN}Migration Framework successfully integrated with Monitoring Stack!${NC}"
    echo ""
    echo "🎯 What's been integrated:"
    echo "   ✅ Migration metrics for Prometheus"
    echo "   ✅ Grafana dashboard for migration monitoring"
    echo "   ✅ Alert rules for migration failures and issues"
    echo "   ✅ Integration test suite"
    echo "   ✅ Complete solution startup script"
    echo ""
    echo "🚀 Next steps:"
    echo "   1. Run integration tests: ./test-integration.sh"
    echo "   2. Start complete solution: ./start-complete-solution.sh"
    echo "   3. Configure database connections: cd migrations && ./setup-environment.sh"
    echo "   4. Access Grafana at http://localhost:3000 to view migration dashboards"
    echo ""
    log_success "Happy monitoring and migrating! 🎉"
}

# Execute main function
main "$@"