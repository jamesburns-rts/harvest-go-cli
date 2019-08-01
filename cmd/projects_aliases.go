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
	"github.com/spf13/cobra"
	"strconv"
)

var projectsAliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "List project aliases",
	Long:  `List project aliases`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		aliases := config.Harvest.ProjectAliases

		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return projectsAliasesOutputSimple(aliases) },
			config.OutputFormatTable:  func() error { return projectsAliasesOutputTable(aliases) },
			config.OutputFormatJson:   func() error { return outputJson(aliases) },
		})
	}),
}

func projectsAliasesOutputSimple(aliases map[string]config.ProjectAlias) error {
	for k := range aliases {
		fmt.Println(k)
	}

	return nil
}
func projectsAliasesOutputTable(aliases map[string]config.ProjectAlias) error {
	table := createTable([]string{"Alias", "ProjectId"})
	for k, v := range aliases {
		table.Append([]string{k, strconv.Itoa(int(v.ProjectId))})
	}
	table.Render()
	return nil
}

func init() {
	projectsCmd.AddCommand(projectsAliasesCmd)
}
