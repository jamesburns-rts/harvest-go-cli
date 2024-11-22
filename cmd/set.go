package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/spf13/cobra"
)

// settable values
var setHarvestArgs config.HarvestProperties
var setCliArgs config.CliProperties

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set config of time",
	Long:  `TODO - longer description`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		// gather inputs
		if setHarvestArgs.AccountId != "" {
			config.Harvest.AccountId = setHarvestArgs.AccountId
		}
		if setHarvestArgs.AccessToken != "" {
			config.Harvest.AccessToken = setHarvestArgs.AccessToken
		}

		if setCliArgs.DefaultOutputFormat != "" {
			if option, ok := config.OutputFormatOptions.Contains(setCliArgs.DefaultOutputFormat); ok {
				config.Cli.DefaultOutputFormat = option
			} else {
				return errors.New("Invalid output format given")
			}
		}

		if setCliArgs.TimeDeltaFormat != "" {
			if option, ok := config.TimeDeltaFormatOptions.Contains(setCliArgs.TimeDeltaFormat); ok {
				config.Cli.TimeDeltaFormat = option
			} else {
				return errors.New("Invalid time format given")
			}
		}

		// write config
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
	setCmd.Flags().StringVar(&setCliArgs.DefaultOutputFormat, "default-output-format", "", fmt.Sprintf(
		"Default output format %v", config.OutputFormatOptions))

	setCmd.Flags().StringVar(&setCliArgs.TimeDeltaFormat, "time-format", "", fmt.Sprintf(
		"Default time delta format %s", config.TimeDeltaFormatOptions))

	setCmd.AddCommand(timeSetProjectAliasCmd)
	timeSetProjectAliasCmd.AddCommand(projectsAliasCmd)

	setCmd.AddCommand(timeSetTaskAliasCmd)
	timeSetTaskAliasCmd.AddCommand(tasksAliasCmd)
}
