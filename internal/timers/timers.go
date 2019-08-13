package timers

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/pkg/errors"
	"log"
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

func SetTimer(t Timer) {
	if Records.Timers == nil {
		Records.Timers = make(map[string]Timer)
	}
	Records.Timers[t.Name] = t
}

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

func (t *Timer) StartedTime() time.Time {
	if t.Started == "" {
		log.Fatal(fmt.Sprintf("Timer %s has no start time", t.Name))
	}

	tm, err := time.Parse(time.RFC3339, t.Started)
	if err != nil {
		log.Fatal(fmt.Sprintf("Timer %s has invalid start time", t.Name))
	}

	return tm
}

func (t *Timer) RunningHours() *Hours {
	dur := t.Duration
	if t.Running {
		dur += Hours(time.Now().Sub(t.StartedTime()).Hours())
	}
	return &dur
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
				started := t.StartedTime()
				_, err := harvest.UpdateEntry(harvest.EntryUpdateOptions{
					Entry:   entry,
					Started: &started,
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
	SetTimer(*t)
	return nil
}

func (t *Timer) Stop(preventSync bool, ctx context.Context) (err error) {
	if !t.Running {
		return nil
	}
	t.Duration = *t.RunningHours()
	t.Running = false
	t.Started = ""

	if t.SyncedEntryId != nil {
		var entry harvest.Entry
		if entry, err = harvest.GetEntry(*t.SyncedEntryId, ctx); err != nil {
			return err
		}

		t.compareNotes(entry.Notes)

		// out of sync
		_, err := harvest.UpdateEntry(harvest.EntryUpdateOptions{
			Entry: entry,
			Hours: &t.Duration,
		}, ctx)
		return err
	}
	SetTimer(*t)
	return nil
}

func (t *Timer) compareNotes(entryNotes string) {

}

func SumTimeOn(names []string) (total Hours) {
	for _, t := range Records.Timers {
		if _, match := util.ContainsIgnoreCase(names, t.Name); match {
			total += *t.RunningHours()
		}
	}
	return total
}
