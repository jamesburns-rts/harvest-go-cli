package util

import (
	"reflect"
	"testing"
	"time"
)

func testTime(month time.Month, date int) time.Time {
	return time.Date(2019, month, date, 13, 23, 49, 23, time.Local)
}

func date(year int, month time.Month, date int) *time.Time {
	d := time.Date(year, month, date, 0, 0, 0, 0, time.Local)
	return &d
}

func TestStringToDate(t *testing.T) {
	type args struct {
		str string
		now time.Time
	}
	tests := []struct {
		name     string
		args     args
		wantDate *time.Time
		wantErr  bool
	}{
		{
			"empty",
			args{"", time.Now()},
			nil,
			true,
		},
		{
			"yest",
			args{"yest", testTime(time.July, 5)},
			date(2019, time.July, 4),
			false,
		},
		{
			"today",
			args{"tod", testTime(time.July, 5)},
			date(2019, time.July, 5),
			false,
		},
		{
			"tomorrow",
			args{"tom", testTime(time.July, 5)},
			date(2019, time.July, 6),
			false,
		},
		{
			"next",
			args{"next", testTime(time.July, 3)},
			date(2019, time.July, 4),
			false,
		},
		{
			"next over weekend",
			args{"next", testTime(time.July, 5)},
			date(2019, time.July, 8),
			false,
		},
		{
			"last",
			args{"last", testTime(time.July, 5)},
			date(2019, time.July, 4),
			false,
		},
		{
			"last over weekend",
			args{"last", testTime(time.July, 1)},
			date(2019, time.June, 28),
			false,
		},
		{
			"mon",
			args{"mon", testTime(time.July, 5)},
			date(2019, time.July, 1),
			false,
		},
		{
			"tue",
			args{"tue", testTime(time.July, 5)},
			date(2019, time.July, 2),
			false,
		},
		{
			"fri",
			args{"fri", testTime(time.July, 5)},
			date(2019, time.June, 28),
			false,
		},
		{
			"-4",
			args{"-4", testTime(time.July, 5)},
			date(2019, time.July, 1),
			false,
		},
		{
			"-3",
			args{"-3", testTime(time.July, 5)},
			date(2019, time.July, 2),
			false,
		},
		{
			"-7",
			args{"fri", testTime(time.July, 5)},
			date(2019, time.June, 28),
			false,
		},
		{
			"all digits",
			args{"20190705", time.Now()},
			date(2019, time.July, 5),
			false,
		},
		{
			"with hyphens",
			args{"2019-07-05", time.Now()},
			date(2019, time.July, 5),
			false,
		},
	}
	for _, tt := range tests {
		gotDate, err := stringToDateFrom(tt.args.str, tt.args.now)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. StringToDate() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(gotDate, tt.wantDate) {
			t.Errorf("%q. StringToDate() = %v, want %v", tt.name, gotDate, tt.wantDate)
		}
	}
}
