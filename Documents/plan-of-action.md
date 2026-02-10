# Backend Action Plan for EnvHub

## Phase 1: Foundation & Core Infrastructure (Week 1)

### 1.1 Project Setup

- Initialize Go module with proper versioning
- Set up project structure (monorepo with `/cmd`, `/pkg`, `/internal`)
- Configure environment-based configs (dev, staging, prod)
- Set up `.gitignore` and basic documentation

### 1.2 Database Design

- Design PostgreSQL schema:
  - `users` (id, email, created_at)
  - `projects` (id, name, owner_id, created_at)
  - `secrets` (id, project_id, environment, encrypted_data, encrypted_dek, version, created_at)
  - `project_members` (project_id, user_id, role)
- Write SQL migrations
- Set up **sqlc** for type-safe database operations
- Configure connection pooling

### 1.3 Crypto Foundation

- Implement `pkg/crypto` package:
  - Master key management (KEK)
  - DEK generation per project
  - `Encrypt()` and `Decrypt()` with AES-256-GCM
  - Key rotation logic
- Write comprehensive unit tests for crypto operations

---

## Phase 2: Authentication & Authorization (Week 2)

### 2.1 Auth Integration

- Integrate **Supabase Auth** (or Clerk)
- Build JWT validation middleware
- Implement user registration/login flow
- Store user sessions securely

### 2.2 RBAC System

- Define roles: Owner, Admin, Developer, Viewer
- Implement permission checks middleware
- Create project membership management logic

---

## Phase 3: Core API Development (Week 3-4)

### 3.1 Router & Middleware Setup

- Set up **Chi** router with middleware chain:
  - Request logging
  - CORS (strict policy)
  - Rate limiting
  - JWT validation
  - Recovery from panics

### 3.2 Critical Endpoints

**Projects:**

- `POST /api/v1/projects` - Create project (generates unique DEK)
- `GET /api/v1/projects` - List user's projects
- `GET /api/v1/projects/:id` - Get project details
- `DELETE /api/v1/projects/:id` - Archive project

**Secrets:**

- `POST /api/v1/projects/:id/secrets` - Create/update secrets for environment
  - Validate JSON payload
  - Encrypt with project's DEK
  - Store encrypted DEK + encrypted secrets
- `GET /api/v1/projects/:id/secrets/:env` - Retrieve decrypted secrets
  - Permission check
  - Decrypt DEK with master key
  - Decrypt secrets with DEK
  - Return JSON map

**Team Management:**

- `POST /api/v1/projects/:id/members` - Add team member
- `DELETE /api/v1/projects/:id/members/:user_id` - Remove member

### 3.3 Shared Types

- Create `pkg/types` package for:
  - Request/response DTOs
  - Domain models
  - Error types

---

## Phase 4: CLI Integration Contract (Week 5)

### 4.1 API for CLI

- `POST /api/v1/auth/cli-login` - Generate long-lived CLI token
- `GET /api/v1/cli/secrets/:project/:env` - Optimized endpoint for CLI
  - Returns flat key-value map
  - Includes caching headers

### 4.2 API Documentation

- Generate OpenAPI/Swagger spec
- Document authentication flow for CLI
- Provide example requests/responses

---

## Phase 5: Security Hardening (Week 6)

### 5.1 Security Checklist

- Enable TLS 1.3 on server
- Implement request validation (sanitize inputs)
- Add audit logging for secret access
- Set up secrets versioning/history
- Implement rate limiting per user/IP

### 5.2 Key Rotation Strategy

- Build master key rotation endpoint (admin-only)
- Implement zero-downtime key rotation
- Add monitoring for failed decryption attempts

---

## Technical Decisions Summary

**Must-Haves:**

- **Database:** PostgreSQL with sqlc
- **Router:** Chi (idiomatic, composable)
- **Encryption:** AES-256-GCM from stdlib
- **Auth:** Supabase Auth
- **Deployment:** Docker + docker-compose for local dev

**File Structure:**

```
/envhub-repo
  /cmd/api/main.go           # API server entry
  /internal/
    /handlers/               # HTTP handlers
    /middleware/            # Auth, logging, etc.
    /repository/            # Database operations
    /service/               # Business logic
  /pkg/
    /crypto/                # Shared encryption
    /types/                 # Shared models
  /migrations/              # SQL migrations
  /config/                  # Config files
```

**Estimated Timeline:** 6 weeks for production-ready backend with comprehensive tests and documentation.
