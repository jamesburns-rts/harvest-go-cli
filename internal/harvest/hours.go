package harvest

import (
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"math"
)

type (
	Hours float64
	MonthSummary struct {
		RequiredHours    Hours
		MonthLoggedHours Hours
		BillableHours    Hours
		NonBillableHours Hours
		WorkedTodayHours Hours
		TodayLoggedHours Hours
	}
)

func (h Hours) Minutes() float64 {
	return 60 * (float64(h) - h.Hours())
}

func (h Hours) Hours() float64 {
	return math.Floor(float64(h))
}

func (h Hours) String() string {
	if config.Cli.TimeDeltaFormat == config.TimeDeltaFormatHuman {
		if h < 1 {
			return fmt.Sprintf("%0.0fm", h.Minutes())
		}
		return fmt.Sprintf("%0.0fh %0.0fm", h.Hours(), h.Minutes())
	}

	// else config.TimeDeltaFormatDecimal or other
	return fmt.Sprintf("%0.2f", float64(h))
}
