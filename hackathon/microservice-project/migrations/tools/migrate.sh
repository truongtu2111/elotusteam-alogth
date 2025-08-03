#!/bin/bash

# Database Migration Tool with Data Ops Controls
# This script provides comprehensive migration management with approval workflows

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/../scripts"
ROLLBACKS_DIR="$SCRIPT_DIR/../rollbacks"
CONFIG_DIR="$SCRIPT_DIR/../config"
DATA_DIR="$SCRIPT_DIR/../data"
LOG_DIR="$SCRIPT_DIR/../logs"

# Create log directory if it doesn't exist
mkdir -p "$LOG_DIR"

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
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_DIR/migration.log"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_DIR/migration.log"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_DIR/migration.log"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_DIR/migration.log"
}

log_debug() {
    if [[ "$DEBUG" == "true" ]]; then
        echo -e "${PURPLE}[DEBUG]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_DIR/migration.log"
    fi
}

# Configuration loading
load_config() {
    local env="$1"
    
    if [[ ! -f "$CONFIG_DIR/environments.yml" ]]; then
        log_error "Configuration file not found: $CONFIG_DIR/environments.yml"
        exit 1
    fi
    
    # Parse YAML configuration (simplified parser)
    eval $(awk -v env="$env" '
        BEGIN { in_env = 0 }
        /^[a-zA-Z_]+:/ { 
            if ($1 == env ":") {
                in_env = 1
            } else {
                in_env = 0
            }
        }
        in_env && /^  [a-zA-Z_]+:/ {
            gsub(/^  /, "")
            gsub(/:/, "")
            key = $1
            value = $2
            for (i = 3; i <= NF; i++) value = value " " $i
            gsub(/"/, "", value)
            print "DB_" toupper(key) "=\"" value "\""
        }
    ' "$CONFIG_DIR/environments.yml")
    
    # Set database connection parameters
    export PGHOST="$DB_HOST"
    export PGPORT="$DB_PORT"
    export PGDATABASE="$DB_DATABASE"
    export PGUSER="$DB_USERNAME"
    
    # Load password from environment variable
    if [[ -n "$DB_PASSWORD_ENV" ]]; then
        export PGPASSWORD="${!DB_PASSWORD_ENV}"
    fi
    
    log_debug "Configuration loaded for environment: $env"
}

# Load approval matrix
load_approval_matrix() {
    if [[ ! -f "$CONFIG_DIR/approval_matrix.yml" ]]; then
        log_error "Approval matrix file not found: $CONFIG_DIR/approval_matrix.yml"
        exit 1
    fi
    
    log_debug "Approval matrix loaded"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL client (psql) is not installed"
        exit 1
    fi
    
    # Check if yq is available for YAML parsing
    if ! command -v yq &> /dev/null; then
        log_warning "yq is not installed. Using simplified YAML parser."
    fi
    
    # Check database connection
    if ! psql -c "SELECT 1;" &> /dev/null; then
        log_error "Cannot connect to database. Check connection parameters."
        log_error "Host: $PGHOST, Port: $PGPORT, Database: $PGDATABASE, User: $PGUSER"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Create migration infrastructure
create_migration_infrastructure() {
    log_info "Creating migration infrastructure..."
    
    # Create migration history table
    psql << 'EOF'
CREATE TABLE IF NOT EXISTS migration_history (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    risk_level VARCHAR(10) DEFAULT 'LOW',
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    applied_by VARCHAR(100) DEFAULT CURRENT_USER,
    approved_by TEXT,
    approval_timestamp TIMESTAMP,
    rolled_back_at TIMESTAMP,
    rolled_back_by VARCHAR(100),
    rollback_reason TEXT,
    checksum VARCHAR(64),
    execution_time_ms INTEGER,
    status VARCHAR(20) DEFAULT 'APPLIED'
);

CREATE TABLE IF NOT EXISTS migration_locks (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(50) NOT NULL,
    locked_by VARCHAR(100) NOT NULL,
    locked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reason TEXT,
    UNIQUE(environment)
);

CREATE TABLE IF NOT EXISTS migration_approvals (
    id SERIAL PRIMARY KEY,
    migration_version VARCHAR(50) NOT NULL,
    approver_role VARCHAR(50) NOT NULL,
    approver_name VARCHAR(100) NOT NULL,
    approved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    comments TEXT,
    UNIQUE(migration_version, approver_role)
);

CREATE INDEX IF NOT EXISTS idx_migration_history_version ON migration_history(version);
CREATE INDEX IF NOT EXISTS idx_migration_history_status ON migration_history(status);
CREATE INDEX IF NOT EXISTS idx_migration_approvals_version ON migration_approvals(migration_version);
EOF
    
    log_success "Migration infrastructure created"
}

# Check migration lock
check_migration_lock() {
    local environment="$1"
    
    local lock_info=$(psql -t -c "SELECT locked_by, locked_at, reason FROM migration_locks WHERE environment = '$environment';")
    
    if [[ -n "$lock_info" ]]; then
        log_error "Migration is locked for environment: $environment"
        log_error "Lock info: $lock_info"
        return 1
    fi
    
    return 0
}

# Acquire migration lock
acquire_migration_lock() {
    local environment="$1"
    local reason="$2"
    local locked_by="${USER:-$(whoami)}"
    
    if ! check_migration_lock "$environment"; then
        return 1
    fi
    
    psql -c "INSERT INTO migration_locks (environment, locked_by, reason) VALUES ('$environment', '$locked_by', '$reason');"
    
    log_info "Migration lock acquired for environment: $environment"
    return 0
}

# Release migration lock
release_migration_lock() {
    local environment="$1"
    
    psql -c "DELETE FROM migration_locks WHERE environment = '$environment';"
    
    log_info "Migration lock released for environment: $environment"
}

# Check approval status
check_approval_status() {
    local version="$1"
    local environment="$2"
    local risk_level="$3"
    
    log_info "Checking approval status for migration: $version"
    
    # Get required approvals from approval matrix
    local required_approvals
    case "$environment" in
        "development")
            required_approvals=1
            ;;
        "staging")
            required_approvals=2
            ;;
        "production")
            case "$risk_level" in
                "LOW")
                    required_approvals=2
                    ;;
                "MEDIUM")
                    required_approvals=3
                    ;;
                "HIGH")
                    required_approvals=4
                    ;;
                *)
                    required_approvals=2
                    ;;
            esac
            ;;
        *)
            required_approvals=1
            ;;
    esac
    
    # Count current approvals
    local current_approvals=$(psql -t -c "SELECT COUNT(*) FROM migration_approvals WHERE migration_version = '$version';")
    current_approvals=$(echo "$current_approvals" | tr -d ' ')
    
    log_info "Required approvals: $required_approvals, Current approvals: $current_approvals"
    
    if [[ "$current_approvals" -ge "$required_approvals" ]]; then
        log_success "Migration $version has sufficient approvals"
        return 0
    else
        log_warning "Migration $version requires $((required_approvals - current_approvals)) more approvals"
        return 1
    fi
}

# Add approval
add_approval() {
    local version="$1"
    local approver_role="$2"
    local approver_name="$3"
    local comments="$4"
    
    psql -c "INSERT INTO migration_approvals (migration_version, approver_role, approver_name, comments) VALUES ('$version', '$approver_role', '$approver_name', '$comments') ON CONFLICT (migration_version, approver_role) DO UPDATE SET approver_name = EXCLUDED.approver_name, approved_at = CURRENT_TIMESTAMP, comments = EXCLUDED.comments;"
    
    log_success "Approval added for migration $version by $approver_name ($approver_role)"
}

# Get applied migrations
get_applied_migrations() {
    psql -t -c "SELECT version FROM migration_history WHERE status = 'APPLIED' ORDER BY version;" | tr -d ' '
}

# Get migration metadata
get_migration_metadata() {
    local migration_file="$1"
    
    # Extract metadata from migration file comments
    local description=$(grep "^-- Description:" "$migration_file" | sed 's/^-- Description: //')
    local risk_level=$(grep "^-- Risk Level:" "$migration_file" | sed 's/^-- Risk Level: //')
    local estimated_duration=$(grep "^-- Estimated Duration:" "$migration_file" | sed 's/^-- Estimated Duration: //')
    
    echo "$description|$risk_level|$estimated_duration"
}

# Validate migration script
validate_migration_script() {
    local migration_file="$1"
    local version=$(basename "$migration_file" .sql)
    
    log_info "Validating migration script: $version"
    
    # Check if file exists
    if [[ ! -f "$migration_file" ]]; then
        log_error "Migration file not found: $migration_file"
        return 1
    fi
    
    # Check if rollback script exists
    local rollback_file="$ROLLBACKS_DIR/${version}_rollback.sql"
    if [[ ! -f "$rollback_file" ]]; then
        log_warning "Rollback script not found: $rollback_file"
    fi
    
    # Validate SQL syntax (dry run)
    if ! psql --set ON_ERROR_STOP=1 --set AUTOCOMMIT=off -f "$migration_file" --dry-run &> /dev/null; then
        log_error "Migration script has syntax errors"
        return 1
    fi
    
    log_success "Migration script validation passed"
    return 0
}

# Apply migration
apply_migration() {
    local migration_file="$1"
    local environment="$2"
    local force="$3"
    
    local version=$(basename "$migration_file" .sql)
    local start_time=$(date +%s%3N)
    
    log_info "Applying migration: $version"
    
    # Get migration metadata
    local metadata=$(get_migration_metadata "$migration_file")
    local description=$(echo "$metadata" | cut -d'|' -f1)
    local risk_level=$(echo "$metadata" | cut -d'|' -f2)
    local estimated_duration=$(echo "$metadata" | cut -d'|' -f3)
    
    # Check if already applied
    local applied=$(psql -t -c "SELECT COUNT(*) FROM migration_history WHERE version = '$version' AND status = 'APPLIED';" | tr -d ' ')
    if [[ "$applied" -gt 0 ]]; then
        log_warning "Migration $version is already applied"
        return 0
    fi
    
    # Check approvals for production
    if [[ "$environment" == "production" && "$force" != "true" ]]; then
        if ! check_approval_status "$version" "$environment" "$risk_level"; then
            log_error "Migration $version does not have sufficient approvals for production deployment"
            return 1
        fi
    fi
    
    # Validate migration script
    if ! validate_migration_script "$migration_file"; then
        return 1
    fi
    
    # Calculate checksum
    local checksum=$(sha256sum "$migration_file" | awk '{print $1}')
    
    # Create backup if required
    if [[ "$DB_BACKUP_REQUIRED" == "true" ]]; then
        log_info "Creating backup before migration..."
        local backup_file="$LOG_DIR/backup_${version}_$(date +%Y%m%d_%H%M%S).sql"
        pg_dump > "$backup_file"
        log_success "Backup created: $backup_file"
    fi
    
    # Apply migration with transaction
    log_info "Executing migration: $version"
    
    if psql --set ON_ERROR_STOP=1 --set AUTOCOMMIT=off << EOF
BEGIN;

-- Execute migration script
\i $migration_file

-- Update migration history
INSERT INTO migration_history (version, description, risk_level, checksum, applied_by) 
VALUES ('$version', '$description', '$risk_level', '$checksum', CURRENT_USER)
ON CONFLICT (version) DO UPDATE SET 
    status = 'APPLIED',
    applied_at = CURRENT_TIMESTAMP,
    applied_by = CURRENT_USER,
    checksum = EXCLUDED.checksum;

COMMIT;
EOF
    then
        local end_time=$(date +%s%3N)
        local execution_time=$((end_time - start_time))
        
        # Update execution time
        psql -c "UPDATE migration_history SET execution_time_ms = $execution_time WHERE version = '$version';"
        
        log_success "Migration $version applied successfully in ${execution_time}ms"
        
        # Send notification
        send_notification "migration_deployed" "Migration $version deployed to $environment" "$version" "$environment"
        
        return 0
    else
        log_error "Migration $version failed"
        
        # Mark as failed
        psql -c "UPDATE migration_history SET status = 'FAILED' WHERE version = '$version';"
        
        # Send failure notification
        send_notification "migration_failed" "Migration $version failed in $environment" "$version" "$environment"
        
        # Auto-rollback if configured
        if [[ "$DB_AUTO_ROLLBACK_ON_FAILURE" == "true" ]]; then
            log_info "Auto-rollback is enabled, attempting rollback..."
            rollback_migration "$version" "$environment" "true"
        fi
        
        return 1
    fi
}

# Rollback migration
rollback_migration() {
    local version="$1"
    local environment="$2"
    local auto_rollback="$3"
    
    log_info "Rolling back migration: $version"
    
    local rollback_file="$ROLLBACKS_DIR/${version}_rollback.sql"
    
    # Check if rollback file exists
    if [[ ! -f "$rollback_file" ]]; then
        log_error "Rollback file not found: $rollback_file"
        return 1
    fi
    
    # Check if migration was applied
    local applied=$(psql -t -c "SELECT COUNT(*) FROM migration_history WHERE version = '$version' AND status = 'APPLIED';" | tr -d ' ')
    if [[ "$applied" -eq 0 ]]; then
        log_error "Migration $version was not applied or already rolled back"
        return 1
    fi
    
    # Check approval for manual rollback
    if [[ "$auto_rollback" != "true" && "$environment" == "production" ]]; then
        log_warning "Manual rollback in production requires approval"
        # In a real implementation, this would check for rollback approvals
    fi
    
    # Apply rollback
    log_info "Executing rollback: $version"
    
    if psql --set ON_ERROR_STOP=1 --set AUTOCOMMIT=off << EOF
BEGIN;

-- Execute rollback script
\i $rollback_file

-- Update migration history
UPDATE migration_history 
SET status = 'ROLLED_BACK',
    rolled_back_at = CURRENT_TIMESTAMP,
    rolled_back_by = CURRENT_USER,
    rollback_reason = CASE WHEN '$auto_rollback' = 'true' THEN 'Auto-rollback due to failure' ELSE 'Manual rollback' END
WHERE version = '$version';

COMMIT;
EOF
    then
        log_success "Migration $version rolled back successfully"
        
        # Send notification
        send_notification "rollback_initiated" "Migration $version rolled back in $environment" "$version" "$environment"
        
        return 0
    else
        log_error "Rollback $version failed"
        return 1
    fi
}

# Send notification
send_notification() {
    local event_type="$1"
    local message="$2"
    local version="$3"
    local environment="$4"
    
    log_debug "Sending notification: $event_type - $message"
    
    # Slack notification (if webhook URL is configured)
    if [[ -n "$SLACK_WEBHOOK_URL" ]]; then
        local color
        case "$event_type" in
            "migration_deployed") color="good" ;;
            "migration_failed") color="danger" ;;
            "rollback_initiated") color="warning" ;;
            *) color="#439FE0" ;;
        esac
        
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"$color\",\"text\":\"$message\"}]}" \
            "$SLACK_WEBHOOK_URL" &> /dev/null || true
    fi
    
    # Email notification (if configured)
    if [[ -n "$EMAIL_RECIPIENTS" ]]; then
        echo "$message" | mail -s "Database Migration: $event_type" "$EMAIL_RECIPIENTS" &> /dev/null || true
    fi
}

# Run migrations
run_migrations() {
    local environment="$1"
    local target_version="$2"
    local force="$3"
    
    log_info "Starting migration process for environment: $environment"
    
    # Acquire migration lock
    if ! acquire_migration_lock "$environment" "Migration process"; then
        exit 1
    fi
    
    # Ensure lock is released on exit
    trap "release_migration_lock '$environment'" EXIT
    
    # Get list of applied migrations
    local applied_migrations=($(get_applied_migrations))
    
    # Get list of available migrations
    local available_migrations=($(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort))
    
    if [[ ${#available_migrations[@]} -eq 0 ]]; then
        log_warning "No migration scripts found in $MIGRATIONS_DIR"
        return 0
    fi
    
    local migrations_applied=0
    
    for migration_file in "${available_migrations[@]}"; do
        local version=$(basename "$migration_file" .sql)
        
        # Skip if already applied
        if [[ " ${applied_migrations[@]} " =~ " ${version} " ]]; then
            log_info "Skipping already applied migration: $version"
            continue
        fi
        
        # Apply migration
        if apply_migration "$migration_file" "$environment" "$force"; then
            migrations_applied=$((migrations_applied + 1))
        else
            log_error "Migration process failed at $version"
            exit 1
        fi
        
        # Stop if target version reached
        if [[ "$version" == "$target_version" ]]; then
            break
        fi
    done
    
    if [[ $migrations_applied -eq 0 ]]; then
        log_info "No new migrations to apply"
    else
        log_success "Migration process completed. Applied $migrations_applied migrations."
    fi
}

# Show migration status
show_status() {
    local environment="$1"
    
    log_info "Migration Status for $environment"
    echo "==========================================="
    
    # Migration summary
    psql -c "
SELECT 
    COUNT(*) as total_migrations,
    COUNT(CASE WHEN status = 'APPLIED' THEN 1 END) as applied_migrations,
    COUNT(CASE WHEN status = 'ROLLED_BACK' THEN 1 END) as rolled_back_migrations,
    COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed_migrations,
    MAX(applied_at) as last_migration_date
FROM migration_history;
"
    
    echo ""
    log_info "Recent Migrations"
    echo "==========================================="
    
    # Recent migrations
    psql -c "
SELECT 
    version,
    description,
    risk_level,
    status,
    applied_at,
    execution_time_ms
FROM migration_history 
ORDER BY applied_at DESC 
LIMIT 10;
"
    
    # Pending migrations
    echo ""
    log_info "Pending Migrations"
    echo "==========================================="
    
    local applied_migrations=($(get_applied_migrations))
    local available_migrations=($(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | sort))
    
    local pending_count=0
    
    for migration_file in "${available_migrations[@]}"; do
        local version=$(basename "$migration_file" .sql)
        
        if [[ ! " ${applied_migrations[@]} " =~ " ${version} " ]]; then
            local metadata=$(get_migration_metadata "$migration_file")
            local description=$(echo "$metadata" | cut -d'|' -f1)
            local risk_level=$(echo "$metadata" | cut -d'|' -f2)
            echo "- $version: $description (Risk: $risk_level)"
            pending_count=$((pending_count + 1))
        fi
    done
    
    if [[ $pending_count -eq 0 ]]; then
        log_success "No pending migrations"
    else
        log_warning "$pending_count pending migrations found"
    fi
}

# Usage information
usage() {
    echo "Database Migration Tool with Data Ops Controls"
    echo ""
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  migrate <env> [version]     Apply migrations to environment"
    echo "  rollback <env> <version>    Rollback specific migration"
    echo "  status <env>                Show migration status"
    echo "  approve <version> <role> <name> [comments]  Add approval"
    echo "  lock <env> [reason]         Acquire migration lock"
    echo "  unlock <env>                Release migration lock"
    echo "  validate <file>             Validate migration script"
    echo ""
    echo "Environments: development, staging, production"
    echo ""
    echo "Options:"
    echo "  --force                     Force migration without approval checks"
    echo "  --debug                     Enable debug logging"
    echo "  --dry-run                   Show what would be done without executing"
    echo ""
    echo "Examples:"
    echo "  $0 migrate development"
    echo "  $0 migrate production 005_add_analytics_tables"
    echo "  $0 rollback staging 004_user_profiles"
    echo "  $0 approve 005_add_analytics_tables data_ops_lead john.doe 'Approved after review'"
    echo "  $0 status production"
}

# Main function
main() {
    local command="$1"
    shift
    
    # Parse global options
    local force="false"
    local dry_run="false"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                force="true"
                shift
                ;;
            --debug)
                export DEBUG="true"
                shift
                ;;
            --dry-run)
                dry_run="true"
                shift
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
    
    case "$command" in
        "migrate")
            local environment="$1"
            local target_version="$2"
            
            if [[ -z "$environment" ]]; then
                log_error "Environment is required"
                usage
                exit 1
            fi
            
            load_config "$environment"
            load_approval_matrix
            check_prerequisites
            create_migration_infrastructure
            
            if [[ "$dry_run" == "true" ]]; then
                log_info "Dry run mode - showing what would be done"
                show_status "$environment"
            else
                run_migrations "$environment" "$target_version" "$force"
            fi
            ;;
            
        "rollback")
            local environment="$1"
            local version="$2"
            
            if [[ -z "$environment" || -z "$version" ]]; then
                log_error "Environment and version are required for rollback"
                usage
                exit 1
            fi
            
            load_config "$environment"
            check_prerequisites
            
            if [[ "$dry_run" == "true" ]]; then
                log_info "Dry run mode - would rollback migration: $version"
            else
                rollback_migration "$version" "$environment" "false"
            fi
            ;;
            
        "status")
            local environment="$1"
            
            if [[ -z "$environment" ]]; then
                log_error "Environment is required"
                usage
                exit 1
            fi
            
            load_config "$environment"
            check_prerequisites
            show_status "$environment"
            ;;
            
        "approve")
            local version="$1"
            local role="$2"
            local name="$3"
            local comments="$4"
            
            if [[ -z "$version" || -z "$role" || -z "$name" ]]; then
                log_error "Version, role, and name are required for approval"
                usage
                exit 1
            fi
            
            # Use development config for approval operations
            load_config "development"
            check_prerequisites
            create_migration_infrastructure
            
            add_approval "$version" "$role" "$name" "$comments"
            ;;
            
        "lock")
            local environment="$1"
            local reason="$2"
            
            if [[ -z "$environment" ]]; then
                log_error "Environment is required"
                usage
                exit 1
            fi
            
            load_config "$environment"
            check_prerequisites
            create_migration_infrastructure
            
            acquire_migration_lock "$environment" "${reason:-Manual lock}"
            ;;
            
        "unlock")
            local environment="$1"
            
            if [[ -z "$environment" ]]; then
                log_error "Environment is required"
                usage
                exit 1
            fi
            
            load_config "$environment"
            check_prerequisites
            
            release_migration_lock "$environment"
            ;;
            
        "validate")
            local migration_file="$1"
            
            if [[ -z "$migration_file" ]]; then
                log_error "Migration file is required"
                usage
                exit 1
            fi
            
            # Use development config for validation
            load_config "development"
            check_prerequisites
            
            validate_migration_script "$migration_file"
            ;;
            
        "")
            log_error "Command is required"
            usage
            exit 1
            ;;
            
        *)
            log_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"