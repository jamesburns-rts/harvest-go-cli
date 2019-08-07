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
	"github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var entriesUpdateProject string
var entriesUpdateTask string
var entriesUpdateHours string
var entriesUpdateNotes string
var entriesUpdateDate string
var entriesUpdateAppendNotes bool
var entriesUpdateAppendHours bool
var entriesUpdateSelectTask bool
var entriesUpdateConfirm bool
var entriesUpdateLast bool
var entriesUpdateClearNotes bool

var entriesUpdateCmd = &cobra.Command{
	Use:   "update [ENTRY_ID]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Update time entry",
	Long:  `Update time entry`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var op harvest.EntryUpdateOptions

		// select entry
		if op.Entry, err = entriesUpdateGetEntry(args, ctx); err != nil {
			return err
		}

		// project and task
		if entriesUpdateSelectTask {
			if op.ProjectId, op.TaskId, err = selectProjectAndTaskFrom(entriesUpdateProject, entriesUpdateTask, ctx); err != nil {
				return err
			}
			entriesUpdateProject = strconv.FormatInt(*op.ProjectId, 10)
			entriesUpdateTask = strconv.FormatInt(*op.TaskId, 10)
		}

		// append
		if entriesUpdateAppendHours {
			if h, err := types.ParseHours(entriesUpdateHours); err != nil {
				return errors.Wrap(err, "for --hours")
			} else {
				*h += op.Entry.Hours
				entriesUpdateHours = h.Duration().String()
			}
		}
		if entriesUpdateAppendNotes {
			entriesUpdateNotes = fmt.Sprintf("%s\n%s", op.Entry.Notes, entriesUpdateNotes)
		}

		// confirm
		if entriesUpdateConfirm {
			if err = entriesUpdateConfirmEntry(op.Entry); err != nil {
				return err
			}
		}

		// parse
		if op.TaskId, op.ProjectId, err = harvest.ParseTaskId(entriesUpdateTask); err != nil {
			return errors.Wrap(err, "for --task")
		}
		if entriesUpdateProject != "" {
			if op.ProjectId, err = harvest.ParseProjectId(entriesUpdateProject); err != nil {
				return errors.Wrap(err, "for --project")
			}
		}
		if op.Hours, err = types.ParseHours(entriesUpdateHours); err != nil {
			return errors.Wrap(err, "for --hours")
		}
		if entriesUpdateNotes != "" || entriesUpdateClearNotes {
			op.Notes = &entriesUpdateNotes
		}
		if op.Date, err = util.StringToDate(entriesUpdateDate); err != nil {
			return errors.Wrap(err, "for --date")
		}

		// update entry
		if entry, err := harvest.UpdateEntry(op, ctx); err != nil {
			return errors.Wrap(err, "problem updating entry")
		} else {
			return printWithFormat(outputMap{
				config.OutputFormatSimple: func() error { return outputSuccess() },
				config.OutputFormatTable:  func() error { return outputEntryTable(entry) },
				config.OutputFormatJson:   func() error { return outputJson(entry) },
			})
		}
	}),
}

func init() {
	entriesCmd.AddCommand(entriesUpdateCmd)
	entriesUpdateCmd.Flags().StringVarP(&entriesUpdateProject, "project", "p", "", "")
	entriesUpdateCmd.Flags().StringVarP(&entriesUpdateTask, "task", "t", "", "")
	entriesUpdateCmd.Flags().StringVarP(&entriesUpdateHours, "hours", "H", "", "")
	entriesUpdateCmd.Flags().StringVarP(&entriesUpdateNotes, "message", "m", "", "")
	entriesUpdateCmd.Flags().StringVarP(&entriesUpdateDate, "date", "d", "", "")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateAppendNotes, "append-notes", false, "")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateAppendHours, "append-hours", false, "")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateSelectTask, "select-task", false, "")
	entriesUpdateCmd.Flags().BoolVarP(&entriesUpdateConfirm, "confirm", "c", false, "")
	entriesUpdateCmd.Flags().BoolVarP(&entriesUpdateLast, "last", "l", false, "")
	entriesUpdateCmd.Flags().BoolVar(&entriesUpdateClearNotes, "clear-notes", false, "")
}

func entriesUpdateConfirmEntry(entry harvest.Entry) error {
	if entriesUpdateProject == "" {
		entriesUpdateProject = strconv.FormatInt(entry.Project.ID, 10)
	}
	if entriesUpdateTask == "" {
		entriesUpdateTask = strconv.FormatInt(entry.Task.ID, 10)
	}
	if entriesUpdateHours == "" {
		entriesUpdateHours = entry.Hours.Duration().String()
	}
	if entriesUpdateNotes == "" && !entriesUpdateClearNotes {
		entriesUpdateNotes = entry.Notes
	}
	if entriesUpdateDate == "" {
		entriesUpdateDate = entry.Date
	}

	fields := []prompt.Confirmation{
		{"Project", &entriesUpdateProject, validProjectId},
		{"Task", &entriesUpdateTask, validTaskId},
		{"Hours", &entriesUpdateHours, validHours},
		{"Notes", &entriesUpdateNotes, validNotes},
		{"Date", &entriesUpdateDate, validDate},
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
		entries, err := harvest.GetEntries(nil, ctx)
		if err != nil {
			return entry, err
		}
		if entriesUpdateLast {
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
