package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	. "github.com/jamesburns-rts/harvest-go-cli/internal/types"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type CobraFunc func(cmd *cobra.Command, args []string)
type CobraFuncWithCtx func(cmd *cobra.Command, args []string, ctx context.Context) error
type TimeFunc func(cmd *cobra.Command, args []string, ctx context.Context)

func withCtx(f CobraFuncWithCtx) CobraFunc {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	return func(cmd *cobra.Command, args []string) {
		err := f(cmd, args, ctx)
		if err != nil {
			if strings.Contains(err.Error(), util.QuitError.Error()) {
				fmt.Println("Exited.")
			} else {
				fmt.Printf("An error occurred: %v\n", err)
			}
		}

		signal.Stop(c)
		cancel()
	}
}

// createTable utility to create table with common properties across project
func createTable(columns []string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.Configure(func(cfg *tablewriter.Config) {
		cfg.Widths.Global = 150
	})
	if columns != nil {
		table.Header(columns)
	}
	return table
}

func outputJson(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("problem marshalling data to json: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func writeConfig() error {
	viper.Set("harvest", config.Harvest)
	viper.Set("cli", config.Cli)
	viper.Set("timers", timers.Records)

	if !fileExists(cfgFile) {
		f, err := os.Create(cfgFile)
		if err != nil {
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	}

	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("problem saving config: %w", err)
	}

	return nil
}

func getAndSaveUserId(ctx context.Context) (*int64, error) {
	if config.Harvest.UserId != nil {
		return config.Harvest.UserId, nil
	}

	id, err := harvest.GetCurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	config.Harvest.UserId = id
	err = writeConfig()
	return id, err
}

func getTaskProjectId(taskId int64, ctx context.Context) (*int64, error) {

	projects, err := harvest.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("problem getting projects for taskId: %w", err)
	}
	var tasksProjects []harvest.Project
	for _, p := range projects {
		for _, t := range p.Tasks {
			if t.ID == taskId {
				tasksProjects = append(tasksProjects, p)
			}
		}
	}
	if len(tasksProjects) == 0 {
		return nil, errors.New("no project found for task id")
	} else if len(tasksProjects) == 1 {
		return &tasksProjects[0].ID, nil
	} else {
		selected, err := prompt.ForSelection("Matched multiple projects, select one", tasksProjects)
		return &projects[selected].ID, err
	}
}

func selectProjectAndTaskFrom(project, task string, ctx context.Context) (projectId, taskId *int64, err error) {

	if task != "" {
		if taskId, projectId, err = harvest.ParseTaskId(task); err != nil {
			return nil, nil, err
		}
		if projectId != nil {
			return taskId, projectId, nil
		}
	}

	if project != "" {
		if projectId, err = harvest.ParseProjectId(project); err != nil {
			return nil, nil, err
		}
	}

	if projectId == nil {
		if projectId, err = selectProject(ctx); err != nil {
			return nil, nil, err
		}
	}
	if taskId == nil {
		if taskId, err = selectTask(*projectId, ctx); err != nil {
			return nil, nil, err
		}
	}
	return projectId, taskId, nil
}

func selectProjectAndTask(ctx context.Context) (projectId, taskId *int64, err error) {
	projects, err := harvest.GetProjects(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("problem getting projects: %w", err)
	}
	selected, err := prompt.ForSelection("Select Project", projects)
	if err != nil {
		return nil, nil, err
	}
	project := projects[selected]
	selected, err = prompt.ForSelection("Select Task", project.Tasks)
	if err != nil {
		return nil, nil, err
	}
	return &project.ID, &project.Tasks[selected].ID, err
}

func selectProject(ctx context.Context) (projectId *int64, err error) {
	projects, err := harvest.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("problem getting projects: %w", err)
	}
	selected, err := prompt.ForSelection("Select Project", projects)
	if err != nil {
		return nil, err
	}
	return &projects[selected].ID, nil
}

func selectTask(projectId int64, ctx context.Context) (taskId *int64, err error) {
	project, err := harvest.GetProject(projectId, ctx)
	if err != nil {
		return nil, err
	}
	selected, err := prompt.ForSelection("Select Task", project.Tasks)
	if err != nil {
		return nil, err
	}
	return &project.Tasks[selected].ID, err
}

func selectProjectAlias() (alias string, err error) {
	var aliases []string
	for _, p := range config.Harvest.Projects {
		aliases = append(aliases, p.Name)
	}
	if selection, err := prompt.ForSelection("Alias", aliases); err != nil {
		return "", err
	} else {
		return aliases[selection], nil
	}
}

func selectTaskAlias() (alias string, err error) {
	var aliases []string
	for _, t := range config.Harvest.Tasks {
		aliases = append(aliases, t.Name)
	}
	if selection, err := prompt.ForSelection("Alias", aliases); err != nil {
		return "", err
	} else {
		return aliases[selection], nil
	}
}

func fmtHours(h *Hours) string {
	if h == nil {
		return "n/a"
	}
	if config.Cli.TimeDeltaFormat == config.TimeDeltaFormatHuman {
		if *h < 1 && *h > -1 {
			return fmt.Sprintf("%0.0fm", h.Minutes())
		} else if h.Minutes() == 0 {
			return fmt.Sprintf("%0.0fh", h.Hours())
		} else {
			return fmt.Sprintf("%0.0fh %0.0fm", h.Hours(), h.Minutes())
		}
	}

	// else config.TimeDeltaFormatDecimal or other
	str := fmt.Sprintf("%0.2f", float64(*h))
	str = strings.TrimRight(str, "0")
	return strings.TrimRight(str, ".")
}

func validProjectId(str string) error {
	p, err := harvest.ParseProjectId(str)
	if p == nil {
		return errors.New("not valid")
	}
	return err
}

func validTaskId(str string) error {
	t, _, err := harvest.ParseTaskId(str)
	if t == nil {
		return errors.New("not valid")
	}
	return err
}

func validHours(str string) error {
	h, err := ParseHours(str)
	if h == nil {
		return errors.New("not valid")
	}
	return err
}

func validDate(str string) error {
	d, err := util.StringToDate(str)
	if d == nil {
		return errors.New("not valid")
	}
	return err
}

func validNotes(str string) error {
	return nil
}

func validAlias(str string) error {
	if str == "" || strings.ContainsAny(str, " \t\n") {
		return errors.New("Must input word with no spaces")
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}

func ifOr[T any](t bool, a T, b T) T {
	if t {
		return a
	}
	return b
}

type timerCompletionOptions struct {
	pos               int
	excludeRunning    bool
	excludeNotRunning bool
}

func timerCompletionFunc(opt timerCompletionOptions) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > opt.pos {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var timerNames []string
		for _, t := range timers.Records.Timers {
			if opt.excludeRunning && t.Running {
				continue
			}
			if opt.excludeNotRunning && !t.Running {
				continue
			}
			timerNames = append(timerNames, t.Name)
		}
		return timerNames, cobra.ShellCompDirectiveNoFileComp
	}
}

func projectCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var aliases []string
	for _, p := range config.Harvest.Projects {
		aliases = append(aliases, p.Name)
	}
	return aliases, cobra.ShellCompDirectiveNoFileComp
}

func taskCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var aliases []string
	for _, p := range config.Harvest.Tasks {
		aliases = append(aliases, p.Name)
	}
	return aliases, cobra.ShellCompDirectiveNoFileComp
}
