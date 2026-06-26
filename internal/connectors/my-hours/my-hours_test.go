package myhours_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	myhours "polymetrics.ai/internal/connectors/my-hours"
)

// newServer builds a fake My Hours API: a tokens/login endpoint that returns a
// bearer token in exchange for credentials, plus the data endpoints. It records
// the Authorization header and api-version header seen on data requests, and the
// DateFrom values seen on the Reports/activity (time_logs) endpoint so the test
// can assert date-window batching.
func newServer(t *testing.T) (*httptest.Server, *seen) {
	t.Helper()
	s := &seen{}
	mux := http.NewServeMux()
	mux.HandleFunc("/tokens/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("login method = %s, want POST", r.Method)
		}
		s.loginCalls++
		_, _ = w.Write([]byte(`{"accessToken":"tok_abc","refreshToken":"ref_xyz","expiresIn":432000}`))
	})
	mux.HandleFunc("/Clients", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		_, _ = w.Write([]byte(`[{"id":1,"name":"Acme","archived":false},{"id":2,"name":"Globex","archived":true}]`))
	})
	mux.HandleFunc("/Users/getAll", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		_, _ = w.Write([]byte(`[{"id":10,"name":"Ada","email":"ada@example.com","active":true}]`))
	})
	mux.HandleFunc("/Reports/activity", func(w http.ResponseWriter, r *http.Request) {
		s.record(r)
		from := r.URL.Query().Get("DateFrom")
		s.dateFroms = append(s.dateFroms, from)
		// Return one distinct log per window so the test can confirm both
		// windows were fetched and merged.
		switch {
		case from == "2026-01-01":
			_, _ = w.Write([]byte(`[{"logId":100,"date":"2026-01-15","projectName":"P1","logDuration":3600}]`))
		default:
			_, _ = w.Write([]byte(`[{"logId":200,"date":"2026-02-15","projectName":"P2","logDuration":7200}]`))
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, s
}

type seen struct {
	loginCalls int
	auth       string
	apiVersion string
	dateFroms  []string
}

func (s *seen) record(r *http.Request) {
	s.auth = r.Header.Get("Authorization")
	s.apiVersion = r.Header.Get("api-version")
}

func cfg(srvURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":   srvURL,
			"email":      "ada@example.com",
			"start_date": "2026-01-01",
		},
		Secrets: map[string]string{"password": "s3cr3t"},
	}
}

// TestReadClientsAuthenticates verifies the login token exchange, the Bearer
// auth header and api-version header on the data request, and record mapping for
// a top-level-array stream.
func TestReadClientsAuthenticates(t *testing.T) {
	srv, s := newServer(t)
	c := myhours.New()

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: cfg(srv.URL)}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if s.loginCalls == 0 {
		t.Fatal("expected a tokens/login exchange")
	}
	if s.auth != "Bearer tok_abc" {
		t.Fatalf("Authorization = %q, want Bearer tok_abc", s.auth)
	}
	if s.apiVersion != "1.0" {
		t.Fatalf("api-version = %q, want 1.0", s.apiVersion)
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2", len(got))
	}
	if got[0]["id"] == nil || got[0]["name"] == nil {
		t.Fatalf("record missing id/name: %+v", got[0])
	}
}

// TestReadTimeLogsBatchesAcrossWindows verifies the time_logs stream pages
// across multiple date windows (logs_batch_size) and merges the records from
// each window — the connector's pagination story.
func TestReadTimeLogsBatchesAcrossWindows(t *testing.T) {
	srv, s := newServer(t)
	c := myhours.New()

	rc := cfg(srv.URL)
	// 31-day batch over a ~2 month range forces at least two windows.
	rc.Config["logs_batch_size"] = "31"
	rc.Config["end_date"] = "2026-02-28"

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "time_logs", Config: rc}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(s.dateFroms) < 2 {
		t.Fatalf("expected >=2 date windows, got DateFroms=%v", s.dateFroms)
	}
	if s.dateFroms[0] != "2026-01-01" {
		t.Fatalf("first DateFrom = %q, want 2026-01-01", s.dateFroms[0])
	}
	if len(got) != 2 {
		t.Fatalf("records = %d, want 2 (one per window)", len(got))
	}
	if got[0]["logId"] == nil || got[0]["log_duration"] == nil {
		t.Fatalf("time log record missing mapped fields: %+v", got[0])
	}
}

// TestFixtureMode confirms the credential-free deterministic path works without
// any network access (mode=fixture), required for conformance.
func TestFixtureMode(t *testing.T) {
	c := myhours.New()
	rc := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: rc}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read fixture: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("fixture records = %d, want 2", len(got))
	}
	if err := c.Check(context.Background(), rc); err != nil {
		t.Fatalf("Check fixture: %v", err)
	}
}

func TestBaseURLValidation(t *testing.T) {
	c := myhours.New()
	rc := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "ftp://evil", "email": "x", "start_date": "2026-01-01"},
		Secrets: map[string]string{"password": "p"},
	}
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "clients", Config: rc}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "base_url") {
		t.Fatalf("expected base_url scheme validation error, got %v", err)
	}
}

func TestMetadataAndRegistry(t *testing.T) {
	c := myhours.New()
	if c.Name() != "my-hours" {
		t.Fatalf("Name() = %q, want my-hours", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Catalog || !caps.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", caps)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) < 5 {
		t.Fatalf("streams = %d, want >=5", len(cat.Streams))
	}
	r := connectors.NewRegistry()
	if _, ok := r.Get("my-hours"); !ok {
		t.Fatal("registry did not resolve my-hours (self-registration)")
	}
}
