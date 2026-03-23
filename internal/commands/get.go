package commands

import (
	"fmt"

	"github.com/Grovy-3170/cli-with/internal/service"
	"github.com/spf13/cobra"
)

// GetCmd creates the get command.
func GetCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "get [KEY_NAME]",
		Short: "Get the value of a specific key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := GetUsername(cfg)
			if err != nil {
				return err
			}

			keyName := args[0]

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

			value, exists := keys[keyName]
			if !exists {
				return fmt.Errorf("key '%s' not found", keyName)
			}

			fmt.Println(value)
			return nil
		},
	}
}
