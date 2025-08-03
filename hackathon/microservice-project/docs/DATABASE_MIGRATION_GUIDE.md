# 🗄️ Database Migration Framework

This guide provides a comprehensive framework for managing database migrations with data ops team controls, ensuring safe and controlled database schema changes.

## 🎯 Overview

The migration framework includes:
- **Automated Migration Scripts** - Version-controlled schema changes
- **Data Ops Approval Workflow** - Multi-stage approval process
- **Rollback Mechanisms** - Safe rollback procedures
- **Migration Monitoring** - Real-time migration tracking
- **Environment Management** - Dev/Staging/Production controls

---

## 📋 Migration Workflow

### Phase 1: Development
```
Developer → Create Migration → Local Testing → PR Creation
```

### Phase 2: Review & Approval
```
PR Review → Data Ops Review → Staging Deployment → Production Approval
```

### Phase 3: Deployment
```
Scheduled Deployment → Monitoring → Validation → Completion
```

---

## 🏗️ Migration Framework Structure

### Directory Structure
```
migrations/
├── scripts/
│   ├── 001_initial_schema.sql
│   ├── 002_add_user_profiles.sql
│   ├── 003_analytics_tables.sql
│   └── ...
├── rollbacks/
│   ├── 001_initial_schema_rollback.sql
│   ├── 002_add_user_profiles_rollback.sql
│   └── ...
├── data/
│   ├── seed_data.sql
│   ├── test_data.sql
│   └── production_data.sql
├── config/
│   ├── migration_config.yml
│   ├── environments.yml
│   └── approval_matrix.yml
└── tools/
    ├── migrate.sh
    ├── rollback.sh
    ├── validate.sh
    └── monitor.sh
```

---

## 📝 Migration Script Standards

### 1. Naming Convention
```
{version}_{description}.sql

Examples:
001_initial_schema.sql
002_add_user_profiles.sql
003_create_analytics_tables.sql
004_add_file_metadata_columns.sql
```

### 2. Migration Script Template
```sql
-- Migration: {version}_{description}
-- Author: {author}
-- Date: {date}
-- Description: {detailed_description}
-- Estimated Duration: {duration}
-- Risk Level: {LOW|MEDIUM|HIGH}
-- Rollback Available: {YES|NO}

-- Pre-migration checks
DO $$
BEGIN
    -- Check prerequisites
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'migration_history') THEN
        RAISE EXCEPTION 'Migration history table not found';
    END IF;
END $$;

-- Begin transaction
BEGIN;

-- Migration steps
-- Step 1: Create new tables
CREATE TABLE IF NOT EXISTS new_feature_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Step 2: Add indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_new_feature_name ON new_feature_table(name);

-- Step 3: Insert initial data (if needed)
INSERT INTO new_feature_table (name) VALUES ('default_value') ON CONFLICT DO NOTHING;

-- Update migration history
INSERT INTO migration_history (version, description, applied_at, applied_by) 
VALUES ('{version}', '{description}', CURRENT_TIMESTAMP, CURRENT_USER);

-- Commit transaction
COMMIT;

-- Post-migration validation
DO $$
BEGIN
    -- Validate migration success
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'new_feature_table') THEN
        RAISE EXCEPTION 'Migration validation failed: new_feature_table not created';
    END IF;
END $$;
```

### 3. Rollback Script Template
```sql
-- Rollback: {version}_{description}
-- Author: {author}
-- Date: {date}
-- Description: Rollback for {original_description}
-- Risk Level: {LOW|MEDIUM|HIGH}

-- Pre-rollback checks
DO $$
BEGIN
    -- Check if migration was applied
    IF NOT EXISTS (SELECT 1 FROM migration_history WHERE version = '{version}') THEN
        RAISE EXCEPTION 'Migration {version} was not applied, cannot rollback';
    END IF;
END $$;

-- Begin transaction
BEGIN;

-- Rollback steps (reverse order of migration)
-- Step 1: Remove data
DELETE FROM new_feature_table WHERE name = 'default_value';

-- Step 2: Drop indexes
DROP INDEX IF EXISTS idx_new_feature_name;

-- Step 3: Drop tables
DROP TABLE IF EXISTS new_feature_table;

-- Update migration history
UPDATE migration_history 
SET rolled_back_at = CURRENT_TIMESTAMP, rolled_back_by = CURRENT_USER 
WHERE version = '{version}';

-- Commit transaction
COMMIT;

-- Post-rollback validation
DO $$
BEGIN
    -- Validate rollback success
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'new_feature_table') THEN
        RAISE EXCEPTION 'Rollback validation failed: new_feature_table still exists';
    END IF;
END $$;
```

---

## 🔐 Data Ops Approval Workflow

### 1. Approval Matrix
```yaml
# config/approval_matrix.yml
approval_matrix:
  development:
    required_approvals: 1
    approvers:
      - "developer"
      - "tech_lead"
    
  staging:
    required_approvals: 2
    approvers:
      - "tech_lead"
      - "data_ops_lead"
    auto_deploy: true
    
  production:
    required_approvals: 3
    approvers:
      - "tech_lead"
      - "data_ops_lead"
      - "engineering_manager"
    auto_deploy: false
    maintenance_window_required: true

risk_levels:
  LOW:
    required_approvals: 2
    can_auto_deploy: true
    
  MEDIUM:
    required_approvals: 3
    can_auto_deploy: false
    requires_staging_validation: true
    
  HIGH:
    required_approvals: 4
    can_auto_deploy: false
    requires_staging_validation: true
    requires_maintenance_window: true
    requires_rollback_plan: true
```

### 2. Approval Process

#### Step 1: Migration Request
```yaml
# .github/PULL_REQUEST_TEMPLATE/migration.md
---
name: Database Migration Request
about: Request approval for database migration
title: '[MIGRATION] {version}_{description}'
labels: ['migration', 'data-ops-review']
assignees: ['@data-ops-team']
---

## Migration Details
- **Version**: {version}
- **Description**: {description}
- **Risk Level**: {LOW|MEDIUM|HIGH}
- **Estimated Duration**: {duration}
- **Rollback Available**: {YES|NO}

## Impact Assessment
- **Tables Affected**: 
- **Data Loss Risk**: {YES|NO}
- **Downtime Required**: {YES|NO}
- **Performance Impact**: {LOW|MEDIUM|HIGH}

## Testing
- [ ] Local testing completed
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Performance testing completed

## Rollback Plan
- **Rollback Script**: {path_to_rollback_script}
- **Rollback Duration**: {duration}
- **Data Recovery Plan**: {description}

## Approvals Required
- [ ] Tech Lead (@tech-lead)
- [ ] Data Ops Lead (@data-ops-lead)
- [ ] Engineering Manager (@eng-manager) [HIGH risk only]

## Deployment Schedule
- **Preferred Date**: {date}
- **Preferred Time**: {time}
- **Maintenance Window**: {YES|NO}
```

#### Step 2: Automated Checks
```yaml
# .github/workflows/migration-validation.yml
name: Migration Validation

on:
  pull_request:
    paths:
      - 'migrations/**'

jobs:
  validate-migration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Validate Migration Format
        run: |
          ./migrations/tools/validate.sh
          
      - name: Check Rollback Script
        run: |
          ./migrations/tools/check_rollback.sh
          
      - name: Test Migration on Staging DB
        run: |
          ./migrations/tools/test_migration.sh staging
          
      - name: Performance Impact Analysis
        run: |
          ./migrations/tools/performance_check.sh
          
      - name: Security Scan
        run: |
          ./migrations/tools/security_scan.sh
```

---

## 🛠️ Migration Tools

### 1. Migration Execution Tool
```bash
#!/bin/bash
# migrations/tools/migrate.sh

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/../scripts"
CONFIG_DIR="$SCRIPT_DIR/../config"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
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

# Load configuration
load_config() {
    local env="$1"
    if [[ -f "$CONFIG_DIR/environments.yml" ]]; then
        # Parse YAML config (simplified)
        DB_HOST=$(grep -A 10 "$env:" "$CONFIG_DIR/environments.yml" | grep "host:" | awk '{print $2}')
        DB_PORT=$(grep -A 10 "$env:" "$CONFIG_DIR/environments.yml" | grep "port:" | awk '{print $2}')
        DB_NAME=$(grep -A 10 "$env:" "$CONFIG_DIR/environments.yml" | grep "database:" | awk '{print $2}')
        DB_USER=$(grep -A 10 "$env:" "$CONFIG_DIR/environments.yml" | grep "username:" | awk '{print $2}')
    else
        log_error "Configuration file not found: $CONFIG_DIR/environments.yml"
        exit 1
    fi
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        log_error "psql is not installed"
        exit 1
    fi
    
    # Check database connection
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" &> /dev/null; then
        log_error "Cannot connect to database"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Create migration history table
create_migration_table() {
    log_info "Creating migration history table..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
CREATE TABLE IF NOT EXISTS migration_history (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    applied_by VARCHAR(100) DEFAULT CURRENT_USER,
    rolled_back_at TIMESTAMP,
    rolled_back_by VARCHAR(100),
    checksum VARCHAR(64)
);
EOF
    
    log_success "Migration history table ready"
}

# Get applied migrations
get_applied_migrations() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT version FROM migration_history WHERE rolled_back_at IS NULL ORDER BY version;"
}

# Apply migration
apply_migration() {
    local migration_file="$1"
    local version=$(basename "$migration_file" .sql)
    
    log_info "Applying migration: $version"
    
    # Calculate checksum
    local checksum=$(sha256sum "$migration_file" | awk '{print $1}')
    
    # Apply migration
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"; then
        # Update checksum in migration history
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
            "UPDATE migration_history SET checksum = '$checksum' WHERE version = '$version';"
        
        log_success "Migration $version applied successfully"
        return 0
    else
        log_error "Migration $version failed"
        return 1
    fi
}

# Main migration function
run_migrations() {
    local target_version="$1"
    
    log_info "Starting migration process..."
    
    # Get list of applied migrations
    local applied_migrations=($(get_applied_migrations))
    
    # Get list of available migrations
    local available_migrations=($(ls "$MIGRATIONS_DIR"/*.sql | sort))
    
    for migration_file in "${available_migrations[@]}"; do
        local version=$(basename "$migration_file" .sql)
        
        # Skip if already applied
        if [[ " ${applied_migrations[@]} " =~ " ${version} " ]]; then
            log_info "Skipping already applied migration: $version"
            continue
        fi
        
        # Apply migration
        if ! apply_migration "$migration_file"; then
            log_error "Migration process failed at $version"
            exit 1
        fi
        
        # Stop if target version reached
        if [[ "$version" == "$target_version" ]]; then
            break
        fi
    done
    
    log_success "Migration process completed"
}

# Usage
usage() {
    echo "Usage: $0 <environment> [target_version]"
    echo "Environments: development, staging, production"
    echo "Example: $0 production 005_add_analytics_tables"
}

# Main
main() {
    if [[ $# -lt 1 ]]; then
        usage
        exit 1
    fi
    
    local environment="$1"
    local target_version="$2"
    
    # Load configuration
    load_config "$environment"
    
    # Check prerequisites
    check_prerequisites
    
    # Create migration table
    create_migration_table
    
    # Run migrations
    run_migrations "$target_version"
}

# Execute main function
main "$@"
```

### 2. Rollback Tool
```bash
#!/bin/bash
# migrations/tools/rollback.sh

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROLLBACKS_DIR="$SCRIPT_DIR/../rollbacks"
CONFIG_DIR="$SCRIPT_DIR/../config"

# Source common functions
source "$SCRIPT_DIR/migrate.sh"

# Rollback migration
rollback_migration() {
    local version="$1"
    local rollback_file="$ROLLBACKS_DIR/${version}_rollback.sql"
    
    log_info "Rolling back migration: $version"
    
    # Check if rollback file exists
    if [[ ! -f "$rollback_file" ]]; then
        log_error "Rollback file not found: $rollback_file"
        exit 1
    fi
    
    # Check if migration was applied
    local applied=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT COUNT(*) FROM migration_history WHERE version = '$version' AND rolled_back_at IS NULL;")
    
    if [[ "$applied" -eq 0 ]]; then
        log_error "Migration $version was not applied or already rolled back"
        exit 1
    fi
    
    # Apply rollback
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$rollback_file"; then
        log_success "Migration $version rolled back successfully"
        return 0
    else
        log_error "Rollback $version failed"
        return 1
    fi
}

# Main rollback function
run_rollback() {
    local version="$1"
    
    if [[ -z "$version" ]]; then
        log_error "Version is required for rollback"
        usage
        exit 1
    fi
    
    log_info "Starting rollback process for version: $version"
    
    # Confirm rollback
    read -p "Are you sure you want to rollback migration $version? (yes/no): " confirm
    if [[ "$confirm" != "yes" ]]; then
        log_info "Rollback cancelled"
        exit 0
    fi
    
    # Perform rollback
    rollback_migration "$version"
    
    log_success "Rollback process completed"
}

# Usage
usage() {
    echo "Usage: $0 <environment> <version>"
    echo "Environments: development, staging, production"
    echo "Example: $0 production 005_add_analytics_tables"
}

# Main
main() {
    if [[ $# -lt 2 ]]; then
        usage
        exit 1
    fi
    
    local environment="$1"
    local version="$2"
    
    # Load configuration
    load_config "$environment"
    
    # Check prerequisites
    check_prerequisites
    
    # Run rollback
    run_rollback "$version"
}

# Execute main function
main "$@"
```

---

## 📊 Migration Monitoring

### 1. Migration Status Dashboard
```sql
-- Create migration monitoring views
CREATE OR REPLACE VIEW migration_status AS
SELECT 
    version,
    description,
    applied_at,
    applied_by,
    CASE 
        WHEN rolled_back_at IS NOT NULL THEN 'ROLLED_BACK'
        ELSE 'APPLIED'
    END as status,
    rolled_back_at,
    rolled_back_by
FROM migration_history
ORDER BY version;

CREATE OR REPLACE VIEW migration_summary AS
SELECT 
    COUNT(*) as total_migrations,
    COUNT(CASE WHEN rolled_back_at IS NULL THEN 1 END) as applied_migrations,
    COUNT(CASE WHEN rolled_back_at IS NOT NULL THEN 1 END) as rolled_back_migrations,
    MAX(applied_at) as last_migration_date
FROM migration_history;
```

### 2. Migration Monitoring Script
```bash
#!/bin/bash
# migrations/tools/monitor.sh

# Monitor migration status
monitor_migrations() {
    local environment="$1"
    
    log_info "Migration Status for $environment"
    echo "==========================================="
    
    # Load configuration
    load_config "$environment"
    
    # Get migration summary
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
        "SELECT * FROM migration_summary;"
    
    echo ""
    log_info "Recent Migrations"
    echo "==========================================="
    
    # Get recent migrations
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
        "SELECT version, description, status, applied_at FROM migration_status ORDER BY applied_at DESC LIMIT 10;"
}

# Check for pending migrations
check_pending_migrations() {
    local environment="$1"
    
    # Load configuration
    load_config "$environment"
    
    # Get applied migrations
    local applied_migrations=($(get_applied_migrations))
    
    # Get available migrations
    local available_migrations=($(ls "$MIGRATIONS_DIR"/*.sql | sort))
    
    local pending_count=0
    
    log_info "Pending Migrations for $environment"
    echo "==========================================="
    
    for migration_file in "${available_migrations[@]}"; do
        local version=$(basename "$migration_file" .sql)
        
        if [[ ! " ${applied_migrations[@]} " =~ " ${version} " ]]; then
            echo "- $version"
            pending_count=$((pending_count + 1))
        fi
    done
    
    if [[ $pending_count -eq 0 ]]; then
        log_success "No pending migrations"
    else
        log_warning "$pending_count pending migrations found"
    fi
}

# Main
main() {
    local environment="$1"
    local action="$2"
    
    if [[ -z "$environment" ]]; then
        echo "Usage: $0 <environment> [status|pending]"
        exit 1
    fi
    
    case "$action" in
        "pending")
            check_pending_migrations "$environment"
            ;;
        "status"|"")
            monitor_migrations "$environment"
            ;;
        *)
            echo "Unknown action: $action"
            echo "Available actions: status, pending"
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"
```

---

## 🔒 Security & Compliance

### 1. Security Checklist
```yaml
# Security validation checklist
security_checks:
  - name: "SQL Injection Prevention"
    description: "Check for parameterized queries and input validation"
    automated: true
    
  - name: "Privilege Escalation"
    description: "Ensure migrations don't grant excessive privileges"
    automated: true
    
  - name: "Data Exposure"
    description: "Check for potential data leaks in migration scripts"
    automated: false
    
  - name: "Backup Verification"
    description: "Ensure backups are available before migration"
    automated: true
    
  - name: "Rollback Testing"
    description: "Verify rollback scripts work correctly"
    automated: true
```

### 2. Compliance Requirements
```yaml
# Compliance requirements
compliance:
  audit_trail:
    required: true
    retention_period: "7 years"
    fields:
      - migration_version
      - applied_by
      - applied_at
      - approval_chain
      - rollback_history
      
  change_management:
    required_approvals:
      - technical_review
      - data_ops_approval
      - security_review
      - business_approval
      
  data_protection:
    pii_handling:
      - encryption_at_rest
      - encryption_in_transit
      - access_logging
      - data_masking
```

---

## 📈 Best Practices

### 1. Migration Design Principles
- **Backward Compatibility**: Ensure migrations don't break existing functionality
- **Idempotency**: Migrations should be safe to run multiple times
- **Atomicity**: Use transactions to ensure all-or-nothing execution
- **Performance**: Consider impact on production systems
- **Rollback Safety**: Always provide rollback scripts

### 2. Testing Strategy
```yaml
testing_levels:
  unit_tests:
    - migration_script_syntax
    - rollback_script_syntax
    - data_validation
    
  integration_tests:
    - full_migration_cycle
    - rollback_verification
    - performance_impact
    
  staging_tests:
    - production_data_subset
    - load_testing
    - monitoring_validation
    
  production_validation:
    - smoke_tests
    - data_integrity_checks
    - performance_monitoring
```

### 3. Emergency Procedures
```yaml
emergency_procedures:
  migration_failure:
    immediate_actions:
      - stop_migration_process
      - assess_data_integrity
      - notify_stakeholders
      - initiate_rollback_if_safe
      
  data_corruption:
    immediate_actions:
      - isolate_affected_systems
      - restore_from_backup
      - investigate_root_cause
      - implement_data_recovery
      
  performance_degradation:
    immediate_actions:
      - monitor_system_metrics
      - identify_bottlenecks
      - apply_performance_fixes
      - consider_rollback_if_severe
```

---

## 🎯 Implementation Checklist

### Phase 1: Setup (Week 1)
- [ ] Create migration directory structure
- [ ] Implement migration tools (migrate.sh, rollback.sh)
- [ ] Set up approval workflow
- [ ] Create migration templates
- [ ] Configure environments

### Phase 2: Integration (Week 2)
- [ ] Integrate with CI/CD pipeline
- [ ] Set up automated testing
- [ ] Configure monitoring and alerting
- [ ] Train data ops team
- [ ] Document procedures

### Phase 3: Production (Week 3)
- [ ] Deploy to staging environment
- [ ] Conduct end-to-end testing
- [ ] Perform security review
- [ ] Get final approvals
- [ ] Deploy to production

### Phase 4: Optimization (Ongoing)
- [ ] Monitor migration performance
- [ ] Gather feedback from teams
- [ ] Optimize approval workflows
- [ ] Enhance automation
- [ ] Update documentation

---

**📞 Support**: For migration issues or questions, contact the Data Ops team or refer to the troubleshooting section in this guide.