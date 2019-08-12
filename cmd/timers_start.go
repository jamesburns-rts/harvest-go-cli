package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
	"time"
)

var timersStartTask taskArg
var timersStartEntryId int64
var timersStartNotes string

var timersStartCmd = &cobra.Command{
	Use:   "start NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Start a timer",
	Long:  `Start a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		return timersStart(args[0], timersStartNotes, timersStartTask, timersStartEntryId, ctx)
	}),
}

func timersStart(name, notes string, task taskArg, entryId int64, ctx context.Context) error {
	var t timers.Timer
	var exists bool
	if t, exists = timers.Records.Timers[name]; !exists {
		t = timers.Timer{
			Name:  name,
			Notes: notes,
			//SyncedTaskId *int64 `yaml,json:"syncedTaskId"`
		}
		t.SetStarted(time.Now())
	} else {
		t.Notes += notes
	}

	if task.str != "" {
		// todo check for change?
		t.SyncedProjectId = task.projectId
		t.SyncedTaskId = task.taskId
	}

	if entryId > 0 {
		// todo check for change?
		t.SyncedEntryId = &entryId
	}

	if err := t.Start(timersDoNotSync, ctx); err != nil {
		return err
	}

	timers.SetTimer(t)

	_ = printWithFormat(outputMap{
		config.OutputFormatSimple: func() error { return timersStartSimple(exists, t) },
		config.OutputFormatTable:  func() error { return timersStartSimple(exists, t) },
		config.OutputFormatJson: func() error {
			t.Duration = *t.RunningHours()
			return outputJson(t)
		},
	})

	return writeConfig()
}

func timersStartSimple(exists bool, t timers.Timer) error {
	resumedStr := "started"
	if exists {
		resumedStr = fmt.Sprintf("resumed from %s", fmtHours(&t.Duration))
	}
	fmt.Printf("Timer %s %s at %s\n", t.Name, resumedStr, t.StartedTime().Format(time.Kitchen))
	return nil
}

func init() {
	timersCmd.AddCommand(timersStartCmd)
	timersStartCmd.Flags().VarP(&timersStartTask, "task", "t",
		"Associate timer with a task and sync the timer with harvest")
	timersStartCmd.Flags().Int64VarP(&timersStartEntryId, "entry", "e", -1,
		"Associate timer with a time entry and sync the timer with harvest")
	timersStartCmd.Flags().StringVarP(&timersStartNotes, "notes", "n", "", "Append notes to the timer")
}
