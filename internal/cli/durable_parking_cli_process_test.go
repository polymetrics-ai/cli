package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/coordination"
)

const durableParkingCLIHelperEnv = "POLYMETRICS_DURABLE_PARKING_CLI_HELPER"

func init() {
	if os.Getenv(durableParkingCLIHelperEnv) == "" {
		return
	}
	connectors.RegisterDefaultRegistryBuilder(func() *connectors.Registry {
		registry := bundleregistry.New()
		registry.Register(durableParkingCLIEngineConnector())
		return registry
	})
}

func durableParkingCLIEngineConnector() connectors.Connector {
	rawSchema := json.RawMessage(`{
		"type":"object",
		"x-primary-key":["id"],
		"x-cursor-field":"updated_at",
		"properties":{"id":{"type":"string"},"name":{"type":"string"},"updated_at":{"type":"string"}}
	}`)
	schema, err := engine.CompileSchema(rawSchema)
	if err != nil {
		panic(err)
	}
	limit, window := 1000, 60
	return engine.New(engine.Bundle{
		Name: "sample",
		Metadata: engine.Metadata{Name: "sample", DisplayName: "Parking Fixture", IntegrationType: "source", Capabilities: engine.Capabilities{
			Check: true, Read: true,
		}},
		HTTP: engine.HTTPBase{
			URL:      os.Getenv("POLYMETRICS_DURABLE_PARKING_CLI_URL"),
			Check:    &engine.RequestSpec{Method: http.MethodGet, Path: "/customers"},
			ErrorMap: []engine.ErrorRule{{Status: http.StatusTooManyRequests, Class: "rate_limited", Hint: "fixture reset is authoritative"}},
		},
		Streams: []engine.StreamSpec{{Name: "customers", Path: "/customers", Records: engine.RecordsSpec{Path: "data"}, SchemaRef: "schemas/customers.json"}},
		Schemas: map[string]*engine.StreamSchema{"customers": {Schema: schema, PrimaryKey: schema.PrimaryKeys(), CursorField: schema.CursorFieldName(), Raw: rawSchema}},
		RateLimits: &connsdk.RateLimits{SchemaVersion: 1, State: connsdk.RateLimitStateDeclared, Policies: []connsdk.RateLimitPolicy{{
			ID: "fixture-account", Selector: connsdk.RateLimitSelector{All: true},
			Scope:   connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "account_id"},
			Budgets: []connsdk.RateLimitBudget{{Model: connsdk.RateLimitBudgetFixedWindow, Dimension: connsdk.RateLimitBudgetSustained, Unit: connsdk.RateLimitBudgetRequests, Limit: &limit, WindowSeconds: &window}},
		}}},
	}, nil)
}

func TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess(t *testing.T) {
	if mode := os.Getenv(durableParkingCLIHelperEnv); mode != "" {
		runDurableParkingCLIHelper(t, mode)
		return
	}
	root := t.TempDir()
	var providerSends atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := providerSends.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if attempt >= 2 && attempt <= 6 {
			if attempt == 6 {
				w.Header().Set("Retry-After", "4")
			} else {
				w.Header().Set("Retry-After", "0")
			}
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"cli","updated_at":"2026-08-15T00:00:00Z"}]}`))
	}))
	t.Cleanup(server.Close)
	setup := durableParkingCLICommand(root, server.URL, "setup", "")
	output, err := setup.StdoutPipe()
	if err != nil {
		t.Fatal("could not capture parking CLI setup")
	}
	setup.Stderr = os.Stderr
	if err := setup.Start(); err != nil {
		t.Fatal("could not start parking CLI setup")
	}
	ready := make([]byte, 512)
	n, readErr := output.Read(ready)
	line := string(ready[:n])
	if readErr != nil || !strings.HasPrefix(line, "parked:") {
		rest, _ := io.ReadAll(output)
		_ = setup.Process.Kill()
		t.Fatalf("parking CLI setup did not persist its record: output=%q read=%v", line+string(rest), readErr)
	}
	runID := strings.TrimSpace(strings.TrimPrefix(line, "parked:"))
	if err := setup.Process.Kill(); err != nil {
		t.Fatal("could not kill parking CLI setup")
	}
	if err := setup.Wait(); err == nil {
		t.Fatal("killed parking CLI setup unexpectedly succeeded")
	}
	if got := providerSends.Load(); got != 6 {
		t.Fatalf("pre-restart CLI provider sends = %d, want one success plus five terminal-429 attempts", got)
	}
	checkpointBefore := durableParkingCLIStreamStates(t, root)
	tableBefore := durableParkingCLITable(t, root)
	if err := durableParkingCLICommand(root, server.URL, "admission", runID).Run(); err != nil {
		t.Fatal("restarted CLI admission helper failed")
	}
	if got := providerSends.Load(); got != 6 {
		t.Fatalf("parked CLI provider sends = %d, want 6 (zero new sends)", got)
	}
	if after := durableParkingCLIStreamStates(t, root); !bytes.Equal(after, checkpointBefore) {
		t.Fatal("fenced CLI admission advanced a checkpoint")
	}
	if after := durableParkingCLITable(t, root); !bytes.Equal(after, tableBefore) {
		t.Fatal("parked CLI admission wrote the destination table")
	}

	t.Run("edge: concurrent durable reopeners share one resume claim", func(t *testing.T) {
		time.Sleep(4100 * time.Millisecond)
		resumeCommands := []*exec.Cmd{
			durableParkingCLICommand(root, server.URL, "resume-race", runID),
			durableParkingCLICommand(root, server.URL, "resume-race", runID),
		}
		for _, command := range resumeCommands {
			if err := command.Start(); err != nil {
				t.Fatal("could not start concurrent CLI resume helper")
			}
		}
		for _, command := range resumeCommands {
			if err := command.Wait(); err != nil {
				t.Fatal("concurrent CLI resume helper failed")
			}
		}
		if got := providerSends.Load(); got != 7 {
			t.Fatalf("CLI provider sends after concurrent resume = %d, want 7 (one claim winner)", got)
		}
		if err := durableParkingCLICommand(root, server.URL, "verify-resume", runID).Run(); err != nil {
			t.Fatal("restarted CLI resume verification failed")
		}
		if got := providerSends.Load(); got != 7 {
			t.Fatalf("CLI provider sends after resumed replay = %d, want 7", got)
		}
	})
}

func durableParkingCLICommand(root, serverURL, mode, runID string) *exec.Cmd {
	command := exec.Command(os.Args[0], "-test.run=^TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess$")
	command.Env = append(os.Environ(), durableParkingCLIHelperEnv+"="+mode,
		"POLYMETRICS_DURABLE_PARKING_CLI_ROOT="+root,
		"POLYMETRICS_DURABLE_PARKING_CLI_URL="+serverURL,
		"POLYMETRICS_DURABLE_PARKING_CLI_RUN="+runID)
	return command
}

func runDurableParkingCLIHelper(t *testing.T, mode string) {
	root := os.Getenv("POLYMETRICS_DURABLE_PARKING_CLI_ROOT")
	runID := os.Getenv("POLYMETRICS_DURABLE_PARKING_CLI_RUN")
	run := func(args ...string) (int, string) {
		var stdout, stderr bytes.Buffer
		code := cli.Run(args, &stdout, &stderr)
		return code, stdout.String() + stderr.String()
	}
	switch mode {
	case "setup":
		commands := [][]string{
			{"init", "--root", root, "--json"},
			{"credentials", "add", "parking", "--connector", "sample", "--config", "account_id=fixture-account", "--root", root, "--json"},
			{"credentials", "add", "warehouse", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
			{"connections", "create", "parking-cli", "--source", "sample:parking", "--destination", "warehouse:warehouse", "--stream", "customers", "--primary-key", "id", "--cursor", "updated_at", "--table", "parking_cli_customers", "--root", root, "--json"},
			{"etl", "run", "--connection", "parking-cli", "--stream", "customers", "--root", root, "--json"},
		}
		for _, args := range commands {
			if code, _ := run(args...); code != 0 {
				t.Fatal("parking CLI setup command failed")
			}
		}
		if code, _ := run("etl", "run", "--connection", "parking-cli", "--stream", "customers", "--root", root, "--json"); code == 0 {
			t.Fatal("terminal 429 did not park the CLI run")
		}
		parked := durableParkingCLILatestRun(t, root)
		if parked.Status != string(coordination.RateParkingOutcomeParkedRateLimit) {
			t.Fatalf("latest CLI run status = %q", parked.Status)
		}
		fmt.Printf("parked:%s\n", parked.ID)
		select {}
	case "admission":
		before := durableParkingCLIStreamStates(t, root)
		code, output := run("etl", "run", "--connection", "parking-cli", "--stream", "customers", "--root", root, "--json")
		if code == 0 || !strings.Contains(output, coordination.ErrRateLimitParked.Error()) {
			t.Fatal("restarted CLI did not return the parked typed refusal")
		}
		if after := durableParkingCLIStreamStates(t, root); !bytes.Equal(after, before) {
			t.Fatal("restarted CLI parked refusal advanced the checkpoint")
		}
	case "resume-race":
		code, output := run("etl", "status", runID, "--root", root, "--json")
		if code != 0 {
			t.Fatalf("CLI app.Open could not participate in durable resume claim: exit=%d output=%q", code, output)
		}
	case "verify-resume":
		code, output := run("etl", "status", runID, "--root", root, "--json")
		if code != 0 || !strings.Contains(output, `"status": "resumed"`) {
			t.Fatal("CLI app.Open did not resume the durable parked run")
		}
	default:
		t.Fatalf("unknown parking CLI helper mode %q", mode)
	}
}

type durableParkingCLIRun struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func durableParkingCLILatestRun(t *testing.T, root string) durableParkingCLIRun {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatal("could not read CLI run state")
	}
	var state struct {
		Runs []durableParkingCLIRun `json:"runs"`
	}
	if err := json.Unmarshal(data, &state); err != nil || len(state.Runs) == 0 {
		t.Fatal("CLI run state had an unexpected shape")
	}
	return state.Runs[len(state.Runs)-1]
}

func durableParkingCLIStreamStates(t *testing.T, root string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatal("could not read CLI checkpoint state")
	}
	var state struct {
		StreamStates json.RawMessage `json:"stream_states"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal("could not decode CLI checkpoint state")
	}
	return append([]byte(nil), state.StreamStates...)
}

func durableParkingCLITable(t *testing.T, root string) []byte {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(root, ".polymetrics", "warehouse", "*", "*", "*", "tables", "parking_cli_customers.parquet"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("parking CLI destination table paths = %#v, %v", paths, err)
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal("could not read parking CLI destination table")
	}
	return data
}
