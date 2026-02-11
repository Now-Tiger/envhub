package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// GenerateMasterKey creates a new 256-bit master key (KEK)
// This should be called once during initial setup and stored securely
func GenerateMasterKey() (*MasterKey, error) {
	key := make([]byte, AES256KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	return &MasterKey{
		Key:     key,
		Version: 1,
	}, nil
}

// GenerateDataKey creates a new 256-bit data key (DEK)
// This is called for each new project
func GenerateDataKey() (*DataKey, error) {
	key := make([]byte, AES256KeySize)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate data key: %w", err)
	}

	return &DataKey{
		Key:     key,
		Version: 1,
	}, nil
}

// MasterKeyFromBase64 loads a master key from base64 string
// Used to load KEK from environment variable or config
func MasterKeyFromBase64(encoded string) (*MasterKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)

	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(decoded) != AES256KeySize {
		return nil, ErrInvalidKeySize
	}

	return &MasterKey{
		Key:     decoded,
		Version: 1,
	}, nil
}

// ToBase64 encodes the master key to base64 for storage
func (mk *MasterKey) ToBase64() string { return base64.StdEncoding.EncodeToString(mk.Key) }

// Validate checks if the master key is valid
func (mk *MasterKey) Validate() error {
	if len(mk.Key) != AES256KeySize {
		return ErrInvalidKeySize
	}

	if mk.Version < 1 {
		return ErrInvalidMasterKey
	}

	return nil
}

// Validate checks if the data key is valid
func (dk *DataKey) Validate() error {
	if len(dk.Key) != AES256KeySize {
		return ErrInvalidKeySize
	}

	if dk.Version < 1 {
		return fmt.Errorf("invalid data key version")
	}

	return nil
}
