#!/bin/bash

# CI/CD Pipeline Setup Script
# This script helps set up and manage the CI/CD pipeline for the microservice project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    
    if ! command_exists git; then
        missing_tools+=("git")
    fi
    
    if ! command_exists go; then
        missing_tools+=("go")
    fi
    
    if ! command_exists docker; then
        missing_tools+=("docker")
    fi
    
    if ! command_exists docker-compose; then
        missing_tools+=("docker-compose")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools and run this script again."
        exit 1
    fi
    
    log_success "All prerequisites are installed"
}

# Validate project structure
validate_project_structure() {
    log_info "Validating project structure..."
    
    local required_dirs=("services" "tests" "shared" ".github/workflows")
    local required_files=("Makefile" "docker-compose.yml" "go.mod")
    
    for dir in "${required_dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            log_error "Required directory not found: $dir"
            exit 1
        fi
    done
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_error "Required file not found: $file"
            exit 1
        fi
    done
    
    log_success "Project structure is valid"
}

# Check workflow files
check_workflows() {
    log_info "Checking CI/CD workflow files..."
    
    local workflows=("ci-cd.yml" "dependency-update.yml" "release.yml" "monitoring.yml")
    
    for workflow in "${workflows[@]}"; do
        if [ ! -f ".github/workflows/$workflow" ]; then
            log_warning "Workflow file not found: $workflow"
        else
            log_success "Found workflow: $workflow"
        fi
    done
}

# Test local build
test_local_build() {
    log_info "Testing local build..."
    
    if ! make build; then
        log_error "Local build failed"
        exit 1
    fi
    
    log_success "Local build successful"
}

# Test local tests
test_local_tests() {
    log_info "Running local tests..."
    
    if ! make test-unit; then
        log_error "Unit tests failed"
        exit 1
    fi
    
    log_success "Unit tests passed"
    
    # Run integration tests if services are available
    if docker-compose ps | grep -q "Up"; then
        log_info "Services are running, testing integration tests..."
        if ! make test-integration; then
            log_warning "Integration tests failed (this might be expected if services are not properly configured)"
        else
            log_success "Integration tests passed"
        fi
    else
        log_info "Services not running, skipping integration tests"
    fi
}

# Check Docker setup
check_docker_setup() {
    log_info "Checking Docker setup..."
    
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker is not running"
        exit 1
    fi
    
    log_success "Docker is running"
    
    # Test Docker Compose
    if ! docker-compose config >/dev/null 2>&1; then
        log_error "Docker Compose configuration is invalid"
        exit 1
    fi
    
    log_success "Docker Compose configuration is valid"
}

# Setup GitHub repository
setup_github_repo() {
    log_info "Setting up GitHub repository..."
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Check if remote origin exists
    if ! git remote get-url origin >/dev/null 2>&1; then
        log_warning "No remote origin found. Please add a GitHub remote:"
        log_info "git remote add origin https://github.com/username/repository.git"
        return
    fi
    
    local remote_url=$(git remote get-url origin)
    log_success "GitHub remote found: $remote_url"
    
    # Check if workflows are pushed
    if git ls-remote --heads origin | grep -q "main\|master"; then
        log_success "Repository has main/master branch"
    else
        log_warning "Repository doesn't have main/master branch"
    fi
}

# Generate secrets template
generate_secrets_template() {
    log_info "Generating secrets template..."
    
    cat > .github-secrets-template.txt << EOF
# GitHub Secrets Template
# Add these secrets to your GitHub repository settings

# Container Registry (if using external registry)
DOCKER_REGISTRY_USERNAME=your_username
DOCKER_REGISTRY_PASSWORD=your_password

# Kubernetes (if deploying to Kubernetes)
KUBE_CONFIG=base64_encoded_kubeconfig

# Notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
EMAIL_SMTP_PASSWORD=your_smtp_password

# Cloud Providers (if applicable)
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
GCP_SERVICE_ACCOUNT_KEY=base64_encoded_service_account_json

# Monitoring (if using external services)
DATADOG_API_KEY=your_datadog_api_key
NEW_RELIC_LICENSE_KEY=your_newrelic_license_key

# Security Scanning (if using external services)
SNYK_TOKEN=your_snyk_token
SONARQUBE_TOKEN=your_sonarqube_token
EOF
    
    log_success "Secrets template generated: .github-secrets-template.txt"
    log_info "Please review and add the required secrets to your GitHub repository"
}

# Create environment files
create_env_files() {
    log_info "Creating environment files..."
    
    # Development environment
    if [ ! -f ".env.development" ]; then
        cat > .env.development << EOF
# Development Environment
ENVIRONMENT=development
DEBUG=true

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=microservice_dev

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Services
AUTH_SERVICE_PORT=8001
FILE_SERVICE_PORT=8002
USER_SERVICE_PORT=8003
ANALYTICS_SERVICE_PORT=8004
API_GATEWAY_PORT=8000
EOF
        log_success "Created .env.development"
    fi
    
    # Testing environment
    if [ ! -f ".env.testing" ]; then
        cat > .env.testing << EOF
# Testing Environment
ENVIRONMENT=testing
DEBUG=false

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=microservice_test

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Services
AUTH_SERVICE_PORT=8001
FILE_SERVICE_PORT=8002
USER_SERVICE_PORT=8003
ANALYTICS_SERVICE_PORT=8004
API_GATEWAY_PORT=8000
EOF
        log_success "Created .env.testing"
    fi
}

# Setup pre-commit hooks
setup_pre_commit_hooks() {
    log_info "Setting up pre-commit hooks..."
    
    mkdir -p .git/hooks
    
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Pre-commit hook for Go projects
set -e

echo "Running pre-commit checks..."

# Check if Go files have been modified
go_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [ -n "$go_files" ]; then
    echo "Checking Go files..."
    
    # Format Go code
    echo "Formatting Go code..."
    gofmt -w $go_files
    git add $go_files
    
    # Run go vet
    echo "Running go vet..."
    go vet ./...
    
    # Run tests
    echo "Running tests..."
    go test -short ./...
    
    echo "Pre-commit checks passed!"
else
    echo "No Go files to check"
fi
EOF
    
    chmod +x .git/hooks/pre-commit
    log_success "Pre-commit hooks installed"
}

# Main setup function
setup_cicd() {
    log_info "Setting up CI/CD pipeline..."
    
    check_prerequisites
    validate_project_structure
    check_workflows
    check_docker_setup
    setup_github_repo
    generate_secrets_template
    create_env_files
    setup_pre_commit_hooks
    
    log_success "CI/CD pipeline setup completed!"
    
    echo
    log_info "Next steps:"
    echo "1. Review and add GitHub secrets from .github-secrets-template.txt"
    echo "2. Configure branch protection rules in GitHub"
    echo "3. Test the pipeline by creating a pull request"
    echo "4. Review the CI/CD documentation in docs/CI-CD-PIPELINE.md"
}

# Test CI/CD pipeline locally
test_pipeline() {
    log_info "Testing CI/CD pipeline locally..."
    
    test_local_build
    test_local_tests
    
    log_info "Testing Docker build..."
    if ! docker-compose build; then
        log_error "Docker build failed"
        exit 1
    fi
    log_success "Docker build successful"
    
    log_info "Testing services startup..."
    docker-compose up -d
    sleep 10
    
    # Check if services are healthy
    local services=("auth-service" "file-service" "user-service" "analytics-service")
    for service in "${services[@]}"; do
        if docker-compose ps | grep "$service" | grep -q "Up"; then
            log_success "$service is running"
        else
            log_warning "$service is not running properly"
        fi
    done
    
    docker-compose down
    log_success "Pipeline test completed"
}

# Show help
show_help() {
    echo "CI/CD Pipeline Setup Script"
    echo
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  setup     Set up the CI/CD pipeline (default)"
    echo "  test      Test the pipeline locally"
    echo "  check     Check prerequisites and project structure"
    echo "  secrets   Generate secrets template"
    echo "  hooks     Setup pre-commit hooks"
    echo "  help      Show this help message"
    echo
}

# Main script
main() {
    case "${1:-setup}" in
        setup)
            setup_cicd
            ;;
        test)
            test_pipeline
            ;;
        check)
            check_prerequisites
            validate_project_structure
            check_workflows
            ;;
        secrets)
            generate_secrets_template
            ;;
        hooks)
            setup_pre_commit_hooks
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