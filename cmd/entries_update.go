/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var entriesUpdateProject projectArg
var entriesUpdateTask taskArg
var entriesUpdateDuration hoursArg
var entriesUpdateNotes stringArg
var entriesUpdateDate dateArg
var entriesUpdateAppendNotes bool
var entriesUpdateAppendHours bool
var entriesUpdateSelectTask bool
var entriesUpdateConfirm bool
var entriesUpdateLast bool
var entriesUpdateLastOf taskArg
var entriesUpdateClearNotes bool

var entriesUpdateCmd = &cobra.Command{
	Use:   "update [ENTRY_ID]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Update time entry",
	Long: `Update time entry where the time entry chosen by either ENTRY_ID, --last, --last-of, 
or selection (if none provided)`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var op harvest.EntryUpdateOptions

		// select entry
		if op.Entry, err = entriesUpdateGetEntry(args, ctx); err != nil {
			return err
		}

		// project and task
		if entriesUpdateSelectTask {
			if op.ProjectId, op.TaskId, err = selectProjectAndTaskFrom(entriesUpdateProject.str, entriesUpdateTask.str, ctx); err != nil {
				return err
			}
			entriesUpdateProject.SetId(op.ProjectId)
			entriesUpdateTask.SetId(op.TaskId, op.ProjectId)
		}

		// append
		if entriesUpdateAppendHours {
			h := op.Entry.Hours
			h += *entriesUpdateDuration.hours
			entriesUpdateDuration.SetHours(&h)
		}
		if entriesUpdateAppendNotes {
			entriesUpdateNotes.str = fmt.Sprintf("%s\n%s", op.Entry.Notes, entriesUpdateNotes.str)
		}

		// confirm
		if entriesUpdateConfirm {
			if err = entriesUpdateConfirmEntry(op.Entry); err != nil {
				return err
			}
		}

		// parse
		op.TaskId = entriesUpdateTask.taskId
		op.ProjectId = entriesUpdateTask.projectId
		if entriesUpdateProject.projectId != nil {
			op.ProjectId = entriesUpdateProject.projectId
		}

		op.Hours = entriesUpdateDuration.hours
		op.Date = entriesUpdateDate.date
		if entriesUpdateNotes.str != "" || entriesUpdateClearNotes {
			op.Notes = &entriesUpdateNotes.str
		}

		// update entry
		if entry, err := harvest.UpdateEntry(op, ctx); err != nil {
			return errors.Wrap(err, "problem updating entry")
		} else {
			return printWithFormat(outputMap{
				config.OutputFormatSimple: func() error { return outputSuccess(entry) },
				config.OutputFormatTable:  func() error { return outputEntryTable(entry) },
				config.OutputFormatJson:   func() error { return outputJson(entry) },
			})
		}
	}),
}

func init() {
	entriesCmd.AddCommand(entriesUpdateCmd)
	entriesUpdateCmd.Flags().VarP(&entriesUpdateProject, "project", "p", "Project to move entry to")
	entriesUpdateCmd.Flags().VarP(&entriesUpdateTask, "task", "t", "Task to move entry to")
	entriesUpdateCmd.Flags().VarP(&entriesUpdateDuration, "duration", "D", "Duration to update entry's to (or append)")
	entriesUpdateCmd.Flags().VarP(&entriesUpdateNotes, "message", "m", "Message to update entry's to (or append)")
	entriesUpdateCmd.Flags().VarP(&entriesUpdateDate, "date", "d", "Date to update entry's to (see root's DATES section)")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateAppendNotes, "append-notes", false, "Append notes instead of replacing")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateAppendHours, "append-hours", false, "Append hours instead of replacing")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateSelectTask, "select-task", false, "Select project/task to update to")
	entriesUpdateCmd.Flags().BoolVarP(&entriesUpdateConfirm, "confirm", "c", false, "Confirm all fields before updating")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateLast, "last", false, "Update last time entry")
	entriesUpdateCmd.Flags().VarP(&entriesUpdateLastOf, "last-of", "l", "Update last time entry of given task")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateClearNotes, "clear-notes", false, "Set the notes to empty")
}

func entriesUpdateConfirmEntry(entry harvest.Entry) error {
	if entriesUpdateTask.str == "" {
		entriesUpdateTask.SetId(&entry.Task.ID, &entry.Project.ID)
	}
	if entriesUpdateProject.str == "" {
		entriesUpdateProject.SetId(&entry.Project.ID)
	}
	if entriesUpdateDuration.str == "" {
		entriesUpdateDuration.SetHours(&entry.Hours)
	}
	if entriesUpdateNotes.str == "" && !entriesUpdateClearNotes {
		entriesUpdateNotes.str = entry.Notes
	}
	if entriesUpdateDate.str == "" {
		_ = entriesUpdateDate.Set(entry.Date)
	}

	fields := []prompt.Confirmation{
		{"Project", &entriesUpdateProject},
		{"Task", &entriesUpdateTask},
		{"Duration", &entriesUpdateDuration},
		{"Notes", &entriesUpdateNotes},
		{"Date", &entriesUpdateDate},
	}
	if err := prompt.ConfirmAll(fields); err != nil {
		return err
	}
	return nil
}

func entriesUpdateGetEntry(args []string, ctx context.Context) (entry harvest.Entry, err error) {

	if len(args) > 0 {
		if entryId, err := strconv.ParseInt(args[0], 10, 64); err != nil {
			return entry, errors.Wrap(err, "problem with ENTRY_ID")
		} else {
			if entry, err = harvest.GetEntry(entryId, ctx); err != nil {
				return entry, err
			}
		}
	} else {
		op := &harvest.EntryListOptions{
			ProjectId: entriesUpdateLastOf.projectId,
			TaskId:    entriesUpdateLastOf.taskId,
		}
		entries, err := harvest.GetEntries(op, ctx)
		if err != nil {
			return entry, err
		}

		if entriesUpdateLastOf.str != "" || entriesUpdateLast {
			if len(entries) == 0 {
				return entry, errors.New("no entries found")
			}
			entry = entries[0]

		} else {
			selected, err := prompt.ForSelection("Select entry", entries)
			if err != nil {
				return entry, err
			}
			entry = entries[selected]
		}
	}
	return entry, nil
}
