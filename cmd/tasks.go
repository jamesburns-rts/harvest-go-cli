package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksProject projectArg
var tasksAll bool

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "List tasks of a project",
	Long:  `For the given project, list the tasks with their associated IDs`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		projectId := tasksProject.projectId

		if projectId == nil && !tasksAll {
			if projectId, err = selectProject(ctx); err != nil {
				return err
			}
		}

		// get tasks
		var tasks []harvest.Task
		if tasks, err = harvest.GetTasks(projectId, ctx); err != nil {
			return fmt.Errorf("getting tasks: %w", err)
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return tasksOutputSimple(tasks) },
			config.OutputFormatTable:  func() error { return tasksOutputTable(tasks) },
			config.OutputFormatJson:   func() error { return outputJson(tasks) },
		})
	}),
}

func tasksOutputSimple(tasks []harvest.Task) error {
	for _, task := range tasks {
		fmt.Printf("%v %v %v\n", task.ProjectId, task.ID, task.Name)
	}
	return nil
}

func tasksOutputTable(tasks []harvest.Task) error {
	table := createTable([]string{"Project ID", "ID", "Task Name"})
	for _, task := range tasks {
		table.Append([]string{
			strconv.Itoa(int(task.ProjectId)),
			strconv.Itoa(int(task.ID)),
			task.Name,
		})
	}
	table.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.Flags().VarP(&tasksProject, "project", "p", "Project ID or alias")
	tasksCmd.Flags().BoolVarP(&tasksAll, "all", "A", false, "Show tasks from all projects")
}
