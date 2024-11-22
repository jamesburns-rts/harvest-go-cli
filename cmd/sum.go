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
	"maps"
	"slices"
	"time"
)

var sumTask taskArg
var sumToDate dateArg
var sumFromDate dateArg

type sumEntry struct {
	Task     harvest.Task
	Billable bool
	Start    time.Time
	Hours    Hours
}

var sumCmd = &cobra.Command{
	Use:               "sum [PROJECT]",
	Short:             "Totals hours for a project or task, broken up by task",
	ValidArgsFunction: projectCompletionFunc,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var sumProject projectArg
		if len(args) > 0 {
			if err := sumProject.Set(args[0]); err != nil {
				return err
			}
			if sumTask.projectId != nil && *sumTask.projectId != *sumProject.projectId {
				return errors.New("provided project and task project do not match")
			}
		} else if sumTask.projectId != nil {
			sumProject.projectId = sumTask.projectId
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
			To:        sumToDate.date,
			From:      sumFromDate.date,
			TaskId:    sumTask.taskId,
			ProjectId: sumProject.projectId,
			UserId:    userId,
		}

		// get entries
		entries, err := harvest.GetEntries(&options, ctx)
		if err != nil {
			return err
		}

		sums := make(map[int64]sumEntry)
		for _, e := range entries {
			key := e.Task.ID
			date, _ := time.Parse(time.DateOnly, e.Date)
			if _, ok := sums[key]; !ok {
				sums[key] = sumEntry{
					Task:     e.Task,
					Billable: e.Billable,
					Start:    date,
					Hours:    e.Hours,
				}
				continue
			}

			sum := sums[key]
			sum.Hours += e.Hours
			if date.Before(sum.Start) {
				sum.Start = date
			}
			sums[key] = sum
		}

		list := slices.SortedFunc(maps.Values(sums), func(a sumEntry, b sumEntry) int {
			if b.Hours > a.Hours {
				return 1
			}
			if b.Hours < a.Hours {
				return -1
			}
			return 0
		})

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return sumOutputSimple(list) },
			config.OutputFormatTable:  func() error { return sumOutputTable(list) },
			config.OutputFormatJson:   func() error { return outputJson(list) },
		})
	}),
}

func init() {
	rootCmd.AddCommand(sumCmd)
	sumCmd.Flags().VarP(&sumTask, "task", "t", "Task ID/alias by which to filter")
	sumCmd.Flags().Var(&sumToDate, "to", "Date by which to filter by entries on or before [see date section in root]")
	sumCmd.Flags().Var(&sumFromDate, "from", "Date by which to filter by entries on or after [see date section in root]")

	_ = sumCmd.RegisterFlagCompletionFunc("task", taskCompletionFunc)
}

func sumOutputSimple(sums []sumEntry) error {
	for _, s := range sums {
		fmt.Printf("%v\t%v\n", fmtHours(&s.Hours), s.Task.Name)
	}
	return nil
}

func sumOutputTable(sums []sumEntry) error {
	table := createTable([]string{"Task Name", "Billable", "Start", "Time Spent", "Per Week"})
	divider := []string{"-------------------------------------", "--------", "---------", "----------", "--------"}
	total := sumEntry{
		Task:  harvest.Task{Name: "Total"},
		Start: time.Now(),
	}
	totalBillable := sumEntry{
		Task:     harvest.Task{Name: "Total Billable"},
		Billable: true,
		Start:    time.Now(),
	}

	fmtRow := func(s sumEntry) []string {
		perWeek := s.Hours / Hours(float32(util.WeekdaysBetween(s.Start, time.Now()))/5)
		return []string{
			s.Task.Name,
			ifOr(s.Billable, "Y", "N"),
			s.Start.Format(time.DateOnly),
			fmtHours(&s.Hours),
			fmtHours(&perWeek),
		}
	}

	for _, s := range sums {
		table.Append(fmtRow(s))
		total.Hours += s.Hours
		if s.Start.Before(total.Start) {
			total.Start = s.Start
		}
		if s.Billable {
			totalBillable.Hours += s.Hours
			if s.Start.Before(totalBillable.Start) {
				totalBillable.Start = s.Start
			}
		}
	}

	table.Append(divider)

	totalRow := fmtRow(total)
	totalRow[1] = "Y+N"
	table.Append(totalRow)
	table.Append(fmtRow(totalBillable))
	table.Render()
	return nil
}
