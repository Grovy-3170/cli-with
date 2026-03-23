package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Grovy-3170/cli-with/internal/commands"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Interfaces for dependency injection (used by tests)
type passwordReader interface {
	ReadPassword(prompt string) (string, error)
}

type confirmationReader interface {
	ReadConfirmation(prompt string) (string, error)
}

// Terminal implementations
type terminalPasswordReader struct{}

func (t terminalPasswordReader) ReadPassword(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s", prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd())) // #nosec G115
	fmt.Fprintln(os.Stderr)
	return string(password), err
}

type terminalConfirmationReader struct{}

func (t terminalConfirmationReader) ReadConfirmation(prompt string) (string, error) {
	fmt.Fprintf(os.Stderr, "%s", prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(response), nil
}

// Global state (replaceable for testing)
var (
	user              string
	passwordFile      string
	keyValue          string
	passwordInput     passwordReader     = terminalPasswordReader{}
	confirmationInput confirmationReader = terminalConfirmationReader{}
)

func getPassword(prompt string) (string, error) {
	if passwordFile != "" {
		data, err := os.ReadFile(passwordFile) // #nosec G304
		if err != nil {
			return "", fmt.Errorf("reading password file: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return passwordInput.ReadPassword(prompt)
}

// newRootCmd creates and returns the root command with all subcommands registered.
// This is also used by tests.
func newRootCmd() *cobra.Command {
	// Create command config with dependencies
	cfg := &commands.Config{
		User:         &user,
		PasswordFile: &passwordFile,
		KeyValue:     &keyValue,
		GetPassword:  getPassword,
		ReadPassword: passwordInput.ReadPassword,
		ReadConfirm:  confirmationInput.ReadConfirmation,
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "with",
		Short: "Run any command with your secrets",
		Long: `Run any command with your secrets. No leaks. No drama.

Store your API keys and secrets in an encrypted vault, then run commands
with them injected as environment variables — isolated to the subprocess only.`,
		Version:      Version,
		SilenceUsage: true,
	}

	// Register global flags
	rootCmd.PersistentFlags().StringVar(&user, "user", "", "Username for the vault")
	rootCmd.PersistentFlags().StringVar(&passwordFile, "password-file", "", "Path to password file")

	// Register all commands
	rootCmd.AddCommand(
		commands.InitCmd(cfg),
		commands.SetCmd(cfg),
		commands.ListCmd(cfg),
		commands.GetCmd(cfg),
		commands.RemoveCmd(cfg),
		commands.ExecCmd(cfg),
		commands.VersionCmd(Version),
	)

	return rootCmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
