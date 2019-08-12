package cmd

import (
	"context"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var timersCmd = &cobra.Command{
	Use:   "timers",
	Short: "List timers",
	Long:  `List timers`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		return outputJson(timers.Records)
	}),
}

func init() {
	rootCmd.AddCommand(timersCmd)
}
