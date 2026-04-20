# EnvHub Documentation

Comprehensive technical documentation for the EnvHub platform.

## Overview

EnvHub is a secure secrets management platform for developers and teams. It provides a REST API and CLI for managing environment variables with enterprise-grade encryption.

## Documentation Structure

### Phase 1: Foundation & Core Infrastructure
- [Phase 1: Foundation & Core Infrastructure](phase-1-foundation.md)

**Key Topics:**
- Project structure and architecture
- Database schema design
- Cryptographic foundation (AES-256-GCM)
- Configuration management

### Phase 2: Authentication & Authorization
- [Phase 2: Authentication & Authorization](phase-2-auth.md)

**Key Topics:**
- JWT authentication
- API tokens for CLI
- Role-Based Access Control (RBAC)
- Middleware implementation

### Phase 3: Core API Development
- [Phase 3: Core API Development](phase-3-core-api.md)

**Key Topics:**
- Projects API
- Secrets API with encryption
- Environments API
- Team management API
- Service layer architecture

### Phase 4: CLI Integration
- [Phase 4: CLI Integration](phase-4-cli.md)

**Key Topics:**
- CLI-optimized endpoints
- Authentication patterns
- CI/CD integration examples
- Performance optimization

### Phase 5: Security Hardening
- [Phase 5: Security Hardening](phase-5-security.md)

### Plans & Subscriptions
- [Plans & Subscriptions](plans-subscriptions.md)

### CLI User Guide
- [CLI User Guide](cli-user-guide.md)

**Key Topics:**
- Installation methods
- Command reference
- Configuration
- CI/CD integration
- Troubleshooting

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- 32-byte master encryption key

### Running the Server
```bash
export DB_HOST=localhost
export DB_USER=envhub
export DB_PASSWORD=secret
export DB_NAME=envhub
export JWT_SECRET=your-256-bit-secret
export MASTER_ENCRYPTION_KEY=base64-encoded-32-byte-key

go run cmd/api/main.go
```

### API Endpoints

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/health` | GET | No | Health check |
| `/api/v1/auth/register` | POST | No | Register user |
| `/api/v1/auth/login` | POST | No | Login |
| `/api/v1/auth/me` | GET | JWT | Get current user |
| `/api/v1/projects` | POST | JWT | Create project |
| `/api/v1/projects/:id/secrets` | POST | JWT | Create secret |
| `/api/v1/cli/secrets/:project/:env` | GET | JWT/Token | Get secrets for CLI |

## Security Model

```
┌─────────────────────────────────────────────────────────────┐
│                      EnvHub Security                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐     ┌──────────────┐     ┌───────────┐  │
│  │   Client     │────▶│    API       │────▶│  Database │  │
│  │  (CLI/Web)   │     │   Server     │     │   (Postgres)  │
│  └──────────────┘     └──────────────┘     └───────────┘  │
│         │                     │                     │       │
│         │                ┌────▼────┐                │       │
│         │                │  JWT    │                │       │
│         │                │  Token  │                │       │
│         │                └─────────┘                │       │
│         │                                          │       │
│         │            ┌──────────────┐              │       │
│         │            │   Secrets    │              │       │
│         │            │  (Encrypted) │              │       │
│         │            └──────────────┘              │       │
│         │                                          │       │
│         │  ┌──────────────┐   ┌───────────────┐   │       │
│         └──│  Master Key  │◀──│  Project DEK  │   │       │
│            │   (KEK)     │   │   (Encrypted) │   │       │
│            └──────────────┘   └───────────────┘   │       │
│                                                   │       │
└───────────────────────────────────────────────────┴───────┘
```

## Architecture

### Layer Structure
```
┌────────────────────────────────────────┐
│           Handlers (HTTP)              │
│    internal/handlers/*.go              │
├────────────────────────────────────────┤
│           Services (Business)           │
│    internal/service/*.go               │
├────────────────────────────────────────┤
│           Repository (DB)              │
│    internal/repository/*.go           │
├────────────────────────────────────────┤
│           Database (PostgreSQL)         │
│    migrations/*.sql                    │
└────────────────────────────────────────┘
```

### Encryption Flow
1. Master Key encrypts Project DEK
2. Project DEK encrypts Secrets
3. Only Master Key can decrypt DEK
4. DEK needed to decrypt Secrets

## Database Schema

### Core Tables
- `users` - User accounts
- `organizations` - Multi-tenant organizations
- `organization_members` - User-org relationships
- `projects` - Secret containers
- `environments` - Environment definitions
- `secrets` - Encrypted key-values
- `secret_history` - Audit trail
- `api_tokens` - CLI tokens
- `access_logs` - Security logs

## Contributing

See contributing guidelines for details on code style and commit conventions.
