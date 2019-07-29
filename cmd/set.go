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
	"github.com/spf13/cobra"
)

// settable values
var configArgs config.HarvestProperties

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set config of time",
	Long:  `TODO - longer description`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		if configArgs.AccountId != "" {
			config.Harvest.AccountId = configArgs.AccountId
		}
		if configArgs.AccessToken != "" {
			config.Harvest.AccessToken = configArgs.AccessToken
		}

		return writeConfig()
	}),
}

var timeSetProjectAliasCmd = &cobra.Command{
	Use:   "project",
	Short: "Set project stuff",
	Long:  ``,
}

var timeSetTaskAliasCmd = &cobra.Command{
	Use:   "task",
	Short: "Set task stuff",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.Flags().StringVar(&configArgs.AccessToken, "harvest-access-token", "", "Harvest API Access token")
	setCmd.Flags().StringVar(&configArgs.AccountId, "harvest-account-id", "", "Harvest API account ID")

	setCmd.AddCommand(timeSetProjectAliasCmd)
	timeSetProjectAliasCmd.AddCommand(projectsAliasCmd)

	setCmd.AddCommand(timeSetTaskAliasCmd)
	timeSetTaskAliasCmd.AddCommand(tasksAliasCmd)
}
