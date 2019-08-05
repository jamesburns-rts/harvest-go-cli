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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var projectsAliasCmd = &cobra.Command{
	Use:   "alias [PROJECT_ID] [ALIAS]",
	Args:  cobra.ExactArgs(2),
	Short: "Alias a project ID",
	Long:  `Alias a project ID to a friendly string the can be used anywhere`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var projectId int64
		var alias string

		// gather inputs
		if projectId, err = strconv.ParseInt(args[0], 10, 64); err != nil {
			return errors.Wrap(err, "for [projectId]")
		}
		alias = args[1]

		// set alias
		config.Harvest.ProjectAliases[alias] = config.ProjectAlias{
			ProjectId: projectId,
		}

		return writeConfig()
	}),
}

var timeProjectsAliasDeleteCmd = &cobra.Command{
	Use:   "delete [Alias]",
	Args:  cobra.ExactArgs(1),
	Short: "Delete a project ID alias",
	Long:  ``,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		alias := args[0]
		delete(config.Harvest.ProjectAliases, alias)

		return writeConfig()
	}),
}

func init() {
	projectsCmd.AddCommand(projectsAliasCmd)
	projectsAliasCmd.AddCommand(timeProjectsAliasDeleteCmd)
}
