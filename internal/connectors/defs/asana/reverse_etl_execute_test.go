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
	"sort"
	"strings"
	"sync"
	"testing"

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

	// blockedDestructiveOperations is the number of destructive_action rows
	// that must remain unbound until cli-delete-confirmation-foundation-r1
	// ships the destructive-write confirm gate.
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
		Config:  map[string]string{"base_url": baseURL},
		Secrets: map[string]string{"access_token": "synthetic-test-token"},
	}
}

// TestNoReverseETLOperationRemainsBlocked fails while any endpoint still
// carries the typed reverse-ETL blocked reason. This is the RED assertion the
// authoring work turns green, and it stays as regression cover: re-blocking a
// promoted operation puts the reason back and fails here.
func TestNoReverseETLOperationRemainsBlocked(t *testing.T) {
	b := loadBundle(t)
	var blocked []string
	for _, ep := range b.Surface.Endpoints {
		if ep.Operation != nil && ep.Operation.Reason == blockedReverseETLReason {
			blocked = append(blocked, fmt.Sprintf("%s %s", ep.Method, ep.Path))
		}
	}
	if len(blocked) > 0 {
		sort.Strings(blocked)
		t.Fatalf("%d reverse-ETL operations are still blocked by an unimplemented typed write action:\n  %s",
			len(blocked), strings.Join(blocked, "\n  "))
	}
}

// TestDestructiveOperationsStayBlocked pins the destructive rows that must NOT
// be promoted before the destructive-write confirm gate exists, and requires
// every bound DELETE to carry a typed confirmation challenge.
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
		t.Fatalf("destructive_action rows still blocked = %d, want %d (promoting a destructive operation needs the destructive-write confirm gate)",
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
				assertRedactedFieldsHidden(t, action, built.RedactedRecord)
			} else if err := commandrunner.Preflight(conn, path); err == nil {
				t.Fatalf("Preflight(%q) = nil for a command marked %q, want it refused", cmd.Path, cmd.Availability)
			}

			// execute: the real engine issues the request.
			capture.Reset()
			result, err := engine.Write(context.Background(), replay,
				connectors.WriteRequest{Action: action.Name, Config: cfg},
				[]connectors.Record{connectors.Record(fixture.Record)},
				engine.HooksFor(b.Name))
			if err != nil {
				t.Fatalf("engine.Write(%q) = %v, want nil", action.Name, err)
			}
			if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
				t.Fatalf("engine.Write(%q) result = %+v, want 1 written 0 failed", action.Name, result)
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

// assertRedactedFieldsHidden proves the action's redaction declarations reach
// the operator-visible plan sample, which is the record a reviewer reads before
// approving the mutation.
func assertRedactedFieldsHidden(t *testing.T, action engine.WriteAction, redacted connectors.Record) {
	t.Helper()
	for _, field := range action.RedactFields {
		target := strings.TrimPrefix(strings.TrimSpace(field), "record.")
		value, ok := lookupRecordPath(map[string]any(redacted), strings.Split(target, "."))
		if !ok {
			continue
		}
		if s, isString := value.(string); !isString || (s != "***" && s != "redacted") {
			t.Errorf("redact_fields declares %q but the approval sample still shows %v", field, value)
		}
	}
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

func lookupRecordPath(record map[string]any, parts []string) (any, bool) {
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
