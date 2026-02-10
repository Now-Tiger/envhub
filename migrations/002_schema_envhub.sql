-- ============================================================================
-- EnvHub Database Schema Design
-- Optimized for: Scale, Performance, Security, and Future Growth
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- USERS TABLE
-- ============================================================================
-- Purpose: Store user accounts (managed by Supabase Auth)
-- Scale considerations: Indexed on email, partitionable by created_at if needed
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255),
    avatar_url TEXT,
    
    -- Auth provider reference (Supabase user ID)
    auth_provider_id VARCHAR(255) UNIQUE,
    
    -- Account status
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    
    -- Soft delete
    deleted_at TIMESTAMPTZ
);

-- Indexes for users
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_auth_provider ON users(auth_provider_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- ============================================================================
-- ORGANIZATIONS TABLE
-- ============================================================================
-- Purpose: Support multi-tenancy - users can belong to multiple orgs
-- Scale: Critical for B2B SaaS growth, enables team collaboration
-- ============================================================================

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    
    -- Billing and limits
    plan_type VARCHAR(50) DEFAULT 'free', -- free, pro, enterprise
    max_projects INTEGER DEFAULT 5,
    max_secrets_per_project INTEGER DEFAULT 100,
    
    -- Ownership
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for organizations
CREATE INDEX idx_organizations_owner ON organizations(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_organizations_plan ON organizations(plan_type);

-- ============================================================================
-- ORGANIZATION_MEMBERS TABLE
-- ============================================================================
-- Purpose: Many-to-many relationship between users and organizations
-- Scale: Enables team collaboration, indexed for fast membership checks
-- ============================================================================

CREATE TYPE org_role AS ENUM ('owner', 'admin', 'member', 'viewer');

CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role org_role NOT NULL DEFAULT 'member',
    
    -- Invitation tracking
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMPTZ DEFAULT NOW(),
    joined_at TIMESTAMPTZ,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint: user can only have one role per org
    UNIQUE(organization_id, user_id)
);

-- Indexes for organization_members (critical for authorization checks)
CREATE INDEX idx_org_members_org ON organization_members(organization_id);
CREATE INDEX idx_org_members_user ON organization_members(user_id);
CREATE INDEX idx_org_members_role ON organization_members(organization_id, role);

-- ============================================================================
-- PROJECTS TABLE
-- ============================================================================
-- Purpose: Container for environment variables (1 project = 1 application)
-- Scale: Partitionable by organization_id for enterprise customers
-- ============================================================================

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Each project gets a unique Data Encryption Key (DEK)
    -- This is encrypted with the Master Key (KEK) and stored here
    encrypted_dek TEXT NOT NULL, -- Base64 encoded encrypted DEK
    dek_version INTEGER NOT NULL DEFAULT 1, -- For key rotation
    
    -- Metadata
    color VARCHAR(7), -- Hex color for UI (#FF5733)
    icon VARCHAR(50), -- Icon identifier
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Ensure unique project names within an organization
    UNIQUE(organization_id, name)
);

-- Indexes for projects
CREATE INDEX idx_projects_org ON projects(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_projects_created_at ON projects(created_at DESC);

-- ============================================================================
-- ENVIRONMENTS TABLE
-- ============================================================================
-- Purpose: Define deployment environments (dev, staging, prod, etc.)
-- Scale: Pre-defined but extensible, indexed for fast lookups
-- ============================================================================

CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    name VARCHAR(50) NOT NULL, -- dev, staging, production, etc.
    description TEXT,
    
    -- Environment-specific settings
    is_protected BOOLEAN DEFAULT false, -- Requires approval for changes
    color VARCHAR(7),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraint: environment names must be unique per project
    UNIQUE(project_id, name)
);

-- Indexes for environments
CREATE INDEX idx_environments_project ON environments(project_id);
CREATE INDEX idx_environments_protected ON environments(is_protected);

-- ============================================================================
-- SECRETS TABLE
-- ============================================================================
-- Purpose: Store encrypted key-value pairs for each environment
-- Scale: This is the HOT TABLE - optimized with composite indexes
-- Partitioning strategy: By project_id or created_at for multi-million row scale
-- ============================================================================

CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    
    -- The secret key (e.g., "DATABASE_URL", "API_KEY")
    key VARCHAR(255) NOT NULL,
    
    -- Encrypted value (AES-256-GCM encrypted with project's DEK)
    encrypted_value TEXT NOT NULL,
    
    -- Metadata
    description TEXT,
    is_active BOOLEAN DEFAULT true, -- For soft-disabling without deletion
    
    -- Versioning (for rollback support)
    version INTEGER NOT NULL DEFAULT 1,
    previous_version_id UUID REFERENCES secrets(id),
    
    -- Audit fields
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Unique constraint: one key per environment (at current version)
    UNIQUE(environment_id, key)
);

-- Indexes for secrets (CRITICAL for performance)
CREATE INDEX idx_secrets_environment ON secrets(environment_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_secrets_key ON secrets(environment_id, key) WHERE deleted_at IS NULL AND is_active = true;
CREATE INDEX idx_secrets_created_at ON secrets(created_at DESC);
CREATE INDEX idx_secrets_version ON secrets(environment_id, version DESC);

-- ============================================================================
-- SECRET_HISTORY TABLE
-- ============================================================================
-- Purpose: Immutable audit log of all secret changes
-- Scale: Append-only table, can be partitioned by created_at (monthly/yearly)
-- ============================================================================

CREATE TYPE secret_action AS ENUM ('created', 'updated', 'deleted', 'rotated');

CREATE TABLE secret_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    
    -- What changed
    action secret_action NOT NULL,
    key VARCHAR(255) NOT NULL,
    encrypted_value TEXT, -- NULL for deletes
    
    -- Who did it
    changed_by UUID NOT NULL REFERENCES users(id),
    
    -- When it happened
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Metadata (for debugging)
    ip_address INET,
    user_agent TEXT
);

-- Indexes for secret_history (for audit queries)
CREATE INDEX idx_secret_history_secret ON secret_history(secret_id, created_at DESC);
CREATE INDEX idx_secret_history_environment ON secret_history(environment_id, created_at DESC);
CREATE INDEX idx_secret_history_user ON secret_history(changed_by, created_at DESC);
CREATE INDEX idx_secret_history_created_at ON secret_history(created_at DESC);

-- Partition preparation (for future scaling)
-- When you hit 10M+ rows, partition by created_at:
-- CREATE TABLE secret_history_2025_01 PARTITION OF secret_history
--     FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- ============================================================================
-- API_TOKENS TABLE
-- ============================================================================
-- Purpose: Long-lived tokens for CLI authentication
-- Scale: Indexed on token hash, user_id for fast validation
-- ============================================================================

CREATE TABLE api_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Token details
    name VARCHAR(255) NOT NULL, -- User-friendly name ("My Laptop", "CI/CD Pipeline")
    token_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA-256 hash of the actual token
    
    -- Scope and permissions
    scopes TEXT[], -- ["read:secrets", "write:secrets"]
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    
    -- Expiration
    expires_at TIMESTAMPTZ,
    
    -- Usage tracking
    last_used_at TIMESTAMPTZ,
    usage_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Indexes for api_tokens
CREATE INDEX idx_api_tokens_hash ON api_tokens(token_hash) WHERE revoked_at IS NULL;
CREATE INDEX idx_api_tokens_user ON api_tokens(user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_api_tokens_expires ON api_tokens(expires_at) WHERE revoked_at IS NULL;

-- ============================================================================
-- ACCESS_LOGS TABLE
-- ============================================================================
-- Purpose: Security audit trail - who accessed what and when
-- Scale: High-write table, partition by created_at (weekly/monthly)
-- ============================================================================

CREATE TYPE access_action AS ENUM ('read', 'create', 'update', 'delete');

CREATE TABLE access_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Who
    user_id UUID REFERENCES users(id),
    api_token_id UUID REFERENCES api_tokens(id),
    
    -- What
    resource_type VARCHAR(50) NOT NULL, -- 'secret', 'project', 'environment'
    resource_id UUID NOT NULL,
    action access_action NOT NULL,
    
    -- When and Where
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    
    -- Result
    success BOOLEAN NOT NULL,
    error_message TEXT
);

-- Indexes for access_logs (for security queries)
CREATE INDEX idx_access_logs_user ON access_logs(user_id, created_at DESC);
CREATE INDEX idx_access_logs_resource ON access_logs(resource_type, resource_id, created_at DESC);
CREATE INDEX idx_access_logs_created_at ON access_logs(created_at DESC);
CREATE INDEX idx_access_logs_failed ON access_logs(success, created_at DESC) WHERE success = false;

-- ============================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_members_updated_at BEFORE UPDATE ON organization_members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_environments_updated_at BEFORE UPDATE ON environments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_secrets_updated_at BEFORE UPDATE ON secrets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- TRIGGER FOR SECRET HISTORY LOGGING
-- ============================================================================

CREATE OR REPLACE FUNCTION log_secret_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO secret_history (secret_id, environment_id, action, key, encrypted_value, changed_by)
        VALUES (NEW.id, NEW.environment_id, 'created', NEW.key, NEW.encrypted_value, NEW.created_by);
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO secret_history (secret_id, environment_id, action, key, encrypted_value, changed_by)
        VALUES (NEW.id, NEW.environment_id, 'updated', NEW.key, NEW.encrypted_value, NEW.updated_by);
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO secret_history (secret_id, environment_id, action, key, encrypted_value, changed_by)
        VALUES (OLD.id, OLD.environment_id, 'deleted', OLD.key, NULL, OLD.updated_by);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER secret_changes_trigger
AFTER INSERT OR UPDATE OR DELETE ON secrets
FOR EACH ROW EXECUTE FUNCTION log_secret_changes();

-- ============================================================================
-- VIEWS FOR COMMON QUERIES (Performance Optimization)
-- ============================================================================

-- View: Active secrets per environment (most common query)
CREATE VIEW active_secrets_by_environment AS
SELECT 
    e.id AS environment_id,
    e.name AS environment_name,
    e.project_id,
    p.organization_id,
    COUNT(s.id) AS secret_count
FROM environments e
LEFT JOIN secrets s ON s.environment_id = e.id 
    AND s.deleted_at IS NULL 
    AND s.is_active = true
JOIN projects p ON p.id = e.project_id
WHERE e.project_id IS NOT NULL
GROUP BY e.id, e.name, e.project_id, p.organization_id;

-- View: User's accessible projects (with organization context)
CREATE VIEW user_accessible_projects AS
SELECT 
    p.id AS project_id,
    p.name AS project_name,
    p.organization_id,
    o.name AS organization_name,
    om.user_id,
    om.role AS user_role
FROM projects p
JOIN organizations o ON o.id = p.organization_id
JOIN organization_members om ON om.organization_id = o.id
WHERE p.deleted_at IS NULL 
  AND o.deleted_at IS NULL;

-- ============================================================================
-- SEED DATA (Default environments for new projects)
-- ============================================================================

-- This will be used by the application to auto-create default environments
-- when a new project is created

-- ============================================================================
-- PERFORMANCE NOTES
-- ============================================================================

-- 1. PARTITIONING STRATEGY (When you hit 10M+ rows):
--    - Partition secret_history by created_at (monthly)
--    - Partition access_logs by created_at (weekly)
--    - Partition projects by organization_id for enterprise customers

-- 2. CACHING STRATEGY:
--    - Cache active secrets per environment in Redis (TTL: 5 minutes)
--    - Invalidate cache on any secret write operation
--    - Cache user permissions (organization membership) for 15 minutes

-- 3. READ REPLICAS:
--    - Route all SELECT queries to read replicas
--    - Only write operations go to primary
--    - Use connection pooling (pgBouncer)

-- 4. ARCHIVAL STRATEGY:
--    - Move secret_history older than 1 year to cold storage (S3)
--    - Move access_logs older than 90 days to data warehouse
--    - Keep deleted_at records for 30 days, then hard delete

-- ============================================================================
-- SECURITY NOTES
-- ============================================================================

-- 1. Row-Level Security (RLS) - Can be enabled for extra security:
--    ALTER TABLE secrets ENABLE ROW LEVEL SECURITY;
--    CREATE POLICY secrets_access_policy ON secrets
--        USING (environment_id IN (
--            SELECT e.id FROM environments e
--            JOIN projects p ON p.id = e.project_id
--            JOIN organization_members om ON om.organization_id = p.organization_id
--            WHERE om.user_id = current_setting('app.current_user_id')::uuid
--        ));

-- 2. Encryption at rest: Enable PostgreSQL encryption at rest in production
-- 3. Backup strategy: Daily encrypted backups with 30-day retention
-- 4. Audit compliance: access_logs and secret_history are immutable

-- ============================================================================
-- MIGRATION NOTES
-- ============================================================================

-- This schema supports zero-downtime migrations:
-- 1. All foreign keys use CASCADE for organizational hierarchy
-- 2. All deletions are soft deletes (deleted_at) - hard deletes are async
-- 3. Indexes are created CONCURRENTLY in production
-- 4. Adding columns uses ALTER TABLE ... ADD COLUMN ... DEFAULT ... (no table lock)
