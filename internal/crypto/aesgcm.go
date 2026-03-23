package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const (
	// KeySize is the required key size for AES-256 (32 bytes)
	KeySize = 32
	// NonceSize is the standard nonce size for GCM (12 bytes)
	NonceSize = 12
)

var (
	ErrInvalidKeySize   = errors.New("key must be 32 bytes for AES-256")
	ErrInvalidCiphertext = errors.New("ciphertext too short")
)

// AESGCM implements the Cipher interface using AES-256-GCM.
// This provides both confidentiality and authenticity for encrypted data.
type AESGCM struct{}

// NewAESGCM creates a new AES-256-GCM cipher instance.
func NewAESGCM() *AESGCM {
	return &AESGCM{}
}

// Encrypt encrypts plaintext using AES-256-GCM with the provided key.
// The key must be exactly 32 bytes for AES-256.
// Returns nonce || ciphertext, where nonce is 12 bytes prepended to the ciphertext.
// A new random nonce is generated for each encryption to ensure uniqueness.
func (a *AESGCM) Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, err := a.GenerateNonce()
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext that was encrypted with AES-256-GCM.
// Expects the ciphertext to be in format: nonce || ciphertext (12-byte nonce prepended).
// Verifies the authentication tag during decryption - will fail if ciphertext was tampered.
func (a *AESGCM) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}

	if len(ciphertext) < NonceSize+16 {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := ciphertext[:NonceSize]
	actualCiphertext := ciphertext[NonceSize:]

	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// GenerateNonce generates a cryptographically secure random 12-byte nonce.
// This should be called for each encryption operation to ensure nonce uniqueness.
func (a *AESGCM) GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
