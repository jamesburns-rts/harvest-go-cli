package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var outputFormat string

type rootSummary struct {
	harvest.MonthSummary
	WorkedTodayHours *Hours
	Timers           []timers.Timer
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "harvest",
	Short: "A commandline tool for all things Harvest Time Tracking",
	Long: `   __ _____   ___ _   ________________    _________      _______   ____
  / // / _ | / _ \ | / / __/ __/_  __/___/ ___/ __ \____/ ___/ /  /  _/
 / _  / __ |/ , _/ |/ / _/_\ \  / / /___/ (_ / /_/ /___/ /__/ /___/ /  
/_//_/_/ |_/_/|_||___/___/___/ /_/      \___/\____/    \___/____/___/  
                                                                    
A commandline tool for all things Harvest Time Tracking

ALIASES
Projects and project tasks can be aliased to easy to remember words. These words 
can then be used anywhere a project/task is needed. See 'harvest projects alias'
and 'harvest tasks alias'

HOURS
Inputs of a duration type can be a couple formats:
 - 1h15m: Human readable version with the number of hours and/or minutes
 - 1.25: Decimal number of hours

DATES
Inputs of date type can be a few formats:
 - yyyy-mm-dd: Standard ISO format, but hyphens are optional
 - '-N': where N is any integer "days ago"
 - 'mon[day]: Date of last Monday (only 'mon' is required)'
 - 'tue[sday], etc: Date of Tuesday or other day of the week'
 - 'yest[erday]: Date of the day before today
`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		userId, err := getAndSaveUserId(ctx)
		if err != nil {
			return err
		}

		var harvestSummary harvest.MonthSummary
		var workedTodayHours *Hours

		// calculate monthly summary
		harvestSummary, err = harvest.CalculateMonthSummary(time.Now(), userId, ctx)
		if err != nil && outputFormat != config.OutputFormatJson {
			fmt.Println(fmt.Errorf("calculating summary: %w", err))
		}

		arrived := timers.Records.ArrivedTime()
		if arrived != nil && util.SameDay(*arrived, time.Now()) {
			calc := Hours(time.Now().Sub(*arrived).Hours())
			workedTodayHours = &calc

			*workedTodayHours -= timers.SumTimeOn([]string{"lunch", "break"})
		}

		summary := rootSummary{
			MonthSummary:     harvestSummary,
			WorkedTodayHours: workedTodayHours,
			Timers:           timers.Records.Timers,
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return rootOutputSimple(summary) },
			config.OutputFormatTable:  func() error { return rootOutputTable(summary) },
			config.OutputFormatJson: func() error {
				for k, v := range summary.Timers {
					if v.Running {
						v.Duration = *v.RunningHours()
						summary.Timers[k] = v
					}
				}
				return outputJson(summary)
			},
		})
	}),
}

func rootOutputSimple(s rootSummary) error {

	var shortWeekMessage string
	if s.ShortWeek >= 0 {
		shortWeekMessage = fmt.Sprintf("You are %s short for today", fmtHours(&s.ShortWeek))
	} else {
		s.ShortWeek *= -1
		shortWeekMessage = fmt.Sprintf("You are %s over for today", fmtHours(&s.ShortWeek))
	}

	var shortMessage string
	if s.Short >= 0 {
		shortMessage = "-" + fmtHours(&s.Short)
	} else {
		s.Short *= -1
		shortMessage = fmtHours(&s.Short)
	}

	fmt.Printf(`
    Month Required Hours: %v
    Month Logged Hours: %v
    Month Delta: %v
    Week Logged Hours: %v

    Month Billable Hours: %v (%0.1f%%)
    Month NonBillable Hours: %v

    Time worked: %v
    Logged today: %v
`,
		fmtHours(&s.RequiredHours),
		fmtHours(&s.MonthLoggedHours),
		shortMessage,
		fmtHours(&s.WeekLoggedHours),
		fmtHours(&s.BillableHours),
		100*s.BillableHours/s.MonthLoggedHours,
		fmtHours(&s.NonBillableHours),
		fmtHours(s.WorkedTodayHours),
		fmtHours(&s.TodayLoggedHours),
	)

	if len(timers.Records.Timers) > 0 {
		fmt.Println()
		_ = timersSimple()
	}

	fmt.Println()
	fmt.Println(shortWeekMessage)

	return nil
}
func rootOutputTable(s rootSummary) error {
	table := createTable(nil)
	_ = table.Append([][]string{
		{"Month Required Hours", fmtHours(&s.RequiredHours)},
		{"Month Logged Hours", fmtHours(&s.MonthLoggedHours)},
		{"Month Billable Hours", fmt.Sprintf("%v (%0.1f%%)", s.BillableHours, 100*s.BillableHours/s.MonthLoggedHours)},
		{"Month NonBillable Hours", fmtHours(&s.NonBillableHours)},
		{"Time worked", fmtHours(s.WorkedTodayHours)},
		{"Logged today", fmtHours(&s.TodayLoggedHours)},
		{"Hours to go", fmtHours(&s.Short)},
	})
	_ = table.Render()

	if len(timers.Records.Timers) > 0 {
		fmt.Println("Timers:")
		_ = timersTable()
	}

	return nil
}

func CreateCommand() *cobra.Command {
	return rootCmd
}

func init() {
	//cobra.OnInitialize(initConfig)
	initConfig()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file (default is $HOME/.harvest.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "", fmt.Sprintf(
		"Format of output [%s]", strings.Join(config.OutputFormatOptions, ", ")))

	_ = rootCmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.OutputFormatOptions, cobra.ShellCompDirectiveNoFileComp
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
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

	config.Harvest = conf.Harvest
	config.Cli = conf.Cli
	timers.Records = conf.Timers

	// clear old timers
	oldTimers := slices.Clone(timers.Records.Timers)
	for _, v := range oldTimers {
		if !util.SameDay(v.StartedTime(), time.Now()) {
			timers.Delete(v.Name)
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
