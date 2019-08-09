package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
	"time"
)

var arrivedShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the time arrived at work",
	Long:  `Show the time arrived at work`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		timeArrived := timers.Records.ArrivedTime()

		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return arrivedShowSimple(timeArrived) },
			config.OutputFormatTable:  func() error { return arrivedShowSimple(timeArrived) },
			config.OutputFormatJson:   func() error { return outputJson(timeArrived) },
		})
	}),
}

func arrivedShowSimple(t *time.Time) error {
	if t != nil {
		fmt.Println(formatArrived(*t))
	} else {
		fmt.Println("No arrived time set")
	}
	return nil
}

func init() {
	arrivedCmd.AddCommand(arrivedShowCmd)
}
