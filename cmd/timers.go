package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"time"
)

var timersDoNotSync bool
var timersJustNames bool

var timersCmd = &cobra.Command{
	Use:   "timers",
	Short: "List timers",
	Long:  `List timers`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return timersSimple() },
			config.OutputFormatTable:  func() error { return timersTable() },
			config.OutputFormatJson:   func() error { return timersJson() },
		})
	}),
}

func timersSimple() error {
	if timersJustNames {
		for t := range timers.Records.Timers {
			fmt.Println(t)
		}
	} else {
		for _, t := range timers.Records.Timers {

			duration := fmtHours(t.RunningHours())
			if t.Running {
				duration += " - running"
			}

			fmt.Printf("%s: %s\n", t.Name, duration)
		}
	}
	return nil
}

func timersTable() error {
	if timersJustNames {
		table := createTable([]string{"Name"})
		for t := range timers.Records.Timers {
			table.Append([]string{t})
		}
		table.Render()
	} else {
		table := createTable([]string{"Name", "Duration", "Started"})
		for _, t := range timers.Records.Timers {

			started := ""
			if t.Running {
				if t.Started == "" {
					return errors.New("No start time for running timer")
				}
				started = t.StartedTime().Format(time.Kitchen)
			}
			table.Append([]string{t.Name, fmtHours(t.RunningHours()), started})
		}
		table.Render()
	}
	return nil
}

func timersJson() error {
	if timersJustNames {
		var names []string
		for t := range timers.Records.Timers {
			names = append(names, t)
		}
		return outputJson(names)
	} else {
		return outputJson(timers.Records.Timers)
	}
}

func init() {
	rootCmd.AddCommand(timersCmd)
	timersCmd.PersistentFlags().BoolVar(&timersDoNotSync, "do-not-sync", false, "Prevent syncing with harvest timers")
	timersCmd.PersistentFlags().BoolVarP(&timersJustNames, "just-names", "l", false, "Just print the timer names")
}
