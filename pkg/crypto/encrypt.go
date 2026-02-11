package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encrypt encrypts plaintext using AES-256-GCM
// The nonce is prepended to the ciphertext for easy decryption
// Format: [nonce (12 bytes)][ciphertext + auth tag]
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Validate key size
	if len(key) != AES256KeySize {
		return nil, ErrInvalidKeySize
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create GCM: %v", ErrEncryptionFailed, err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("%w: nonce generation failed: %v", ErrEncryptionFailed, err)
	}

	// Encrypt and authenticate
	// Seal appends the ciphertext and tag to nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// EncryptString encrypts a string and returns base64-encoded ciphertext
func EncryptString(plaintext string, key []byte) (string, error) {
	ciphertext, err := Encrypt([]byte(plaintext), key)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// EncryptWithDEK encrypts data using a Data Encryption Key
func EncryptWithDEK(plaintext []byte, dek *DataKey) ([]byte, error) {
	if err := dek.Validate(); err != nil {
		return nil, err
	}
	return Encrypt(plaintext, dek.Key)
}

// EncryptDEK encrypts a Data Encryption Key with the Master Key
// This is used to store DEKs securely in the database
func EncryptDEK(dek *DataKey, masterKey *MasterKey) (string, error) {
	if err := masterKey.Validate(); err != nil {
		return "", err
	}
	if err := dek.Validate(); err != nil {
		return "", err
	}

	// Encrypt the DEK with the master key
	encrypted, err := Encrypt(dek.Key, masterKey.Key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	// Return base64-encoded encrypted DEK
	return base64.StdEncoding.EncodeToString(encrypted), nil
}
