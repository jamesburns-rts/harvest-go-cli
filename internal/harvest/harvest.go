package harvest

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jamesburns-rts/harvest-go-cli/internal/util"

	"github.com/becoded/go-harvest/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type (
	Project struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		ClientName string `json:"client"`
		Billable   bool   `json:"billable"`
		Tasks      []Task `json:"tasks"`
	}

	Task struct {
		ID        int64  `json:"id"`
		ProjectId int64  `json:"projectId"`
		Name      string `json:"name"`
	}

	Entry struct {
		Date         string     `json:"date"`
		Hours        Hours      `json:"hours"`
		ID           int64      `json:"id"`
		Notes        string     `json:"notes"`
		Project      Project    `json:"project"`
		Task         Task       `json:"task"`
		Running      bool       `json:"running"`
		TimerStarted *time.Time `json:"timerStarted"`
		Billable     bool       `json:"billable"`
	}

	EntryListOptions struct {
		To        *time.Time
		From      *time.Time
		ProjectId *int64
		TaskId    *int64
		Running   *bool
		UserId    *int64
	}

	EntryUpdateOptions struct {
		Entry     Entry
		ProjectId *int64
		TaskId    *int64
		Date      *time.Time
		Hours     *Hours
		Notes     *string
		Started   *time.Time
	}

	LogTimeOptions struct {
		TaskId    int64
		ProjectId int64
		Date      time.Time
		Hours     Hours
		Notes     string
	}

	TimerStartOptions struct {
		TaskId    int64
		ProjectId int64
		StartTime *time.Time
		Notes     *string
	}
)

func GetCurrentUserId(ctx context.Context) (*int64, error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, err
	}
	u, _, err := client.User.Current(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}
	return u.ID, nil
}

func createClient(ctx context.Context) (*harvest.APIClient, error) {

	props := config.Harvest

	// check properties are set
	if props.AccessToken == "" {
		return nil, errors.New("you must set harvest access token using harvest set --harvest-access-token")
	}
	if props.AccountId == "" {
		return nil, errors.New("you must set harvest account id using harvest set --harvest-account-id")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: props.AccessToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := harvest.NewAPIClient(tc)
	client.AccountID = props.AccountId

	return client, nil
}

// ParseProjectId either parse the string for an integer or check for an alias
func ParseProjectId(str string) (*int64, error) {
	if str == "" {
		return nil, nil
	}

	if projectAlias, ok := config.GetProjectAlias(str); ok {
		return &projectAlias.ProjectId, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, errors.New("no alias found for " + str)
	}
	return &i, err
}

func ParseTaskId(str string) (taskId, projectId *int64, err error) {
	if str == "" {
		return nil, nil, nil
	}

	if taskAlias, ok := config.GetTaskAlias(str); ok {
		return &taskAlias.TaskId, &taskAlias.ProjectId, nil
	}

	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return &i, nil, nil
	}

	if projectId, err = ParseProjectId(str); err == nil {
		return nil, projectId, nil
	}
	return nil, nil, errors.New("no alias found for " + str)
}

func GetProject(projectId int64, ctx context.Context) (Project, error) {
	projects, err := GetProjects(ctx)
	if err != nil {
		return Project{}, err
	}
	for _, p := range projects {
		if p.ID == projectId {
			return p, nil
		}
	}
	return Project{}, errors.New("Project not found")
}

// GetProjects Get a list of projects and their tasks
func GetProjects(ctx context.Context) (projects []Project, err error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "creating client")
	}
	options := harvest.MyProjectAssignmentListOptions{}
	options.Page = -1
	lastPage := 0

	for options.Page != lastPage {
		options.Page = options.Page + 1

		page, _, err := client.Project.GetMyProjectAssignments(ctx, &options)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range page.UserAssignments {
			if !*p.IsActive {
				continue
			}

			tasks := make([]Task, 0, len(*p.TaskAssignments))
			for _, t := range *p.TaskAssignments {
				if !*t.IsActive {
					continue
				}

				tasks = append(tasks, Task{
					ID:        *t.Task.ID,
					ProjectId: *p.Project.ID,
					Name:      *t.Task.Name,
				})
			}

			projects = append(projects, Project{
				ID:         *p.Project.ID,
				Name:       *p.Project.Name,
				ClientName: *p.Client.Name,
				Billable:   p.Project.IsBillable != nil && *p.Project.IsBillable,
				Tasks:      tasks,
			})
		}
		lastPage = *page.TotalPages - 1
	}

	return projects, nil
}

// GetTasks get a list of tasks (optionally filter by projectId)
func GetTasks(projectId *int64, ctx context.Context) (tasks []Task, err error) {

	projects, err := GetProjects(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting projects")
	}

	for _, project := range projects {
		if projectId == nil || *projectId == project.ID {
			for _, task := range project.Tasks {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

// GetEntries get a list of entries (optionally filtered by the options)
func GetEntries(o *EntryListOptions, ctx context.Context) (entries []Entry, err error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "creating client")
	}

	options := o.toHarvestOptions()

	options.Page = -1
	lastPage := 0
	for options.Page != lastPage {
		options.Page = options.Page + 1

		page, _, err := client.Timesheet.List(ctx, &options)
		if err != nil {
			return entries, err
		}

		for _, e := range page.TimeEntries {

			if !o.includeTask(e.Task.ID) {
				continue
			}

			entries = append(entries, convertEntry(*e))
		}

		lastPage = *page.TotalPages - 1
	}

	return entries, nil
}

func GetEntry(entryId int64, ctx context.Context) (Entry, error) {
	client, err := createClient(ctx)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating client")
	}

	entry, _, err := client.Timesheet.Get(ctx, entryId)
	if err != nil || entry == nil {
		return Entry{}, err
	}

	return convertEntry(*entry), nil
}

func DeleteEntry(entryId int64, ctx context.Context) error {
	client, err := createClient(ctx)
	if err != nil {
		return errors.Wrap(err, "creating client")
	}

	_, err = client.Timesheet.DeleteTimeEntry(ctx, entryId)
	return err
}

func UpdateEntry(options EntryUpdateOptions, ctx context.Context) (Entry, error) {
	client, err := createClient(ctx)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating client")
	}

	harvestOptions := options.toHarvestOptions()

	entry, _, err := client.Timesheet.UpdateTimeEntry(ctx, options.Entry.ID, &harvestOptions)
	if err != nil {
		return Entry{}, errors.Wrap(err, "updating entry")
	}
	return convertEntry(*entry), err
}

func GetTimers(o *EntryListOptions, userId *int64, ctx context.Context) (entries []Entry, err error) {

	if o == nil {
		o = &EntryListOptions{
			UserId: userId,
		}
	}
	o.Running = BoolPtr(true)

	return GetEntries(o, ctx)
}

func StartTimerEntry(o TimerStartOptions, ctx context.Context) (Entry, error) {
	client, err := createClient(ctx)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating client")
	}

	options := o.toHarvestOptions()

	entry, _, err := client.Timesheet.CreateTimeEntryViaStartEndTime(ctx, &options)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating time entry")
	}

	return convertEntry(*entry), nil
}

func StopTimerEntry(entryId int64, ctx context.Context) (Entry, error) {
	client, err := createClient(ctx)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating client")
	}

	entry, _, err := client.Timesheet.StopTimeEntry(ctx, entryId)
	if err != nil {
		return Entry{}, errors.Wrap(err, "stopping time entry")
	}

	return convertEntry(*entry), nil
}

func LogTime(o LogTimeOptions, ctx context.Context) (Entry, error) {
	client, err := createClient(ctx)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating client")
	}

	options := o.toHarvestOptions()

	entry, _, err := client.Timesheet.CreateTimeEntryViaDuration(ctx, &options)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating time entry")
	}

	return convertEntry(*entry), nil
}

func RestartTimeEntry(timeEntryId int64, ctx context.Context) (Entry, error) {
	client, err := createClient(ctx)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating client")
	}

	entry, _, err := client.Timesheet.RestartTimeEntry(ctx, timeEntryId)
	if err != nil {
		return Entry{}, errors.Wrap(err, "restarting time entry")
	}

	return convertEntry(*entry), nil
}

func SubmitWeekUrl(t time.Time, ctx context.Context) (string, error) {
	client, err := createClient(ctx)
	if err != nil {
		return "", errors.Wrap(err, "creating client")
	}

	comp, _, err := client.Company.Get(ctx)
	if err != nil {
		return "", errors.Wrap(err, "getting company")
	}

	return fmt.Sprintf("%s/time/week/%d/%d/%d", *comp.BaseURI, t.Year(), t.Month(), t.Day()), nil
}

func convertEntry(e harvest.TimeEntry) Entry {
	entry := Entry{
		ID:       *e.ID,
		Hours:    Hours(*e.Hours),
		Date:     (*e.SpentDate).String(),
		Running:  *e.IsRunning,
		Billable: *e.Billable,
		Project: Project{
			ID:   *e.Project.ID,
			Name: *e.Project.Name,
		},
		Task: Task{
			ID:        *e.Task.ID,
			ProjectId: *e.Project.ID,
			Name:      *e.Task.Name,
		},
	}
	if e.Hours != nil {
		entry.Hours = Hours(*e.Hours)
	}
	if e.Notes != nil {
		entry.Notes = *e.Notes
	}
	if e.StartedTime != nil {
		entry.TimerStarted = e.TimerStartedAt
	}
	return entry
}

func (o *EntryListOptions) toHarvestOptions() harvest.TimeEntryListOptions {
	var options harvest.TimeEntryListOptions
	options.PerPage = 100

	if o == nil {
		return options
	}

	if o.To != nil {
		options.To = &harvest.Date{Time: *o.To}
	}
	if o.From != nil {
		options.From = &harvest.Date{Time: *o.From}
	}

	options.ProjectID = o.ProjectId
	options.IsRunning = o.Running
	options.UserID = o.UserId

	return options
}

func (o LogTimeOptions) toHarvestOptions() harvest.TimeEntryCreateViaDuration {

	var notes *string
	if o.Notes != "" {
		notes = &o.Notes
	}

	hours := float64(o.Hours)

	return harvest.TimeEntryCreateViaDuration{
		ProjectID: &o.ProjectId,
		TaskID:    &o.TaskId,
		SpentDate: &harvest.Date{Time: o.Date},
		Hours:     &hours,
		Notes:     notes,
	}
}

func (o TimerStartOptions) toHarvestOptions() harvest.TimeEntryCreateViaStartEndTime {
	startTime := time.Now()
	if o.StartTime != nil {
		startTime = *o.StartTime
	}

	return harvest.TimeEntryCreateViaStartEndTime{
		ProjectID:   &o.ProjectId,
		TaskID:      &o.TaskId,
		SpentDate:   &harvest.Date{Time: startTime},
		StartedTime: &harvest.Time{Time: startTime},
		Notes:       o.Notes,
	}
}

func (o *EntryListOptions) includeTask(taskId *int64) bool {
	if o == nil || o.TaskId == nil {
		return true
	}
	if taskId == nil || *taskId == *o.TaskId {
		return true
	}
	return false
}

func (o EntryUpdateOptions) toHarvestOptions() harvest.TimeEntryUpdate {
	var h harvest.TimeEntryUpdate

	o.ProjectId = &o.Entry.Project.ID
	o.TaskId = &o.Entry.Task.ID
	o.Date, _ = util.StringToDate(o.Entry.Date)

	if o.ProjectId != nil {
		h.ProjectID = o.ProjectId
	}
	if o.TaskId != nil {
		h.TaskID = o.TaskId
	}
	if o.Date != nil {
		h.SpentDate = &harvest.Date{Time: *o.Date}
	}
	if o.Started != nil {
		h.StartedTime = &harvest.Time{Time: *o.Started}
	}

	h.Notes = o.Notes
	h.Hours = o.Hours.FloatPtr()

	return h
}
