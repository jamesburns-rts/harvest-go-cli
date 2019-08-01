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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksAliasProjectId string

var tasksAliasCmd = &cobra.Command{
	Use:   "alias [TaskId] [Alias]",
	Args:  cobra.ExactArgs(2),
	Short: "Alias a task ID",
	Long:  `Alias a task ID to a friendly string the can be used anywhere`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var taskId int64
		var alias string
		var projectId *int64

		// gather inputs
		if taskId, err = strconv.ParseInt(args[0], 10, 64); err != nil {
			return errors.Wrap(err, "for [taskID]")
		}
		alias = args[1]

		if tasksAliasProjectId != "" {
			if projectId, err = harvest.GetProjectId(tasksAliasProjectId); err != nil {
				return errors.Wrap(err, "getting project")
			}
		} else {
			if projectId, err = getTaskProjectId(taskId, ctx); err != nil {
				return errors.Wrap(err, "error getting task project")
			}
		}

		// set alias
		config.Harvest.TaskAliases[alias] = config.TaskAlias{
			TaskId:    taskId,
			ProjectId: *projectId,
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
}
