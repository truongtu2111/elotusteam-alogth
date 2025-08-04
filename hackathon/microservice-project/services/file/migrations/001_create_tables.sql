-- File Service Database Schema
-- This file contains the SQL schema for the file service

-- Create extension for UUID generation if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Files table
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    path TEXT NOT NULL,
    url TEXT,
    checksum VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Upload sessions table
CREATE TABLE IF NOT EXISTS upload_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    filename VARCHAR(255) NOT NULL,
    total_size BIGINT NOT NULL,
    uploaded_size BIGINT NOT NULL DEFAULT 0,
    chunk_size INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- File shares table
CREATE TABLE IF NOT EXISTS file_shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    shared_by UUID NOT NULL,
    shared_with UUID,
    token VARCHAR(255) UNIQUE,
    permissions TEXT[] NOT NULL DEFAULT '{"read"}',
    expires_at TIMESTAMP WITH TIME ZONE,
    access_count INTEGER NOT NULL DEFAULT 0,
    max_access_count INTEGER,
    password_hash VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- File versions table
CREATE TABLE IF NOT EXISTS file_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    size BIGINT NOT NULL,
    path TEXT NOT NULL,
    url TEXT,
    checksum VARCHAR(64) NOT NULL,
    changes_description TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Image variants table
CREATE TABLE IF NOT EXISTS image_variants (
    id VARCHAR(255) PRIMARY KEY, -- Composite key: file_id + variant_type
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    variant_type VARCHAR(50) NOT NULL, -- thumbnail, small, medium, large, etc.
    width INTEGER,
    height INTEGER,
    format VARCHAR(10) NOT NULL, -- jpg, png, webp, etc.
    quality INTEGER,
    size BIGINT NOT NULL,
    path TEXT NOT NULL,
    url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_checksum ON files(checksum);
CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files(deleted_at);

CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_id ON upload_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_status ON upload_sessions(status);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_expires_at ON upload_sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_file_shares_file_id ON file_shares(file_id);
CREATE INDEX IF NOT EXISTS idx_file_shares_shared_by ON file_shares(shared_by);
CREATE INDEX IF NOT EXISTS idx_file_shares_shared_with ON file_shares(shared_with);
CREATE INDEX IF NOT EXISTS idx_file_shares_token ON file_shares(token);
CREATE INDEX IF NOT EXISTS idx_file_shares_expires_at ON file_shares(expires_at);
CREATE INDEX IF NOT EXISTS idx_file_shares_deleted_at ON file_shares(deleted_at);

CREATE INDEX IF NOT EXISTS idx_file_versions_file_id ON file_versions(file_id);
CREATE INDEX IF NOT EXISTS idx_file_versions_version_number ON file_versions(file_id, version_number);
CREATE INDEX IF NOT EXISTS idx_file_versions_deleted_at ON file_versions(deleted_at);

CREATE INDEX IF NOT EXISTS idx_image_variants_file_id ON image_variants(file_id);
CREATE INDEX IF NOT EXISTS idx_image_variants_variant_type ON image_variants(variant_type);
CREATE INDEX IF NOT EXISTS idx_image_variants_status ON image_variants(status);

-- Create triggers for updating updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_files_updated_at BEFORE UPDATE ON files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_upload_sessions_updated_at BEFORE UPDATE ON upload_sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_file_shares_updated_at BEFORE UPDATE ON file_shares
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_file_versions_updated_at BEFORE UPDATE ON file_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_image_variants_updated_at BEFORE UPDATE ON image_variants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add constraints
ALTER TABLE file_versions ADD CONSTRAINT unique_file_version UNIQUE (file_id, version_number);
ALTER TABLE image_variants ADD CONSTRAINT unique_file_variant UNIQUE (file_id, variant_type);

-- Add check constraints
ALTER TABLE files ADD CONSTRAINT check_file_size_positive CHECK (size >= 0);
ALTER TABLE upload_sessions ADD CONSTRAINT check_upload_size_positive CHECK (total_size >= 0 AND uploaded_size >= 0);
ALTER TABLE upload_sessions ADD CONSTRAINT check_upload_size_valid CHECK (uploaded_size <= total_size);
ALTER TABLE file_shares ADD CONSTRAINT check_access_count_positive CHECK (access_count >= 0);
ALTER TABLE file_shares ADD CONSTRAINT check_max_access_count_positive CHECK (max_access_count IS NULL OR max_access_count > 0);
ALTER TABLE file_versions ADD CONSTRAINT check_version_number_positive CHECK (version_number > 0);
ALTER TABLE file_versions ADD CONSTRAINT check_version_size_positive CHECK (size >= 0);
ALTER TABLE image_variants ADD CONSTRAINT check_image_dimensions_positive CHECK (width IS NULL OR width > 0);
ALTER TABLE image_variants ADD CONSTRAINT check_image_height_positive CHECK (height IS NULL OR height > 0);
ALTER TABLE image_variants ADD CONSTRAINT check_image_quality_valid CHECK (quality IS NULL OR (quality >= 1 AND quality <= 100));
ALTER TABLE image_variants ADD CONSTRAINT check_image_size_positive CHECK (size >= 0);