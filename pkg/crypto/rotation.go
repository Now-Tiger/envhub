package crypto

import (
	"encoding/base64"
	"fmt"
)

// RotationResult contains the results of a key rotation
type RotationResult struct {
	OldKeyVersion int
	NewKeyVersion int
	ItemsRotated  int
}

// RotateMasterKey handles master key rotation
// This re-encrypts all DEKs with a new master key
func RotateMasterKey(oldMasterKey, newMasterKey *MasterKey, encryptedDEKs []string) ([]string, error) {
	if err := oldMasterKey.Validate(); err != nil {
		return nil, fmt.Errorf("invalid old master key: %w", err)
	}
	if err := newMasterKey.Validate(); err != nil {
		return nil, fmt.Errorf("invalid new master key: %w", err)
	}

	reencryptedDEKs := make([]string, 0, len(encryptedDEKs))

	for _, encryptedDEK := range encryptedDEKs {
		// Decrypt DEK with old master key
		dek, err := DecryptDEK(encryptedDEK, oldMasterKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt DEK during rotation: %w", err)
		}

		// Re-encrypt DEK with new master key
		newEncryptedDEK, err := EncryptDEK(dek, newMasterKey)
		if err != nil {
			return nil, fmt.Errorf("failed to re-encrypt DEK during rotation: %w", err)
		}

		reencryptedDEKs = append(reencryptedDEKs, newEncryptedDEK)
	}

	return reencryptedDEKs, nil
}

// RotateProjectDEK creates a new DEK for a project and re-encrypts all secrets
// Returns the new encrypted DEK and re-encrypted secrets
func RotateProjectDEK(
	oldEncryptedDEK string,
	masterKey *MasterKey,
	encryptedSecrets []string,
) (*DataKey, string, []string, error) {
	// Decrypt old DEK
	oldDEK, err := DecryptDEK(oldEncryptedDEK, masterKey)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to decrypt old DEK: %w", err)
	}

	// Generate new DEK
	newDEK, err := GenerateDataKey()
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to generate new DEK: %w", err)
	}
	newDEK.Version = oldDEK.Version + 1

	// Encrypt new DEK with master key
	newEncryptedDEK, err := EncryptDEK(newDEK, masterKey)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to encrypt new DEK: %w", err)
	}

	// Re-encrypt all secrets with new DEK
	reencryptedSecrets := make([]string, 0, len(encryptedSecrets))
	for _, encryptedSecret := range encryptedSecrets {
		// Decrypt with old DEK
		secretBytes, err := base64.StdEncoding.DecodeString(encryptedSecret)
		if err != nil {
			return nil, "", nil, fmt.Errorf("invalid base64 secret: %w", err)
		}

		plaintext, err := DecryptWithDEK(secretBytes, oldDEK)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to decrypt secret: %w", err)
		}

		// Encrypt with new DEK
		newCiphertext, err := EncryptWithDEK(plaintext, newDEK)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to re-encrypt secret: %w", err)
		}

		reencryptedSecrets = append(reencryptedSecrets, base64.StdEncoding.EncodeToString(newCiphertext))
	}

	return newDEK, newEncryptedDEK, reencryptedSecrets, nil
}
