# Database Migration Framework with Data Operations Controls

## Overview

This comprehensive database migration framework provides enterprise-grade migration management with data operations team controls, approval workflows, monitoring integration, and automated deployment capabilities. It's designed to ensure safe, auditable, and controlled database changes across all environments.

## 🚀 Quick Start

### 1. Initial Setup

```bash
# Navigate to the migrations directory
cd migrations/

# Run the complete setup script
./setup-migration-framework.sh

# Configure database connections
./setup-environment.sh

# Source environment variables
source .env
```

### 2. Initialize Migration Infrastructure

```bash
# Initialize migration tables and infrastructure
./tools/migrate.sh migrate development
```

### 3. Create Your First Migration

```bash
# Interactive migration creation
./tools/create-migration.sh

# Or with parameters
./tools/create-migration.sh --type SCHEMA_CHANGE --description "Add user preferences table"
```

### 4. Start Data Operations Dashboard

```bash
# Install dependencies (if not done during setup)
pip3 install -r requirements.txt

# Start the web dashboard
python3 tools/data-ops-dashboard.py

# Access at http://localhost:5000
```

## 📁 Directory Structure

```
migrations/
├── scripts/                    # Migration SQL files
│   ├── 001_20240101_000000_sample_migration.sql
│   └── ...
├── rollbacks/                  # Rollback SQL files
│   ├── 001_20240101_000000_sample_migration_rollback.sql
│   └── ...
├── data/                       # Data files for migrations
│   └── sample_data.json
├── config/                     # Configuration files
│   ├── environments.yml        # Environment configurations
│   └── approval_matrix.yml     # Approval workflow rules
├── tools/                      # Migration tools and scripts
│   ├── migrate.sh              # Main migration tool
│   ├── create-migration.sh     # Migration template generator
│   └── data-ops-dashboard.py   # Web-based dashboard
├── templates/                  # Migration templates
├── logs/                       # Migration logs
├── backups/                    # Database backups
├── setup-migration-framework.sh # Complete setup script
├── setup-environment.sh        # Environment configuration
├── requirements.txt            # Python dependencies
├── QUICK_START.md             # Quick start guide
├── DATABASE_MIGRATION_GUIDE.md # Comprehensive documentation
└── README.md                  # This file
```

## 🛠️ Core Components

### 1. Migration Management Tool (`migrate.sh`)

The main tool for managing database migrations with comprehensive features:

```bash
# Apply migrations
./tools/migrate.sh migrate <environment>

# Check status
./tools/migrate.sh status <environment>

# Rollback migrations
./tools/migrate.sh rollback <environment> <migration_id>

# Validate migration scripts
./tools/migrate.sh validate <script_path>

# Manage approvals
./tools/migrate.sh approve <migration_id> <role> <approver> "<comment>"

# Lock/unlock environments
./tools/migrate.sh lock <environment> "<reason>"
./tools/migrate.sh unlock <environment>
```

### 2. Migration Creator (`create-migration.sh`)

Generates standardized migration templates:

```bash
# Interactive mode
./tools/create-migration.sh

# Command line mode
./tools/create-migration.sh \
  --type SCHEMA_CHANGE \
  --description "Add user preferences table" \
  --risk LOW \
  --duration "5m" \
  --author "john.doe" \
  --ticket "PROJ-123"
```

### 3. Data Operations Dashboard (`data-ops-dashboard.py`)

Web-based interface for migration management:

- **Real-time Status**: Live migration status across environments
- **Approval Workflow**: Web-based approval management
- **Environment Controls**: Lock/unlock environments
- **Migration History**: Complete audit trail
- **Monitoring Integration**: Performance metrics and alerts
- **User Management**: Role-based access control

**Default Accounts:**
- `admin` / `admin123` (Administrator)
- `data_ops_lead` / `dataops123` (Data Operations Lead)
- `developer` / `dev123` (Developer)

## 🔐 Security & Access Control

### Role-Based Access

- **Developer**: Create migrations, apply to development
- **Data Operations**: Approve staging migrations, manage environments
- **Data Operations Lead**: Approve production migrations, emergency access
- **DBA**: Database-specific approvals, performance review
- **Tech Lead**: Architecture review, high-risk approvals
- **CTO**: Critical production approvals

### Approval Matrix

| Environment | Risk Level | Required Approvals |
|-------------|------------|-------------------|
| Development | Any | 1 (Developer) |
| Staging | Any | 2 (Data Ops Team) |
| Production | LOW | 2 (Data Ops Lead + DBA) |
| Production | MEDIUM | 3 (Data Ops Lead + DBA + Tech Lead) |
| Production | HIGH | 4 (Data Ops Lead + DBA + Tech Lead + CTO) |

### Security Features

- Environment variable-based credentials
- SSL/TLS database connections
- Migration locking mechanisms
- Comprehensive audit logging
- Role-based dashboard access
- Approval workflow enforcement

## 📊 Monitoring Integration

The migration framework integrates with the existing monitoring stack:

### Prometheus Metrics

- Migration execution time
- Success/failure rates
- Environment status
- Approval workflow metrics
- Database performance impact

### Grafana Dashboards

- Migration performance overview
- Environment health status
- Approval workflow tracking
- Historical migration trends

### Alerting

- Failed migration alerts
- Long-running migration warnings
- Environment lock notifications
- Approval deadline reminders

## 🔄 Migration Types

### Schema Changes
```sql
-- Table creation/modification
CREATE TABLE user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index management
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
```

### Data Migrations
```sql
-- Data transformations
UPDATE users 
SET email = LOWER(email) 
WHERE email != LOWER(email);

-- Bulk operations with progress tracking
DO $$
DECLARE
    batch_size INTEGER := 1000;
    processed INTEGER := 0;
    total INTEGER;
BEGIN
    SELECT COUNT(*) INTO total FROM legacy_table;
    
    WHILE processed < total LOOP
        -- Process batch
        INSERT INTO new_table (column1, column2)
        SELECT old_column1, old_column2
        FROM legacy_table
        LIMIT batch_size OFFSET processed;
        
        processed := processed + batch_size;
        RAISE NOTICE 'Processed % of % records', processed, total;
    END LOOP;
END $$;
```

### Hotfixes
```sql
-- Emergency fixes with minimal impact
UPDATE configuration 
SET value = 'fixed_value' 
WHERE key = 'problematic_setting' 
  AND value = 'problematic_value';
```

## 🚨 Emergency Procedures

### Emergency Migration Deployment

```bash
# Force migration without full approval (emergency only)
./tools/migrate.sh migrate production --force --emergency "Critical security fix"

# Emergency rollback
./tools/migrate.sh rollback production <migration_id> --force --emergency
```

### Environment Recovery

```bash
# Unlock stuck environment
./tools/migrate.sh unlock production --force

# Reset migration state
./tools/migrate.sh reset production --confirm

# Restore from backup
./tools/migrate.sh restore production <backup_file>
```

## 📈 Best Practices

### Migration Development

1. **Small, Focused Changes**: Keep migrations small and focused on single changes
2. **Descriptive Naming**: Use clear, descriptive names and comments
3. **Rollback Scripts**: Always create corresponding rollback scripts
4. **Testing**: Test thoroughly in development before promoting
5. **Performance**: Consider performance impact and use appropriate timeouts
6. **Dependencies**: Document migration dependencies and order

### Code Examples

```sql
-- Good: Small, focused migration
-- Migration: Add email verification column
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_users_email_verified ON users(email_verified);

-- Good: Transaction with rollback
BEGIN;

-- Migration logic here
ALTER TABLE orders ADD COLUMN status_updated_at TIMESTAMP;
UPDATE orders SET status_updated_at = updated_at WHERE status_updated_at IS NULL;
ALTER TABLE orders ALTER COLUMN status_updated_at SET NOT NULL;

-- Verify changes
SELECT COUNT(*) FROM orders WHERE status_updated_at IS NULL;
-- Should return 0

COMMIT;
```

### Approval Workflow

1. **Development**: Automatic approval for developers
2. **Staging**: Data operations team review
3. **Production**: Multi-level approval based on risk
4. **Emergency**: Streamlined process with post-deployment review

### Monitoring

1. **Pre-Migration**: Check system health and performance
2. **During Migration**: Monitor execution progress and impact
3. **Post-Migration**: Verify success and performance impact
4. **Ongoing**: Track long-term effects and optimization opportunities

## 🔧 Configuration

### Environment Configuration (`config/environments.yml`)

```yaml
development:
  database:
    host: "${DEV_DB_HOST}"
    port: "${DEV_DB_PORT}"
    database: "${DEV_DB_NAME}"
    user: "${DEV_DB_USER}"
    password: "${DEV_DB_PASSWORD}"
  migration:
    auto_approve: true
    backup_required: false
    timeout: 300
  monitoring:
    enabled: true
    metrics_endpoint: "http://localhost:9090"
```

### Approval Matrix (`config/approval_matrix.yml`)

```yaml
production:
  HIGH:
    required_approvals: 4
    required_roles:
      - data_ops_lead
      - dba
      - tech_lead
      - cto
    timeout_hours: 24
    emergency_override: true
```

## 🚀 Advanced Features

### Automated Testing

```bash
# Run migration tests
./tools/test-migrations.sh development

# Performance testing
./tools/benchmark-migration.sh <migration_file>

# Rollback testing
./tools/test-rollback.sh <migration_id>
```

### Backup Integration

```bash
# Automatic backup before migration
./tools/migrate.sh migrate production --backup

# Manual backup
./tools/backup-database.sh production

# Restore from backup
./tools/restore-database.sh production <backup_file>
```

### Notification Integration

```bash
# Configure Slack notifications
export SLACK_WEBHOOK_URL="https://hooks.slack.com/..."

# Configure email notifications
export EMAIL_RECIPIENTS="admin@company.com,dataops@company.com"
export SMTP_SERVER="smtp.company.com"
```

## 📚 Documentation

- **[QUICK_START.md](QUICK_START.md)**: Quick start guide
- **[DATABASE_MIGRATION_GUIDE.md](DATABASE_MIGRATION_GUIDE.md)**: Comprehensive documentation
- **[MONITORING_SETUP_GUIDE.md](../MONITORING_SETUP_GUIDE.md)**: Monitoring integration
- **[MONITORING_QUICK_REFERENCE.md](../MONITORING_QUICK_REFERENCE.md)**: Monitoring commands

## 🆘 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check environment variables
   echo $DEV_DB_HOST $DEV_DB_USER
   
   # Test connection manually
   psql -h $DEV_DB_HOST -U $DEV_DB_USER -d $DEV_DB_NAME
   ```

2. **Migration Lock Error**
   ```bash
   # Check lock status
   ./tools/migrate.sh status production
   
   # Release lock if needed
   ./tools/migrate.sh unlock production
   ```

3. **Approval Insufficient**
   ```bash
   # Check approval status
   ./tools/migrate.sh status production
   
   # Add approval via dashboard or command line
   ./tools/migrate.sh approve <migration_id> data_ops_lead john.doe "Approved"
   ```

4. **Migration Failed**
   ```bash
   # Check logs
   tail -f logs/migration.log
   
   # Review migration details
   ./tools/migrate.sh status production --verbose
   
   # Rollback if necessary
   ./tools/migrate.sh rollback production <migration_id>
   ```

### Log Locations

- Migration logs: `logs/migration.log`
- Dashboard logs: `logs/dashboard.log`
- Error logs: `logs/error.log`
- Audit logs: `logs/audit.log`

## 🔄 Integration with Existing Systems

### CI/CD Pipeline Integration

```yaml
# .github/workflows/migration.yml
name: Database Migration
on:
  push:
    paths:
      - 'migrations/scripts/**'
      
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Validate Migrations
        run: |
          cd migrations
          ./tools/migrate.sh validate scripts/*.sql
          
  deploy-staging:
    needs: validate
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Staging
        run: |
          cd migrations
          ./tools/migrate.sh migrate staging
```

### Monitoring Stack Integration

The migration framework automatically integrates with:

- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards
- **Alertmanager**: Alert notifications
- **Jaeger**: Distributed tracing

## 🎯 Future Enhancements

### Planned Features

- [ ] Multi-database support (MySQL, MongoDB)
- [ ] Advanced rollback strategies
- [ ] Migration dependency graph
- [ ] Automated performance testing
- [ ] Blue-green deployment support
- [ ] Schema drift detection
- [ ] Automated documentation generation

### Contributing

1. Follow the established patterns and conventions
2. Add comprehensive tests for new features
3. Update documentation for any changes
4. Ensure backward compatibility
5. Follow security best practices

## 📞 Support

### Getting Help

- **Documentation**: Check the comprehensive guides
- **Logs**: Review migration and error logs
- **Dashboard**: Use the web interface for real-time status
- **Team**: Contact the data operations team

### Emergency Contacts

- **Data Operations Lead**: dataops-lead@company.com
- **Database Administrator**: dba@company.com
- **On-call Engineer**: oncall@company.com

---

**Database Migration Framework v1.0**  
*Enterprise-grade database migration management with data operations controls*

Built with ❤️ for safe, reliable, and auditable database changes.