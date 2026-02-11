package crypto

import (
	"bytes"
	"testing"
)

func TestGenerateMasterKey(t *testing.T) {
	mk, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey failed: %v", err)
	}

	if len(mk.Key) != AES256KeySize {
		t.Errorf("Expected key size %d, got %d", AES256KeySize, len(mk.Key))
	}

	if mk.Version != 1 {
		t.Errorf("Expected version 1, got %d", mk.Version)
	}
}

func TestGenerateDataKey(t *testing.T) {
	dk, err := GenerateDataKey()
	if err != nil {
		t.Fatalf("GenerateDataKey failed: %v", err)
	}

	if len(dk.Key) != AES256KeySize {
		t.Errorf("Expected key size %d, got %d", AES256KeySize, len(dk.Key))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"Simple text", "Hello, World!"},
		{"Empty string", ""},
		{"Long text", "This is a longer piece of text that should still work correctly with encryption and decryption"},
		{"Special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"Unicode", "Hello 世界 🌍"},
	}

	// Generate a test key
	dk, err := GenerateDataKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := Encrypt([]byte(tt.plaintext), dk.Key)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Decrypt
			decrypted, err := Decrypt(ciphertext, dk.Key)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if string(decrypted) != tt.plaintext {
				t.Errorf("Expected %q, got %q", tt.plaintext, string(decrypted))
			}
		})
	}
}

func TestEncryptDecryptString(t *testing.T) {
	dk, err := GenerateDataKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	plaintext := "SECRET_API_KEY=sk_test_12345"

	// Encrypt
	encrypted, err := EncryptString(plaintext, dk.Key)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptString(encrypted, dk.Key)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptDecryptDEK(t *testing.T) {
	// Generate master key
	mk, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}

	// Generate data key
	dek, err := GenerateDataKey()
	if err != nil {
		t.Fatalf("Failed to generate data key: %v", err)
	}

	// Encrypt DEK
	encryptedDEK, err := EncryptDEK(dek, mk)
	if err != nil {
		t.Fatalf("EncryptDEK failed: %v", err)
	}

	// Decrypt DEK
	decryptedDEK, err := DecryptDEK(encryptedDEK, mk)
	if err != nil {
		t.Fatalf("DecryptDEK failed: %v", err)
	}

	// Compare keys
	if !bytes.Equal(dek.Key, decryptedDEK.Key) {
		t.Errorf("Decrypted DEK doesn't match original")
	}
}

func TestMasterKeyBase64(t *testing.T) {
	mk, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("Failed to generate master key: %v", err)
	}

	// Convert to base64
	encoded := mk.ToBase64()

	// Load from base64
	loaded, err := MasterKeyFromBase64(encoded)
	if err != nil {
		t.Fatalf("MasterKeyFromBase64 failed: %v", err)
	}

	// Compare
	if !bytes.Equal(mk.Key, loaded.Key) {
		t.Errorf("Loaded key doesn't match original")
	}
}

func TestInvalidKey(t *testing.T) {
	plaintext := []byte("test data")
	invalidKey := []byte("too_short")

	// Should fail with invalid key
	_, err := Encrypt(plaintext, invalidKey)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got %v", err)
	}
}

func TestInvalidCiphertext(t *testing.T) {
	dk, _ := GenerateDataKey()

	// Too short ciphertext
	_, err := Decrypt([]byte("short"), dk.Key)
	if err != ErrInvalidCiphertext {
		t.Errorf("Expected ErrInvalidCiphertext, got %v", err)
	}
}

func TestRotateMasterKey(t *testing.T) {
	// Generate old and new master keys
	oldMK, _ := GenerateMasterKey()
	newMK, _ := GenerateMasterKey()

	// Generate some DEKs and encrypt them with old master key
	dek1, _ := GenerateDataKey()
	dek2, _ := GenerateDataKey()

	encDEK1, _ := EncryptDEK(dek1, oldMK)
	encDEK2, _ := EncryptDEK(dek2, oldMK)

	encryptedDEKs := []string{encDEK1, encDEK2}

	// Rotate
	rotatedDEKs, err := RotateMasterKey(oldMK, newMK, encryptedDEKs)
	if err != nil {
		t.Fatalf("RotateMasterKey failed: %v", err)
	}

	// Verify we can decrypt with new master key
	for i, rotatedDEK := range rotatedDEKs {
		decrypted, err := DecryptDEK(rotatedDEK, newMK)
		if err != nil {
			t.Fatalf("Failed to decrypt rotated DEK %d: %v", i, err)
		}

		// Verify it matches original DEK
		var originalDEK *DataKey
		if i == 0 {
			originalDEK = dek1
		} else {
			originalDEK = dek2
		}

		if !bytes.Equal(decrypted.Key, originalDEK.Key) {
			t.Errorf("Rotated DEK %d doesn't match original", i)
		}
	}
}

func TestRotateProjectDEK(t *testing.T) {
	// Setup
	masterKey, _ := GenerateMasterKey()
	oldDEK, _ := GenerateDataKey()
	encryptedOldDEK, _ := EncryptDEK(oldDEK, masterKey)

	// Create some test secrets
	secret1, _ := EncryptString("DATABASE_URL=postgres://...", oldDEK.Key)
	secret2, _ := EncryptString("API_KEY=sk_test_123", oldDEK.Key)
	secrets := []string{secret1, secret2}

	// Rotate
	newDEK, newEncryptedDEK, reencryptedSecrets, err := RotateProjectDEK(
		encryptedOldDEK,
		masterKey,
		secrets,
	)
	if err != nil {
		t.Fatalf("RotateProjectDEK failed: %v", err)
	}

	// Verify new DEK version
	if newDEK.Version != oldDEK.Version+1 {
		t.Errorf("Expected version %d, got %d", oldDEK.Version+1, newDEK.Version)
	}

	// Verify we can decrypt with new DEK
	decrypted1, err := DecryptString(reencryptedSecrets[0], newDEK.Key)
	if err != nil {
		t.Fatalf("Failed to decrypt secret 1: %v", err)
	}
	if decrypted1 != "DATABASE_URL=postgres://..." {
		t.Errorf("Secret 1 content mismatch")
	}

	// Verify we can decrypt the new DEK with master key
	_, err = DecryptDEK(newEncryptedDEK, masterKey)
	if err != nil {
		t.Fatalf("Failed to decrypt new DEK: %v", err)
	}
}

// Benchmark tests
func BenchmarkEncrypt(b *testing.B) {
	dk, _ := GenerateDataKey()
	plaintext := []byte("This is a test secret value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encrypt(plaintext, dk.Key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	dk, _ := GenerateDataKey()
	plaintext := []byte("This is a test secret value")
	ciphertext, _ := Encrypt(plaintext, dk.Key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decrypt(ciphertext, dk.Key)
	}
}
