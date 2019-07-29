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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksProjectId string

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks of a project",
	Long:  `For the given project, list the tasks with their associated IDs`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		projectId, err := harvest.GetProjectId(tasksProjectId)
		if err != nil {
			return errors.Wrap(err, "for --project")
		}

		tasks, err := harvest.GetTasks(projectId, ctx)
		if err != nil {
			return err
		}

		// print
		format := getOutputFormat()
		if format == config.OutputFormatSimple {
			for _, task := range tasks {
				fmt.Printf("%v %v\n", task.ID, task.Name)
			}

		} else if format == config.OutputFormatJson {
			b, err := json.MarshalIndent(tasks, "", "  ")
			if err != nil {
				return errors.Wrap(err, "problem marshalling projects to json")
			}
			fmt.Println(string(b))

		} else if format == config.OutputFormatTable {

			table := createTable([]string{"ID", "Task Name"})
			for _, task := range tasks {
				table.Append([]string{
					strconv.Itoa(int(task.ID)),
					task.Name,
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
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.Flags().StringVarP(&tasksProjectId, "project", "p", "", "ProjectID")
}
