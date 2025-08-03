#!/bin/bash

# Migration Template Generator
# This script helps developers create standardized migration files with proper metadata

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/../scripts"
ROLLBACKS_DIR="$SCRIPT_DIR/../rollbacks"
DATA_DIR="$SCRIPT_DIR/../data"
TEMPLATES_DIR="$SCRIPT_DIR/../templates"

# Create directories if they don't exist
mkdir -p "$MIGRATIONS_DIR" "$ROLLBACKS_DIR" "$DATA_DIR" "$TEMPLATES_DIR"

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

# Generate migration version
generate_version() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local sequence=$(printf "%03d" $(ls "$MIGRATIONS_DIR"/*.sql 2>/dev/null | wc -l | tr -d ' '))
    echo "${sequence}_${timestamp}"
}

# Get user input with validation
get_input() {
    local prompt="$1"
    local default="$2"
    local validation="$3"
    local value
    
    while true; do
        if [[ -n "$default" ]]; then
            read -p "$prompt [$default]: " value
            value=${value:-$default}
        else
            read -p "$prompt: " value
        fi
        
        if [[ -z "$validation" ]] || eval "$validation"; then
            echo "$value"
            break
        else
            log_error "Invalid input. Please try again."
        fi
    done
}

# Validate risk level
validate_risk_level() {
    [[ "$value" =~ ^(LOW|MEDIUM|HIGH)$ ]]
}

# Validate duration format
validate_duration() {
    [[ "$value" =~ ^[0-9]+[smh]$ ]]
}

# Create migration template
create_migration_template() {
    local version="$1"
    local description="$2"
    local risk_level="$3"
    local estimated_duration="$4"
    local migration_type="$5"
    local author="$6"
    local feature_branch="$7"
    local jira_ticket="$8"
    
    local migration_file="$MIGRATIONS_DIR/${version}_$(echo "$description" | tr ' ' '_' | tr '[:upper:]' '[:lower:]').sql"
    local rollback_file="$ROLLBACKS_DIR/${version}_$(echo "$description" | tr ' ' '_' | tr '[:upper:]' '[:lower:]')_rollback.sql"
    
    # Create migration file
    cat > "$migration_file" << EOF
-- Migration: $version
-- Description: $description
-- Author: $author
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- Risk Level: $risk_level
-- Estimated Duration: $estimated_duration
-- Migration Type: $migration_type
-- Feature Branch: $feature_branch
-- JIRA Ticket: $jira_ticket
--
-- IMPORTANT: This migration should be reviewed and approved before deployment
-- Review checklist:
-- [ ] SQL syntax is correct
-- [ ] Indexes are properly created for performance
-- [ ] Foreign key constraints are properly defined
-- [ ] Data migration preserves data integrity
-- [ ] Rollback script is tested and verified
-- [ ] Performance impact has been assessed
-- [ ] Backup strategy is in place

-- ============================================================================
-- MIGRATION START
-- ============================================================================

-- Enable timing for performance monitoring
\timing on

-- Set statement timeout for safety (adjust as needed)
SET statement_timeout = '30min';

-- Begin transaction
BEGIN;

-- Add your migration SQL here
-- Example:
-- CREATE TABLE example_table (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- CREATE INDEX idx_example_table_name ON example_table(name);

-- Insert sample data if needed
-- INSERT INTO example_table (name) VALUES ('Sample Data');

-- Verify the changes
-- SELECT COUNT(*) FROM example_table;

-- Commit transaction
COMMIT;

-- ============================================================================
-- MIGRATION END
-- ============================================================================

-- Post-migration verification queries
-- Add queries here to verify the migration was successful
-- Example:
-- SELECT 
--     schemaname,
--     tablename,
--     attname,
--     typename,
--     attnotnull
-- FROM pg_stats 
-- JOIN pg_attribute ON pg_stats.attname = pg_attribute.attname
-- JOIN pg_type ON pg_attribute.atttypid = pg_type.oid
-- WHERE schemaname = 'public' AND tablename = 'example_table';

EOF

    # Create rollback file
    cat > "$rollback_file" << EOF
-- Rollback Migration: $version
-- Description: Rollback for $description
-- Author: $author
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- Risk Level: $risk_level
-- JIRA Ticket: $jira_ticket
--
-- IMPORTANT: This rollback script should be tested before the migration is applied
-- Rollback checklist:
-- [ ] All created objects are properly dropped
-- [ ] Data is preserved or properly migrated back
-- [ ] Foreign key constraints are handled correctly
-- [ ] Indexes are properly removed
-- [ ] No orphaned data remains

-- ============================================================================
-- ROLLBACK START
-- ============================================================================

-- Enable timing for performance monitoring
\timing on

-- Set statement timeout for safety
SET statement_timeout = '30min';

-- Begin transaction
BEGIN;

-- Add your rollback SQL here (reverse order of migration)
-- Example:
-- DROP INDEX IF EXISTS idx_example_table_name;
-- DROP TABLE IF EXISTS example_table;

-- Verify the rollback
-- SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'example_table';

-- Commit transaction
COMMIT;

-- ============================================================================
-- ROLLBACK END
-- ============================================================================

-- Post-rollback verification queries
-- Add queries here to verify the rollback was successful
-- Example:
-- SELECT COUNT(*) FROM information_schema.tables 
-- WHERE table_schema = 'public' AND table_name = 'example_table';
-- -- Should return 0 if table was properly dropped

EOF

    log_success "Migration files created:"
    log_info "  Migration: $migration_file"
    log_info "  Rollback:  $rollback_file"
    
    return 0
}

# Create data migration template
create_data_migration_template() {
    local version="$1"
    local description="$2"
    local risk_level="$3"
    local estimated_duration="$4"
    local author="$5"
    local feature_branch="$6"
    local jira_ticket="$7"
    
    local migration_file="$MIGRATIONS_DIR/${version}_$(echo "$description" | tr ' ' '_' | tr '[:upper:]' '[:lower:]').sql"
    local rollback_file="$ROLLBACKS_DIR/${version}_$(echo "$description" | tr ' ' '_' | tr '[:upper:]' '[:lower:]')_rollback.sql"
    local data_file="$DATA_DIR/${version}_data.sql"
    
    # Create migration file for data migration
    cat > "$migration_file" << EOF
-- Data Migration: $version
-- Description: $description
-- Author: $author
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- Risk Level: $risk_level
-- Estimated Duration: $estimated_duration
-- Migration Type: DATA_MIGRATION
-- Feature Branch: $feature_branch
-- JIRA Ticket: $jira_ticket
--
-- IMPORTANT: Data migrations require special attention
-- Data migration checklist:
-- [ ] Backup strategy is in place
-- [ ] Data transformation logic is tested
-- [ ] Performance impact is assessed
-- [ ] Rollback data is preserved
-- [ ] Data integrity constraints are maintained
-- [ ] Large dataset handling is optimized

-- ============================================================================
-- DATA MIGRATION START
-- ============================================================================

-- Enable timing and progress reporting
\timing on
\set ECHO all

-- Set work_mem for large operations
SET work_mem = '256MB';
SET maintenance_work_mem = '1GB';

-- Disable autovacuum during migration
SET autovacuum = off;

-- Begin transaction
BEGIN;

-- Create backup table for rollback
CREATE TABLE IF NOT EXISTS backup_${version}_original_data AS 
SELECT * FROM target_table WHERE 1=0; -- Replace 'target_table' with actual table

-- Backup original data
INSERT INTO backup_${version}_original_data 
SELECT * FROM target_table; -- Add WHERE clause if needed

-- Perform data transformation
-- Example batch processing for large datasets:
-- DO $$
-- DECLARE
--     batch_size INTEGER := 10000;
--     total_rows INTEGER;
--     processed_rows INTEGER := 0;
-- BEGIN
--     SELECT COUNT(*) INTO total_rows FROM target_table;
--     
--     WHILE processed_rows < total_rows LOOP
--         UPDATE target_table 
--         SET column_name = new_value
--         WHERE id IN (
--             SELECT id FROM target_table 
--             WHERE condition
--             ORDER BY id 
--             LIMIT batch_size OFFSET processed_rows
--         );
--         
--         processed_rows := processed_rows + batch_size;
--         
--         -- Log progress
--         RAISE NOTICE 'Processed % of % rows (%.2f%%)', 
--             processed_rows, total_rows, 
--             (processed_rows::float / total_rows * 100);
--         
--         -- Commit batch and start new transaction
--         COMMIT;
--         BEGIN;
--     END LOOP;
-- END $$;

-- Verify data transformation
-- SELECT COUNT(*) as total_rows,
--        COUNT(CASE WHEN new_column IS NOT NULL THEN 1 END) as migrated_rows
-- FROM target_table;

-- Re-enable autovacuum
SET autovacuum = on;

-- Analyze tables for updated statistics
ANALYZE target_table;

-- Commit transaction
COMMIT;

-- ============================================================================
-- DATA MIGRATION END
-- ============================================================================

-- Post-migration verification
-- Add comprehensive verification queries here

EOF

    # Create rollback file for data migration
    cat > "$rollback_file" << EOF
-- Data Migration Rollback: $version
-- Description: Rollback for $description
-- Author: $author
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- Risk Level: $risk_level
-- JIRA Ticket: $jira_ticket

-- ============================================================================
-- DATA ROLLBACK START
-- ============================================================================

-- Enable timing
\timing on

-- Begin transaction
BEGIN;

-- Restore original data from backup
-- TRUNCATE target_table;
-- INSERT INTO target_table SELECT * FROM backup_${version}_original_data;

-- Verify rollback
-- SELECT COUNT(*) FROM target_table;
-- SELECT COUNT(*) FROM backup_${version}_original_data;

-- Clean up backup table
-- DROP TABLE IF EXISTS backup_${version}_original_data;

-- Commit transaction
COMMIT;

-- ============================================================================
-- DATA ROLLBACK END
-- ============================================================================

EOF

    # Create data file template
    cat > "$data_file" << EOF
-- Data file for migration: $version
-- Description: $description
-- Author: $author
-- Created: $(date '+%Y-%m-%d %H:%M:%S')

-- This file contains data to be inserted during migration
-- Use this for reference data, lookup tables, or initial data

-- Example:
-- INSERT INTO lookup_table (code, description) VALUES
-- ('CODE1', 'Description 1'),
-- ('CODE2', 'Description 2'),
-- ('CODE3', 'Description 3');

EOF

    log_success "Data migration files created:"
    log_info "  Migration: $migration_file"
    log_info "  Rollback:  $rollback_file"
    log_info "  Data:      $data_file"
    
    return 0
}

# Create hotfix migration template
create_hotfix_template() {
    local version="$1"
    local description="$2"
    local author="$3"
    local jira_ticket="$4"
    
    local migration_file="$MIGRATIONS_DIR/${version}_hotfix_$(echo "$description" | tr ' ' '_' | tr '[:upper:]' '[:lower:]').sql"
    local rollback_file="$ROLLBACKS_DIR/${version}_hotfix_$(echo "$description" | tr ' ' '_' | tr '[:upper:]' '[:lower:]')_rollback.sql"
    
    # Create hotfix migration file
    cat > "$migration_file" << EOF
-- HOTFIX Migration: $version
-- Description: $description
-- Author: $author
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- Risk Level: HIGH
-- Estimated Duration: 5m
-- Migration Type: HOTFIX
-- JIRA Ticket: $jira_ticket
--
-- ⚠️  CRITICAL HOTFIX - REQUIRES IMMEDIATE ATTENTION ⚠️
-- This is an emergency hotfix migration
-- Hotfix checklist:
-- [ ] Issue is critical and affects production
-- [ ] Fix has been tested in staging environment
-- [ ] Rollback plan is ready and tested
-- [ ] Stakeholders have been notified
-- [ ] Monitoring is in place to verify fix

-- ============================================================================
-- HOTFIX MIGRATION START
-- ============================================================================

-- Enable timing
\timing on

-- Set short timeout for hotfix
SET statement_timeout = '5min';

-- Begin transaction
BEGIN;

-- Add your hotfix SQL here
-- Keep it minimal and focused on the specific issue
-- Example:
-- UPDATE configuration_table 
-- SET config_value = 'fixed_value' 
-- WHERE config_key = 'problematic_setting';

-- Verify the fix
-- SELECT config_key, config_value 
-- FROM configuration_table 
-- WHERE config_key = 'problematic_setting';

-- Commit transaction
COMMIT;

-- ============================================================================
-- HOTFIX MIGRATION END
-- ============================================================================

EOF

    # Create hotfix rollback file
    cat > "$rollback_file" << EOF
-- HOTFIX Rollback: $version
-- Description: Rollback for $description
-- Author: $author
-- Created: $(date '+%Y-%m-%d %H:%M:%S')
-- JIRA Ticket: $jira_ticket

-- ============================================================================
-- HOTFIX ROLLBACK START
-- ============================================================================

-- Enable timing
\timing on

-- Set short timeout
SET statement_timeout = '5min';

-- Begin transaction
BEGIN;

-- Add your rollback SQL here
-- Example:
-- UPDATE configuration_table 
-- SET config_value = 'original_value' 
-- WHERE config_key = 'problematic_setting';

-- Verify the rollback
-- SELECT config_key, config_value 
-- FROM configuration_table 
-- WHERE config_key = 'problematic_setting';

-- Commit transaction
COMMIT;

-- ============================================================================
-- HOTFIX ROLLBACK END
-- ============================================================================

EOF

    log_success "Hotfix migration files created:"
    log_info "  Migration: $migration_file"
    log_info "  Rollback:  $rollback_file"
    
    log_warning "⚠️  This is a HOTFIX migration with HIGH risk level"
    log_warning "⚠️  Ensure proper testing and approval before deployment"
    
    return 0
}

# Interactive migration creation
interactive_create() {
    log_info "Creating new database migration..."
    echo ""
    
    # Get migration details
    local description=$(get_input "Migration description")
    local migration_type=$(get_input "Migration type" "SCHEMA_CHANGE" '[[ "$value" =~ ^(SCHEMA_CHANGE|DATA_MIGRATION|HOTFIX|INDEX_CREATION|CONSTRAINT_ADDITION)$ ]]')
    local author=$(get_input "Author" "$(git config user.name 2>/dev/null || whoami)")
    local feature_branch=$(get_input "Feature branch" "$(git branch --show-current 2>/dev/null || echo 'main')")
    local jira_ticket=$(get_input "JIRA ticket (optional)" "")
    
    # Generate version
    local version=$(generate_version)
    
    case "$migration_type" in
        "HOTFIX")
            create_hotfix_template "$version" "$description" "$author" "$jira_ticket"
            ;;
        "DATA_MIGRATION")
            local risk_level=$(get_input "Risk level" "MEDIUM" 'validate_risk_level')
            local estimated_duration=$(get_input "Estimated duration (e.g., 5m, 30s, 2h)" "10m" 'validate_duration')
            create_data_migration_template "$version" "$description" "$risk_level" "$estimated_duration" "$author" "$feature_branch" "$jira_ticket"
            ;;
        *)
            local risk_level=$(get_input "Risk level" "LOW" 'validate_risk_level')
            local estimated_duration=$(get_input "Estimated duration (e.g., 5m, 30s, 2h)" "5m" 'validate_duration')
            create_migration_template "$version" "$description" "$risk_level" "$estimated_duration" "$migration_type" "$author" "$feature_branch" "$jira_ticket"
            ;;
    esac
    
    echo ""
    log_info "Next steps:"
    log_info "1. Edit the migration file to add your SQL changes"
    log_info "2. Edit the rollback file to add the reverse operations"
    log_info "3. Test both migration and rollback in development environment"
    log_info "4. Get required approvals before deploying to production"
    
    if [[ "$migration_type" == "HOTFIX" ]]; then
        log_warning "5. ⚠️  HOTFIX: Follow emergency deployment procedures"
    fi
}

# Usage information
usage() {
    echo "Migration Template Generator"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --interactive, -i       Interactive migration creation (default)"
    echo "  --type <type>          Migration type (SCHEMA_CHANGE, DATA_MIGRATION, HOTFIX, etc.)"
    echo "  --description <desc>   Migration description"
    echo "  --author <name>        Author name"
    echo "  --risk <level>         Risk level (LOW, MEDIUM, HIGH)"
    echo "  --duration <time>      Estimated duration (e.g., 5m, 30s, 2h)"
    echo "  --branch <name>        Feature branch name"
    echo "  --ticket <id>          JIRA ticket ID"
    echo "  --help, -h             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                     # Interactive mode"
    echo "  $0 --type SCHEMA_CHANGE --description 'Add user preferences table'"
    echo "  $0 --type HOTFIX --description 'Fix critical login issue'"
    echo "  $0 --type DATA_MIGRATION --description 'Migrate user data' --risk HIGH"
}

# Main function
main() {
    local interactive="true"
    local migration_type=""
    local description=""
    local author="$(git config user.name 2>/dev/null || whoami)"
    local risk_level="LOW"
    local estimated_duration="5m"
    local feature_branch="$(git branch --show-current 2>/dev/null || echo 'main')"
    local jira_ticket=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --interactive|-i)
                interactive="true"
                shift
                ;;
            --type)
                migration_type="$2"
                interactive="false"
                shift 2
                ;;
            --description)
                description="$2"
                interactive="false"
                shift 2
                ;;
            --author)
                author="$2"
                shift 2
                ;;
            --risk)
                risk_level="$2"
                shift 2
                ;;
            --duration)
                estimated_duration="$2"
                shift 2
                ;;
            --branch)
                feature_branch="$2"
                shift 2
                ;;
            --ticket)
                jira_ticket="$2"
                shift 2
                ;;
            --help|-h)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
    
    # Run interactive mode or create migration with provided parameters
    if [[ "$interactive" == "true" ]]; then
        interactive_create
    else
        if [[ -z "$migration_type" || -z "$description" ]]; then
            log_error "Migration type and description are required in non-interactive mode"
            usage
            exit 1
        fi
        
        local version=$(generate_version)
        
        case "$migration_type" in
            "HOTFIX")
                create_hotfix_template "$version" "$description" "$author" "$jira_ticket"
                ;;
            "DATA_MIGRATION")
                create_data_migration_template "$version" "$description" "$risk_level" "$estimated_duration" "$author" "$feature_branch" "$jira_ticket"
                ;;
            *)
                create_migration_template "$version" "$description" "$risk_level" "$estimated_duration" "$migration_type" "$author" "$feature_branch" "$jira_ticket"
                ;;
        esac
    fi
}

# Execute main function
main "$@"