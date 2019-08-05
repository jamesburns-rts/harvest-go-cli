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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksProjectId string

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks of a project",
	Long:  `For the given project, list the tasks with their associated IDs`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var projectId *int64

		// gather inputs
		if projectId, err = harvest.ParseProjectId(tasksProjectId); err != nil {
			return errors.Wrap(err, "for --project")
		}

		// get tasks
		var tasks []harvest.Task
		if tasks, err = harvest.GetTasks(projectId, ctx); err != nil {
			return errors.Wrap(err, "getting tasks")
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return tasksOutputSimple(tasks) },
			config.OutputFormatTable:  func() error { return tasksOutputTable(tasks) },
			config.OutputFormatJson:   func() error { return outputJson(tasks) },
		})
	}),
}

func tasksOutputSimple(tasks []harvest.Task) error {
	for _, task := range tasks {
		fmt.Printf("%v %v %v\n", task.ProjectId, task.ID, task.Name)
	}
	return nil
}

func tasksOutputTable(tasks []harvest.Task) error {
	table := createTable([]string{"Project ID", "ID", "Task Name"})
	for _, task := range tasks {
		table.Append([]string{
			strconv.Itoa(int(task.ProjectId)),
			strconv.Itoa(int(task.ID)),
			task.Name,
		})
	}
	table.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.Flags().StringVarP(&tasksProjectId, "project", "p", "", "ProjectID")
}
