#!/usr/bin/env python3
"""
Data Operations Dashboard
A web-based interface for managing database migrations and approvals

This dashboard provides:
- Migration status overview
- Approval workflow management
- Environment controls
- Migration history and monitoring
- Real-time notifications
"""

import os
import sys
import json
import yaml
import sqlite3
import psycopg2
from datetime import datetime, timedelta
from flask import Flask, render_template, request, jsonify, redirect, url_for, flash, session
from flask_login import LoginManager, UserMixin, login_user, logout_user, login_required, current_user
from werkzeug.security import generate_password_hash, check_password_hash
import subprocess
import logging
from functools import wraps
import smtplib
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
import requests

# Configuration
app = Flask(__name__)
app.secret_key = os.environ.get('SECRET_KEY', 'dev-secret-key-change-in-production')

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Login manager setup
login_manager = LoginManager()
login_manager.init_app(app)
login_manager.login_view = 'login'

# Configuration paths
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
CONFIG_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), 'config')
MIGRATIONS_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), 'scripts')
ROLLBACKS_DIR = os.path.join(os.path.dirname(SCRIPT_DIR), 'rollbacks')

# User model
class User(UserMixin):
    def __init__(self, id, username, email, role, permissions):
        self.id = id
        self.username = username
        self.email = email
        self.role = role
        self.permissions = permissions

# Mock user database (in production, use proper database)
USERS = {
    'admin': {
        'id': 1,
        'username': 'admin',
        'email': 'admin@company.com',
        'password_hash': generate_password_hash('admin123'),
        'role': 'data_ops_admin',
        'permissions': ['approve_all', 'manage_locks', 'view_all', 'emergency_deploy']
    },
    'data_ops_lead': {
        'id': 2,
        'username': 'data_ops_lead',
        'email': 'dataops@company.com',
        'password_hash': generate_password_hash('dataops123'),
        'role': 'data_ops_lead',
        'permissions': ['approve_high_risk', 'manage_locks', 'view_all']
    },
    'developer': {
        'id': 3,
        'username': 'developer',
        'email': 'dev@company.com',
        'password_hash': generate_password_hash('dev123'),
        'role': 'developer',
        'permissions': ['create_migrations', 'view_own']
    }
}

@login_manager.user_loader
def load_user(user_id):
    for username, user_data in USERS.items():
        if user_data['id'] == int(user_id):
            return User(
                user_data['id'],
                user_data['username'],
                user_data['email'],
                user_data['role'],
                user_data['permissions']
            )
    return None

# Permission decorator
def requires_permission(permission):
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            if not current_user.is_authenticated:
                return redirect(url_for('login'))
            if permission not in current_user.permissions:
                flash('Insufficient permissions', 'error')
                return redirect(url_for('dashboard'))
            return f(*args, **kwargs)
        return decorated_function
    return decorator

# Database connection helper
def get_db_connection(environment='development'):
    """Get database connection for specified environment"""
    config_file = os.path.join(CONFIG_DIR, 'environments.yml')
    
    try:
        with open(config_file, 'r') as f:
            config = yaml.safe_load(f)
        
        env_config = config.get(environment, {})
        
        conn = psycopg2.connect(
            host=env_config.get('host', 'localhost'),
            port=env_config.get('port', 5432),
            database=env_config.get('database', 'postgres'),
            user=env_config.get('username', 'postgres'),
            password=os.environ.get(env_config.get('password_env', 'DB_PASSWORD'), '')
        )
        
        return conn
    except Exception as e:
        logger.error(f"Database connection failed: {e}")
        return None

# Migration service
class MigrationService:
    @staticmethod
    def get_migration_status(environment):
        """Get migration status for environment"""
        conn = get_db_connection(environment)
        if not conn:
            return None
        
        try:
            cursor = conn.cursor()
            
            # Get migration summary
            cursor.execute("""
                SELECT 
                    COUNT(*) as total_migrations,
                    COUNT(CASE WHEN status = 'APPLIED' THEN 1 END) as applied_migrations,
                    COUNT(CASE WHEN status = 'ROLLED_BACK' THEN 1 END) as rolled_back_migrations,
                    COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed_migrations,
                    MAX(applied_at) as last_migration_date
                FROM migration_history
            """)
            
            summary = cursor.fetchone()
            
            # Get recent migrations
            cursor.execute("""
                SELECT 
                    version,
                    description,
                    risk_level,
                    status,
                    applied_at,
                    execution_time_ms,
                    applied_by
                FROM migration_history 
                ORDER BY applied_at DESC 
                LIMIT 10
            """)
            
            recent_migrations = cursor.fetchall()
            
            # Get pending approvals
            cursor.execute("""
                SELECT 
                    ma.migration_version,
                    mh.description,
                    mh.risk_level,
                    COUNT(ma.approver_role) as current_approvals
                FROM migration_history mh
                LEFT JOIN migration_approvals ma ON mh.version = ma.migration_version
                WHERE mh.status = 'PENDING_APPROVAL'
                GROUP BY ma.migration_version, mh.description, mh.risk_level
            """)
            
            pending_approvals = cursor.fetchall()
            
            return {
                'summary': {
                    'total': summary[0] or 0,
                    'applied': summary[1] or 0,
                    'rolled_back': summary[2] or 0,
                    'failed': summary[3] or 0,
                    'last_migration': summary[4]
                },
                'recent_migrations': recent_migrations,
                'pending_approvals': pending_approvals
            }
            
        except Exception as e:
            logger.error(f"Error getting migration status: {e}")
            return None
        finally:
            conn.close()
    
    @staticmethod
    def get_pending_migrations(environment):
        """Get list of pending migrations"""
        applied_migrations = set()
        
        conn = get_db_connection(environment)
        if conn:
            try:
                cursor = conn.cursor()
                cursor.execute("SELECT version FROM migration_history WHERE status = 'APPLIED'")
                applied_migrations = {row[0] for row in cursor.fetchall()}
            except Exception as e:
                logger.error(f"Error getting applied migrations: {e}")
            finally:
                conn.close()
        
        # Get available migration files
        pending_migrations = []
        
        try:
            for filename in sorted(os.listdir(MIGRATIONS_DIR)):
                if filename.endswith('.sql'):
                    version = filename[:-4]  # Remove .sql extension
                    
                    if version not in applied_migrations:
                        migration_file = os.path.join(MIGRATIONS_DIR, filename)
                        metadata = MigrationService.get_migration_metadata(migration_file)
                        
                        pending_migrations.append({
                            'version': version,
                            'filename': filename,
                            'description': metadata.get('description', ''),
                            'risk_level': metadata.get('risk_level', 'LOW'),
                            'estimated_duration': metadata.get('estimated_duration', ''),
                            'author': metadata.get('author', ''),
                            'created': metadata.get('created', '')
                        })
        except Exception as e:
            logger.error(f"Error getting pending migrations: {e}")
        
        return pending_migrations
    
    @staticmethod
    def get_migration_metadata(migration_file):
        """Extract metadata from migration file"""
        metadata = {}
        
        try:
            with open(migration_file, 'r') as f:
                for line in f:
                    if line.startswith('-- Description:'):
                        metadata['description'] = line.replace('-- Description:', '').strip()
                    elif line.startswith('-- Risk Level:'):
                        metadata['risk_level'] = line.replace('-- Risk Level:', '').strip()
                    elif line.startswith('-- Estimated Duration:'):
                        metadata['estimated_duration'] = line.replace('-- Estimated Duration:', '').strip()
                    elif line.startswith('-- Author:'):
                        metadata['author'] = line.replace('-- Author:', '').strip()
                    elif line.startswith('-- Created:'):
                        metadata['created'] = line.replace('-- Created:', '').strip()
                    elif not line.startswith('--'):
                        break  # Stop reading when we reach actual SQL
        except Exception as e:
            logger.error(f"Error reading migration metadata: {e}")
        
        return metadata
    
    @staticmethod
    def add_approval(version, approver_role, approver_name, comments, environment='development'):
        """Add approval for migration"""
        conn = get_db_connection(environment)
        if not conn:
            return False
        
        try:
            cursor = conn.cursor()
            cursor.execute("""
                INSERT INTO migration_approvals (migration_version, approver_role, approver_name, comments) 
                VALUES (%s, %s, %s, %s)
                ON CONFLICT (migration_version, approver_role) 
                DO UPDATE SET 
                    approver_name = EXCLUDED.approver_name,
                    approved_at = CURRENT_TIMESTAMP,
                    comments = EXCLUDED.comments
            """, (version, approver_role, approver_name, comments))
            
            conn.commit()
            return True
            
        except Exception as e:
            logger.error(f"Error adding approval: {e}")
            return False
        finally:
            conn.close()
    
    @staticmethod
    def execute_migration(version, environment, force=False):
        """Execute migration using the migrate.sh script"""
        try:
            migrate_script = os.path.join(SCRIPT_DIR, 'migrate.sh')
            
            cmd = [migrate_script, 'migrate', environment]
            if force:
                cmd.append('--force')
            
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)
            
            return {
                'success': result.returncode == 0,
                'output': result.stdout,
                'error': result.stderr
            }
            
        except subprocess.TimeoutExpired:
            return {
                'success': False,
                'output': '',
                'error': 'Migration timed out after 5 minutes'
            }
        except Exception as e:
            return {
                'success': False,
                'output': '',
                'error': str(e)
            }

# Notification service
class NotificationService:
    @staticmethod
    def send_slack_notification(message, channel='#data-ops'):
        """Send Slack notification"""
        webhook_url = os.environ.get('SLACK_WEBHOOK_URL')
        if not webhook_url:
            return False
        
        try:
            payload = {
                'channel': channel,
                'text': message,
                'username': 'Data Ops Dashboard'
            }
            
            response = requests.post(webhook_url, json=payload)
            return response.status_code == 200
            
        except Exception as e:
            logger.error(f"Error sending Slack notification: {e}")
            return False
    
    @staticmethod
    def send_email_notification(to_emails, subject, message):
        """Send email notification"""
        smtp_server = os.environ.get('SMTP_SERVER')
        smtp_port = int(os.environ.get('SMTP_PORT', 587))
        smtp_username = os.environ.get('SMTP_USERNAME')
        smtp_password = os.environ.get('SMTP_PASSWORD')
        
        if not all([smtp_server, smtp_username, smtp_password]):
            return False
        
        try:
            msg = MIMEMultipart()
            msg['From'] = smtp_username
            msg['To'] = ', '.join(to_emails)
            msg['Subject'] = subject
            
            msg.attach(MIMEText(message, 'plain'))
            
            server = smtplib.SMTP(smtp_server, smtp_port)
            server.starttls()
            server.login(smtp_username, smtp_password)
            server.send_message(msg)
            server.quit()
            
            return True
            
        except Exception as e:
            logger.error(f"Error sending email notification: {e}")
            return False

# Routes
@app.route('/login', methods=['GET', 'POST'])
def login():
    if request.method == 'POST':
        username = request.form['username']
        password = request.form['password']
        
        user_data = USERS.get(username)
        if user_data and check_password_hash(user_data['password_hash'], password):
            user = User(
                user_data['id'],
                user_data['username'],
                user_data['email'],
                user_data['role'],
                user_data['permissions']
            )
            login_user(user)
            return redirect(url_for('dashboard'))
        else:
            flash('Invalid username or password', 'error')
    
    return render_template('login.html')

@app.route('/logout')
@login_required
def logout():
    logout_user()
    return redirect(url_for('login'))

@app.route('/')
@login_required
def dashboard():
    environments = ['development', 'staging', 'production']
    environment_status = {}
    
    for env in environments:
        status = MigrationService.get_migration_status(env)
        environment_status[env] = status
    
    return render_template('dashboard.html', 
                         environment_status=environment_status,
                         user=current_user)

@app.route('/migrations/<environment>')
@login_required
def migrations(environment):
    if environment not in ['development', 'staging', 'production']:
        flash('Invalid environment', 'error')
        return redirect(url_for('dashboard'))
    
    status = MigrationService.get_migration_status(environment)
    pending = MigrationService.get_pending_migrations(environment)
    
    return render_template('migrations.html',
                         environment=environment,
                         status=status,
                         pending_migrations=pending,
                         user=current_user)

@app.route('/approve/<version>', methods=['POST'])
@login_required
@requires_permission('approve_high_risk')
def approve_migration(version):
    comments = request.form.get('comments', '')
    environment = request.form.get('environment', 'development')
    
    success = MigrationService.add_approval(
        version, 
        current_user.role, 
        current_user.username, 
        comments,
        environment
    )
    
    if success:
        flash(f'Approval added for migration {version}', 'success')
        
        # Send notification
        NotificationService.send_slack_notification(
            f'Migration {version} approved by {current_user.username} ({current_user.role})'
        )
    else:
        flash('Failed to add approval', 'error')
    
    return redirect(url_for('migrations', environment=environment))

@app.route('/execute/<version>', methods=['POST'])
@login_required
@requires_permission('approve_all')
def execute_migration(version):
    environment = request.form.get('environment', 'development')
    force = request.form.get('force') == 'true'
    
    result = MigrationService.execute_migration(version, environment, force)
    
    if result['success']:
        flash(f'Migration {version} executed successfully', 'success')
        
        # Send notification
        NotificationService.send_slack_notification(
            f'Migration {version} executed in {environment} by {current_user.username}'
        )
    else:
        flash(f'Migration execution failed: {result["error"]}', 'error')
    
    return redirect(url_for('migrations', environment=environment))

@app.route('/api/status/<environment>')
@login_required
def api_status(environment):
    """API endpoint for migration status"""
    status = MigrationService.get_migration_status(environment)
    return jsonify(status)

@app.route('/api/pending/<environment>')
@login_required
def api_pending(environment):
    """API endpoint for pending migrations"""
    pending = MigrationService.get_pending_migrations(environment)
    return jsonify(pending)

# Template creation (basic HTML templates)
def create_templates():
    """Create basic HTML templates"""
    templates_dir = os.path.join(os.path.dirname(__file__), 'templates')
    os.makedirs(templates_dir, exist_ok=True)
    
    # Base template
    base_template = '''
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Data Ops Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
        <div class="container">
            <a class="navbar-brand" href="{{ url_for('dashboard') }}">
                <i class="fas fa-database"></i> Data Ops Dashboard
            </a>
            {% if current_user.is_authenticated %}
            <div class="navbar-nav ms-auto">
                <span class="navbar-text me-3">Welcome, {{ current_user.username }}</span>
                <a class="nav-link" href="{{ url_for('logout') }}">Logout</a>
            </div>
            {% endif %}
        </div>
    </nav>
    
    <div class="container mt-4">
        {% with messages = get_flashed_messages(with_categories=true) %}
            {% if messages %}
                {% for category, message in messages %}
                    <div class="alert alert-{{ 'danger' if category == 'error' else category }} alert-dismissible fade show" role="alert">
                        {{ message }}
                        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
                    </div>
                {% endfor %}
            {% endif %}
        {% endwith %}
        
        {% block content %}{% endblock %}
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    {% block scripts %}{% endblock %}
</body>
</html>
    '''
    
    with open(os.path.join(templates_dir, 'base.html'), 'w') as f:
        f.write(base_template)
    
    # Login template
    login_template = '''
{% extends "base.html" %}

{% block content %}
<div class="row justify-content-center">
    <div class="col-md-6">
        <div class="card">
            <div class="card-header">
                <h4 class="mb-0">Login to Data Ops Dashboard</h4>
            </div>
            <div class="card-body">
                <form method="POST">
                    <div class="mb-3">
                        <label for="username" class="form-label">Username</label>
                        <input type="text" class="form-control" id="username" name="username" required>
                    </div>
                    <div class="mb-3">
                        <label for="password" class="form-label">Password</label>
                        <input type="password" class="form-control" id="password" name="password" required>
                    </div>
                    <button type="submit" class="btn btn-primary">Login</button>
                </form>
                
                <div class="mt-4">
                    <small class="text-muted">
                        Demo accounts:<br>
                        admin / admin123<br>
                        data_ops_lead / dataops123<br>
                        developer / dev123
                    </small>
                </div>
            </div>
        </div>
    </div>
</div>
{% endblock %}
    '''
    
    with open(os.path.join(templates_dir, 'login.html'), 'w') as f:
        f.write(login_template)
    
    # Dashboard template
    dashboard_template = '''
{% extends "base.html" %}

{% block content %}
<h1>Data Operations Dashboard</h1>

<div class="row">
    {% for env, status in environment_status.items() %}
    <div class="col-md-4 mb-4">
        <div class="card">
            <div class="card-header d-flex justify-content-between align-items-center">
                <h5 class="mb-0">{{ env.title() }}</h5>
                <span class="badge bg-{{ 'success' if status and status.summary.failed == 0 else 'warning' }}">
                    {{ 'Healthy' if status and status.summary.failed == 0 else 'Issues' }}
                </span>
            </div>
            <div class="card-body">
                {% if status %}
                <div class="row text-center">
                    <div class="col-6">
                        <h4 class="text-primary">{{ status.summary.applied }}</h4>
                        <small>Applied</small>
                    </div>
                    <div class="col-6">
                        <h4 class="text-warning">{{ status.pending_approvals|length }}</h4>
                        <small>Pending</small>
                    </div>
                </div>
                {% else %}
                <p class="text-muted">Unable to connect to database</p>
                {% endif %}
                <div class="mt-3">
                    <a href="{{ url_for('migrations', environment=env) }}" class="btn btn-outline-primary btn-sm">
                        View Details
                    </a>
                </div>
            </div>
        </div>
    </div>
    {% endfor %}
</div>
{% endblock %}
    '''
    
    with open(os.path.join(templates_dir, 'dashboard.html'), 'w') as f:
        f.write(dashboard_template)
    
    # Migrations template
    migrations_template = '''
{% extends "base.html" %}

{% block content %}
<div class="d-flex justify-content-between align-items-center mb-4">
    <h1>{{ environment.title() }} Migrations</h1>
    <a href="{{ url_for('dashboard') }}" class="btn btn-secondary">Back to Dashboard</a>
</div>

{% if status %}
<div class="row mb-4">
    <div class="col-md-3">
        <div class="card text-center">
            <div class="card-body">
                <h4 class="text-primary">{{ status.summary.applied }}</h4>
                <small>Applied</small>
            </div>
        </div>
    </div>
    <div class="col-md-3">
        <div class="card text-center">
            <div class="card-body">
                <h4 class="text-warning">{{ pending_migrations|length }}</h4>
                <small>Pending</small>
            </div>
        </div>
    </div>
    <div class="col-md-3">
        <div class="card text-center">
            <div class="card-body">
                <h4 class="text-danger">{{ status.summary.failed }}</h4>
                <small>Failed</small>
            </div>
        </div>
    </div>
    <div class="col-md-3">
        <div class="card text-center">
            <div class="card-body">
                <h4 class="text-info">{{ status.summary.rolled_back }}</h4>
                <small>Rolled Back</small>
            </div>
        </div>
    </div>
</div>
{% endif %}

{% if pending_migrations %}
<div class="card mb-4">
    <div class="card-header">
        <h5 class="mb-0">Pending Migrations</h5>
    </div>
    <div class="card-body">
        <div class="table-responsive">
            <table class="table table-striped">
                <thead>
                    <tr>
                        <th>Version</th>
                        <th>Description</th>
                        <th>Risk Level</th>
                        <th>Author</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {% for migration in pending_migrations %}
                    <tr>
                        <td><code>{{ migration.version }}</code></td>
                        <td>{{ migration.description }}</td>
                        <td>
                            <span class="badge bg-{{ 'danger' if migration.risk_level == 'HIGH' else 'warning' if migration.risk_level == 'MEDIUM' else 'success' }}">
                                {{ migration.risk_level }}
                            </span>
                        </td>
                        <td>{{ migration.author }}</td>
                        <td>
                            {% if 'approve_high_risk' in user.permissions %}
                            <button class="btn btn-success btn-sm" data-bs-toggle="modal" data-bs-target="#approveModal" 
                                    data-version="{{ migration.version }}" data-description="{{ migration.description }}">
                                <i class="fas fa-check"></i> Approve
                            </button>
                            {% endif %}
                            {% if 'approve_all' in user.permissions %}
                            <button class="btn btn-primary btn-sm" data-bs-toggle="modal" data-bs-target="#executeModal" 
                                    data-version="{{ migration.version }}" data-description="{{ migration.description }}">
                                <i class="fas fa-play"></i> Execute
                            </button>
                            {% endif %}
                        </td>
                    </tr>
                    {% endfor %}
                </tbody>
            </table>
        </div>
    </div>
</div>
{% endif %}

<!-- Approval Modal -->
<div class="modal fade" id="approveModal" tabindex="-1">
    <div class="modal-dialog">
        <div class="modal-content">
            <form method="POST" id="approveForm">
                <div class="modal-header">
                    <h5 class="modal-title">Approve Migration</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <input type="hidden" name="environment" value="{{ environment }}">
                    <p>Are you sure you want to approve migration <strong id="approveVersion"></strong>?</p>
                    <p class="text-muted" id="approveDescription"></p>
                    <div class="mb-3">
                        <label for="comments" class="form-label">Comments</label>
                        <textarea class="form-control" name="comments" rows="3"></textarea>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="submit" class="btn btn-success">Approve</button>
                </div>
            </form>
        </div>
    </div>
</div>

<!-- Execute Modal -->
<div class="modal fade" id="executeModal" tabindex="-1">
    <div class="modal-dialog">
        <div class="modal-content">
            <form method="POST" id="executeForm">
                <div class="modal-header">
                    <h5 class="modal-title">Execute Migration</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <input type="hidden" name="environment" value="{{ environment }}">
                    <p>Are you sure you want to execute migration <strong id="executeVersion"></strong>?</p>
                    <p class="text-muted" id="executeDescription"></p>
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" name="force" value="true" id="forceExecute">
                        <label class="form-check-label" for="forceExecute">
                            Force execution (skip approval checks)
                        </label>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="submit" class="btn btn-primary">Execute</button>
                </div>
            </form>
        </div>
    </div>
</div>
{% endblock %}

{% block scripts %}
<script>
// Handle modal data
document.addEventListener('DOMContentLoaded', function() {
    const approveModal = document.getElementById('approveModal');
    const executeModal = document.getElementById('executeModal');
    
    approveModal.addEventListener('show.bs.modal', function(event) {
        const button = event.relatedTarget;
        const version = button.getAttribute('data-version');
        const description = button.getAttribute('data-description');
        
        document.getElementById('approveVersion').textContent = version;
        document.getElementById('approveDescription').textContent = description;
        document.getElementById('approveForm').action = '/approve/' + version;
    });
    
    executeModal.addEventListener('show.bs.modal', function(event) {
        const button = event.relatedTarget;
        const version = button.getAttribute('data-version');
        const description = button.getAttribute('data-description');
        
        document.getElementById('executeVersion').textContent = version;
        document.getElementById('executeDescription').textContent = description;
        document.getElementById('executeForm').action = '/execute/' + version;
    });
});
</script>
{% endblock %}
    '''
    
    with open(os.path.join(templates_dir, 'migrations.html'), 'w') as f:
        f.write(migrations_template)

if __name__ == '__main__':
    # Create templates directory and files
    create_templates()
    
    # Run the application
    port = int(os.environ.get('PORT', 5000))
    debug = os.environ.get('DEBUG', 'False').lower() == 'true'
    
    print(f"Starting Data Ops Dashboard on port {port}")
    print("Demo accounts:")
    print("  admin / admin123 (Full access)")
    print("  data_ops_lead / dataops123 (Approval access)")
    print("  developer / dev123 (Read-only access)")
    
    app.run(host='0.0.0.0', port=port, debug=debug)