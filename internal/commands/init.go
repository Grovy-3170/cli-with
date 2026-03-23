package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Grovy-3170/cli-with/internal/service"
	"github.com/spf13/cobra"
)

// InitCmd creates the init command.
func InitCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new user vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			var username string
			if *cfg.User != "" {
				username = *cfg.User
			} else {
				fmt.Fprint(os.Stderr, "Create user: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("reading username: %w", err)
				}
				username = strings.TrimSpace(input)
				if username == "" {
					return fmt.Errorf("username cannot be empty")
				}
			}

			svc, err := service.NewVaultService()
			if err != nil {
				return err
			}

			vaultPath := svc.GetVaultPath(username)

			if svc.VaultExists(username) {
				return fmt.Errorf("vault for user '%s' already exists at %s", username, vaultPath)
			}

			password, err := cfg.GetPassword("Enter password for vault: ")
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}

			if len(password) == 0 {
				return fmt.Errorf("password cannot be empty")
			}

			confirmPassword, err := cfg.GetPassword("Confirm password: ")
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}

			if password != confirmPassword {
				return fmt.Errorf("passwords do not match")
			}

			if err := svc.CreateVault(username, password); err != nil {
				return fmt.Errorf("creating vault: %w", err)
			}

			fmt.Printf("Vault initialized successfully for user '%s' at %s\n", username, svc.GetVaultPath(username))
			return nil
		},
	}
}
