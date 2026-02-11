package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// Decrypt decrypts ciphertext using AES-256-GCM
// Expects format: [nonce (12 bytes)][ciphertext + auth tag]
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Validate key size
	if len(key) != AES256KeySize {
		return nil, ErrInvalidKeySize
	}

	// Validate ciphertext size
	if len(ciphertext) < MinEncryptedSize {
		return nil, ErrInvalidCiphertext
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create GCM: %v", ErrDecryptionFailed, err)
	}

	// Extract nonce (first 12 bytes)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrInvalidNonce
	}

	nonce := ciphertext[:nonceSize]
	encryptedData := ciphertext[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: authentication failed or corrupted data: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

// DecryptString decrypts a base64-encoded ciphertext to string
func DecryptString(encodedCiphertext string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}

	plaintext, err := Decrypt(ciphertext, key)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// DecryptWithDEK decrypts data using a Data Encryption Key
func DecryptWithDEK(ciphertext []byte, dek *DataKey) ([]byte, error) {
	if err := dek.Validate(); err != nil {
		return nil, err
	}
	return Decrypt(ciphertext, dek.Key)
}

// DecryptDEK decrypts an encrypted Data Encryption Key using the Master Key
// This is used to retrieve DEKs from the database
func DecryptDEK(encryptedDEK string, masterKey *MasterKey) (*DataKey, error) {
	if err := masterKey.Validate(); err != nil {
		return nil, err
	}

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("invalid base64-encoded DEK: %w", err)
	}

	// Decrypt with master key
	dekBytes, err := Decrypt(ciphertext, masterKey.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DEK: %w", err)
	}

	// Validate decrypted key size
	if len(dekBytes) != AES256KeySize {
		return nil, ErrInvalidKeySize
	}

	return &DataKey{
		Key:     dekBytes,
		Version: 1,
	}, nil
}
