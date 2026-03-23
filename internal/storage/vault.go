package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Grovy-3170/cli-with/internal/crypto"
)

const (
	VaultVersion = 1
	VaultFormat  = "cli-with-vault"
)

var (
	ErrInvalidPassword     = errors.New("invalid password")
	ErrCorruptedVault      = errors.New("corrupted vault file")
	ErrInvalidVaultFormat  = errors.New("invalid vault format")
	ErrInvalidVaultVersion = errors.New("invalid vault version")
	ErrInvalidPermissions  = errors.New("invalid file permissions")
)

type KDFParams struct {
	Algorithm   string `json:"algorithm"`
	Memory      uint32 `json:"memory"`
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism"`
	Salt        []byte `json:"salt"`
}

type EncryptionParams struct {
	Algorithm string `json:"algorithm"`
	Nonce     []byte `json:"nonce"`
}

type Vault struct {
	Version    int              `json:"version"`
	Format     string           `json:"format"`
	CreatedAt  time.Time        `json:"created_at"`
	KDF        KDFParams        `json:"kdf"`
	Encryption EncryptionParams `json:"encryption"`
	Ciphertext []byte           `json:"ciphertext"`

	deriver *crypto.Argon2idDeriver
	cipher  *crypto.AESGCM
}

func Create(password string, keys map[string]string) (*Vault, error) {
	if keys == nil {
		keys = make(map[string]string)
	}

	deriver := crypto.NewArgon2idDeriver()
	cipher := crypto.NewAESGCM()

	salt, err := deriver.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := deriver.DeriveKey([]byte(password), salt)

	plaintext, err := json.Marshal(keys)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize keys: %w", err)
	}

	ciphertext, err := cipher.Encrypt(plaintext, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt keys: %w", err)
	}

	nonce := ciphertext[:crypto.NonceSize]
	actualCiphertext := ciphertext[crypto.NonceSize:]

	return &Vault{
		Version:   VaultVersion,
		Format:    VaultFormat,
		CreatedAt: time.Now().UTC(),
		KDF: KDFParams{
			Algorithm:   "argon2id",
			Memory:      crypto.Argon2idMemory,
			Iterations:  crypto.Argon2idIterations,
			Parallelism: crypto.Argon2idParallelism,
			Salt:        salt,
		},
		Encryption: EncryptionParams{
			Algorithm: "aes-256-gcm",
			Nonce:     nonce,
		},
		Ciphertext: actualCiphertext,
		deriver:    deriver,
		cipher:     cipher,
	}, nil
}

func (v *Vault) Decrypt(password string) (map[string]string, error) {
	if v.Format != VaultFormat {
		return nil, ErrInvalidVaultFormat
	}
	if v.Version != VaultVersion {
		return nil, ErrInvalidVaultVersion
	}

	if v.deriver == nil {
		v.deriver = crypto.NewArgon2idDeriver()
	}
	if v.cipher == nil {
		v.cipher = crypto.NewAESGCM()
	}

	key := v.deriver.DeriveKey([]byte(password), v.KDF.Salt)

	fullCiphertext := make([]byte, 0, len(v.Encryption.Nonce)+len(v.Ciphertext))
	fullCiphertext = append(fullCiphertext, v.Encryption.Nonce...)
	fullCiphertext = append(fullCiphertext, v.Ciphertext...)

	plaintext, err := v.cipher.Decrypt(fullCiphertext, key)
	if err != nil {
		if errors.Is(err, crypto.ErrDecryptionFailed) {
			return nil, ErrInvalidPassword
		}
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	var keys map[string]string
	if err := json.Unmarshal(plaintext, &keys); err != nil {
		return nil, ErrCorruptedVault
	}

	return keys, nil
}

func (v *Vault) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create vault directory: %w", err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize vault: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600) // #nosec G304 -- path from config, not arbitrary input
	if err != nil {
		return fmt.Errorf("failed to create vault file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write vault: %w", err)
	}

	return nil
}

func Load(path string) (*Vault, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat vault file: %w", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		return nil, fmt.Errorf("%w: expected 0600, got %04o", ErrInvalidPermissions, perm)
	}

	data, err := os.ReadFile(path) // #nosec G304 -- path from config, not arbitrary input
	if err != nil {
		return nil, fmt.Errorf("failed to read vault file: %w", err)
	}

	var vault Vault
	if err := json.Unmarshal(data, &vault); err != nil {
		return nil, fmt.Errorf("failed to parse vault: %w", err)
	}

	if vault.Format != VaultFormat {
		return nil, ErrInvalidVaultFormat
	}
	if vault.Version != VaultVersion {
		return nil, ErrInvalidVaultVersion
	}

	vault.deriver = crypto.NewArgon2idDeriver()
	vault.cipher = crypto.NewAESGCM()

	return &vault, nil
}

func (v *Vault) UpdateKeys(password string, keys map[string]string) error {
	if keys == nil {
		keys = make(map[string]string)
	}

	_, err := v.Decrypt(password)
	if err != nil {
		return err
	}

	key := v.deriver.DeriveKey([]byte(password), v.KDF.Salt)

	plaintext, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("failed to serialize keys: %w", err)
	}

	ciphertext, err := v.cipher.Encrypt(plaintext, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt keys: %w", err)
	}

	v.Encryption.Nonce = ciphertext[:crypto.NonceSize]
	v.Ciphertext = ciphertext[crypto.NonceSize:]

	return nil
}
