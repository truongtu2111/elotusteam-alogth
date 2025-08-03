# File Upload Server with JWT Authentication

A Go-based HTTP server that provides JWT authentication and secure file upload functionality.

## Features

### 1. JWT Authentication System
- User registration with username/password
- User login with JWT token generation (HS256 algorithm)
- Token revocation functionality
- Secure password hashing using bcrypt

### 2. File Upload API
- Authenticated file upload endpoint
- Image file validation (content-type checking)
- File size limit enforcement (8MB maximum)
- Temporary file storage in `/tmp` directory
- Comprehensive metadata storage in database

## System Design Analysis

### Functional Requirements

#### Authentication System
- **User Registration**: Accept username and password, validate uniqueness, hash password securely
- **User Login**: Validate credentials, generate JWT token with 24-hour expiry
- **Token Revocation**: Allow users to invalidate their tokens before expiry
- **Token Validation**: Verify JWT signature and expiry for protected endpoints

#### File Upload System
- **Authorization Check**: Validate JWT token before accepting uploads
- **File Validation**: Ensure uploaded files are images and under 8MB
- **File Storage**: Save files to temporary location with unique names
- **Metadata Persistence**: Store file information and HTTP context in database

#### File Access Control System
- **File Ownership**: Each uploaded file has a designated owner (the uploader)
- **Permission Management**: Support for granular file permissions (read, write, share, delete)
- **User-to-File Permissions**: Allow file owners to grant specific permissions to other users
- **Permission Types**:
  - **Read Permission**: Users can view file metadata and see files in listings (frontend/app display)
  - **Download Permission**: Users can download and access the actual file content
  - **Write Permission**: Users can modify/replace the file content
  - **Share Permission**: Users can grant permissions to other users
  - **Delete Permission**: Users can remove the file from the system
- **Permission Inheritance**: Owners automatically have all permissions on their files
- **Permission Revocation**: Owners can revoke previously granted permissions
- **Bulk Permission Management**: Support for applying permissions to multiple files
- **Permission Auditing**: Track permission changes and access history

#### User Management System
- **User Roles**: Support for different user roles (admin, regular user, guest)
- **User Groups**: Allow creation of user groups for easier permission management
- **Group Permissions**: Apply permissions to entire groups rather than individual users
- **User Profile Management**: Users can manage their profile and view their files
- **User Activity Tracking**: Monitor user actions for security and auditing purposes

### Non-Functional Requirements

#### Security
- **Password Security**: Bcrypt hashing with default cost (10 rounds)
- **JWT Security**: HS256 signing with cryptographically secure random secret
- **Input Validation**: Comprehensive validation of all user inputs
- **Authorization**: Bearer token authentication for protected endpoints
- **File Type Validation**: Content-type verification to prevent malicious uploads
- **Access Control Security**:
  - Principle of least privilege enforcement
  - Permission validation on every file operation
  - Secure permission inheritance and delegation
  - Protection against privilege escalation attacks
  - Audit trail for all permission changes
  - Rate limiting for permission-related operations

#### Performance
- **Memory Efficiency**: Streaming file uploads to avoid loading entire files in memory
- **Database Efficiency**: SQLite for lightweight, embedded database operations
- **Token Management**: In-memory revocation store with automatic cleanup
- **Access Control Performance**:
  - Cache frequently accessed permission data
  - Efficient bulk permission operations
  - Optimized permission lookup algorithms

#### High Throughput & Low Latency Architecture

#### Performance Targets
- **Login Operations**: 10,000 requests/second with <100ms latency
- **File Downloads**: 5,000 concurrent downloads with <200ms first byte
- **File Uploads**: 1,000 concurrent uploads with <500ms processing time
- **Registration**: 1,000 requests/second with <150ms response time
- **Permission Checks**: 50,000 requests/second with <10ms latency

#### Hyperscale Architecture (1 Billion Active Users/Second)

For extreme scale requirements of 1 billion active users per second, please refer to the comprehensive [Hyperscale Architecture Document](billion_readme.md) which details:

- Global edge computing networks with 1000+ locations
- Massive horizontal scaling with 100,000+ instances
- Distributed databases with 10,000+ shards
- Multi-level caching with 1TB+ RAM per cluster
- Exabyte-scale storage with 11-nines durability
- AI-powered operations and monitoring

#### Million-Scale Architecture (Production Ready)

For practical production deployments supporting up to 1 million concurrent users, see the detailed [Million-Scale Architecture Guide](million_readme.md) covering:

- Cost-effective infrastructure design
- Proven technology stack
- Gradual scaling strategies
- Operational best practices
- Implementation timeline

#### Capacity Planning

**Current Active Users Per Second (Target Load)**:
- **Peak Login Rate**: 500 users/second
- **Peak Download Rate**: 2,000 files/second
- **Peak Upload Rate**: 200 files/second
- **Peak Registration Rate**: 100 users/second
- **Permission Operations**: 5,000 checks/second
- **Concurrent Active Sessions**: 100,000 users

#### Scalability Architecture

**Database Layer**:
- **Primary Database**: PostgreSQL with read replicas (3-5 replicas)
- **Connection Pooling**: PgBouncer with 1000 max connections
- **Caching Layer**: Redis Cluster for session data and permissions
- **Database Sharding**: Horizontal partitioning by user_id for files table
- **Query Optimization**: Materialized views for complex permission queries

**Application Layer**:
- **Load Balancing**: HAProxy/Nginx with least-connection algorithm
- **Horizontal Scaling**: Auto-scaling groups (5-50 instances)
- **Service Mesh**: Istio for inter-service communication
- **Circuit Breakers**: Hystrix pattern for fault tolerance
- **Async Processing**: Message queues (RabbitMQ/Apache Kafka) for file processing

**Storage Layer**:
- **File Storage**: AWS S3/Google Cloud Storage with CDN (CloudFront/CloudFlare)
- **Multi-Region**: Cross-region replication for disaster recovery
- **Storage Classes**: Intelligent tiering for cost optimization
- **Caching**: Redis for frequently accessed file metadata

**Memory Management**:
- **Application Memory**: 8GB per instance with garbage collection tuning
- **Cache Memory**: 64GB Redis cluster for hot data
- **Buffer Pools**: Optimized database buffer pools (75% of available RAM)

#### Performance Optimizations

**Database Optimizations**:
- **Indexing Strategy**: Composite indexes on (user_id, file_id, permission_type)
- **Query Optimization**: Prepared statements and query plan caching
- **Connection Management**: Connection pooling with health checks
- **Partitioning**: Time-based partitioning for activity logs

**Application Optimizations**:
- **JWT Caching**: In-memory cache for decoded JWT claims
- **Permission Caching**: Multi-level cache (L1: in-memory, L2: Redis)
- **Batch Operations**: Bulk permission checks and updates
- **Streaming**: Chunked file uploads/downloads for large files

**Network Optimization**:
- **HTTP/2**: Enable HTTP/2 for multiplexing
- **Compression**: Gzip/Brotli compression for API responses
- **Keep-Alive**: Connection reuse for reduced latency
- **CDN**: Global content delivery network for static assets

**Adaptive Image Optimization**:
- **Multi-Resolution Generation**: Create multiple sizes (thumbnail, small, medium, large, original)
- **Quality Variants**: Generate different compression levels (low, medium, high quality)
- **Format Optimization**: WebP, AVIF for modern browsers, JPEG fallback
- **Progressive Loading**: Base64 thumbnails, progressive JPEG enhancement
- **Smart Delivery**: Network-aware quality selection based on connection speed
- **Lazy Loading**: Load images on-demand with intersection observer
- **Responsive Images**: Serve appropriate size based on device and viewport

#### Reliability
- **Error Handling**: Comprehensive error responses with appropriate HTTP status codes
- **Data Integrity**: Foreign key constraints and transaction safety
- **Graceful Degradation**: Proper cleanup of resources and temporary files
- **Access Control Reliability**:
  - Atomic permission operations
  - Consistent permission state across concurrent operations
  - Rollback capabilities for failed permission changes
  - Data integrity constraints for permission relationships

#### Maintainability
- **Code Structure**: Clear separation of concerns with dedicated handler functions
- **Documentation**: Comprehensive comments and API documentation
- **Configuration**: Environment variable support for deployment flexibility

### Security Considerations

#### Access Control Implementation
- **Permission Validation**: Every file operation must validate user permissions before execution
- **Owner Privileges**: File owners automatically have all permissions and cannot be revoked
- **Permission Hierarchy**: Admin users can override permissions for system administration
- **Secure Defaults**: New files are private by default (only owner has access)
- **Permission Auditing**: All permission changes are logged with timestamp and actor

#### Data Protection
- **Encryption at Rest**: Consider encrypting sensitive files on disk
- **Secure File Paths**: Use UUIDs or hashed paths to prevent path traversal
- **Temporary File Cleanup**: Automatically clean up temporary files and orphaned data
- **Backup Security**: Ensure backups maintain permission structures

#### API Security
- **Rate Limiting**: Implement rate limits on permission-sensitive endpoints
- **Input Validation**: Strict validation of permission types and user identifiers
- **CORS Configuration**: Proper CORS settings for web client access
- **Session Management**: Secure JWT token handling and expiration

### DDoS & DoS Protection Strategy

#### Network Layer Protection
- **CDN Protection**: CloudFlare/AWS Shield for L3/L4 DDoS mitigation
- **Geographic Filtering**: Block traffic from high-risk countries
- **IP Reputation**: Real-time IP blacklisting from threat intelligence feeds
- **Traffic Shaping**: Bandwidth limiting and traffic prioritization
- **Anycast Network**: Distributed traffic absorption across multiple data centers

#### Application Layer Protection

**Rate Limiting Strategy**:
- **Global Rate Limits**: 1000 requests/minute per IP address
- **Endpoint-Specific Limits**:
  - Login: 10 attempts/minute per IP, 5 attempts/minute per username
  - Registration: 5 registrations/hour per IP
  - File Upload: 100 uploads/hour per user, 10MB/minute per IP
  - File Download: 1000 downloads/hour per user
  - Permission Operations: 500 requests/minute per user

**Advanced Rate Limiting**:
- **Sliding Window**: Time-based rate limiting with burst allowance
- **Token Bucket**: Allow burst traffic within limits
- **Adaptive Limits**: Dynamic rate adjustment based on system load
- **User-Based Limits**: Different limits for authenticated vs anonymous users

**Request Validation & Filtering**:
- **Request Size Limits**: Maximum 10MB per request
- **Header Validation**: Strict HTTP header validation
- **User-Agent Filtering**: Block known bot signatures
- **Referrer Validation**: Validate request origins
- **Content-Type Enforcement**: Strict MIME type validation

#### Behavioral Analysis & Anomaly Detection

**Traffic Pattern Analysis**:
- **Machine Learning Models**: Detect abnormal traffic patterns
- **Baseline Establishment**: Normal traffic pattern learning
- **Anomaly Scoring**: Real-time threat scoring algorithms
- **Adaptive Thresholds**: Dynamic adjustment based on traffic patterns

**Bot Detection**:
- **CAPTCHA Integration**: Challenge-response for suspicious requests
- **Browser Fingerprinting**: Device and browser characteristic analysis
- **JavaScript Challenges**: Client-side computation requirements
- **Behavioral Biometrics**: Mouse movement and typing pattern analysis

**Attack Pattern Recognition**:
- **Signature-Based Detection**: Known attack pattern matching
- **Heuristic Analysis**: Suspicious behavior pattern detection
- **Correlation Analysis**: Multi-vector attack detection
- **Threat Intelligence**: Integration with external threat feeds

#### Infrastructure Hardening

**Load Balancer Configuration**:
- **Health Checks**: Automatic unhealthy instance removal
- **Circuit Breakers**: Prevent cascade failures
- **Failover Mechanisms**: Automatic traffic rerouting
- **Resource Isolation**: Separate critical and non-critical services

**Database Protection**:
- **Connection Limits**: Maximum database connections per service
- **Query Timeouts**: Prevent long-running query attacks
- **Resource Monitoring**: CPU, memory, and I/O usage tracking
- **Backup Systems**: Hot standby for critical data

**Monitoring & Alerting**:
- **Real-Time Dashboards**: Traffic and system health monitoring
- **Automated Alerts**: Threshold-based notification system
- **Incident Response**: Automated mitigation trigger points
- **Forensic Logging**: Detailed attack pattern logging

#### Emergency Response Procedures

**Incident Response Plan**:
1. **Detection**: Automated threat detection and alerting
2. **Assessment**: Rapid threat severity evaluation
3. **Containment**: Immediate traffic filtering and blocking
4. **Mitigation**: Service scaling and traffic rerouting
5. **Recovery**: Service restoration and performance validation
6. **Post-Incident**: Attack analysis and system hardening

**Automated Mitigation**:
- **Auto-Scaling**: Horizontal scaling during traffic spikes
- **Traffic Shedding**: Non-essential request dropping
- **Graceful Degradation**: Reduced functionality maintenance
- **Emergency Mode**: Critical services only operation

**Manual Override Capabilities**:
- **IP Blocking**: Immediate IP/subnet blocking
- **Service Isolation**: Critical service protection
- **Traffic Redirection**: Alternative endpoint routing
- **System Shutdown**: Emergency service termination

### Implementation Notes

#### Permission Resolution Algorithm
1. Check if user is the file owner (grant all permissions: read, download, write, share, delete)
2. Check direct user permissions in `file_permissions` table
3. Check group permissions via `group_members` and `group_file_permissions`
4. Apply most permissive permissions from all sources
5. Cache resolved permissions for performance

#### Permission Usage Guidelines
- **Read Permission**: Required for file listing, metadata viewing, and frontend display
- **Download Permission**: Required for actual file content access and download operations
- **Write Permission**: Required for file content modification and replacement
- **Share Permission**: Required for granting permissions to other users or groups
- **Delete Permission**: Required for file removal operations
- **Multiple Permissions**: Users can have multiple permissions simultaneously (e.g., read + download for view-only access)

#### Database Optimization
- Use database triggers to maintain permission consistency
- Implement soft deletes for audit trail preservation
- Regular cleanup of expired activity logs
- Optimize queries with proper indexing strategy

#### Adaptive Image Optimization Implementation

**Image Processing Pipeline**:
1. **Upload Processing**: Original image validation and metadata extraction
2. **Background Processing**: Asynchronous generation of multiple variants
3. **Storage Strategy**: Organized file structure with variant naming convention
4. **Database Schema**: Track all variants with size, quality, and format metadata
5. **CDN Integration**: Automatic upload of all variants to CDN

**Image Variant Generation**:
- **Sizes**: 150x150 (thumbnail), 400x400 (small), 800x800 (medium), 1200x1200 (large), original
- **Quality Levels**: 60% (low), 80% (medium), 95% (high) JPEG compression
- **Formats**: WebP (primary), AVIF (next-gen), JPEG (fallback)
- **Progressive Enhancement**: Base64 micro-thumbnails for instant loading

**Smart Delivery Algorithm**:
1. **Network Detection**: Client-side connection speed estimation
2. **Device Capability**: Screen resolution and pixel density detection
3. **Browser Support**: Format capability detection (WebP, AVIF support)
4. **Adaptive Selection**: Automatic quality/size selection based on conditions
5. **Fallback Strategy**: Graceful degradation for unsupported formats

**Database Schema Extensions**:
```sql
CREATE TABLE image_variants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_file_id INTEGER NOT NULL,
    variant_type VARCHAR(20) NOT NULL, -- 'thumbnail', 'small', 'medium', 'large', 'original'
    quality_level VARCHAR(10) NOT NULL, -- 'low', 'medium', 'high'
    format VARCHAR(10) NOT NULL, -- 'jpeg', 'webp', 'avif'
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (original_file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX idx_image_variants_lookup ON image_variants(original_file_id, variant_type, quality_level, format);
```

**API Enhancements**:
- **GET /files/{id}/variants**: List all available variants
- **GET /files/{id}/optimal**: Get optimal variant based on client capabilities
- **GET /files/{id}/thumbnail**: Quick thumbnail access
- **Query Parameters**: ?quality=low|medium|high, ?format=webp|jpeg|avif, ?size=thumbnail|small|medium|large

#### Error Handling
- Distinguish between "file not found" and "access denied" for security
- Provide meaningful error messages without exposing system internals
- Log security violations for monitoring and analysis
- Implement graceful degradation for permission system failures
- Handle image processing failures with fallback to original

## API Endpoints

### Authentication Endpoints

#### POST /register
Register a new user account.

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (201 Created):**
```json
{
  "message": "User registered successfully"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid JSON or missing fields
- `409 Conflict`: Username already exists
- `500 Internal Server Error`: Database or hashing error

#### POST /login
Authenticate user and receive JWT token.

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (200 OK):**
```json
{
  "token": "jwt_token_string",
  "user": {
    "id": 1,
    "username": "string"
  }
}
```

**Error Responses:**
- `400 Bad Request`: Invalid JSON
- `401 Unauthorized`: Invalid credentials
- `500 Internal Server Error`: Database error

#### POST /revoke
Revoke the current JWT token.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "message": "Token revoked successfully"
}
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid token
- `405 Method Not Allowed`: Non-POST request

### File Upload Endpoint

#### POST /upload
Upload an image file (requires authentication).

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**Form Data:**
- `data`: Image file (max 8MB)

**Response (200 OK):**
```json
{
  "message": "File uploaded successfully",
  "file_id": 1,
  "filename": "image.jpg",
  "content_type": "image/jpeg",
  "size": 1024576,
  "temp_path": "/tmp/upload_xyz.tmp"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid file, size limit exceeded, or non-image file
- `401 Unauthorized`: Missing or invalid token
- `405 Method Not Allowed`: Non-POST request
- `500 Internal Server Error`: File system or database error

### File Access Control Endpoints

#### GET /files
List files accessible to the authenticated user.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `owned`: boolean - Filter to show only owned files
- `shared`: boolean - Filter to show only shared files
- `permission`: string - Filter by specific permission (read, download, write, share, delete)

**Response (200 OK):**
```json
{
  "files": [
    {
      "id": 1,
      "filename": "image.jpg",
      "owner_id": 1,
      "owner_username": "user1",
      "permissions": ["read", "download", "write"],
      "uploaded_at": "2024-01-01T12:00:00Z",
      "size": 1024576
    }
  ]
}
```

#### POST /files/{file_id}/permissions
Grant permissions to a user for a specific file.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "username": "target_user",
  "permissions": ["read", "download", "write"]
}
```

**Response (200 OK):**
```json
{
  "message": "Permissions granted successfully"
}
```

**Error Responses:**
- `403 Forbidden`: User doesn't own the file or lack share permission
- `404 Not Found`: File or target user not found
- `400 Bad Request`: Invalid permissions specified

#### DELETE /files/{file_id}/permissions/{user_id}
Revoke permissions from a user for a specific file.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "message": "Permissions revoked successfully"
}
```

#### GET /files/{file_id}/permissions
List all users with permissions on a specific file.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "permissions": [
    {
      "user_id": 2,
      "username": "user2",
      "permissions": ["read", "download"],
      "granted_at": "2024-01-01T12:00:00Z",
      "granted_by": "owner_user"
    }
  ]
}
```

#### GET /files/{file_id}/download
Download a file (requires download permission).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
File content with appropriate Content-Type and Content-Disposition headers.

**Error Responses:**
- `403 Forbidden`: User lacks download permission
- `404 Not Found`: File not found

#### PUT /files/{file_id}
Replace file content (requires write permission).

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**Form Data:**
- `data`: New file content

**Response (200 OK):**
```json
{
  "message": "File updated successfully",
  "version": 2
}
```

#### DELETE /files/{file_id}
Delete a file (requires delete permission or ownership).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "message": "File deleted successfully"
}
```

### User Management Endpoints

#### GET /users/profile
Get current user's profile and file statistics.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "user": {
    "id": 1,
    "username": "user1",
    "role": "regular",
    "created_at": "2024-01-01T10:00:00Z"
  },
  "statistics": {
    "files_owned": 5,
    "files_shared_with_me": 3,
    "total_storage_used": 10485760
  }
}
```

#### GET /users/activity
Get user's activity log.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `limit`: number - Maximum number of activities to return (default: 50)
- `offset`: number - Number of activities to skip
- `action`: string - Filter by action type (upload, download, share, etc.)

**Response (200 OK):**
```json
{
  "activities": [
    {
      "id": 1,
      "action": "file_upload",
      "file_id": 1,
      "filename": "image.jpg",
      "timestamp": "2024-01-01T12:00:00Z",
      "ip_address": "192.168.1.1"
    }
  ]
}
```

### Testing Interface

#### GET /
Provides a simple HTML interface for testing all endpoints.

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'regular' CHECK (role IN ('admin', 'regular', 'guest')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### File Metadata Table
```sql
CREATE TABLE file_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    uploaded_by INTEGER NOT NULL,
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    temp_path TEXT NOT NULL,
    user_agent TEXT,
    remote_addr TEXT,
    version INTEGER DEFAULT 1,
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE CASCADE
);
```

### File Permissions Table
```sql
CREATE TABLE file_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    permission_type TEXT NOT NULL CHECK (permission_type IN ('read', 'download', 'write', 'share', 'delete')),
    granted_by INTEGER NOT NULL,
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES file_metadata(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id),
    UNIQUE(file_id, user_id, permission_type)
);
```

### User Groups Table
```sql
CREATE TABLE user_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id)
);
```

### Group Members Table
```sql
CREATE TABLE group_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    added_by INTEGER NOT NULL,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (added_by) REFERENCES users(id),
    UNIQUE(group_id, user_id)
);
```

### Group File Permissions Table
```sql
CREATE TABLE group_file_permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    permission_type TEXT NOT NULL CHECK (permission_type IN ('read', 'download', 'write', 'share', 'delete')),
    granted_by INTEGER NOT NULL,
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES file_metadata(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id),
    UNIQUE(file_id, group_id, permission_type)
);
```

### User Activity Log Table
```sql
CREATE TABLE user_activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    file_id INTEGER,
    target_user_id INTEGER,
    details TEXT,
    ip_address TEXT,
    user_agent TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES file_metadata(id) ON DELETE SET NULL,
    FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL
);
```

### Database Indexes
```sql
-- Performance indexes for common queries
CREATE INDEX idx_file_metadata_uploaded_by ON file_metadata(uploaded_by);
CREATE INDEX idx_file_permissions_file_id ON file_permissions(file_id);
CREATE INDEX idx_file_permissions_user_id ON file_permissions(user_id);
CREATE INDEX idx_group_file_permissions_file_id ON group_file_permissions(file_id);
CREATE INDEX idx_group_file_permissions_group_id ON group_file_permissions(group_id);
CREATE INDEX idx_user_activities_user_id ON user_activities(user_id);
CREATE INDEX idx_user_activities_timestamp ON user_activities(timestamp);
CREATE INDEX idx_group_members_group_id ON group_members(group_id);
CREATE INDEX idx_group_members_user_id ON group_members(user_id);
```

## Building and Running

### Prerequisites
- Go 1.21 or later
- SQLite3 (for database)

### Installation

1. Clone the repository:
```bash
git clone <repository_url>
cd elotusteam-alogth/hackathon
```

2. Download dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o server main.go
```

4. Run the server:
```bash
./server
```

Or run directly with Go:
```bash
go run main.go
```

### Configuration

The server supports the following environment variables:

- `PORT`: Server port (default: 8080)

Example:
```bash
PORT=3000 go run main.go
```

### Testing

1. Open your browser and navigate to `http://localhost:8080`
2. Use the web interface to:
   - Register a new user
   - Login to get a JWT token
   - Upload an image file using the token

Or use curl commands:

```bash
# Register a user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}'

# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass"}'

# Upload a file (replace TOKEN with actual JWT)
curl -X POST http://localhost:8080/upload \
  -H "Authorization: Bearer TOKEN" \
  -F "data=@image.jpg"

# Revoke token
curl -X POST http://localhost:8080/revoke \
  -H "Authorization: Bearer TOKEN"
```

## Architecture Decisions

### Technology Choices

1. **Go Standard Library**: Minimized external dependencies for better security and maintainability
2. **SQLite**: Embedded database for simplicity and portability
3. **JWT with HS256**: Industry-standard authentication with symmetric key signing
4. **Bcrypt**: Proven password hashing algorithm with adaptive cost

### Security Considerations

1. **Password Storage**: Never store plaintext passwords; use bcrypt hashing
2. **JWT Secret**: Generated using cryptographically secure random bytes
3. **Token Expiry**: 24-hour token lifetime to limit exposure window
4. **Input Validation**: Comprehensive validation of all user inputs
5. **File Type Validation**: Multiple layers of content-type checking
6. **Size Limits**: Prevent DoS attacks through large file uploads

### Scalability Considerations

1. **Stateless Design**: Server maintains no session state (except revoked tokens)
2. **Database Abstraction**: Easy to migrate from SQLite to production databases
3. **File Storage**: Temporary storage design allows easy migration to cloud storage
4. **Horizontal Scaling**: Multiple server instances can share database and storage

### Comprehensive Testing Strategy

#### Performance Testing Plan

**Load Testing Scenarios**:
1. **Baseline Load Test**:
   - 1,000 concurrent users
   - 30-minute duration
   - Mixed workload (70% reads, 20% uploads, 10% auth)
   - Target: <200ms average response time

2. **Peak Load Test**:
   - 10,000 concurrent users
   - 60-minute duration
   - Realistic traffic patterns with spikes
   - Target: <500ms 95th percentile response time

3. **Stress Testing**:
   - Gradual load increase until system failure
   - Identify breaking points and bottlenecks
   - Validate graceful degradation
   - Recovery time measurement

4. **Endurance Testing**:
   - 5,000 concurrent users
   - 24-hour duration
   - Memory leak detection
   - Performance degradation monitoring

**Performance Test Tools**:
- **JMeter**: HTTP load testing with complex scenarios
- **Artillery**: Modern load testing with JavaScript
- **k6**: Developer-friendly performance testing
- **Gatling**: High-performance load testing

**Performance Metrics**:
- **Response Time**: Average, median, 95th, 99th percentiles
- **Throughput**: Requests per second, transactions per second
- **Error Rate**: HTTP errors, application errors, timeouts
- **Resource Utilization**: CPU, memory, disk I/O, network
- **Database Performance**: Query execution time, connection pool usage

#### Security Testing Plan

**Vulnerability Assessment**:
1. **OWASP Top 10 Testing**:
   - SQL Injection attempts
   - Cross-Site Scripting (XSS)
   - Cross-Site Request Forgery (CSRF)
   - Security Misconfiguration
   - Sensitive Data Exposure
   - Broken Authentication
   - Insecure Direct Object References
   - Security Headers validation

2. **Authentication & Authorization Testing**:
   - JWT token manipulation
   - Session fixation attacks
   - Privilege escalation attempts
   - Password policy enforcement
   - Account lockout mechanisms
   - Token expiration validation

3. **Input Validation Testing**:
   - Malformed JSON payloads
   - SQL injection in all parameters
   - File upload security (malicious files)
   - Path traversal attempts
   - Buffer overflow testing
   - Unicode and encoding attacks

4. **DDoS Simulation Testing**:
   - Layer 3/4 flood attacks
   - Application layer attacks
   - Slowloris attacks
   - HTTP flood testing
   - Rate limiting validation
   - Circuit breaker testing

**Security Test Tools**:
- **OWASP ZAP**: Automated security scanning
- **Burp Suite**: Manual penetration testing
- **Nmap**: Network discovery and security auditing
- **SQLMap**: SQL injection testing
- **Nikto**: Web server scanner
- **HULK**: HTTP Unbearable Load King for DDoS simulation

#### Functional Testing Plan

**API Testing**:
1. **Positive Test Cases**:
   - Valid authentication flows
   - Successful file operations
   - Permission management workflows
   - User management operations

2. **Negative Test Cases**:
   - Invalid credentials
   - Unauthorized access attempts
   - Malformed requests
   - Resource not found scenarios

3. **Edge Cases**:
   - Maximum file size uploads
   - Concurrent permission modifications
   - Token expiration edge cases
   - Database connection failures

**Integration Testing**:
- Database integration testing
- File storage integration
- Cache layer integration
- External service dependencies

#### Reliability Testing Plan

**Chaos Engineering**:
1. **Infrastructure Failures**:
   - Database server failures
   - Application server crashes
   - Network partitions
   - Storage system failures

2. **Resource Exhaustion**:
   - Memory exhaustion
   - CPU saturation
   - Disk space depletion
   - Network bandwidth saturation

3. **Dependency Failures**:
   - Cache system failures
   - External API timeouts
   - DNS resolution failures
   - SSL certificate expiration

**Disaster Recovery Testing**:
- Backup and restore procedures
- Failover mechanisms
- Data consistency validation
- Recovery time objectives (RTO)
- Recovery point objectives (RPO)

#### Monitoring & Observability Testing

**Metrics Validation**:
- Application performance metrics
- Business metrics accuracy
- Alert threshold validation
- Dashboard functionality

**Logging Testing**:
- Log format consistency
- Sensitive data redaction
- Log aggregation and search
- Audit trail completeness

**Tracing Testing**:
- Distributed tracing accuracy
- Performance bottleneck identification
- Error propagation tracking
- Service dependency mapping

#### Test Automation Strategy

**Continuous Integration Testing**:
- Unit test execution (>90% coverage)
- Integration test automation
- Security scan automation
- Performance regression testing

**Test Environment Management**:
- Production-like test environments
- Data anonymization for testing
- Environment provisioning automation
- Test data management

**Test Reporting**:
- Automated test result reporting
- Performance trend analysis
- Security vulnerability tracking
- Test coverage metrics

### Production Readiness Improvements

1. **Database**: Migrate to PostgreSQL or MySQL for production
2. **File Storage**: Use cloud storage (AWS S3, Google Cloud Storage)
3. **Token Revocation**: Use Redis for distributed token blacklist
4. **Logging**: Add structured logging with correlation IDs
5. **Monitoring**: Add health checks and metrics endpoints
6. **Rate Limiting**: Implement request rate limiting
7. **HTTPS**: Add TLS termination
8. **Configuration**: Use configuration files or environment-based config

## Potential Interview Questions & Answers

### System Design Questions

**Q: How would you scale this system to handle millions of users with high throughput requirements?**

A: Comprehensive scaling strategy:
1. **Database Layer**: PostgreSQL with 3-5 read replicas, PgBouncer connection pooling (1000 max connections), horizontal sharding by user_id
2. **Application Layer**: Auto-scaling groups (5-50 instances), HAProxy load balancing with least-connection algorithm, service mesh (Istio)
3. **Storage Layer**: AWS S3/Google Cloud Storage with CDN (CloudFront), multi-region replication, intelligent tiering
4. **Caching**: Redis Cluster for session data and permissions, multi-level cache (L1: in-memory, L2: Redis)
5. **Performance Targets**: 10,000 login requests/second, 5,000 concurrent downloads, 50,000 permission checks/second
6. **Capacity Planning**: Support 100,000 concurrent active sessions with 500 peak logins/second

**Q: How would you architect the system to handle 1 billion active users per second (hyperscale)?**

A: For billion-scale architecture, please refer to our comprehensive [Hyperscale Architecture Document](billion_readme.md) which covers:
1. **Global Edge Computing**: 1000+ edge locations with 95% request processing at edge
2. **Massive Horizontal Scaling**: 1,000,000+ containerized instances with AI-powered auto-scaling
3. **Distributed Databases**: Apache Cassandra with 100,000+ shards and cross-region replication
4. **Exabyte Storage**: 100+ EB distributed globally with 11-nines durability
5. **AI-Powered Operations**: Machine learning for predictive scaling and automated operations
6. **Network Infrastructure**: 1000+ Tbps bandwidth with quantum-resistant security

For practical million-scale implementations, see our [Million-Scale Architecture Guide](million_readme.md) with proven, cost-effective solutions.

**Q: How do you handle DDoS attacks and ensure system availability?**

A: Multi-layer DDoS protection strategy:
1. **Network Layer**: CloudFlare/AWS Shield for L3/L4 mitigation, geographic filtering, IP reputation blacklisting
2. **Application Layer**: Advanced rate limiting (10 login attempts/minute per IP, 100 uploads/hour per user), sliding window with burst allowance
3. **Behavioral Analysis**: Machine learning for traffic pattern analysis, bot detection with CAPTCHA, attack pattern recognition
4. **Infrastructure Hardening**: Circuit breakers, health checks, automated failover, resource isolation
5. **Emergency Response**: Automated incident response plan, traffic shedding, graceful degradation, manual override capabilities

**Q: How do you handle token revocation in a distributed system?**

A: Enhanced token management for distributed systems:
1. **Redis Blacklist**: Centralized token blacklist with TTL across all instances
2. **Database Approach**: Store revoked tokens with cleanup jobs and proper indexing
3. **Short-lived Tokens**: Refresh token pattern with 15-minute access tokens
4. **JWT Claims**: Version-based revocation with incremental claim updates
5. **Performance**: Cache decoded JWT claims in memory, batch revocation operations

**Q: Explain your comprehensive access control system design.**

A: Multi-layered permission system:
1. **Permission Types**: Read (metadata/listing), Download (file content), Write (modify), Share (grant permissions), Delete (remove)
2. **Permission Sources**: Direct user permissions, group-based permissions, owner privileges
3. **Resolution Algorithm**: Owner check → direct permissions → group permissions → most permissive wins
4. **Performance**: Multi-level caching, materialized views for complex queries, bulk operations
5. **Security**: Principle of least privilege, audit trail, atomic operations, rollback capabilities

### Technical Questions

**Q: Why did you choose HS256 over RS256 for JWT signing?**

A: HS256 (HMAC) vs RS256 (RSA) trade-offs:
- **HS256 Pros**: Simpler implementation, faster signing/verification, single secret
- **HS256 Cons**: Shared secret, harder to distribute verification
- **RS256 Pros**: Public key verification, better for microservices
- **RS256 Cons**: More complex, slower, key management overhead

For this single-service application, HS256 is appropriate. For microservices, RS256 would be better.

**Q: How do you ensure uploaded files are actually images?**

A: Multiple validation layers:
1. **Content-Type Header**: Check HTTP Content-Type header
2. **File Signature**: Use `http.DetectContentType()` to read file magic bytes
3. **File Extension**: Validate file extension (additional check)
4. **Image Processing**: Attempt to decode image headers (most robust)
5. **Malware Scanning**: Implement virus/malware detection for production
6. **Size Validation**: Enforce strict file size limits (8MB maximum)

**Q: What happens if the server crashes during file upload?**

A: Current behavior and improvements:
- **Current**: Temporary files remain in `/tmp`, database transaction may be incomplete
- **Improvements**: 
  1. Use database transactions for atomicity
  2. Implement cleanup job for orphaned temp files
  3. Add upload resumption capability
  4. Use cloud storage with atomic operations
  5. Implement circuit breakers for graceful degradation

**Q: How do you implement comprehensive security testing?**

A: Multi-layered security testing approach:
1. **OWASP Top 10 Testing**: SQL injection, XSS, CSRF, security misconfiguration validation
2. **Authentication Testing**: JWT manipulation, session fixation, privilege escalation attempts
3. **Input Validation**: Malformed payloads, path traversal, buffer overflow testing
4. **DDoS Simulation**: Layer 3/4 floods, application layer attacks, slowloris testing
5. **Tools**: OWASP ZAP, Burp Suite, Nmap, SQLMap, Nikto for comprehensive coverage
6. **Automated Scanning**: CI/CD integration with security scan automation

### Performance Questions

**Q: How would you optimize file upload performance for high throughput?**

A: Comprehensive performance optimization:
1. **Streaming**: Already implemented - don't load entire file in memory
2. **Parallel Processing**: Process metadata storage while streaming file
3. **CDN Integration**: CloudFront/CloudFlare for global file delivery
4. **Chunked Upload**: Implement resumable uploads for large files (>100MB)
5. **Background Processing**: Async image processing (thumbnails, compression)
6. **Performance Targets**: 1,000 concurrent uploads with <500ms processing time
7. **Storage Optimization**: Intelligent tiering and compression
8. **Network Optimization**: HTTP/2, connection keep-alive, Gzip/Brotli compression

**Q: How do you optimize image delivery for poor network conditions while maintaining quality?**

A: Adaptive image optimization strategy:
1. **Multi-Resolution Generation**: Create 5 size variants (150x150 thumbnail to original) during upload processing
2. **Quality Variants**: Generate 3 compression levels (60%, 80%, 95%) for each size
3. **Format Optimization**: WebP primary, AVIF next-gen, JPEG fallback with browser detection
4. **Smart Delivery Algorithm**: Client-side network speed detection + device capability assessment
5. **Progressive Enhancement**: Base64 micro-thumbnails for instant loading, progressive JPEG enhancement
6. **API Design**: /files/{id}/optimal endpoint with automatic variant selection, query parameters for manual control
7. **Database Schema**: image_variants table tracking all variants with metadata
8. **Fallback Strategy**: Graceful degradation to original file if processing fails

**Q: How do you handle database connection pooling and optimization?**

A: Production-ready database optimization:
1. **Connection Pool**: PgBouncer with 1000 max connections, health checks
2. **Read Replicas**: 3-5 PostgreSQL replicas for read scaling
3. **Query Optimization**: Prepared statements, query plan caching, composite indexes
4. **Sharding Strategy**: Horizontal partitioning by user_id for files table
5. **Caching**: Redis Cluster for frequently accessed data
6. **Monitoring**: Connection pool metrics, query performance tracking
7. **Buffer Management**: 75% of available RAM for database buffers

**Q: How do you implement comprehensive performance testing?**

A: Multi-faceted performance testing strategy:
1. **Load Testing**: 1,000-10,000 concurrent users with JMeter, Artillery, k6
2. **Stress Testing**: Gradual load increase until system failure points
3. **Endurance Testing**: 24-hour duration with 5,000 concurrent users
4. **Performance Metrics**: Response time percentiles, throughput, error rates, resource utilization
5. **Bottleneck Identification**: Database query analysis, memory leak detection
6. **Capacity Planning**: Support 100,000 concurrent sessions with defined SLAs
7. **Automated Testing**: CI/CD integration with performance regression detection

**Q: What are your specific performance targets and how do you achieve them?**

A: Detailed performance targets and implementation:
1. **Login Operations**: 10,000 requests/second with <100ms latency via JWT caching
2. **File Downloads**: 5,000 concurrent downloads with <200ms first byte via CDN
3. **Permission Checks**: 50,000 requests/second with <10ms latency via multi-level caching
4. **Database Performance**: Materialized views for complex permission queries
5. **Memory Management**: 8GB per instance with optimized garbage collection
6. **Auto-scaling**: Horizontal scaling (5-50 instances) based on load metrics

## Dependencies

- `github.com/golang-jwt/jwt/v5`: JWT token generation and validation
- `golang.org/x/crypto/bcrypt`: Password hashing
- `github.com/mattn/go-sqlite3`: SQLite database driver

All dependencies are well-maintained, widely-used libraries with good security track records.