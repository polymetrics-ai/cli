// Package zendesksupport holds the Zendesk Support bundle's connector-local
// executable-parity test. The bundle itself is pure JSON; this file exists only
// so the promoted reverse-ETL surface is proven by execution rather than by
// inspection.
package zendesksupport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/warehouse"
)

const (
	bundleName = "zendesk-support"

	// blockedReverseETLReason is the api_surface.json reason this lane exists
	// to retire. Every endpoint still carrying it is an operation the audit
	// counted as supported-but-unreachable.
	blockedReverseETLReason = "Blocked by default until a connector-local typed reverse-ETL action implements plan, preview, explicit approval, and execute."

	// noRequestContractReason and the two named bulk/batch blockers are where the rows this
	// lane could NOT promote went. They are counted, not hand-waved: a bounded
	// record schema has to come from the pinned OpenAPI source the bundle
	// already cites, and inventing one would reproduce exactly the
	// implemented-but-unreachable hole this repository is closing.
	noRequestContractReason = "Blocked by default until the pinned Zendesk Support OpenAPI source declares a request body for this operation. A typed reverse-ETL action requires a bounded record schema, and https://developer.zendesk.com/zendesk/oas.yaml documents no request contract for it, so one cannot be derived without inventing the payload shape."

	ticketsBulkBatchSplitReason = "Blocked by default until the shared write surface can bind separate bulk and batch actions to this one endpoint and render record-array selectors into the ids query parameter. The pinned Zendesk Support OAS declares two closed typed oneOf arms; no free-form payload is involved."

	usersBulkBatchSplitReason = "Blocked by default until the shared write surface can bind separate bulk and batch actions to this one endpoint and render record-array selectors into ids or external_ids query parameters. The pinned Zendesk Support OAS declares two closed typed oneOf arms; no free-form payload is involved."

	// promotedReverseETLOperations plus the two deferred counts must equal the
	// ledger rows that originally carried blockedReverseETLReason, so the
	// shortfall can never be hidden by rewording a reason string.
	promotedReverseETLOperations = 62
	deferredNoRequestContract    = 98
	deferredBulkBatchSplit       = 2
	originalBlockedReverseETL    = 162

	// reverseETLBoundEndpoints is the measured number of api_surface.json rows
	// bound to a write action, pinned on its own so the ledger arithmetic can
	// never be satisfied by dropping covered_by blocks and re-blocking them
	// under a freshly worded reason.
	reverseETLBoundEndpoints = 90

	// blockedDestructiveOperations is the number of destructive_action rows
	// that remain unbound pending connector-local action, command, and fixture
	// authoring; the shared gate alone does not promote them.
	blockedDestructiveOperations = 88
)

func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(os.DirFS(".."), bundleName)
	if err != nil {
		t.Fatalf("load %s bundle: %v", bundleName, err)
	}
	return b
}

func runtimeConfig(baseURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config:              map[string]string{"base_url": baseURL},
		Secrets:             map[string]string{"access_token": "synthetic-test-token"},
		CredentialRevision:  "zendesk-support-fixture-credential-revision",
		ConfigurationDigest: "zendesk-support-fixture-configuration-digest",
		WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
	}
}

// TestReverseETLLedgerReconciles accounts for every ledger row that originally
// carried the typed reverse-ETL blocked reason. A promoted row is bound to a
// write action; a row that could not be promoted must carry one of the two
// precise, cited reasons and be counted here. Two independent things are
// pinned: the MEASURED number of endpoints bound to a write action, and the
// arithmetic that the three buckets still add up to originalBlockedReverseETL.
// Neither alone is enough - the count alone would not notice a reworded
// blocked reason, and the arithmetic alone would not notice a covered_by block
// being dropped and re-blocked under a third reason string.
func TestReverseETLLedgerReconciles(t *testing.T) {
	b := loadBundle(t)

	var stillGeneric []string
	promoted, noContract, bulkBatchSplit := 0, 0, 0
	for _, ep := range b.Surface.Endpoints {
		if ep.CoveredBy != nil && ep.CoveredBy.Write != "" {
			promoted++
			continue
		}
		if ep.Operation == nil {
			continue
		}
		switch ep.Operation.Reason {
		case blockedReverseETLReason:
			stillGeneric = append(stillGeneric, fmt.Sprintf("%s %s", ep.Method, ep.Path))
		case noRequestContractReason:
			noContract++
		case ticketsBulkBatchSplitReason, usersBulkBatchSplitReason:
			bulkBatchSplit++
		}
	}

	if len(stillGeneric) > 0 {
		sort.Strings(stillGeneric)
		t.Fatalf("%d reverse-ETL operations still carry the generic blocked reason instead of being promoted or given a cited blocker:\n  %s",
			len(stillGeneric), strings.Join(stillGeneric, "\n  "))
	}
	if noContract != deferredNoRequestContract {
		t.Errorf("rows blocked on a missing request contract = %d, want %d", noContract, deferredNoRequestContract)
	}
	if bulkBatchSplit != deferredBulkBatchSplit {
		t.Errorf("rows blocked on the named bulk/batch action split = %d, want %d", bulkBatchSplit, deferredBulkBatchSplit)
	}
	if promoted != reverseETLBoundEndpoints {
		t.Errorf("api_surface rows bound to a write action = %d, want %d", promoted, reverseETLBoundEndpoints)
	}
	if got := promotedReverseETLOperations + noContract + bulkBatchSplit; got != originalBlockedReverseETL {
		t.Fatalf("promoted(%d) + deferred(%d+%d) = %d, want %d ledger rows accounted for",
			promotedReverseETLOperations, noContract, bulkBatchSplit, got, originalBlockedReverseETL)
	}
}

// TestDestructiveOperationsStayBlocked pins the destructive rows that must NOT
// be promoted by this foundation PR, and requires every already-bound DELETE
// to carry a typed confirmation challenge through the shared gate.
func TestDestructiveOperationsStayBlocked(t *testing.T) {
	b := loadBundle(t)
	actions := map[string]engine.WriteAction{}
	for _, a := range b.Writes {
		actions[a.Name] = a
	}

	unbound := 0
	for _, ep := range b.Surface.Endpoints {
		if ep.Operation != nil && ep.Operation.Model == "destructive_action" {
			unbound++
		}
		if ep.CoveredBy == nil || ep.CoveredBy.Write == "" {
			continue
		}
		if !strings.EqualFold(ep.Method, http.MethodDelete) {
			continue
		}
		action, ok := actions[ep.CoveredBy.Write]
		if !ok {
			t.Errorf("endpoint %s %s is covered by unknown write %q", ep.Method, ep.Path, ep.CoveredBy.Write)
			continue
		}
		if strings.TrimSpace(action.Confirm) == "" {
			t.Errorf("bound DELETE endpoint %s %s uses write %q with no typed confirmation challenge",
				ep.Method, ep.Path, action.Name)
		}
	}
	if unbound != blockedDestructiveOperations {
		t.Fatalf("destructive_action rows still unbound = %d, want %d (connector binding is outside this foundation PR)",
			unbound, blockedDestructiveOperations)
	}
}

// TestReverseETLWriteActionsExecute drives every declared write action through
// the surface an operator actually uses: the real commandrunner resolves the
// command, builds the record from the command's own flags, produces a preview,
// and the real engine issues the HTTP request against a capture server. It
// asserts execution, not declaration.
func TestReverseETLWriteActionsExecute(t *testing.T) {
	b := loadBundle(t)
	if len(b.Writes) == 0 {
		t.Fatalf("bundle declares no write actions")
	}

	capture := newCaptureServer()
	defer capture.Close()

	cfg := runtimeConfig(capture.URL)
	replay := b
	conn := engine.New(replay, engine.HooksFor(b.Name))
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() = %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	if _, err := application.AddCredential(context.Background(), app.AddCredentialRequest{
		Name: b.Name + "-fixture", Connector: b.Name,
		Config: map[string]string{"base_url": capture.URL}, Secrets: map[string]string{"access_token": "synthetic-test-token"},
	}); err != nil {
		t.Fatalf("AddCredential(%s) = %v", b.Name, err)
	}

	commandsByWrite := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range conn.CommandSurface().Commands {
		if cmd.Write == "" {
			continue
		}
		if prior, ok := commandsByWrite[cmd.Write]; ok {
			t.Fatalf("write %q is referenced by two commands (%q and %q)", cmd.Write, prior.Path, cmd.Path)
		}
		commandsByWrite[cmd.Write] = cmd
	}

	for _, action := range b.Writes {
		t.Run(action.Name, func(t *testing.T) {
			cmd, ok := commandsByWrite[action.Name]
			if !ok {
				t.Fatalf("write action %q has no cli_surface command", action.Name)
			}
			fixture := loadWriteFixture(t, action.Name)
			path := strings.Fields(cmd.Path)

			// An "implemented" command is a promise the runtime keeps: it must
			// resolve, build a record from its own flags, and stage a preview
			// for approval. A command still marked "partial" must be refused,
			// so the availability field can never overstate reachability.
			if cmd.Availability == "implemented" {
				if err := commandrunner.Preflight(conn, path); err != nil {
					t.Fatalf("Preflight(%q) = %v, want nil", cmd.Path, err)
				}
				built, err := commandrunner.BuildWriteCommand(context.Background(), conn, commandrunner.Request{
					Path:    path,
					Flags:   flagsFromRecord(cmd, fixture.Record),
					Config:  cfg,
					Preview: true,
				})
				if err != nil {
					t.Fatalf("BuildWriteCommand(%q) = %v, want nil", cmd.Path, err)
				}
				if !built.ApprovalRequired {
					t.Fatalf("command %q does not require approval", cmd.Path)
				}
				if built.Preview == nil || built.Preview.RecordsStaged != 1 {
					t.Fatalf("command %q preview = %+v, want 1 staged record", cmd.Path, built.Preview)
				}
				assertRedactFieldsLoadCompatible(t, action, cmd, fixture.Record)
				assertCommandRecordPreserved(t, built.Record, built.RedactedRecord)
			} else if err := commandrunner.Preflight(conn, path); err == nil {
				t.Fatalf("Preflight(%q) = nil for a command marked %q, want it refused", cmd.Path, cmd.Availability)
			}

			// execute: the real engine issues the request.
			capture.Reset()
			records := []connectors.Record{connectors.Record(fixture.Record)}
			hooks := engine.HooksFor(b.Name)
			written, failed := 0, 0
			if engine.DestructiveTargetForWrite(b.Name, action).RequiresApproval() {
				table := "fixture_" + action.Name
				// Written through the real Parquet writer: a hand-written JSONL
				// table is a format pm refuses, so the fixture would fail the
				// reverse plan before the action under test ran at all.
				warehousePath := filepath.Join(application.ProjectDir(), "warehouse", table+warehouse.TableFileExt)
				if err := warehouse.WriteTable(context.Background(), warehousePath, []warehouse.Row{warehouse.Row(fixture.Record)}); err != nil {
					t.Fatalf("WriteTable(%q fixture) = %v", action.Name, err)
				}
				mappings := make(map[string]string, len(fixture.Record))
				for field := range fixture.Record {
					mappings[field] = field
				}
				plan, err := application.PlanReverseETL(context.Background(), app.PlanReverseETLRequest{
					Name: action.Name, SourceTable: table, DestinationConnector: b.Name,
					DestinationCredential: b.Name + "-fixture", Action: action.Name, Mappings: mappings,
				})
				if err != nil {
					t.Fatalf("PlanReverseETL(%q) = %v", action.Name, err)
				}
				plan, preview, err := application.PreviewReversePlan(context.Background(), plan.ID, nil)
				if err != nil {
					t.Fatalf("PreviewReversePlan(%q) = %v", action.Name, err)
				}
				if preview.Digest == "" || plan.ApprovalToken == "" {
					t.Fatalf("PreviewReversePlan(%q) did not produce genuine preview approval", action.Name)
				}
				run, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{
					PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
					Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
				})
				if err != nil {
					t.Fatalf("RunReverseETL(%q) = %v", action.Name, err)
				}
				written, failed = run.RecordsSucceeded, run.RecordsFailed
			} else {
				result, err := engine.Write(context.Background(), replay, connectors.WriteRequest{Action: action.Name, Config: cfg}, records, hooks)
				if err != nil {
					t.Fatalf("engine.Write(%q) = %v, want nil", action.Name, err)
				}
				written, failed = result.RecordsWritten, result.RecordsFailed
			}
			if written != 1 || failed != 0 {
				t.Fatalf("write(%q) result = %d written %d failed, want 1 written 0 failed", action.Name, written, failed)
			}
			got := capture.Last()
			if got == nil {
				t.Fatalf("engine.Write(%q) sent no HTTP request", action.Name)
			}
			if !strings.EqualFold(got.Method, fixture.Expect.Method) {
				t.Fatalf("method = %q, want %q", got.Method, fixture.Expect.Method)
			}
			if got.Path != fixture.Expect.Path {
				t.Fatalf("path = %q, want %q", got.Path, fixture.Expect.Path)
			}
			for key, want := range fixture.Expect.Body {
				val, ok := got.Body[key]
				if !ok {
					t.Fatalf("request body is missing key %q", key)
				}
				if fmt.Sprint(val) != fmt.Sprint(want) {
					t.Fatalf("request body[%q] = %v, want %v", key, val, want)
				}
			}
		})
	}
}

func TestZendeskFoundationWriteBoundsAreEnforced(t *testing.T) {
	b := loadBundle(t)
	tests := []struct {
		name   string
		action string
		record connectors.Record
	}{
		{
			name:   "create users maxItems 100",
			action: "create_many_users",
			record: connectors.Record{"users": repeatedObjects(101, func(i int) map[string]any { return map[string]any{"name": fmt.Sprintf("Fixture %d", i)} })},
		},
		{
			name:   "upsert users maxItems 100",
			action: "create_or_update_many_users",
			record: connectors.Record{"users": repeatedObjects(101, func(i int) map[string]any { return map[string]any{"external_id": fmt.Sprintf("fixture-%d", i)} })},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateWrite(context.Background(), b, connectors.WriteRequest{Action: tt.action}, []connectors.Record{tt.record})
			if err == nil {
				t.Fatalf("ValidateWrite(%q) = nil for 101 items", tt.action)
			}
		})
	}
}

// UserMergeInput deliberately has no required property in the pinned OAS:
// callers can identify an existing user by id, email, or external_id. The
// selected merge arm must retain that flexibility rather than inventing an
// external_id-only requirement while making the shape closed.
func TestZendeskFoundationMergeUserAcceptsOASIdentifierVariants(t *testing.T) {
	b := loadBundle(t)
	for name, user := range map[string]map[string]any{
		"id":    {"id": 101},
		"email": {"email": "fixture@example.invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			err := engine.ValidateWrite(context.Background(), b, connectors.WriteRequest{Action: "create_or_update_many_users"}, []connectors.Record{{"users": []any{user}}})
			if err != nil {
				t.Fatalf("ValidateWrite(%q) = %v, want UserMergeInput identifier accepted", name, err)
			}
		})
	}
}

func repeatedObjects(count int, build func(int) map[string]any) []any {
	objects := make([]any, 0, count)
	for i := 0; i < count; i++ {
		objects = append(objects, build(i))
	}
	return objects
}

func TestZendeskFoundationRequestFieldsCitePinnedOAS(t *testing.T) {
	b := loadBundle(t)
	foundation := map[string]bool{
		"update_permission_policy":            true,
		"bulk_set_agent_attribute_values_job": true,
		"create_many_users":                   true,
		"create_or_update_many_users":         true,
		"request_user_create":                 true,
	}
	for _, action := range b.Writes {
		if !foundation[action.Name] {
			continue
		}
		var schema map[string]any
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("unmarshal %q record schema: %v", action.Name, err)
		}
		assertSchemaFieldsCitePinnedOAS(t, action.Name, schema)
	}
}

// TestZendeskFoundationOperationRequestFieldsCitePinnedOAS keeps the
// operation ledger as auditable as the executable write definitions. The
// operations remain the canonical per-endpoint request contracts even after a
// write action has been promoted, so a field must not be cited in one place
// and become anonymous in the other.
func TestZendeskFoundationOperationRequestFieldsCitePinnedOAS(t *testing.T) {
	b := loadBundle(t)
	foundation := map[string]bool{
		"zendesk-support.update_permission_policy":            true,
		"zendesk-support.bulk_set_agent_attribute_values_job": true,
		"zendesk-support.create_many_users":                   true,
		"zendesk-support.create_or_update_many_users":         true,
		"zendesk-support.request_user_create":                 true,
	}
	for _, operation := range b.Operations {
		if !foundation[operation.ID] {
			continue
		}
		if operation.REST == nil {
			t.Errorf("foundation operation %q has no REST request contract", operation.ID)
			continue
		}
		var schema map[string]any
		if err := json.Unmarshal(operation.REST.BodySchema, &schema); err != nil {
			t.Fatalf("unmarshal %q body schema: %v", operation.ID, err)
		}
		assertSchemaFieldsCitePinnedOAS(t, operation.ID, schema)
	}
}

// TestZendeskFoundationOperationBodiesMatchExecutableSchemas keeps the OAS
// operation ledger and the action the operator can invoke on one contract.
// Path parameters are record fields for the write action but not request-body
// fields, so the permission-policy action is the one intentional exception.
func TestZendeskFoundationOperationBodiesMatchExecutableSchemas(t *testing.T) {
	b := loadBundle(t)
	actions := map[string]engine.WriteAction{}
	for _, action := range b.Writes {
		actions[action.Name] = action
	}
	pairs := map[string]string{
		"zendesk-support.update_permission_policy":            "update_permission_policy",
		"zendesk-support.bulk_set_agent_attribute_values_job": "bulk_set_agent_attribute_values_job",
		"zendesk-support.create_many_users":                   "create_many_users",
		"zendesk-support.create_or_update_many_users":         "create_or_update_many_users",
		"zendesk-support.request_user_create":                 "request_user_create",
	}
	for _, operation := range b.Operations {
		actionName, ok := pairs[operation.ID]
		if !ok {
			continue
		}
		action, ok := actions[actionName]
		if !ok {
			t.Fatalf("operation %q refers to missing action %q", operation.ID, actionName)
		}
		if operation.REST == nil {
			t.Fatalf("operation %q has no REST contract", operation.ID)
		}
		var want, got map[string]any
		if err := json.Unmarshal(action.RecordSchema, &want); err != nil {
			t.Fatalf("unmarshal %q record schema: %v", actionName, err)
		}
		if err := json.Unmarshal(operation.REST.BodySchema, &got); err != nil {
			t.Fatalf("unmarshal %q body schema: %v", operation.ID, err)
		}
		delete(want, "$schema")
		delete(got, "title")
		if actionName == "update_permission_policy" {
			properties := want["properties"].(map[string]any)
			delete(properties, "custom_object_key")
			delete(properties, "id")
			delete(want, "required")
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("operation %q body schema differs from executable action %q", operation.ID, actionName)
		}
	}
}

func assertSchemaFieldsCitePinnedOAS(t *testing.T, path string, schema map[string]any) {
	t.Helper()
	if properties, ok := schema["properties"].(map[string]any); ok {
		for name, rawChild := range properties {
			child, ok := rawChild.(map[string]any)
			if !ok {
				t.Errorf("%s.%s is not an object schema", path, name)
				continue
			}
			if description, _ := child["description"].(string); !strings.Contains(description, "https://developer.zendesk.com/zendesk/oas.yaml#") {
				t.Errorf("%s.%s lacks a pinned OAS citation", path, name)
			}
			assertSchemaFieldsCitePinnedOAS(t, path+"."+name, child)
		}
	}
	if items, ok := schema["items"].(map[string]any); ok {
		assertSchemaFieldsCitePinnedOAS(t, path+"[]", items)
	}
}

func assertRedactFieldsLoadCompatible(t *testing.T, action engine.WriteAction, cmd connectors.CommandSurfaceCommand, record map[string]any) {
	t.Helper()
	for _, field := range cmd.RedactFields {
		if strings.TrimSpace(field) == "" {
			t.Errorf("command %q loaded an empty redact field", cmd.Path)
		}
	}
	for _, field := range action.RedactFields {
		target := strings.TrimPrefix(strings.TrimSpace(field), "record.")
		if _, ok := lookupMapPath(record, strings.Split(target, ".")); !ok {
			t.Errorf("write action %q loaded redact field %q that cannot resolve in the sanitized record", action.Name, field)
		}
	}
}

func assertCommandRecordPreserved(t *testing.T, original, preserved connectors.Record) {
	t.Helper()
	if !reflect.DeepEqual(original, preserved) {
		t.Errorf("command record = %#v, want complete record %#v", preserved, original)
	}
}

// lookupMapPath mirrors the engine's record-path resolver exactly: it descends
// maps only and has no array-index support, so an array-nested declaration
// fails here for the same reason it is inert in the engine.
func lookupMapPath(record map[string]any, parts []string) (any, bool) {
	var current any = record
	for _, part := range parts {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

type writeFixture struct {
	Record map[string]any `json:"record"`
	Expect struct {
		Method string         `json:"method"`
		Path   string         `json:"path"`
		Body   map[string]any `json:"body,omitempty"`
	} `json:"expect"`
}

func loadWriteFixture(t *testing.T, action string) writeFixture {
	t.Helper()
	raw, err := os.ReadFile("fixtures/writes/" + action + ".json")
	if err != nil {
		t.Fatalf("read sanitized fixture for %q: %v", action, err)
	}
	var fx writeFixture
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&fx); err != nil {
		t.Fatalf("parse fixture for %q: %v", action, err)
	}
	return fx
}

// flagsFromRecord reverse-maps a fixture record onto the command's declared
// flags, so BuildWriteCommand exercises the same flag plumbing the CLI uses.
func flagsFromRecord(cmd connectors.CommandSurfaceCommand, record map[string]any) map[string][]string {
	flags := map[string][]string{}
	for _, flag := range cmd.Flags {
		target, ok := strings.CutPrefix(flag.MapsTo, "record.")
		if !ok || target == "" {
			continue
		}
		value, found := lookupRecordPath(record, strings.Split(target, "."))
		if !found || value == nil {
			continue
		}
		switch typed := value.(type) {
		case []any:
			var parts []string
			for _, item := range typed {
				if _, isObject := item.(map[string]any); isObject {
					parts = nil
					break
				}
				parts = append(parts, fmt.Sprint(item))
			}
			if len(parts) > 0 {
				flags[flag.Name] = []string{strings.Join(parts, ",")}
			}
		case map[string]any:
			// Objects have no flag representation; the record carries them.
		default:
			flags[flag.Name] = []string{fmt.Sprint(typed)}
		}
	}
	return flags
}

// lookupRecordPath walks a dotted record path, descending through array
// indices the same way commandrunner's record-path builder does, so a flag
// mapped to record.users.0.email resolves against the fixture.
func lookupRecordPath(record map[string]any, parts []string) (any, bool) {
	var current any = record
	for _, part := range parts {
		if index, err := strconv.Atoi(part); err == nil {
			items, ok := current.([]any)
			if !ok || index < 0 || index >= len(items) {
				return nil, false
			}
			current = items[index]
			continue
		}
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

type capturedRequest struct {
	Method string
	Path   string
	Body   map[string]any
}

type captureServer struct {
	*httptest.Server
	mu   sync.Mutex
	last *capturedRequest
}

func newCaptureServer() *captureServer {
	c := &captureServer{}
	c.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		captured := &capturedRequest{Method: r.Method, Path: r.URL.Path}
		if len(raw) > 0 {
			var body map[string]any
			dec := json.NewDecoder(bytes.NewReader(raw))
			dec.UseNumber()
			if err := dec.Decode(&body); err == nil {
				captured.Body = body
			}
		}
		c.mu.Lock()
		c.last = captured
		c.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	return c
}

func (c *captureServer) Reset() {
	c.mu.Lock()
	c.last = nil
	c.mu.Unlock()
}

func (c *captureServer) Last() *capturedRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.last
}
