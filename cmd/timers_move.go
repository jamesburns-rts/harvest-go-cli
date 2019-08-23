package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"time"
)

var timersMoveHours hoursArg

var timersMoveCmd = &cobra.Command{
	Use:   "move ORIGIn DESTINATION",
	Args:  cobra.ExactArgs(2),
	Short: "Move a timer",
	Long:  `Move a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		originName, destinationName := args[0], args[1]
		if origin, ok := timers.Get(originName); ok {

			if timersMoveHours.hours != nil {

				var destination timers.Timer
				if destination, ok = timers.Get(destinationName); !ok {
					destination.Name = destinationName
					destination.SetStarted(time.Now())
				}
				destination.Duration += *timersMoveHours.hours
				origin.Duration -= *timersMoveHours.hours
				timers.Set(origin)
				timers.Set(destination)

				fmt.Printf("Moved %s of %s to %s\n", timersMoveHours.str, originName, destinationName)
			} else {
				timers.Delete(originName)
				origin.Name = destinationName
				timers.Set(origin)
				fmt.Printf("Moved %s to %s\n", originName, destinationName)
			}
		} else {
			return errors.New("timer does not exist")
		}
		return writeConfig()
	}),
}

func init() {
	timersCmd.AddCommand(timersMoveCmd)
	timersMoveCmd.Flags().VarP(&timersMoveHours, "hours", "H", "Amount of duration to move (default all)")
}
