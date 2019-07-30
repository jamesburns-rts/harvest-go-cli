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
	"strings"
)

// settable values
var setHarvestArgs config.HarvestProperties
var setCliArgs config.CliProperties

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set config of time",
	Long:  `TODO - longer description`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		if setHarvestArgs.AccountId != "" {
			config.Harvest.AccountId = setHarvestArgs.AccountId
		}
		if setHarvestArgs.AccessToken != "" {
			config.Harvest.AccessToken = setHarvestArgs.AccessToken
		}
		if setCliArgs.DefaultOutputFormat != "" {
			setCliArgs.DefaultOutputFormat = strings.ToLower(setCliArgs.DefaultOutputFormat)
			switch setCliArgs.DefaultOutputFormat {
			case config.OutputFormatTable:
				fallthrough
			case config.OutputFormatJson:
				fallthrough
			case config.OutputFormatSimple:
				config.Cli.DefaultOutputFormat = setCliArgs.DefaultOutputFormat
			default:
				return errors.New("Invalid output format given")
			}
		}
		if setCliArgs.TimeDeltaFormat != "" {
			setCliArgs.TimeDeltaFormat = strings.ToLower(setCliArgs.TimeDeltaFormat)
			switch setCliArgs.TimeDeltaFormat {
			case config.TimeDeltaFormatHuman:
				fallthrough
			case config.TimeDeltaFormatDecimal:
				config.Cli.TimeDeltaFormat = setCliArgs.TimeDeltaFormat
			default:
				return errors.New("Invalid time format given " + setCliArgs.TimeDeltaFormat)
			}
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
	setCmd.Flags().StringVar(&setHarvestArgs.AccessToken, "harvest-access-token", "", "Harvest API Access token")
	setCmd.Flags().StringVar(&setHarvestArgs.AccountId, "harvest-account-id", "", "Harvest API account ID")
	setCmd.Flags().StringVar(&setCliArgs.DefaultOutputFormat, "default-output-format", "",
		"Default output format "+outputFormatOptions)
	setCmd.Flags().StringVar(&setCliArgs.TimeDeltaFormat, "time-format", "",
		fmt.Sprintf("Default output format [%s, %s]", config.TimeDeltaFormatDecimal, config.TimeDeltaFormatHuman))

	setCmd.AddCommand(timeSetProjectAliasCmd)
	timeSetProjectAliasCmd.AddCommand(projectsAliasCmd)

	setCmd.AddCommand(timeSetTaskAliasCmd)
	timeSetTaskAliasCmd.AddCommand(tasksAliasCmd)
}
