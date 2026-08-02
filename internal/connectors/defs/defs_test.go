package defs

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestProductionEmbedLoadsRuntimeBundles(t *testing.T) {
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	if len(bundles) == 0 {
		t.Fatal("LoadAll(FS) returned zero bundles")
	}

	var github *engine.Bundle
	for i := range bundles {
		if bundles[i].Name == "github" {
			github = &bundles[i]
			break
		}
	}
	if github == nil {
		t.Fatal("LoadAll(FS) missing github bundle")
	}
	if github.Metadata.Name != "github" {
		t.Fatalf("github metadata name = %q", github.Metadata.Name)
	}
	if len(github.Streams) == 0 {
		t.Fatal("github bundle has zero streams")
	}
	if github.Docs == "" {
		t.Fatal("github bundle docs are empty")
	}
	if github.Surface != nil {
		t.Fatal("production embed should not include api_surface.json")
	}
	if github.Fixtures != nil {
		t.Fatal("production embed should not include fixtures")
	}
}

func TestAirtableRuntimeBundleSafetyContract(t *testing.T) {
	airtable := loadAirtableBundle(t)
	if airtable.Metadata.Capabilities.Query || airtable.Metadata.Capabilities.CDC {
		t.Fatalf("airtable capabilities query=%v cdc=%v, want both false", airtable.Metadata.Capabilities.Query, airtable.Metadata.Capabilities.CDC)
	}

	valid := []connectors.Record{{"records": []any{"recA1B2C3D4E5F6G7"}}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "delete_multiple_records"}, valid); err != nil {
		t.Fatalf("ValidateWrite valid delete_multiple_records: %v", err)
	}
	unsafe := []connectors.Record{{"records": []any{"recA&records[]=recB"}}}
	err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "delete_multiple_records"}, unsafe)
	if err == nil {
		t.Fatal("ValidateWrite unsafe delete_multiple_records = nil, want error")
	}
	if !strings.Contains(err.Error(), "pattern") {
		t.Fatalf("ValidateWrite unsafe error = %q, want pattern validation", err.Error())
	}

	hyperDBDelete := findAirtableWrite(t, airtable, "hyperdb_delete_records_by_primary_keys")
	if hyperDBDelete.Confirm != "destructive" {
		t.Fatalf("hyperdb_delete_records_by_primary_keys confirm = %q, want destructive", hyperDBDelete.Confirm)
	}
	validHyperDBDelete := []connectors.Record{{"enterprise_account_id": "ent_fixture", "data_table_id": "dtbl_fixture", "primaryKeysForDelete": []any{"pk_fixture"}}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "hyperdb_delete_records_by_primary_keys"}, validHyperDBDelete); err != nil {
		t.Fatalf("ValidateWrite valid hyperdb_delete_records_by_primary_keys: %v", err)
	}
	invalidHyperDBDelete := []connectors.Record{{"enterprise_account_id": "ent_fixture", "data_table_id": "dtbl_fixture", "primaryKeys": []any{"pk_fixture"}}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "hyperdb_delete_records_by_primary_keys"}, invalidHyperDBDelete); err == nil {
		t.Fatal("ValidateWrite legacy primaryKeys hyperdb delete = nil, want schema error")
	}

	noopCases := []struct {
		action string
		record connectors.Record
	}{
		{"update_table", connectors.Record{}},
		{"update_field", connectors.Record{"table_id": "tbl_fixture", "column_id": "fld_fixture"}},
		{"manage_user", connectors.Record{"enterprise_account_id": "ent_fixture", "id": "usr_fixture"}},
		{"update_workspace_restrictions", connectors.Record{"workspace_id": "wsp_fixture"}},
	}
	for _, tt := range noopCases {
		t.Run(tt.action+" rejects no-op", func(t *testing.T) {
			err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: tt.action}, []connectors.Record{tt.record})
			if err == nil {
				t.Fatalf("ValidateWrite %s no-op = nil, want minProperties error", tt.action)
			}
			if !strings.Contains(err.Error(), "minProperties") {
				t.Fatalf("ValidateWrite %s no-op error = %q, want minProperties", tt.action, err.Error())
			}
		})
	}
}

func TestAirtableDeleteUsersByEmailRedactsPathValue(t *testing.T) {
	airtable := loadAirtableBundle(t)
	action := findAirtableWrite(t, airtable, "delete_users_by_email")
	if !containsString(action.RedactFields, "email") {
		t.Fatalf("delete_users_by_email redact_fields = %v, want email", action.RedactFields)
	}

	rawEmail := "sensitive.user@example.invalid"
	encodedEmail := url.QueryEscape(rawEmail)
	record := connectors.Record{"enterprise_account_id": "ent_fixture", "email": rawEmail}
	runtime := connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "fixture-token"}}
	preview, err := engine.DryRunWrite(context.Background(), airtable, connectors.WriteRequest{Action: "delete_users_by_email", Config: runtime}, []connectors.Record{record}, nil)
	if err != nil {
		t.Fatalf("DryRunWrite delete_users_by_email: %v", err)
	}
	warnings := strings.Join(preview.Warnings, " | ")
	for _, leaked := range []string{rawEmail, encodedEmail} {
		if strings.Contains(warnings, leaked) {
			t.Fatalf("DryRunWrite warnings leaked %q in %q", leaked, warnings)
		}
	}
	if !strings.Contains(warnings, "email=redacted") {
		t.Fatalf("DryRunWrite warnings = %q, want redacted email query", warnings)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("email"); got != rawEmail {
			t.Fatalf("email query = %q, want %q", got, rawEmail)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"cannot delete ` + rawEmail + ` or ` + encodedEmail + `"}`))
	}))
	t.Cleanup(srv.Close)
	airtable.HTTP.URL = srv.URL

	result, err := engine.Write(context.Background(), airtable, connectors.WriteRequest{Action: "delete_users_by_email", Config: runtime}, []connectors.Record{record}, nil)
	if err == nil {
		t.Fatal("Write delete_users_by_email = nil, want HTTP error")
	}
	if result.RecordsFailed != 1 || result.RecordsWritten != 0 {
		t.Fatalf("Write result = %+v, want one failed record", result)
	}
	msg := err.Error()
	for _, leaked := range []string{rawEmail, encodedEmail} {
		if strings.Contains(msg, leaked) {
			t.Fatalf("Write error leaked %q in %q", leaked, msg)
		}
	}
	if !strings.Contains(msg, "redacted") {
		t.Fatalf("Write error = %q, want redaction marker", msg)
	}
}

func loadAirtableBundle(t *testing.T) engine.Bundle {
	t.Helper()
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	for i := range bundles {
		if bundles[i].Name == "airtable" {
			return bundles[i]
		}
	}
	t.Fatal("LoadAll(FS) missing airtable bundle")
	return engine.Bundle{}
}

func findAirtableWrite(t *testing.T, bundle engine.Bundle, name string) engine.WriteAction {
	t.Helper()
	for _, action := range bundle.Writes {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("airtable write action %q missing", name)
	return engine.WriteAction{}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestProductionEmbedExcludesConformanceArtifacts(t *testing.T) {
	for _, path := range []string{"github/api_surface.json", "github/fixtures"} {
		if _, err := fs.Stat(FS, path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("fs.Stat(%q) err = %v, want fs.ErrNotExist", path, err)
		}
	}
}
