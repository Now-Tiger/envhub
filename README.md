# EnvHub Backend

**EnvHub** is a **secure environment variable management platform** that replaces insecure practices like sharing `.env` files via Slack. It provides encrypted storage for API keys, database credentials, and secrets across multiple environments (dev/staging/prod). Developers use a CLI tool to inject secrets directly into their applications at runtime, ensuring sensitive data never touches disk in plaintext.

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional)

## Quick Start

```bash
# Start services
make up

# View logs
make logs

# Stop services
make down
```

## Development

```bash
# Install dependencies
go mod download

# Run tests
make test

# Connect to database
make psql
```

## Project Structure

```
/envhub-backend
  /cmd/api              # API server entry point
  /internal             # Private application code
  /pkg                  # Public libraries
  /migrations           # Database migrations
  /config               # Configuration files
```

## Environment Variables

Copy `.env.example` to `.env` and configure as needed.
