package timers

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/pkg/errors"
	"time"
)

type (
	Timer struct {
		Name            string `json:"name"`
		Running         bool   `json:"running"`
		Started         string `json:"started"`
		Duration        Hours  `json:"duration"`
		SyncedProjectId *int64 `json:"syncedProjectId"`
		SyncedTaskId    *int64 `json:"syncedTaskId"`
		SyncedEntryId   *int64 `json:"syncedEntryId"`
		Notes           string `json:"notes"`
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

func (t *Timer) SetStarted(tm time.Time) {
	t.Started = tm.Format(time.RFC3339)
}

func (t *Timer) ClearStarted(tm time.Time) {
	t.Started = ""
}

func (t *Timer) StartedTime() *time.Time {
	if t.Started == "" {
		return nil
	}

	tm, err := time.Parse(time.RFC3339, t.Started)
	if err != nil {
		fmt.Println("Warning: Bad time format in 'timers.timers[].started'")
		return nil
	}

	return &tm
}

func (t *Timer) Start(preventSync bool, ctx context.Context) (err error) {
	if t.Running {
		return nil
	}
	t.Running = true
	t.SetStarted(time.Now())
	if t.SyncedTaskId != nil && !preventSync {

		if t.SyncedProjectId == nil {
			return errors.New("timer has taskId but no projectId")
		}

		if t.SyncedEntryId != nil {
			var entry harvest.Entry
			if entry, err = harvest.GetEntry(*t.SyncedEntryId, ctx); err != nil {
				return err
			}

			t.compareNotes(entry.Notes)
			t.Duration = entry.Hours
			t.SyncedTaskId = &entry.Task.ID
			t.SyncedProjectId = &entry.Project.ID

			// out of sync
			if entry.Running {
				t.SetStarted(*entry.TimerStarted)
			} else {
				_, err := harvest.UpdateEntry(harvest.EntryUpdateOptions{
					Entry:   entry,
					Started: t.StartedTime(),
				}, ctx)
				return err
			}
		} else {

			startTime := t.StartedTime().Add(-t.Duration.Duration())

			entry, err := harvest.StartTimerEntry(harvest.TimerStartOptions{
				TaskId:    *t.SyncedTaskId,
				ProjectId: *t.SyncedProjectId,
				StartTime: &startTime,
				Notes:     &t.Notes,
			}, ctx)
			if err != nil {
				return err
			}
			t.SyncedEntryId = &entry.ID
		}
	}
	return nil
}

func (t *Timer) Stop(preventSync bool, ctx context.Context) (err error) {
	if !t.Running {
		return nil
	}
	t.Running = false
	startedTime := t.StartedTime()
	if startedTime == nil {
		return errors.New("no start time started noted")
	}
	//dur := t.Duration.Duration() + time.Now().Sub(*startedTime)
	//t.Started = ""
	return nil
}

func (t *Timer) compareNotes(entryNotes string) {

}

func Start(name, notes string, taskId, projectId, entryId *int64) (Timer, error) {
	if existing, ok := Records.Timers[name]; ok {
		if taskId != nil {
			existing.SyncedTaskId = taskId
			existing.SyncedProjectId = projectId
			existing.SyncedEntryId = entryId
		}
		if !existing.Running {
			existing.Running = true
			existing.SetStarted(time.Now())
		}
		if notes != "" {
			existing.Notes += "\n" + notes
		}
	} else {
	}
	return Timer{}, nil
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
		Duration:      entry.Hours,
		SyncedTaskId:  &entry.Task.ID,
		SyncedEntryId: &entry.ID,
		Notes:         entry.Notes,
	}
	timer.SetStarted(*entry.TimerStarted)

	Records.Timers[name] = timer

	return timer, nil
}
