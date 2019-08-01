package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"time"
)

var logMessage string
var logDate string

var logCmd = &cobra.Command{
	Use:   "log [task] [duration]",
	Args:  cobra.ExactArgs(2),
	Short: "Log a time entry",
	Long:  `Log a time entry`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		taskId, err := harvest.GetTaskId(args[0])
		if err != nil {
			return err
		}

		date, err := util.StringToDate(logDate)
		if err != nil {
			return err
		}

		duration, err := time.ParseDuration(args[1])
		if err != nil {
			return errors.Wrap(err, "problem parsing duration")
		}

		entry, err := harvest.LogTime(harvest.LogTimeOptions{
			TaskId: *taskId,
			Date:   *date,
			Hours:  Hours(duration.Hours()),
			Notes:  logMessage,
		}, ctx)

		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == config.OutputFormatSimple {
			fmt.Println("Successful")

		} else if format == config.OutputFormatJson {
			return outputJson(entry)

		} else if format == config.OutputFormatTable {
			table := createTable([]string{"Key", "Value"})
			table.AppendBulk([][]string{
				{"ID", fmt.Sprintf("%v", entry.ID),},
				{"Project", entry.Project.Name,},
				{"Task", entry.Task.Name,},
				{"Message", entry.Notes,},
				{"Date", entry.Date,},
				{"Hours", entry.Hours.String(),},
			})
			table.Render()
		} else {
			return errors.New("unrecognized --format " + format)
		}

		return nil
	}),
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVarP(&logMessage, "message", "m", "", "Add notes to the time entry")
	logCmd.Flags().StringVarP(&logDate, "date", "d", "today", "Set the date for the entry")
}
