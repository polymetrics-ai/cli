package coordination

import (
	"context"
	"errors"
	"fmt"
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
		_ = writeUnixRateBudgetResponse(conn, unixRateBudgetResponse{Version: unixRateBudgetProtocolVersion, Kind: unixRateBudgetReady, Ready: true})
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
	runDir, socketPath := owner.runDir, owner.socketPath
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

	results := make(chan string, 8)
	for range 8 {
		go func() {
			child := exec.Command(os.Args[0], "-test.run=^TestRateBudgetCoordinatorHelper$", "-test.v")
			child.Env = []string{
				rateBudgetHelperEnv + "=1",
				rateBudgetHelperPathEnv + "=" + client.socketPath,
				rateBudgetHelperEpochEnv + "=" + client.epoch,
			}
			output, err := child.CombinedOutput()
			if err != nil {
				results <- fmt.Sprintf("child error: %v: %s", err, output)
				return
			}
			results <- string(output)
		}()
	}
	grants, refusals := 0, 0
	for range 8 {
		output := <-results
		switch {
		case strings.Contains(output, "rate-budget-helper=granted"):
			grants++
		case strings.Contains(output, "rate-budget-helper=refused"):
			refusals++
		default:
			t.Fatalf("helper returned unexpected result: %s", output)
		}
	}
	if grants != 3 || refusals != 5 {
		t.Fatalf("shared tiny budget grants=%d refusals=%d, want 3 and 5", grants, refusals)
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

func TestRateBudgetCoordinatorHelper(t *testing.T) {
	if os.Getenv(rateBudgetHelperEnv) != "1" {
		return
	}
	client := &UnixRateBudgetCoordinatorClient{
		socketPath: os.Getenv(rateBudgetHelperPathEnv),
		epoch:      os.Getenv(rateBudgetHelperEpochEnv),
	}
	batch := connsdk.ReservationBatch{Policies: []connsdk.ReservationPolicy{testReservationPolicy(t, "tiny", "opaque-shared-scope", 3)}}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	decision, err := client.Decide(ctx, batch)
	if err != nil {
		t.Fatalf("helper decide: %v", err)
	}
	if decision.Granted {
		fmt.Println("rate-budget-helper=granted")
		return
	}
	fmt.Println("rate-budget-helper=refused")
}
