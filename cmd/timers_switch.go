package cmd

import (
	"context"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var timersSwitchTask taskArg
var timersSwitchEntryId int64
var timersSwitchNotes string

var timersSwitchCmd = &cobra.Command{
	Use:   "switch NAME",
	Args:  cobra.MaximumNArgs(1),
	Short: "Switch a timer",
	Long:  `Switch a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		// stop all other timers
		for _, t := range timers.Records.Timers {
			if t.Running && t.Name != name {
				if err := timersStop(t.Name, ctx); err != nil {
					return err
				}
			}
		}

		if name != "" {
			if existing, ok := timers.Records.Timers[name]; !ok || !existing.Running {
				return timersStart(name, timersSwitchNotes, timersSwitchTask, timersSwitchEntryId, ctx)
			} else {
				return timersStop(name, ctx)
			}
		}
		return nil
	}),
}

func init() {
	timersCmd.AddCommand(timersSwitchCmd)
	timersSwitchCmd.Flags().VarP(&timersStartTask, "task", "t",
		"Associate timer with a task and sync the timer with harvest")
	timersSwitchCmd.Flags().Int64VarP(&timersStartEntryId, "entry", "e", -1,
		"Associate timer with a time entry and sync the timer with harvest")
	timersSwitchCmd.Flags().StringVarP(&timersStartNotes, "notes", "n", "", "Append notes to the timer")
}
