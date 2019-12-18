/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
	Long: ``,
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
