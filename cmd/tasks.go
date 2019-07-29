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
	"github.com/jamesburns-rts/harvest-go-cli/internal/time"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

var tasksProjectId string
var tasksFormat string

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks of a project",
	Long:  `For the given project, list the tasks with their associated IDs`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		tasksFormat = strings.ToLower(tasksFormat)

		projectId, err := time.GetProjectId(tasksProjectId)
		if err != nil {
			return errors.Wrap(err, "for --project")
		}

		tasks, err := time.GetTasks(projectId, ctx)
		if err != nil {
			return err
		}

		if tasksFormat == formatSimple {
			for _, task := range tasks {
				fmt.Printf("%v %v\n", task.ID, task.Name)
			}

		} else if tasksFormat == formatJson {
			b, err := json.MarshalIndent(tasks, "", "  ")
			if err != nil {
				return errors.Wrap(err, "problem marshalling projects to json")
			}
			fmt.Println(string(b))

		} else if tasksFormat == formatTable {

			table := createTable([]string{"ID", "Task Name"})
			for _, task := range tasks {
				table.Append([]string{
					strconv.Itoa(int(task.ID)),
					task.Name,
				})
			}
			table.Render()

		} else {
			return errors.New("unrecognized --format " + tasksFormat)
		}
		return nil
	}),
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.Flags().StringVarP(&tasksProjectId, "project", "p", "", "ProjectID")
	tasksCmd.Flags().StringVarP(&tasksFormat, "format", "f", formatTable, "Format of output "+formatOptions)
}
