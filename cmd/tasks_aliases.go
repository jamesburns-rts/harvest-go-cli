package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/spf13/cobra"
	"strconv"
)

var tasksAliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "List task aliases",
	Long:  `List task aliases`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) error {

		aliases := config.Harvest.Tasks

		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return tasksAliasesOutputSimple(aliases) },
			config.OutputFormatTable:  func() error { return tasksAliasesOutputTable(aliases) },
			config.OutputFormatJson:   func() error { return outputJson(aliases) },
		})
	}),
}

func tasksAliasesOutputSimple(aliases []config.TaskAlias) error {
	for _, t := range aliases {
		fmt.Println(t.Name)
	}
	return nil
}

func tasksAliasesOutputTable(aliases []config.TaskAlias) error {
	table := createTable([]string{"Alias", "ProjectId", "TaskId"})
	for _, v := range aliases {
		table.Append([]string{
			v.Name,
			strconv.Itoa(int(v.ProjectId)),
			strconv.Itoa(int(v.TaskId)),
		})
	}
	table.Render()
	return nil
}

func init() {
	tasksCmd.AddCommand(tasksAliasesCmd)
}
