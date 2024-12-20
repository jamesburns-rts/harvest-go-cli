package cmd

import (
	"context"

	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var timersSwitchTask taskArg
var timersSwitchEntryId int64
var timersSwitchNotes string
var timersSwitchHours hoursArg

var timersSwitchCmd = &cobra.Command{
	Use:               "switch NAME",
	Args:              cobra.MaximumNArgs(1),
	Short:             "Switch a timer",
	Long:              `Switch a timer`,
	ValidArgsFunction: timerCompletionFunc(timerCompletionOptions{}),
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		// stop all other timers
		for _, t := range timers.Records.Timers {
			if t.Running && t.Name != name {
				if err := timersStop(t.Name, hoursArg{}, "", ctx); err != nil {
					return err
				}
			}
		}

		if name != "" {
			if existing, ok := timers.Get(name); !ok || !existing.Running {
				return timersStart(name, timersSwitchNotes, timersSwitchTask, timersSwitchEntryId, timersSwitchHours, ctx)
			} else {
				return timersStop(name, timersSwitchHours, timersSwitchNotes, ctx)
			}
		}
		return nil
	}),
}

func init() {
	timersCmd.AddCommand(timersSwitchCmd)
	timersSwitchCmd.Flags().VarP(&timersSwitchTask, "task", "t",
		"Associate timer with a task and sync the timer with harvest")
	timersSwitchCmd.Flags().Int64VarP(&timersSwitchEntryId, "entry", "e", -1,
		"Associate timer with a time entry and sync the timer with harvest")
	timersSwitchCmd.Flags().StringVarP(&timersSwitchNotes, "notes", "n", "", "Append notes to the timer")
	timersSwitchCmd.Flags().VarP(&timersSwitchHours, "hours", "H",
		"Start/stop the timer with the given hours appended")

	_ = timersSwitchCmd.RegisterFlagCompletionFunc("task", taskCompletionFunc)
}
