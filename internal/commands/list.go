package commands

import (
	"fmt"

	"github.com/Grovy-3170/cli-with/internal/service"
	"github.com/spf13/cobra"
)

// ListCmd creates the list command.
func ListCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all keys for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := GetUsername(cfg)
			if err != nil {
				return err
			}

			svc, err := service.NewVaultService()
			if err != nil {
				return err
			}

			password, err := cfg.GetPassword("Enter password for vault: ")
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}

			keys, _, err := svc.LoadAndDecrypt(username, password)
			if err != nil {
				return err
			}

			if len(keys) == 0 {
				fmt.Println("No keys stored")
				return nil
			}

			for name := range keys {
				fmt.Println(name)
			}

			return nil
		},
	}
}
