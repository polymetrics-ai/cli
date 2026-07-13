package gsd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunnerEmitsHeartbeatDuringSilentLiveProcess(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var heartbeats int
	var questions int
	runner, err := NewRunner(Config{
		Command:           []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir:           t.TempDir(),
		GSDHome:           t.TempDir(),
		Model:             "openai-codex/gpt-5.6-sol",
		Thinking:          "high",
		Timeout:           5 * time.Second,
		HeartbeatInterval: 10 * time.Millisecond,
		MaxEventBytes:     1024,
		Environment:       []string{"GO_WANT_RUNNER_HELPER=1", "RUNNER_HELPER_MODE=silent", "GH_TOKEN=must-not-reach-child"},
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	result := runner.Run(context.Background(), "auto", nil, Observer{
		Heartbeat: func(Heartbeat) {
			mu.Lock()
			heartbeats++
			mu.Unlock()
		},
		Question: func(context.Context, Question) (UIResponse, error) {
			questions++
			return UIResponse{Cancelled: true}, nil
		},
	})
	if result.Terminal != TerminalSuccess {
		t.Fatalf("terminal=%s error=%v", result.Terminal, result.Err)
	}
	mu.Lock()
	defer mu.Unlock()
	if heartbeats == 0 {
		t.Fatal("expected at least one supervisor heartbeat")
	}
	if questions != 0 {
		t.Fatalf("fire-and-forget UI updates reached human gate: %d", questions)
	}
}

func TestRunnerClassifiesBlockedExit(t *testing.T) {
	t.Parallel()

	runner, err := NewRunner(Config{
		Command:           []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir:           t.TempDir(),
		GSDHome:           t.TempDir(),
		Model:             "openai-codex/gpt-5.6-sol",
		Thinking:          "high",
		Timeout:           5 * time.Second,
		HeartbeatInterval: 10 * time.Millisecond,
		MaxEventBytes:     1024,
		Environment:       []string{"GO_WANT_RUNNER_HELPER=1", "RUNNER_HELPER_MODE=blocked"},
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	for attempt := 0; attempt < 25; attempt++ {
		result := runner.Run(context.Background(), "next", nil, Observer{})
		if result.Terminal != TerminalBlocked || result.ExitCode != 10 {
			t.Fatalf("attempt=%d result=%+v", attempt, result)
		}
	}
}

func TestRunnerRejectsUnsupportedCommandAndModel(t *testing.T) {
	t.Parallel()

	_, err := NewRunner(Config{Command: []string{"gsd"}, WorkDir: t.TempDir(), GSDHome: t.TempDir(), Model: "openai-codex/gpt-5.5", Thinking: "high"})
	if err == nil {
		t.Fatal("expected model downgrade to fail")
	}
	runner, err := NewRunner(Config{Command: []string{"gsd"}, WorkDir: t.TempDir(), GSDHome: t.TempDir(), Model: "openai-codex/gpt-5.6-sol", Thinking: "high"})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	result := runner.Run(context.Background(), "plan", nil, Observer{})
	if result.Terminal != TerminalRejected {
		t.Fatalf("unsupported command terminal=%s", result.Terminal)
	}
	result = runner.Run(context.Background(), "recover", nil, Observer{})
	if result.Terminal != TerminalRejected {
		t.Fatalf("destructive recover terminal=%s", result.Terminal)
	}
}

func TestRunnerKeepsHeartbeatsWhileHumanGateWaitsAndCancelsBeforeUpstreamFallback(t *testing.T) {
	t.Parallel()

	var heartbeats atomic.Int64
	runner, err := NewRunner(Config{
		Command:           []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir:           t.TempDir(),
		GSDHome:           t.TempDir(),
		Model:             "openai-codex/gpt-5.6-sol",
		Thinking:          "high",
		Timeout:           90 * time.Millisecond,
		HeartbeatInterval: 10 * time.Millisecond,
		MaxEventBytes:     1024,
		Environment:       []string{"GO_WANT_RUNNER_HELPER=1", "RUNNER_HELPER_MODE=question"},
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	result := runner.Run(context.Background(), "next", nil, Observer{
		Heartbeat: func(Heartbeat) { heartbeats.Add(1) },
		Question: func(ctx context.Context, _ Question) (UIResponse, error) {
			<-ctx.Done()
			return UIResponse{Cancelled: true}, ctx.Err()
		},
	})
	if result.Terminal != TerminalTimeout {
		t.Fatalf("terminal=%s error=%v", result.Terminal, result.Err)
	}
	if got := heartbeats.Load(); got < 2 {
		t.Fatalf("heartbeats while awaiting gate=%d want >=2", got)
	}
}

func TestRunnerQueriesSupportedSurface(t *testing.T) {
	t.Parallel()

	runner, err := NewRunner(Config{
		Command: []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir: t.TempDir(), GSDHome: t.TempDir(), Model: "openai-codex/gpt-5.6-sol", Thinking: "high",
		Environment: []string{"GO_WANT_RUNNER_HELPER=1", "RUNNER_HELPER_MODE=query"},
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	snapshot, err := runner.Query(context.Background())
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if snapshot.MilestoneID != "M001" || snapshot.Next.Action != "stop" {
		t.Fatalf("snapshot=%+v", snapshot)
	}
}

func TestRunnerUsesNativeDBToMarkdownRepairCommand(t *testing.T) {
	t.Parallel()
	workDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(workDir, ".gsd"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, ".gsd", "notifications.jsonl"), []byte(`{"message":"gsd rebuild markdown: rebuilt markdown projections from the canonical DB\n  Rendered: 1\n  Skipped: 0"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner, err := NewRunner(Config{
		Command: []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir: workDir, GSDHome: t.TempDir(), Model: "openai-codex/gpt-5.6-sol", Thinking: "high",
		Environment: []string{"GO_WANT_RUNNER_HELPER=1", "RUNNER_HELPER_MODE=repair"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.RebuildMarkdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestRebuildNotificationFailsOnProjectionErrors(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "notifications.jsonl")
	if err := os.WriteFile(path, []byte(`{"message":"gsd rebuild markdown: rebuilt markdown projections from the canonical DB\n  Errors: 1"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := validateRebuildNotification(path); err == nil {
		t.Fatal("expected projection error to fail maintenance")
	}
}

func TestRunnerHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_RUNNER_HELPER") != "1" {
		return
	}
	args := strings.Join(os.Args, " ")
	if os.Getenv("GH_TOKEN") != "" || os.Getenv("GIT_CONFIG_COUNT") != "2" {
		fmt.Fprintln(os.Stderr, "governed environment was not enforced")
		os.Exit(1)
	}
	if os.Getenv("RUNNER_HELPER_MODE") != "repair" && (!strings.Contains(args, "headless") || (os.Getenv("RUNNER_HELPER_MODE") != "query" &&
		os.Getenv("RUNNER_HELPER_MODE") != "repair" &&
		(!strings.Contains(args, "--model openai-codex/gpt-5.6-sol") ||
			!strings.Contains(args, "--response-timeout") || !strings.Contains(args, "--max-restarts 0")))) {
		fmt.Fprintln(os.Stderr, "missing governed headless flags")
		os.Exit(1)
	}
	switch os.Getenv("RUNNER_HELPER_MODE") {
	case "repair":
		if !strings.Contains(args, "--no-session --print /gsd rebuild markdown") {
			os.Exit(1)
		}
		os.Exit(0)
	case "silent":
		fmt.Println(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.6-sol"},"source":"restore"}`)
		fmt.Println(`{"type":"thinking_level_select","level":"high","previousLevel":"off"}`)
		fmt.Println(`{"type":"extension_ui_request","id":"status-1","method":"setStatus","message":"working"}`)
		time.Sleep(60 * time.Millisecond)
		os.Exit(0)
	case "blocked":
		os.Exit(10)
	case "question":
		fmt.Println(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.6-sol"},"source":"restore"}`)
		fmt.Println(`{"type":"thinking_level_select","level":"high","previousLevel":"off"}`)
		fmt.Println(`{"type":"extension_ui_request","id":"gate-1","method":"confirm","title":"Continue?"}`)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	case "query":
		fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"phase":"complete","nextAction":"Done","blockers":[]},"next":{"action":"stop","unitType":"","unitId":""}}`)
		os.Exit(0)
	default:
		os.Exit(1)
	}
}
