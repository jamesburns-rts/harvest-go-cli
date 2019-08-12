package cmd

import (
	"context"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
	"time"
)

var timersStartTask taskArg
var timersStartEntryId int64
var timersStartDoNotSync bool
var timersStartNotes string

var timersStartCmd = &cobra.Command{
	Use:   "start NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Start a timer",
	Long:  `Start a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		name := args[0]

		if existing, ok := timers.Records.Timers[name]; ok {
			existing.Running = true
			existing.SetStarted(time.Now())
			existing.Notes += timersStartNotes
		} else {
			t := timers.Timer{
				Name:    name,
				Running: true,
				//SyncedTaskId *int64 `yaml,json:"syncedTaskId"`
				Notes: timersStartNotes,
			}
			t.SetStarted(time.Now())
		}

		return writeConfig()
	}),
}

func init() {
	rootCmd.AddCommand(timersStartCmd)
	timersStartCmd.Flags().VarP(&timersStartTask, "task", "t",
		"Associate timer with a task and sync the timer with harvest")
	timersStartCmd.Flags().Int64VarP(&timersStartEntryId, "entry", "e", -1,
		"Associate timer with a time entry and sync the timer with harvest")
	timersStartCmd.Flags().BoolVar(&timersStartDoNotSync, "do-not-sync", false, "Prevent syncing with harvest timers")
	timersStartCmd.Flags().StringVarP(&timersStartNotes, "notes", "n", "", "Append notes to the timer")
}
