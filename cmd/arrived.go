package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var arrivedCmd = &cobra.Command{
	Use:   "arrived [time]",
	Short: "Mark time arrived at work",
	Long:  `Mark time arrived at work`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {
		t, err := parseTime(args)
		if err != nil {
			return errors.New("expected time format of hh:mm")
		}

		config.Timers.SetArrived(t)
		fmt.Printf("Marking time arrived as %s\n", config.Timers.Arrived)
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
