package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/harvest"
	"github.com/spf13/cobra"
	"strconv"
)

var entriesDeleteCmd = &cobra.Command{
	Use:   "delete ENTRY_ID",
	Args:  cobra.ExactArgs(1),
	Short: "Delete time entry",
	Long:  `Delete time entry where ENTRY_ID is a positive integer`,
	Run: withCtx(func(cmd *cobra.Command, args []string, ctx context.Context) (err error) {

		var entryId int64

		if entryId, err = strconv.ParseInt(args[0], 10, 64); err != nil {
			return fmt.Errorf("problem with ENTRY_ID: %w", err)
		}

		// delete entry
		if err = harvest.DeleteEntry(entryId, ctx); err != nil {
			return fmt.Errorf("problem deleting entry: %w", err)
		}

		return nil
	}),
}

func init() {
	entriesCmd.AddCommand(entriesDeleteCmd)
}
