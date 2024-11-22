package cmd

import (
	"context"
	"fmt"

	"errors"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var timersStopHours hoursArg
var timersStopNotes string

var timersStopCmd = &cobra.Command{
	Use:   "stop NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Stop a timer",
	Long:  `Stop a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		return timersStop(args[0], timersStopHours, timersStopNotes, ctx)

	}),
}

func timersStop(name string, hours hoursArg, notesArg string, ctx context.Context) error {

	if t, ok := timers.Get(name); ok {
		if err := t.Stop(timersDoNotSync, ctx); err != nil {
			return err
		}

		if hours.hours != nil {
			t.Duration += *hours.hours
		}

		t.AppendNotes(notesArg)

		_ = printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return timersStopSimple(t) },
			config.OutputFormatTable:  func() error { return timersStopSimple(t) },
			config.OutputFormatJson:   func() error { return outputJson(t) },
		})

		return writeConfig()
	} else {
		return errors.New("no timer exists")
	}
}

func timersStopSimple(t timers.Timer) error {
	fmt.Printf("Timer %s stopped at %s\n", t.Name, fmtHours(t.RunningHours()))
	return nil
}

func init() {
	timersCmd.AddCommand(timersStopCmd)
	timersStopCmd.Flags().VarP(&timersStopHours, "hours", "H", "Stop the timer with the given hours appended")
	timersStopCmd.Flags().StringVarP(&timersStopNotes, "notes", "n", "", "Append notes to the timer")
}
