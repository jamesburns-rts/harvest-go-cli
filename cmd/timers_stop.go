package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var timersStopCmd = &cobra.Command{
	Use:   "stop NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Stop a timer",
	Long:  `Stop a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		return timersStop(args[0], ctx)

	}),
}

func timersStop(name string, ctx context.Context) error {

	if t, ok := timers.Records.Timers[name]; ok {
		if err := t.Stop(timersDoNotSync, ctx); err != nil {
			return err
		}
		_ = printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return timersStopSimple(t) },
			config.OutputFormatTable:  func() error { return timersStopSimple(t) },
			config.OutputFormatJson:   func() error { return outputJson(t) },
		})
	} else {
		return errors.New("no timer exists")
	}

	return writeConfig()
}

func timersStopSimple(t timers.Timer) error {
	fmt.Printf("Timer %s stopped at %s\n", t.Name, fmtHours(t.RunningHours()))
	return nil
}

func init() {
	timersCmd.AddCommand(timersStopCmd)
}
