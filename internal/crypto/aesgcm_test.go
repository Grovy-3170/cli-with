package crypto

import (
	"crypto/rand"
	"testing"
)

func TestAESGCM_EncryptDecrypt_Roundtrip(t *testing.T) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("this is a secret message that needs encryption")
	ciphertext, err := cipher.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := cipher.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted text does not match original\nwant: %s\ngot: %s", plaintext, decrypted)
	}
}

func TestAESGCM_Encrypt_EmptyPlaintext(t *testing.T) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte{}
	ciphertext, err := cipher.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := cipher.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty decrypted text, got: %v", decrypted)
	}
}

func TestAESGCM_Encrypt_InvalidKeySize(t *testing.T) {
	cipher := NewAESGCM()
	tests := []struct {
		name    string
		keySize int
	}{
		{"too short - 16 bytes", 16},
		{"too short - 24 bytes", 24},
		{"too long - 64 bytes", 64},
		{"empty key", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			_, err := cipher.Encrypt([]byte("test"), key)
			if err != ErrInvalidKeySize {
				t.Errorf("expected ErrInvalidKeySize, got: %v", err)
			}
		})
	}
}

func TestAESGCM_Decrypt_InvalidKeySize(t *testing.T) {
	cipher := NewAESGCM()

	validKey := make([]byte, KeySize)
	rand.Read(validKey)
	validCiphertext, _ := cipher.Encrypt([]byte("test"), validKey)

	tests := []struct {
		name    string
		keySize int
	}{
		{"too short - 16 bytes", 16},
		{"too short - 24 bytes", 24},
		{"too long - 64 bytes", 64},
		{"empty key", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			_, err := cipher.Decrypt(validCiphertext, key)
			if err != ErrInvalidKeySize {
				t.Errorf("expected ErrInvalidKeySize, got: %v", err)
			}
		})
	}
}

func TestAESGCM_Decrypt_TamperedCiphertext(t *testing.T) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("secret data")
	ciphertext, err := cipher.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	tests := []struct {
		name       string
		tamperFunc func([]byte) []byte
	}{
		{
			name: "modify nonce byte",
			tamperFunc: func(ct []byte) []byte {
				tampered := make([]byte, len(ct))
				copy(tampered, ct)
				tampered[0] ^= 0xFF
				return tampered
			},
		},
		{
			name: "modify ciphertext byte",
			tamperFunc: func(ct []byte) []byte {
				tampered := make([]byte, len(ct))
				copy(tampered, ct)
				tampered[NonceSize+5] ^= 0xFF
				return tampered
			},
		},
		{
			name: "modify last byte (part of auth tag)",
			tamperFunc: func(ct []byte) []byte {
				tampered := make([]byte, len(ct))
				copy(tampered, ct)
				tampered[len(tampered)-1] ^= 0xFF
				return tampered
			},
		},
		{
			name: "truncate ciphertext",
			tamperFunc: func(ct []byte) []byte {
				return ct[:len(ct)-1]
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tampered := tt.tamperFunc(ciphertext)
			_, err := cipher.Decrypt(tampered, key)
			if err != ErrDecryptionFailed && err != ErrInvalidCiphertext {
				t.Errorf("expected decryption to fail, got error: %v", err)
			}
		})
	}
}

func TestAESGCM_Decrypt_WrongKey(t *testing.T) {
	cipher := NewAESGCM()

	correctKey := make([]byte, KeySize)
	wrongKey := make([]byte, KeySize)
	if _, err := rand.Read(correctKey); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	if _, err := rand.Read(wrongKey); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("secret message")
	ciphertext, err := cipher.Encrypt(plaintext, correctKey)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = cipher.Decrypt(ciphertext, wrongKey)
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed, got: %v", err)
	}
}

func TestAESGCM_Encrypt_UniqueNonces(t *testing.T) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := []byte("same plaintext encrypted twice")

	ciphertext1, err := cipher.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("first Encrypt failed: %v", err)
	}

	ciphertext2, err := cipher.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("second Encrypt failed: %v", err)
	}

	if string(ciphertext1) == string(ciphertext2) {
		t.Error("same plaintext produced identical ciphertext - nonce reuse detected!")
	}

	nonce1 := ciphertext1[:NonceSize]
	nonce2 := ciphertext2[:NonceSize]
	if string(nonce1) == string(nonce2) {
		t.Error("nonces should be unique but are identical")
	}

	decrypted1, err := cipher.Decrypt(ciphertext1, key)
	if err != nil {
		t.Fatalf("Decrypt ciphertext1 failed: %v", err)
	}
	decrypted2, err := cipher.Decrypt(ciphertext2, key)
	if err != nil {
		t.Fatalf("Decrypt ciphertext2 failed: %v", err)
	}

	if string(decrypted1) != string(plaintext) || string(decrypted2) != string(plaintext) {
		t.Error("decrypted texts don't match original plaintext")
	}
}

func TestAESGCM_GenerateNonce(t *testing.T) {
	cipher := NewAESGCM()

	nonce1, err := cipher.GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}

	if len(nonce1) != NonceSize {
		t.Errorf("expected nonce size %d, got %d", NonceSize, len(nonce1))
	}

	nonce2, err := cipher.GenerateNonce()
	if err != nil {
		t.Fatalf("second GenerateNonce failed: %v", err)
	}

	if string(nonce1) == string(nonce2) {
		t.Error("consecutive nonces should be unique")
	}
}

func TestAESGCM_Decrypt_TooShortCiphertext(t *testing.T) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	rand.Read(key)

	tests := []struct {
		name string
		size int
	}{
		{"empty", 0},
		{"1 byte", 1},
		{"nonce only - 12 bytes", NonceSize},
		{"nonce + partial tag - 20 bytes", NonceSize + 8},
		{"nonce + incomplete tag - 27 bytes", NonceSize + 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext := make([]byte, tt.size)
			_, err := cipher.Decrypt(ciphertext, key)
			if err != ErrInvalidCiphertext {
				t.Errorf("expected ErrInvalidCiphertext, got: %v", err)
			}
		})
	}
}

func TestAESGCM_LargePlaintext(t *testing.T) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	plaintext := make([]byte, 1024*1024)
	if _, err := rand.Read(plaintext); err != nil {
		t.Fatalf("failed to generate plaintext: %v", err)
	}

	ciphertext, err := cipher.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := cipher.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if len(decrypted) != len(plaintext) {
		t.Errorf("decrypted length mismatch: want %d, got %d", len(plaintext), len(decrypted))
	}

	for i := range plaintext {
		if plaintext[i] != decrypted[i] {
			t.Errorf("decrypted data mismatch at byte %d", i)
			break
		}
	}
}

func TestAESGCM_ImplementsCipherInterface(t *testing.T) {
	var _ Cipher = NewAESGCM()
}

func BenchmarkAESGCM_Encrypt(b *testing.B) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	rand.Read(key)

	plaintext := make([]byte, 1024)
	rand.Read(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Encrypt(plaintext, key)
	}
}

func BenchmarkAESGCM_Decrypt(b *testing.B) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	rand.Read(key)

	plaintext := make([]byte, 1024)
	rand.Read(plaintext)

	ciphertext, _ := cipher.Encrypt(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Decrypt(ciphertext, key)
	}
}

func BenchmarkAESGCM_GenerateNonce(b *testing.B) {
	cipher := NewAESGCM()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.GenerateNonce()
	}
}

func BenchmarkAESGCM_Encrypt_Large(b *testing.B) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	rand.Read(key)

	plaintext := make([]byte, 64*1024)
	rand.Read(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Encrypt(plaintext, key)
	}
}

func BenchmarkAESGCM_Decrypt_Large(b *testing.B) {
	cipher := NewAESGCM()
	key := make([]byte, KeySize)
	rand.Read(key)

	plaintext := make([]byte, 64*1024)
	rand.Read(plaintext)

	ciphertext, _ := cipher.Encrypt(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Decrypt(ciphertext, key)
	}
}
