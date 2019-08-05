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
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksAliasTaskId string
var tasksAliasProjectId string
var tasksAliasNotes string
var tasksAliasDuration string

var tasksAliasCmd = &cobra.Command{
	Use:   "alias [Alias] [TaskId]",
	Args:  cobra.MaximumNArgs(2),
	Short: "Alias a task ID",
	Long:  `Alias a task ID to a friendly string the can be used anywhere`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		// get alias
		var alias string
		if len(args) > 0 {
			alias = args[0]
		} else {
			if alias, err = prompt.ForWord("Alias Name"); err != nil {
				return err
			}
		}

		var projectId *int64
		var taskId *int64
		var defaultNotes *string
		var defaultDuration *Hours

		// check for existing
		if taskAlias, ok := config.Harvest.TaskAliases[alias]; ok {
			projectId = &taskAlias.ProjectId
			taskId = &taskAlias.TaskId
			defaultNotes = taskAlias.DefaultNotes
			defaultDuration = taskAlias.DefaultDuration
		}

		// get projectId maybe
		if tasksAliasProjectId != "" {
			if projectId, err = harvest.ParseProjectId(tasksAliasProjectId); err != nil {
				return errors.Wrap(err, "getting project")
			}
		}

		// get task ID
		if len(args) > 1 {
			if id, err := strconv.ParseInt(args[1], 10, 64); err != nil {
				return errors.Wrap(err, "for [taskID]")
			} else {
				taskId = &id
			}
		}
		if tasksAliasTaskId != "" {
			if id, err := strconv.ParseInt(tasksAliasTaskId, 10, 64); err != nil {
				return errors.Wrap(err, "for --task")
			} else {
				taskId = &id
			}
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
				return errors.Wrap(err, "error getting task project")
			}
		}

		if tasksAliasNotes != "" {
			defaultNotes = &tasksAliasNotes
		}
		if tasksAliasDuration != "" {
			var duration Hours
			if duration, err = ParseHours(tasksAliasDuration); err != nil {
				return errors.Wrap(err, "for --default-duration")
			}
			defaultDuration = &duration
		}

		// set alias
		config.Harvest.TaskAliases[alias] = config.TaskAlias{
			TaskId:          *taskId,
			ProjectId:       *projectId,
			DefaultNotes:    defaultNotes,
			DefaultDuration: defaultDuration,
		}

		return writeConfig()
	}),
}

var timeTasksAliasDeleteCmd = &cobra.Command{
	Use:   "delete [Alias]",
	Args:  cobra.ExactArgs(1),
	Short: "Delete a task ID alias",
	Long:  ``,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		alias := args[0]
		delete(config.Harvest.TaskAliases, alias)

		return writeConfig()
	}),
}

func init() {
	tasksCmd.AddCommand(tasksAliasCmd)
	tasksAliasCmd.AddCommand(timeTasksAliasDeleteCmd)

	tasksAliasCmd.Flags().StringVarP(&tasksAliasProjectId, "project", "p", "", "project ID/alias the task is for")
	tasksAliasCmd.Flags().StringVarP(&tasksAliasTaskId, "task", "t", "", "Task ID the task is for")
	tasksAliasCmd.Flags().StringVarP(&tasksAliasNotes, "default-notes", "m", "", "Default notes to use when logging time")
	tasksAliasCmd.Flags().StringVarP(&tasksAliasDuration, "default-duration", "d", "", "Default duration to use when logging time")
}
