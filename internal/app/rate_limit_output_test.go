package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("RateLimit-Remaining", "99")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	instance, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	bundle, err := engine.Load(os.DirFS(filepath.Join("..", "connectors", "engine", "testdata", "rate-limit-enforcement")), "paced")
	if err != nil {
		t.Fatalf("Load declared rate-limit test bundle: %v", err)
	}
	instance.Registry().Register(engine.New(bundle, nil))
	ctx := context.Background()
	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "paced-rate-limit",
		Connector: "paced",
		Config: map[string]string{
			"base_url":   server.URL,
			"account_id": "test-account",
			"tier":       "pro",
			"auth_type":  "oauth",
		},
	}); err != nil {
		t.Fatalf("AddCredential paced: %v", err)
	}
	if _, err := instance.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-rate-limit",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("AddCredential warehouse: %v", err)
	}
	if _, err := instance.CreateConnection(ctx, app.CreateConnectionRequest{
		Name:        "paced_to_warehouse",
		Source:      app.EndpointConfig{Connector: "paced", Credential: "paced-rate-limit"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-rate-limit"},
		Streams: map[string]app.StreamConfig{
			"widgets": {SyncMode: "full_refresh_overwrite", DestinationTable: "widgets"},
		},
	}); err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}

	run, err := instance.RunETL(ctx, app.RunETLRequest{Connection: "paced_to_warehouse", Stream: "widgets"})
	if err != nil {
		t.Fatalf("RunETL: %v", err)
	}
	if len(run.RateLimit.Connectors) != 2 {
		t.Fatalf("rate-limit connector count = %d, want source and destination", len(run.RateLimit.Connectors))
	}
	var paced connectors.RateLimitConnectorSummary
	for _, summary := range run.RateLimit.Connectors {
		if summary.Connector == "paced" {
			paced = summary
		}
	}
	if paced.Declaration != connectors.RateLimitDeclarationDeclared || paced.RequestCount != 1 {
		t.Fatalf("paced rate-limit summary = %+v", paced)
	}
	if paced.RequestLatencyMS <= 0 || paced.PacingWaitMS != 0 || paced.ProviderWaitMS != 0 || paced.Provider429Observed != 0 {
		t.Fatalf("ordinary request latency was not separate from rate-limit waits: %+v", paced)
	}
	if len(paced.Policies) != 1 || paced.Policies[0].ID != "widgets-points" || paced.Policies[0].SelectionReason != "endpoint+tier+auth_type" {
		t.Fatalf("selected policy summary = %+v", paced.Policies)
	}
	if paced.Policies[0].ProviderRemaining == nil || *paced.Policies[0].ProviderRemaining != 99 {
		t.Fatalf("provider remaining = %v, want 99", paced.Policies[0].ProviderRemaining)
	}
	human := strings.Join(run.RateLimit.HumanLines(), "\n")
	for _, want := range []string{
		"connector=paced declaration=declared",
		"widgets-points,subject_kind=account,selected_by=endpoint+tier+auth_type,provider_remaining=99",
		"local_pacing_wait=0ms",
		"provider_429_observed=0",
		"request_latency=",
	} {
		if !strings.Contains(human, want) {
			t.Fatalf("human rate-limit summary missing %q:\n%s", want, human)
		}
	}
}
