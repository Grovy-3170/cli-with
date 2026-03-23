// Package commands contains all CLI command implementations.
package commands

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Grovy-3170/cli-with/internal/service"
)

// Config holds shared configuration and dependencies for all commands.
type Config struct {
	User         *string
	PasswordFile *string
	KeyValue     *string
	GetPassword  func(prompt string) (string, error)
	ReadPassword func(prompt string) (string, error)
	ReadConfirm  func(prompt string) (string, error)
}

// validKeyNameRegex validates key names: must start with letter or underscore,
// followed by letters, digits, or underscores
var validKeyNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidateKeyName checks if a key name is valid.
func ValidateKeyName(name string) error {
	if name == "" {
		return fmt.Errorf("key name cannot be empty")
	}
	if !validKeyNameRegex.MatchString(name) {
		return fmt.Errorf("invalid key name '%s': must start with letter or underscore, followed by letters, digits, or underscores", name)
	}
	return nil
}

// SelectUser prompts the user to select from available vaults.
func SelectUser() (string, error) {
	svc, err := service.NewVaultService()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	entries, err := os.ReadDir(svc.GetVaultDir())
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no users found, run 'with init' first")
		}
		return "", fmt.Errorf("failed to list users: %w", err)
	}

	var users []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".vault") {
			username := strings.TrimSuffix(entry.Name(), ".vault")
			users = append(users, username)
		}
	}

	if len(users) == 0 {
		return "", fmt.Errorf("no users found, run 'with init' first")
	}
	if len(users) == 1 {
		return users[0], nil
	}

	fmt.Fprintln(os.Stderr, "Select user:")
	for i, u := range users {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, u)
	}
	fmt.Fprint(os.Stderr, "Enter number: ")

	var choice int
	_, err = fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(users) {
		return "", fmt.Errorf("invalid selection")
	}

	return users[choice-1], nil
}

// GetUsername returns the username from config or prompts for selection.
func GetUsername(cfg *Config) (string, error) {
	if *cfg.User != "" {
		return *cfg.User, nil
	}
	return SelectUser()
}
