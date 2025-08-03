# Local CI/CD Pipeline

This document explains how to set up and use the local CI/CD pipeline that mirrors the GitHub Actions workflow and automatically runs on code pushes.

## Overview

The local CI/CD pipeline allows you to:
- Run the same tests and checks locally that run in GitHub Actions
- Automatically trigger pipeline on git push (via Git hooks)
- Get immediate feedback before pushing to remote repository
- Generate detailed reports and coverage analysis
- Customize pipeline behavior for different scenarios

## Quick Start

### 1. Setup the Pipeline

```bash
# Make the script executable (if not already done)
chmod +x scripts/local-cicd.sh

# Install required tools
./scripts/local-cicd.sh install-tools

# Setup Git hooks for automatic execution
./scripts/local-cicd.sh setup-hooks
```

### 2. Run the Pipeline

```bash
# Run quick pipeline (quality checks + unit tests)
./scripts/local-cicd.sh quick

# Run full pipeline (all stages)
./scripts/local-cicd.sh full

# Run custom stages
./scripts/local-cicd.sh custom quality,unit,build
```

### 3. Automatic Execution

Once Git hooks are set up, the pipeline will automatically run:
- **Before push**: Quick pipeline (quality + unit tests)
- **After commit**: Full pipeline (optional, disabled by default)

## Pipeline Stages

### 1. Code Quality & Security
- **Go Formatting**: Checks and fixes code formatting
- **Go Vet**: Static analysis for common errors
- **Linting**: Code quality checks with golangci-lint
- **Security Scanning**: Vulnerability detection with gosec
- **Dependency Check**: Vulnerability scanning with nancy

### 2. Unit Tests
- **Test Execution**: Runs all unit tests with race detection
- **Coverage Analysis**: Generates coverage reports
- **Coverage Threshold**: Validates minimum coverage (configurable)

### 3. Integration Tests
- **Service Setup**: Starts PostgreSQL and Redis containers
- **Database Tests**: Tests database interactions
- **Service Communication**: Tests inter-service communication
- **Cleanup**: Stops test services

### 4. Security Tests
- **Security Test Suite**: Runs dedicated security tests
- **Vulnerability Assessment**: Additional security validations

### 5. Build Docker Images
- **Service Images**: Builds Docker images for all microservices
- **Multi-platform**: Supports different architectures
- **Tagging**: Tags images with git commit hash

### 6. End-to-End Tests
- **Full Stack**: Starts all services with docker-compose
- **API Testing**: Tests complete user journeys
- **Health Checks**: Validates service health endpoints
- **Cleanup**: Stops all services

### 7. Performance Tests (Optional)
- **Benchmarks**: Runs performance benchmarks
- **Load Testing**: Basic load testing scenarios
- **Metrics**: Collects performance metrics

### 8. Chaos Engineering Tests (Optional)
- **Resilience Testing**: Tests system resilience
- **Failure Scenarios**: Simulates various failure conditions
- **Recovery Testing**: Validates recovery mechanisms

## Configuration

### Pipeline Configuration

Customize the pipeline behavior by editing `.local-cicd.config`:

```bash
# Edit configuration
vim .local-cicd.config
```

Key configuration options:

```bash
# Default stages to run
DEFAULT_STAGES="quality,unit,integration,security,build,e2e"

# Coverage threshold
COVERAGE_THRESHOLD=80

# Enable/disable Git hooks
ENABLE_PRE_PUSH_HOOK=true
ENABLE_POST_COMMIT_HOOK=false

# Notification settings
ENABLE_NOTIFICATIONS=false
NOTIFICATION_WEBHOOK_URL="https://hooks.slack.com/..."
```

### Branch-Specific Configuration

Different settings for different branches:

```bash
# Main branch - strict requirements
[branch.main]
STAGES="quality,unit,integration,security,build,e2e,performance"
COVERAGE_THRESHOLD=85
FAIL_ON_SECURITY_ISSUES=true

# Feature branches - faster feedback
[branch.feature/*]
STAGES="quality,unit,integration"
COVERAGE_THRESHOLD=75
FAIL_ON_SECURITY_ISSUES=false
```

## Usage Examples

### Basic Usage

```bash
# Quick check before committing
./scripts/local-cicd.sh quick

# Full pipeline before major changes
./scripts/local-cicd.sh full

# Only run tests
./scripts/local-cicd.sh custom unit,integration

# Only build and test
./scripts/local-cicd.sh custom quality,unit,build
```

### Development Workflow

```bash
# 1. Make changes to code
vim services/auth/handler.go

# 2. Run quick pipeline
./scripts/local-cicd.sh quick

# 3. If tests pass, commit changes
git add .
git commit -m "feat: add new authentication method"

# 4. Push changes (pre-push hook runs automatically)
git push origin feature/new-auth
```

### Git Hooks Integration

Once hooks are set up:

```bash
# This will automatically run quick pipeline before push
git push origin main

# If pipeline fails, push is aborted
# Fix issues and try again
./scripts/local-cicd.sh quick  # Debug locally
git push origin main           # Try push again
```

## Reports and Logs

### Generated Files

- **Coverage Report**: `coverage.html` - Visual coverage report
- **Pipeline Logs**: `.local-cicd-logs/pipeline_TIMESTAMP.log`
- **HTML Report**: `.local-cicd-logs/report_TIMESTAMP.html`
- **Test Results**: Various test output files

### Viewing Reports

```bash
# Open coverage report
open coverage.html

# View latest pipeline log
tail -f .local-cicd-logs/pipeline_*.log

# Open latest HTML report
open .local-cicd-logs/report_*.html
```

### Log Management

```bash
# Clean old logs and artifacts
./scripts/local-cicd.sh clean

# View log directory
ls -la .local-cicd-logs/
```

## Tool Installation

### Required Tools

```bash
# Install Go tools automatically
./scripts/local-cicd.sh install-tools
```

This installs:
- `golangci-lint` - Code linting
- `gosec` - Security scanning
- `nancy` - Dependency vulnerability scanning

### Manual Installation

```bash
# golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# nancy
go install github.com/sonatypecommunity/nancy@latest
```

### Docker Requirements

- Docker Desktop or Docker Engine
- docker-compose
- Sufficient disk space for images

## Troubleshooting

### Common Issues

#### 1. Docker Not Running
```bash
Error: Docker is not running
```
**Solution**: Start Docker Desktop or Docker daemon

#### 2. Port Conflicts
```bash
Error: Port 5432 already in use
```
**Solution**: Stop conflicting services or change ports in docker-compose.yml

#### 3. Permission Denied
```bash
Error: Permission denied
```
**Solution**: Make script executable
```bash
chmod +x scripts/local-cicd.sh
```

#### 4. Missing Tools
```bash
Warning: golangci-lint not found
```
**Solution**: Install missing tools
```bash
./scripts/local-cicd.sh install-tools
```

#### 5. Test Failures
```bash
Error: Unit tests failed
```
**Solution**: Check test logs and fix failing tests
```bash
# View detailed logs
cat .local-cicd-logs/pipeline_*.log

# Run tests manually for debugging
go test -v ./tests/unit/...
```

### Debug Mode

For detailed debugging:

```bash
# Set debug mode in configuration
echo "LOG_LEVEL=debug" >> .local-cicd.config

# Run with verbose output
./scripts/local-cicd.sh quick 2>&1 | tee debug.log
```

### Performance Issues

```bash
# Reduce parallel jobs
echo "PARALLEL_JOBS=2" >> .local-cicd.config

# Skip heavy stages for development
./scripts/local-cicd.sh custom quality,unit

# Use short mode for performance tests
echo "PERFORMANCE_TEST_MODE=short" >> .local-cicd.config
```

## Integration with IDEs

### VS Code

Add tasks to `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Local CI/CD - Quick",
            "type": "shell",
            "command": "./scripts/local-cicd.sh",
            "args": ["quick"],
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "Local CI/CD - Full",
            "type": "shell",
            "command": "./scripts/local-cicd.sh",
            "args": ["full"],
            "group": "test"
        }
    ]
}
```

### GoLand/IntelliJ

Add run configurations:
1. Go to Run → Edit Configurations
2. Add new Shell Script configuration
3. Set script path to `scripts/local-cicd.sh`
4. Set arguments to `quick` or `full`

## Customization

### Custom Stages

Add custom stages by modifying the script:

```bash
# Add custom stage function
stage_custom_checks() {
    log_stage "Custom Checks"
    
    # Your custom logic here
    run_with_log "your-custom-command" "Custom validation"
}

# Add to pipeline
case "$stage" in
    custom-checks)
        if ! stage_custom_checks; then
            failed_stages+=("Custom Checks")
        fi
        ;;
esac
```

### Custom Notifications

Add Slack notifications:

```bash
# In .local-cicd.config
ENABLE_NOTIFICATIONS=true
NOTIFICATION_WEBHOOK_URL="https://hooks.slack.com/services/..."
NOTIFY_ON_FAILURE=true
```

### Environment-Specific Settings

Create environment-specific configs:

```bash
# .local-cicd.development.config
STAGES="quality,unit"
COVERAGE_THRESHOLD=70

# .local-cicd.production.config
STAGES="quality,unit,integration,security,build,e2e,performance,chaos"
COVERAGE_THRESHOLD=90
```

## Best Practices

### Development Workflow

1. **Run Quick Pipeline Frequently**: Use `quick` mode during development
2. **Full Pipeline Before Push**: Run `full` mode before important pushes
3. **Fix Issues Immediately**: Don't ignore pipeline failures
4. **Monitor Coverage**: Keep test coverage above threshold
5. **Review Reports**: Check generated reports for insights

### Performance Optimization

1. **Use Appropriate Stages**: Don't run unnecessary stages during development
2. **Parallel Execution**: Configure parallel jobs based on your system
3. **Docker Optimization**: Use Docker layer caching
4. **Selective Testing**: Run only affected tests when possible

### Security

1. **Regular Scans**: Run security scans regularly
2. **Update Dependencies**: Keep dependencies up to date
3. **Review Security Reports**: Address security findings promptly
4. **Secure Configuration**: Don't commit sensitive configuration

## Advanced Usage

### Conditional Execution

Run different pipelines based on changes:

```bash
# Only run if Go files changed
if git diff --name-only HEAD~1 | grep -q '\.go$'; then
    ./scripts/local-cicd.sh full
else
    echo "No Go files changed, skipping pipeline"
fi
```

### Integration with Make

Add to Makefile:

```makefile
.PHONY: ci-quick ci-full ci-custom

ci-quick:
	./scripts/local-cicd.sh quick

ci-full:
	./scripts/local-cicd.sh full

ci-custom:
	./scripts/local-cicd.sh custom $(STAGES)
```

### Parallel Pipeline Execution

Run multiple pipelines in parallel:

```bash
# Run different test suites in parallel
./scripts/local-cicd.sh custom unit &
./scripts/local-cicd.sh custom integration &
./scripts/local-cicd.sh custom security &
wait
```

## Support

For issues or questions:

1. Check the troubleshooting section
2. Review pipeline logs in `.local-cicd-logs/`
3. Run with debug mode enabled
4. Check GitHub Actions workflow for comparison
5. Create an issue with detailed logs

---

*The local CI/CD pipeline ensures consistent quality and provides fast feedback during development, making it easier to maintain high code quality and catch issues early.*