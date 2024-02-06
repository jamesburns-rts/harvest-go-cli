package harvest

import (
	"context"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"time"
)

type (
	MonthSummary struct {
		RequiredHours    Hours
		MonthLoggedHours Hours
		WeekLoggedHours  Hours
		BillableHours    Hours
		NonBillableHours Hours
		TodayLoggedHours Hours
		Short            Hours
		ShortWeek        Hours
	}
)

func CalculateMonthSummary(t time.Time, userId *int64, ctx context.Context) (MonthSummary, error) {

	startOfMonth := util.StartOfMonth(t)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	startOfWeek := util.StartOfWeek(t)
	startOfNextWeek := startOfWeek.AddDate(0, 0, 7)

	earliestPoint := startOfMonth
	if startOfWeek.Before(startOfMonth) {
		earliestPoint = startOfWeek
	}
	latestPoint := startOfNextMonth
	if startOfNextWeek.After(startOfNextMonth) {
		latestPoint = startOfNextWeek
	}

	entries, err := GetEntries(&EntryListOptions{
		From:   &earliestPoint,
		To:     &latestPoint,
		UserId: userId,
	}, ctx)
	if err != nil {
		return MonthSummary{}, err
	}

	weekDaysInMonth := util.WeekdaysBetween(startOfMonth, startOfNextMonth)

	summary := MonthSummary{
		RequiredHours: Hours(8 * weekDaysInMonth),
	}

	for _, e := range entries {
		date, _ := util.StringToDate(e.Date)

		// check for today
		if util.SameDay(*date, time.Now()) {
			summary.TodayLoggedHours += e.Hours
		}

		// total hours for month
		if !date.Before(startOfMonth) && date.Before(startOfNextMonth) {
			if e.Billable {
				summary.BillableHours += e.Hours
			} else {
				summary.NonBillableHours += e.Hours
			}
		}
		if !date.Before(startOfWeek) && date.Before(startOfNextWeek) {
			summary.WeekLoggedHours += e.Hours
		}
	}

	summary.MonthLoggedHours = summary.BillableHours + summary.NonBillableHours

	// short
	daysSoFar := util.WeekdaysBetween(startOfMonth, time.Now())
	weekdaysInWeekSoFar := util.WeekdaysBetween(startOfWeek, t.AddDate(0, 0, 1))
	summary.Short = Hours(daysSoFar*8) - summary.MonthLoggedHours
	summary.ShortWeek = Hours(weekdaysInWeekSoFar*8) - summary.WeekLoggedHours

	return summary, nil
}
