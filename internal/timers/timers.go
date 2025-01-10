package timers

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"errors"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
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
		Arrived string  `yaml,json:"arrived"`
		Timers  []Timer `yaml,json:"timers"`
	}
)

var Records TrackingRecords

func Set(t Timer) {
	for i, v := range Records.Timers {
		if v.Name == t.Name {
			Records.Timers[i] = t
			return
		}
	}
	Records.Timers = append(Records.Timers, t)
}

func Get(name string) (Timer, bool) {
	for _, v := range Records.Timers {
		if v.Name == name {
			return v, true
		}
	}
	return Timer{}, false
}

func Delete(name string) {
	for i, t := range Records.Timers {
		if t.Name == name {
			Records.Timers[i] = Records.Timers[len(Records.Timers)-1]
			Records.Timers = Records.Timers[:len(Records.Timers)-1]
			return
		}
	}
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
					Notes:   &t.Notes,
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
	Set(*t)
	return nil
}

func (t *Timer) Stop(preventSync bool, ctx context.Context) (err error) {
	if !t.Running {
		return nil
	}
	t.Duration = *t.RunningHours()
	t.Running = false

	if t.SyncedEntryId != nil {
		var entry harvest.Entry
		if entry, err = harvest.StopTimerEntry(*t.SyncedEntryId, ctx); err != nil {
			return err
		}

		t.compareNotes(entry.Notes)

		// out of sync
		if math.Abs(float64(entry.Hours-t.Duration)) > 0.1 {
			_, err := harvest.UpdateEntry(harvest.EntryUpdateOptions{
				Entry: entry,
				Hours: &t.Duration,
			}, ctx)
			if err != nil {
				return err
			}
		}
	}
	Set(*t)
	return nil
}

func (t *Timer) compareNotes(entryNotes string) {
	lines := strings.Split(t.Notes, "\n")
	for _, line := range strings.Split(entryNotes, "\n") {
		if _, ok := util.Contains(lines, line); !ok {
			lines = append(lines, line)
		}
	}
	t.Notes = strings.Join(lines, "\n")
}

func (t *Timer) AppendNotes(notes string) {
	if notes == "" {
		return
	}
	if t.Notes == "" {
		t.Notes = notes
		return
	}
	t.Notes += "\n" + notes
}

func SumTimeOn(names []string) (total Hours) {
	for _, t := range Records.Timers {
		if _, match := util.ContainsIgnoreCase(names, t.Name); match {
			total += *t.RunningHours()
		}
	}
	return total
}
