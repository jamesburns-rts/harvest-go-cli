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
	"encoding/json"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var entriesProject string
var entriesTask string
var entriesToDate string
var entriesFromDate string

var entriesCmd = &cobra.Command{
	Use:     "entries",
	Aliases: []string{"list"},
	Short:   "List time entries",
	Long:    `List time entries you have entered already`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		var err error
		options := harvest.EntryListOptions{}
		if options.ProjectId, err = harvest.GetProjectId(entriesProject); err != nil {
			return errors.Wrap(err, "for --project: ")
		}
		if options.TaskId, err = harvest.GetTaskId(entriesTask); err != nil {
			return errors.Wrap(err, "for --task: ")
		}
		if options.To, err = util.StringToDate(entriesToDate); err != nil {
			return errors.Wrap(err, "for --to: ")
		}
		if options.From, err = util.StringToDate(entriesFromDate); err != nil {
			return errors.Wrap(err, "for --from: ")
		}

		entries, err := harvest.GetEntries(&options, ctx)
		if err != nil {
			return err
		}


		// print
		format := getOutputFormat()
		if format == config.OutputFormatSimple {
			for _, entry := range entries {
				fmt.Printf("%v %v %v %s %0.2f %v\n", entry.ID, entry.Project.ID, entry.Task.ID, entry.Date, entry.Hours, entry.Notes)
			}

		} else if format == config.OutputFormatJson {
			b, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return errors.Wrap(err, "problem marshalling entries to json")
			}
			fmt.Println(string(b))

		} else if format == config.OutputFormatTable {
			table := createTable([]string{"ID", "Project Name", "Date", "Task Name", "Hours", "Notes"})
			for _, entry := range entries {

				table.Append([]string{
					strconv.Itoa(int(entry.ID)),
					entry.Project.Name,
					entry.Date,
					entry.Task.Name,
					entry.Hours.String(),
					entry.Notes,
				})
			}
			table.Render()

		} else {
			return errors.New("unrecognized --format " + format)
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(entriesCmd)
	entriesCmd.Flags().StringVarP(&entriesProject, "project", "p", "", "Project ID/alias by which to filter")
	entriesCmd.Flags().StringVarP(&entriesTask, "task", "t", "", "Task ID/alias by which to filter")
	entriesCmd.Flags().StringVar(&entriesToDate, "to", "", "Date by which to filter by entries on or before [see date section in root]")
	entriesCmd.Flags().StringVar(&entriesFromDate, "from", "", "Date by which to filter by entries on or after [see date section in root]")
}
