package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

type failedRateLimitConnector struct{}

func (failedRateLimitConnector) Name() string { return "failed-rate-limit" }

func (failedRateLimitConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "failed-rate-limit",
		DisplayName:     "Failed Rate Limit Test Connector",
		IntegrationType: "api",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (failedRateLimitConnector) Check(context.Context, connectors.RuntimeConfig) error {
	return nil
}

func (failedRateLimitConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: "failed-rate-limit", Streams: []connectors.Stream{{Name: "widgets"}}}, nil
}

func (failedRateLimitConnector) Read(_ context.Context, req connectors.ReadRequest, _ func(connectors.Record) error) error {
	report := req.Config.RateLimitReport
	report.Declare("failed-rate-limit", connectors.RateLimitDeclarationDeclared)
	report.RecordPolicySelection("failed-rate-limit", "widgets-points", "account", "endpoint")
	report.RecordProviderObservation("failed-rate-limit", "widgets-points", connsdk.RateLimitObservation{
		Limit:        100,
		HasLimit:     true,
		Remaining:    77,
		HasRemaining: true,
	})
	report.RecordPacingWait("failed-rate-limit", 3*time.Millisecond)
	report.RecordProvider429Observed("failed-rate-limit")
	report.RecordProvider429Wait("failed-rate-limit", 5*time.Millisecond, true)
	report.RecordRequestLatency("failed-rate-limit", 7*time.Millisecond)
	return errors.New("token=failed-rate-limit-token-must-not-escape https://provider.example.test/widgets?token=failed-rate-limit-token-must-not-escape: body-must-not-escape")
}

func (failedRateLimitConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func TestETLRunFailureReportsSafeRateLimitCarrier(t *testing.T) {
	a := newFailedRateLimitTestApp(t)

	human, humanErr, humanCode := runFailedRateLimitETL(t, a, false)
	if humanCode != 1 {
		t.Fatalf("human failed ETL exit = %d, want 1", humanCode)
	}
	if humanErr != "" {
		t.Fatalf("human failed ETL stderr = %q, want empty", humanErr)
	}
	for _, want := range []string{
		"ETL run run_",
		"failed: read=0 loaded=0 failed=0",
		"Rate limits: connector=failed-rate-limit declaration=declared",
		"widgets-points,subject_kind=account,selected_by=endpoint,provider_limit=100,provider_remaining=77",
		"local_pacing_wait=3ms",
		"provider_429_observed=1",
		"provider_429_honored=1",
		"provider_429_wait=5ms",
		"request_latency=7ms",
	} {
		if !strings.Contains(human, want) {
			t.Fatalf("human failed ETL output missing %q:\n%s", want, human)
		}
	}

	structured, structuredErr, structuredCode := runFailedRateLimitETL(t, a, true)
	if structuredCode != 1 {
		t.Fatalf("JSON failed ETL exit = %d, want 1", structuredCode)
	}
	if structuredErr != "" {
		t.Fatalf("JSON failed ETL stderr = %q, want empty", structuredErr)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(structured), &payload); err != nil {
		t.Fatalf("unmarshal failed ETL JSON: %v\n%s", err, structured)
	}
	var kind string
	if err := json.Unmarshal(payload["kind"], &kind); err != nil {
		t.Fatalf("unmarshal failed ETL kind: %v", err)
	}
	if kind != "ETLRun" {
		t.Fatalf("failed ETL kind = %q, want ETLRun", kind)
	}
	var run struct {
		ID        string                      `json:"id"`
		Status    string                      `json:"status"`
		RateLimit connectors.RateLimitSummary `json:"rate_limit"`
	}
	if err := json.Unmarshal(payload["run"], &run); err != nil {
		t.Fatalf("unmarshal failed ETL run: %v", err)
	}
	if run.ID == "" || run.Status != "failed" {
		t.Fatalf("failed ETL carrier = %+v, want failed run with ID", run)
	}
	if len(run.RateLimit.Connectors) != 2 {
		t.Fatalf("failed ETL rate-limit connector count = %d, want 2", len(run.RateLimit.Connectors))
	}
	var failedRateLimit, warehouse connectors.RateLimitConnectorSummary
	for _, connector := range run.RateLimit.Connectors {
		switch connector.Connector {
		case "failed-rate-limit":
			failedRateLimit = connector
		case "warehouse":
			warehouse = connector
		}
	}
	if failedRateLimit.Declaration != connectors.RateLimitDeclarationDeclared || failedRateLimit.Provider429Observed != 1 || failedRateLimit.Provider429Honored != 1 || failedRateLimit.ProviderWaitMS != 5 || failedRateLimit.RequestLatencyMS != 7 {
		t.Fatalf("failed ETL JSON rate-limit summary = %+v", failedRateLimit)
	}
	if len(failedRateLimit.Policies) != 1 || failedRateLimit.Policies[0].ID != "widgets-points" || failedRateLimit.Policies[0].SubjectKind != "account" || failedRateLimit.Policies[0].SelectionReason != "endpoint" || failedRateLimit.Policies[0].ProviderLimit == nil || *failedRateLimit.Policies[0].ProviderLimit != 100 || failedRateLimit.Policies[0].ProviderRemaining == nil || *failedRateLimit.Policies[0].ProviderRemaining != 77 {
		t.Fatalf("failed ETL JSON policy summary = %+v", failedRateLimit.Policies)
	}
	if warehouse.Declaration != connectors.RateLimitDeclarationUndeclared {
		t.Fatalf("warehouse declaration = %q, want undeclared", warehouse.Declaration)
	}
	var carrier map[string]json.RawMessage
	if err := json.Unmarshal(payload["run"], &carrier); err != nil {
		t.Fatalf("unmarshal failed ETL carrier: %v", err)
	}
	for _, omitted := range []string{"error", "checkpoint"} {
		if _, ok := carrier[omitted]; ok {
			t.Fatalf("failed ETL carrier included %q", omitted)
		}
	}
	if _, ok := payload["error"]; ok {
		t.Fatal("failed ETL envelope included error")
	}
	for _, forbidden := range []string{
		"failed-rate-limit-token-must-not-escape",
		"runtime-subject-must-not-escape",
		"provider.example.test",
		"body-must-not-escape",
	} {
		if strings.Contains(human+humanErr+structured+structuredErr, forbidden) {
			t.Fatalf("failed ETL output leaked %q", forbidden)
		}
	}
}

func TestETLRunPreflightFailureDoesNotInventRunCarrier(t *testing.T) {
	a := newFailedRateLimitTestApp(t)
	var stdout, stderr bytes.Buffer
	err := runETL(context.Background(), a, []string{"run", "--connection", "missing", "--stream", "widgets"}, &stdout, true, config.Config{})
	if err == nil {
		t.Fatal("runETL error = nil, want preflight failure")
	}
	if code := writeError(&stdout, &stderr, err, true); code != 1 {
		t.Fatalf("preflight failed ETL exit = %d, want 1", code)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal preflight error JSON: %v\n%s", err, stdout.String())
	}
	var kind string
	if err := json.Unmarshal(payload["kind"], &kind); err != nil {
		t.Fatalf("unmarshal preflight error kind: %v", err)
	}
	if kind != "Error" {
		t.Fatalf("preflight failed ETL kind = %q, want Error", kind)
	}
	if _, ok := payload["run"]; ok {
		t.Fatal("preflight failed ETL unexpectedly emitted a run carrier")
	}
}

func TestETLStatusFailureUsesSafeRateLimitCarrier(t *testing.T) {
	a := newFailedRateLimitTestApp(t)
	structured, structuredErr, structuredCode := runFailedRateLimitETL(t, a, true)
	if structuredCode != 1 || structuredErr != "" {
		t.Fatalf("failed ETL result = code %d stderr %q, want code 1 with empty stderr", structuredCode, structuredErr)
	}
	var failedResult struct {
		Run struct {
			ID string `json:"id"`
		} `json:"run"`
	}
	if err := json.Unmarshal([]byte(structured), &failedResult); err != nil {
		t.Fatalf("unmarshal failed ETL: %v", err)
	}
	if failedResult.Run.ID == "" {
		t.Fatal("failed ETL run ID is empty")
	}

	for _, jsonOut := range []bool{false, true} {
		var stdout bytes.Buffer
		if err := runETL(context.Background(), a, []string{"status", failedResult.Run.ID}, &stdout, jsonOut, config.Config{}); err != nil {
			t.Fatalf("failed ETL status json=%t: %v", jsonOut, err)
		}
		output := stdout.String()
		for _, forbidden := range []string{
			"failed-rate-limit-token-must-not-escape",
			"runtime-subject-must-not-escape",
			"provider.example.test",
			"body-must-not-escape",
		} {
			if strings.Contains(output, forbidden) {
				t.Fatalf("failed ETL status json=%t leaked %q", jsonOut, forbidden)
			}
		}
		if !jsonOut {
			for _, want := range []string{
				"ETL run " + failedResult.Run.ID + " failed: read=0 loaded=0 failed=0",
				"Rate limits: connector=failed-rate-limit declaration=declared",
				"provider_429_observed=1",
				"request_latency=7ms",
			} {
				if !strings.Contains(output, want) {
					t.Fatalf("failed ETL status human output missing %q:\n%s", want, output)
				}
			}
			continue
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal([]byte(output), &payload); err != nil {
			t.Fatalf("unmarshal failed ETL status JSON: %v\n%s", err, output)
		}
		var carrier map[string]json.RawMessage
		if err := json.Unmarshal(payload["run"], &carrier); err != nil {
			t.Fatalf("unmarshal failed ETL status carrier: %v", err)
		}
		if len(carrier) != 3 {
			t.Fatalf("failed ETL status carrier field count = %d, want 3: %s", len(carrier), payload["run"])
		}
		for _, want := range []string{"id", "status", "rate_limit"} {
			if _, ok := carrier[want]; !ok {
				t.Fatalf("failed ETL status carrier missing %q", want)
			}
		}
		for _, omitted := range []string{"error", "checkpoint"} {
			if _, ok := carrier[omitted]; ok {
				t.Fatalf("failed ETL status carrier included %q", omitted)
			}
		}
	}
}

func TestETLRuntimeRecordingFailureReportsCompletedRun(t *testing.T) {
	for _, jsonOut := range []bool{false, true} {
		a := newRuntimeFailureTestApp(t)
		stdout, stderr, code := runRuntimeRecordingFailureETL(t, a, jsonOut)
		if code != 1 {
			t.Fatalf("runtime recording failure json=%t exit = %d, want 1", jsonOut, code)
		}
		if stderr != "" {
			t.Fatalf("runtime recording failure json=%t stderr = %q, want empty", jsonOut, stderr)
		}
		for _, forbidden := range []string{"runtime-password-must-not-escape", "127.0.0.1:1"} {
			if strings.Contains(stdout+stderr, forbidden) {
				t.Fatalf("runtime recording failure json=%t leaked %q", jsonOut, forbidden)
			}
		}
		if !jsonOut {
			for _, want := range []string{
				"ETL run run_",
				"completed: read=3 loaded=3 failed=0 runtime_recorded=false runtime_recording=failed",
				"Rate limits: connector=sample declaration=undeclared",
			} {
				if !strings.Contains(stdout, want) {
					t.Fatalf("runtime recording failure human output missing %q:\n%s", want, stdout)
				}
			}
			continue
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
			t.Fatalf("unmarshal runtime recording failure JSON: %v\n%s", err, stdout)
		}
		var kind, runtimeRecording string
		var runtimeRecorded bool
		if err := json.Unmarshal(payload["kind"], &kind); err != nil {
			t.Fatalf("unmarshal runtime recording kind: %v", err)
		}
		if err := json.Unmarshal(payload["runtime_recorded"], &runtimeRecorded); err != nil {
			t.Fatalf("unmarshal runtime_recorded: %v", err)
		}
		if err := json.Unmarshal(payload["runtime_recording"], &runtimeRecording); err != nil {
			t.Fatalf("unmarshal runtime_recording: %v", err)
		}
		if kind != "ETLRun" || runtimeRecorded || runtimeRecording != "failed" {
			t.Fatalf("runtime recording failure envelope = kind %q recorded %t recording %q", kind, runtimeRecorded, runtimeRecording)
		}
		if _, ok := payload["error"]; ok {
			t.Fatal("runtime recording failure envelope included error")
		}
		var carrier map[string]json.RawMessage
		if err := json.Unmarshal(payload["run"], &carrier); err != nil {
			t.Fatalf("unmarshal completed run carrier: %v", err)
		}
		for _, omitted := range []string{"checkpoint", "error", "connection", "stream"} {
			if _, ok := carrier[omitted]; ok {
				t.Fatalf("runtime recording failure carrier included %q", omitted)
			}
		}
		var run struct {
			ID        string                      `json:"id"`
			Status    string                      `json:"status"`
			RateLimit connectors.RateLimitSummary `json:"rate_limit"`
		}
		if err := json.Unmarshal(payload["run"], &run); err != nil {
			t.Fatalf("unmarshal completed run: %v", err)
		}
		if run.ID == "" || run.Status != "completed" || len(run.RateLimit.Connectors) != 2 {
			t.Fatalf("runtime recording failure run = %+v", run)
		}
	}
}

func TestCompletedETLRunCarrierOmitsUnboundedFields(t *testing.T) {
	run := app.Run{
		ID:                 "run_safe_carrier",
		Type:               "etl",
		Connection:         "https://connection.example.test/body-must-not-escape",
		Stream:             "https://stream.example.test/body-must-not-escape",
		Status:             "completed",
		RecordsRead:        1,
		RecordsTransformed: 1,
		RecordsLoaded:      1,
		Checkpoint: map[string]string{
			"cursor": "https://provider.example.test/body-must-not-escape",
		},
		Error: "runtime-password-must-not-escape",
		RateLimit: connectors.RateLimitSummary{Connectors: []connectors.RateLimitConnectorSummary{{
			Connector:   "sample",
			Declaration: connectors.RateLimitDeclarationUndeclared,
		}}},
		StartedAt:   time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC),
		CompletedAt: time.Date(2026, time.August, 6, 12, 0, 1, 0, time.UTC),
	}
	var stdout bytes.Buffer
	if err := writeCompletedETLRun(&stdout, run, true, false, true); err != nil {
		t.Fatalf("writeCompletedETLRun: %v", err)
	}
	output := stdout.String()
	for _, forbidden := range []string{
		"connection.example.test",
		"stream.example.test",
		"provider.example.test",
		"body-must-not-escape",
		"runtime-password-must-not-escape",
	} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("completed ETL carrier leaked %q", forbidden)
		}
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("unmarshal completed ETL carrier: %v\n%s", err, output)
	}
	var carrier map[string]json.RawMessage
	if err := json.Unmarshal(payload["run"], &carrier); err != nil {
		t.Fatalf("unmarshal completed ETL run: %v", err)
	}
	for _, omitted := range []string{"checkpoint", "error", "connection", "stream"} {
		if _, ok := carrier[omitted]; ok {
			t.Fatalf("completed ETL carrier included %q", omitted)
		}
	}
	for _, want := range []string{"id", "type", "status", "records_read", "records_transformed", "records_loaded", "records_failed", "rate_limit", "started_at", "completed_at"} {
		if _, ok := carrier[want]; !ok {
			t.Fatalf("completed ETL carrier missing %q", want)
		}
	}
}

func newFailedRateLimitTestApp(t *testing.T) *app.App {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	a.Registry().Register(failedRateLimitConnector{})
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "failed-rate-limit-source",
		Connector: "failed-rate-limit",
		Config:    map[string]string{"account_id": "runtime-subject-must-not-escape"},
		Secrets:   map[string]string{"token": "failed-rate-limit-token-must-not-escape"},
	}); err != nil {
		t.Fatalf("AddCredential source: %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "failed-rate-limit-warehouse",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential warehouse: %v", err)
	}
	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "failed_rate_limit_to_warehouse",
		Source:      app.EndpointConfig{Connector: "failed-rate-limit", Credential: "failed-rate-limit-source"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "failed-rate-limit-warehouse"},
		Streams: map[string]app.StreamConfig{
			"widgets": {SyncMode: "full_refresh_overwrite", DestinationTable: "widgets"},
		},
	}); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	return a
}

func newRuntimeFailureTestApp(t *testing.T) *app.App {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "runtime-failure-source",
		Connector: "sample",
		Secrets:   map[string]string{"token": "runtime-source-token-must-not-escape"},
	}); err != nil {
		t.Fatalf("AddCredential source: %v", err)
	}
	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "runtime-failure-warehouse",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential warehouse: %v", err)
	}
	if _, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "runtime_failure_to_warehouse",
		Source:      app.EndpointConfig{Connector: "sample", Credential: "runtime-failure-source"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "runtime-failure-warehouse"},
		Streams: map[string]app.StreamConfig{
			"customers": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, CursorField: "updated_at", DestinationTable: "runtime_failure_customers"},
		},
	}); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	return a
}

func runFailedRateLimitETL(t *testing.T, a *app.App, jsonOut bool) (stdout, stderr string, code int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	err := runETL(context.Background(), a, []string{"run", "--connection", "failed_rate_limit_to_warehouse", "--stream", "widgets"}, &outBuf, jsonOut, config.Config{})
	if err == nil {
		t.Fatal("runETL error = nil, want failed run")
	}
	code = writeError(&outBuf, &errBuf, err, jsonOut)
	return outBuf.String(), errBuf.String(), code
}

func runRuntimeRecordingFailureETL(t *testing.T, a *app.App, jsonOut bool) (stdout, stderr string, code int) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cfg := config.Config{Runtime: config.RuntimeConfig{
		PostgresURL:   "postgres://runtime-user:runtime-password-must-not-escape@127.0.0.1:1/polymetrics?sslmode=disable",
		DragonflyAddr: "127.0.0.1:1",
		TemporalAddr:  "127.0.0.1:1",
	}}
	err := runETL(context.Background(), a, []string{"run", "--connection", "runtime_failure_to_warehouse", "--stream", "customers", "--runtime"}, &outBuf, jsonOut, cfg)
	if err == nil {
		t.Fatal("runETL error = nil, want runtime recording failure")
	}
	code = writeError(&outBuf, &errBuf, err, jsonOut)
	return outBuf.String(), errBuf.String(), code
}
