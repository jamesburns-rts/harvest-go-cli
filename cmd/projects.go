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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var projectsFormat string

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects",
	Long:  `List projects and their associated IDs`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		projectsFormat = strings.ToLower(projectsFormat)

		// get projects
		projects, err := time.GetProjects(ctx)
		if err != nil {
			return err
		}

		// print
		if projectsFormat == formatSimple {
			for _, proj := range projects {
				fmt.Printf("%v %v\n", proj.ID, proj.Name)
			}

		} else if projectsFormat == formatJson {
			b, err := json.MarshalIndent(projects, "", "  ")
			if err != nil {
				return errors.Wrap(err, "problem marshalling projects to json")
			}
			fmt.Println(string(b))

		} else if projectsFormat == formatTable {

			table := createTable([]string{"ID", "Project Name"})
			for _, proj := range projects {
				table.Append([]string{
					strconv.Itoa(int(proj.ID)),
					proj.Name,
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
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.Flags().StringVarP(&projectsFormat, "format", "f", formatTable, "Format of output "+formatOptions)
}
