package commands

import (
	"fmt"

	"github.com/Grovy-3170/cli-with/internal/service"
	"github.com/spf13/cobra"
)

// SetCmd creates the set command.
func SetCmd(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [KEY_NAME]",
		Short: "Set or update an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username, err := GetUsername(cfg)
			if err != nil {
				return err
			}

			keyName := args[0]

			if err := ValidateKeyName(keyName); err != nil {
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

			keys, vault, err := svc.LoadAndDecrypt(username, password)
			if err != nil {
				return err
			}

			var val string
			if *cfg.KeyValue != "" {
				val = *cfg.KeyValue
			} else {
				val, err = cfg.ReadPassword("Enter value for " + keyName + ": ")
				if err != nil {
					return fmt.Errorf("reading key value: %w", err)
				}
			}

			if keys == nil {
				keys = make(map[string]string)
			}
			keys[keyName] = val

			if err := svc.UpdateAndSave(username, password, vault, keys); err != nil {
				return err
			}

			fmt.Printf("Key '%s' set successfully\n", keyName)
			return nil
		},
	}

	cmd.Flags().StringVar(cfg.KeyValue, "value", "", "Value for the key (if not provided, will prompt)")
	return cmd
}
