package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var arrivedCmd = &cobra.Command{
	Use:   "arrived [TIME]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Mark time arrived at work",
	Long:  `Mark time arrived at work`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var timeArrived time.Time

		// gather inputs
		if timeArrived, err = parseTime(args); err != nil {
			return errors.New("for [time]: expected time format of hh:mm")
		}

		// set time arrived
		timers.Records.SetArrived(timeArrived)

		// output
		fmt.Printf("Marking time arrived as %s\n", formatArrived(timeArrived))
		return writeConfig()
	}),
}

func init() {
	rootCmd.AddCommand(arrivedCmd)
}

func parseTime(args []string) (t time.Time, err error) {
	t = time.Now()

	if len(args) > 0 {
		str := strings.ToUpper(args[0])

		var tm time.Time
		if !strings.HasSuffix(str, "PM") && !strings.HasSuffix(str, "AM") {
			tm, err = time.Parse("15:04", str)
		} else {
			tm, err = time.Parse(time.Kitchen, str)
		}
		if err != nil {
			return t, err
		}
		return time.Date(t.Year(), t.Month(), t.Day(), tm.Hour(), tm.Minute(), 0, 0, t.Location()), nil
	}

	return t, nil
}

func formatArrived(t time.Time) string {
	if util.SameDay(t, time.Now()) {
		return t.Format(time.Kitchen)
	} else {
		return t.Format("Mon Jan _2 3:04PM")
	}
}
