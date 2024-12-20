package util

import (
	"errors"
	"fmt"
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

// StartOfMonth StartOfDay gets the time at midnight on the first of the month
func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

func StartOfWeek(t time.Time) time.Time {
	var delta int
	if t.Weekday() == time.Sunday {
		delta = -6
	} else {
		delta = -int(t.Weekday() - time.Monday)
	}
	return StartOfDay(t.AddDate(0, 0, delta))
}

// WeekdaysBetween get the total number of workable weekdays in the month
// includes start but not stop
func WeekdaysBetween(start, stop time.Time) int {
	if start.After(stop) {
		start, stop = stop, start
	}
	totalDays := 0
	for !SameDay(start, stop) {
		switch start.Weekday() {
		case time.Monday:
			fallthrough
		case time.Tuesday:
			fallthrough
		case time.Wednesday:
			fallthrough
		case time.Thursday:
			fallthrough
		case time.Friday:
			totalDays++
		default:
			break
		}
		start = start.AddDate(0, 0, 1)
	}
	return totalDays
}

// SameDay checks the two times are of the same day
func SameDay(t1, t2 time.Time) bool {
	year1, month1, date1 := t1.Date()
	year2, month2, date2 := t2.Date()
	return year1 == year2 && month1 == month2 && date1 == date2
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

func dateOfLastWeekDayFrom(weekday time.Weekday, now time.Time) time.Time {
	t := now.AddDate(0, 0, -1)
	for t.Weekday() != weekday {
		t = t.AddDate(0, 0, -1)
	}
	return StartOfDay(t)
}

func stringToDateFrom(str string, now time.Time) (*time.Time, error) {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)

	if str == "" {
		return nil, errors.New("empty string given")
	}

	now = StartOfDay(now)

	// if yyyy-mm-dd or something similar
	if justMonthAndDate.MatchString(str) {
		year := now.Format("2006")
		if strings.ContainsRune(str, '-') {
			year = year + "-"
		}
		str = year + str
	}
	if validDate.MatchString(str) {
		str = strings.ReplaceAll(str, "-", "")
		t, err := time.ParseInLocation(validDateLayout, str, time.Local)
		if err != nil {
			return nil, fmt.Errorf("expecting %s (or with hyphens): %w", validDateLayout, err)
		}
		return &t, nil
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
		return nil, fmt.Errorf("unexpected date word: %s", str)
	}

	t := now.AddDate(0, 0, int(delta))
	return &t, nil
}

func StringToTime(str string) (time.Time, error) {
	now := time.Now()

	str = strings.ToUpper(str)
	str = strings.ReplaceAll(str, "P.M.", "PM")
	str = strings.ReplaceAll(str, "A.M.", "AM")

	layouts := []string{time.Kitchen, "15:04"}
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, str, time.Local)
		if err == nil {
			return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, t.Location()), nil
		}
	}
	return now, errors.New("expected format of hh:mm or h:mmPM")
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
var justMonthAndDate = regexp.MustCompile(`^\d{2}-?\d{2}$`)
var validDate = regexp.MustCompile(`^\d{4}-?\d{2}-?\d{2}$`)
var validDelta = regexp.MustCompile(`^-?\d+$`)

const validDateLayout = "20060102"
