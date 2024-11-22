package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/spf13/cobra"
	"math"
	"slices"
	"time"
)

var weeklyTask taskArg
var weeklyToDate dateArg
var weeklyFromDate dateArg
var weeklyBillableOnly bool
var weeklyExpected hoursArg

type weeklyEntry struct {
	Week  time.Time
	Hours Hours
}

var weeklyCmd = &cobra.Command{
	Use:               "weekly [PROJECT]",
	Short:             "Totals hours for a project or task per week",
	ValidArgsFunction: projectCompletionFunc,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var sumProject projectArg
		if len(args) > 0 {
			if err := sumProject.Set(args[0]); err != nil {
				return err
			}
			if weeklyTask.projectId != nil && *weeklyTask.projectId != *sumProject.projectId {
				return errors.New("provided project and task project do not match")
			}
		} else if weeklyTask.projectId != nil {
			sumProject.projectId = weeklyTask.projectId
		} else {
			if sumProject.projectId, err = selectProject(ctx); err != nil {
				return err
			}
		}

		userId, err := getAndSaveUserId(ctx)
		if err != nil {
			return err
		}

		// get entries
		options := harvest.EntryListOptions{
			To:        weeklyToDate.date,
			From:      weeklyFromDate.date,
			TaskId:    weeklyTask.taskId,
			ProjectId: sumProject.projectId,
			UserId:    userId,
		}

		// get entries
		entries, err := harvest.GetEntries(&options, ctx)
		if err != nil {
			return err
		}

		sums := make(map[time.Time]Hours)
		for _, e := range entries {
			if weeklyBillableOnly && !e.Billable {
				continue
			}
			date, _ := time.Parse(time.DateOnly, e.Date)
			week := util.StartOfWeek(date)
			sums[week] += e.Hours
		}

		weeks := make([]weeklyEntry, 0, len(sums))
		for w, s := range sums {
			weeks = append(weeks, weeklyEntry{
				Week:  w,
				Hours: s,
			})
		}
		weeks = slices.SortedFunc(slices.Values(weeks), func(a weeklyEntry, b weeklyEntry) int {
			if b.Week.Before(a.Week) {
				return 1
			}
			if b.Week.After(a.Week) {
				return -1
			}
			return 0
		})

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return weeklyOutputSimple(weeks) },
			config.OutputFormatTable:  func() error { return weeklyOutputTable(weeks) },
			config.OutputFormatJson:   func() error { return outputJson(weeks) },
		})
	}),
}

func init() {
	rootCmd.AddCommand(weeklyCmd)
	weeklyCmd.Flags().VarP(&weeklyTask, "task", "t", "Task ID/alias by which to filter")
	weeklyCmd.Flags().Var(&weeklyToDate, "to", "Date by which to filter by entries on or before [see date section in root]")
	weeklyCmd.Flags().Var(&weeklyFromDate, "from", "Date by which to filter by entries on or after [see date section in root]")
	weeklyCmd.Flags().BoolVar(&weeklyBillableOnly, "billable-only", false, "Only include billable entries")
	weeklyCmd.Flags().Var(&weeklyExpected, "expected", "Number of hours per week you're supposed to work - shows an extra total")

	_ = weeklyCmd.RegisterFlagCompletionFunc("task", taskCompletionFunc)
}

func calculateExpectedWeeklyTotalHours(weeks []weeklyEntry) *Hours {
	if weeklyExpected.hours == nil {
		return nil
	}
	start := weeklyFromDate.date
	if start == nil {
		if len(weeks) == 0 {
			return nil
		}
		start = &weeks[0].Week
	}
	*start = util.StartOfWeek(*start)
	weekCount := math.Ceil(float64(util.WeekdaysBetween(*start, time.Now())) / 5.0)

	return ptr(Hours(weekCount * float64(*weeklyExpected.hours)))
}

func weeklyOutputSimple(weeks []weeklyEntry) error {
	total := Hours(0)
	for _, s := range weeks {
		fmt.Printf("%v\t%v\n", s.Week.Format(time.DateOnly), fmtHours(&s.Hours))
		total += s.Hours
	}
	expectedHours := calculateExpectedWeeklyTotalHours(weeks)
	if expectedHours == nil {
		return nil
	}
	fmt.Println()
	fmt.Printf("Total:\t\t%s\n", fmtHours(&total))
	fmt.Printf("Expected:\t%s\n", fmtHours(expectedHours))
	diff := Hours(float64(total) - float64(*expectedHours))
	if float64(diff) > 0 {
		fmt.Printf("Over:\t\t%s\n", fmtHours(&diff))
	} else {
		fmt.Printf("Under:\t\t%s\n", fmtHours(&diff))
	}
	return nil
}

func weeklyOutputTable(weeks []weeklyEntry) error {
	table := createTable([]string{"Monday", "Total"})
	for _, w := range weeks {
		table.Append([]string{
			w.Week.Format(time.DateOnly),
			fmtHours(&w.Hours),
		})
	}
	table.Render()
	return nil
}
