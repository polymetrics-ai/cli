// Package certifytiming runs and summarizes the narrowly-scoped cold test
// commands that exercise the connector certification harness. It intentionally
// preserves Go's raw -json stream so CI logs remain independently inspectable.
package certifytiming

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const maxEventLineSize = 4 * 1024 * 1024

// Target describes one cold go-test invocation and the package event expected
// in its JSON stream.
type Target struct {
	Name       string
	GoPackage  string
	ImportPath string
	Run        string
}

// SlowTest is one completed test from a target's go test JSON stream.
type SlowTest struct {
	Name    string
	Elapsed time.Duration
}

// TargetResult is the parsed passing result for one target.
type TargetResult struct {
	Name        string
	Package     string
	Elapsed     time.Duration
	WallElapsed time.Duration
	SlowTests   []SlowTest
}

// Report is the aggregate of all targeted cold test invocations.
type Report struct {
	Targets []TargetResult
}

// TotalElapsed is the sum of Go's final package elapsed values. It reports
// package test cost and drives the per-package diagnostic summary.
func (r Report) TotalElapsed() time.Duration {
	var total time.Duration
	for _, target := range r.Targets {
		total += target.Elapsed
	}
	return total
}

// TotalWallElapsed is the sum of the wall time for each targeted cold go test
// process. It includes startup and compilation for those commands, making it
// the independent time backstop for the deterministic invocation contracts.
func (r Report) TotalWallElapsed() time.Duration {
	var total time.Duration
	for _, target := range r.Targets {
		total += target.WallElapsed
	}
	return total
}

// ExecuteTarget writes the raw go test JSON stream for target to raw. The
// injected form keeps parser/summary tests deterministic without executing Go.
type ExecuteTarget func(ctx context.Context, target Target, raw io.Writer) error

// DefaultTargets names the two narrow test commands covered by the cost
// contract. The CLI target uses its focused test prefix so unrelated CLI test
// work cannot distort certification-harness timing.
func DefaultTargets() []Target {
	return []Target{
		{
			Name:       "certify-harness",
			GoPackage:  "./internal/connectors/certify",
			ImportPath: "polymetrics.ai/internal/connectors/certify",
		},
		{
			Name:       "certify-cli",
			GoPackage:  "./internal/cli",
			ImportPath: "polymetrics.ai/internal/cli",
			Run:        "^TestCertifyCLI",
		},
	}
}

// GoTestArgs constructs a cold, JSON-emitting command without a shell.
func GoTestArgs(target Target) []string {
	args := []string{"test", "-count=1", "-json"}
	if target.Run != "" {
		args = append(args, "-run", target.Run)
	}
	return append(args, target.GoPackage)
}

// ExecuteGoTest runs target and streams its raw stdout events to raw. Stderr
// stays visible in CI, including compiler failures that cannot produce JSON.
func ExecuteGoTest(ctx context.Context, target Target, raw io.Writer) error {
	cmd := exec.CommandContext(ctx, "go", GoTestArgs(target)...)
	cmd.Stdout = raw
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type event struct {
	Action  string
	Package string
	Test    string
	Elapsed float64
}

// ParseTarget accepts only a complete passing package event stream. A malformed
// event, missing final package completion, or failed package is a hard error;
// this prevents timing output from accidentally looking healthy after a broken
// test command.
func ParseTarget(target Target, raw io.Reader) (TargetResult, error) {
	if target.Name == "" || target.ImportPath == "" {
		return TargetResult{}, fmt.Errorf("certify timing: target requires name and import path")
	}

	result := TargetResult{Name: target.Name, Package: target.ImportPath}
	scanner := bufio.NewScanner(raw)
	scanner.Buffer(make([]byte, 0, 64*1024), maxEventLineSize)

	var sawPackage bool
	var completed bool
	var passed bool
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var parsed event
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			return TargetResult{}, fmt.Errorf("certify timing: parse go test event line %d for %s: %w", line, target.Name, err)
		}
		if parsed.Package != target.ImportPath {
			continue
		}
		sawPackage = true

		if parsed.Test != "" && (parsed.Action == "pass" || parsed.Action == "fail") {
			result.SlowTests = append(result.SlowTests, SlowTest{
				Name:    parsed.Test,
				Elapsed: secondsToDuration(parsed.Elapsed),
			})
		}
		if parsed.Test == "" && (parsed.Action == "pass" || parsed.Action == "fail") {
			completed = true
			passed = parsed.Action == "pass"
			result.Elapsed = secondsToDuration(parsed.Elapsed)
		}
	}
	if err := scanner.Err(); err != nil {
		return TargetResult{}, fmt.Errorf("certify timing: read go test events for %s: %w", target.Name, err)
	}
	if !sawPackage {
		return TargetResult{}, fmt.Errorf("certify timing: missing events for target %s (%s)", target.Name, target.ImportPath)
	}
	if !completed {
		return TargetResult{}, fmt.Errorf("certify timing: missing package completion for target %s", target.Name)
	}
	if !passed {
		return TargetResult{}, fmt.Errorf("certify timing: target %s failed", target.Name)
	}

	sort.SliceStable(result.SlowTests, func(i, j int) bool {
		if result.SlowTests[i].Elapsed != result.SlowTests[j].Elapsed {
			return result.SlowTests[i].Elapsed > result.SlowTests[j].Elapsed
		}
		return result.SlowTests[i].Name < result.SlowTests[j].Name
	})
	return result, nil
}

// Run executes each target, mirrors every raw Go test event into output, and
// appends a compact deterministic summary after all targets pass.
func Run(ctx context.Context, targets []Target, execute ExecuteTarget, output io.Writer) (Report, error) {
	if len(targets) == 0 {
		return Report{}, fmt.Errorf("certify timing: no targets configured")
	}
	if execute == nil {
		return Report{}, fmt.Errorf("certify timing: target executor is required")
	}
	if output == nil {
		return Report{}, fmt.Errorf("certify timing: output writer is required")
	}

	report := Report{Targets: make([]TargetResult, 0, len(targets))}
	for _, target := range targets {
		var raw bytes.Buffer
		started := time.Now()
		execErr := execute(ctx, target, io.MultiWriter(&raw, output))
		wallElapsed := time.Since(started)
		result, parseErr := ParseTarget(target, &raw)
		if parseErr != nil {
			return report, parseErr
		}
		if execErr != nil {
			return report, fmt.Errorf("certify timing: run target %s: %w", target.Name, execErr)
		}
		result.WallElapsed = wallElapsed
		report.Targets = append(report.Targets, result)
	}
	RenderSummary(output, report)
	return report, nil
}

// RenderSummary prints target totals and their five slowest completed tests.
func RenderSummary(output io.Writer, report Report) {
	_, _ = fmt.Fprintln(output, "certify timing summary")
	for _, target := range report.Targets {
		_, _ = fmt.Fprintf(output, "  %s package=%s elapsed=%s wall_elapsed=%s\n", target.Name, target.Package, formatDuration(target.Elapsed), formatDuration(target.WallElapsed))
		limit := len(target.SlowTests)
		if limit > 5 {
			limit = 5
		}
		for _, test := range target.SlowTests[:limit] {
			_, _ = fmt.Fprintf(output, "    slow_test=%s elapsed=%s\n", test.Name, formatDuration(test.Elapsed))
		}
	}
	_, _ = fmt.Fprintf(output, "  total elapsed=%s wall_elapsed=%s\n", formatDuration(report.TotalElapsed()), formatDuration(report.TotalWallElapsed()))
}

// CheckDurationBudget enforces the secondary measured duration backstop. A
// zero duration leaves enforcement disabled until the CI measurements exist.
func CheckDurationBudget(report Report, allowed time.Duration) error {
	if allowed == 0 {
		return nil
	}
	if allowed < 0 {
		return fmt.Errorf("certify timing: allowed duration must be positive")
	}
	for _, target := range report.Targets {
		if target.WallElapsed <= 0 {
			return fmt.Errorf("certify timing: target %s has no measured wall duration", target.Name)
		}
	}
	observed := report.TotalWallElapsed()
	if observed > allowed {
		return fmt.Errorf("certify timing duration budget exceeded: observed %s, allowed %s", formatDuration(observed), formatDuration(allowed))
	}
	return nil
}

func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fs", d.Seconds())
}
