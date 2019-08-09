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
	"github.com/spf13/cobra"
	"strconv"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects",
	Long:  `List projects and their associated IDs`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		// get projects
		projects, err := harvest.GetProjects(ctx)
		if err != nil {
			return err
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return projectsOutputSimple(projects) },
			config.OutputFormatTable:  func() error { return projectsOutputTable(projects) },
			config.OutputFormatJson:   func() error { return outputJson(projects) },
		})
	}),
}

func projectsOutputSimple(projects []harvest.Project) error {
	for _, p := range projects {
		fmt.Printf("%v %v\n", p.ID, p.Name)
	}
	return nil
}

func projectsOutputTable(projects []harvest.Project) error {
	table := createTable([]string{"ID", "Project Name"})
	for _, proj := range projects {
		table.Append([]string{
			strconv.Itoa(int(proj.ID)),
			proj.Name,
		})
	}
	table.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}
