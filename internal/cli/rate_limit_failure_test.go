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
