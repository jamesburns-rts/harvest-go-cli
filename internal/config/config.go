package config

import (
	"fmt"
	"strings"

	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
)

type (
	HarvestProperties struct {
		AccessToken    string                  `yaml,json:"accessToken"`
		AccountId      string                  `yaml,json:"accountId"`
		ProjectAliases map[string]ProjectAlias `yaml,json:"projectAliases"`
		TaskAliases    map[string]TaskAlias    `yaml,json:"taskAliases"`
		Projects       []ProjectAlias          `yaml,json:"projects"`
		Tasks          []TaskAlias             `yaml,json:"tasks"`
		SyncTimers     *bool                   `yaml,json:"syncTimers"`
		UserId         *int64                  `yaml,json:"userId"`
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

func GetTaskAlias(name string) (TaskAlias, bool) {
	for _, t := range Harvest.Tasks {
		if t.Name == name {
			return t, true
		}
	}
	return TaskAlias{}, false
}

func SetTaskAlias(a TaskAlias) {
	for i, t := range Harvest.Tasks {
		if t.Name == a.Name {
			Harvest.Tasks[i] = a
			return
		}
	}
	Harvest.Tasks = append(Harvest.Tasks, a)
}

func DeleteTaskAlias(name string) {
	for i, t := range Harvest.Tasks {
		if t.Name == name {
			Harvest.Tasks = append(Harvest.Tasks[:i], Harvest.Tasks[i+1:]...)
			return
		}
	}
}

func GetProjectAlias(name string) (ProjectAlias, bool) {
	for _, p := range Harvest.Projects {
		if p.Name == name {
			return p, true
		}
	}
	return ProjectAlias{}, false
}

func SetProjectAlias(a ProjectAlias) {
	for i, p := range Harvest.Projects {
		if p.Name == a.Name {
			Harvest.Projects[i] = a
			return
		}
	}
	Harvest.Projects = append(Harvest.Projects, a)
}

func DeleteProjectAlias(name string) {
	for i, p := range Harvest.Projects {
		if p.Name == name {
			Harvest.Projects = append(Harvest.Projects[:i], Harvest.Projects[i+1:]...)
			return
		}
	}
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
