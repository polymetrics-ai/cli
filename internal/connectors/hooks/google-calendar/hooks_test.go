package googlecalendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

type testNoAuthHooks struct{ Hooks }

func (testNoAuthHooks) Authenticator(context.Context, connectors.RuntimeConfig, engine.AuthSpec) (connsdk.Authenticator, error) {
	return connsdk.AuthFunc(func(context.Context, *http.Request) error { return nil }), nil
}

func TestHooksRegistered(t *testing.T) {
	h := ExplicitFactory()
	if h == nil {
		t.Fatal("registered hooks = nil")
	}
	if h.ConnectorName() != "google-calendar" {
		t.Fatalf("ConnectorName() = %q", h.ConnectorName())
	}
	if _, ok := h.(engine.AuthHook); !ok {
		t.Fatal("hooks do not implement AuthHook")
	}
	if _, ok := h.(engine.CheckHook); ok {
		t.Fatal("hooks should not override declarative Check")
	}
	if _, ok := h.(engine.StreamHook); ok {
		t.Fatal("hooks should not override declarative ReadStream")
	}
}

func TestAuthenticatorRefreshTokenExchange(t *testing.T) {
	var sawForm bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Fatalf("grant_type = %q", r.Form.Get("grant_type"))
		}
		if r.Form.Get("client_id") != "calendar-client" || r.Form.Get("client_secret") != "calendar-client-secret" || r.Form.Get("refresh_token") != "calendar-refresh" {
			t.Fatalf("unexpected OAuth form keys")
		}
		sawForm = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "calendar-access", "expires_in": 3600})
	}))
	defer srv.Close()

	auth, err := Hooks{}.Authenticator(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "live"}}, engine.AuthSpec{TokenURL: srv.URL, ClientID: "calendar-client", ClientSecret: "calendar-client-secret", Token: "calendar-refresh"})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "https://example.test/calendar", nil)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !sawForm {
		t.Fatal("OAuth refresh endpoint was not called")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer calendar-access" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestAuthenticatorLiveSyntheticValuesStillRefreshes(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "calendar-access", "expires_in": 3600})
	}))
	defer srv.Close()

	auth, err := Hooks{}.Authenticator(context.Background(), connectors.RuntimeConfig{}, engine.AuthSpec{TokenURL: srv.URL, ClientID: "synthetic-fixture-secret", ClientSecret: "synthetic-fixture-secret", Token: "synthetic-fixture-secret"})
	if err != nil {
		t.Fatalf("Authenticator: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "https://example.test/calendar", nil)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !called {
		t.Fatal("expected live auth to call refresh endpoint")
	}
	if got := req.Header.Get("Authorization"); got != "Bearer calendar-access" {
		t.Fatalf("Authorization = %q", got)
	}
}

func TestAuthenticatorMissingRefreshTokenErrorIsRedacted(t *testing.T) {
	_, err := Hooks{}.Authenticator(context.Background(), connectors.RuntimeConfig{Config: map[string]string{"mode": "live"}}, engine.AuthSpec{TokenURL: "https://oauth2.googleapis.com/token", ClientID: "calendar-client", ClientSecret: "calendar-client-secret"})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "calendar-client-secret") {
		t.Fatalf("error leaked secret: %v", err)
	}
	if !strings.Contains(err.Error(), "refresh token is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestFreeBusyOperationDirectReadRejectsInvalidTimeBounds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("server should not be called for invalid typed bounds")
	}))
	defer srv.Close()

	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	b.HTTP.URL = srv.URL + "/calendar/v3"
	conn := engine.New(b, testNoAuthHooks{})
	_, err = conn.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "google-calendar.freebusy.query",
		Config:    connectors.RuntimeConfig{ProjectDir: "__polymetrics_execution_fixture__"},
		Body: map[string]any{
			"timeMin": "not-a-date-time",
			"timeMax": "2020-01-02T00:00:00Z",
			"items":   []any{map[string]any{"id": "primary"}},
		},
		MaxBytes: 4096,
	})
	if err == nil {
		t.Fatal("OperationDirectRead accepted invalid timeMin")
	}
}

func TestFreeBusyOperationDirectReadFixtureOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/calendar/v3/freeBusy" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["timeMin"] != "2020-01-01T00:00:00Z" || body["timeMax"] != "2020-01-02T00:00:00Z" {
			t.Fatalf("unexpected time bounds: %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"kind":        "calendar#freeBusy",
			"calendars":   map[string]any{"primary": map[string]any{"busy": []any{}}},
			"accessToken": "provider-sensitive-fixture",
		})
	}))
	defer srv.Close()

	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	b.HTTP.URL = srv.URL + "/calendar/v3"
	conn := engine.New(b, testNoAuthHooks{})
	direct, err := conn.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "google-calendar.freebusy.query",
		Config:    connectors.RuntimeConfig{ProjectDir: "__polymetrics_execution_fixture__"},
		Body: map[string]any{
			"timeMin": "2020-01-01T00:00:00Z",
			"timeMax": "2020-01-02T00:00:00Z",
			"items":   []any{map[string]any{"id": "primary"}},
		},
		MaxBytes: 4096,
	})
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	body, ok := direct.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want object", direct.Body)
	}
	calendars, ok := body["calendars"].(map[string]any)
	if !ok {
		t.Fatalf("calendars = %#v, want object", body["calendars"])
	}
	if _, ok := calendars["primary"]; !ok {
		t.Fatalf("calendars = %#v, want caller-supplied primary key", calendars)
	}
	if body["accessToken"] != "provider-sensitive-fixture" {
		t.Fatalf("ordinary unclassified accessToken was changed: %#v", body)
	}
}

func TestWriteActionProviderConstraints(t *testing.T) {
	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	conn := engine.New(b, Hooks{})
	tests := []struct {
		name        string
		action      string
		record      connectors.Record
		wantErrText string
	}{
		{
			name:   "import requires iCalendar UID",
			action: "import_event",
			record: connectors.Record{
				"calendar_id": "calendar-fixture",
				"summary":     "Fixture event",
				"start":       map[string]any{"dateTime": "2030-01-01T10:00:00Z"},
				"end":         map[string]any{"dateTime": "2030-01-01T11:00:00Z"},
			},
			wantErrText: "iCalUID",
		},
		{
			name:   "ownership transfer rejects false admin access",
			action: "transfer_calendar_ownership",
			record: connectors.Record{
				"calendar_id":      "calendar-fixture",
				"new_data_owner":   "owner@example.invalid",
				"use_admin_access": false,
			},
			wantErrText: "enum",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := conn.ValidateWrite(context.Background(), connectors.WriteRequest{Action: tc.action}, []connectors.Record{tc.record})
			if err == nil {
				t.Fatalf("ValidateWrite(%s) accepted invalid record", tc.action)
			}
			if !strings.Contains(err.Error(), tc.wantErrText) {
				t.Fatalf("ValidateWrite(%s) error = %q, want %q", tc.action, err, tc.wantErrText)
			}
		})
	}

	watchActions := []struct {
		name   string
		action string
		record connectors.Record
	}{
		{
			name:   "ACL watch",
			action: "watch_acl",
			record: connectors.Record{"calendar_id": "calendar-fixture", "id": "channel-fixture", "type": "web_hook"},
		},
		{
			name:   "calendar list watch",
			action: "watch_calendar_list",
			record: connectors.Record{"id": "channel-fixture", "type": "web_hook"},
		},
		{
			name:   "event watch",
			action: "watch_events",
			record: connectors.Record{"calendar_id": "calendar-fixture", "id": "channel-fixture", "type": "web_hook"},
		},
		{
			name:   "settings watch",
			action: "watch_settings",
			record: connectors.Record{"id": "channel-fixture", "type": "web_hook"},
		},
	}
	invalidAddresses := []struct {
		name        string
		address     string
		wantErrText string
	}{
		{name: "non-HTTPS scheme", address: "http://example.invalid/hook", wantErrText: "pattern"},
		{name: "missing host", address: "https://", wantErrText: "pattern"},
		{name: "whitespace in authority", address: "https://not a URL", wantErrText: "pattern"},
		{name: "malformed URI escape", address: "https://example.invalid/%ZZ", wantErrText: "format"},
	}
	recordWithAddress := func(source connectors.Record, address string) connectors.Record {
		record := make(connectors.Record, len(source)+1)
		for key, value := range source {
			record[key] = value
		}
		record["address"] = address
		return record
	}
	for _, action := range watchActions {
		t.Run(action.name, func(t *testing.T) {
			for _, invalid := range invalidAddresses {
				t.Run(invalid.name, func(t *testing.T) {
					err := conn.ValidateWrite(context.Background(), connectors.WriteRequest{Action: action.action}, []connectors.Record{recordWithAddress(action.record, invalid.address)})
					if err == nil {
						t.Fatalf("ValidateWrite(%s) accepted %q", action.action, invalid.address)
					}
					if !strings.Contains(err.Error(), invalid.wantErrText) {
						t.Fatalf("ValidateWrite(%s) error = %q, want %s validation", action.action, err, invalid.wantErrText)
					}
				})
			}
			err := conn.ValidateWrite(context.Background(), connectors.WriteRequest{Action: action.action}, []connectors.Record{recordWithAddress(action.record, "https://example.invalid/calendar-hook")})
			if err != nil {
				t.Fatalf("ValidateWrite(%s) rejected valid HTTPS callback: %v", action.action, err)
			}
		})
	}
}

func TestCommandRedactionsReachWriteActions(t *testing.T) {
	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	actions := make(map[string]engine.WriteAction, len(b.Writes))
	for _, action := range b.Writes {
		actions[action.Name] = action
	}
	for _, command := range b.CLISurface.Commands {
		if command.Write == "" || len(command.RedactFields) == 0 {
			continue
		}
		action, ok := actions[command.Write]
		if !ok {
			t.Fatalf("command %q references missing write action %q", command.Path, command.Write)
		}
		redacted := make(map[string]bool, len(action.RedactFields))
		for _, field := range action.RedactFields {
			redacted[field] = true
		}
		for _, field := range command.RedactFields {
			if !redacted[field] {
				t.Fatalf("command %q redacts %q but write action %q does not", command.Path, field, action.Name)
			}
		}
	}
}

func TestBundleMetadataMatchesExecutableSurface(t *testing.T) {
	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.ReleaseStage != "alpha" {
		t.Fatalf("release stage = %q, want alpha", b.Metadata.ReleaseStage)
	}
	if !b.Metadata.Capabilities.Write {
		t.Fatal("write capability = false with executable writes")
	}
	want := map[string]struct{}{
		"delete_acl_rule": {}, "insert_acl_rule": {}, "patch_acl_rule": {}, "update_acl_rule": {}, "watch_acl": {},
		"delete_calendar_list_entry": {}, "insert_calendar_list_entry": {}, "patch_calendar_list_entry": {}, "update_calendar_list_entry": {}, "watch_calendar_list": {},
		"clear_calendar": {}, "delete_calendar": {}, "insert_calendar": {}, "patch_calendar": {}, "transfer_calendar_ownership": {}, "update_calendar": {},
		"stop_channel": {},
		"delete_event": {}, "import_event": {}, "insert_event": {}, "move_event": {}, "patch_event": {}, "quick_add_event": {}, "update_event": {}, "watch_events": {},
		"watch_settings": {},
	}
	if len(b.Writes) != len(want) {
		t.Fatalf("writes = %d, want %d", len(b.Writes), len(want))
	}
	for _, action := range b.Writes {
		if _, ok := want[action.Name]; !ok {
			t.Fatalf("unexpected write action %q", action.Name)
		}
		delete(want, action.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing write actions: %#v", want)
	}
}

func TestWriteActionsSendDeclaredRecordQueries(t *testing.T) {
	tests := []struct {
		name             string
		action           string
		record           connectors.Record
		requiresApproval bool
		wantPath         string
		wantQuery        map[string]string
	}{
		{
			name:   "transfer calendar ownership",
			action: "transfer_calendar_ownership",
			record: connectors.Record{
				"calendar_id":      "calendar-fixture",
				"new_data_owner":   "owner@example.invalid",
				"use_admin_access": true,
			},
			requiresApproval: true,
			wantPath:         "/calendar/v3/calendars/calendar-fixture/transferOwnership",
			wantQuery: map[string]string{
				"newDataOwner":   "owner@example.invalid",
				"useAdminAccess": "true",
			},
		},
		{
			name:   "move event",
			action: "move_event",
			record: connectors.Record{
				"calendar_id": "calendar-fixture",
				"event_id":    "event-fixture",
				"destination": "destination-fixture",
			},
			wantPath:  "/calendar/v3/calendars/calendar-fixture/events/event-fixture/move",
			wantQuery: map[string]string{"destination": "destination-fixture"},
		},
		{
			name:   "quick add event",
			action: "quick_add_event",
			record: connectors.Record{
				"calendar_id": "calendar-fixture",
				"text":        "fixture meeting tomorrow",
			},
			wantPath:  "/calendar/v3/calendars/calendar-fixture/events/quickAdd",
			wantQuery: map[string]string{"text": "fixture meeting tomorrow"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, closeServer := newFixtureConnector(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %q, want POST", r.Method)
				}
				if r.URL.Path != tt.wantPath {
					t.Errorf("path = %q, want %q", r.URL.Path, tt.wantPath)
				}
				for key, want := range tt.wantQuery {
					if got := r.URL.Query().Get(key); got != want {
						t.Errorf("query[%q] = %q, want %q", key, got, want)
					}
				}
				_ = json.NewEncoder(w).Encode(map[string]any{})
			}))
			defer closeServer()

			records := []connectors.Record{tt.record}
			req := connectors.WriteRequest{
				Action: tt.action,
				Config: connectors.RuntimeConfig{ProjectDir: "__polymetrics_execution_fixture__"},
			}
			if tt.requiresApproval {
				req = approvedFixtureWriteRequest(t, conn, req, records)
			}
			result, err := conn.Write(context.Background(), req, records)
			if err != nil {
				t.Fatalf("Write: %v", err)
			}
			if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
				t.Fatalf("Write result = %+v, want one written record", result)
			}
		})
	}
}

func approvedFixtureWriteRequest(t *testing.T, conn *engine.Connector, req connectors.WriteRequest, records []connectors.Record) connectors.WriteRequest {
	t.Helper()
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatalf("NewFixtureWriteApprovalAuthority: %v", err)
	}
	req.Config.CredentialRevision, err = authority.CredentialRevision("google-calendar-query-shape-fixture", req.Config.Secrets)
	if err != nil {
		t.Fatalf("CredentialRevision: %v", err)
	}
	req.Config.ConfigurationDigest, err = authority.ConfigurationDigest("google-calendar-query-shape-fixture", req.Config.Config)
	if err != nil {
		t.Fatalf("ConfigurationDigest: %v", err)
	}
	req.Config.WriteApprovalScope = connectors.WriteApprovalScopeFixture

	preview, err := conn.DryRunWrite(context.Background(), req, records)
	if err != nil {
		t.Fatalf("DryRunWrite: %v", err)
	}
	const approvalToken = "google-calendar-query-shape-approval"
	confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID:        "rplan_google_calendar_query_shape_fixture",
		PlanHash:      strings.Repeat("a", 64),
		PreviewDigest: preview.Digest,
		ApprovalToken: approvalToken,
		Target:        preview.ApprovalTarget,
		Confirmation:  confirmation,
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant: %v", err)
	}
	req.Approval, err = authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID:        grant.PlanID,
		PlanHash:      grant.PlanHash,
		PreviewDigest: grant.PreviewDigest,
		ApprovalToken: approvalToken,
		Target:        grant.Target,
		Confirmation:  grant.Confirmation,
	})
	if err != nil {
		t.Fatalf("VerifyWriteGrant: %v", err)
	}
	return req
}

func TestEventsInitialReadDoesNotApplyImplicitCutoff(t *testing.T) {
	tests := []struct {
		name           string
		startDate      string
		wantUpdatedMin string
	}{
		{name: "unfiltered fresh read"},
		{name: "explicit lower bound", startDate: "2020-01-01T00:00:00Z", wantUpdatedMin: "2020-01-01T00:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUpdatedMin string
			conn, closeServer := newFixtureConnector(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/calendar/v3/calendars/primary/events" {
					t.Errorf("path = %q", r.URL.Path)
				}
				gotUpdatedMin = r.URL.Query().Get("updatedMin")
				_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"id": "event-1", "updated": "2026-01-01T00:00:00Z"}}})
			}))
			defer closeServer()

			config := map[string]string{}
			if tt.startDate != "" {
				config["start_date"] = tt.startDate
			}
			records := 0
			err := conn.Read(context.Background(), connectors.ReadRequest{
				Stream: "events",
				Config: connectors.RuntimeConfig{
					ProjectDir: "__polymetrics_execution_fixture__",
					Config:     config,
				},
			}, func(connectors.Record) error {
				records++
				return nil
			})
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if records != 1 {
				t.Fatalf("records = %d, want 1", records)
			}
			if gotUpdatedMin != tt.wantUpdatedMin {
				t.Fatalf("updatedMin = %q, want %q", gotUpdatedMin, tt.wantUpdatedMin)
			}
		})
	}
}

func TestLegacyStreamsProjectDeclaredSchemas(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		record        map[string]any
		excludedField string
	}{
		{
			name:          "calendar_list",
			path:          "/calendar/v3/users/me/calendarList",
			record:        map[string]any{"id": "calendar-1", "summary": "Calendar", "backgroundColor": "#ffffff"},
			excludedField: "backgroundColor",
		},
		{
			name:          "events",
			path:          "/calendar/v3/calendars/primary/events",
			record:        map[string]any{"id": "event-1", "updated": "2026-01-01T00:00:00Z", "conferenceData": map[string]any{}},
			excludedField: "conferenceData",
		},
		{
			name:          "settings",
			path:          "/calendar/v3/users/me/settings",
			record:        map[string]any{"id": "timezone", "value": "UTC", "providerOnly": true},
			excludedField: "providerOnly",
		},
		{
			name:          "acl",
			path:          "/calendar/v3/calendars/primary/acl",
			record:        map[string]any{"id": "rule-1", "role": "reader", "scope": map[string]any{"type": "default"}, "providerOnly": true},
			excludedField: "providerOnly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, closeServer := newFixtureConnector(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("path = %q, want %q", r.URL.Path, tt.path)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{tt.record}})
			}))
			defer closeServer()

			var got connectors.Record
			err := conn.Read(context.Background(), connectors.ReadRequest{
				Stream: tt.name,
				Config: connectors.RuntimeConfig{ProjectDir: "__polymetrics_execution_fixture__"},
			}, func(record connectors.Record) error {
				got = record
				return nil
			})
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if got == nil {
				t.Fatal("no record emitted")
			}
			if _, ok := got[tt.excludedField]; ok {
				t.Fatalf("record retained undeclared field %q: %#v", tt.excludedField, got)
			}
		})
	}
}

func TestSettingsReadFollowsNextPageToken(t *testing.T) {
	requests := 0
	conn, closeServer := newFixtureConnector(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/calendar/v3/users/me/settings" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("maxResults"); got != "250" {
			t.Errorf("maxResults = %q, want 250", got)
		}
		switch token := r.URL.Query().Get("pageToken"); token {
		case "":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"nextPageToken": "settings-page-2",
				"items":         []any{map[string]any{"id": "timezone", "value": "UTC"}},
			})
		case "settings-page-2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []any{map[string]any{"id": "weekStart", "value": "1"}},
			})
		default:
			t.Errorf("pageToken = %q", token)
			_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
		}
	}))
	defer closeServer()

	var records []connectors.Record
	err := conn.Read(context.Background(), connectors.ReadRequest{
		Stream: "settings",
		Config: connectors.RuntimeConfig{ProjectDir: "__polymetrics_execution_fixture__"},
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
}

func newFixtureConnector(t *testing.T, handler http.Handler) (*engine.Connector, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	b, err := engine.Load(defs.FS, "google-calendar")
	if err != nil {
		srv.Close()
		t.Fatalf("Load: %v", err)
	}
	b.HTTP.URL = srv.URL + "/calendar/v3"
	return engine.New(b, testNoAuthHooks{}), srv.Close
}
