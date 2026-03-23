package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	VaultDir string
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	vaultDir := os.Getenv("WITH_VAULT_DIR")
	if vaultDir == "" {
		vaultDir = filepath.Join(homeDir, ".config", "cli-with", "users")
	}

	return &Config{
		VaultDir: vaultDir,
	}, nil
}

func (c *Config) GetVaultPath(username string) string {
	return filepath.Join(c.VaultDir, username+".vault")
}

func (c *Config) EnsureVaultDir() error {
	if err := os.MkdirAll(c.VaultDir, 0700); err != nil {
		return fmt.Errorf("creating vault directory: %w", err)
	}
	return nil
}
