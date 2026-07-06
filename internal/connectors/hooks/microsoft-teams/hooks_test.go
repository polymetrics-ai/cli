package microsoftteams_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	microsoftteams "polymetrics.ai/internal/connectors/hooks/microsoft-teams"
)

// newTestRuntime builds a minimal *engine.Runtime pointed at srv with a
// Bundle whose Schemas[streamName] projects exactly props -- enough for
// ReadStream to run without needing a full loaded bundle (mirrors
// hooks/sentry/hooks_test.go's helper).
func newTestRuntime(srv *httptest.Server, streamName string, props []string, cfg connectors.RuntimeConfig) *engine.Runtime {
	sch := schemaWithProperties(props)
	b := &engine.Bundle{
		Name:    "microsoft-teams",
		Schemas: map[string]*engine.StreamSchema{streamName: sch},
	}
	return &engine.Runtime{
		Requester: &connsdk.Requester{BaseURL: srv.URL},
		Bundle:    b,
		Config:    cfg,
	}
}

func schemaWithProperties(props []string) *engine.StreamSchema {
	doc := map[string]any{
		"$schema":    "http://json-schema.org/draft-07/schema#",
		"type":       "object",
		"properties": map[string]any{},
	}
	propsMap := doc["properties"].(map[string]any)
	for _, p := range props {
		propsMap[p] = map[string]any{"type": []string{"string", "boolean", "null"}}
	}
	sch, err := engine.CompileSchema(mustMarshal(doc))
	if err != nil {
		panic(err)
	}
	return &engine.StreamSchema{Schema: sch}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// TestReadStream_UsersFollowsNextLinkPagination is the red-first test: the
// hook must follow the literal "@odata.nextLink" key across two pages of
// value[] and map every record via the schema-declared properties.
func TestReadStream_UsersFollowsNextLinkPagination(t *testing.T) {
	var calls int
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("$skiptoken") {
		case "":
			next := srv.URL + "/users?$skiptoken=page2"
			_, _ = w.Write([]byte(`{"value":[{"id":"u1","displayName":"Ada","userPrincipalName":"ada@x.com"},{"id":"u2","displayName":"Grace","userPrincipalName":"grace@x.com"}],"@odata.nextLink":"` + next + `"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"value":[{"id":"u3","displayName":"Kay","userPrincipalName":"kay@x.com"}]}`))
		default:
			t.Errorf("unexpected $skiptoken=%q", r.URL.Query().Get("$skiptoken"))
			_, _ = w.Write([]byte(`{"value":[]}`))
		}
	})

	h := microsoftteams.Hooks{}
	rt := newTestRuntime(srv, "users", []string{"id", "display_name", "user_principal_name"}, connectors.RuntimeConfig{})

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "users"}, connectors.ReadRequest{Stream: "users"}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream returned handled=false, want true")
	}
	if calls != 2 {
		t.Fatalf("graph calls = %d, want 2 (nextLink pagination)", calls)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 across 2 pages", len(got))
	}
	for _, rec := range got {
		if rec["id"] == nil || rec["display_name"] == nil {
			t.Fatalf("record missing mapped id/display_name: %+v", rec)
		}
	}
}

// TestReadStream_TeamDeviceUsageReportSendsPeriod verifies the
// team_device_usage_report stream sends period + $format, honoring a
// config override.
func TestReadStream_TeamDeviceUsageReportSendsPeriod(t *testing.T) {
	var gotPeriod, gotFormat string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPeriod = r.URL.Query().Get("period")
		gotFormat = r.URL.Query().Get("$format")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":[{"id":"d1","userPrincipalName":"ada@x.com","isDeleted":false}]}`))
	}))
	defer srv.Close()

	h := microsoftteams.Hooks{}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"period": "D30"}}
	rt := newTestRuntime(srv, "team_device_usage_report", []string{"id", "user_principal_name", "is_deleted"}, cfg)

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "team_device_usage_report"}, connectors.ReadRequest{Stream: "team_device_usage_report", Config: cfg}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream returned handled=false, want true")
	}
	if gotPeriod != "D30" {
		t.Fatalf("period = %q, want D30", gotPeriod)
	}
	if gotFormat != "application/json" {
		t.Fatalf("$format = %q, want application/json", gotFormat)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestReadStream_MaxPagesCapsRequests verifies the config-driven max_pages
// cap stops pagination early even when the server still offers a next
// link.
func TestReadStream_MaxPagesCapsRequests(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		next := r.URL.Scheme + "://" + r.Host + "/groups?page=" + r.URL.Query().Get("$skiptoken") + "x"
		_, _ = w.Write([]byte(`{"value":[{"id":"g1","displayName":"Group"}],"@odata.nextLink":"` + next + `"}`))
	}))
	defer srv.Close()

	h := microsoftteams.Hooks{}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"max_pages": "1"}}
	rt := newTestRuntime(srv, "groups", []string{"id", "display_name"}, cfg)

	var got []connectors.Record
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "groups"}, connectors.ReadRequest{Stream: "groups", Config: cfg}, rt, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if !handled {
		t.Fatal("ReadStream returned handled=false, want true")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (max_pages=1 cap)", calls)
	}
	if len(got) != 1 {
		t.Fatalf("records = %d, want 1", len(got))
	}
}

// TestReadStream_UnknownStreamFallsBack verifies an unrecognized stream
// name returns handled=false rather than erroring, keeping the declarative
// path an honest fallback.
func TestReadStream_UnknownStreamFallsBack(t *testing.T) {
	h := microsoftteams.Hooks{}
	rt := &engine.Runtime{Bundle: &engine.Bundle{Name: "microsoft-teams"}}
	handled, err := h.ReadStream(context.Background(), engine.StreamSpec{Name: "nonexistent"}, connectors.ReadRequest{Stream: "nonexistent"}, rt, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream: %v", err)
	}
	if handled {
		t.Fatal("ReadStream returned handled=true for an unknown stream, want false")
	}
}
