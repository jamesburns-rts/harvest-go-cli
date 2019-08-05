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
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var entriesDeleteCmd = &cobra.Command{
	Use:   "delete ENTRY_ID",
	Args:  cobra.ExactArgs(1),
	Short: "Delete time entry",
	Long:  `Delete time entry`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var entryId int64

		if entryId, err = strconv.ParseInt(args[0], 10, 64); err != nil {
			return errors.Wrap(err, "problem with [entryId]")
		}

		// delete entry
		if err = harvest.DeleteEntry(entryId, ctx); err != nil {
			return errors.Wrap(err, "problem delete entry")
		}

		return nil
	}),
}

func init() {
	entriesCmd.AddCommand(entriesDeleteCmd)
}
