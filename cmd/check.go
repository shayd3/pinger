package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shayd3/pinger/internal/checker"
	"github.com/shayd3/pinger/internal/config"
	"github.com/shayd3/pinger/internal/worker"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [urls...]",
	Short: "Check health of endpoints",
	Long: `Check the health status of one or more endpoints.

Examples:
  pinger check https://google.com
  pinger check https://google.com https://github.com
  pinger check -c config.yaml
  pinger check https://google.com --json`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	var targets []config.Target

	if cfgFile != "" {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
		targets = cfg.Targets
		if concurrency == 10 && cfg.Concurrency > 0 {
			concurrency = cfg.Concurrency
		}
	} else if len(args) > 0 {
		for _, url := range args {
			targets = append(targets, config.Target{
				Name:           url,
				URL:            url,
				Type:           config.CheckTypeHTTP,
				Timeout:        time.Duration(timeout) * time.Second,
				ExpectedStatus: 200,
			})
		}
	} else {
		return fmt.Errorf("no targets specified. Provide URLs or use -c config.yaml")
	}

	ctx := context.Background()
	results := worker.Run(ctx, targets, concurrency)

	if jsonOutput {
		return outputJSON(results)
	}
	return outputTable(results)
}

func outputJSON(results []checker.Result) error {
	output := struct {
		Timestamp time.Time        `json:"timestamp"`
		Results   []checker.Result `json:"results"`
		Summary   struct {
			Total     int `json:"total"`
			Healthy   int `json:"healthy"`
			Unhealthy int `json:"unhealthy"`
		} `json:"summary"`
	}{
		Timestamp: time.Now(),
		Results:   results,
	}

	output.Summary.Total = len(results)
	for _, r := range results {
		if r.IsHealthy() {
			output.Summary.Healthy++
		} else {
			output.Summary.Unhealthy++
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputTable(results []checker.Result) error {
	green := "\033[32m"
	red := "\033[31m"
	yellow := "\033[33m"
	reset := "\033[0m"
	bold := "\033[1m"

	fmt.Printf("\n%s%-40s %-12s %-10s %-12s %s%s\n",
		bold, "TARGET", "STATUS", "CODE", "LATENCY", "ERROR", reset)
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────")

	healthy := 0
	unhealthy := 0

	for _, r := range results {
		statusColor := green
		statusIcon := "✓"

		switch r.Status {
		case checker.StatusHealthy:
			healthy++
		case checker.StatusUnhealthy:
			statusColor = red
			statusIcon = "✗"
			unhealthy++
		case checker.StatusTimeout:
			statusColor = yellow
			statusIcon = "⏱"
			unhealthy++
		case checker.StatusError:
			statusColor = red
			statusIcon = "!"
			unhealthy++
		}

		name := r.Name
		if len(name) > 38 {
			name = name[:35] + "..."
		}

		codeStr := "-"
		if r.StatusCode > 0 {
			codeStr = fmt.Sprintf("%d", r.StatusCode)
		}

		latencyStr := fmt.Sprintf("%dms", r.LatencyDuration().Milliseconds())

		errStr := r.Error
		if len(errStr) > 30 {
			errStr = errStr[:27] + "..."
		}

		fmt.Printf("%-40s %s%s %-10s%s %-10s %-12s %s\n",
			name, statusColor, statusIcon, string(r.Status), reset,
			codeStr, latencyStr, errStr)
	}

	fmt.Println()
	fmt.Printf("%sSummary:%s %d total, %s%d healthy%s, %s%d unhealthy%s\n",
		bold, reset, len(results), green, healthy, reset, red, unhealthy, reset)

	if unhealthy > 0 {
		os.Exit(1)
	}
	return nil
}
