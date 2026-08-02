package defs

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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

	for _, stream := range []string{"scim_groups", "scim_users"} {
		if hasAirtableStream(airtable, stream) {
			t.Fatalf("airtable stream %q should stay blocked until SCIM startIndex pagination is enforceable", stream)
		}
	}
	if hasAirtableStream(airtable, "enterprise_users") {
		t.Fatal("airtable stream enterprise_users should stay blocked until required id/email query filtering is enforceable")
	}
	webhookPayloads := findAirtableStream(t, airtable, "webhook_payloads")
	if webhookPayloads.Pagination == nil || webhookPayloads.Pagination.StopPath != "mightHaveMore" {
		t.Fatalf("webhook_payloads pagination = %+v, want stop_path mightHaveMore", webhookPayloads.Pagination)
	}

	blockedArrayActions := []string{
		"add_base_collaborator",
		"add_interface_collaborator",
		"add_workspace_collaborator",
		"create_base",
		"create_records",
		"create_scim_group",
		"create_scim_user",
		"create_table",
		"delete_multiple_records",
		"grant_admin_access",
		"hyperdb_delete_records_by_primary_keys",
		"hyperdb_upsert_records_by_primary_keys",
		"manage_user_batched",
		"manage_user_membership",
		"move_user_groups",
		"move_workspaces",
		"patch_scim_group",
		"patch_scim_user",
		"put_scim_group",
		"put_scim_user",
		"revoke_admin_access",
		"revoke_enterprise_personal_access_tokens",
		"update_multiple_records",
		"update_multiple_records_put",
		"update_workspace_ai_allowlist",
	}
	for _, action := range blockedArrayActions {
		if hasAirtableWrite(airtable, action) {
			t.Fatalf("airtable write %q should stay blocked until non-empty arrays are enforceable", action)
		}
	}

	validCreateField := []connectors.Record{{"table_id": "tbl_fixture", "name": "Fixture Field", "type": "singleLineText"}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_field"}, validCreateField); err != nil {
		t.Fatalf("ValidateWrite valid create_field: %v", err)
	}
	invalidCreateField := []connectors.Record{{"table_id": "tbl_fixture", "name": "Fixture Field", "type": "unsupportedFixtureType"}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_field"}, invalidCreateField); err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("ValidateWrite unsupported create_field type error = %v, want enum validation", err)
	}
	checkboxCreateField := []connectors.Record{{"table_id": "tbl_fixture", "name": "Visited", "type": "checkbox"}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_field"}, checkboxCreateField); err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("ValidateWrite checkbox create_field error = %v, want variant block enum validation", err)
	}

	validAuditLogRequest := []connectors.Record{{"enterprise_account_id": "ent00000000000000", "timePeriod": "2021-01-31"}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_audit_log_request"}, validAuditLogRequest); err != nil {
		t.Fatalf("ValidateWrite valid create_audit_log_request: %v", err)
	}
	invalidAuditLogRequest := []connectors.Record{{"enterprise_account_id": "ent00000000000000", "timePeriod": "2021/01"}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_audit_log_request"}, invalidAuditLogRequest); err == nil || !strings.Contains(err.Error(), "pattern") {
		t.Fatalf("ValidateWrite invalid create_audit_log_request timePeriod error = %v, want pattern validation", err)
	}

	validWebhookIncludeAll := airtableWebhookIncludeRecord("all")
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_webhook"}, validWebhookIncludeAll); err != nil {
		t.Fatalf("ValidateWrite valid create_webhook include all: %v", err)
	}
	validWebhookIncludeArray := airtableWebhookIncludeRecord([]any{"fld_fixture"})
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_webhook"}, validWebhookIncludeArray); err != nil {
		t.Fatalf("ValidateWrite valid create_webhook include array: %v", err)
	}
	invalidWebhookInclude := airtableWebhookIncludeRecord("everything")
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "create_webhook"}, invalidWebhookInclude); err == nil || !strings.Contains(err.Error(), "pattern") {
		t.Fatalf("ValidateWrite invalid create_webhook include string error = %v, want pattern validation", err)
	}

	invalidWorkspacePermission := []connectors.Record{{"workspace_id": "wsp_fixture", "user_or_group_id": "usr_fixture", "permissionLevel": "admin"}}
	if err := engine.ValidateWrite(context.Background(), airtable, connectors.WriteRequest{Action: "update_workspace_collaborator"}, invalidWorkspacePermission); err == nil || !strings.Contains(err.Error(), "enum") {
		t.Fatalf("ValidateWrite invalid update_workspace_collaborator permission error = %v, want enum validation", err)
	}

	noopCases := []struct {
		action string
		record connectors.Record
	}{
		{"update_table", connectors.Record{}},
		{"update_field", connectors.Record{"table_id": "tbl_fixture", "column_id": "fld_fixture"}},
		{"update_record", connectors.Record{"id": "rec_fixture", "fields": map[string]any{}}},
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

func TestAirtableHyperDBDirectReadAllowsDocumentedBodies(t *testing.T) {
	airtable := loadAirtableBundle(t)
	op := findAirtableOperation(t, airtable, "hyperdb_table_read_records")
	if op.REST == nil || op.REST.BodySchema == nil {
		t.Fatalf("hyperdb_table_read_records REST/body_schema = %+v", op.REST)
	}
	if !containsString(op.AuthScopes, "hyperDB.records:read") || containsString(op.AuthScopes, "data.records:read") {
		t.Fatalf("hyperdb_table_read_records auth scopes = %v, want hyperDB.records:read only", op.AuthScopes)
	}
	var bodySchema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(op.REST.BodySchema, &bodySchema); err != nil {
		t.Fatalf("decode hyperdb body schema: %v", err)
	}
	if containsString(bodySchema.Required, "primaryKeys") {
		t.Fatalf("hyperdb body schema required = %v, want optional primaryKeys", bodySchema.Required)
	}

	cmd := findAirtableCommand(t, airtable, "hyperdb get-records")
	primaryKeyFlag := findAirtableCommandFlag(t, cmd, "primary-key")
	if primaryKeyFlag.Required {
		t.Fatal("hyperdb get-records --primary-key should be optional")
	}

	tests := []struct {
		name string
		body map[string]any
		want map[string]any
	}{
		{
			name: "empty body",
			body: map[string]any{},
			want: map[string]any{},
		},
		{
			name: "cursor only",
			body: map[string]any{"cursor": "itr_fixture_2"},
			want: map[string]any{"cursor": "itr_fixture_2"},
		},
		{
			name: "primary keys",
			body: map[string]any{"primaryKeys": []any{"pk_fixture_1"}},
			want: map[string]any{"primaryKeys": []any{"pk_fixture_1"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sawBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want POST", r.Method)
				}
				if r.URL.Path != "/v0/ent_fixture/dtbl_fixture/getRecords" {
					t.Fatalf("path = %s, want /v0/ent_fixture/dtbl_fixture/getRecords", r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&sawBody); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"records":[]}`))
			}))
			t.Cleanup(srv.Close)

			bundle := airtable
			bundle.HTTP.URL = srv.URL
			_, err := engine.OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation: "hyperdb_table_read_records",
				Config: connectors.RuntimeConfig{
					Secrets: map[string]string{"api_key": "fixture-token"},
				},
				PathParams: map[string]string{
					"enterpriseAccountId": "ent_fixture",
					"dataTableId":         "dtbl_fixture",
				},
				Body:         tt.body,
				MaxBytes:     1024,
				OutputPolicy: "json_redacted",
			}, nil)
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			if !reflect.DeepEqual(sawBody, tt.want) {
				t.Fatalf("request body = %+v, want %+v", sawBody, tt.want)
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

func airtableWebhookIncludeRecord(include any) []connectors.Record {
	return []connectors.Record{{
		"specification": map[string]any{
			"options": map[string]any{
				"filters":  map[string]any{"dataTypes": []any{"tableData"}},
				"includes": map[string]any{"includeCellValuesInFieldIds": include},
			},
		},
	}}
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

func findAirtableOperation(t *testing.T, bundle engine.Bundle, id string) engine.OperationSpec {
	t.Helper()
	for _, op := range bundle.Operations {
		if op.ID == id {
			return op
		}
	}
	t.Fatalf("airtable operation %q missing", id)
	return engine.OperationSpec{}
}

func findAirtableCommand(t *testing.T, bundle engine.Bundle, path string) engine.CLICommand {
	t.Helper()
	if bundle.CLISurface == nil {
		t.Fatal("airtable cli_surface missing")
	}
	for _, cmd := range bundle.CLISurface.Commands {
		if cmd.Path == path {
			return cmd
		}
	}
	t.Fatalf("airtable command %q missing", path)
	return engine.CLICommand{}
}

func findAirtableCommandFlag(t *testing.T, cmd engine.CLICommand, name string) engine.CLIFlag {
	t.Helper()
	for _, flag := range cmd.Flags {
		if flag.Name == name {
			return flag
		}
	}
	t.Fatalf("airtable command %q flag %q missing", cmd.Path, name)
	return engine.CLIFlag{}
}

func hasAirtableWrite(bundle engine.Bundle, name string) bool {
	for _, action := range bundle.Writes {
		if action.Name == name {
			return true
		}
	}
	return false
}

func findAirtableStream(t *testing.T, bundle engine.Bundle, name string) engine.StreamSpec {
	t.Helper()
	for _, stream := range bundle.Streams {
		if stream.Name == name {
			return stream
		}
	}
	t.Fatalf("airtable stream %q missing", name)
	return engine.StreamSpec{}
}

func hasAirtableStream(bundle engine.Bundle, name string) bool {
	for _, stream := range bundle.Streams {
		if stream.Name == name {
			return true
		}
	}
	return false
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
