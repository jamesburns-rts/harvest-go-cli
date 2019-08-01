package harvest

import (
	"context"
	//"github.com/becoded/go-harvest/harvest"
	"github.com/jamesburns-rts/go-harvest/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"log"
	"strconv"
	"time"
)

type (
	Project struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Billable bool   `json:"billable"`
		Tasks    []Task `json:"tasks"`
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
	}

	EntryListOptions struct {
		To        *time.Time
		From      *time.Time
		ProjectId *int64
		TaskId    *int64
		Running   *bool
	}

	LogTimeOptions struct {
		TaskId    int64
		ProjectId int64
		Date      time.Time
		Hours     Hours
		Notes     string
	}
)

func createClient(ctx context.Context) (*harvest.HarvestClient, error) {

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
	client := harvest.NewHarvestClient(tc)
	client.AccountId = props.AccountId

	return client, nil
}

// GetProjectId either parse the string for an integer or check for an alias
func GetProjectId(str string) (*int64, error) {
	if str == "" {
		return nil, nil
	}

	if projectAlias, ok := config.Harvest.ProjectAliases[str]; ok {
		return &projectAlias.ProjectId, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, errors.New("no alias found for " + str)
	}
	return &i, err
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
					ID:        *t.Task.Id,
					ProjectId: *p.Project.Id,
					Name:      *t.Task.Name,
				})
			}

			projects = append(projects, Project{
				ID:       *p.Project.Id,
				Name:     *p.Project.Name,
				Billable: p.Project.IsBillable != nil && *p.Project.IsBillable,
				Tasks:    tasks,
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
			log.Fatal(err)
		}

		for _, e := range page.TimeEntries {

			if !o.includeTask(e.Task.Id) {
				continue
			}

			entries = append(entries, convertEntry(*e))
		}

		lastPage = *page.TotalPages - 1
	}

	return entries, nil
}

func DeleteEntry(entryId int64, ctx context.Context) error {
	client, err := createClient(ctx)
	if err != nil {
		return errors.Wrap(err, "creating client")
	}

	_, err = client.Timesheet.DeleteTimeEntry(ctx, entryId)
	return err
}

func GetTimers(o *EntryListOptions, ctx context.Context) (entries []Entry, err error) {

	if o == nil {
		o = &EntryListOptions{}
	}
	o.Running = BoolPtr(true)

	return GetEntries(o, ctx)
}

func LogTime(o LogTimeOptions, ctx context.Context) (Entry, error) {
	client, err := createClient(ctx)
	if err != nil {
		return Entry{}, errors.Wrap(err, "creating client")
	}

	options, err := o.toHarvestOptions(ctx)
	if err != nil {
		return Entry{}, err
	}

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

func convertEntry(e harvest.TimeEntry) Entry {
	entry := Entry{
		ID:      *e.Id,
		Hours:   Hours(*e.Hours),
		Date:    (*e.SpentDate).String(),
		Running: *e.IsRunning,
		Project: Project{
			ID:   *e.Project.Id,
			Name: *e.Project.Name,
		},
		Task: Task{
			ID:        *e.Task.Id,
			ProjectId: *e.Project.Id,
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
		entry.TimerStarted = &e.StartedTime.Time
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

	options.ProjectId = o.ProjectId
	options.IsRunning = o.Running

	return options
}

func (o LogTimeOptions) toHarvestOptions(ctx context.Context) (harvest.TimeEntryCreateViaDuration, error) {

	var notes *string
	if o.Notes != "" {
		notes = &o.Notes
	}

	hours := float64(o.Hours)

	return harvest.TimeEntryCreateViaDuration{
		ProjectId: &o.ProjectId,
		TaskId:    &o.TaskId,
		SpentDate: &harvest.Date{Time: o.Date},
		Hours:     &hours,
		Notes:     notes,
	}, nil
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
