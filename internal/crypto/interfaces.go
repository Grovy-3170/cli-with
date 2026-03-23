package crypto

import "errors"

var (
	ErrInvalidKey       = errors.New("invalid key")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrInvalidNonce     = errors.New("invalid nonce")
	ErrKeyNotFound      = errors.New("key not found")
)

type KeyDerivationFunc interface {
	DeriveKey(password, salt []byte) []byte
	Verify(password []byte, key []byte) bool
	GenerateSalt() ([]byte, error)
}

type Cipher interface {
	Encrypt(plaintext []byte, key []byte) ([]byte, error)
	Decrypt(ciphertext []byte, key []byte) ([]byte, error)
	GenerateNonce() ([]byte, error)
}
