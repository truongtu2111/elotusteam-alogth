# Microservice Project

A comprehensive microservice architecture built with Go, featuring authentication, file management, user management, and analytics services.

## Project Structure

```
microservice-project/
├── services/
│   ├── auth/                 # Authentication service
│   │   ├── main.go
│   │   ├── config/
│   │   ├── domain/
│   │   ├── infrastructure/
│   │   └── presentation/
│   ├── file/                 # File management service
│   │   └── main.go
│   ├── user/                 # User management service
│   │   └── main.go
│   └── analytics/            # Analytics service
│       └── main.go
├── shared/                   # Shared libraries and utilities
│   ├── config/
│   ├── data/
│   ├── domain/
│   └── utils/
├── tests/                    # Test files
│   ├── integration/
│   └── unit/
├── docker-compose.yml        # Docker composition
├── nginx.conf               # API Gateway configuration
├── Makefile                 # Build and run commands
└── README.md
```

## Services

### 1. Authentication Service (Port 8081)
- User registration and login
- JWT token management
- Session management
- Password reset functionality

### 2. File Service (Port 8082)
- File upload and download
- File metadata management
- File permissions
- File storage operations

### 3. User Service (Port 8083)
- User profile management
- User CRUD operations
- User groups and permissions

### 4. Analytics Service (Port 8085)
- User activity tracking
- System metrics collection
- Reporting and dashboards

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose (optional)
- PostgreSQL (if running locally)
- Redis (if running locally)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd microservice-project
```

2. Install dependencies:
```bash
make install-deps
```

### Running the Services

#### Option 1: Using Make Commands

```bash
# Run individual services
make run-auth      # Auth service on port 8081
make run-file      # File service on port 8082
make run-user      # User service on port 8083
make run-analytics # Analytics service on port 8085
```

#### Option 2: Using Docker Compose

```bash
# Start all services with dependencies
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

#### Option 3: Manual Go Run

```bash
# Terminal 1 - Auth Service
SERVER_PORT=8081 go run ./services/auth

# Terminal 2 - File Service
SERVER_PORT=8082 go run ./services/file

# Terminal 3 - User Service
SERVER_PORT=8083 go run ./services/user

# Terminal 4 - Analytics Service
SERVER_PORT=8085 go run ./services/analytics
```

### API Gateway

When using Docker Compose, an Nginx API Gateway is available on port 8080:

- Auth endpoints: `http://localhost:8080/api/v1/auth/`
- File endpoints: `http://localhost:8080/api/v1/files/`
- User endpoints: `http://localhost:8080/api/v1/users/`
- Profile endpoints: `http://localhost:8080/api/v1/profile/`
- Analytics endpoints: `http://localhost:8080/api/v1/analytics/`

### Health Checks

Each service provides a health check endpoint:

- Auth: `http://localhost:8081/health`
- File: `http://localhost:8082/health`
- User: `http://localhost:8083/health`
- Analytics: `http://localhost:8085/health`
- API Gateway: `http://localhost:8080/health`

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration
```

## Development

### Building

```bash
# Build all services
make build

# Clean build artifacts
make clean
```

### Code Quality

```bash
# Run linter
make lint
```

## Environment Variables

### Common Variables
- `SERVER_HOST`: Server host (default: localhost)
- `SERVER_PORT`: Server port (varies by service)
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_NAME`: Database name
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password

### Auth Service Specific
- `JWT_SECRET`: JWT signing secret
- `REDIS_HOST`: Redis host
- `REDIS_PORT`: Redis port

## Architecture

This project follows Clean Architecture principles:

- **Domain Layer**: Business logic and entities
- **Infrastructure Layer**: Database, external services
- **Presentation Layer**: HTTP handlers, middleware
- **Shared Layer**: Common utilities and interfaces

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run tests and linting
6. Submit a pull request

## License

This project is licensed under the MIT License.

A comprehensive microservice architecture for file upload system with JWT authentication, implementing SOLID principles and Clean Architecture.

## Architecture Overview

This project implements a scalable microservice architecture with the following key principles:

- **SOLID Principles**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Test-Driven Development (TDD)**: Comprehensive test coverage with integration, security, and performance tests
- **Abstracted Communication**: Pluggable sync/async communication layers
- **Abstracted Data Sources**: Switchable database and storage implementations
- **CI/CD Pipeline**: Automated testing, building, and deployment

## Project Structure

```
microservice-project/
├── services/                    # Microservices
│   ├── auth-service/           # Authentication & Authorization
│   ├── file-service/           # File Management
│   ├── permission-service/     # Access Control
│   ├── user-service/           # User Management
│   ├── image-processing-service/ # Image Processing
│   └── notification-service/   # Notifications
├── shared/                     # Shared libraries
│   ├── communication/          # Sync/Async communication abstractions
│   ├── data/                   # Data source abstractions
│   ├── security/               # Security utilities
│   └── monitoring/             # Observability
├── infrastructure/             # Infrastructure as Code
│   ├── docker/                 # Docker configurations
│   ├── kubernetes/             # K8s manifests
│   └── terraform/              # Cloud infrastructure
├── tests/                      # Test suites
│   ├── integration/            # Integration tests
│   ├── security/               # Security tests
│   ├── performance/            # Load/stress tests
│   └── e2e/                    # End-to-end tests
├── scripts/                    # Automation scripts
├── docs/                       # Documentation
└── .github/workflows/          # CI/CD pipelines
```

## Technology Stack

### Local Development
- **Language**: Go 1.21+
- **Database**: PostgreSQL with Docker
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **Storage**: MinIO (S3-compatible)
- **Monitoring**: Prometheus + Grafana
- **Tracing**: Jaeger

### Production
- **Container Orchestration**: Kubernetes
- **Database**: PostgreSQL (managed service)
- **Cache**: Redis Cluster
- **Message Queue**: Apache Kafka
- **Storage**: AWS S3 / Google Cloud Storage
- **Monitoring**: Prometheus + Grafana + AlertManager
- **Tracing**: Jaeger / AWS X-Ray
- **Service Mesh**: Istio

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- kubectl (for K8s deployment)
- Make

### Local Development Setup

```bash
# Clone the repository
git clone <repository-url>
cd microservice-project

# Start infrastructure services
make infra-up

# Run tests
make test-all

# Start all services
make services-up

# Run integration tests
make test-integration
```

## Testing Strategy

### Test Pyramid
1. **Unit Tests** (70%): Fast, isolated tests for business logic
2. **Integration Tests** (20%): Service integration and API tests
3. **End-to-End Tests** (10%): Full system workflow tests

### Test Types
- **Security Tests**: Authentication, authorization, input validation
- **Performance Tests**: Load testing, stress testing, benchmark tests
- **Contract Tests**: API contract validation between services
- **Chaos Tests**: Resilience and fault tolerance testing

## CI/CD Pipeline

### Continuous Integration
1. Code quality checks (linting, formatting)
2. Security scanning (SAST, dependency check)
3. Unit tests execution
4. Integration tests
5. Build Docker images
6. Push to registry

### Continuous Deployment
1. Deploy to staging environment
2. Run E2E tests
3. Security tests
4. Performance tests
5. Deploy to production (blue-green)
6. Health checks and monitoring

## Performance Targets

- **Authentication**: 10,000 requests/second, <100ms latency
- **File Upload**: 1,000 concurrent uploads, <500ms processing
- **File Download**: 5,000 concurrent downloads, <200ms first byte
- **Permission Checks**: 50,000 requests/second, <10ms latency
- **Availability**: 99.9% uptime
- **Scalability**: Horizontal scaling to handle 1M+ users

## Security Features

- JWT-based authentication with refresh tokens
- Role-based access control (RBAC)
- File permission system with granular controls
- Input validation and sanitization
- Rate limiting and DDoS protection
- Audit logging and monitoring
- Encryption at rest and in transit

## Monitoring & Observability

- **Metrics**: Prometheus with custom business metrics
- **Logging**: Structured logging with correlation IDs
- **Tracing**: Distributed tracing across services
- **Alerting**: Automated alerts for SLA violations
- **Dashboards**: Real-time operational dashboards

## Contributing

1. Follow TDD approach: Write tests first
2. Implement SOLID principles
3. Maintain clean architecture boundaries
4. Add comprehensive documentation
5. Ensure all tests pass before submitting PR

## License

MIT License - see LICENSE file for details