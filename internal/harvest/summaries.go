package harvest

import (
	"context"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"time"
)

type (
	MonthSummary struct {
		RequiredHours    float64
		MonthLoggedHours float64
		BillableHours    float64
		NonBillableHours float64
		WorkedTodayHours float64
		TodayLoggedHours float64
	}
)

func CalculateMonthSummary(t time.Time, ctx context.Context) (MonthSummary, error) {

	startOfMonth := util.StartOfMonth(t)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)

	entries, err := GetEntries(&EntryListOptions{
		From: &startOfMonth,
		To:   &startOfNextMonth,
	}, ctx)
	if err != nil {
		return MonthSummary{}, err
	}

	weekDays := util.WeekdaysBetween(startOfMonth, startOfNextMonth)

	summary := MonthSummary{
		RequiredHours: float64(8 * weekDays),
	}

	for _, e := range entries {
		date, _ := util.StringToDate(e.Date)

		// check for today
		if util.SameDay(*date, time.Now()) {
			summary.TodayLoggedHours += e.Hours
		}

		// total hours
		if e.Project.Billable {
			summary.BillableHours += e.Hours
		} else {
			summary.NonBillableHours += e.Hours
		}
	}

	summary.MonthLoggedHours = summary.BillableHours + summary.NonBillableHours

	arrived := config.Timers.ArrivedTime()
	if arrived != nil && util.SameDay(*arrived, time.Now()) {
		summary.TodayLoggedHours = time.Now().Sub(*arrived).Hours()
	}

	return summary, nil
}
