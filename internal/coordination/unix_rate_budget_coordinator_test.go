package coordination

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	rateBudgetHelperEnv      = "PM_RATE_BUDGET_HELPER"
	rateBudgetHelperModeEnv  = "PM_RATE_BUDGET_HELPER_MODE"
	rateBudgetHelperPathEnv  = "PM_RATE_BUDGET_HELPER_SOCKET"
	rateBudgetHelperEpochEnv = "PM_RATE_BUDGET_HELPER_EPOCH"
)

func newUnixRateBudgetTestListener(t *testing.T) (*net.UnixListener, string) {
	t.Helper()
	runDir, err := os.MkdirTemp("/tmp", "pmrb-test-")
	if err != nil {
		t.Fatalf("create rate-budget test directory: %v", err)
	}
	socketPath := filepath.Join(runDir, "s")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		_ = os.Remove(runDir)
		t.Fatalf("listen for rate-budget test: %v", err)
	}
	t.Cleanup(func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		_ = os.Remove(runDir)
	})
	return listener, socketPath
}

func TestUnixRateBudgetCoordinatorClientCancellationInterruptsStalledExchange(t *testing.T) {
	listener, socketPath := newUnixRateBudgetTestListener(t)
	accepted := make(chan struct{})
	release := make(chan struct{}, 1)
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.AcceptUnix()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		close(accepted)
		<-release
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case release <- struct{}{}:
		default:
		}
		select {
		case <-serverDone:
		case <-time.After(time.Second):
			t.Error("stalled rate-budget test server did not stop")
		}
	})

	client := &UnixRateBudgetCoordinatorClient{socketPath: socketPath, epoch: "test-epoch"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- client.Ready(ctx) }()
	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("rate-budget test server did not accept the client")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled exchange error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled exchange did not return promptly")
	}
}

func TestUnixRateBudgetCoordinatorClientCancellationWinsResponseRace(t *testing.T) {
	listener, socketPath := newUnixRateBudgetTestListener(t)
	requestRead := make(chan struct{})
	release := make(chan struct{}, 1)
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		conn, err := listener.AcceptUnix()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		if _, err := readUnixRateBudgetRequest(conn); err != nil {
			return
		}
		close(requestRead)
		<-release
		_ = writeUnixRateBudgetResponse(conn, unixRateBudgetResponse{
			Version: unixRateBudgetProtocolVersion,
			Kind:    unixRateBudgetReady,
			Ready:   true,
		})
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		select {
		case release <- struct{}{}:
		default:
		}
		select {
		case <-serverDone:
		case <-time.After(time.Second):
			t.Error("response-race rate-budget test server did not stop")
		}
	})

	client := &UnixRateBudgetCoordinatorClient{socketPath: socketPath, epoch: "test-epoch"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- client.Ready(ctx) }()
	select {
	case <-requestRead:
	case <-time.After(time.Second):
		t.Fatal("rate-budget test server did not read the request")
	}
	cancel()
	release <- struct{}{}
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("response after cancellation error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled response race did not return promptly")
	}
}

func TestUnixRateBudgetCoordinatorMultiProcessTinyBudget(t *testing.T) {
	owner, client, err := StartUnixRateBudgetCoordinator(context.Background(), UnixRateBudgetCoordinatorOptions{
		MaxInFlight: 8,
		LeaseTTL:    time.Second,
	})
	if err != nil {
		t.Fatalf("start shared coordinator: %v", err)
	}
	t.Cleanup(func() { _ = owner.Close() })
	runDir := owner.runDir
	socketPath := owner.socketPath
	if info, err := os.Stat(runDir); err != nil {
		t.Fatalf("stat shared coordinator run directory: %v", err)
	} else if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("shared coordinator run directory mode = %o, want 700", got)
	}
	if info, err := os.Stat(socketPath); err != nil {
		t.Fatalf("stat shared coordinator socket: %v", err)
	} else if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("shared coordinator socket mode = %o, want 600", got)
	}

	shared := runRateBudgetHelpers(t, rateBudgetHelperConfig{
		mode:       "shared",
		socketPath: client.socketPath,
		epoch:      client.epoch,
	})
	if grants, refusals := countRateBudgetHelperResults(shared); grants != 3 || refusals != 5 {
		t.Fatalf("shared helper results = %v, want 3 grants and 5 typed refusals", shared)
	}

	local := runRateBudgetHelpers(t, rateBudgetHelperConfig{mode: "process_local"})
	if grants, refusals := countRateBudgetHelperResults(local); grants != 8 || refusals != 0 {
		t.Fatalf("process-local helper results = %v, want 8 grants", local)
	}

	if err := owner.Close(); err != nil {
		t.Fatalf("close shared coordinator: %v", err)
	}
	if _, err := os.Stat(runDir); !os.IsNotExist(err) {
		t.Fatal("shared coordinator run directory remains after close")
	}
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Fatal("shared coordinator socket remains after close")
	}
}

func TestSharedRateBudgetScopesRemainIndependent(t *testing.T) {
	owner, client, err := StartUnixRateBudgetCoordinator(context.Background(), UnixRateBudgetCoordinatorOptions{
		MaxInFlight: 4,
		LeaseTTL:    time.Second,
	})
	if err != nil {
		t.Fatalf("start shared coordinator: %v", err)
	}
	t.Cleanup(func() { _ = owner.Close() })

	ctx := context.Background()
	scopeA := testReservationPolicy(t, "core", "opaque-scope-a", 1)
	scopeB := testReservationPolicy(t, "core", "opaque-scope-b", 1)
	first, err := client.Decide(ctx, testReservationBatch(scopeA))
	if err != nil || !first.Granted {
		t.Fatalf("scope A initial decision = %+v, %v", first, err)
	}
	if err := client.Finish(ctx, first.Lease, connsdk.CompletionObservation{Attempted: true}); err != nil {
		t.Fatalf("finish scope A: %v", err)
	}

	secondA, err := client.Decide(ctx, testReservationBatch(scopeA))
	if err != nil {
		t.Fatalf("scope A second decision: %v", err)
	}
	if secondA.Granted || secondA.NotBefore.IsZero() {
		t.Fatalf("scope A second decision = %+v, want a typed non-grant", secondA)
	}

	firstB, err := client.Decide(ctx, testReservationBatch(scopeB))
	if err != nil || !firstB.Granted {
		t.Fatalf("independent scope B decision = %+v, %v", firstB, err)
	}
}

func TestSharedRateBudgetOwnerCrashFailsClosed(t *testing.T) {
	owner, client, err := StartUnixRateBudgetCoordinator(context.Background(), UnixRateBudgetCoordinatorOptions{
		MaxInFlight: 1,
		LeaseTTL:    time.Second,
	})
	if err != nil {
		t.Fatalf("start shared coordinator: %v", err)
	}
	batch := testReservationBatch(testReservationPolicy(t, "core", "opaque-scope-a", 2))
	granted, err := client.Decide(context.Background(), batch)
	if err != nil || !granted.Granted {
		t.Fatalf("initial shared decision = %+v, %v", granted, err)
	}
	if err := owner.Close(); err != nil {
		t.Fatalf("crash owner: %v", err)
	}

	if _, err := client.Decide(context.Background(), batch); err == nil {
		t.Fatal("owner loss admitted work instead of failing closed")
	} else if !IsSharedCoordinatorUnavailable(err) {
		t.Fatalf("owner loss error is not typed unavailable: %T", err)
	}

	freshOwner, freshClient, err := StartUnixRateBudgetCoordinator(context.Background(), UnixRateBudgetCoordinatorOptions{
		MaxInFlight: 1,
		LeaseTTL:    time.Second,
	})
	if err != nil {
		t.Fatalf("start fresh coordinator: %v", err)
	}
	t.Cleanup(func() { _ = freshOwner.Close() })
	oldEpochClient := &UnixRateBudgetCoordinatorClient{socketPath: freshClient.socketPath, epoch: client.epoch}
	if err := oldEpochClient.Ready(context.Background()); err == nil {
		t.Fatal("old client joined a fresh coordinator epoch")
	} else if !IsSharedCoordinatorEpochMismatch(err) {
		t.Fatalf("old epoch error is not typed mismatch: %T", err)
	}
}

type rateBudgetHelperConfig struct {
	mode       string
	socketPath string
	epoch      string
}

func runRateBudgetHelpers(t *testing.T, cfg rateBudgetHelperConfig) []string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve helper executable: %v", err)
	}
	resultReader, resultWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("create helper result pipe: %v", err)
	}
	defer func() { _ = resultReader.Close() }()

	const helpers = 8
	type helperProcess struct {
		command *exec.Cmd
		start   *os.File
	}
	processes := make([]helperProcess, 0, helpers)
	for range helpers {
		startReader, startWriter, err := os.Pipe()
		if err != nil {
			t.Fatalf("create helper start pipe: %v", err)
		}
		command := exec.Command(executable, "-test.run=^TestUnixRateBudgetCoordinatorHelper$", "-test.v=false")
		command.Env = append(os.Environ(),
			rateBudgetHelperEnv+"=1",
			rateBudgetHelperModeEnv+"="+cfg.mode,
			rateBudgetHelperPathEnv+"="+cfg.socketPath,
			rateBudgetHelperEpochEnv+"="+cfg.epoch,
		)
		command.ExtraFiles = []*os.File{startReader, resultWriter}
		command.Stdout = io.Discard
		command.Stderr = io.Discard
		if err := command.Start(); err != nil {
			_ = startReader.Close()
			_ = startWriter.Close()
			t.Fatalf("start helper process: %v", err)
		}
		_ = startReader.Close()
		processes = append(processes, helperProcess{command: command, start: startWriter})
	}
	if err := resultWriter.Close(); err != nil {
		t.Fatalf("close parent result writer: %v", err)
	}
	for _, process := range processes {
		if _, err := process.start.Write([]byte{1}); err != nil {
			t.Fatalf("release helper process: %v", err)
		}
		if err := process.start.Close(); err != nil {
			t.Fatalf("close helper start pipe: %v", err)
		}
	}

	done := make(chan error, 1)
	go func() {
		for _, process := range processes {
			if err := process.command.Wait(); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	resultBytes, readErr := io.ReadAll(resultReader)
	if err := <-done; err != nil {
		t.Fatal("a rate-budget helper process failed")
	}
	if readErr != nil {
		t.Fatalf("read helper results: %v", readErr)
	}
	results := strings.Fields(string(resultBytes))
	if len(results) != helpers {
		t.Fatalf("helper result count = %d, want %d", len(results), helpers)
	}
	return results
}

func countRateBudgetHelperResults(results []string) (grants, refusals int) {
	for _, result := range results {
		switch result {
		case "grant":
			grants++
		case "refusal":
			refusals++
		default:
			return -1, -1
		}
	}
	return grants, refusals
}

func TestUnixRateBudgetCoordinatorHelper(t *testing.T) {
	if os.Getenv(rateBudgetHelperEnv) != "1" {
		return
	}
	start := os.NewFile(uintptr(3), "rate-budget-helper-start")
	result := os.NewFile(uintptr(4), "rate-budget-helper-result")
	if start == nil || result == nil {
		t.Fatal("helper file descriptors are unavailable")
	}
	defer func() { _ = start.Close() }()
	defer func() { _ = result.Close() }()
	var gate [1]byte
	if _, err := io.ReadFull(start, gate[:]); err != nil {
		t.Fatal("helper did not receive release signal")
	}

	batch := testReservationBatch(testReservationPolicy(t, "tiny", "opaque-shared-scope", 3))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var decision connsdk.AdmissionDecision
	var err error
	switch os.Getenv(rateBudgetHelperModeEnv) {
	case "shared":
		client := &UnixRateBudgetCoordinatorClient{
			socketPath: os.Getenv(rateBudgetHelperPathEnv),
			epoch:      os.Getenv(rateBudgetHelperEpochEnv),
		}
		decision, err = client.Decide(ctx, batch)
		if err == nil && decision.Granted {
			err = client.Finish(ctx, decision.Lease, connsdk.CompletionObservation{Attempted: true})
		}
	case "process_local":
		coordinator := NewRateBudgetCoordinator(nil, RateBudgetCoordinatorOptions{MaxInFlight: 8, LeaseTTL: time.Second})
		decision, err = coordinator.Decide(ctx, batch)
		if err == nil && decision.Granted {
			err = coordinator.Finish(ctx, decision.Lease, connsdk.CompletionObservation{Attempted: true})
		}
	default:
		err = fmt.Errorf("unknown helper mode")
	}
	if err != nil {
		_, _ = fmt.Fprintln(result, "error")
		return
	}
	if decision.Granted {
		_, _ = fmt.Fprintln(result, "grant")
		return
	}
	_, _ = fmt.Fprintln(result, "refusal")
}
