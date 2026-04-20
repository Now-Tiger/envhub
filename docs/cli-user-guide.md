# EnvHub CLI - User Guide

## Overview

EnvHub CLI is a professional command-line tool for managing environment variables and secrets. It provides a fast, developer-friendly interface to interact with the EnvHub API for pulling, pushing, and managing secrets across projects and environments.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/Now-Tiger/envhub.git
cd envhub

# Build the CLI
go build -o envhub ./cmd/cli

# Add to PATH
sudo mv envhub /usr/local/bin/envhub
```

### Using Go Install

```bash
go install github.com/Now-Tiger/envhub/cmd/cli@latest
```

### Using Homebrew (macOS)

```bash
brew install envhub/tap/envhub
```

## Quick Start

### 1. Login to EnvHub

```bash
envhub login
```

This displays instructions for authenticating:

```
Login to EnvHub

1. Start the API server:
   docker-compose up -d

2. Open browser:
   http://localhost:8080

3. Register and create a project

4. Get a CLI token:
   - Login via web UI
   - Use CLI login endpoint or create API token

5. Set the token:
   export ENVUB_TOKEN=<your-token>
```

### 2. Set Your API Token

```bash
# Option 1: Environment variable
export ENVUB_TOKEN=your_token_here

# Option 2: Save to config (recommended)
envhub config set token your_token_here
```

### 3. Pull Secrets

```bash
# Basic usage
envhub pull --project myapp --env production

# Output to .env file
envhub pull -p myapp -e staging --output .env

# Output in JSON format
envhub pull -p myapp -e prod --format json
```

### 4. Push Secrets

```bash
# Push from .env file
envhub push --project myapp --env production --input .env

# Push from JSON file
envhub push -p myapp -e staging --input secrets.json --format json
```

## Commands Reference

### Global Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config` | - | Path to config file | `~/.envhub/config.yaml` |
| `--verbose` | `-v` | Enable verbose output | `false` |
| `--no-color` | - | Disable colored output | `false` |
| `--output` | `-o` | Output in JSON format | `false` |

### login

Display login instructions and authentication help.

```bash
envhub login
envhub login --browser  # Open browser for web login
```

### pull

Pull secrets from EnvHub for a specific project and environment.

```bash
envhub pull [flags]
```

**Flags:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--project` | `-p` | Project name (required) | - |
| `--env` | `-e` | Environment name | `development` |
| `--output` | `-o` | Output file path | stdout |
| `--format` | - | Output format: env, json, yaml, docker | `env` |

**Examples:**

```bash
# Pull from production environment
envhub pull --project myapp --env production

# Pull and save to .env file
envhub pull -p myapp -e staging --output .env

# Pull in JSON format
envhub pull -p myapp -e prod --format json

# Pull to stdout
envhub pull -p myapp
```

**Output Formats:**

- `env` (default): KEY=value format
- `json`: JSON object
- `yaml`: YAML format
- `docker`: Docker-compatible format (export KEY=value)

### push

Push secrets to EnvHub for a specific project and environment.

```bash
envhub push [flags]
```

**Flags:**

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--project` | `-p` | Project name (required) | - |
| `--env` | `-e` | Environment name | `development` |
| `--input` | `-i` | Input file path (required) | - |
| `--format` | - | Input format: env, json | `env` |

**Examples:**

```bash
# Push from .env file
envhub push --project myapp --env production --input .env

# Push from JSON file
envhub push -p myapp -e staging --input secrets.json --format json
```

### list

List all projects accessible to the current user.

```bash
envhub list
envhub list --org myorg  # Filter by organization
envhub list -o           # Output as JSON
```

### whoami

Display information about the currently authenticated user.

```bash
envhub whoami
```

**Example Output:**

```
Logged in as: user@example.com
```

### logout

Clear cached credentials and logout.

```bash
envhub logout
```

### version

Display version information.

```bash
envhub version
```

**Example Output:**

```
EnvHub CLI version 0.1.0
Build: abc123
```

### config

Manage CLI configuration.

```bash
# Show current configuration
envhub config show

# Set a configuration value
envhub config set token your_token_here
envhub config set api_url https://api.envhub.dev

# Open config in editor
envhub config edit
```

## Configuration

### Config File Location

- Linux/macOS: `~/.envhub/config.yaml`
- Windows: `%USERPROFILE%\.envhub\config.yaml`

### Config File Structure

```yaml
api_url: http://localhost:8080
token: your_encrypted_token_here
verbose: false
no_color: false
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ENVUB_TOKEN` | API token for authentication |
| `ENVUB_CONFIG` | Path to config file |
| `ENVUB_API_URL` | API base URL (overrides config) |

## Authentication

### Getting an API Token

1. **Via Web UI:**
   - Login to EnvHub at http://localhost:8080
   - Navigate to Settings → API Tokens
   - Create a new token

2. **Via CLI Login:**
   ```bash
   # Get token from login endpoint
   export ENVUB_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"you@example.com","password":"password"}' | jq -r '.token')
   ```

3. **Via CLI Login Command:**
   ```bash
   # Generate CLI-specific token
   curl -X POST http://localhost:8080/api/v1/auth/cli-login \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"project_id":"project-uuid"}'
   ```

### Token Security

- Tokens are stored encrypted in the config file
- File permissions are set to 0600 (owner read/write only)
- Never commit tokens to version control

## Output Examples

### Standard Output

```bash
$ envhub pull --project myapp --env production
ℹ Pulling secrets from myapp/production...
✓ Secrets written to .env
```

```bash
$ envhub whoami
✓ Logged in as: user@example.com
```

### JSON Output

```bash
$ envhub list -o
{
  "projects": [
    {"id": "uuid", "name": "myapp", " environments": 3},
    {"id": "uuid", "name": "webapp", "environments": 2}
  ]
}
```

### Error Output

```bash
$ envhub pull --project myapp
✗ project name required
Usage: envhub pull --project <name>
```

```bash
$ envhub pull --project myapp --env production
✗ API returned status 401: authentication required
Hint: Set your API token:
  • Export: export ENVUB_TOKEN=your_token
  • Login:  envhub login
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Pull secrets
  run: |
    echo "${{ secrets.ENVUB_TOKEN }}" > token.txt
    envhub pull --project myapp --env production --output .env
  env:
    ENVUB_TOKEN_FILE: token.txt
```

### GitLab CI

```yaml
pull-secrets:
  script:
    - envhub pull --project $CI_PROJECT_NAME --env production --output .env
  variables:
    ENVUB_TOKEN: $ENVUB_TOKEN
```

### Docker

```bash
# Run as Docker container
docker run --rm -it \
  -e ENVUB_TOKEN=$ENVUB_TOKEN \
  envhub/cli pull -p myapp -e production
```

## Troubleshooting

### Common Issues

**"authentication required"**
- Set `ENVUB_TOKEN` environment variable
- Run `envhub login` for instructions

**"project name required"**
- Ensure `--project` or `-p` flag is provided

**"API returned status 404"**
- Verify the project exists
- Check the environment name is correct

**"connection refused"**
- Ensure the API server is running
- Check `api_url` in config

### Verbose Mode

Enable verbose logging for debugging:

```bash
envhub -v pull --project myapp --env production
```

### Config File Issues

Reset configuration:

```bash
rm ~/.envhub/config.yaml
envhub login
```

## Shell Completions

### Bash

```bash
# Generate completion script
envhub completion bash > /etc/bash_completion.d/envhub

# Or add to .bashrc
echo 'source <(envhub completion bash)' >> ~/.bashrc
```

### Zsh

```bash
# Generate completion script
envhub completion zsh > "${fpath[1]}/_envhub"

# Or add to .zshrc
echo 'source <(envhub completion zsh)' >> ~/.zshrc
```

### Fish

```bash
# Generate completion script
envhub completion fish > ~/.config/fish/completions/envhub.fish
```

## Rate Limits

The API enforces rate limiting:
- 100 requests per minute per IP
- Per-user rate limiting available on paid plans

## Security Best Practices

1. **Never commit tokens** to version control
2. **Use environment variables** for CI/CD pipelines
3. **Rotate tokens** periodically
4. **Use short-lived tokens** for CLI access
5. **Enable 2FA** on your EnvHub account
