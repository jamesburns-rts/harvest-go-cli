package cmd

import (
	"context"
	"fmt"

	"errors"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var timersDeleteCmd = &cobra.Command{
	Use:               "delete NAME",
	Args:              cobra.MinimumNArgs(1),
	Short:             "Delete a timer",
	Long:              `Delete a timer`,
	ValidArgsFunction: timerCompletionFunc(timerCompletionOptions{}),
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		for _, name := range args {
			if t, ok := timers.Get(name); ok {
				timers.Delete(name)
				_ = printWithFormat(outputMap{
					config.OutputFormatSimple: func() error { return timersDeleteSimple(t) },
					config.OutputFormatTable:  func() error { return timersDeleteSimple(t) },
					config.OutputFormatJson:   func() error { return outputJson(t) },
				})
			} else {
				return errors.New("no timer exists")
			}
		}
		return writeConfig()
	}),
}

func timersDeleteSimple(t timers.Timer) error {
	fmt.Printf("Timer %s deleted at %s\n", t.Name, fmtHours(t.RunningHours()))
	return nil
}

func init() {
	timersCmd.AddCommand(timersDeleteCmd)
}
