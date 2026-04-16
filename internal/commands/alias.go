package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/Grovy-3170/cli-with/internal/aliases"
	"github.com/Grovy-3170/cli-with/internal/config"
	"github.com/spf13/cobra"
)

// AliasCmd creates the `with alias` command and its subcommands.
func AliasCmd(cfg *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage saved shortcuts for `with exec`",
		Long: `Manage saved shortcuts for ` + "`with exec`" + `.

Use ` + "`with alias add`" + ` to register a named shortcut, then activate all saved
shortcuts as native shell aliases by adding one line to your shell config:

    eval "$(with alias shell)"        # bash/zsh
    with alias shell --shell fish | source  # fish`,
	}
	cmd.AddCommand(
		aliasAddCmd(cfg),
		aliasListCmd(),
		aliasRemoveCmd(),
		aliasShellCmd(),
	)
	return cmd
}

func loadStore() (*aliases.Store, string, error) {
	c, err := config.Load()
	if err != nil {
		return nil, "", fmt.Errorf("loading config: %w", err)
	}
	path := c.GetAliasPath()
	s, err := aliases.Load(path)
	if err != nil {
		return nil, "", err
	}
	return s, path, nil
}

func aliasAddCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "add NAME -- COMMAND [ARGS...]",
		Short: "Save a named `with exec` shortcut",
		Long: `Save a named ` + "`with exec`" + ` shortcut. Global flags (--user, --password,
--password-file) are captured at save time; the command after -- is what the
alias will run.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := ValidateKeyName(name); err != nil {
				return err
			}
			command := args[1:]

			store, path, err := loadStore()
			if err != nil {
				return err
			}

			a := aliases.Alias{
				Command:      command,
				User:         *cfg.User,
				PasswordFile: *cfg.PasswordFile,
			}
			if cfg.PasswordChanged != nil && cfg.PasswordChanged() {
				a.Password = *cfg.Password
				a.PasswordSet = true
			}
			store.Set(name, a)

			if err := store.Save(path); err != nil {
				return err
			}
			fmt.Printf("Alias '%s' saved\n", name)
			return nil
		},
	}
}

func aliasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all saved aliases",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, _, err := loadStore()
			if err != nil {
				return err
			}
			names := store.Names()
			if len(names) == 0 {
				fmt.Println("No aliases saved")
				return nil
			}
			for _, n := range names {
				a := store.Aliases[n]
				line := fmt.Sprintf("%s → %s", n, strings.Join(a.Command, " "))
				if a.User != "" {
					line += fmt.Sprintf(" [user: %s]", a.User)
				}
				fmt.Println(line)
			}
			return nil
		},
	}
}

func aliasRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove NAME",
		Short: "Remove a saved alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			store, path, err := loadStore()
			if err != nil {
				return err
			}
			if !store.Remove(name) {
				return fmt.Errorf("alias '%s' not found", name)
			}
			if err := store.Save(path); err != nil {
				return err
			}
			fmt.Printf("Alias '%s' removed\n", name)
			return nil
		},
	}
}

func aliasShellCmd() *cobra.Command {
	var shellFlag string
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Print shell alias lines for all saved aliases",
		Long: `Print shell alias lines for all saved aliases on stdout.

For bash/zsh, add this to ~/.bashrc or ~/.zshrc:

    eval "$(with alias shell)"

For fish, add to ~/.config/fish/config.fish:

    with alias shell --shell fish | source`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, _, err := loadStore()
			if err != nil {
				return err
			}
			shell := detectShell(shellFlag)
			for _, name := range store.Names() {
				fmt.Println(aliases.ShellLine(name, store.Aliases[name], shell))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&shellFlag, "shell", "", "Shell dialect: bash, zsh, or fish (default: autodetect from $SHELL)")
	return cmd
}

func detectShell(flag string) aliases.ShellType {
	name := strings.ToLower(flag)
	if name == "" {
		s := os.Getenv("SHELL")
		switch {
		case strings.Contains(s, "fish"):
			name = "fish"
		case strings.Contains(s, "zsh"):
			name = "zsh"
		default:
			name = "bash"
		}
	}
	switch name {
	case "fish":
		return aliases.ShellFish
	case "zsh":
		return aliases.ShellZsh
	default:
		return aliases.ShellBash
	}
}
