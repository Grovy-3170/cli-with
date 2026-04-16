package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	VaultDir  string
	ConfigDir string
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "cli-with")

	vaultDir := os.Getenv("WITH_VAULT_DIR")
	if vaultDir == "" {
		vaultDir = filepath.Join(configDir, "users")
	}

	return &Config{
		VaultDir:  vaultDir,
		ConfigDir: configDir,
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

// GetAliasPath returns the path to the aliases file. If WITH_ALIAS_FILE is set,
// it takes precedence — primarily a testability hook, but also lets users
// relocate the file explicitly.
func (c *Config) GetAliasPath() string {
	if v := os.Getenv("WITH_ALIAS_FILE"); v != "" {
		return v
	}
	return filepath.Join(c.ConfigDir, "aliases.json")
}
