# EnvHub Crypto Package

Production-grade encryption for EnvHub using AES-256-GCM.

## Architecture

### Two-Tier Key Hierarchy

```
Master Key (KEK)
    └─> Data Key 1 (DEK) → Project 1 Secrets
    └─> Data Key 2 (DEK) → Project 2 Secrets
    └─> Data Key 3 (DEK) → Project 3 Secrets
```

- **Master Key (KEK)**: Single key that encrypts all Data Keys
- **Data Key (DEK)**: Unique key per project that encrypts secrets

### Benefits

1. **Key Rotation**: Rotate master key without re-encrypting all secrets
2. **Isolation**: Compromise of one DEK doesn't affect other projects
3. **Performance**: Secrets encrypted/decrypted with DEK (no master key needed)

## Usage

### Initialize Master Key (Once)

```go
// Generate new master key
masterKey, err := crypto.GenerateMasterKey()
if err != nil {
    log.Fatal(err)
}

// Save to environment variable (base64)
os.Setenv("MASTER_ENCRYPTION_KEY", masterKey.ToBase64())
```

### Create New Project

```go
// Load master key from env
masterKey, err := crypto.MasterKeyFromBase64(os.Getenv("MASTER_ENCRYPTION_KEY"))

// Generate DEK for new project
dek, err := crypto.GenerateDataKey()

// Encrypt DEK with master key for storage
encryptedDEK, err := crypto.EncryptDEK(dek, masterKey)

// Save encryptedDEK to database project.encrypted_dek column
```

### Encrypt Secret

```go
// Load project's encrypted DEK from database
encryptedDEK := project.EncryptedDEK

// Decrypt DEK with master key
dek, err := crypto.DecryptDEK(encryptedDEK, masterKey)

// Encrypt secret with DEK
encryptedSecret, err := crypto.EncryptString("DATABASE_URL=postgres://...", dek.Key)

// Save encryptedSecret to database secrets.encrypted_value column
```

### Decrypt Secret

```go
// Load project's encrypted DEK
dek, err := crypto.DecryptDEK(project.EncryptedDEK, masterKey)

// Decrypt secret
plaintext, err := crypto.DecryptString(secret.EncryptedValue, dek.Key)

// Use plaintext (never log or persist!)
```

## Security Best Practices

1. **Master Key Storage**: Store in AWS KMS, HashiCorp Vault, or secure env var
2. **Never Log Keys**: Keys should never appear in logs
3. **Rotate Regularly**: Use `RotateMasterKey()` quarterly
4. **Audit Access**: Log all decrypt operations
5. **Limit Scope**: Use separate master keys for dev/staging/prod

## Testing

```bash
go test -v ./pkg/crypto
go test -bench=. ./pkg/crypto
```

## Algorithm Details

- **Cipher**: AES-256
- **Mode**: GCM (Galois/Counter Mode)
- **Key Size**: 256 bits (32 bytes)
- **Nonce Size**: 96 bits (12 bytes)
- **Tag Size**: 128 bits (16 bytes)

GCM provides both confidentiality and authenticity.
