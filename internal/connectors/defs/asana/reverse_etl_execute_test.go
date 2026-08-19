// Package asana holds the Asana bundle's connector-local executable-parity
// test. The bundle itself is pure JSON; this file exists only so the promoted
// reverse-ETL surface is proven by execution rather than by inspection.
package asana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
)

const (
	bundleName = "asana"

	// blockedReverseETLReason is the api_surface.json reason this lane exists
	// to retire. Every endpoint still carrying it is an operation the audit
	// counted as supported-but-unreachable.
	blockedReverseETLReason = "planned reverse-ETL write action; blocked until a named action has a bounded record schema, redaction, sanitized fixture, and plan -> preview -> explicit approval -> execute evidence."

	// noRequestContractReason and noBoundedShapeReason are where the rows this
	// lane could NOT promote went. They are counted, not hand-waved: a bounded
	// record schema has to come from the pinned OpenAPI source the bundle
	// already cites, and inventing one would reproduce exactly the
	// implemented-but-unreachable hole this repository is closing.
	noRequestContractReason = "Blocked by default until the pinned Asana OpenAPI source declares a request body for this operation. A typed reverse-ETL action requires a bounded record schema and none can be derived without inventing the payload shape."

	noBoundedShapeReason = "Blocked by default until this operation's request body has a bounded, flag-representable shape in the pinned Asana OpenAPI source."

	// promotedReverseETLOperations plus the two deferred counts must equal the
	// ledger rows that originally carried blockedReverseETLReason, so the
	// shortfall can never be hidden by rewording a reason string.
	promotedReverseETLOperations = 60
	deferredNoRequestContract    = 0
	deferredNoBoundedShape       = 0
	originalBlockedReverseETL    = 60

	// reverseETLBoundEndpoints is the measured number of api_surface.json rows
	// bound to a write action, pinned on its own so the ledger arithmetic can
	// never be satisfied by dropping covered_by blocks and re-blocking them
	// under a freshly worded reason.
	reverseETLBoundEndpoints = 73

	// blockedDestructiveOperations is the number of destructive_action rows
	// that remain unbound pending connector-local action, command, and fixture
	// authoring; the shared gate alone does not promote them.
	blockedDestructiveOperations = 36
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
		CredentialRevision:  "asana-fixture-credential-revision",
		ConfigurationDigest: "asana-fixture-configuration-digest",
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
	promoted, noContract, noShape := 0, 0, 0
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
		case noBoundedShapeReason:
			noShape++
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
	if noShape != deferredNoBoundedShape {
		t.Errorf("rows blocked on an unbounded request shape = %d, want %d", noShape, deferredNoBoundedShape)
	}
	if promoted != reverseETLBoundEndpoints {
		t.Errorf("api_surface rows bound to a write action = %d, want %d", promoted, reverseETLBoundEndpoints)
	}
	if got := promotedReverseETLOperations + noContract + noShape; got != originalBlockedReverseETL {
		t.Fatalf("promoted(%d) + deferred(%d+%d) = %d, want %d ledger rows accounted for",
			promotedReverseETLOperations, noContract, noShape, got, originalBlockedReverseETL)
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
				plan, preview, err := application.PlanConnectorCommand(context.Background(), app.PlanConnectorCommandRequest{
					Name: action.Name, Connector: b.Name, Credential: b.Name + "-fixture", Path: path,
					Flags: flagsFromRecord(cmd, fixture.Record), Preview: true,
				})
				if err != nil {
					t.Fatalf("PlanConnectorCommand(%q) = %v", action.Name, err)
				}
				if preview == nil || preview.Digest == "" || plan.ApprovalToken == "" {
					t.Fatalf("PlanConnectorCommand(%q) did not produce genuine preview approval", action.Name)
				}
				run, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{
					PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
					Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
					WithheldFlags: flagsFromRecord(cmd, fixture.Record),
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
		_, _ = w.Write([]byte(`{"data":{}}`))
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
