// certifytiming emits raw cold go-test events and a concise certification
// harness timing summary for CI and local Verify diagnostics.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"polymetrics.ai/internal/certifytiming"
)

func main() {
	maxDuration := flag.String("max-duration", "", "optional total wall-time duration budget, for example 3m30s")
	flag.Parse()
	if flag.NArg() != 0 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: certifytiming [--max-duration <duration>]")
		os.Exit(2)
	}

	allowed, err := parseDuration(*maxDuration)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	report, err := certifytiming.Run(context.Background(), certifytiming.DefaultTargets(), certifytiming.ExecuteGoTest, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := certifytiming.CheckDurationBudget(report, allowed); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseDuration(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("certify timing: invalid --max-duration %q: %w", raw, err)
	}
	if duration <= 0 {
		return 0, fmt.Errorf("certify timing: --max-duration must be positive")
	}
	return duration, nil
}
