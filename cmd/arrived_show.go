package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/spf13/cobra"
)

var arrivedShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the time arrived at work",
	Long:  `Show the time arrived at work`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		t := config.Tracking.ArrivedTime()
		if t != nil {
			fmt.Println(formatArrived(*t))
		} else {
			fmt.Println("No arrived time set")
		}
		return nil
	}),
}

func init() {
	arrivedCmd.AddCommand(arrivedShowCmd)
}
