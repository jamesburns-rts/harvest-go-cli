package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/jamesburns-rts/harvest-go-cli/internal/util"
	"github.com/spf13/cobra"
)

var submitWeek dateArg

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a timesheet week",
	Long:  ``,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {
		url, err := harvest.SubmitWeekUrl(time.Now(), ctx)
		if err != nil {
			return err
		}
		fmt.Println(url)
		return util.OpenURL(url)
	}),
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().Var(&submitWeek, "week", "week [see date section in root]")
}
