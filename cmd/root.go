package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "data-automation-service",
	Short: "Data Retrieval and Automation Service",
	Long:  `A Go-based CLI tool to extract complex reports from PostgreSQL and export them to JSON, or push to an API.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
