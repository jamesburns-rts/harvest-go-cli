package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var timersRenameCmd = &cobra.Command{
	Use:   "rename NAME NEW_NAME",
	Args:  cobra.ExactArgs(2),
	Short: "Rename a timer",
	Long:  `Rename a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		origName, newName := args[0], args[1]
		if timer, ok := timers.Records.Timers[origName]; ok {
			delete(timers.Records.Timers, origName)
			timer.Name = newName
			timers.SetTimer(timer)
			fmt.Printf("Moved %s to %s\n", origName, newName)
		} else {
			return errors.New("timer does not exist")
		}
		return writeConfig()
	}),
}

func init() {
	timersCmd.AddCommand(timersRenameCmd)
}
