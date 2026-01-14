package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	concurrency int
	timeout     int
	jsonOutput  bool
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "pinger",
	Short: "A concurrent service health checker",
	Long: `Pinger is a CLI tool for checking the health of services and endpoints.

It uses goroutines to perform concurrent health checks against HTTP endpoints,
TCP ports, and DNS records. Perfect for monitoring service dependencies,
validating deployments, and integrating into CI/CD pipelines.

Examples:
  pinger check -c config.yaml
  pinger check https://google.com https://github.com
  pinger check -c config.yaml --json`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "n", 10, "number of concurrent workers")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 5, "timeout in seconds")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")
}
