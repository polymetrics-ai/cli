package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/telemetry"
)

const version = "0.1.0"

type fileConfig struct {
	GSDCommand       []string `json:"gsd_command"`
	WorkDir          string   `json:"work_dir"`
	GSDHome          string   `json:"gsd_home"`
	StateDir         string   `json:"state_dir"`
	CoordinatorModel string   `json:"coordinator_model"`
	TimeoutSeconds   int      `json:"timeout_seconds"`
	HeartbeatSeconds int      `json:"heartbeat_seconds"`
	MaxEventBytes    int      `json:"max_event_bytes"`
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "shepherd:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: shepherd <version|query|run|start> [flags]")
	}
	if args[0] == "version" {
		fmt.Println(version)
		return nil
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	configPath := flags.String("config", "", "path to governed Shepherd JSON config")
	command := flags.String("command", "next", "supported GSD headless command")
	issue := flags.Int("issue", 0, "GitHub issue number")
	contextPath := flags.String("context", "", "validated milestone context file")
	auto := flags.Bool("auto", false, "continue into GSD auto-mode after milestone intake")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	config, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	if err := gsd.ValidateRuntimeSettings(config.GSDHome, config.CoordinatorModel, "high"); err != nil {
		return fmt.Errorf("runtime admission: %w", err)
	}
	runner, err := gsd.NewRunner(gsd.Config{
		Command: config.GSDCommand, WorkDir: config.WorkDir, GSDHome: config.GSDHome,
		Model: config.CoordinatorModel, Thinking: "high",
		Timeout:           time.Duration(config.TimeoutSeconds) * time.Second,
		HeartbeatInterval: time.Duration(config.HeartbeatSeconds) * time.Second,
		MaxEventBytes:     config.MaxEventBytes,
	})
	if err != nil {
		return err
	}
	switch args[0] {
	case "query":
		snapshot, err := runner.Query(ctx)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(snapshot)
	case "run":
		return runHeadless(ctx, runner, config, *command, nil)
	case "start":
		if *issue <= 0 {
			return errors.New("--issue is required")
		}
		path, err := governedPath(config.WorkDir, *contextPath)
		if err != nil {
			return fmt.Errorf("context: %w", err)
		}
		milestoneArgs := []string{"--context", path}
		if *auto {
			milestoneArgs = append(milestoneArgs, "--auto")
		}
		return runHeadless(ctx, runner, config, "new-milestone", milestoneArgs)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func loadConfig(path string) (fileConfig, error) {
	if path == "" {
		return fileConfig{}, errors.New("--config is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, err
	}
	var config fileConfig
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return fileConfig{}, fmt.Errorf("decode config: %w", err)
	}
	if !filepath.IsAbs(config.WorkDir) || !filepath.IsAbs(config.GSDHome) || !filepath.IsAbs(config.StateDir) {
		return fileConfig{}, errors.New("work_dir, gsd_home, and state_dir must be absolute")
	}
	if config.CoordinatorModel == "" {
		config.CoordinatorModel = "openai-codex/gpt-5.6-sol"
	}
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 3600
	}
	if config.HeartbeatSeconds <= 0 {
		config.HeartbeatSeconds = 15
	}
	if config.MaxEventBytes <= 0 {
		config.MaxEventBytes = 256 * 1024
	}
	return config, nil
}

func runHeadless(ctx context.Context, runner *gsd.Runner, config fileConfig, command string, args []string) error {
	if err := os.MkdirAll(config.StateDir, 0o700); err != nil {
		return err
	}
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	runID := fmt.Sprintf("run-%d", time.Now().UTC().UnixNano())
	lease, err := authority.AcquireLease(ctx, runID, "shepherd", time.Now().UTC(), time.Duration(config.TimeoutSeconds+60)*time.Second)
	if err != nil {
		return err
	}
	_ = lease
	activity, err := telemetry.Open(ctx, filepath.Join(config.StateDir, "activity", "segments"))
	if err != nil {
		return err
	}
	defer activity.Close()
	var sequence atomic.Uint64
	appendActivity := func(kind, status, unit, model, tool string, at time.Time) {
		seq := sequence.Add(1)
		idInput := runID + ":" + strconv.FormatUint(seq, 10) + ":" + kind
		hash := sha256.Sum256([]byte(idInput))
		_, _ = activity.Append(ctx, telemetry.Activity{
			ID: hex.EncodeToString(hash[:]), RunID: runID, UnitID: unit, Kind: kind,
			Status: status, Model: model, Tool: tool, At: at,
		})
	}
	observer := gsd.Observer{
		Event: func(event gsd.Event) {
			kind := mapEventKind(event.Kind)
			appendActivity(kind, event.Status, event.UnitID, event.Model, event.Tool, event.At)
			fmt.Fprintf(os.Stderr, "%s kind=%s unit=%s tool=%s model=%s\n", event.At.Format(time.RFC3339), event.Kind, event.UnitID, event.Tool, event.Model)
		},
		Heartbeat: func(heartbeat gsd.Heartbeat) {
			appendActivity("heartbeat", "alive", "", "", heartbeat.InFlightTool, heartbeat.At)
			fmt.Fprintf(os.Stderr, "%s heartbeat alive=%t in_flight_tool=%s\n", heartbeat.At.Format(time.RFC3339), heartbeat.ProcessAlive, heartbeat.InFlightTool)
		},
		Question: terminalQuestion,
	}
	result := runner.Run(ctx, command, args, observer)
	if result.Terminal == gsd.TerminalSuccess || result.Terminal == gsd.TerminalBlocked {
		snapshot, queryErr := runner.Query(ctx)
		if queryErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = fmt.Errorf("post-run query reconciliation failed: %w", queryErr)
		} else {
			terminal, reconcileErr := gsd.Reconcile(command, result, snapshot)
			result.Terminal = terminal
			if reconcileErr != nil {
				result.Err = reconcileErr
			}
		}
	}
	appendActivity("run.terminal", string(result.Terminal), "", "", "", result.Ended)
	encoded, _ := json.Marshal(struct {
		RunID    string       `json:"run_id"`
		Terminal gsd.Terminal `json:"terminal"`
		ExitCode int          `json:"exit_code"`
	}{RunID: runID, Terminal: result.Terminal, ExitCode: result.ExitCode})
	fmt.Println(string(encoded))
	if result.Terminal != gsd.TerminalSuccess && result.Terminal != gsd.TerminalBlocked {
		return fmt.Errorf("GSD terminal=%s exit=%d: %w", result.Terminal, result.ExitCode, result.Err)
	}
	return nil
}

func terminalQuestion(_ context.Context, question gsd.Question) (gsd.UIResponse, error) {
	fmt.Fprintf(os.Stderr, "\nHUMAN GATE [%s] %s\n", question.Method, question.Title)
	for i, option := range question.Options {
		fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, option)
	}
	fmt.Fprint(os.Stderr, "Response (number, yes/no, text, or cancel): ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return gsd.UIResponse{Cancelled: true}, errors.New("explicit human response unavailable")
	}
	answer := strings.TrimSpace(line)
	if answer == "" || strings.EqualFold(answer, "cancel") {
		return gsd.UIResponse{Cancelled: true}, nil
	}
	if question.Method == "confirm" {
		confirmed := strings.EqualFold(answer, "yes") || strings.EqualFold(answer, "y")
		return gsd.UIResponse{Confirmed: &confirmed}, nil
	}
	if question.Method == "select" {
		choice, err := strconv.Atoi(answer)
		if err != nil || choice < 1 || choice > len(question.Options) {
			return gsd.UIResponse{Cancelled: true}, errors.New("selection must be a listed number")
		}
		return gsd.UIResponse{Value: question.Options[choice-1]}, nil
	}
	return gsd.UIResponse{Value: answer}, nil
}

func governedPath(root, path string) (string, error) {
	if path == "" {
		return "", errors.New("path is required")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	clean := filepath.Clean(path)
	relative, err := filepath.Rel(root, clean)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes work directory")
	}
	info, err := os.Stat(clean)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("path must be an existing regular file")
	}
	return clean, nil
}

func mapEventKind(kind gsd.EventKind) string {
	switch kind {
	case gsd.EventAgentStart:
		return "run.started"
	case gsd.EventTurnStart:
		return "unit.started"
	case gsd.EventToolStart:
		return "tool.started"
	case gsd.EventToolEnd:
		return "tool.terminal"
	case gsd.EventAgentEnd:
		return "unit.terminal"
	default:
		return "transition"
	}
}
