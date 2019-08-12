package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"time"
)

var arrivedClear bool

var arrivedCmd = &cobra.Command{
	Use:   "arrived [TIME]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Mark time arrived at work",
	Long:  `Mark time arrived at work. [TIME] should be of format hh:mm but defaults to the current time.`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		if arrivedClear {
			timers.Records.Arrived = ""
			fmt.Println("cleared")
		} else {

			timeArrived := time.Now()

			// gather inputs
			if len(args) > 0 {
				if timeArrived, err = util.StringToTime(args[0]); err != nil {
					return errors.New("for [TIME]")
				}
			}

			// set time arrived
			timers.Records.SetArrived(timeArrived)

			_ = printWithFormat(outputMap{
				config.OutputFormatSimple: func() error { return arrivedShowSimple(&timeArrived) },
				config.OutputFormatTable:  func() error { return arrivedShowSimple(&timeArrived) },
				config.OutputFormatJson:   func() error { return outputJson(timeArrived) },
			})
		}
		return writeConfig()

	}),
}

func init() {
	rootCmd.AddCommand(arrivedCmd)
	arrivedCmd.Flags().BoolVarP(&arrivedClear, "clear", "c", false, "Clear the current arrived time")
}

func formatArrived(t time.Time) string {
	if util.SameDay(t, time.Now()) {
		return t.Format(time.Kitchen)
	} else {
		return t.Format("Mon Jan _2 3:04PM")
	}
}
