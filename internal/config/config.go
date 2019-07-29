package config

import (
	"fmt"
	"time"
)

type (
	HarvestProperties struct {
		AccessToken    string           `yaml:"accessToken"json:"accessToken"`
		AccountId      string           `yaml:"accountId"json:"accessToken"`
		ProjectAliases map[string]int64 `yaml:"projectAliases"json:"projectAliases"`
		TaskAliases    map[string]int64 `yaml:"taskAliases"json:"taskAliases"`
	}

	CliProperties struct {
		TimeDeltaFormat     string
		DefaultOutputFormat string
	}

	TimerRecords struct {
		Arrived string `yaml:"arrived"json:"arrived"`
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
var Timers TimerRecords

func (r *TimerRecords) SetArrived(t time.Time) {
	r.Arrived = t.Format(time.RFC3339)
}

func (r *TimerRecords) ArrivedTime() *time.Time {
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
