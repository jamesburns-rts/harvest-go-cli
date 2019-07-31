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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksAliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "List task aliases",
	Long:  `List task aliases`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		format := getOutputFormat()
		aliases := config.Harvest.TaskAliases
		if format == config.OutputFormatSimple {
			for k := range aliases {
				fmt.Println(k)
			}

		} else if format == config.OutputFormatJson {
			return outputJson(aliases)

		} else if format == config.OutputFormatTable {
			table := createTable([]string{"Alias", "ProjectId", "TaskId"})
			for k, v := range aliases {
				table.Append([]string{
					k,
					strconv.Itoa(int(v.ProjectId)),
					strconv.Itoa(int(v.TaskId)),
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
	tasksCmd.AddCommand(tasksAliasesCmd)
}
