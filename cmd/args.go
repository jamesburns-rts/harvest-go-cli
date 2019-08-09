package cmd

import (
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"strconv"
	"time"
)

type (
	taskArg struct {
		str       string
		taskId    *int64
		projectId *int64
	}
	projectArg struct {
		str       string
		projectId *int64
	}
	hoursArg struct {
		str   string
		hours *Hours
	}
	stringArg struct {
		str string
	}
	dateArg struct {
		str  string
		date *time.Time
	}
)

// task arg
func (t *taskArg) String() string {
	return t.str
}

func (t *taskArg) Set(str string) (err error) {
	t.str = str
	t.projectId, t.taskId, err = harvest.ParseTaskId(str)
	return err
}

func (t *taskArg) Type() string {
	return "task"
}

func (t *taskArg) SetId(taskId, projectId *int64) {

	t.taskId = taskId
	t.projectId = projectId
	if taskId == nil {
		t.str = ""
	} else {
		t.str = strconv.FormatInt(*taskId, 10)
	}
}

// project arg
func (p *projectArg) String() string {
	return p.str
}

func (p *projectArg) Set(str string) error {
	if projectId, err := harvest.ParseProjectId(str); err != nil {
		return err
	} else {
		p.SetId(projectId)
	}
	return nil
}

func (p *projectArg) Type() string {
	return "project"
}

func (p *projectArg) SetId(projectId *int64) {
	p.projectId = projectId
	if projectId == nil {
		p.str = ""
	} else {
		p.str = strconv.FormatInt(*projectId, 10)
	}
}

// hours arg
func (h *hoursArg) String() string {
	return h.str
}

func (h *hoursArg) Set(str string) (err error) {
	h.str = str
	h.hours, err = ParseHours(str)
	return err
}

func (h *hoursArg) Type() string {
	return "hours"
}

func (h *hoursArg) SetHours(hours *Hours) {
	h.hours = hours
	if hours == nil {
		h.str = ""
	} else {
		h.str = hours.Duration().String()
	}
}

func (h *hoursArg) prompt(title string) error {
	str, err := prompt.ForString(title, validHours)
	if err != nil {
		return err
	}
	return h.Set(str)
}

// string arg
func (s *stringArg) String() string {
	return s.str
}

func (s *stringArg) Set(str string) (err error) {
	s.str = str
	return nil
}

func (s *stringArg) Type() string {
	return "string"
}

func (s *stringArg) prompt(title string) error {
	str, err := prompt.ForString(title, nil)
	if err != nil {
		return err
	}
	return s.Set(str)
}

// date arg
func (d *dateArg) String() string {
	return d.str
}

func (d *dateArg) Set(str string) (err error) {
	d.str = str
	d.date, err = util.StringToDate(str)
	return err
}

func (d *dateArg) Type() string {
	return "date"
}

func (d *dateArg) SetDate(date *time.Time) {
	d.date = date
	if date == nil {
		d.str = ""
	} else {
		d.str = date.Format("2006-01-02")
	}
}
