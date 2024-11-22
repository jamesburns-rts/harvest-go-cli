package cmd

import (
	"context"
	"fmt"
	"time"

	"errors"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/spf13/cobra"
)

var timersMoveHours hoursArg

var timersMoveCmd = &cobra.Command{
	Use:     "move ORIGIN DESTINATION",
	Aliases: []string{"mv"},
	Args:    cobra.RangeArgs(2, 3),
	Short:   "Move a timer",
	Long:    `Move a timer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		originName, destinationName := args[0], args[1]
		if len(args) > 2 && timersMoveHours.str == "" {
			if err := timersMoveHours.Set(args[2]); err != nil {
				return err
			}
		}

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

			} else {
				var destination timers.Timer
				if destination, ok = timers.Get(destinationName); !ok {
					destination = origin
					destination.Name = destinationName
				} else {
					destination.Duration += *origin.RunningHours()
				}

				timers.Delete(originName)
				timers.Set(destination)
				timersMoveHours.str = fmtHours(origin.RunningHours())
			}

			fmt.Printf("Moved %s of %s to %s\n", timersMoveHours.str, originName, destinationName)
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
