package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/spf13/cobra"
	"strconv"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects",
	Long:  `List projects and their associated IDs`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		// get projects
		projects, err := harvest.GetProjects(ctx)
		if err != nil {
			return err
		}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return projectsOutputSimple(projects) },
			config.OutputFormatTable:  func() error { return projectsOutputTable(projects) },
			config.OutputFormatJson:   func() error { return outputJson(projects) },
		})
	}),
}

func projectsOutputSimple(projects []harvest.Project) error {
	for _, p := range projects {
		fmt.Printf("%v %v\n", p.ID, p.Name)
	}
	return nil
}

func projectsOutputTable(projects []harvest.Project) error {
	table := createTable([]string{"ID", "Project Name"})
	for _, proj := range projects {
		table.Append([]string{
			strconv.Itoa(int(proj.ID)),
			proj.Name,
		})
	}
	table.Render()
	return nil
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}
