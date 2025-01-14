package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
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

		userId, err := getAndSaveUserId(ctx)
		if err != nil {
			return err
		}

		options := harvest.EntryListOptions{
			To:        entriesToDate.date,
			From:      entriesFromDate.date,
			TaskId:    entriesTask.taskId,
			ProjectId: entriesTask.projectId,
			UserId:    userId,
		}

		if entriesProject.projectId != nil {
			options.ProjectId = entriesProject.projectId
		}

		if len(args) > 0 {
			var onDate *time.Time
			if onDate, err = util.StringToDate(args[0]); err != nil {
				return fmt.Errorf("for [DATE]: %w", err)
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

	_ = entriesCmd.RegisterFlagCompletionFunc("project", projectCompletionFunc)
	_ = entriesCmd.RegisterFlagCompletionFunc("task", taskCompletionFunc)
}
