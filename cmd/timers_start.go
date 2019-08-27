package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/spf13/cobra"
	"time"
)

var timersStartTask taskArg
var timersStartEntryId int64
var timersStartNotes string
var timersStartHours hoursArg

var timersStartCmd = &cobra.Command{
	Use:   "start NAME",
	Args:  cobra.ExactArgs(1),
	Short: "Start a timer",
	Long:  `Start a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		return timersStart(args[0], timersStartNotes, timersStartTask, timersStartEntryId, timersStartHours, ctx)
	}),
}

func timersStart(name, notes string, task taskArg, entryId int64, hours hoursArg, ctx context.Context) error {
	var t timers.Timer
	var exists bool
	if t, exists = timers.Get(name); !exists {
		t = timers.Timer{
			Name:  name,
			Notes: notes,
			//SyncedTaskId *int64 `yaml,json:"syncedTaskId"`
		}
		t.SetStarted(time.Now())
	} else if t.Notes == "" {
		t.Notes = notes
	} else {
		t.Notes += "\n" + notes
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

	if hours.hours != nil {
		t.Duration += *hours.hours
	}

	if err := t.Start(timersDoNotSync, ctx); err != nil {
		return err
	}

	timers.Set(t)

	_ = printWithFormat(outputMap{
		config.OutputFormatSimple: func() error { return timersStartSimple(t) },
		config.OutputFormatTable:  func() error { return timersStartSimple(t) },
		config.OutputFormatJson: func() error {
			t.Duration = *t.RunningHours()
			return outputJson(t)
		},
	})

	return writeConfig()
}

func timersStartSimple(t timers.Timer) error {
	resumedStr := "started"
	if t.Duration > types.Hours(0) {
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
	timersStartCmd.Flags().VarP(&timersStartHours, "hours", "H",
		"Start the timer with the given hours already clocked (or appended)")
}
