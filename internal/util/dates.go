package util

import (
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StartOfDay gets the time at midnight the night before
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// StringToDate takes a string of many forms and returns the parsed time
// valid forms include:
// * yyyy-mm-dd
// * yyyymmdd
// * today
// * Monday (the last Monday)
// * fri (the last Friday)
// * -2 (two days ago)
// * yest (yesterday)
// * next (next weekday)
// * last (last weekday)
func StringToDate(str string) (date *time.Time, err error) {
	return stringToDateFrom(str, time.Now())
}

// DateOfLastWeekDay Returns the date of the last occurrence of the given weekday
func DateOfLastWeekDay(weekday time.Weekday) time.Time {
	return dateOfLastWeekDayFrom(weekday, time.Now())
}

func dateOfLastWeekDayFrom(weekday time.Weekday, now time.Time) time.Time {
	t := now.AddDate(0, 0, -1)
	for t.Weekday() != weekday {
		t = t.AddDate(0, 0, -1)
	}
	return StartOfDay(t)
}

func stringToDateFrom(str string, now time.Time) (date *time.Time, err error) {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)

	if str == "" {
		return nil, err
	}

	now = StartOfDay(now)

	defer func() {
		if date == nil {
			err = errors.New("invalid date given [" + str + "]")
		}
	}()

	if str == "" {
		return
	}

	// if yyyy-mm-dd or something similar
	if validDate.MatchString(str) {
		str = strings.ReplaceAll(str, "-", "")
		if t, err := time.ParseInLocation(validDateLayout, str, time.Local); err == nil {
			date = &t
		}
		return
	}

	// check days of week
	for k, v := range dayOfWeekMap {
		if strings.HasPrefix(str, k) {
			t := dateOfLastWeekDayFrom(v, now)
			return &t, nil
		}
	}

	// check relative

	var delta int64
	switch {
	case strings.HasPrefix(str, "last"):
		delta = -1
		if now.Weekday() == time.Monday {
			delta = -3
		}
	case strings.HasPrefix(str, "yest"):
		delta = -1
	case strings.HasPrefix(str, "tod"):
		delta = 0
	case strings.HasPrefix(str, "tom"):
		delta = 1
	case strings.HasPrefix(str, "next"):
		delta = 1
		if now.Weekday() == time.Friday {
			delta = 3
		}
	case validDelta.MatchString(str):
		delta, _ = strconv.ParseInt(str, 10, 64)
	default:
		return
	}

	t := now.AddDate(0, 0, int(delta))
	date = &t

	return
}

var dayOfWeekMap = map[string]time.Weekday{
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
	"sun": time.Sunday,
}
var validDate = regexp.MustCompile(`^\d{4}-?\d{2}-?\d{2}$`)
var validDelta = regexp.MustCompile(`^-?\d+$`)

const validDateLayout = "20060102"
