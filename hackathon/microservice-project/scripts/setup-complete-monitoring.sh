#!/bin/bash

# Complete Monitoring Setup Script
# This script automates the entire monitoring stack deployment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MONITORING_DIR="$PROJECT_ROOT/monitoring"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"
WAIT_TIMEOUT=300  # 5 minutes

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for service to be ready
wait_for_service() {
    local service_name="$1"
    local url="$2"
    local timeout="$3"
    local counter=0
    
    print_status "Waiting for $service_name to be ready..."
    
    while [ $counter -lt $timeout ]; do
        if curl -s "$url" >/dev/null 2>&1; then
            print_success "$service_name is ready!"
            return 0
        fi
        
        sleep 5
        counter=$((counter + 5))
        echo -n "."
    done
    
    print_error "$service_name failed to start within $timeout seconds"
    return 1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command_exists docker-compose; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker daemon is not running. Please start Docker first."
        exit 1
    fi
    
    print_success "All prerequisites are met"
}

# Function to create monitoring directories
setup_directories() {
    print_status "Setting up monitoring directories..."
    
    # Create main monitoring directories
    mkdir -p "$MONITORING_DIR"/{prometheus,grafana,alertmanager}
    mkdir -p "$MONITORING_DIR/grafana/provisioning"/{datasources,dashboards}
    mkdir -p "$MONITORING_DIR/grafana/dashboards"/{system,application,performance,alerts}
    
    # Set proper permissions
    chmod -R 755 "$MONITORING_DIR"
    
    print_success "Monitoring directories created"
}

# Function to start monitoring services
start_monitoring_services() {
    print_status "Starting monitoring services..."
    
    cd "$MONITORING_DIR"
    
    # Pull latest images
    print_status "Pulling latest Docker images..."
    docker-compose pull
    
    # Start services
    print_status "Starting monitoring stack..."
    docker-compose up -d
    
    print_success "Monitoring services started"
}

# Function to verify services
verify_services() {
    print_status "Verifying monitoring services..."
    
    # Wait for services to be ready
    wait_for_service "Prometheus" "http://localhost:9090/-/ready" 60
    wait_for_service "Grafana" "http://localhost:3000/api/health" 60
    wait_for_service "Alertmanager" "http://localhost:9093/-/ready" 30
    wait_for_service "Jaeger" "http://localhost:16686" 30
    wait_for_service "Node Exporter" "http://localhost:9100/metrics" 30
    wait_for_service "cAdvisor" "http://localhost:8080/metrics" 30
    
    print_success "All monitoring services are running"
}

# Function to check microservice metrics
check_microservice_metrics() {
    print_status "Checking microservice metrics endpoints..."
    
    local services=("auth-service:8081" "file-service:8082" "user-service:8083" "analytics-service:8085")
    local available_services=0
    
    for service in "${services[@]}"; do
        local service_name=$(echo "$service" | cut -d':' -f1)
        local port=$(echo "$service" | cut -d':' -f2)
        
        if curl -s "http://localhost:$port/metrics" >/dev/null 2>&1; then
            print_success "$service_name metrics endpoint is available"
            available_services=$((available_services + 1))
        else
            print_warning "$service_name metrics endpoint is not available (service may not be running)"
        fi
    done
    
    if [ $available_services -eq 0 ]; then
        print_warning "No microservice metrics endpoints are available. You may need to start the microservices separately."
    else
        print_success "$available_services out of ${#services[@]} microservice metrics endpoints are available"
    fi
}

# Function to setup Grafana
setup_grafana() {
    print_status "Setting up Grafana..."
    
    # Wait a bit more for Grafana to fully initialize
    sleep 10
    
    # Check if Grafana is accessible
    if curl -s -u admin:admin "http://localhost:3000/api/datasources" >/dev/null 2>&1; then
        print_success "Grafana is accessible with default credentials"
    else
        print_warning "Grafana may not be fully ready yet. You can access it manually at http://localhost:3000"
    fi
}

# Function to display access information
display_access_info() {
    print_success "\n🎉 Monitoring stack deployment completed successfully!"
    
    echo -e "\n${BLUE}📊 Access Points:${NC}"
    echo -e "  • Grafana Dashboard: ${GREEN}http://localhost:3000${NC} (admin/admin)"
    echo -e "  • Prometheus: ${GREEN}http://localhost:9090${NC}"
    echo -e "  • Alertmanager: ${GREEN}http://localhost:9093${NC}"
    echo -e "  • Jaeger Tracing: ${GREEN}http://localhost:16686${NC}"
    echo -e "  • cAdvisor: ${GREEN}http://localhost:8080${NC}"
    echo -e "  • Node Exporter: ${GREEN}http://localhost:9100${NC}"
    
    echo -e "\n${BLUE}📁 Important Files:${NC}"
    echo -e "  • Setup Guide: ${GREEN}MONITORING_SETUP_GUIDE.md${NC}"
    echo -e "  • Quick Reference: ${GREEN}MONITORING_QUICK_REFERENCE.md${NC}"
    echo -e "  • Configuration: ${GREEN}monitoring/docker-compose.yml${NC}"
    
    echo -e "\n${BLUE}🔧 Next Steps:${NC}"
    echo -e "  1. Change Grafana admin password: ${YELLOW}docker exec -it microservice-project-grafana-1 grafana-cli admin reset-admin-password newpassword${NC}"
    echo -e "  2. Configure alert channels in Alertmanager"
    echo -e "  3. Start microservices: ${YELLOW}docker-compose up -d${NC} (in project root)"
    echo -e "  4. Import custom dashboards as needed"
    
    echo -e "\n${BLUE}📚 Documentation:${NC}"
    echo -e "  • Complete setup guide: ${GREEN}./MONITORING_SETUP_GUIDE.md${NC}"
    echo -e "  • Quick reference: ${GREEN}./MONITORING_QUICK_REFERENCE.md${NC}"
    
    echo -e "\n${BLUE}🆘 Troubleshooting:${NC}"
    echo -e "  • Check service status: ${YELLOW}docker-compose ps${NC}"
    echo -e "  • View logs: ${YELLOW}docker-compose logs -f [service-name]${NC}"
    echo -e "  • Restart services: ${YELLOW}docker-compose restart${NC}"
}

# Function to cleanup on error
cleanup_on_error() {
    print_error "Setup failed. Cleaning up..."
    cd "$MONITORING_DIR" 2>/dev/null && docker-compose down 2>/dev/null || true
    exit 1
}

# Main execution
main() {
    echo -e "${BLUE}🚀 Starting Complete Monitoring Stack Setup${NC}\n"
    
    # Set trap for cleanup on error
    trap cleanup_on_error ERR
    
    # Execute setup steps
    check_prerequisites
    setup_directories
    start_monitoring_services
    verify_services
    setup_grafana
    check_microservice_metrics
    
    # Display final information
    display_access_info
    
    print_success "\n✅ Monitoring stack setup completed successfully!"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            echo "Complete Monitoring Stack Setup Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --help, -h     Show this help message"
            echo "  --cleanup      Stop and remove all monitoring services"
            echo "  --status       Show status of monitoring services"
            echo "  --restart      Restart all monitoring services"
            echo ""
            echo "This script will:"
            echo "  1. Check prerequisites (Docker, Docker Compose)"
            echo "  2. Create monitoring directories"
            echo "  3. Start monitoring services (Prometheus, Grafana, etc.)"
            echo "  4. Verify all services are running"
            echo "  5. Display access information"
            exit 0
            ;;
        --cleanup)
            print_status "Stopping and removing monitoring services..."
            cd "$MONITORING_DIR" && docker-compose down -v
            print_success "Monitoring services stopped and removed"
            exit 0
            ;;
        --status)
            print_status "Checking monitoring services status..."
            cd "$MONITORING_DIR" && docker-compose ps
            exit 0
            ;;
        --restart)
            print_status "Restarting monitoring services..."
            cd "$MONITORING_DIR" && docker-compose restart
            print_success "Monitoring services restarted"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
    shift
done

# Run main function
main