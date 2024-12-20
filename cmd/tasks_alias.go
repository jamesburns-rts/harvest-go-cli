package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksAliasTaskId int64
var tasksAliasProject projectArg
var tasksAliasNotes stringArg
var tasksAliasDuration hoursArg

var tasksAliasCmd = &cobra.Command{
	Use:   "alias [ALIAS [TASK_ID]]",
	Args:  cobra.MaximumNArgs(2),
	Short: "Alias a task ID",
	Long:  `Alias a task ID to a friendly string the can be used anywhere`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		// get alias
		var alias string
		if len(args) > 0 {
			alias = args[0]
		} else {
			if alias, err = prompt.ForString("Alias Name", validAlias); err != nil {
				return err
			}
		}

		var projectId *int64
		var taskId *int64
		var defaultNotes *string
		var defaultDuration *Hours

		// check for existing
		if taskAlias, ok := config.GetTaskAlias(alias); ok {
			projectId = &taskAlias.ProjectId
			taskId = &taskAlias.TaskId
			defaultNotes = taskAlias.DefaultNotes
			defaultDuration = taskAlias.DefaultDuration
		}

		// get projectId maybe
		if tasksAliasProject.str != "" {
			projectId = tasksAliasProject.projectId
		}

		// get task ID
		if len(args) > 1 {
			if id, err := strconv.ParseInt(args[1], 10, 64); err != nil {
				return fmt.Errorf("for [TASK_ID]: %w", err)
			} else {
				taskId = &id
			}
		}
		if tasksAliasTaskId != -1 {
			taskId = &tasksAliasTaskId
		}
		if tasksAliasDuration.str != "" {
			defaultDuration = tasksAliasDuration.hours
		}

		if taskId == nil {
			if projectId == nil {
				if projectId, taskId, err = selectProjectAndTask(ctx); err != nil {
					return err
				}
			} else {
				if taskId, err = selectTask(*projectId, ctx); err != nil {
					return err
				}
			}

		} else if projectId == nil {
			// select project
			if projectId, err = getTaskProjectId(*taskId, ctx); err != nil {
				return fmt.Errorf("error getting task project: %w", err)
			}
		}

		if tasksAliasNotes.str != "" {
			defaultNotes = &tasksAliasNotes.str
		}

		// set alias
		config.SetTaskAlias(config.TaskAlias{
			Name:            alias,
			TaskId:          *taskId,
			ProjectId:       *projectId,
			DefaultNotes:    defaultNotes,
			DefaultDuration: defaultDuration,
		})

		return writeConfig()
	}),
}

var timeTasksAliasDeleteCmd = &cobra.Command{
	Use:   "delete [ALIAS]",
	Args:  cobra.ExactArgs(1),
	Short: "Delete a task ID alias",
	Long:  ``,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var alias string
		if len(args) > 0 {
			alias = args[0]
		} else {
			if alias, err = selectTaskAlias(); err != nil {
				return err
			}
		}
		config.DeleteTaskAlias(alias)
		return writeConfig()
	}),
}

func init() {
	tasksCmd.AddCommand(tasksAliasCmd)
	tasksAliasCmd.AddCommand(timeTasksAliasDeleteCmd)

	tasksAliasCmd.Flags().VarP(&tasksAliasProject, "project", "p", "project ID/alias the task is for")
	tasksAliasCmd.Flags().Int64VarP(&tasksAliasTaskId, "task", "t", -1, "Task ID the task is for")
	tasksAliasCmd.Flags().VarP(&tasksAliasNotes, "default-notes", "n", "Default notes to use when logging time")
	tasksAliasCmd.Flags().VarP(&tasksAliasDuration, "default-hours", "H", "Default duration to use when logging time")

	_ = tasksAliasCmd.RegisterFlagCompletionFunc("project", projectCompletionFunc)
}
