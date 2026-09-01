package ashby

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestConnectorContract(t *testing.T) {
	assertConnectorContract(t, New(), "ashby")
}

func TestOperationDirectReadPreflightDelegatesToDeclaredAshbyOperation(t *testing.T) {
	preflighter, ok := New().(connectors.OperationDirectReadPreflighter)
	if !ok {
		t.Fatal("Ashby connector does not expose operation direct-read preflight")
	}
	if err := preflighter.PreflightOperationDirectRead("ashby.direct.candidate.search", http.MethodPost, "/candidate.search", 16<<20, "json_redacted"); err != nil {
		t.Fatalf("PreflightOperationDirectRead valid command: %v", err)
	}
	if err := preflighter.PreflightOperationDirectRead("ashby.direct.candidate.search", http.MethodGet, "/candidate.search", 16<<20, "json_redacted"); err == nil || !strings.Contains(err.Error(), "does not match declared operation method") {
		t.Fatalf("PreflightOperationDirectRead method mismatch = %v, want declared operation rejection", err)
	}
}

func TestOperationClassificationsMatchAshbySemantics(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../../defs"), "ashby")
	if err != nil {
		t.Fatalf("load Ashby disk bundle: %v", err)
	}
	if len(bundle.Streams) != 71 || len(bundle.Operations) != 9 || len(bundle.Writes) != 98 {
		t.Fatalf("implemented counts = streams:%d direct_reads:%d writes:%d, want 71/9/98", len(bundle.Streams), len(bundle.Operations), len(bundle.Writes))
	}

	for _, stream := range bundle.Streams {
		if stream.Name == "referral_form_info" {
			t.Fatal("referralForm.info remains executable as an ETL stream")
		}
	}
	blockedWritePaths := map[string]bool{
		"/applicationForm.submit":   false,
		"/report.generate":          false,
		"/user.interviewerSettings": false,
	}
	for _, action := range bundle.Writes {
		if _, blocked := blockedWritePaths[action.Path]; blocked {
			t.Fatalf("%s remains executable as write action %s", action.Path, action.Name)
		}
	}

	wantDirect := map[string]string{
		"user.interviewerSettings": "/user.interviewerSettings",
		"report.generate":          "/report.generate",
	}
	for _, operation := range bundle.Operations {
		wantPath, ok := wantDirect[operation.Summary]
		if !ok {
			continue
		}
		if operation.Kind != "rest_read" || operation.OutputPolicy != "json_redacted" || operation.Approval != "none" || operation.REST == nil || operation.REST.Path != wantPath || operation.REST.MaxBytes != 1<<20 {
			t.Fatalf("operation %s = %+v, want bounded json_redacted REST read", operation.Summary, operation)
		}
		delete(wantDirect, operation.Summary)
	}
	if len(wantDirect) != 0 {
		t.Fatalf("missing direct-read operation classifications: %v", wantDirect)
	}

	wantBlocked := map[string]string{
		"/referralForm.info":      "ashby-referral-form-info-side-effect-foundation",
		"/applicationForm.submit": "ashby-application-form-typed-multipart-foundation",
	}
	blockedCount := 0
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.Operation != nil {
			blockedCount++
		}
		foundation, ok := wantBlocked[endpoint.Path]
		if !ok {
			continue
		}
		if endpoint.Operation == nil || endpoint.Operation.Status != "blocked" || !endpoint.Operation.BlockedByDefault || !strings.Contains(endpoint.Operation.Reason, foundation) {
			t.Fatalf("surface %s = %+v, want named blocked foundation %s", endpoint.Path, endpoint.Operation, foundation)
		}
		delete(wantBlocked, endpoint.Path)
	}
	if blockedCount != 34 || len(wantBlocked) != 0 {
		t.Fatalf("blocked ledger count = %d, missing = %v; want 34 and both named blockers", blockedCount, wantBlocked)
	}

	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, command := range New().(connectors.CommandSurfaceProvider).CommandSurface().Commands {
		commands[command.Path] = command
	}
	for _, path := range []string{"referral-form info", "application-form submit"} {
		if _, ok := commands[path]; ok {
			t.Fatalf("blocked command %q remains executable", path)
		}
	}
	for path, operation := range map[string]string{
		"user interviewer-settings": "ashby.direct.user.interviewer.settings",
		"report generate":           "ashby.direct.report.generate",
	} {
		command, ok := commands[path]
		if !ok {
			t.Fatalf("direct-read command %q not found", path)
		}
		if command.Intent != "direct_read" || command.Operation != operation || command.Write != "" || command.OutputPolicy != "json_redacted" || command.Approval != "none" {
			t.Fatalf("command %q = %+v, want direct read %s", path, command, operation)
		}
		if path == "report generate" && command.Summary != "Start an Ashby report generation or check an existing request." {
			t.Fatalf("command %q summary = %q, want bounded report help", path, command.Summary)
		}
		for _, flag := range command.Flags {
			if !strings.HasPrefix(flag.MapsTo, "body.") {
				t.Fatalf("command %q flag --%s maps to %q, want body field", path, flag.Name, flag.MapsTo)
			}
		}
	}
}

func TestSyncTokenCommandsUseFullRefreshHelp(t *testing.T) {
	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, command := range New().(connectors.CommandSurfaceProvider).CommandSurface().Commands {
		if command.Stream != "" {
			commands[command.Stream] = command
		}
	}
	found := 0
	for stream, endpoint := range ashbyStreamEndpoints {
		if _, ok := endpoint.requestFields["syncToken"]; !ok {
			continue
		}
		found++
		command, ok := commands[stream]
		if !ok {
			t.Fatalf("sync-token stream command %s not found", stream)
		}
		if !strings.Contains(command.Summary, "Full-refresh-only") || !strings.Contains(command.Summary, "ashby-sync-token-checkpoint-foundation") || strings.Contains(strings.ToLower(command.Summary), "incremental") {
			t.Fatalf("stream %s summary = %q, want connector-owned full-refresh blocker help", stream, command.Summary)
		}
		if !strings.Contains(command.Notes, "ashby-sync-token-checkpoint-foundation") {
			t.Fatalf("stream %s notes = %q, want sync-token foundation", stream, command.Notes)
		}
	}
	if found == 0 {
		t.Fatal("no sync-token-capable Ashby stream endpoints found")
	}
}

func TestCommandSurfaceSummariesRenderAsPlainTerminalHelp(t *testing.T) {
	for _, command := range New().(connectors.CommandSurfaceProvider).CommandSurface().Commands {
		summary := strings.TrimSpace(command.Summary)
		if summary == "" {
			t.Fatalf("command %q has an empty summary", command.Path)
		}
		if strings.ContainsAny(summary, "\r\n") {
			t.Fatalf("command %q summary contains a line break: %q", command.Path, summary)
		}
		for _, marker := range []string{"](", "**", "`"} {
			if strings.Contains(summary, marker) {
				t.Fatalf("command %q summary contains raw Markdown marker %q: %q", command.Path, marker, summary)
			}
		}
		if strings.HasPrefix(summary, ">") {
			t.Fatalf("command %q summary contains a raw Markdown blockquote: %q", command.Path, summary)
		}
		if len(summary) > 160 {
			t.Fatalf("command %q summary is %d bytes, want at most 160: %q", command.Path, len(summary), summary)
		}
	}

	for _, command := range New().(connectors.CommandSurfaceProvider).CommandSurface().Commands {
		if command.Path != "application create" {
			continue
		}
		const want = "Consider a candidate for a job (e.g. when sourcing a candidate for a job posting)."
		if command.Summary != want {
			t.Fatalf("application create summary = %q, want %q", command.Summary, want)
		}
		return
	}
	t.Fatal("application create command not found")
}

func TestSemanticDirectReadsReturnAshbyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/user.interviewerSettings":
			if body["userId"] != "user_fixture" {
				t.Fatalf("body[userId] = %v, want user_fixture", body["userId"])
			}
			_, _ = w.Write([]byte(`{"success":true,"results":{"timezone":"UTC","apiToken":"synthetic-response-token"}}`))
		case "/report.generate":
			if body["reportId"] != "report_fixture" {
				t.Fatalf("body[reportId] = %v, want report_fixture", body["reportId"])
			}
			_, _ = w.Write([]byte(`{"success":true,"results":{"status":"pending","requestId":"request_fixture"}}`))
		default:
			t.Fatalf("unexpected direct-read path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	tests := []struct {
		name      string
		operation string
		body      map[string]any
		wantField string
		wantValue any
	}{
		{name: "interviewer settings", operation: "ashby.direct.user.interviewer.settings", body: map[string]any{"userId": "user_fixture"}, wantField: "timezone", wantValue: "UTC"},
		{name: "report generation result", operation: "ashby.direct.report.generate", body: map[string]any{"reportId": "report_fixture"}, wantField: "requestId", wantValue: "request_fixture"},
	}
	reader := New().(connectors.OperationDirectReader)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reader.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
				Operation: tt.operation,
				Config: connectors.RuntimeConfig{
					Config:  map[string]string{"base_url": server.URL},
					Secrets: map[string]string{"api_key": "test_key"},
				},
				Body:         tt.body,
				OutputPolicy: "json_redacted",
				MaxBytes:     1 << 20,
			})
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			body, ok := result.Body.(map[string]any)
			if !ok {
				t.Fatalf("result body = %T, want object", result.Body)
			}
			results, ok := body["results"].(map[string]any)
			if !ok || results[tt.wantField] != tt.wantValue {
				t.Fatalf("results = %+v, want %s=%v", body["results"], tt.wantField, tt.wantValue)
			}
			if tt.operation == "ashby.direct.user.interviewer.settings" {
				if results["apiToken"] != "synthetic-response-token" {
					t.Fatalf("results = %+v, want ordinary unclassified field preserved", results)
				}
			}
		})
	}
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
	if _, ok := c.(connectors.OperationDirectReadPreflighter); !ok {
		t.Fatalf("%s must expose operation direct-read preflight", wantName)
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

func TestValidateWriteRejectsNestedArbitraryFields(t *testing.T) {
	validator := New().(connectors.WriteValidator)
	record := connectors.Record{
		"candidateId": "candidate_fixture",
		"jobId":       "job_fixture",
		"applicationHistory": []any{map[string]any{
			"stageId":        "stage_fixture",
			"stageNumber":    1,
			"enteredStageAt": "2026-01-01T00:00:00Z",
			"arbitrary":      "blocked",
		}},
	}
	err := validator.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "create_application"}, []connectors.Record{record})
	if err == nil || !strings.Contains(err.Error(), "additional property") {
		t.Fatalf("ValidateWrite nested arbitrary field error = %v, want additional property rejection", err)
	}
}

func TestValidateWritePreservesDocumentedMapFields(t *testing.T) {
	validator := New().(connectors.WriteValidator)
	record := connectors.Record{
		"name":      "Fixture Department",
		"extraData": map[string]any{"documented_map_key": "fixture value"},
	}
	if err := validator.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "create_department"}, []connectors.Record{record}); err != nil {
		t.Fatalf("ValidateWrite documented map field: %v", err)
	}
}

func TestAnonymizeCandidateRequiresDestructiveConfirmation(t *testing.T) {
	def := New().(connectors.DefinitionProvider).Definition()
	for _, action := range def.WriteActions {
		if action.Name != "anonymize_candidate" {
			continue
		}
		if action.Confirm != "destructive" {
			t.Fatalf("anonymize_candidate confirm = %q, want destructive", action.Confirm)
		}
		return
	}
	t.Fatal("anonymize_candidate write action not found")
}

func TestDestructiveValueWritesRequireConfirmation(t *testing.T) {
	def := New().(connectors.DefinitionProvider).Definition()
	confirmByAction := map[string]string{}
	for _, action := range def.WriteActions {
		confirmByAction[action.Name] = action.Confirm
	}
	tests := []struct {
		name string
	}{
		{name: "create_application"},
		{name: "change_application_stage"},
		{name: "change_application_stage_2"},
		{name: "update_application_history"},
		{name: "set_job_status"},
		{name: "set_offer_status"},
		{name: "create_opening"},
		{name: "set_opening_archived"},
		{name: "set_opening_opening_state"},
		{name: "update_custom_field_selectable_values"},
		{name: "update_assessment"},
		{name: "discard_sequence"},
		{name: "update_interview_schedule"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirm, ok := confirmByAction[tt.name]
			if !ok {
				t.Fatalf("write action not found")
			}
			if confirm != "destructive" {
				t.Fatalf("confirm = %q, want destructive", confirm)
			}
		})
	}
}

func TestCommandSurfaceDocumentsSignedURLPreservation(t *testing.T) {
	surface := New().(connectors.CommandSurfaceProvider).CommandSurface()
	want := map[string]bool{"notetaker-transcript info": false, "file info": false}
	for _, cmd := range surface.Commands {
		if _, ok := want[cmd.Path]; !ok {
			continue
		}
		if cmd.OutputPolicy != "json_redacted" {
			t.Fatalf("command %q output policy = %q, want json_redacted", cmd.Path, cmd.OutputPolicy)
		}
		if !strings.Contains(cmd.Risk, "signed URL fields are preserved") {
			t.Fatalf("command %q risk = %q, want signed URL preservation", cmd.Path, cmd.Risk)
		}
		want[cmd.Path] = true
	}
	for path, found := range want {
		if !found {
			t.Fatalf("command %q not found", path)
		}
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

func TestHiringTeamRoleListDefaultsNamesOnlyTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hiringTeamRole.list" {
			t.Fatalf("path = %q, want /hiringTeamRole.list", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := body["namesOnly"]; got != true {
			t.Fatalf("body[namesOnly] = %v, want true", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"results":["Hiring Team Role fixture"],"moreDataAvailable":false}`))
	}))
	defer server.Close()

	var records []connectors.Record
	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "hiring_team_role_list",
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
	if len(records) != 1 || records[0]["value"] != "Hiring Team Role fixture" {
		t.Fatalf("records = %+v, want scalar role title value", records)
	}
}

func TestHiringTeamRoleListBlocksNamesOnlyFalseVariant(t *testing.T) {
	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "hiring_team_role_list",
		Config: connectors.RuntimeConfig{Secrets: map[string]string{"api_key": "test_key"}},
		Query:  map[string]string{"namesOnly": "false"},
	}, func(connectors.Record) error { return nil })
	if err == nil {
		t.Fatal("Read accepted namesOnly=false")
	}
	if !strings.Contains(err.Error(), "ashby_hiring_team_role_list_names_only_false") {
		t.Fatalf("error = %v, want named variant-schema blocker", err)
	}
}

// TestReadDefaultsToExhaustivePagination pins the unbounded default: an unset
// max_pages must keep following nextCursor while the provider still reports
// moreDataAvailable, rather than silently truncating an ETL read.
func TestReadDefaultsToExhaustivePagination(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/candidate.list" {
			t.Errorf("path = %q, want /candidate.list", r.URL.Path)
		}
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch requestCount {
		case 1:
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"first","updatedAt":"2026-01-03T00:00:00Z"}],"moreDataAvailable":true,"nextCursor":"opaque-page-2"}`))
		case 2:
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"second","updatedAt":"2026-01-04T00:00:00Z"}],"moreDataAvailable":true,"nextCursor":"opaque-page-3"}`))
		case 3:
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"third","updatedAt":"2026-01-05T00:00:00Z"}],"moreDataAvailable":false}`))
		default:
			t.Errorf("unexpected request %d after the provider stopped advertising more data", requestCount)
			_, _ = w.Write([]byte(`{"success":true,"results":[],"moreDataAvailable":false}`))
		}
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
	if requestCount != 3 {
		t.Fatalf("request count = %d, want 3", requestCount)
	}
	if len(records) != 3 {
		t.Fatalf("emitted %d records = %+v, want 3", len(records), records)
	}
	for i, want := range []string{"first", "second", "third"} {
		if records[i]["id"] != want {
			t.Errorf("records[%d][id] = %v, want %s", i, records[i]["id"], want)
		}
	}
}

// TestDefaultMaxPagesMatchesSpecDefault guards the pair that together decide
// the effective default: the engine materializes spec.json's declared default
// into RuntimeConfig.Config, so a spec default that disagrees with
// ashbyDefaultMaxPages would silently win on the declarative read path.
func TestDefaultMaxPagesMatchesSpecDefault(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "defs", "ashby", "spec.json"))
	if err != nil {
		t.Fatalf("read spec.json: %v", err)
	}
	var spec struct {
		Properties map[string]struct {
			Default string `json:"default"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("decode spec.json: %v", err)
	}
	declared := spec.Properties["max_pages"].Default
	if declared == "" {
		t.Fatal("spec.json max_pages declares no default")
	}
	resolved, err := ashbyMaxPages(connectors.RuntimeConfig{Config: map[string]string{"max_pages": declared}})
	if err != nil {
		t.Fatalf("ashbyMaxPages(%q): %v", declared, err)
	}
	if resolved != ashbyDefaultMaxPages {
		t.Fatalf("spec default %q resolves to %d pages, want ashbyDefaultMaxPages %d", declared, resolved, ashbyDefaultMaxPages)
	}
	if resolved != 0 {
		t.Fatalf("default max pages = %d, want 0 (unbounded)", resolved)
	}
}

func TestReadDoesNotInferIncrementalStateFromTimestamps(t *testing.T) {
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
			_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"old","updatedAt":"2026-01-01T00:00:00Z"},{"id":"equal","updatedAt":"2026-01-02T00:00:00Z"},{"id":"new","updatedAt":"2026-01-03T00:00:00Z"}],"moreDataAvailable":true,"nextCursor":"opaque-page-2"}`))
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
	wantIDs := []string{"old", "equal", "new", "next"}
	if len(records) != len(wantIDs) {
		t.Fatalf("emitted %d records = %+v, want ids %v", len(records), records, wantIDs)
	}
	for i, wantID := range wantIDs {
		if got := records[i]["id"]; got != wantID {
			t.Fatalf("record %d id = %v, want %s", i, got, wantID)
		}
	}
}

func TestReadBlocksSyncTokenWithoutCheckpointFoundation(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		_, _ = w.Write([]byte(`{"success":true,"results":[],"moreDataAvailable":false}`))
	}))
	defer server.Close()

	tests := []struct {
		name   string
		query  map[string]string
		state  map[string]string
		config map[string]string
	}{
		{name: "query", query: map[string]string{"syncToken": "opaque-sync-token"}},
		{name: "state", state: map[string]string{"syncToken": "opaque-sync-token"}},
		{name: "config", config: map[string]string{"sync_token": "opaque-sync-token"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]string{"base_url": server.URL}
			for key, value := range tt.config {
				config[key] = value
			}
			err := New().Read(context.Background(), connectors.ReadRequest{
				Stream: "candidates",
				Config: connectors.RuntimeConfig{
					Config:  config,
					Secrets: map[string]string{"api_key": "test_key"},
				},
				Query: tt.query,
				State: tt.state,
			}, func(connectors.Record) error { return nil })
			if err == nil || !strings.Contains(err.Error(), "ashby-sync-token-checkpoint-foundation") {
				t.Fatalf("Read syncToken error = %v, want named checkpoint foundation blocker", err)
			}
		})
	}
	if requestCount != 0 {
		t.Fatalf("provider request count = %d, want 0", requestCount)
	}
}

func TestReadStopsAtPageBoundBeforeCursorReuse(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		_, _ = w.Write([]byte(`{"success":true,"results":[],"moreDataAvailable":true,"nextCursor":"bounded-page-token"}`))
	}))
	defer server.Close()

	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "candidates",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL, "max_pages": "2"},
			Secrets: map[string]string{"api_key": "test_key"},
		},
	}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("Read bounded repeated cursor: %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("provider request count = %d, want bounded 2", requestCount)
	}
}

func TestReadRejectsRepeatedPageCursorBeforeReuse(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"candidate_fixture"}],"moreDataAvailable":true,"nextCursor":"repeated-page-token"}`))
	}))
	defer server.Close()

	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "candidates",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL, "max_pages": "3"},
			Secrets: map[string]string{"api_key": "test_key"},
		},
	}, func(connectors.Record) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "repeated pagination cursor") {
		t.Fatalf("Read repeated cursor error = %v, want cycle rejection", err)
	}
	if requestCount != 2 {
		t.Fatalf("provider request count = %d, want 2 before cursor reuse", requestCount)
	}
}

func TestAshbyCatalogAdvertisesFullRefreshOnly(t *testing.T) {
	c := New()
	catalog, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	for _, stream := range catalog.Streams {
		if len(stream.CursorFields) != 0 {
			t.Fatalf("stream %s cursor fields = %v, want none until opaque syncToken checkpointing exists", stream.Name, stream.CursorFields)
		}
	}
	for _, mode := range connectors.ManifestOf(c).SyncModes {
		if strings.HasPrefix(mode, "incremental_") {
			t.Fatalf("manifest sync mode %q advertises unsupported Ashby incremental state", mode)
		}
	}
}

func TestApplicationListHistoryIsPaginationOnly(t *testing.T) {
	if fields := cursorFields(ashbyStreamEndpoints["application_list_history"]); len(fields) != 0 {
		t.Fatalf("application_list_history cursor fields = %v, want pagination-only stream", fields)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/application.listHistory" {
			t.Fatalf("path = %q, want /application.listHistory", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got := body["applicationId"]; got != "application_fixture" {
			t.Fatalf("body[applicationId] = %v, want application_fixture", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"results":[{"id":"old","enteredStageAt":"2026-01-01T00:00:00Z"},{"id":"new","enteredStageAt":"2026-01-03T00:00:00Z"}],"moreDataAvailable":false}`))
	}))
	defer server.Close()

	var records []connectors.Record
	err := New().Read(context.Background(), connectors.ReadRequest{
		Stream: "application_list_history",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL, "max_pages": "1", "start_date": "2026-01-02T00:00:00Z"},
			Secrets: map[string]string{"api_key": "test_key"},
		},
		Query: map[string]string{"applicationId": "application_fixture"},
		State: map[string]string{"cursor": "2026-01-02T00:00:00Z"},
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantIDs := []string{"old", "new"}
	if len(records) != len(wantIDs) {
		t.Fatalf("emitted %d records = %+v, want ids %v", len(records), records, wantIDs)
	}
	for i, wantID := range wantIDs {
		if got := records[i]["id"]; got != wantID {
			t.Fatalf("record %d id = %v, want %s", i, got, wantID)
		}
	}
}

func TestCommandSurfaceMarksRequiredStreamSelectors(t *testing.T) {
	surface := New().(connectors.CommandSurfaceProvider).CommandSurface()
	commandsByStream := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range surface.Commands {
		if cmd.Stream != "" {
			commandsByStream[cmd.Stream] = cmd
		}
	}
	for stream, endpoint := range ashbyStreamEndpoints {
		for _, field := range endpoint.requiredFields {
			cmd, ok := commandsByStream[stream]
			if !ok {
				t.Fatalf("stream command %s not found", stream)
			}
			mapsTo := "query." + field
			found := false
			for _, flag := range cmd.Flags {
				if flag.MapsTo != mapsTo {
					continue
				}
				found = true
				if !flag.Required {
					t.Fatalf("stream %s flag --%s required = false, want true", stream, flag.Name)
				}
			}
			if !found {
				t.Fatalf("stream %s missing CLI flag for %s", stream, mapsTo)
			}
		}
	}
}

func TestCommandSurfaceBlocksRepeatableStreamArrayVariants(t *testing.T) {
	surface := New().(connectors.CommandSurfaceProvider).CommandSurface()
	blocked := 0
	for _, cmd := range surface.Commands {
		if cmd.Intent != "etl" {
			continue
		}
		for _, flag := range cmd.Flags {
			if flag.Type == "string_array" {
				t.Fatalf("stream command %q exposes unsupported repeatable flag --%s", cmd.Path, flag.Name)
			}
		}
		if strings.Contains(cmd.Notes, "connector-stream-repeatable-array-foundation") {
			blocked++
		}
	}
	if blocked == 0 {
		t.Fatal("no Ashby stream command records the repeatable-array foundation blocker")
	}
}

func TestAshbyResultRecordsRejectsErrorEnvelopes(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "false success", body: `{"success":false,"error":"fixture"}`, wantErr: "success=false"},
		{name: "missing success", body: `{"results":[]}`, wantErr: "missing success"},
		{name: "missing results", body: `{"success":true,"moreDataAvailable":false}`, wantErr: "missing results"},
		{name: "malformed success", body: `{"success":"false","results":[]}`, wantErr: "success field must be boolean"},
		{name: "malformed results", body: `{"success":true,"results":1}`, wantErr: "unsupported type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, records, err := ashbyResultRecords([]byte(tt.body))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ashbyResultRecords error = %v, want %q", err, tt.wantErr)
			}
			if len(records) != 0 {
				t.Fatalf("records = %+v, want none on error", records)
			}
		})
	}
}

func TestAshbyOperationDirectReadPreservesEngineResultOnEnvelopeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/candidate.search" {
			t.Fatalf("path = %q, want declared candidate search route", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("X-Provider-Trace", "first")
		w.Header().Add("X-Provider-Trace", "second")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":false,"error":"logical failure","requestId":"ashby-occurrence-9007199254740993"}`))
	}))
	t.Cleanup(server.Close)

	result, err := New().(connectors.OperationDirectReader).OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "ashby.direct.candidate.search",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL},
			Secrets: map[string]string{"api_key": "test_key"},
		},
		Body:         map[string]any{"email": "candidate@example.invalid"},
		OutputPolicy: "json_redacted",
		MaxBytes:     1024,
	})
	if err == nil || !strings.Contains(err.Error(), "success=false") {
		t.Fatalf("OperationDirectRead error = %v, want logical Ashby envelope failure", err)
	}
	if result.Status != http.StatusOK || result.Receipt == nil || !result.Receipt.ResponseReceived {
		t.Fatalf("OperationDirectRead result = %#v, want retained received provider response", result)
	}
	if got := result.Receipt.Headers["X-Provider-Trace"].Values; len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("Ashby receipt headers = %#v, want repeatable provider trace", result.Receipt.Headers)
	}
}

func TestAshbySiblingPathsRejectCredentialSafeErrorEnvelopes(t *testing.T) {
	const sensitive = "synthetic-sensitive-response-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"error":{"apiToken":"` + sensitive + `"}}`))
	}))
	defer server.Close()
	cfg := connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": server.URL},
		Secrets: map[string]string{"api_key": "test_key"},
	}

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "check", run: func() error { return New().Check(context.Background(), cfg) }},
		{name: "direct read", run: func() error {
			_, err := New().(connectors.OperationDirectReader).OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
				Operation:    "ashby.direct.candidate.search",
				Config:       cfg,
				Body:         map[string]any{"email": "candidate@example.invalid"},
				OutputPolicy: "json_redacted",
			})
			return err
		}},
		{name: "write", run: func() error {
			result, err := New().Write(context.Background(), connectors.WriteRequest{Action: "create_candidate_tag", Config: cfg}, []connectors.Record{{"title": "Fixture Tag"}})
			if result.RecordsWritten != 0 || result.RecordsFailed != 1 {
				return fmt.Errorf("write result = %+v, want one failed record", result)
			}
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil || !strings.Contains(err.Error(), "ashby response success=false") {
				t.Fatalf("error = %v, want Ashby success-envelope rejection", err)
			}
			if strings.Contains(err.Error(), sensitive) {
				t.Fatalf("error persisted sensitive response content: %v", err)
			}
		})
	}
}

func TestWriteAcceptsSuccessfulAshbyEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/candidateTag.create" {
			t.Fatalf("path = %q, want /candidateTag.create", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["title"] != "Fixture Tag" {
			t.Fatalf("body[title] = %v, want Fixture Tag", body["title"])
		}
		_, _ = w.Write([]byte(`{"success":true,"results":{"id":"tag_fixture"}}`))
	}))
	defer server.Close()

	result, err := New().Write(context.Background(), connectors.WriteRequest{
		Action: "create_candidate_tag",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL},
			Secrets: map[string]string{"api_key": "test_key"},
		},
	}, []connectors.Record{{"title": "Fixture Tag"}})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		t.Fatalf("write result = %+v, want one successful record", result)
	}
}

func TestDestructiveWriteUsesSameHookAwarePreviewAtExecution(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/application.delete" {
			t.Fatalf("path = %q, want /application.delete", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"success":true,"results":{}}`))
	}))
	defer server.Close()

	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		t.Fatalf("NewFixtureWriteApprovalAuthority() error = %v", err)
	}
	cfg := connectors.RuntimeConfig{
		Config:             map[string]string{"base_url": server.URL},
		Secrets:            map[string]string{"api_key": "test_key"},
		WriteApprovalScope: connectors.WriteApprovalScopeFixture,
	}
	cfg.CredentialRevision, err = authority.CredentialRevision("ashby-fixture", cfg.Secrets)
	if err != nil {
		t.Fatalf("CredentialRevision() error = %v", err)
	}
	cfg.ConfigurationDigest, err = authority.ConfigurationDigest("ashby-fixture", cfg.Config)
	if err != nil {
		t.Fatalf("ConfigurationDigest() error = %v", err)
	}
	records := []connectors.Record{{"applicationId": "application_fixture"}}
	req := connectors.WriteRequest{Action: "delete_application", Config: cfg}
	preview, err := New().(connectors.DryRunWriter).DryRunWrite(context.Background(), req, records)
	if err != nil {
		t.Fatalf("DryRunWrite() error = %v", err)
	}
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID: "rplan_fixture", PlanHash: strings.Repeat("a", 64), PreviewDigest: preview.Digest,
		ApprovalToken: "fixture-token", Target: preview.ApprovalTarget,
		Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		t.Fatalf("IssueWriteGrant() error = %v", err)
	}
	req.Approval, err = authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID: grant.PlanID, PlanHash: grant.PlanHash, PreviewDigest: preview.Digest,
		ApprovalToken: "fixture-token", Target: preview.ApprovalTarget, Confirmation: grant.Confirmation,
	})
	if err != nil {
		t.Fatalf("VerifyWriteGrant() error = %v", err)
	}

	result, err := New().Write(context.Background(), req, records)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 || calls != 1 {
		t.Fatalf("Write() result = %+v, calls = %d, want one successful request", result, calls)
	}
}

func TestCommandSurfaceDirectReadsDoNotRedactNonCredentialFields(t *testing.T) {
	surface := New().(connectors.CommandSurfaceProvider).CommandSurface()
	for _, cmd := range surface.Commands {
		if cmd.Intent != "direct_read" {
			continue
		}
		if len(cmd.RedactFields) != 0 {
			t.Fatalf("direct-read command %q redact_fields = %v, want none", cmd.Path, cmd.RedactFields)
		}
	}
}

func TestOperationDirectReadPreservesNonCredentialIdentityFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/candidate.search":
			_, _ = w.Write([]byte(`{"success":true,"results":[{"email":"candidate@example.invalid","name":"Ada Candidate","apiToken":"synthetic-response-token"}]}`))
		case "/job.search":
			_, _ = w.Write([]byte(`{"success":true,"results":[{"title":"Fixture Job","requisitionId":"REQ-123"}]}`))
		case "/user.search":
			_, _ = w.Write([]byte(`{"success":true,"results":[{"email":"user@example.invalid","name":"Ada User"}]}`))
		default:
			t.Fatalf("path = %q, want Ashby direct-read search endpoint", r.URL.Path)
		}
	}))
	defer server.Close()

	reader := New().(connectors.OperationDirectReader)
	tests := []struct {
		name      string
		operation string
		body      map[string]any
		want      map[string]any
	}{
		{name: "candidate", operation: "ashby.direct.candidate.search", body: map[string]any{"email": "candidate@example.invalid"}, want: map[string]any{"email": "candidate@example.invalid", "name": "Ada Candidate", "apiToken": "synthetic-response-token"}},
		{name: "job", operation: "ashby.direct.job.search", body: map[string]any{"title": "Fixture Job"}, want: map[string]any{"requisitionId": "REQ-123"}},
		{name: "user", operation: "ashby.direct.user.search", body: map[string]any{"email": "user@example.invalid"}, want: map[string]any{"email": "user@example.invalid", "name": "Ada User"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reader.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
				Operation: tt.operation,
				Config: connectors.RuntimeConfig{
					Config:  map[string]string{"base_url": server.URL},
					Secrets: map[string]string{"api_key": "test_key"},
				},
				Body:         tt.body,
				OutputPolicy: "json_redacted",
				MaxBytes:     1 << 20,
			})
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			body, ok := result.Body.(map[string]any)
			if !ok {
				t.Fatalf("result body = %T, want object", result.Body)
			}
			results, ok := body["results"].([]any)
			if !ok || len(results) != 1 {
				t.Fatalf("results = %#v, want one result", body["results"])
			}
			row, ok := results[0].(map[string]any)
			if !ok {
				t.Fatalf("result row = %T, want object", results[0])
			}
			for field, want := range tt.want {
				if got := row[field]; got != want {
					t.Fatalf("row[%s] = %v, want %v (row=%+v)", field, got, want, row)
				}
			}
		})
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

func TestOperationDirectReadRejectsEmptyJobSearch(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		t.Fatalf("unexpected provider request %s", r.URL.Path)
	}))
	defer server.Close()

	reader := New().(connectors.OperationDirectReader)
	_, err := reader.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "ashby.direct.job.search",
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL},
			Secrets: map[string]string{"api_key": "test_key"},
		},
		Body:         map[string]any{},
		OutputPolicy: "json_redacted",
		MaxBytes:     1 << 20,
	})
	if err == nil || !strings.Contains(err.Error(), "minProperties") {
		t.Fatalf("OperationDirectRead empty job.search error = %v, want minProperties", err)
	}
	if requestCount != 0 {
		t.Fatalf("provider request count = %d, want 0", requestCount)
	}
}

func TestReadInfoStreamsRequireOneDocumentedSelector(t *testing.T) {
	tests := []struct {
		name       string
		stream     string
		wantFields string
	}{
		{name: "application info", stream: "application_info", wantFields: "applicationId, submittedFormInstanceId"},
		{name: "candidate info", stream: "candidate_info", wantFields: "id, externalMappingId"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				t.Fatalf("unexpected provider request %s", r.URL.Path)
			}))
			defer server.Close()

			err := New().Read(context.Background(), connectors.ReadRequest{
				Stream: tt.stream,
				Config: connectors.RuntimeConfig{
					Config:  map[string]string{"base_url": server.URL, "max_pages": "1"},
					Secrets: map[string]string{"api_key": "test_key"},
				},
			}, func(record connectors.Record) error { return nil })
			if err == nil || !strings.Contains(err.Error(), tt.wantFields) {
				t.Fatalf("Read(%s) error = %v, want required selector %s", tt.stream, err, tt.wantFields)
			}
			if requestCount != 0 {
				t.Fatalf("provider request count = %d, want 0", requestCount)
			}
		})
	}
}

func TestOperationDirectReadPreservesAshbySignedURLs(t *testing.T) {
	const transcriptURL = "https://download.example.invalid/transcripts/interview_fixture.json?expires=1893456000&transcript_id=nt_fixture&signature=sanitized"
	const fileURL = "https://download.example.invalid/files/resume_fixture.pdf?expires=1893456000&file_id=file_fixture&signature=sanitized"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/notetakerTranscript.info":
			if body["notetakerTranscriptId"] != "nt_fixture" {
				t.Fatalf("body[notetakerTranscriptId] = %v, want nt_fixture", body["notetakerTranscriptId"])
			}
			_, _ = w.Write([]byte(`{"success":true,"results":{"transcriptUrl":"` + transcriptURL + `","apiToken":"synthetic-response-token"}}`))
		case "/file.info":
			if body["fileHandle"] != "file_fixture" {
				t.Fatalf("body[fileHandle] = %v, want file_fixture", body["fileHandle"])
			}
			_, _ = w.Write([]byte(`{"success":true,"results":{"url":"` + fileURL + `","mimeType":"application/pdf","apiToken":"synthetic-response-token"}}`))
		default:
			t.Fatalf("path = %q, want Ashby signed URL direct-read endpoint", r.URL.Path)
		}
	}))
	defer server.Close()

	tests := []struct {
		name      string
		operation string
		body      map[string]any
		field     string
		wantURL   string
	}{
		{name: "notetaker transcript", operation: "ashby.direct.notetaker.transcript.info", body: map[string]any{"notetakerTranscriptId": "nt_fixture"}, field: "transcriptUrl", wantURL: transcriptURL},
		{name: "file", operation: "ashby.direct.file.info", body: map[string]any{"fileHandle": "file_fixture"}, field: "url", wantURL: fileURL},
	}
	reader := New().(connectors.OperationDirectReader)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reader.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
				Operation: tt.operation,
				Config: connectors.RuntimeConfig{
					Config:  map[string]string{"base_url": server.URL},
					Secrets: map[string]string{"api_key": "test_key"},
				},
				Body:         tt.body,
				OutputPolicy: "json_redacted",
				MaxBytes:     1 << 20,
			})
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			body, ok := result.Body.(map[string]any)
			if !ok {
				t.Fatalf("result body = %T, want object", result.Body)
			}
			results, ok := body["results"].(map[string]any)
			if !ok {
				t.Fatalf("results = %T, want object", body["results"])
			}
			if got := results[tt.field]; got != tt.wantURL {
				t.Fatalf("results[%s] = %v, want %s", tt.field, got, tt.wantURL)
			}
			if results["apiToken"] != "synthetic-response-token" {
				t.Fatalf("apiToken = %+v, want ordinary unclassified field preserved", results)
			}
		})
	}
}
