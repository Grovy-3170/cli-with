package service

import (
	"fmt"
	"os"

	"github.com/Grovy-3170/cli-with/internal/config"
	"github.com/Grovy-3170/cli-with/internal/storage"
)

// VaultService provides high-level vault operations.
// It wraps config and storage to reduce duplication in command handlers.
type VaultService struct {
	config *config.Config
}

// NewVaultService creates a new VaultService with loaded config.
func NewVaultService() (*VaultService, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &VaultService{config: cfg}, nil
}

// GetVaultPath returns the vault file path for a user.
func (s *VaultService) GetVaultPath(username string) string {
	return s.config.GetVaultPath(username)
}

// VaultExists checks if a vault exists for the given user.
func (s *VaultService) VaultExists(username string) bool {
	vaultPath := s.config.GetVaultPath(username)
	_, err := os.Stat(vaultPath)
	return err == nil
}

// EnsureVaultDir creates the vault directory if it doesn't exist.
func (s *VaultService) EnsureVaultDir() error {
	return s.config.EnsureVaultDir()
}

// CreateVault creates a new vault for a user with the given password.
func (s *VaultService) CreateVault(username, password string) error {
	vaultPath := s.config.GetVaultPath(username)

	if _, err := os.Stat(vaultPath); err == nil {
		return fmt.Errorf("vault for user '%s' already exists at %s", username, vaultPath)
	}

	if err := s.config.EnsureVaultDir(); err != nil {
		return fmt.Errorf("creating vault directory: %w", err)
	}

	vault, err := storage.Create(password, nil)
	if err != nil {
		return fmt.Errorf("creating vault: %w", err)
	}

	if err := vault.Save(vaultPath); err != nil {
		return fmt.Errorf("saving vault: %w", err)
	}

	return nil
}

// LoadAndDecrypt loads a vault and decrypts it with the given password.
// Returns the decrypted keys and the vault (for potential updates).
func (s *VaultService) LoadAndDecrypt(username, password string) (map[string]string, *storage.Vault, error) {
	vaultPath := s.config.GetVaultPath(username)

	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("vault for user '%s' does not exist", username)
	}

	vault, err := storage.Load(vaultPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading vault: %w", err)
	}

	keys, err := vault.Decrypt(password)
	if err != nil {
		return nil, nil, fmt.Errorf("decrypting vault: %w", err)
	}

	return keys, vault, nil
}

// UpdateAndSave updates the vault with new keys and saves it.
func (s *VaultService) UpdateAndSave(username, password string, vault *storage.Vault, keys map[string]string) error {
	if err := vault.UpdateKeys(password, keys); err != nil {
		return fmt.Errorf("updating vault: %w", err)
	}

	vaultPath := s.config.GetVaultPath(username)
	if err := vault.Save(vaultPath); err != nil {
		return fmt.Errorf("saving vault: %w", err)
	}

	return nil
}

// DeleteVault removes the vault file for a user.
func (s *VaultService) DeleteVault(username string) error {
	vaultPath := s.config.GetVaultPath(username)
	if err := os.Remove(vaultPath); err != nil {
		return fmt.Errorf("deleting vault: %w", err)
	}
	return nil
}

// GetVaultDir returns the vault directory path.
func (s *VaultService) GetVaultDir() string {
	return s.config.VaultDir
}
