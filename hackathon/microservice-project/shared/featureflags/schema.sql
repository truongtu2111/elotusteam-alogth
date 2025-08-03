Get -- Feature Flags Database Schema
-- This file contains the SQL schema for the feature flags system

-- Create extension for UUID generation if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Feature flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT false,
    rollout_percentage INTEGER NOT NULL DEFAULT 0 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    conditions JSONB,
    variants JSONB,
    default_variant VARCHAR(255),
    value JSONB,
    environment VARCHAR(100) NOT NULL DEFAULT 'development',
    service VARCHAR(100),
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled ON feature_flags(enabled);
CREATE INDEX IF NOT EXISTS idx_feature_flags_environment ON feature_flags(environment);
CREATE INDEX IF NOT EXISTS idx_feature_flags_service ON feature_flags(service);
CREATE INDEX IF NOT EXISTS idx_feature_flags_tags ON feature_flags USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_feature_flags_expires_at ON feature_flags(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_feature_flags_created_at ON feature_flags(created_at);
CREATE INDEX IF NOT EXISTS idx_feature_flags_updated_at ON feature_flags(updated_at);

-- Feature flag events table for analytics
CREATE TABLE IF NOT EXISTS feature_flag_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    service VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- 'evaluation', 'exposure', 'conversion'
    result BOOLEAN NOT NULL,
    variant VARCHAR(255),
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    duration INTEGER, -- in milliseconds
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255)
);

-- Create indexes for analytics queries
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_flag_id ON feature_flag_events(flag_id);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_user_id ON feature_flag_events(user_id);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_service ON feature_flag_events(service);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_event_type ON feature_flag_events(event_type);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_timestamp ON feature_flag_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_result ON feature_flag_events(result);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_variant ON feature_flag_events(variant);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_flag_timestamp ON feature_flag_events(flag_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_user_timestamp ON feature_flag_events(user_id, timestamp);

-- Feature flag audit log table
CREATE TABLE IF NOT EXISTS feature_flag_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'created', 'updated', 'deleted', 'enabled', 'disabled'
    old_value JSONB,
    new_value JSONB,
    changed_by VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reason TEXT,
    metadata JSONB
);

-- Create indexes for audit log
CREATE INDEX IF NOT EXISTS idx_feature_flag_audit_log_flag_id ON feature_flag_audit_log(flag_id);
CREATE INDEX IF NOT EXISTS idx_feature_flag_audit_log_action ON feature_flag_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_feature_flag_audit_log_changed_by ON feature_flag_audit_log(changed_by);
CREATE INDEX IF NOT EXISTS idx_feature_flag_audit_log_timestamp ON feature_flag_audit_log(timestamp);

-- Feature flag user segments table (for advanced targeting)
CREATE TABLE IF NOT EXISTS feature_flag_user_segments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    conditions JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255),
    updated_by VARCHAR(255)
);

-- Create indexes for user segments
CREATE INDEX IF NOT EXISTS idx_feature_flag_user_segments_name ON feature_flag_user_segments(name);
CREATE INDEX IF NOT EXISTS idx_feature_flag_user_segments_created_at ON feature_flag_user_segments(created_at);

-- Feature flag segment assignments table
CREATE TABLE IF NOT EXISTS feature_flag_segment_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id VARCHAR(255) NOT NULL,
    segment_id UUID NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    rollout_percentage INTEGER NOT NULL DEFAULT 100 CHECK (rollout_percentage >= 0 AND rollout_percentage <= 100),
    variant VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    FOREIGN KEY (flag_id) REFERENCES feature_flags(id) ON DELETE CASCADE,
    FOREIGN KEY (segment_id) REFERENCES feature_flag_user_segments(id) ON DELETE CASCADE,
    UNIQUE(flag_id, segment_id)
);

-- Create indexes for segment assignments
CREATE INDEX IF NOT EXISTS idx_feature_flag_segment_assignments_flag_id ON feature_flag_segment_assignments(flag_id);
CREATE INDEX IF NOT EXISTS idx_feature_flag_segment_assignments_segment_id ON feature_flag_segment_assignments(segment_id);

-- Feature flag cache invalidation table (for distributed cache management)
CREATE TABLE IF NOT EXISTS feature_flag_cache_invalidation (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id VARCHAR(255),
    invalidation_type VARCHAR(50) NOT NULL, -- 'flag', 'all', 'service'
    service VARCHAR(100),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed BOOLEAN NOT NULL DEFAULT false
);

-- Create indexes for cache invalidation
CREATE INDEX IF NOT EXISTS idx_feature_flag_cache_invalidation_flag_id ON feature_flag_cache_invalidation(flag_id);
CREATE INDEX IF NOT EXISTS idx_feature_flag_cache_invalidation_service ON feature_flag_cache_invalidation(service);
CREATE INDEX IF NOT EXISTS idx_feature_flag_cache_invalidation_timestamp ON feature_flag_cache_invalidation(timestamp);
CREATE INDEX IF NOT EXISTS idx_feature_flag_cache_invalidation_processed ON feature_flag_cache_invalidation(processed);

-- Create trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_feature_flags_updated_at 
    BEFORE UPDATE ON feature_flags 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_feature_flag_user_segments_updated_at 
    BEFORE UPDATE ON feature_flag_user_segments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create trigger function for audit logging
CREATE OR REPLACE FUNCTION feature_flag_audit_trigger()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO feature_flag_audit_log (flag_id, action, new_value, changed_by)
        VALUES (NEW.id, 'created', to_jsonb(NEW), COALESCE(NEW.created_by, 'system'));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO feature_flag_audit_log (flag_id, action, old_value, new_value, changed_by)
        VALUES (NEW.id, 'updated', to_jsonb(OLD), to_jsonb(NEW), COALESCE(NEW.updated_by, 'system'));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO feature_flag_audit_log (flag_id, action, old_value, changed_by)
        VALUES (OLD.id, 'deleted', to_jsonb(OLD), 'system');
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Create audit trigger
CREATE TRIGGER feature_flag_audit_trigger
    AFTER INSERT OR UPDATE OR DELETE ON feature_flags
    FOR EACH ROW EXECUTE FUNCTION feature_flag_audit_trigger();

-- Create views for common queries

-- View for active flags
CREATE OR REPLACE VIEW active_feature_flags AS
SELECT *
FROM feature_flags
WHERE enabled = true
  AND (expires_at IS NULL OR expires_at > NOW());

-- View for flag statistics
CREATE OR REPLACE VIEW feature_flag_statistics AS
SELECT 
    f.id,
    f.name,
    f.environment,
    f.service,
    f.enabled,
    f.rollout_percentage,
    COUNT(e.id) as total_evaluations,
    COUNT(CASE WHEN e.result = true THEN 1 END) as enabled_evaluations,
    COUNT(DISTINCT e.user_id) as unique_users,
    MAX(e.timestamp) as last_evaluation
FROM feature_flags f
LEFT JOIN feature_flag_events e ON f.id = e.flag_id
GROUP BY f.id, f.name, f.environment, f.service, f.enabled, f.rollout_percentage;

-- View for recent flag changes
CREATE OR REPLACE VIEW recent_feature_flag_changes AS
SELECT 
    a.flag_id,
    f.name as flag_name,
    a.action,
    a.changed_by,
    a.timestamp,
    a.reason
FROM feature_flag_audit_log a
JOIN feature_flags f ON a.flag_id = f.id
WHERE a.timestamp >= NOW() - INTERVAL '7 days'
ORDER BY a.timestamp DESC;

-- Insert some sample data for testing (optional)
-- Uncomment the following lines if you want sample data

/*
-- Sample feature flags
INSERT INTO feature_flags (id, name, description, enabled, rollout_percentage, environment, service, created_by)
VALUES 
    ('new-dashboard', 'New Dashboard', 'Enable the new dashboard UI', true, 50, 'development', 'frontend', 'admin'),
    ('advanced-analytics', 'Advanced Analytics', 'Enable advanced analytics features', false, 0, 'development', 'analytics', 'admin'),
    ('beta-features', 'Beta Features', 'Enable beta features for testing', true, 10, 'development', 'all', 'admin'),
    ('performance-optimization', 'Performance Optimization', 'Enable performance optimizations', true, 100, 'production', 'backend', 'admin');

-- Sample user segments
INSERT INTO feature_flag_user_segments (name, description, conditions, created_by)
VALUES 
    ('beta-users', 'Beta testing users', '{"attributes": {"user_type": "beta"}}', 'admin'),
    ('premium-users', 'Premium subscription users', '{"attributes": {"subscription": "premium"}}', 'admin'),
    ('internal-users', 'Internal company users', '{"attributes": {"email": {"endsWith": "@company.com"}}}', 'admin');
*/

-- Grant permissions (adjust as needed for your setup)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO feature_flag_service;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO feature_flag_service;

-- Create indexes for performance (additional ones based on usage patterns)
CREATE INDEX IF NOT EXISTS idx_feature_flags_name_environment ON feature_flags(name, environment);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_flag_user ON feature_flag_events(flag_id, user_id);
CREATE INDEX IF NOT EXISTS idx_feature_flag_events_daily_stats ON feature_flag_events(flag_id, DATE(timestamp));

-- Comments for documentation
COMMENT ON TABLE feature_flags IS 'Main table storing feature flag definitions and configurations';
COMMENT ON TABLE feature_flag_events IS 'Table storing all feature flag evaluation events for analytics';
COMMENT ON TABLE feature_flag_audit_log IS 'Audit log for tracking all changes to feature flags';
COMMENT ON TABLE feature_flag_user_segments IS 'User segments for advanced targeting';
COMMENT ON TABLE feature_flag_segment_assignments IS 'Assignments of flags to user segments';
COMMENT ON TABLE feature_flag_cache_invalidation IS 'Cache invalidation events for distributed systems';

COMMENT ON COLUMN feature_flags.rollout_percentage IS 'Percentage of users who should see this flag enabled (0-100)';
COMMENT ON COLUMN feature_flags.conditions IS 'JSON conditions for flag evaluation (user attributes, etc.)';
COMMENT ON COLUMN feature_flags.variants IS 'JSON definition of flag variants for A/B testing';
COMMENT ON COLUMN feature_flags.value IS 'JSON value returned when flag is enabled (for configuration flags)';
COMMENT ON COLUMN feature_flag_events.event_type IS 'Type of event: evaluation, exposure, conversion';
COMMENT ON COLUMN feature_flag_events.duration IS 'Time taken for flag evaluation in milliseconds';