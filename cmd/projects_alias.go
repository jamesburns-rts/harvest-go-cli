package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/prompt"
	"github.com/spf13/cobra"
	"strconv"
)

var projectsAliasCmd = &cobra.Command{
	Use:   "alias [ALIAS [PROJECT_ID]]",
	Args:  cobra.MaximumNArgs(2),
	Short: "Alias a project ID",
	Long:  `Alias a project ID to a friendly string the can be used anywhere`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var projectId int64

		// get alias
		var alias string
		if len(args) > 0 {
			alias = args[0]
		} else {
			if alias, err = prompt.ForString("Alias Name", validAlias); err != nil {
				return err
			}
		}

		// get project
		if len(args) > 1 {
			if projectId, err = strconv.ParseInt(args[1], 10, 64); err != nil {
				return fmt.Errorf("for [PROJECT_ID]: %w", err)
			}
		} else {
			if p, err := selectProject(ctx); err != nil {
				return err
			} else {
				projectId = *p
			}

		}

		// set alias
		config.SetProjectAlias(config.ProjectAlias{
			Name:      alias,
			ProjectId: projectId,
		})

		return writeConfig()
	}),
}

var timeProjectsAliasDeleteCmd = &cobra.Command{
	Use:   "delete [ALIAS]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Delete a project ID alias",
	Long:  ``,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var alias string
		if len(args) > 0 {
			alias = args[0]
		} else {
			if alias, err = selectProjectAlias(); err != nil {
				return err
			}
		}
		config.DeleteProjectAlias(alias)

		return writeConfig()
	}),
}

func init() {
	projectsCmd.AddCommand(projectsAliasCmd)
	projectsAliasCmd.AddCommand(timeProjectsAliasDeleteCmd)
}
