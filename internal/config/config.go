package config

import (
	"fmt"
	"time"
)

type (
	HarvestProperties struct {
		AccessToken    string           `yaml,json:"accessToken"`
		AccountId      string           `yaml,json:"accountId"`
		ProjectAliases map[string]int64 `yaml,json:"projectAliases"`
		TaskAliases    map[string]int64 `yaml,json:"taskAliases"`
		SyncTimers     *bool            `yaml,json:"syncTimers"`
	}

	CliProperties struct {
		TimeDeltaFormat     string `yaml,json:"timeDeltaFormat"`
		DefaultOutputFormat string `yaml,json:"defaultOutputFormat"`
		DisplayAliases      *bool `yaml,json:"displayAliases"`
	}

	Timer struct {
		Name string `yaml,json:"name"`
	}

	TrackingRecords struct {
		Arrived string  `yaml,json:"arrived"`
		Timers  []Timer `yaml,json:"timers"`
	}
)

const (
	TimeDeltaFormatDecimal = "decimal"
	TimeDeltaFormatHuman   = "human"

	OutputFormatJson   = "json"
	OutputFormatSimple = "simple"
	OutputFormatTable  = "table"
)

var Harvest HarvestProperties
var Cli CliProperties
var Tracking TrackingRecords

func (r *TrackingRecords) SetArrived(t time.Time) {
	r.Arrived = t.Format(time.RFC3339)
}

func (r *TrackingRecords) ArrivedTime() *time.Time {
	if r.Arrived == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, r.Arrived)
	if err != nil {
		fmt.Println("Warning: Bad time format in 'timers.arrived'")
		return nil
	}

	return &t
}
