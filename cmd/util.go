package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/timers"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"os"
	"os/signal"
	"strconv"
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
			fmt.Printf("An error occurred: %v\n", err)
		}

		signal.Stop(c)
		cancel()
	}
}

// createTable utility to create table with common properties across project
func createTable(columns []string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(150)
	if columns != nil {
		table.SetHeader(columns)
	}
	return table
}

func outputJson(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errors.Wrap(err, "problem marshalling data to json")
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
	viper.Set("tracking", timers.Records)

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
		return errors.Wrap(err, "problem saving config")
	}

	return nil
}

func formatHours(hours float64) string {
	if config.Cli.TimeDeltaFormat == config.TimeDeltaFormatHuman {
		minutes := 60 * (hours - math.Floor(hours))
		if hours < 1 {
			return fmt.Sprintf("%0.0fm", minutes)
		}
		return fmt.Sprintf("%0.0fh %0.0fm", math.Floor(hours), minutes)
	}

	// else config.TimeDeltaFormatDecimal or other
	return fmt.Sprintf("%0.2f", hours)
}

func getTaskProjectId(taskId int64, ctx context.Context) (*int64, error) {

	projects, err := harvest.GetProjects(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "problem getting projects for taskId")
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
		selected := prompt.ForSelection("Matched multiple projects, select one", tasksProjects)
		return &projects[selected].ID, nil
	}
}

func getTaskAndProjectId(str string) (taskId, projectId *int64, err error) {
	if str == "" {
		return nil, nil, nil
	}

	if taskAlias, ok := config.Harvest.TaskAliases[str]; ok {
		return &taskAlias.TaskId, &taskAlias.ProjectId, nil
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return nil, nil, errors.New("no alias found for " + str)
	}
	return &i, nil, err
}
