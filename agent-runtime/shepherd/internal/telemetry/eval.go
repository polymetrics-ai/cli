package telemetry

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type EvalSummary struct {
	Runs                 int            `json:"runs"`
	ToolCalls            int            `json:"tool_calls"`
	Heartbeats           int            `json:"heartbeats"`
	MaxHeartbeatGapMS    int64          `json:"max_heartbeat_gap_ms"`
	HeartbeatSLOBreaches int            `json:"heartbeat_slo_breaches"`
	InputTokens          int64          `json:"input_tokens"`
	OutputTokens         int64          `json:"output_tokens"`
	CostMicros           int64          `json:"cost_micros"`
	TerminalCounts       map[string]int `json:"terminal_counts"`
	ValidationCounts     map[string]int `json:"validation_counts"`
}

type Sink interface {
	Export(context.Context, []Activity) error
}

func ReadDirectory(directory string) ([]Activity, error) {
	paths, err := filepath.Glob(filepath.Join(directory, "activity-*.jsonl"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	var activities []Activity
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 4096), 256*1024)
		for scanner.Scan() {
			var activity Activity
			if err := json.Unmarshal(scanner.Bytes(), &activity); err != nil {
				_ = file.Close()
				return nil, fmt.Errorf("decode activity segment: %w", err)
			}
			if err := validate(activity); err != nil {
				_ = file.Close()
				return nil, err
			}
			activities = append(activities, activity)
		}
		if err := scanner.Err(); err != nil {
			_ = file.Close()
			return nil, err
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	}
	return activities, nil
}

func Evaluate(activities []Activity) EvalSummary {
	summary := EvalSummary{TerminalCounts: make(map[string]int), ValidationCounts: make(map[string]int)}
	runs := make(map[string]struct{})
	lastHeartbeat := make(map[string]time.Time)
	starts := make(map[string]time.Time)
	for _, activity := range activities {
		runs[activity.RunID] = struct{}{}
		summary.InputTokens += activity.InputTokens
		summary.OutputTokens += activity.OutputTokens
		summary.CostMicros += activity.CostMicros
		switch activity.Kind {
		case "run.started":
			starts[activity.RunID] = activity.At
		case "tool.started":
			summary.ToolCalls++
		case "heartbeat":
			summary.Heartbeats++
			previous := lastHeartbeat[activity.RunID]
			if previous.IsZero() {
				previous = starts[activity.RunID]
			}
			if !previous.IsZero() {
				gap := activity.At.Sub(previous).Milliseconds()
				if gap > summary.MaxHeartbeatGapMS {
					summary.MaxHeartbeatGapMS = gap
				}
				if gap > 15_000 {
					summary.HeartbeatSLOBreaches++
				}
			}
			lastHeartbeat[activity.RunID] = activity.At
		case "run.terminal":
			summary.TerminalCounts[activity.Status]++
			previous := lastHeartbeat[activity.RunID]
			if previous.IsZero() {
				previous = starts[activity.RunID]
			}
			if !previous.IsZero() {
				gap := activity.At.Sub(previous).Milliseconds()
				if gap > summary.MaxHeartbeatGapMS {
					summary.MaxHeartbeatGapMS = gap
				}
				if gap > 15_000 {
					summary.HeartbeatSLOBreaches++
				}
			}
		case "validation":
			summary.ValidationCounts[activity.Status]++
		}
	}
	summary.Runs = len(runs)
	return summary
}

func Export(ctx context.Context, sink Sink, directory string) error {
	activities, err := ReadDirectory(directory)
	if err != nil {
		return err
	}
	return sink.Export(ctx, activities)
}
