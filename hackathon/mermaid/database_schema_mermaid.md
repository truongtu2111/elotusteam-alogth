# Database Schema - File Upload System (Mermaid)

## Entity Relationship Diagram

```mermaid
erDiagram
    users {
        int id PK
        varchar username UK
        varchar password_hash
        varchar role
        timestamp created_at
        timestamp last_login
        boolean is_active
    }

    files {
        int id PK
        int owner_id FK
        varchar filename
        varchar original_name
        varchar content_type
        int file_size
        varchar file_path
        varchar upload_ip
        text user_agent
        timestamp created_at
        boolean is_deleted
    }

    user_groups {
        int id PK
        varchar group_name
        text description
        int created_by FK
        timestamp created_at
    }

    group_members {
        int id PK
        int group_id FK
        int user_id FK
        timestamp joined_at
    }

    file_permissions {
        int id PK
        int file_id FK
        int user_id FK
        boolean can_read
        boolean can_download
        boolean can_write
        boolean can_share
        timestamp granted_at
    }

    group_file_permissions {
        int id PK
        int file_id FK
        int group_id FK
        boolean can_read
        boolean can_download
        boolean can_write
        boolean can_share
        timestamp granted_at
    }

    image_variants {
        int id PK
        int original_file_id FK
        varchar variant_type
        varchar quality_level
        varchar format
        varchar file_path
        int file_size
        int width
        int height
    }

    activity_logs {
        int id PK
        int user_id FK
        int file_id FK
        varchar action
        varchar ip_address
        timestamp timestamp
    }

    revoked_tokens {
        int id PK
        varchar token_id UK
        int user_id FK
        timestamp revoked_at
    }

    %% Relationships
    users ||--o{ files : owns
    users ||--o{ user_groups : creates
    users ||--o{ group_members : belongs_to
    users ||--o{ file_permissions : has
    users ||--o{ activity_logs : performs
    users ||--o{ revoked_tokens : has
    
    user_groups ||--o{ group_members : contains
    user_groups ||--o{ group_file_permissions : has
    
    files ||--o{ file_permissions : has
    files ||--o{ group_file_permissions : has
    files ||--o{ image_variants : generates
    files ||--o{ activity_logs : tracked_in
```

## Database Architecture Overview

```mermaid
flowchart TD
    subgraph Core["🏗️ Core Tables"]
        Users["👤 users<br/>User Management"]
        Files["📁 files<br/>File Metadata"]
        Groups["👥 user_groups<br/>Group Management"]
        Members["🔗 group_members<br/>Group Membership"]
    end

    subgraph Permissions["🔐 Permission System"]
        FilePerm["🔑 file_permissions<br/>Individual Access"]
        GroupPerm["🔑 group_file_permissions<br/>Group Access"]
    end

    subgraph Support["🛠️ Supporting Tables"]
        Variants["🖼️ image_variants<br/>Image Processing"]
        Logs["📊 activity_logs<br/>Audit Trail"]
        Tokens["🚫 revoked_tokens<br/>Security"]
    end

    subgraph Performance["⚡ Performance Layer"]
        Indexes["📇 Composite Indexes<br/>Query Optimization"]
        Partitions["📂 Table Partitioning<br/>Data Management"]
        Pool["🏊 Connection Pooling<br/>PgBouncer"]
        Cache["🚀 Redis Cache<br/>Query Caching"]
    end

    subgraph Security["🛡️ Security Features"]
        Encryption["🔒 Password Hashing<br/>bcrypt + salt"]
        JWT["🎫 JWT Management<br/>Token Blacklist"]
        Audit["📝 Complete Audit<br/>Activity Tracking"]
        RBAC["👮 Role-Based Access<br/>Granular Permissions"]
    end

    subgraph Scaling["📈 Scalability"]
        Sharding["🔀 Database Sharding<br/>Horizontal Scaling"]
        Replicas["📚 Read Replicas<br/>Load Distribution"]
        Archive["📦 Data Archival<br/>Cold Storage"]
        Backup["💾 Backup Strategy<br/>Cross-Region"]
    end

    %% Connections
    Core --> Permissions
    Core --> Support
    Permissions --> Performance
    Support --> Performance
    Performance --> Security
    Security --> Scaling
```

## Permission Flow Diagram

```mermaid
flowchart LR
    subgraph User["👤 User Access"]
        U1["User Request"]
        U2["Check Permissions"]
        U3["Grant/Deny Access"]
    end

    subgraph Individual["🔑 Individual Permissions"]
        I1["file_permissions"]
        I2["Direct User Access"]
    end

    subgraph Group["👥 Group Permissions"]
        G1["group_members"]
        G2["group_file_permissions"]
        G3["Inherited Access"]
    end

    subgraph Actions["⚡ Permission Types"]
        A1["📖 can_read"]
        A2["⬇️ can_download"]
        A3["✏️ can_write"]
        A4["🔗 can_share"]
    end

    U1 --> U2
    U2 --> Individual
    U2 --> Group
    Individual --> I1 --> I2
    Group --> G1 --> G2 --> G3
    I2 --> Actions
    G3 --> Actions
    Actions --> U3
```

## Data Flow & Operations

```mermaid
sequenceDiagram
    participant User as 👤 User
    participant App as 🖥️ Application
    participant Auth as 🔐 Auth System
    participant DB as 🗄️ Database
    participant Cache as 🚀 Cache
    participant Storage as 💾 File Storage
    participant Audit as 📊 Audit Log

    User->>App: Upload File Request
    App->>Auth: Validate JWT Token
    Auth->>DB: Check revoked_tokens
    DB-->>Auth: Token Valid
    Auth-->>App: User Authenticated
    
    App->>DB: Check user permissions
    DB->>Cache: Query permission cache
    Cache-->>DB: Cache miss
    DB-->>App: Permission granted
    
    App->>Storage: Store file
    Storage-->>App: File stored
    
    App->>DB: Insert file metadata
    App->>DB: Create image variants
    App->>Audit: Log upload activity
    
    App-->>User: Upload successful
    
    Note over Cache: Update permission cache
    Note over Audit: Track all operations
```

## Index Strategy

```mermaid
graph TD
    subgraph Indexes["📇 Database Indexes"]
        A["🔍 Primary Indexes"]
        B["⚡ Composite Indexes"]
        C["📊 Performance Indexes"]
    end

    A --> A1["users.id (PK)"]
    A --> A2["files.id (PK)"]
    A --> A3["user_groups.id (PK)"]
    
    B --> B1["(user_id, file_id, permission)"]
    B --> B2["(original_file_id, variant_type)"]
    B --> B3["(group_id, user_id)"]
    
    C --> C1["files.owner_id"]
    C --> C2["files.created_at"]
    C --> C3["files.content_type"]
    C --> C4["activity_logs.timestamp"]
```

## Monitoring & Maintenance

```mermaid
graph LR
    subgraph Monitor["📊 Monitoring"]
        M1["🐌 Slow Query Log"]
        M2["📈 Index Usage"]
        M3["🔗 Connection Metrics"]
        M4["💾 Storage Usage"]
    end

    subgraph Maintain["🔧 Maintenance"]
        MT1["🧹 VACUUM"]
        MT2["🔄 REINDEX"]
        MT3["📊 ANALYZE"]
        MT4["✂️ Partition Pruning"]
    end

    Monitor --> Maintain
```