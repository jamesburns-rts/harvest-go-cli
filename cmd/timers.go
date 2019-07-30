package cmd

import (
	"context"
	"github.com/spf13/cobra"
)

var timersCmd = &cobra.Command{
	Use:   "timers",
	Short: "List timers",
	Long:  `List timers`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(timersCmd)
}
