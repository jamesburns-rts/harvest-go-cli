package main

import (
	"fmt"
	"github.com/jamesburns-rts/harvest-go-cli/cmd"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func main() {
	rootCmd := cmd.CreateCommand()
	rootCmd.Use = filepath.Base(os.Args[0])
	rootCmd.AddCommand(versionCommand)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var version = "unknown" // filled in by goreleaser

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Prints the version string",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
