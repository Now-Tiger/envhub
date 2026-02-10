-- ============================================================================
-- EnvHub Seed Data
-- This file populates the database with sample data for testing
-- ============================================================================

-- Insert test users
INSERT INTO users (id, email, full_name, auth_provider_id, is_active, email_verified, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'john.doe@example.com', 'John Doe', 'auth0|507f1f77bcf86cd799439011', true, true, NOW() - INTERVAL '30 days'),
('550e8400-e29b-41d4-a716-446655440002', 'jane.smith@example.com', 'Jane Smith', 'auth0|507f1f77bcf86cd799439012', true, true, NOW() - INTERVAL '25 days'),
('550e8400-e29b-41d4-a716-446655440003', 'bob.johnson@example.com', 'Bob Johnson', 'auth0|507f1f77bcf86cd799439013', true, true, NOW() - INTERVAL '20 days');

-- Insert test organizations
INSERT INTO organizations (id, name, slug, plan_type, max_projects, max_secrets_per_project, owner_id, created_at) VALUES
('660e8400-e29b-41d4-a716-446655440001', 'Acme Corporation', 'acme-corp', 'enterprise', 50, 500, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '30 days'),
('660e8400-e29b-41d4-a716-446655440002', 'StartupXYZ', 'startup-xyz', 'pro', 20, 200, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '25 days'),
('660e8400-e29b-41d4-a716-446655440003', 'Freelance Dev', 'freelance-dev', 'free', 5, 100, '550e8400-e29b-41d4-a716-446655440003', NOW() - INTERVAL '20 days');

-- Insert organization members
INSERT INTO organization_members (id, organization_id, user_id, role, invited_by, invited_at, joined_at) VALUES
-- Acme Corporation members
('770e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'owner', NULL, NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
('770e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'admin', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '28 days', NOW() - INTERVAL '28 days'),
('770e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440003', 'member', '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '25 days', NOW() - INTERVAL '25 days'),
-- StartupXYZ members
('770e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', 'owner', NULL, NOW() - INTERVAL '25 days', NOW() - INTERVAL '25 days'),
-- Freelance Dev members
('770e8400-e29b-41d4-a716-446655440005', '660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440003', 'owner', NULL, NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days');

-- Insert test projects (with dummy encrypted DEKs - in production, these would be real encrypted keys)
INSERT INTO projects (id, organization_id, name, description, encrypted_dek, dek_version, color, created_at) VALUES
('880e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440001', 'Main API', 'Primary backend API service', 'ENC_DEK_BASE64_PLACEHOLDER_001', 1, '#3B82F6', NOW() - INTERVAL '29 days'),
('880e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', 'Mobile App', 'iOS and Android mobile application', 'ENC_DEK_BASE64_PLACEHOLDER_002', 1, '#10B981', NOW() - INTERVAL '27 days'),
('880e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440001', 'Analytics Service', 'Data analytics and reporting', 'ENC_DEK_BASE64_PLACEHOLDER_003', 1, '#F59E0B', NOW() - INTERVAL '25 days'),
('880e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440002', 'Web Platform', 'Main web application', 'ENC_DEK_BASE64_PLACEHOLDER_004', 1, '#8B5CF6', NOW() - INTERVAL '24 days'),
('880e8400-e29b-41d4-a716-446655440005', '660e8400-e29b-41d4-a716-446655440003', 'Portfolio Site', 'Personal portfolio website', 'ENC_DEK_BASE64_PLACEHOLDER_005', 1, '#EC4899', NOW() - INTERVAL '20 days');

-- Insert environments for each project
INSERT INTO environments (id, project_id, name, description, is_protected, color, created_at) VALUES
-- Main API environments
('990e8400-e29b-41d4-a716-446655440001', '880e8400-e29b-41d4-a716-446655440001', 'development', 'Development environment', false, '#3B82F6', NOW() - INTERVAL '29 days'),
('990e8400-e29b-41d4-a716-446655440002', '880e8400-e29b-41d4-a716-446655440001', 'staging', 'Staging environment', true, '#F59E0B', NOW() - INTERVAL '29 days'),
('990e8400-e29b-41d4-a716-446655440003', '880e8400-e29b-41d4-a716-446655440001', 'production', 'Production environment', true, '#EF4444', NOW() - INTERVAL '29 days'),
-- Mobile App environments
('990e8400-e29b-41d4-a716-446655440004', '880e8400-e29b-41d4-a716-446655440002', 'development', 'Development environment', false, '#3B82F6', NOW() - INTERVAL '27 days'),
('990e8400-e29b-41d4-a716-446655440005', '880e8400-e29b-41d4-a716-446655440002', 'production', 'Production environment', true, '#EF4444', NOW() - INTERVAL '27 days'),
-- Analytics Service environments
('990e8400-e29b-41d4-a716-446655440006', '880e8400-e29b-41d4-a716-446655440003', 'development', 'Development environment', false, '#3B82F6', NOW() - INTERVAL '25 days'),
('990e8400-e29b-41d4-a716-446655440007', '880e8400-e29b-41d4-a716-446655440003', 'production', 'Production environment', true, '#EF4444', NOW() - INTERVAL '25 days'),
-- Web Platform environments
('990e8400-e29b-41d4-a716-446655440008', '880e8400-e29b-41d4-a716-446655440004', 'development', 'Development environment', false, '#3B82F6', NOW() - INTERVAL '24 days'),
('990e8400-e29b-41d4-a716-446655440009', '880e8400-e29b-41d4-a716-446655440004', 'staging', 'Staging environment', true, '#F59E0B', NOW() - INTERVAL '24 days'),
('990e8400-e29b-41d4-a716-446655440010', '880e8400-e29b-41d4-a716-446655440004', 'production', 'Production environment', true, '#EF4444', NOW() - INTERVAL '24 days'),
-- Portfolio Site environments
('990e8400-e29b-41d4-a716-446655440011', '880e8400-e29b-41d4-a716-446655440005', 'production', 'Production environment', false, '#EF4444', NOW() - INTERVAL '20 days');

-- Insert sample secrets (with dummy encrypted values)
-- Note: In production, these would be encrypted with AES-256-GCM using the project's DEK
INSERT INTO secrets (id, environment_id, key, encrypted_value, description, is_active, version, created_by, created_at) VALUES
-- Main API - Development
('aa0e8400-e29b-41d4-a716-446655440001', '990e8400-e29b-41d4-a716-446655440001', 'DATABASE_URL', 'ENC_postgresql://dev_user:dev_pass@localhost:5432/dev_db', 'Development database connection string', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '29 days'),
('aa0e8400-e29b-41d4-a716-446655440002', '990e8400-e29b-41d4-a716-446655440001', 'REDIS_URL', 'ENC_redis://localhost:6379', 'Redis cache connection', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '29 days'),
('aa0e8400-e29b-41d4-a716-446655440003', '990e8400-e29b-41d4-a716-446655440001', 'JWT_SECRET', 'ENC_dev_secret_key_12345', 'JWT signing secret', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '29 days'),
('aa0e8400-e29b-41d4-a716-446655440004', '990e8400-e29b-41d4-a716-446655440001', 'AWS_ACCESS_KEY_ID', 'ENC_AKIAIOSFODNN7EXAMPLE', 'AWS access key', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '28 days'),
('aa0e8400-e29b-41d4-a716-446655440005', '990e8400-e29b-41d4-a716-446655440001', 'AWS_SECRET_ACCESS_KEY', 'ENC_wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY', 'AWS secret key', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '28 days'),
-- Main API - Production
('aa0e8400-e29b-41d4-a716-446655440006', '990e8400-e29b-41d4-a716-446655440003', 'DATABASE_URL', 'ENC_postgresql://prod_user:prod_pass@prod-db.example.com:5432/prod_db', 'Production database connection string', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '29 days'),
('aa0e8400-e29b-41d4-a716-446655440007', '990e8400-e29b-41d4-a716-446655440003', 'REDIS_URL', 'ENC_redis://prod-redis.example.com:6379', 'Redis cache connection', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '29 days'),
('aa0e8400-e29b-41d4-a716-446655440008', '990e8400-e29b-41d4-a716-446655440003', 'JWT_SECRET', 'ENC_super_secure_prod_key_xyz789', 'JWT signing secret', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '29 days'),
('aa0e8400-e29b-41d4-a716-446655440009', '990e8400-e29b-41d4-a716-446655440003', 'STRIPE_SECRET_KEY', 'ENC_sk_live_51HzXXXXXXXXXXXXXXXXXXXX', 'Stripe payment secret', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '27 days'),
('aa0e8400-e29b-41d4-a716-446655440010', '990e8400-e29b-41d4-a716-446655440003', 'SENDGRID_API_KEY', 'ENC_SG.XXXXXXXXXXXXXXXXXXXX', 'SendGrid email service', true, 1, '550e8400-e29b-41d4-a716-446655440001', NOW() - INTERVAL '26 days'),
-- Mobile App - Development
('aa0e8400-e29b-41d4-a716-446655440011', '990e8400-e29b-41d4-a716-446655440004', 'API_BASE_URL', 'ENC_http://localhost:8080/api/v1', 'Backend API endpoint', true, 1, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '27 days'),
('aa0e8400-e29b-41d4-a716-446655440012', '990e8400-e29b-41d4-a716-446655440004', 'GOOGLE_MAPS_API_KEY', 'ENC_AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXX', 'Google Maps integration', true, 1, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '27 days'),
-- Mobile App - Production
('aa0e8400-e29b-41d4-a716-446655440013', '990e8400-e29b-41d4-a716-446655440005', 'API_BASE_URL', 'ENC_https://api.example.com/v1', 'Backend API endpoint', true, 1, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '27 days'),
('aa0e8400-e29b-41d4-a716-446655440014', '990e8400-e29b-41d4-a716-446655440005', 'GOOGLE_MAPS_API_KEY', 'ENC_AIzaSyPRODXXXXXXXXXXXXXXXXXXXXXX', 'Google Maps integration', true, 1, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '27 days'),
('aa0e8400-e29b-41d4-a716-446655440015', '990e8400-e29b-41d4-a716-446655440005', 'FIREBASE_CONFIG', 'ENC_{"apiKey":"AIzaSyXXX","authDomain":"xxx"}', 'Firebase configuration', true, 1, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '26 days'),
-- Web Platform - Development
('aa0e8400-e29b-41d4-a716-446655440016', '990e8400-e29b-41d4-a716-446655440008', 'NEXT_PUBLIC_API_URL', 'ENC_http://localhost:3000', 'Next.js API URL', true, 1, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '24 days'),
('aa0e8400-e29b-41d4-a716-446655440017', '990e8400-e29b-41d4-a716-446655440008', 'DATABASE_URL', 'ENC_postgresql://web_dev:password@localhost:5432/web_dev', 'Development database', true, 1, '550e8400-e29b-41d4-a716-446655440002', NOW() - INTERVAL '24 days');

-- Insert API tokens (for CLI authentication)
INSERT INTO api_tokens (id, user_id, name, token_hash, scopes, organization_id, expires_at, last_used_at, created_at) VALUES
('bb0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'Johns Laptop', 'sha256_hash_of_token_12345abcdef', ARRAY['read:secrets', 'write:secrets'], '660e8400-e29b-41d4-a716-446655440001', NOW() + INTERVAL '90 days', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '15 days'),
('bb0e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'CI/CD Pipeline', 'sha256_hash_of_token_67890ghijkl', ARRAY['read:secrets'], '660e8400-e29b-41d4-a716-446655440001', NOW() + INTERVAL '365 days', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '10 days'),
('bb0e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', 'Janes MacBook', 'sha256_hash_of_token_mnopqr', ARRAY['read:secrets', 'write:secrets'], '660e8400-e29b-41d4-a716-446655440002', NOW() + INTERVAL '90 days', NOW() - INTERVAL '1 day', NOW() - INTERVAL '8 days');

-- Insert some access logs (for audit trail demonstration)
INSERT INTO access_logs (id, user_id, resource_type, resource_id, action, created_at, ip_address, user_agent, success) VALUES
('cc0e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'secret', 'aa0e8400-e29b-41d4-a716-446655440001', 'read', NOW() - INTERVAL '2 hours', '192.168.1.100', 'EnvHub-CLI/1.0.0', true),
('cc0e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'secret', 'aa0e8400-e29b-41d4-a716-446655440006', 'read', NOW() - INTERVAL '1 hour', '192.168.1.100', 'EnvHub-CLI/1.0.0', true),
('cc0e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', 'secret', 'aa0e8400-e29b-41d4-a716-446655440011', 'update', NOW() - INTERVAL '3 hours', '10.0.0.50', 'Mozilla/5.0 (Dashboard)', true),
('cc0e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', 'project', '880e8400-e29b-41d4-a716-446655440001', 'read', NOW() - INTERVAL '30 minutes', '192.168.1.100', 'Mozilla/5.0 (Dashboard)', true),
('cc0e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', 'secret', 'aa0e8400-e29b-41d4-a716-446655440016', 'read', NOW() - INTERVAL '15 minutes', '172.16.0.20', 'EnvHub-CLI/1.0.0', false);

-- Summary output
DO $$
DECLARE
    user_count INTEGER;
    org_count INTEGER;
    project_count INTEGER;
    env_count INTEGER;
    secret_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO org_count FROM organizations;
    SELECT COUNT(*) INTO project_count FROM projects;
    SELECT COUNT(*) INTO env_count FROM environments;
    SELECT COUNT(*) INTO secret_count FROM secrets;
    
    RAISE NOTICE '';
    RAISE NOTICE '=== Seed Data Summary ===';
    RAISE NOTICE 'Users created: %', user_count;
    RAISE NOTICE 'Organizations created: %', org_count;
    RAISE NOTICE 'Projects created: %', project_count;
    RAISE NOTICE 'Environments created: %', env_count;
    RAISE NOTICE 'Secrets created: %', secret_count;
    RAISE NOTICE '========================';
    RAISE NOTICE '';
END $$;
