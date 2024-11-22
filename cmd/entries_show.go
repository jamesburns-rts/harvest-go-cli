package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
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
			return fmt.Errorf("invalid entry ID: %w", err)
		}

		entry, err := harvest.GetEntry(entryId, ctx)
		if err != nil {
			return fmt.Errorf("getting entry: %w", err)
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
