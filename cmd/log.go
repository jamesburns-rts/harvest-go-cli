package cmd

import (
	"context"
	"fmt"

	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var logTask taskArg
var logProject projectArg
var logNotes stringArg
var logDate dateArg
var logDuration hoursArg
var logConfirm bool
var logTimer string

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
				return fmt.Errorf("for [TASK]: %w", err)
			}
		}
		if len(args) > 1 {
			if err = logDuration.Set(args[1]); err != nil {
				return fmt.Errorf("for [DURATION]: %w", err)
			}
		}

		if logTimer != "" {
			if timer, ok := timers.Get(logTimer); !ok {
				return fmt.Errorf("timer %s does not exist", logTimer)
			} else {
				if logTask.taskId == nil {
					logTask.SetId(timer.SyncedTaskId, timer.SyncedProjectId)
				}
				if logDuration.hours == nil {
					logDuration.SetHours(timer.RunningHours())
				}
				if logNotes.str == "" {
					logNotes.str = timer.Notes
				}
				defer func() {
					if err == nil {
						timers.Delete(timer.Name)
						err = writeConfig()
					}
				}()
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
		} else if logTask.projectId != nil {
			if logTask.taskId, err = selectTask(*logTask.projectId, ctx); err != nil {
				return err
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
		if alias, ok := config.GetTaskAlias(logTask.str); ok {

			if logDuration.hours == nil {
				logDuration.hours = alias.DefaultDuration
			}
			if logNotes.str == "" && alias.DefaultNotes != nil {
				logNotes.str = *alias.DefaultNotes
			}
		}

		// prompt for missing arguments
		var confirms []prompt.Confirmation
		if logDuration.hours == nil || logConfirm {
			confirms = append(confirms, prompt.Confirmation{Title: "Duration", Value: &logDuration})
		}
		if logNotes.str == "" || logConfirm {
			confirms = append(confirms, prompt.Confirmation{Title: "Notes", Value: &logNotes})
		}
		if logConfirm {
			confirms = append(confirms, prompt.Confirmation{Title: "Date", Value: &logDate})
		}
		if err = prompt.ConfirmAll(confirms); err != nil {
			return err
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
			config.OutputFormatSimple: func() error { return outputSuccess(entry) },
			config.OutputFormatTable:  func() error { return outputEntryTable(entry) },
			config.OutputFormatJson:   func() error { return outputJson(entry) },
		})
	}),
}

func entryOutputRows(entry harvest.Entry) [][]string {
	return [][]string{
		{"ID", fmt.Sprintf("%v", entry.ID)},
		{"Project", entry.Project.Name},
		{"Task", entry.Task.Name},
		{"Notes", entry.Notes},
		{"Date", entry.Date},
		{"Duration", fmtHours(&entry.Hours)},
	}
}

func outputSuccess(entry harvest.Entry) error {
	fmt.Println()
	for _, r := range entryOutputRows(entry) {
		fmt.Printf("%s: %s\n", r[0], r[1])
	}
	return nil
}

func outputEntryTable(entry harvest.Entry) error {
	table := createTable([]string{"Key", "Value"})
	for _, row := range entryOutputRows(entry) {
		_ = table.Append(row)
	}
	_ = table.Render()
	return nil
}

func init() {
	_ = logDate.Set("today")
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().VarP(&logProject, "project", "p", "Set project (see root's ALIASES section)")
	logCmd.Flags().VarP(&logTask, "task", "t", "Set the task (see root's ALIASES section)")
	logCmd.Flags().VarP(&logNotes, "notes", "n", "Add notes to the time entry")
	logCmd.Flags().VarP(&logDate, "date", "d", "Set the date for the entry")
	logCmd.Flags().VarP(&logDuration, "hours", "H", "Set the duration for the entry")
	logCmd.Flags().BoolVarP(&logConfirm, "confirm", "c", false, "Confirm all the values before logging")
	logCmd.Flags().StringVarP(&logTimer, "timer", "T", "", "Get data from timer while creating record")

	_ = logCmd.RegisterFlagCompletionFunc("project", projectCompletionFunc)
	_ = logCmd.RegisterFlagCompletionFunc("task", taskCompletionFunc)
}
