package crypto

import (
	"errors"
)

// Common errors
var (
	ErrInvalidKeySize      = errors.New("crypto: invalid key size")
	ErrEncryptionFailed    = errors.New("crypto: encryption failed")
	ErrDecryptionFailed    = errors.New("crypto: decryption failed")
	ErrInvalidCiphertext   = errors.New("crypto: invalid ciphertext format")
	ErrInvalidNonce        = errors.New("crypto: invalid nonce size")
	ErrKeyDerivationFailed = errors.New("crypto: key derivation failed")
	ErrInvalidMasterKey    = errors.New("crypto: invalid master key")
)

// Key sizes in bytes
const (
	// AES-256 requires 32 bytes (256 bits)
	AES256KeySize = 32

	// GCM nonce size (96 bits / 12 bytes is standard)
	GCMNonceSize = 12

	// Minimum encrypted data size (nonce + ciphertext + tag)
	// GCM tag is 16 bytes
	MinEncryptedSize = GCMNonceSize + 16
)

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	// Ciphertext includes nonce prepended (first 12 bytes)
	Ciphertext []byte

	// Version for future algorithm changes
	Version int
}

// MasterKey represents the Key Encryption Key (KEK)
type MasterKey struct {
	Key     []byte
	Version int
}

// DataKey represents a Data Encryption Key (DEK)
type DataKey struct {
	Key     []byte
	Version int
}
