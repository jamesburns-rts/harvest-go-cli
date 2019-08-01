package timers

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"time"
)

type (
	Timer struct {
		Name          string `yaml,json:"name"`
		Running       bool   `yaml,json:"running"`
		Started       string `yaml,json:"started"`
		Duration      string `yaml,json:"duration"`
		SyncedTaskId  *int64 `yaml,json:"syncedTaskId"`
		SyncedEntryId *int64 `yaml,json:"syncedEntryId"`
		Notes         string `yaml,json:"notes"`
	}

	TrackingRecords struct {
		Arrived string           `yaml,json:"arrived"`
		Timers  map[string]Timer `yaml,json:"timers"`
	}
)

var Records TrackingRecords

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

func (r *Timer) SetStarted(t time.Time) {
	r.Started = t.Format(time.RFC3339)
}

func (r *Timer) ClearStarted(t time.Time) {
	r.Started = ""
}

func (r *Timer) StartedTime() *time.Time {
	if r.Started == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, r.Started)
	if err != nil {
		fmt.Println("Warning: Bad time format in 'timers.timers[].started'")
		return nil
	}

	return &t
}

func StartSynced(name, notes string, timeEntryId int64, ctx context.Context) (Timer, error) {

	entry, err := harvest.RestartTimeEntry(timeEntryId, ctx)
	if err != nil {
		return Timer{}, err
	}

	if notes != "" {
		entry.Notes = fmt.Sprintf("%s\n%s", entry.Notes, notes)
	}

	timer := Timer{
		Name:          name,
		Running:       entry.Running,
		Duration:      entry.Hours.Duration().String(),
		SyncedTaskId:  &entry.Task.ID,
		SyncedEntryId: &entry.ID,
		Notes:         entry.Notes,
	}
	timer.SetStarted(*entry.TimerStarted)

	Records.Timers[name] = timer

	return timer, nil
}
