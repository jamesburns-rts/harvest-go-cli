package time

import (
	"context"
	"github.com/becoded/go-harvest/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"log"
	"strconv"
	"time"
)

type (
	Project struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Tasks []Task `json:"tasks"`
	}

	Task struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	Entry struct {
		Date    string  `json:"date"`
		Hours   float64 `json:"hours"`
		ID      int64   `json:"id"`
		Notes   string  `json:"notes"`
		Project Project `json:"project"`
		Task    Task    `json:"task"`
	}

	EntryListOptions struct {
		To        *time.Time
		From      *time.Time
		ProjectId *int64
		TaskId    *int64
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

	if taskId, ok := config.Harvest.ProjectAliases[str]; ok {
		return &taskId, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, errors.New("no alias found for " + str)
	}
	return &i, err
}

// GetTaskId either parse the string for an integer or check for an alias
func GetTaskId(str string) (*int64, error) {
	if str == "" {
		return nil, nil
	}

	if taskId, ok := config.Harvest.TaskAliases[str]; ok {
		return &taskId, nil
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
					ID:   *t.Task.Id,
					Name: *t.Task.Name,
				})
			}

			projects = append(projects, Project{
				ID:    *p.Project.Id,
				Name:  *p.Project.Name,
				Tasks: tasks,
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

			entry := Entry{
				ID:    *e.Id,
				Hours: *e.Hours,
				Date:  (*e.SpentDate).String(),
				Project: Project{
					ID:   *e.Project.Id,
					Name: *e.Project.Name,
				},
				Task: Task{
					ID:   *e.Task.Id,
					Name: *e.Task.Name,
				},
			}
			if e.Hours != nil {
				entry.Hours = *e.Hours
			}
			if e.Notes != nil {
				entry.Notes = *e.Notes
			}

			entries = append(entries, entry)
		}

		lastPage = *page.TotalPages - 1
	}

	return entries, nil
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
	if o.ProjectId != nil {
		options.ProjectId = o.ProjectId
	}

	return options
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
