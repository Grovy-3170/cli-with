package commands

import "github.com/spf13/cobra"

// VersionCmd creates the version command.
func VersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version)
		},
	}
}
