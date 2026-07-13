package gsd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunnerCorrelatesSemanticQuestionIDAndRespondsToRPCRequest(t *testing.T) {
	t.Parallel()

	runner, err := NewRunner(Config{
		Command: []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir: t.TempDir(), GSDHome: t.TempDir(), StateDir: t.TempDir(), Model: "openai-codex/gpt-5.6-sol", Thinking: "high",
		Timeout: 5 * time.Second, HeartbeatInterval: 25 * time.Millisecond, MaxEventBytes: 4096,
		Environment: []string{"GO_WANT_RUNNER_HELPER=1", "RUNNER_HELPER_MODE=semantic-question"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var observed Question
	result := runner.Run(context.Background(), "next", nil, Observer{
		Question: func(_ context.Context, question Question) (UIResponse, error) {
			observed = question
			return UIResponse{Value: "Confirm (Recommended)"}, nil
		},
	})
	if result.Terminal != TerminalSuccess {
		t.Fatalf("terminal=%s error=%v stderr=%s", result.Terminal, result.Err, result.Stderr)
	}
	if observed.ID != "depth_verification_M001-n6ms9v_confirm" || observed.RequestID != "rpc-42" {
		t.Fatalf("question=%+v", observed)
	}
}

func TestRunnerEmitsHeartbeatDuringSilentLiveProcess(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var heartbeats int
	var questions int
	runner, err := NewRunner(Config{
		Command:           []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir:           t.TempDir(),
		GSDHome:           t.TempDir(),
		StateDir:          t.TempDir(),
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
		StateDir:          t.TempDir(),
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

	_, err := NewRunner(Config{Command: []string{"gsd"}, WorkDir: t.TempDir(), GSDHome: t.TempDir(), StateDir: t.TempDir(), Model: "openai-codex/gpt-5.5", Thinking: "high"})
	if err == nil {
		t.Fatal("expected model downgrade to fail")
	}
	runner, err := NewRunner(Config{
		Command: []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir: t.TempDir(), GSDHome: t.TempDir(), StateDir: t.TempDir(), Model: "openai-codex/gpt-5.6-sol", Thinking: "high",
		Environment: []string{"GO_WANT_RUNNER_HELPER=1", "RUNNER_HELPER_MODE=silent"},
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}
	result := runner.Run(context.Background(), "discuss", []string{"M001"}, Observer{})
	if result.Terminal != TerminalSuccess {
		t.Fatalf("targeted discuss terminal=%s error=%v", result.Terminal, result.Err)
	}
	result = runner.Run(context.Background(), "discuss", []string{"M001", "--resume", "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9"}, Observer{})
	if result.Terminal != TerminalSuccess {
		t.Fatalf("resumed discuss terminal=%s error=%v", result.Terminal, result.Err)
	}
	result = runner.Run(context.Background(), "plan", nil, Observer{})
	if result.Terminal != TerminalRejected {
		t.Fatalf("unsupported command terminal=%s", result.Terminal)
	}
	result = runner.Run(context.Background(), "recover", nil, Observer{})
	if result.Terminal != TerminalRejected {
		t.Fatalf("destructive recover terminal=%s", result.Terminal)
	}
}

func TestRunnerRequiresDeliveryStateOutsideProject(t *testing.T) {
	t.Parallel()

	workDir := t.TempDir()
	_, err := NewRunner(Config{
		Command: []string{"gsd"}, WorkDir: workDir, GSDHome: t.TempDir(),
		StateDir: filepath.Join(workDir, ".shepherd"),
		Model:    "openai-codex/gpt-5.6-sol", Thinking: "high",
	})
	if err == nil {
		t.Fatal("expected state inside the project to be rejected")
	}
}

func TestRunnerBridgesOfficialHeadlessResumeToPiSessions(t *testing.T) {
	t.Parallel()
	gsdHome := t.TempDir()
	_, err := NewRunner(Config{
		Command: []string{"gsd"}, WorkDir: t.TempDir(), GSDHome: gsdHome, StateDir: t.TempDir(),
		Model: "openai-codex/gpt-5.6-sol", Thinking: "high",
	})
	if err != nil {
		t.Fatal(err)
	}
	target, err := os.Readlink(filepath.Join(gsdHome, "sessions"))
	if err != nil {
		t.Fatal(err)
	}
	if target != filepath.Join("agent", "sessions") {
		t.Fatalf("resume bridge=%q", target)
	}
}

func TestRunnerKeepsHeartbeatsWhileHumanGateWaitsAndCancelsBeforeUpstreamFallback(t *testing.T) {
	t.Parallel()

	var heartbeats atomic.Int64
	runner, err := NewRunner(Config{
		Command:           []string{os.Args[0], "-test.run=TestRunnerHelperProcess", "--"},
		WorkDir:           t.TempDir(),
		GSDHome:           t.TempDir(),
		StateDir:          t.TempDir(),
		Model:             "openai-codex/gpt-5.6-sol",
		Thinking:          "high",
		Timeout:           500 * time.Millisecond,
		HeartbeatInterval: 25 * time.Millisecond,
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
		WorkDir: t.TempDir(), GSDHome: t.TempDir(), StateDir: t.TempDir(), Model: "openai-codex/gpt-5.6-sol", Thinking: "high",
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
		WorkDir: workDir, GSDHome: t.TempDir(), StateDir: t.TempDir(), Model: "openai-codex/gpt-5.6-sol", Thinking: "high",
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
	if os.Getenv("GH_TOKEN") != "" || os.Getenv("GIT_CONFIG_COUNT") != "5" ||
		os.Getenv("GSD_STATE_DIR") == "" ||
		os.Getenv("GIT_CONFIG_KEY_3") != "user.name" ||
		os.Getenv("GIT_CONFIG_VALUE_3") != "Polymetrics Shepherd" ||
		os.Getenv("GIT_CONFIG_KEY_4") != "user.email" ||
		os.Getenv("GIT_CONFIG_VALUE_4") != "shepherd@localhost.invalid" {
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
	case "semantic-question":
		fmt.Println(`{"type":"model_select","model":{"provider":"openai-codex","id":"gpt-5.6-sol"},"source":"restore"}`)
		fmt.Println(`{"type":"thinking_level_select","level":"high","previousLevel":"off"}`)
		fmt.Println(`{"type":"tool_execution_start","toolName":"ask_user_questions","toolCallId":"tool-1","input":{"questions":[{"id":"depth_verification_M001-n6ms9v_confirm","header":"Depth check","question":"Confirm that the printed depth verification is sufficient.","options":[{"label":"Confirm (Recommended)"},{"label":"Decline"}]}]}}`)
		fmt.Println(`{"type":"extension_ui_request","id":"rpc-42","method":"select","title":"Depth check: Confirm that the printed depth verification is sufficient.","options":["Confirm (Recommended)","Decline"]}`)
		var response struct {
			Type  string `json:"type"`
			ID    string `json:"id"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(bufio.NewReader(os.Stdin)).Decode(&response); err != nil || response.Type != "extension_ui_response" || response.ID != "rpc-42" || response.Value != "Confirm (Recommended)" {
			fmt.Fprintf(os.Stderr, "invalid response: %+v err=%v\n", response, err)
			os.Exit(1)
		}
		os.Exit(0)
	case "query":
		fmt.Print(`{"state":{"activeMilestone":{"id":"M001"},"phase":"complete","nextAction":"Done","blockers":[]},"next":{"action":"stop","unitType":"","unitId":""}}`)
		os.Exit(0)
	default:
		os.Exit(1)
	}
}
