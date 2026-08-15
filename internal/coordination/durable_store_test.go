package coordination

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const durableCoordinationHelperEnv = "POLYMETRICS_DURABLE_COORDINATION_HELPER"

// TestDurableCoordinationStoresSurviveKilledWriterProcess is intentionally a
// parent/child process test. Reconstructing coordinators in one process would
// not prove that the exact crash these mechanisms guard against is survivable.
func TestDurableCoordinationStoresSurviveKilledWriterProcess(t *testing.T) {
	if os.Getenv(durableCoordinationHelperEnv) == "1" {
		runDurableCoordinationWriter(t)
		return
	}

	dir := t.TempDir()
	command := exec.Command(os.Args[0], "-test.run=^TestDurableCoordinationStoresSurviveKilledWriterProcess$")
	command.Env = append(os.Environ(), durableCoordinationHelperEnv+"=1", "POLYMETRICS_DURABLE_COORDINATION_DIR="+dir)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe: %v", err)
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		t.Fatalf("start writer process: %v", err)
	}
	ready, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || ready != "durable\n" {
		_ = command.Process.Kill()
		t.Fatalf("writer readiness = %q, %v", ready, err)
	}
	if err := command.Process.Kill(); err != nil {
		t.Fatalf("kill writer process: %v", err)
	}
	if err := command.Wait(); err == nil {
		t.Fatal("killed writer process unexpectedly exited successfully")
	}

	authStore, err := OpenFileAuthCohortHealthStore(filepath.Join(dir, "auth.json"))
	if err != nil {
		t.Fatalf("reopen auth store after killed process: %v", err)
	}
	cohort := testAuthCohortKey(t, "durable-process-restart")
	restartedAuth := NewAuthCohortCoordinator(authStore)
	if _, err := restartedAuth.Admit(context.Background(), cohort); !errors.Is(err, ErrAuthCohortFenced) {
		t.Fatalf("post-crash auth admission = %v, want ErrAuthCohortFenced", err)
	}

	parkingStore, err := OpenFileRateParkingStore(filepath.Join(dir, "parking.json"))
	if err != nil {
		t.Fatalf("reopen parking store after killed process: %v", err)
	}
	resumed := make(chan ParkedRateLimitRun, 1)
	restartedParking := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
		Store: parkingStore,
		Resume: func(_ context.Context, run ParkedRateLimitRun) error {
			resumed <- run
			return nil
		},
	})
	if err := restartedParking.Start(context.Background()); err != nil {
		t.Fatalf("restart parking coordinator: %v", err)
	}
	t.Cleanup(restartedParking.Close)
	select {
	case run := <-resumed:
		if run.RunID != "durable-run" {
			t.Fatalf("resumed run = %q, want durable-run", run.RunID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("parked run was not resumed by the restarted process")
	}
	if runs, err := parkingStore.List(); err != nil || len(runs) != 0 {
		t.Fatalf("parking store after restart resume = %#v, %v; want empty", runs, err)
	}
}

func runDurableCoordinationWriter(t *testing.T) {
	dir := os.Getenv("POLYMETRICS_DURABLE_COORDINATION_DIR")
	if dir == "" {
		t.Fatal("helper coordination directory is unavailable")
	}
	authStore, err := OpenFileAuthCohortHealthStore(filepath.Join(dir, "auth.json"))
	if err != nil {
		t.Fatalf("open auth store: %v", err)
	}
	auth := NewAuthCohortCoordinator(authStore)
	member, err := auth.Admit(context.Background(), testAuthCohortKey(t, "durable-process-restart"))
	if err != nil {
		t.Fatalf("admit writer member: %v", err)
	}
	if err := auth.Report(member, AuthenticationOutcomeVerifiedInvalid); err != nil {
		t.Fatalf("persist fence: %v", err)
	}

	parkingStore, err := OpenFileRateParkingStore(filepath.Join(dir, "parking.json"))
	if err != nil {
		t.Fatalf("open parking store: %v", err)
	}
	parking := NewRateParkingCoordinator(RateParkingCoordinatorOptions{
		Store: parkingStore,
		Resume: func(context.Context, ParkedRateLimitRun) error { return nil },
	})
	if err := parking.Start(context.Background()); err != nil {
		t.Fatalf("start parking coordinator: %v", err)
	}
	if _, err := parking.Park(context.Background(), RateParkingRequest{
		RunID:      "durable-run",
		Scope:      connectors.RateLimitScopeKey("durable-scope"),
		Checkpoint: testParkedCheckpoint(time.Now().UTC().Add(-time.Second)),
		ResetAt:    time.Now().UTC().Add(-time.Millisecond),
		Reason:     connsdk.RateLimitObservationSourceRetryAfter,
	}); err != nil {
		t.Fatalf("persist parked run: %v", err)
	}

	if _, err := fmt.Fprintln(os.Stdout, "durable"); err != nil {
		t.Fatalf("signal durability: %v", err)
	}
	select {}
}

func TestFileCoordinationStoresRejectSchemaDriftWithoutMutation(t *testing.T) {
	for _, name := range []string{"auth.json", "parking.json"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(path, []byte(`{"schema_version":999,"records":{}}`), 0o600); err != nil {
				t.Fatalf("write drifted store: %v", err)
			}
			var err error
			if name == "auth.json" {
				_, err = OpenFileAuthCohortHealthStore(path)
			} else {
				_, err = OpenFileRateParkingStore(path)
			}
			if !errors.Is(err, ErrCoordinationStoreSchema) {
				t.Fatalf("schema drift error = %v, want ErrCoordinationStoreSchema", err)
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read refused store: %v", readErr)
			}
			if string(data) != `{"schema_version":999,"records":{}}` {
				t.Fatalf("schema refusal mutated durable state: %q", data)
			}
		})
	}
}
