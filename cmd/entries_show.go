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
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
)

var entriesShowCmd = &cobra.Command{
	Use:   "show [ENTRY_ID]",
	Args:  cobra.ExactArgs(1),
	Short: "Show a time entry",
	Long:  `Show a time entry`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		entryId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return errors.Wrap(err, "invalid entry ID")
		}

		entry, err := harvest.GetEntry(entryId, ctx)
		if err != nil {
			return errors.Wrap(err, "getting entry")
		}

		entries := []harvest.Entry{entry}

		// print
		return printWithFormat(outputMap{
			config.OutputFormatSimple: func() error { return entriesOutputSimple(entries) },
			config.OutputFormatTable:  func() error { return entriesOutputTable(entries) },
			config.OutputFormatJson:   func() error { return outputJson(entry) },
		})
	}),
}

func init() {
	entriesCmd.AddCommand(entriesShowCmd)
}
