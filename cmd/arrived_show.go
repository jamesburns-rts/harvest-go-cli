package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var arrivedShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the time arrived at work",
	Long:  `Show the time arrived at work`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		timeArrived := timers.Records.ArrivedTime()

		if timeArrived != nil {
			fmt.Println(formatArrived(*timeArrived))
		} else {
			fmt.Println("No arrived time set")
		}
		return nil
	}),
}

func init() {
	arrivedCmd.AddCommand(arrivedShowCmd)
}
