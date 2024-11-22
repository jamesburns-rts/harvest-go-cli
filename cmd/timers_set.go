package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
	"time"
)

var timersSetTask taskArg
var timersSetEntryId int64
var timersSetNotes string
var timersSetHours hoursArg
var timersSetAdd hoursArg

var timersSetCmd = &cobra.Command{
	Use:               "set NAME",
	Args:              cobra.RangeArgs(1, 2),
	Short:             "Set/alter values of a timer",
	Long:              `Set/alter values of a timer`,
	ValidArgsFunction: timerCompletionFunc(timerCompletionOptions{}),
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		name := args[0]
		if len(args) > 1 && timersSetHours.str == "" {
			if err := timersSetHours.Set(args[1]); err != nil {
				return err
			}
		}

		if timer, ok := timers.Get(name); ok {
			timer.Notes += timersSetNotes

			if timersSetHours.hours != nil {
				timer.Duration = *timersSetHours.hours
				if timer.Running {
					timer.SetStarted(time.Now())
				}
			} else if timersSetAdd.hours != nil {
				timer.Duration += *timersSetAdd.hours
			}

			timers.Set(timer)

			_ = printWithFormat(outputMap{
				config.OutputFormatSimple: func() error { return timersSetSimple(timer) },
				config.OutputFormatTable:  func() error { return timersSetSimple(timer) },
				config.OutputFormatJson: func() error {
					timer.Duration = *timer.RunningHours()
					return outputJson(timer)
				},
			})

			return writeConfig()

		} else {
			return timersStart(name, timersSetNotes, timersSetTask, timersSetEntryId, timersSetHours, ctx)
		}
	}),
}

func timersSetSimple(t timers.Timer) error {
	if t.Running {
		return timersStartSimple(t)
	}
	fmt.Printf("Timer %s set to %s\n", t.Name, fmtHours(&t.Duration))
	return nil
}

func init() {
	timersCmd.AddCommand(timersSetCmd)
	timersSetCmd.Flags().VarP(&timersSetTask, "task", "t",
		"Associate timer with a task and sync the timer with harvest")
	timersSetCmd.Flags().Int64VarP(&timersSetEntryId, "entry", "e", -1,
		"Associate timer with a time entry and sync the timer with harvest")
	timersSetCmd.Flags().StringVarP(&timersSetNotes, "notes", "n", "", "Append notes to the timer")
	timersSetCmd.Flags().VarP(&timersSetHours, "hours", "H", "Set the duration of the timer")
	timersSetCmd.Flags().VarP(&timersSetAdd, "add", "a", "Add a duration (or negative duration) to the timer")

	_ = timersSetCmd.RegisterFlagCompletionFunc("task", taskCompletionFunc)
}
