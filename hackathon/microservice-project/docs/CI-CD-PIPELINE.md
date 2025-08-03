# CI/CD Pipeline Documentation

This document describes the comprehensive CI/CD pipeline implemented for the microservice project using GitHub Actions.

## Overview

The CI/CD pipeline consists of multiple workflows that handle different aspects of the development lifecycle:

1. **Main CI/CD Pipeline** (`ci-cd.yml`) - Core testing, building, and deployment
2. **Dependency Management** (`dependency-update.yml`) - Automated dependency updates and security scanning
3. **Release Management** (`release.yml`) - Automated releases and versioning
4. **Monitoring & Health Checks** (`monitoring.yml`) - Continuous monitoring and alerting

## Workflows

### 1. Main CI/CD Pipeline (`ci-cd.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches
- Manual dispatch

**Jobs:**

#### Code Quality & Security
- Runs Go linters (`golangci-lint`)
- Performs security scanning with `nancy` and `gosec`
- Checks code formatting and runs `go vet`

#### Unit Tests
- Executes all unit tests with race detection
- Generates code coverage reports
- Uploads coverage to artifacts

#### Integration Tests
- Sets up PostgreSQL and Redis services
- Runs integration tests against real databases
- Tests inter-service communication

#### Security Tests
- Runs dedicated security test suite
- Performs additional security scans

#### Performance Tests
- Executes performance benchmarks (main branch only)
- Measures response times and throughput

#### Build Docker Images
- Builds Docker images for all services
- Pushes images to container registry
- Uses multi-platform builds (amd64, arm64)

#### End-to-End Tests
- Starts services using Docker Compose
- Runs complete user journey tests
- Tests API endpoints and workflows

#### Chaos Engineering Tests
- Runs chaos tests (main branch only)
- Tests system resilience and recovery

#### Deployment
- **Staging**: Deploys on `develop` branch pushes
- **Production**: Deploys on `main` branch pushes
- Uses environment protection rules

#### Notification
- Sends deployment status notifications
- Alerts team on failures

### 2. Dependency Management (`dependency-update.yml`)

**Triggers:**
- Daily schedule (2 AM UTC)
- Manual dispatch

**Jobs:**

#### Update Go Dependencies
- Updates all Go dependencies
- Runs tests to ensure compatibility
- Creates pull request with changes

#### Security Vulnerability Scan
- Scans for known vulnerabilities using `nancy` and `gosec`
- Generates security reports
- Fails on critical vulnerabilities

#### Docker Security Scan
- Scans Docker images with Trivy
- Uploads SARIF results
- Checks for container vulnerabilities

#### License Compliance Check
- Scans dependencies for license compliance
- Generates license reports
- Ensures legal compliance

### 3. Release Management (`release.yml`)

**Triggers:**
- Git tags matching `v*.*.*` pattern
- Manual dispatch with version input

**Jobs:**

#### Validate Release
- Validates version format
- Determines if it's a pre-release
- Sets release metadata

#### Full Test Suite
- Runs comprehensive test suite
- Ensures release quality

#### Build Release Artifacts
- Builds binaries for multiple platforms
- Creates distribution archives
- Supports Linux, macOS, and Windows

#### Build Docker Images
- Builds and tags release images
- Pushes to container registry
- Creates semantic version tags

#### Generate Changelog
- Automatically generates changelog
- Extracts changes since last release

#### Create GitHub Release
- Creates GitHub release with artifacts
- Includes changelog and installation instructions
- Attaches binary distributions

#### Deploy to Production
- Deploys stable releases to production
- Skips pre-releases

### 4. Monitoring & Health Checks (`monitoring.yml`)

**Triggers:**
- Every 15 minutes (scheduled)
- Manual dispatch

**Jobs:**

#### Service Health Check
- Checks health endpoints for all services
- Monitors response times
- Tests both staging and production

#### API Endpoint Testing
- Tests critical API endpoints
- Validates authentication flows
- Ensures proper error responses

#### Database Connectivity
- Checks database connections
- Validates database health

#### Performance Monitoring
- Runs basic load tests
- Monitors response times
- Detects performance degradation

#### Security Monitoring
- Checks security headers
- Validates SSL/TLS configuration
- Monitors for security issues

#### Generate Monitoring Report
- Creates comprehensive monitoring report
- Uploads results as artifacts

#### Alert on Failures
- Sends alerts when checks fail
- Integrates with notification systems

## Configuration

### Required Secrets

Add these secrets to your GitHub repository:

```bash
# Container Registry
GITHUB_TOKEN  # Automatically provided by GitHub

# Deployment (if using external services)
DOCKER_REGISTRY_USERNAME
DOCKER_REGISTRY_PASSWORD
KUBE_CONFIG  # For Kubernetes deployments

# Notifications
SLACK_WEBHOOK_URL
EMAIL_SMTP_PASSWORD

# External Services (if applicable)
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
GCP_SERVICE_ACCOUNT_KEY
```

### Environment Variables

Configure these in your workflow files:

```yaml
env:
  GO_VERSION: '1.21'
  DOCKER_REGISTRY: ghcr.io
  STAGING_URL: https://staging-api.example.com
  PRODUCTION_URL: https://api.example.com
```

### Branch Protection Rules

Recommended branch protection settings:

**Main Branch:**
- Require pull request reviews
- Require status checks to pass
- Require branches to be up to date
- Include administrators
- Required status checks:
  - `code-quality`
  - `unit-tests`
  - `integration-tests`
  - `security-tests`

**Develop Branch:**
- Require pull request reviews
- Require status checks to pass
- Required status checks:
  - `unit-tests`
  - `integration-tests`

## Usage

### Development Workflow

1. **Feature Development:**
   ```bash
   git checkout -b feature/new-feature
   # Make changes
   git commit -m "feat: add new feature"
   git push origin feature/new-feature
   ```

2. **Create Pull Request:**
   - Open PR to `develop` branch
   - CI pipeline runs automatically
   - Review and merge after checks pass

3. **Release to Staging:**
   ```bash
   git checkout develop
   git merge feature/new-feature
   git push origin develop
   # Automatically deploys to staging
   ```

4. **Release to Production:**
   ```bash
   git checkout main
   git merge develop
   git push origin main
   # Automatically deploys to production
   ```

### Creating Releases

1. **Automatic Release (Recommended):**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   # Release workflow runs automatically
   ```

2. **Manual Release:**
   - Go to Actions tab in GitHub
   - Select "Release Management" workflow
   - Click "Run workflow"
   - Enter version and options

### Monitoring

- Health checks run automatically every 15 minutes
- View monitoring reports in Actions artifacts
- Set up notifications for alerts

### Troubleshooting

#### Common Issues

1. **Test Failures:**
   - Check test logs in Actions
   - Run tests locally: `make test`
   - Fix issues and push again

2. **Build Failures:**
   - Check build logs
   - Verify Dockerfile syntax
   - Test locally: `make build`

3. **Deployment Failures:**
   - Check deployment logs
   - Verify environment configuration
   - Check service health endpoints

4. **Security Scan Failures:**
   - Review security reports
   - Update vulnerable dependencies
   - Fix security issues

#### Debugging

1. **Enable Debug Logging:**
   ```yaml
   env:
     ACTIONS_STEP_DEBUG: true
     ACTIONS_RUNNER_DEBUG: true
   ```

2. **SSH into Runner (for debugging):**
   ```yaml
   - name: Setup tmate session
     uses: mxschmitt/action-tmate@v3
   ```

## Best Practices

### Code Quality
- Write comprehensive tests
- Follow Go coding standards
- Use meaningful commit messages
- Keep pull requests small and focused

### Security
- Regularly update dependencies
- Review security scan results
- Use least privilege principles
- Rotate secrets regularly

### Performance
- Monitor application metrics
- Set up performance budgets
- Optimize Docker images
- Use caching effectively

### Deployment
- Use blue-green deployments
- Implement health checks
- Have rollback procedures
- Monitor post-deployment

## Metrics and Monitoring

### Key Metrics
- Build success rate
- Test coverage percentage
- Deployment frequency
- Lead time for changes
- Mean time to recovery
- Change failure rate

### Dashboards
- GitHub Actions dashboard
- Application performance monitoring
- Infrastructure monitoring
- Security monitoring

## Integration with External Tools

### Supported Integrations
- **Slack**: Notifications and alerts
- **Email**: Deployment notifications
- **PagerDuty**: Critical alerts
- **Datadog/New Relic**: Application monitoring
- **SonarQube**: Code quality analysis
- **Snyk**: Security vulnerability scanning

### Adding New Integrations
1. Add required secrets
2. Update workflow files
3. Test integration
4. Document configuration

## Maintenance

### Regular Tasks
- Review and update dependencies
- Update workflow configurations
- Review security scan results
- Monitor pipeline performance
- Update documentation

### Quarterly Reviews
- Analyze pipeline metrics
- Optimize workflow performance
- Review security practices
- Update tooling versions
- Team retrospectives

## Support

For issues or questions:
1. Check this documentation
2. Review workflow logs
3. Search existing issues
4. Create new issue with details
5. Contact DevOps team

---

*This pipeline is designed to be robust, secure, and maintainable. Regular reviews and updates ensure it continues to meet the team's needs.*