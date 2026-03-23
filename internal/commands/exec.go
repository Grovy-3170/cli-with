package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/Grovy-3170/cli-with/internal/service"
	"github.com/spf13/cobra"
)

// minimalEnvVars is the list of environment variables to inherit from the parent process
var minimalEnvVars = []string{"PATH", "HOME", "USER", "SHELL", "TMPDIR"}

// ExecCmd creates the exec command.
func ExecCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "exec -- [command]",
		Short: "Execute a command with keys as environment variables",
		Args:  cobra.MinimumNArgs(1),
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

			password, err := cfg.GetPassword("Enter password for vault: ")
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}

			keys, _, err := svc.LoadAndDecrypt(username, password)
			if err != nil {
				return err
			}

			env := BuildMinimalEnv(keys)

			execCmd := exec.Command(args[0], args[1:]...) // #nosec G204
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			execCmd.Env = env

			if err := execCmd.Start(); err != nil {
				return fmt.Errorf("starting command: %w", err)
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan)

			waitChan := make(chan error, 1)
			go func() {
				waitChan <- execCmd.Wait()
			}()

			for {
				select {
				case sig := <-sigChan:
					if execCmd.Process != nil {
						_ = execCmd.Process.Signal(sig)
					}
				case err := <-waitChan:
					signal.Stop(sigChan)

					if exitErr, ok := err.(*exec.ExitError); ok {
						if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
							os.Exit(status.ExitStatus())
						}
						os.Exit(1)
					}
					if err != nil {
						return fmt.Errorf("waiting for command: %w", err)
					}
					os.Exit(0)
				}
			}
		},
	}
}

// BuildMinimalEnv creates an environment with minimal system vars plus the given keys.
func BuildMinimalEnv(keys map[string]string) []string {
	env := make([]string, 0, len(minimalEnvVars)+len(keys))

	for _, key := range minimalEnvVars {
		if value, exists := os.LookupEnv(key); exists {
			env = append(env, key+"="+value)
		}
	}

	for key, value := range keys {
		env = append(env, key+"="+value)
	}

	return env
}
