# EnvHub CLI Documentation

This document describes the CLI integration flow for EnvHub.

## Authentication Flow

The EnvHub CLI uses long-lived tokens (30-day expiry) for authentication. Here's how to set up CLI access:

### Step 1: Generate CLI Token

Use the CLI login endpoint to generate a token:

```bash
# Using JWT token from web login
curl -X POST https://api.envhub.dev/api/v1/auth/cli-login \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"project_id": "550e8400-e29b-41d4-a716-446655440000"}'
```

**Response:**
```json
{
  "token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "expires_at": "2024-02-15T10:30:00Z"
}
```

### Step 2: Configure CLI

Store the token in your environment:

```bash
# Bash/Zsh
export ENVHUB_TOKEN="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"

# Fish
set -x ENVHUB_TOKEN "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"

# Windows (PowerShell)
$env:ENVHUB_TOKEN="a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
```

### Step 3: Fetch Secrets

Retrieve secrets for a specific project and environment:

```bash
curl -X GET "https://api.envhub.dev/api/v1/cli/secrets/my-project/production" \
  -H "Authorization: Bearer $ENVHUB_TOKEN"
```

**Response:**
```json
{
  "project": "my-project",
  "environment": "production",
  "secrets": {
    "DATABASE_URL": "postgresql://user:pass@localhost:5432/db",
    "API_KEY": "sk-live-xxxxxxxx",
    "STRIPE_SECRET": "sk_live_xxxxxxxx"
  },
  "retrieved_at": "2024-01-15T10:30:00Z",
  "version": "1.0"
}
```

### Step 4: Inject into Process

Use the secrets with your application:

```bash
# Export to environment and run command
export $(curl -s -H "Authorization: Bearer $ENVHUB_TOKEN" \
  "https://api.envhub.dev/api/v1/cli/secrets/my-project/production" | \
  jq -r '.secrets | to_entries | .[] | "\(.key)=\(.value)"')

# Run your application
./my-app
```

## Caching

The CLI endpoint includes caching headers for performance optimization:

- **Cache-Control**: `private, max-age=60, must-revalidate`
- **X-EnvHub-Cache**: `MISS` (or `HIT` for cached responses)

For conditional requests with ETags:

```bash
# First request - gets ETag
curl -v "https://api.envhub.dev/api/v1/cli/secrets/my-project/production" \
  -H "Authorization: Bearer $ENVHUB_TOKEN" \
  2>&1 | grep -i ETag

# Subsequent requests - use If-None-Match
curl -v "https://api.envhub.dev/api/v1/cli/secrets/my-project/production" \
  -H "Authorization: Bearer $ENVHUB_TOKEN" \
  -H "If-None-Match: \"<etag-value>\"" \
  2>&1
```

## Example: Using with Docker

```dockerfile
# In your Dockerfile
RUN pip install envhub-cli

# In your entrypoint script
#!/bin/bash
source <(envhub pull my-project production)
exec "$@"
```

## Error Handling

| Status Code | Meaning | Resolution |
|-------------|---------|-------------|
| 401 | Invalid/expired token | Generate a new token via CLI login |
| 404 | Project or environment not found | Verify project name and environment |
| 429 | Rate limit exceeded | Wait and retry |

## Security Notes

1. **Never commit tokens** to version control
2. **Rotate tokens regularly** via the dashboard
3. **Use environment-specific tokens** for production
4. **Set appropriate token expiry** based on security requirements
