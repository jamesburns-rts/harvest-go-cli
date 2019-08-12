package cmd

import (
	"context"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var timersStopDoNotSync bool

var timersStopCmd = &cobra.Command{
	Use:   "stop NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Stop a timer",
	Long:  `Stop a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		name := args[0]

		if existing, ok := timers.Records.Timers[name]; ok {
			if err := existing.Stop(false, ctx); err != nil {
				return err
			}
		} else {
			return errors.New("no timer exists")
		}

		return writeConfig()
	}),
}

func init() {
	timersCmd.AddCommand(timersStopCmd)
	timersStopCmd.Flags().BoolVar(&timersStopDoNotSync, "do-not-sync", false, "Prevent syncing with harvest timers")
}
