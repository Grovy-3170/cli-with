package crypto

import (
	"crypto/rand"
	"crypto/subtle"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters following OWASP recommendations
const (
	Argon2idMemory      = 64 * 1024 // 64 MiB
	Argon2idIterations  = 3
	Argon2idParallelism = 1
	Argon2idKeyLength   = 32 // 256 bits for AES-256
	Argon2idSaltLength  = 16 // 128 bits
)

// Argon2idDeriver implements KeyDerivationFunc using Argon2id.
// Argon2id is the recommended algorithm for password hashing and key derivation,
// providing resistance against GPU and side-channel attacks.
type Argon2idDeriver struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
	saltLength  int
}

// NewArgon2idDeriver creates a new Argon2idDeriver with secure default parameters.
func NewArgon2idDeriver() *Argon2idDeriver {
	return &Argon2idDeriver{
		memory:      Argon2idMemory,
		iterations:  Argon2idIterations,
		parallelism: Argon2idParallelism,
		keyLength:   Argon2idKeyLength,
		saltLength:  Argon2idSaltLength,
	}
}

// DeriveKey derives a cryptographic key from the given password and salt using Argon2id.
// The same password and salt will always produce the same key.
func (d *Argon2idDeriver) DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		d.iterations,
		d.memory,
		d.parallelism,
		d.keyLength,
	)
}

// Verify performs constant-time comparison to prevent timing attacks.
func (d *Argon2idDeriver) Verify(password []byte, key []byte) bool {
	if len(key) != int(d.keyLength) {
		return false
	}
	return subtle.ConstantTimeCompare(password, key) == 1
}

// VerifyWithSalt derives a key from password+salt and compares it to the expected key.
// This is the proper way to verify a password when you have the stored salt.
func (d *Argon2idDeriver) VerifyWithSalt(password, salt, expectedKey []byte) bool {
	if len(expectedKey) != int(d.keyLength) {
		return false
	}
	derivedKey := d.DeriveKey(password, salt)
	return subtle.ConstantTimeCompare(derivedKey, expectedKey) == 1
}

// GenerateSalt generates a cryptographically secure random salt.
// The salt is used to ensure unique keys even for the same password.
func (d *Argon2idDeriver) GenerateSalt() ([]byte, error) {
	salt := make([]byte, d.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

// GetKeyLength returns the configured key length in bytes.
func (d *Argon2idDeriver) GetKeyLength() uint32 {
	return d.keyLength
}

// GetSaltLength returns the configured salt length in bytes.
func (d *Argon2idDeriver) GetSaltLength() int {
	return d.saltLength
}
