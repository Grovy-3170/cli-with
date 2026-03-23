package commands

import (
	"fmt"

	"github.com/Grovy-3170/cli-with/internal/service"
	"github.com/spf13/cobra"
)

// RemoveCmd creates the remove command.
func RemoveCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [KEY_NAME]",
		Short: "Remove a key or entire user vault",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := GetUsername(cfg)
			if err != nil {
				return err
			}

			svc, err := service.NewVaultService()
			if err != nil {
				return err
			}

			if !svc.VaultExists(username) {
				return fmt.Errorf("vault for user '%s' does not exist", username)
			}

			if len(args) == 1 {
				keyName := args[0]

				password, err := cfg.GetPassword("Enter password for vault: ")
				if err != nil {
					return fmt.Errorf("reading password: %w", err)
				}

				keys, vault, err := svc.LoadAndDecrypt(username, password)
				if err != nil {
					return err
				}

				if _, exists := keys[keyName]; !exists {
					return fmt.Errorf("key '%s' not found", keyName)
				}

				delete(keys, keyName)

				if err := svc.UpdateAndSave(username, password, vault, keys); err != nil {
					return err
				}

				fmt.Printf("Key '%s' removed successfully\n", keyName)
				return nil
			}

			response, err := cfg.ReadConfirm("Are you sure you want to delete entire vault? (yes/no): ")
			if err != nil {
				return fmt.Errorf("reading confirmation: %w", err)
			}

			if response != "yes" {
				fmt.Println("Vault deletion cancelled")
				return nil
			}

			if err := svc.DeleteVault(username); err != nil {
				return err
			}

			fmt.Printf("Vault for user '%s' deleted successfully\n", username)
			return nil
		},
	}
}
