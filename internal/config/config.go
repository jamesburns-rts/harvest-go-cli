package config

import (
	"fmt"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"strings"
)

type (
	HarvestProperties struct {
		AccessToken    string                  `yaml,json:"accessToken"`
		AccountId      string                  `yaml,json:"accountId"`
		ProjectAliases map[string]ProjectAlias `yaml,json:"projectAliases"`
		TaskAliases    map[string]TaskAlias    `yaml,json:"taskAliases"`
		SyncTimers     *bool                   `yaml,json:"syncTimers"`
	}

	ProjectAlias struct {
		Name      string `yaml,json:"name"`
		ProjectId int64  `yaml,json:"projectId"`
	}

	TaskAlias struct {
		Name            string  `yaml,json:"name"`
		TaskId          int64   `yaml,json:"taskId"`
		ProjectId       int64   `yaml,json:"projectId"`
		DefaultNotes    *string `yaml,json:"defaultNotes"`
		DefaultDuration *Hours  `yaml,json:"defaultDuration"`
	}

	CliProperties struct {
		TimeDeltaFormat     string `yaml,json:"timeDeltaFormat"`
		DefaultOutputFormat string `yaml,json:"defaultOutputFormat"`
		IdDisplay           string `yaml,json:"displayAliases"`
	}
)

var Harvest HarvestProperties
var Cli CliProperties

func SetTaskAlias(a TaskAlias) {
	if Harvest.TaskAliases == nil {
		Harvest.TaskAliases = make(map[string]TaskAlias)
	}
	Harvest.TaskAliases[a.Name] = a
}

func SetProjectAlias(a ProjectAlias) {
	if Harvest.ProjectAliases == nil {
		Harvest.ProjectAliases = make(map[string]ProjectAlias)
	}
	Harvest.ProjectAliases[a.Name] = a
}

type Options []string

const (
	TimeDeltaFormatDecimal = "decimal"
	TimeDeltaFormatHuman   = "human"

	OutputFormatJson   = "json"
	OutputFormatSimple = "simple"
	OutputFormatTable  = "table"
)

var OutputFormatOptions = Options{
	OutputFormatJson,
	OutputFormatSimple,
	OutputFormatTable,
}

var TimeDeltaFormatOptions = Options{
	TimeDeltaFormatDecimal,
	TimeDeltaFormatHuman,
}

func (o Options) String() string {
	return fmt.Sprintf("[%s]", strings.Join([]string(o), ", "))
}

func (o Options) Contains(str string) (string, bool) {
	return util.ContainsIgnoreCase([]string(o), str)
}
