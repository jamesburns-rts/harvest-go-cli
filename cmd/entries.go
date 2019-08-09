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
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

var entriesProject projectArg
var entriesTask taskArg
var entriesToDate dateArg
var entriesFromDate dateArg

var entriesCmd = &cobra.Command{
	Use:   "entries [DATE]",
	Args:  cobra.MaximumNArgs(1),
	Short: "List time entries",
	Long: `List time entries you have entered already. For formats of DATE, see the DATES 
section of the root command.`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		options := harvest.EntryListOptions{
			To:        entriesToDate.date,
			From:      entriesFromDate.date,
			TaskId:    entriesTask.taskId,
			ProjectId: entriesTask.projectId,
		}

		if entriesProject.projectId != nil {
			options.ProjectId = entriesProject.projectId
		}

		if len(args) > 0 {
			var onDate *time.Time
			if onDate, err = util.StringToDate(args[0]); err != nil {
				return errors.Wrap(err, "for [DATE]")
			}
			options.From = onDate
			options.To = onDate
		}

		// get entries
		entries, err := harvest.GetEntries(&options, ctx)
		if err != nil {
			return err
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return entriesOutputSimple(entries) },
			config.OutputFormatTable:  func() error { return entriesOutputTable(entries) },
			config.OutputFormatJson:   func() error { return outputJson(entries) },
		})
	}),
}

func entriesOutputSimple(entries []harvest.Entry) error {
	for _, entry := range entries {
		fmt.Printf("%v %v %v %s %0.2f %v\n", entry.ID, entry.Project.ID, entry.Task.ID, entry.Date, entry.Hours, entry.Notes)
	}
	return nil
}

func entriesOutputTable(entries []harvest.Entry) error {
	table := createTable([]string{"ID", "Project Name", "Date", "Task Name", "Hours", "Notes"})
	for _, entry := range entries {

		table.Append([]string{
			strconv.Itoa(int(entry.ID)),
			entry.Project.Name,
			entry.Date,
			entry.Task.Name,
			fmtHours(&entry.Hours),
			entry.Notes,
		})
	}
	table.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(entriesCmd)
	entriesCmd.Flags().VarP(&entriesProject, "project", "p", "Project ID/alias by which to filter")
	entriesCmd.Flags().VarP(&entriesTask, "task", "t", "Task ID/alias by which to filter")
	entriesCmd.Flags().Var(&entriesToDate, "to", "Date by which to filter by entries on or before [see date section in root]")
	entriesCmd.Flags().Var(&entriesFromDate, "from", "Date by which to filter by entries on or after [see date section in root]")
}
