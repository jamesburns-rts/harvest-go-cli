package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"time"
)

var logMessage string
var logDate string

var logCmd = &cobra.Command{
	Use:   "log [task] [duration]",
	Args:  cobra.ExactArgs(2),
	Short: "Log a time entry",
	Long:  `Log a time entry`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		taskId, projectId, err := getTaskAndProjectId(args[0])
		if err != nil {
			return err
		}

		if projectId == nil {
			if projectId, err = getTaskProjectId(*taskId, ctx); err != nil {
				return err
			}
		}

		date, err := util.StringToDate(logDate)
		if err != nil {
			return err
		}

		duration, err := time.ParseDuration(args[1])
		if err != nil {
			return errors.Wrap(err, "problem parsing duration")
		}

		entry, err := harvest.LogTime(harvest.LogTimeOptions{
			TaskId:    *taskId,
			ProjectId: *projectId,
			Date:      *date,
			Hours:     Hours(duration.Hours()),
			Notes:     logMessage,
		}, ctx)

		if err != nil {
			return err
		}

		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return logOutputSimple() },
			config.OutputFormatTable:  func() error { return logOutputTable(entry) },
			config.OutputFormatJson:   func() error { return outputJson(entry) },
		})
	}),
}

func logOutputSimple() error {
	fmt.Println("Successful")
	return nil
}

func logOutputTable(entry harvest.Entry) error {
	table := createTable([]string{"Key", "Value"})
	table.AppendBulk([][]string{
		{"ID", fmt.Sprintf("%v", entry.ID),},
		{"Project", entry.Project.Name,},
		{"Task", entry.Task.Name,},
		{"Message", entry.Notes,},
		{"Date", entry.Date,},
		{"Hours", entry.Hours.String(),},
	})
	table.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVarP(&logMessage, "message", "m", "", "Add notes to the time entry")
	logCmd.Flags().StringVarP(&logDate, "date", "d", "today", "Set the date for the entry")
}
