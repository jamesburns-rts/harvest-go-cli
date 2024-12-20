package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/spf13/cobra"
	"strconv"
)

var projectsAliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "List project aliases",
	Long:  `List project aliases`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		// get aliases
		aliases := config.Harvest.Projects

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return projectsAliasesOutputSimple(aliases) },
			config.OutputFormatTable:  func() error { return projectsAliasesOutputTable(aliases) },
			config.OutputFormatJson:   func() error { return outputJson(aliases) },
		})
	}),
}

func projectsAliasesOutputSimple(aliases []config.ProjectAlias) error {
	for _, p := range aliases {
		fmt.Println(p.Name)
	}
	return nil
}

func projectsAliasesOutputTable(aliases []config.ProjectAlias) error {
	table := createTable([]string{"Alias", "ProjectId"})
	for _, v := range aliases {
		table.Append([]string{v.Name, strconv.Itoa(int(v.ProjectId))})
	}
	table.Render()
	return nil
}

func init() {
	projectsCmd.AddCommand(projectsAliasesCmd)
}
