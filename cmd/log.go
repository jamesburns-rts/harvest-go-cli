package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"time"
)

var logTask string
var logMessage string
var logDate string
var logDuration string

var logCmd = &cobra.Command{
	Use:   "log [TASK] [DURATION]",
	Args:  cobra.MaximumNArgs(2),
	Short: "Log a time entry",
	Long:  `Log a time entry`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var taskId *int64
		var projectId *int64
		var date *time.Time
		var duration *Hours

		if logTask == "" && len(args) > 0 {
			logTask = args[0]
		}
		if logDuration == "" && len(args) > 1 {
			logDuration = args[1]
		}

		// select project/task

		if logTask != "" {
			if taskId, projectId, err = parseTaskAndProjectId(logTask); err != nil {
				return errors.Wrap(err, "for [task]")
			}
			if projectId == nil {
				if projectId, err = getTaskProjectId(*taskId, ctx); err != nil {
					return err
				}
			}
		} else {
			if projectId, taskId, err = selectProjectAndTask(ctx); err != nil {
				return err
			}
		}

		// get date
		if date, err = util.StringToDate(logDate); err != nil {
			return errors.Wrap(err, "for --date")
		}

		// get defaults
		if alias, ok := config.Harvest.TaskAliases[logTask]; ok {
			duration = alias.DefaultDuration
			if alias.DefaultNotes != nil && logMessage == "" {
				logMessage = *alias.DefaultNotes
			}
		}

		// get duration
		if logDuration == "" {
			logDuration, err = prompt.ForStringWithValidation("Duration", func(s string) error {
				_, e := ParseHours(s)
				return e
			})
			if err != nil {
				return err
			}
		}

		duration = new(Hours)
		if *duration, err = ParseHours(logDuration); err != nil {
			return errors.Wrap(err, "for [duration]")
		}

		// get message
		if logMessage == "" {
			if logMessage, err = prompt.ForString("Notes"); err != nil {
				return err
			}
		}

		// log time
		entry, err := harvest.LogTime(harvest.LogTimeOptions{
			TaskId:    *taskId,
			ProjectId: *projectId,
			Date:      *date,
			Hours:     *duration,
			Notes:     logMessage,
		}, ctx)

		if err != nil {
			return err
		}

		// print
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
		{"ID", fmt.Sprintf("%v", entry.ID)},
		{"Project", entry.Project.Name},
		{"Task", entry.Task.Name},
		{"Message", entry.Notes},
		{"Date", entry.Date},
		{"Hours", fmtHours(&entry.Hours)},
	})
	table.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVarP(&logTask, "task", "t", "", "Set the task")
	logCmd.Flags().StringVarP(&logMessage, "message", "m", "", "Add notes to the time entry")
	logCmd.Flags().StringVarP(&logDate, "date", "d", "today", "Set the date for the entry")
	logCmd.Flags().StringVarP(&logDuration, "duration", "D", "", "Set the duration for the entry")
}
