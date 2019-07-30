package cmd

import (
	"context"
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/internal/config"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"math"
	"os"
	"os/signal"
)

type CobraFunc func(cmd *cobra.Command, args []string)
type CobraFuncWithCtx func(cmd *cobra.Command, args []string, ctx context.Context) error
type TimeFunc func(cmd *cobra.Command, args []string, ctx context.Context)

func withCtx(f CobraFuncWithCtx) CobraFunc {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	return func(cmd *cobra.Command, args []string) {
		err := f(cmd, args, ctx)
		if err != nil {
			fmt.Printf("An error occurred: %v\n", err)
		}

		signal.Stop(c)
		cancel()
	}
}

// createTable utility to create table with common properties across project
func createTable(columns []string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetColWidth(150)
	if columns != nil {
		table.SetHeader(columns)
	}
	return table
}

const (
	outputFormatOptions = "[" +
		config.OutputFormatJson + ", " +
		config.OutputFormatSimple + ", " +
		config.OutputFormatTable +
		"]"
)

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func writeConfig() error {
	viper.Set("harvest", config.Harvest)
	viper.Set("cli", config.Cli)
	viper.Set("timers", config.Timers)

	if !fileExists(cfgFile) {
		f, err := os.Create(cfgFile)
		if err != nil {
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	}

	if err := viper.WriteConfig(); err != nil {
		return errors.Wrap(err, "problem saving config")
	}

	return nil
}

func formatHours(hours float64) string {
	if config.Cli.TimeDeltaFormat == config.TimeDeltaFormatHuman {
		minutes := 60 * (hours - math.Floor(hours))
		if hours < 1 {
			return fmt.Sprintf("%0.0fm", minutes)
		}
		return fmt.Sprintf("%0.0fh %0.0fm", math.Floor(hours), minutes)
	}

	// else config.TimeDeltaFormatDecimal or other
	return fmt.Sprintf("%0.2f", hours)
}
