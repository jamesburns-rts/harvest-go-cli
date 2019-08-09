package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var logTask taskArg
var logProject projectArg
var logNotes stringArg
var logDate dateArg
var logDuration hoursArg

var logCmd = &cobra.Command{
	Use:   "log [TASK [DURATION]]",
	Args:  cobra.MaximumNArgs(2),
	Short: "Log a time entry",
	Long: `Log a time entry for the given task. For TASK see root's ALIASES section and for 
DURATION see root's HOURS section. The task is selected from either TASK, --task, or 
--project.`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		if len(args) > 0 {
			if err = logTask.Set(args[0]); err != nil {
				return errors.Wrap(err, "for [TASK]")
			}
		}
		if len(args) > 1 {
			if err = logDuration.Set(args[1]); err != nil {
				return errors.Wrap(err, "for [DURATION]")
			}
		}

		// select project/task
		if logTask.projectId == nil {
			logTask.projectId = logProject.projectId
		}

		if logTask.taskId != nil {
			if logTask.projectId == nil {
				if logTask.projectId, err = getTaskProjectId(*logTask.taskId, ctx); err != nil {
					return err
				}
			}
		} else if logProject.projectId != nil {
			if logTask.taskId, err = selectTask(*logProject.projectId, ctx); err != nil {
				return err
			}
		} else {
			if logTask.projectId, logTask.taskId, err = selectProjectAndTask(ctx); err != nil {
				return err
			}
		}

		// get defaults
		if alias, ok := config.Harvest.TaskAliases[logTask.str]; ok {

			if logDuration.hours == nil {
				logDuration.hours = alias.DefaultDuration
			}
			if logNotes.str == "" && alias.DefaultNotes != nil {
				logNotes.str = *alias.DefaultNotes
			}
		}

		// prompt for duration if still not there
		if logDuration.hours == nil {
			if err = logDuration.prompt("Duration"); err != nil {
				return err
			}
		}

		// get message
		if logNotes.str == "" {
			if err = logNotes.prompt("Notes"); err != nil {
				return err
			}
		}

		// log time
		entry, err := harvest.LogTime(harvest.LogTimeOptions{
			TaskId:    *logTask.taskId,
			ProjectId: *logTask.projectId,
			Date:      *logDate.date,
			Hours:     *logDuration.hours,
			Notes:     logNotes.str,
		}, ctx)

		if err != nil {
			return err
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return outputSuccess() },
			config.OutputFormatTable:  func() error { return outputEntryTable(entry) },
			config.OutputFormatJson:   func() error { return outputJson(entry) },
		})
	}),
}

func outputSuccess() error {
	fmt.Println("Successful")
	return nil
}

func outputEntryTable(entry harvest.Entry) error {
	table := createTable([]string{"Key", "Value"})
	table.AppendBulk([][]string{
		{"ID", fmt.Sprintf("%v", entry.ID)},
		{"Project", entry.Project.Name},
		{"Task", entry.Task.Name},
		{"Notes", entry.Notes},
		{"Date", entry.Date},
		{"Hours", fmtHours(&entry.Hours)},
	})
	table.Render()
	return nil
}

func init() {
	_ = logDate.Set("today")
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().VarP(&logProject, "project", "p", "Set project (see root's ALIASES section)")
	logCmd.Flags().VarP(&logTask, "task", "t", "Set the task (see root's ALIASES section)")
	logCmd.Flags().VarP(&logNotes, "notes", "n", "Add notes to the time entry")
	logCmd.Flags().VarP(&logDate, "date", "d", "Set the date for the entry")
	logCmd.Flags().VarP(&logDuration, "duration", "D", "Set the duration for the entry")
}
