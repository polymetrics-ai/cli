package gitlab

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
)

// The retained source lock declares gitlab.com as the provider origin. The
// witness intentionally keeps that production origin and redirects only the
// TLS dial to a local double, so config.base_url never becomes a test-only raw
// route escape hatch.
const gitLabMLflowProviderBaseURL = "https://gitlab.com/api/v4"

const gitLabMLflowFixtureToken = "fixture-gitlab-mlflow-token"

func TestGitLabMLflowMetricsHistoryETLMaterializesThroughDuckDB(t *testing.T) {
	fixture := newGitLabMLflowETLFixture(t, false)
	application := fixture.openApplication(t)
	connection := fixture.createWarehouseConnection(t, application, "gitlab_mlflow_history_to_warehouse", "gitlab_mlflow_metrics_history")

	run, err := application.RunETL(context.Background(), app.RunETLRequest{
		Connection: connection.Name,
		Stream:     "mlflow_metrics_history",
		BatchSize:  1,
	})
	if err != nil {
		t.Fatalf("run source-bound GitLab MLflow ETL: %v", err)
	}
	if run.Status != "completed" || run.RecordsRead != 2 || run.RecordsLoaded != 2 {
		t.Fatalf("GitLab MLflow ETL run = %+v, want completed two-record DuckDB materialization", run)
	}
	fixture.assertTwoPageSourceRoute(t)

	rows, err := application.QueryTable(context.Background(), app.QueryTableRequest{
		Connection: connection.Name,
		Table:      "gitlab_mlflow_metrics_history",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("query GitLab MLflow DuckDB table: %v", err)
	}
	if len(rows) != 2 || rows[0]["key"] != "accuracy" || rows[1]["key"] != "accuracy" {
		t.Fatalf("GitLab MLflow DuckDB rows = %#v, want two source metrics", rows)
	}
}

func TestGitLabMLflowMetricsHistoryETLNonSuccessDoesNotMaterialize(t *testing.T) {
	fixture := newGitLabMLflowETLFixture(t, true)
	application := fixture.openApplication(t)
	connection := fixture.createWarehouseConnection(t, application, "gitlab_mlflow_history_failure", "gitlab_mlflow_metrics_history_failure")

	run, err := application.RunETL(context.Background(), app.RunETLRequest{
		Connection: connection.Name,
		Stream:     "mlflow_metrics_history",
		BatchSize:  1,
	})
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("GitLab MLflow 502 run error = %v, want source non-success before materialization", err)
	}
	if run.Status != "failed" || run.RecordsRead != 0 || run.RecordsLoaded != 0 {
		t.Fatalf("GitLab MLflow 502 run = %+v, want failed zero-record result", run)
	}
	fixture.assertFailureSourceRoute(t)

	rows, queryErr := application.QueryTable(context.Background(), app.QueryTableRequest{
		Connection: connection.Name,
		Table:      "gitlab_mlflow_metrics_history_failure",
		Limit:      10,
	})
	if queryErr == nil || len(rows) != 0 {
		t.Fatalf("failed GitLab MLflow ETL materialized rows=%#v err=%v, want no warehouse result", rows, queryErr)
	}
}

func TestGitLabMLflowMetricsHistoryETLPreflightStopsBeforeProviderIO(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("initialize GitLab MLflow preflight project: %v", err)
	}
	spy := &gitLabMLflowNoNetworkTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = spy
	t.Cleanup(func() { http.DefaultTransport = previous })

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"gitlab", "mlflow-metrics-history", "list",
		"--project-id", "42",
		"--run-id", "run-one",
		"--metric-key", "accuracy",
		"--root", root,
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("GitLab MLflow command without credential unexpectedly succeeded: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if got := spy.requests.Load(); got != 0 {
		t.Fatalf("GitLab MLflow credential/preflight failure made %d provider requests, want zero", got)
	}
}

type gitLabMLflowNoNetworkTransport struct {
	requests atomic.Int64
}

func (spy *gitLabMLflowNoNetworkTransport) RoundTrip(*http.Request) (*http.Response, error) {
	spy.requests.Add(1)
	return nil, fmt.Errorf("unexpected GitLab MLflow provider I/O")
}

type gitLabMLflowETLFixture struct {
	t       *testing.T
	badRead bool
	server  *httptest.Server

	mu    sync.Mutex
	calls []gitLabMLflowHTTPCall
}

type gitLabMLflowHTTPCall struct {
	Path  string
	Query string
}

func newGitLabMLflowETLFixture(t *testing.T, badRead bool) *gitLabMLflowETLFixture {
	t.Helper()
	fixture := &gitLabMLflowETLFixture{t: t, badRead: badRead}
	fixture.server = httptest.NewTLSServer(http.HandlerFunc(fixture.serveHTTP))
	t.Cleanup(fixture.server.Close)
	fixture.installSourceBoundOriginRedirect(t)
	return fixture
}

func (f *gitLabMLflowETLFixture) installSourceBoundOriginRedirect(t *testing.T) {
	t.Helper()
	serverTransport, ok := f.server.Client().Transport.(*http.Transport)
	if !ok {
		t.Fatal("GitLab MLflow httptest client does not expose an HTTP transport")
	}
	transport := serverTransport.Clone()
	transport.Proxy = nil
	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	} else {
		transport.TLSClientConfig = transport.TLSClientConfig.Clone()
	}
	transport.TLSClientConfig.InsecureSkipVerify = true //nolint:gosec // local httptest-only dial redirect
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if address == "gitlab.com:443" {
			return (&net.Dialer{}).DialContext(ctx, network, f.server.Listener.Addr().String())
		}
		return nil, fmt.Errorf("GitLab MLflow fixture blocked unexpected outbound dial to %q", address)
	}
	previous := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
}

func (f *gitLabMLflowETLFixture) openApplication(t *testing.T) *app.App {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("initialize GitLab MLflow ETL project: %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("open GitLab MLflow ETL project: %v", err)
	}
	ctx := context.Background()
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "gitlab-mlflow-local",
		Connector: "gitlab",
		Config: map[string]string{
			"base_url":          gitLabMLflowProviderBaseURL,
			"project_id":        "42",
			"mlflow_run_id":     "run-one",
			"mlflow_metric_key": "accuracy",
		},
		Secrets: map[string]string{"access_token": gitLabMLflowFixtureToken},
	}); err != nil {
		t.Fatalf("add GitLab MLflow credential: %v", err)
	}
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatalf("add GitLab MLflow warehouse credential: %v", err)
	}
	return application
}

func (f *gitLabMLflowETLFixture) createWarehouseConnection(t *testing.T, application *app.App, name, table string) app.Connection {
	t.Helper()
	connection, err := application.CreateConnection(context.Background(), app.CreateConnectionRequest{
		Name:        name,
		Source:      app.EndpointConfig{Connector: "gitlab", Credential: "gitlab-mlflow-local"},
		Destination: app.EndpointConfig{Connector: "warehouse", Credential: "warehouse-local"},
		Streams: map[string]app.StreamConfig{"mlflow_metrics_history": {
			SyncMode:         "full_append",
			DestinationTable: table,
		}},
	})
	if err != nil {
		t.Fatalf("create GitLab MLflow full-refresh connection: %v", err)
	}
	return connection
}

func (f *gitLabMLflowETLFixture) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/api/v4/projects/42/ml/mlflow/api/2.0/mlflow/metrics/get-history" {
		f.failf("unexpected GitLab MLflow fixture request %s %s", r.Method, r.URL.String())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if got := r.Header.Get("Authorization"); got != "Bearer "+gitLabMLflowFixtureToken {
		f.failf("GitLab MLflow authorization = %q, want source credential bearer token", got)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	query := r.URL.Query()
	if query.Get("run_id") != "run-one" || query.Get("metric_key") != "accuracy" || query.Get("max_results") != "1000" {
		f.failf("GitLab MLflow query = %q, want exact retained run_id/metric_key/default max_results", r.URL.RawQuery)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	f.mu.Lock()
	f.calls = append(f.calls, gitLabMLflowHTTPCall{Path: r.URL.Path, Query: r.URL.RawQuery})
	f.mu.Unlock()

	if f.badRead {
		f.writeJSON(w, http.StatusBadGateway, map[string]any{"message": "fixture source failure"})
		return
	}
	pageToken := query.Get("page_token")
	switch pageToken {
	case "":
		f.writeJSON(w, http.StatusOK, map[string]any{
			"metrics":         []any{map[string]any{"key": "accuracy", "step": 1, "timestamp": 1000, "value": 0.5}},
			"next_page_token": "page-two",
		})
	case "page-two":
		f.writeJSON(w, http.StatusOK, map[string]any{
			"metrics": []any{map[string]any{"key": "accuracy", "step": 2, "timestamp": 2000, "value": 0.75}},
		})
	default:
		f.failf("GitLab MLflow page_token = %q, want empty or source continuation page-two", pageToken)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (f *gitLabMLflowETLFixture) assertTwoPageSourceRoute(t *testing.T) {
	t.Helper()
	calls := f.snapshotCalls()
	if len(calls) != 2 {
		t.Fatalf("GitLab MLflow provider calls = %#v, want first and page_token continuation requests", calls)
	}
	for index, call := range calls {
		if call.Path != "/api/v4/projects/42/ml/mlflow/api/2.0/mlflow/metrics/get-history" || !strings.Contains(call.Query, "run_id=run-one") || !strings.Contains(call.Query, "metric_key=accuracy") || !strings.Contains(call.Query, "max_results=1000") {
			t.Fatalf("GitLab MLflow provider call %d=%#v, want exact source-bound path/query", index, call)
		}
	}
	if strings.Contains(calls[0].Query, "page_token=") || !strings.Contains(calls[1].Query, "page_token=page-two") {
		t.Fatalf("GitLab MLflow provider pagination calls=%#v, want source page_token -> next_page_token continuation", calls)
	}
}

func (f *gitLabMLflowETLFixture) assertFailureSourceRoute(t *testing.T) {
	t.Helper()
	calls := f.snapshotCalls()
	if len(calls) == 0 {
		t.Fatal("GitLab MLflow 502 made no source request")
	}
	for _, call := range calls {
		if call.Path != "/api/v4/projects/42/ml/mlflow/api/2.0/mlflow/metrics/get-history" || strings.Contains(call.Query, "page_token=") {
			t.Fatalf("GitLab MLflow 502 source calls=%#v, want only initial source requests", calls)
		}
	}
}

func (f *gitLabMLflowETLFixture) snapshotCalls() []gitLabMLflowHTTPCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]gitLabMLflowHTTPCall(nil), f.calls...)
}

func (f *gitLabMLflowETLFixture) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		f.failf("encode GitLab MLflow fixture response: %v", err)
	}
}

func (f *gitLabMLflowETLFixture) failf(format string, args ...any) {
	f.t.Errorf(format, args...)
}
