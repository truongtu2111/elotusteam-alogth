#!/bin/bash

# Database Migration Framework Setup Script
# This script sets up the complete migration framework with data ops controls

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

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

log_header() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_header "Checking Prerequisites"
    
    local missing_deps=()
    
    # Check for required commands
    if ! command -v psql &> /dev/null; then
        missing_deps+=("postgresql-client")
    fi
    
    if ! command -v python3 &> /dev/null; then
        missing_deps+=("python3")
    fi
    
    if ! command -v pip3 &> /dev/null; then
        missing_deps+=("python3-pip")
    fi
    
    if ! command -v git &> /dev/null; then
        missing_deps+=("git")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Please install the missing dependencies and run this script again."
        
        if [[ "$(uname)" == "Darwin" ]]; then
            log_info "On macOS, you can install dependencies using Homebrew:"
            log_info "  brew install postgresql python3 git"
        elif [[ "$(uname)" == "Linux" ]]; then
            log_info "On Ubuntu/Debian, you can install dependencies using:"
            log_info "  sudo apt-get update"
            log_info "  sudo apt-get install postgresql-client python3 python3-pip git"
        fi
        
        exit 1
    fi
    
    log_success "All prerequisites are satisfied"
}

# Setup directory structure
setup_directories() {
    log_header "Setting Up Directory Structure"
    
    local directories=(
        "$SCRIPT_DIR/scripts"
        "$SCRIPT_DIR/rollbacks"
        "$SCRIPT_DIR/data"
        "$SCRIPT_DIR/config"
        "$SCRIPT_DIR/tools"
        "$SCRIPT_DIR/templates"
        "$SCRIPT_DIR/logs"
        "$SCRIPT_DIR/backups"
    )
    
    for dir in "${directories[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            log_info "Created directory: $dir"
        else
            log_info "Directory already exists: $dir"
        fi
    done
    
    log_success "Directory structure setup completed"
}

# Install Python dependencies
install_python_dependencies() {
    log_header "Installing Python Dependencies"
    
    # Create requirements.txt for the dashboard
    cat > "$SCRIPT_DIR/requirements.txt" << 'EOF'
Flask==2.3.3
Flask-Login==0.6.3
psycopg2-binary==2.9.7
PyYAML==6.0.1
requests==2.31.0
Werkzeug==2.3.7
EOF
    
    log_info "Installing Python dependencies..."
    
    # Install dependencies
    if pip3 install -r "$SCRIPT_DIR/requirements.txt" --user; then
        log_success "Python dependencies installed successfully"
    else
        log_warning "Failed to install some Python dependencies. Dashboard may not work properly."
    fi
}

# Setup configuration files
setup_configuration() {
    log_header "Setting Up Configuration Files"
    
    # Check if configuration files already exist
    if [[ -f "$SCRIPT_DIR/config/environments.yml" ]]; then
        log_info "Configuration files already exist"
        return 0
    fi
    
    log_info "Configuration files will be created during the migration setup process"
    log_success "Configuration setup completed"
}

# Make scripts executable
make_scripts_executable() {
    log_header "Making Scripts Executable"
    
    local scripts=(
        "$SCRIPT_DIR/tools/migrate.sh"
        "$SCRIPT_DIR/tools/create-migration.sh"
        "$SCRIPT_DIR/setup-migration-framework.sh"
    )
    
    for script in "${scripts[@]}"; do
        if [[ -f "$script" ]]; then
            chmod +x "$script"
            log_info "Made executable: $(basename "$script")"
        else
            log_warning "Script not found: $script"
        fi
    done
    
    log_success "Scripts made executable"
}

# Create sample migration
create_sample_migration() {
    log_header "Creating Sample Migration"
    
    local sample_migration="$SCRIPT_DIR/scripts/001_20240101_000000_sample_migration.sql"
    local sample_rollback="$SCRIPT_DIR/rollbacks/001_20240101_000000_sample_migration_rollback.sql"
    
    if [[ -f "$sample_migration" ]]; then
        log_info "Sample migration already exists"
        return 0
    fi
    
    # Create sample migration
    cat > "$sample_migration" << 'EOF'
-- Migration: 001_20240101_000000_sample_migration
-- Description: Sample migration to demonstrate the framework
-- Author: Migration Framework
-- Created: 2024-01-01 00:00:00
-- Risk Level: LOW
-- Estimated Duration: 1m
-- Migration Type: SCHEMA_CHANGE
-- Feature Branch: main
-- JIRA Ticket: SAMPLE-001
--
-- This is a sample migration to demonstrate the migration framework
-- It creates a simple table for demonstration purposes

-- ============================================================================
-- MIGRATION START
-- ============================================================================

-- Enable timing for performance monitoring
\timing on

-- Set statement timeout for safety
SET statement_timeout = '5min';

-- Begin transaction
BEGIN;

-- Create sample table
CREATE TABLE IF NOT EXISTS sample_table (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_sample_table_name ON sample_table(name);

-- Insert sample data
INSERT INTO sample_table (name, description) VALUES
('Sample Record 1', 'This is a sample record created by the migration framework'),
('Sample Record 2', 'This demonstrates data insertion during migration')
ON CONFLICT DO NOTHING;

-- Verify the changes
SELECT COUNT(*) as record_count FROM sample_table;

-- Commit transaction
COMMIT;

-- ============================================================================
-- MIGRATION END
-- ============================================================================

-- Post-migration verification
SELECT 
    schemaname,
    tablename,
    attname,
    typename,
    attnotnull
FROM pg_stats 
JOIN pg_attribute ON pg_stats.attname = pg_attribute.attname
JOIN pg_type ON pg_attribute.atttypid = pg_type.oid
WHERE schemaname = 'public' AND tablename = 'sample_table'
LIMIT 5;
EOF
    
    # Create sample rollback
    cat > "$sample_rollback" << 'EOF'
-- Rollback Migration: 001_20240101_000000_sample_migration
-- Description: Rollback for sample migration
-- Author: Migration Framework
-- Created: 2024-01-01 00:00:00
-- Risk Level: LOW
-- JIRA Ticket: SAMPLE-001
--
-- This rollback script removes the sample table and data

-- ============================================================================
-- ROLLBACK START
-- ============================================================================

-- Enable timing for performance monitoring
\timing on

-- Set statement timeout for safety
SET statement_timeout = '5min';

-- Begin transaction
BEGIN;

-- Drop index
DROP INDEX IF EXISTS idx_sample_table_name;

-- Drop table (this will also remove all data)
DROP TABLE IF EXISTS sample_table;

-- Verify the rollback
SELECT COUNT(*) FROM information_schema.tables 
WHERE table_schema = 'public' AND table_name = 'sample_table';
-- Should return 0 if table was properly dropped

-- Commit transaction
COMMIT;

-- ============================================================================
-- ROLLBACK END
-- ============================================================================
EOF
    
    log_success "Sample migration created"
    log_info "  Migration: $sample_migration"
    log_info "  Rollback:  $sample_rollback"
}

# Create environment setup script
create_env_setup() {
    log_header "Creating Environment Setup Script"
    
    cat > "$SCRIPT_DIR/setup-environment.sh" << 'EOF'
#!/bin/bash

# Environment Setup Script for Database Migrations
# This script helps set up environment variables for database connections

set -e

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

echo "Database Migration Environment Setup"
echo "===================================="
echo ""

log_info "This script will help you set up environment variables for database connections."
log_info "You can either export these variables in your shell or add them to your .env file."
echo ""

# Get database connection details
read -p "Development Database Host [localhost]: " DEV_HOST
DEV_HOST=${DEV_HOST:-localhost}

read -p "Development Database Port [5432]: " DEV_PORT
DEV_PORT=${DEV_PORT:-5432}

read -p "Development Database Name [postgres]: " DEV_DB
DEV_DB=${DEV_DB:-postgres}

read -p "Development Database User [postgres]: " DEV_USER
DEV_USER=${DEV_USER:-postgres}

read -s -p "Development Database Password: " DEV_PASSWORD
echo ""

echo ""
log_info "Staging Database Configuration (press Enter to use same as development):"

read -p "Staging Database Host [$DEV_HOST]: " STAGING_HOST
STAGING_HOST=${STAGING_HOST:-$DEV_HOST}

read -p "Staging Database Port [$DEV_PORT]: " STAGING_PORT
STAGING_PORT=${STAGING_PORT:-$DEV_PORT}

read -p "Staging Database Name [$DEV_DB]: " STAGING_DB
STAGING_DB=${STAGING_DB:-$DEV_DB}

read -p "Staging Database User [$DEV_USER]: " STAGING_USER
STAGING_USER=${STAGING_USER:-$DEV_USER}

read -s -p "Staging Database Password [$DEV_PASSWORD]: " STAGING_PASSWORD
STAGING_PASSWORD=${STAGING_PASSWORD:-$DEV_PASSWORD}
echo ""

echo ""
log_info "Production Database Configuration:"

read -p "Production Database Host: " PROD_HOST
read -p "Production Database Port [5432]: " PROD_PORT
PROD_PORT=${PROD_PORT:-5432}

read -p "Production Database Name: " PROD_DB
read -p "Production Database User: " PROD_USER
read -s -p "Production Database Password: " PROD_PASSWORD
echo ""

# Create .env file
cat > .env << EOL
# Database Migration Environment Variables
# Generated on $(date)

# Development Database
export DEV_DB_HOST="$DEV_HOST"
export DEV_DB_PORT="$DEV_PORT"
export DEV_DB_NAME="$DEV_DB"
export DEV_DB_USER="$DEV_USER"
export DEV_DB_PASSWORD="$DEV_PASSWORD"

# Staging Database
export STAGING_DB_HOST="$STAGING_HOST"
export STAGING_DB_PORT="$STAGING_PORT"
export STAGING_DB_NAME="$STAGING_DB"
export STAGING_DB_USER="$STAGING_USER"
export STAGING_DB_PASSWORD="$STAGING_PASSWORD"

# Production Database
export PROD_DB_HOST="$PROD_HOST"
export PROD_DB_PORT="$PROD_PORT"
export PROD_DB_NAME="$PROD_DB"
export PROD_DB_USER="$PROD_USER"
export PROD_DB_PASSWORD="$PROD_PASSWORD"

# Notification Settings (optional)
# export SLACK_WEBHOOK_URL="your-slack-webhook-url"
# export EMAIL_RECIPIENTS="admin@company.com,dataops@company.com"
# export SMTP_SERVER="smtp.company.com"
# export SMTP_PORT="587"
# export SMTP_USERNAME="notifications@company.com"
# export SMTP_PASSWORD="your-smtp-password"
EOL

log_success "Environment configuration saved to .env file"
log_info "To use these variables, run: source .env"
log_warning "Remember to add .env to your .gitignore file to avoid committing sensitive information"

echo ""
log_info "Next steps:"
log_info "1. Review and source the .env file: source .env"
log_info "2. Test database connections"
log_info "3. Run the migration framework setup"
EOF
    
    chmod +x "$SCRIPT_DIR/setup-environment.sh"
    
    log_success "Environment setup script created: setup-environment.sh"
}

# Create quick start guide
create_quick_start_guide() {
    log_header "Creating Quick Start Guide"
    
    cat > "$SCRIPT_DIR/QUICK_START.md" << 'EOF'
# Database Migration Framework - Quick Start Guide

## Overview

This migration framework provides comprehensive database migration management with data operations team controls, approval workflows, and monitoring capabilities.

## Quick Setup

### 1. Environment Setup

```bash
# Run the environment setup script
./setup-environment.sh

# Source the environment variables
source .env
```

### 2. Initialize Migration Infrastructure

```bash
# Initialize the migration infrastructure in development
./tools/migrate.sh migrate development
```

### 3. Create Your First Migration

```bash
# Interactive migration creation
./tools/create-migration.sh

# Or create with parameters
./tools/create-migration.sh --type SCHEMA_CHANGE --description "Add user preferences table"
```

### 4. Apply Migrations

```bash
# Apply to development
./tools/migrate.sh migrate development

# Apply to staging (requires approvals)
./tools/migrate.sh migrate staging

# Apply to production (requires multiple approvals)
./tools/migrate.sh migrate production
```

### 5. Start Data Ops Dashboard

```bash
# Install Python dependencies
pip3 install -r requirements.txt

# Start the dashboard
python3 tools/data-ops-dashboard.py

# Access at http://localhost:5000
# Default accounts:
#   admin / admin123
#   data_ops_lead / dataops123
#   developer / dev123
```

## Common Commands

### Migration Management

```bash
# Check migration status
./tools/migrate.sh status development

# Validate a migration script
./tools/migrate.sh validate scripts/001_example.sql

# Rollback a migration
./tools/migrate.sh rollback development 001_example
```

### Approval Workflow

```bash
# Add approval for a migration
./tools/migrate.sh approve 001_example data_ops_lead john.doe "Approved after review"

# Lock environment for maintenance
./tools/migrate.sh lock production "Scheduled maintenance"

# Unlock environment
./tools/migrate.sh unlock production
```

### Data Operations Dashboard

The web dashboard provides:

- **Migration Status Overview**: Real-time status across all environments
- **Approval Management**: Web-based approval workflow
- **Environment Controls**: Lock/unlock environments
- **Migration History**: Complete audit trail
- **Real-time Monitoring**: Live updates and notifications

## Migration Types

### Schema Changes
- Table creation/modification
- Index management
- Constraint additions

### Data Migrations
- Data transformations
- Bulk data operations
- Reference data updates

### Hotfixes
- Emergency fixes
- Critical issue resolution
- Fast-track deployment

## Risk Levels

- **LOW**: Simple schema changes, low impact
- **MEDIUM**: Data modifications, moderate impact
- **HIGH**: Complex operations, high impact

## Approval Matrix

| Environment | Risk Level | Required Approvals |
|-------------|------------|-------------------|
| Development | Any | 1 (Developer) |
| Staging | Any | 2 (Data Ops) |
| Production | LOW | 2 (Data Ops Lead + DBA) |
| Production | MEDIUM | 3 (Data Ops Lead + DBA + Tech Lead) |
| Production | HIGH | 4 (Data Ops Lead + DBA + Tech Lead + CTO) |

## Best Practices

1. **Always create rollback scripts** for every migration
2. **Test migrations in development** before promoting
3. **Use descriptive migration names** and comments
4. **Follow the approval process** for production deployments
5. **Monitor migration performance** and impact
6. **Keep migrations small and focused** on single changes
7. **Use transactions** to ensure atomicity
8. **Document breaking changes** and dependencies

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check environment variables
   - Verify database credentials
   - Ensure database is running

2. **Migration Lock Error**
   - Check for existing locks: `./tools/migrate.sh status <env>`
   - Release stuck locks: `./tools/migrate.sh unlock <env>`

3. **Approval Insufficient**
   - Check approval status in dashboard
   - Add required approvals
   - Use `--force` flag for emergency deployments

4. **Migration Failed**
   - Check migration logs in `logs/migration.log`
   - Review rollback script
   - Use rollback command if needed

### Getting Help

- Check the comprehensive guides in the documentation
- Review migration logs for detailed error information
- Use the data ops dashboard for real-time monitoring
- Contact the data operations team for assistance

## Security Considerations

- Store database passwords in environment variables
- Use SSL connections for production databases
- Implement proper access controls
- Audit all migration activities
- Backup databases before major migrations

## Next Steps

1. Read the complete documentation in `DATABASE_MIGRATION_GUIDE.md`
2. Set up monitoring and alerting
3. Configure notification channels (Slack, email)
4. Establish backup and recovery procedures
5. Train team members on the migration process
EOF
    
    log_success "Quick start guide created: QUICK_START.md"
}

# Test database connection
test_database_connection() {
    log_header "Testing Database Connection"
    
    if [[ -f "$SCRIPT_DIR/.env" ]]; then
        log_info "Loading environment variables from .env file..."
        source "$SCRIPT_DIR/.env"
    fi
    
    # Test development database connection
    if [[ -n "$DEV_DB_HOST" && -n "$DEV_DB_PASSWORD" ]]; then
        log_info "Testing development database connection..."
        
        if PGPASSWORD="$DEV_DB_PASSWORD" psql -h "$DEV_DB_HOST" -p "${DEV_DB_PORT:-5432}" -U "${DEV_DB_USER:-postgres}" -d "${DEV_DB_NAME:-postgres}" -c "SELECT 1;" &> /dev/null; then
            log_success "Development database connection successful"
        else
            log_warning "Development database connection failed"
            log_info "Please check your database configuration and ensure the database is running"
        fi
    else
        log_warning "Development database configuration not found"
        log_info "Run ./setup-environment.sh to configure database connections"
    fi
}

# Display completion message
show_completion_message() {
    log_header "Setup Complete!"
    
    echo -e "${GREEN}Database Migration Framework has been successfully set up!${NC}"
    echo ""
    echo "📁 Directory Structure:"
    echo "   migrations/"
    echo "   ├── scripts/          # Migration SQL files"
    echo "   ├── rollbacks/        # Rollback SQL files"
    echo "   ├── data/             # Data files for migrations"
    echo "   ├── config/           # Configuration files"
    echo "   ├── tools/            # Migration tools and scripts"
    echo "   ├── templates/        # Migration templates"
    echo "   ├── logs/             # Migration logs"
    echo "   └── backups/          # Database backups"
    echo ""
    echo "🛠️  Available Tools:"
    echo "   ./tools/migrate.sh              # Main migration tool"
    echo "   ./tools/create-migration.sh     # Migration template generator"
    echo "   ./tools/data-ops-dashboard.py   # Web-based dashboard"
    echo "   ./setup-environment.sh          # Environment configuration"
    echo ""
    echo "📚 Documentation:"
    echo "   QUICK_START.md                  # Quick start guide"
    echo "   DATABASE_MIGRATION_GUIDE.md     # Comprehensive guide"
    echo "   MONITORING_SETUP_GUIDE.md       # Monitoring setup"
    echo ""
    echo "🚀 Next Steps:"
    echo "   1. Configure database connections: ./setup-environment.sh"
    echo "   2. Initialize migration infrastructure: ./tools/migrate.sh migrate development"
    echo "   3. Create your first migration: ./tools/create-migration.sh"
    echo "   4. Start the data ops dashboard: python3 tools/data-ops-dashboard.py"
    echo ""
    echo "💡 Quick Commands:"
    echo "   # Check status"
    echo "   ./tools/migrate.sh status development"
    echo ""
    echo "   # Create migration"
    echo "   ./tools/create-migration.sh"
    echo ""
    echo "   # Apply migrations"
    echo "   ./tools/migrate.sh migrate development"
    echo ""
    echo "   # Start dashboard"
    echo "   python3 tools/data-ops-dashboard.py"
    echo ""
    log_success "Happy migrating! 🎉"
}

# Main execution
main() {
    local skip_deps="false"
    local skip_test="false"
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-deps)
                skip_deps="true"
                shift
                ;;
            --skip-test)
                skip_test="true"
                shift
                ;;
            --help|-h)
                echo "Database Migration Framework Setup"
                echo ""
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --skip-deps    Skip Python dependency installation"
                echo "  --skip-test    Skip database connection test"
                echo "  --help, -h     Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    echo -e "${CYAN}"
    echo "██████╗  █████╗ ████████╗ █████╗ ██████╗  █████╗ ███████╗███████╗"
    echo "██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔════╝"
    echo "██║  ██║███████║   ██║   ███████║██████╔╝███████║███████╗█████╗  "
    echo "██║  ██║██╔══██║   ██║   ██╔══██║██╔══██╗██╔══██║╚════██║██╔══╝  "
    echo "██████╔╝██║  ██║   ██║   ██║  ██║██████╔╝██║  ██║███████║███████╗"
    echo "╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝"
    echo ""
    echo "███╗   ███╗██╗ ██████╗ ██████╗  █████╗ ████████╗██╗ ██████╗ ███╗   ██╗"
    echo "████╗ ████║██║██╔════╝ ██╔══██╗██╔══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║"
    echo "██╔████╔██║██║██║  ███╗██████╔╝███████║   ██║   ██║██║   ██║██╔██╗ ██║"
    echo "██║╚██╔╝██║██║██║   ██║██╔══██╗██╔══██║   ██║   ██║██║   ██║██║╚██╗██║"
    echo "██║ ╚═╝ ██║██║╚██████╔╝██║  ██║██║  ██║   ██║   ██║╚██████╔╝██║ ╚████║"
    echo "╚═╝     ╚═╝╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝"
    echo ""
    echo "███████╗██████╗  █████╗ ███╗   ███╗███████╗██╗    ██╗ ██████╗ ██████╗ ██╗  ██╗"
    echo "██╔════╝██╔══██╗██╔══██╗████╗ ████║██╔════╝██║    ██║██╔═══██╗██╔══██╗██║ ██╔╝"
    echo "█████╗  ██████╔╝███████║██╔████╔██║█████╗  ██║ █╗ ██║██║   ██║██████╔╝█████╔╝ "
    echo "██╔══╝  ██╔══██╗██╔══██║██║╚██╔╝██║██╔══╝  ██║███╗██║██║   ██║██╔══██╗██╔═██╗ "
    echo "██║     ██║  ██║██║  ██║██║ ╚═╝ ██║███████╗╚███╔███╔╝╚██████╔╝██║  ██║██║  ██╗"
    echo "╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝ ╚══╝╚══╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝"
    echo -e "${NC}"
    echo ""
    echo -e "${BLUE}Setting up Database Migration Framework with Data Ops Controls${NC}"
    echo ""
    
    # Run setup steps
    check_prerequisites
    setup_directories
    
    if [[ "$skip_deps" != "true" ]]; then
        install_python_dependencies
    fi
    
    setup_configuration
    make_scripts_executable
    create_sample_migration
    create_env_setup
    create_quick_start_guide
    
    if [[ "$skip_test" != "true" ]]; then
        test_database_connection
    fi
    
    show_completion_message
}

# Execute main function
main "$@"