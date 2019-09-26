package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:       "completion [SHELL]",
	Args:      cobra.ExactArgs(1),
	Short:     "Generate a shell completion script",
	Long:      `Generate a shell completion script`,
	ValidArgs: []string{"bash", "zsh", "PowerShell"},
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "PowerShell":
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		case "zsh":
			fallthrough
		default:
			return rootCmd.GenZshCompletion(os.Stdout)
		}
	}),
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
