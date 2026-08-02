package ashby

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "ashby")
}

func assertConnectorContract(t *testing.T, c connectors.Connector, wantName string) {
	t.Helper()
	if c == nil {
		t.Fatal("New() = nil")
	}
	if got := c.Name(); got != wantName {
		t.Fatalf("Name() = %q, want %q", got, wantName)
	}
	meta := c.Metadata()
	if meta.Name != wantName {
		t.Fatalf("Metadata().Name = %q, want %q", meta.Name, wantName)
	}
	caps := meta.Capabilities
	if !caps.Check || !caps.Catalog || !caps.Read {
		t.Fatalf("capabilities = %+v, want Check, Catalog, and Read", caps)
	}
	if !caps.Write {
		t.Fatalf("%s must expose typed reverse-ETL write capability", wantName)
	}
	if _, ok := c.(connectors.WriteValidator); !ok {
		t.Fatalf("%s must validate typed write records", wantName)
	}
	if _, ok := c.(connectors.DryRunWriter); !ok {
		t.Fatalf("%s must dry-run typed write records", wantName)
	}
	if _, ok := c.(connectors.OperationDirectReader); !ok {
		t.Fatalf("%s must expose bounded operation direct reads", wantName)
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if cat.Connector != wantName {
		t.Fatalf("Catalog().Connector = %q, want %q", cat.Connector, wantName)
	}
	if len(cat.Streams) < 70 {
		t.Fatalf("Catalog returned %d streams, want Ashby parity stream coverage", len(cat.Streams))
	}
}

func TestValidateWriteAndDryRun(t *testing.T) {
	c := New()
	validator := c.(connectors.WriteValidator)
	dryRunner := c.(connectors.DryRunWriter)
	req := connectors.WriteRequest{Action: "add_candidate_tag", Config: connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://api.ashbyhq.com"}}}
	records := []connectors.Record{{"candidateId": "candidate_fixture", "tagId": "tag_fixture"}}
	if err := validator.ValidateWrite(context.Background(), req, records); err != nil {
		t.Fatalf("ValidateWrite: %v", err)
	}
	preview, err := dryRunner.DryRunWrite(context.Background(), req, records)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	if preview.RecordsStaged != 1 || preview.Action != "add_candidate_tag" {
		t.Fatalf("preview = %+v, want one staged add_candidate_tag", preview)
	}
}

func TestValidateWriteAcceptsCustomFieldValueUnion(t *testing.T) {
	validator := New().(connectors.WriteValidator)
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://api.ashbyhq.com"}}
	tests := []struct {
		name   string
		action string
		record connectors.Record
	}{
		{
			name:   "single string",
			action: "set_custom_field_value",
			record: connectors.Record{"objectId": "candidate_fixture", "objectType": "Candidate", "fieldId": "field_fixture", "fieldValue": "text"},
		},
		{
			name:   "single number",
			action: "set_custom_field_value",
			record: connectors.Record{"objectId": "candidate_fixture", "objectType": "Candidate", "fieldId": "field_fixture", "fieldValue": 12.5},
		},
		{
			name:   "single boolean",
			action: "set_custom_field_value",
			record: connectors.Record{"objectId": "candidate_fixture", "objectType": "Candidate", "fieldId": "field_fixture", "fieldValue": true},
		},
		{
			name:   "single array",
			action: "set_custom_field_value",
			record: connectors.Record{"objectId": "candidate_fixture", "objectType": "Candidate", "fieldId": "field_fixture", "fieldValue": []any{"A", "B"}},
		},
		{
			name:   "single object",
			action: "set_custom_field_value",
			record: connectors.Record{"objectId": "candidate_fixture", "objectType": "Candidate", "fieldId": "field_fixture", "fieldValue": map[string]any{"country": "USA", "city": "San Francisco"}},
		},
		{
			name:   "single null",
			action: "set_custom_field_value",
			record: connectors.Record{"objectId": "candidate_fixture", "objectType": "Candidate", "fieldId": "field_fixture", "fieldValue": nil},
		},
		{
			name:   "plural string",
			action: "set_custom_field_values",
			record: connectors.Record{"objectId": "candidate_fixture", "objectType": "Candidate", "values": []any{map[string]any{"fieldId": "field_fixture", "fieldValue": "text"}}},
		},
		{
			name:   "user single string",
			action: "set_user_custom_field_value",
			record: connectors.Record{"userId": "user_fixture", "fieldId": "field_fixture", "fieldValue": "text"},
		},
		{
			name:   "user plural string",
			action: "set_user_custom_field_values",
			record: connectors.Record{"userId": "user_fixture", "values": []any{map[string]any{"fieldId": "field_fixture", "fieldValue": "text"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := connectors.WriteRequest{Action: tt.action, Config: cfg}
			if err := validator.ValidateWrite(context.Background(), req, []connectors.Record{tt.record}); err != nil {
				t.Fatalf("ValidateWrite(%s): %v", tt.action, err)
			}
		})
	}
}

func TestCustomFieldValueCommandsArePartial(t *testing.T) {
	surface := New().(connectors.CommandSurfaceProvider).CommandSurface()
	want := map[string]bool{
		"set_custom_field_value":       false,
		"set_custom_field_values":      false,
		"set_user_custom_field_value":  false,
		"set_user_custom_field_values": false,
	}
	for _, cmd := range surface.Commands {
		if _, ok := want[cmd.Write]; !ok {
			continue
		}
		if cmd.Availability != "partial" {
			t.Fatalf("command %q availability = %q, want partial", cmd.Path, cmd.Availability)
		}
		if !strings.Contains(cmd.Notes, "fieldValue union") {
			t.Fatalf("command %q notes = %q, want fieldValue union note", cmd.Path, cmd.Notes)
		}
		for _, flag := range cmd.Flags {
			if strings.Contains(flag.MapsTo, "fieldValue") {
				t.Fatalf("command %q flag --%s maps to fieldValue despite partial union coverage", cmd.Path, flag.Name)
			}
		}
		want[cmd.Write] = true
	}
	for write, found := range want {
		if !found {
			t.Fatalf("custom field command for write %q not found", write)
		}
	}
}

func TestValidateWriteRequiresUploadHandles(t *testing.T) {
	validator := New().(connectors.WriteValidator)
	cfg := connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://api.ashbyhq.com"}}
	tests := []struct {
		name        string
		action      string
		handleField string
	}{
		{name: "resume", action: "upload_candidate_resume", handleField: "resumeHandle"},
		{name: "file", action: "upload_candidate_file", handleField: "fileHandle"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := connectors.WriteRequest{Action: tt.action, Config: cfg}
			missing := connectors.Record{"candidateId": "candidate_fixture"}
			if err := validator.ValidateWrite(context.Background(), req, []connectors.Record{missing}); err == nil {
				t.Fatalf("ValidateWrite(%s) without %s returned nil", tt.action, tt.handleField)
			}
			valid := connectors.Record{"candidateId": "candidate_fixture", tt.handleField: "handle_fixture"}
			if err := validator.ValidateWrite(context.Background(), req, []connectors.Record{valid}); err != nil {
				t.Fatalf("ValidateWrite(%s) with %s: %v", tt.action, tt.handleField, err)
			}
		})
	}
}

func TestReadOmitsLimitWhenUndocumented(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/apiKey.info" {
			t.Errorf("path = %q, want /apiKey.info", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if _, ok := body["limit"]; ok {
			t.Errorf("body[limit] = %v, want omitted for apiKey.info", body["limit"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"results":{"title":"fixture","createdAt":"2026-01-01T00:00:00Z","scopes":[]},"moreDataAvailable":false}`))
	}))
	defer server.Close()

	var records []connectors.Record
	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "api_key_info",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL, "max_pages": "1"},
			Secrets: map[string]string{"api_key": "test_key"},
		},
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("emitted %d records = %+v, want 1", len(records), records)
	}
}

func TestReadDefaultsToOnePage(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/candidate.list" {
			t.Errorf("path = %q, want /candidate.list", r.URL.Path)
		}
		requestCount++
		if requestCount > 1 {
			t.Errorf("unexpected request %d with default max_pages", requestCount)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"first","updatedAt":"2026-01-03T00:00:00Z"}],"moreDataAvailable":true,"nextCursor":"opaque-page-2"}`))
	}))
	defer server.Close()

	var records []connectors.Record
	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "candidates",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL},
			Secrets: map[string]string{"api_key": "test_key"},
		},
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("request count = %d, want 1", requestCount)
	}
	if len(records) != 1 {
		t.Fatalf("emitted %d records = %+v, want 1", len(records), records)
	}
}

func TestReadUsesStateAsLowerBoundNotPageCursor(t *testing.T) {
	var requestBodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/candidate.list" {
			t.Errorf("path = %q, want /candidate.list", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		requestBodies = append(requestBodies, body)
		w.Header().Set("Content-Type", "application/json")
		switch len(requestBodies) {
		case 1:
			if _, ok := body["cursor"]; ok {
				t.Errorf("first request cursor = %v, want no Ashby page cursor from saved state", body["cursor"])
			}
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"old","updatedAt":"2026-01-01T00:00:00Z"},{"id":"new","updatedAt":"2026-01-03T00:00:00Z"}],"moreDataAvailable":true,"nextCursor":"opaque-page-2"}`))
		case 2:
			if got := body["cursor"]; got != "opaque-page-2" {
				t.Errorf("second request cursor = %v, want opaque-page-2", got)
			}
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"next","updatedAt":"2026-01-04T00:00:00Z"}],"moreDataAvailable":false}`))
		default:
			t.Errorf("unexpected request body %d: %+v", len(requestBodies), body)
			_, _ = w.Write([]byte(`{"success":true,"results":[],"moreDataAvailable":false}`))
		}
	}))
	defer server.Close()

	var records []connectors.Record
	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "candidates",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL, "max_pages": "2"},
			Secrets: map[string]string{"api_key": "test_key"},
		},
		State: map[string]string{"cursor": "2026-01-02T00:00:00Z"},
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(requestBodies) != 2 {
		t.Fatalf("request count = %d, want 2", len(requestBodies))
	}
	wantIDs := []string{"new", "next"}
	if len(records) != len(wantIDs) {
		t.Fatalf("emitted %d records = %+v, want ids %v", len(records), records, wantIDs)
	}
	for i, wantID := range wantIDs {
		if got := records[i]["id"]; got != wantID {
			t.Fatalf("record %d id = %v, want %s", i, got, wantID)
		}
	}
}

func TestOperationDirectReadUsesFixedSearchPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/candidate.search" {
			t.Fatalf("path = %q, want /candidate.search", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "test_key" || pass != "" {
			t.Fatalf("basic auth = (%q,%q,%v), want Ashby key as username with blank password", user, pass, ok)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["email"] != "candidate@example.invalid" {
			t.Fatalf("body[email] = %v, want candidate@example.invalid", body["email"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"results":[]}`))
	}))
	defer server.Close()

	reader := New().(connectors.OperationDirectReader)
	result, err := reader.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "ashby.direct.candidate.search",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL},
			Secrets: map[string]string{"api_key": "test_key"},
		},
		Body:         map[string]any{"email": "candidate@example.invalid"},
		OutputPolicy: "json_redacted",
		MaxBytes:     1 << 20,
	})
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.Status)
	}
}
