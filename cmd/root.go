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
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var cfgFile string
var outputFormat string

type rootSummary struct {
	harvest.MonthSummary
	WorkedTodayHours *Hours
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "harvest",
	Short: "A commandline tool for all things Harvest Time Tracking",
	Long:  `TODO - longer description`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var harvestSummary harvest.MonthSummary
		var workedTodayHours *Hours

		// calculate monthly summary
		if harvestSummary, err = harvest.CalculateMonthSummary(time.Now(), ctx); err != nil {
			return errors.Wrap(err, "calculating summary")
		}

		arrived := timers.Records.ArrivedTime()
		if arrived != nil && util.SameDay(*arrived, time.Now()) {
			calc := Hours(time.Now().Sub(*arrived).Hours())
			workedTodayHours = &calc
		}

		summary := rootSummary{
			MonthSummary:     harvestSummary,
			WorkedTodayHours: workedTodayHours,
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return rootOutputSimple(summary) },
			config.OutputFormatTable:  func() error { return rootOutputTable(summary) },
			config.OutputFormatJson:   func() error { return outputJson(summary) },
		})
	}),
}

func rootOutputSimple(s rootSummary) error {
	fmt.Printf(`
    Month Required Hours: %v
    Month Logged Hours: %v

    Month Billable Hours: %v (%0.1f%%)
    Month NonBillable Hours: %v

    Time worked: %v
    Logged today: %v
`,
		s.RequiredHours,
		s.MonthLoggedHours,
		s.BillableHours,
		100*s.BillableHours/s.MonthLoggedHours,
		s.NonBillableHours,
		s.WorkedTodayHours,
		s.TodayLoggedHours,
	)
	return nil
}
func rootOutputTable(s rootSummary) error {
	table := createTable(nil)
	table.AppendBulk([][]string{
		{"Month Required Hours", fmtHours(&s.RequiredHours)},
		{"Month Logged Hours", fmtHours(&s.MonthLoggedHours)},
		{"Month Billable Hours", fmt.Sprintf("%v (%0.1f%%)", s.BillableHours, 100*s.BillableHours/s.MonthLoggedHours)},
		{"Month NonBillable Hours", fmtHours(&s.NonBillableHours)},
		{"Time worked", fmtHours(s.WorkedTodayHours)},
		{"Logged today", fmtHours(&s.TodayLoggedHours)},
	})
	table.Render()
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	rootCmd.Use = filepath.Base(os.Args[0])
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.harvest.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "", fmt.Sprintf(
		"Format of output [%s]", strings.Join(config.OutputFormatOptions, ", ")))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".harvest" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".harvest")
		cfgFile = path.Join(home, ".harvest.yml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	_ = viper.ReadInConfig()

	conf := &viperConfig{}
	if err := viper.Unmarshal(conf); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// initialize maps and things
	if conf.Harvest.ProjectAliases == nil {
		conf.Harvest.ProjectAliases = make(map[string]config.ProjectAlias)
	}
	if conf.Harvest.TaskAliases == nil {
		conf.Harvest.TaskAliases = make(map[string]config.TaskAlias)
	}
	if timers.Records.Timers == nil {
		timers.Records.Timers = make(map[string]timers.Timer)
	}

	config.Harvest = conf.Harvest
	config.Cli = conf.Cli
	timers.Records = conf.Timers

	// clear old timers
	for k, v := range timers.Records.Timers {
		if !util.SameDay(*v.StartedTime(), time.Now()) {
			delete(timers.Records.Timers, k)
		}
	}
}

type viperConfig struct {
	Harvest config.HarvestProperties
	Cli     config.CliProperties
	Timers  timers.TrackingRecords
}

func getOutputFormat() string {
	if outputFormat != "" {
		return strings.ToLower(outputFormat)
	}
	if config.Cli.DefaultOutputFormat != "" {
		return config.Cli.DefaultOutputFormat
	}
	return config.OutputFormatTable
}

type outputMap map[string]func() error

func printWithFormat(supportedFormats map[string]func() error) error {
	format := getOutputFormat()
	if f, ok := supportedFormats[format]; ok {
		return f()
	} else {
		return errors.New("unsupported --format " + format)
	}
}
