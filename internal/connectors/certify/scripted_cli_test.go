package certify_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// scriptedCLI is a deterministic test-only implementation of the CLI seam.
// It returns complete, synthetic envelopes while validating every command
// family and its required arguments. It never calls cli.Run, a provider, or
// credentials. The real full-sweep and sample/outbox write-lifecycle proofs
// remain separately counted by TestMain.
type scriptedCLI struct {
	t *testing.T

	root         string
	calls        [][]string
	seen         map[string]int
	plans        map[string]scriptedPlan
	previewed    map[string]bool
	consumed     map[string]bool
	replays      map[string]int
	connections  map[string]string
	nextPlan     int
	schedule     string
	scheduleFlow string
	protocols    []string
}

type scriptedPlan struct {
	action        string
	approvalToken string
}

func newScriptedCLI(t *testing.T, protocols ...string) *scriptedCLI {
	t.Helper()
	return &scriptedCLI{
		t:           t,
		seen:        make(map[string]int),
		plans:       make(map[string]scriptedPlan),
		previewed:   make(map[string]bool),
		consumed:    make(map[string]bool),
		replays:     make(map[string]int),
		connections: make(map[string]string),
		protocols:   protocols,
	}
}

// scriptedSampleRunner replaces duplicated real CLI runs in the focused
// stage-contract tests. The retained full-sweep and sample/outbox write-
// lifecycle tests are intentionally not routed through this helper: they
// remain the package's real CLI proofs.
func scriptedSampleRunner(t *testing.T, opts certify.Options) (*certify.Runner, *scriptedCLI) {
	t.Helper()
	driver := newScriptedCLI(t, scriptedProtocols(opts)...)
	driver.install(t)
	return certify.NewRunner(opts), driver
}

func scriptedProtocols(opts certify.Options) []string {
	protocols := []string{
		"init",
		"connectors_list",
		"connectors_inspect",
		"credentials_test",
		"connections_create",
		"catalog_refresh",
		"etl_run",
		"query_run",
		"flow_plan",
		"flow_preview",
		"flow_run",
		"flow_status",
		"schedule_create",
		"schedule_list",
		"schedule_install",
		"schedule_fire",
		"schedule_remove",
	}
	if opts.Write {
		protocols = append(protocols,
			"reverse_plan",
			"reverse_preview",
			"reverse_run",
			"reverse_replay",
		)
	}
	return protocols
}

func (s *scriptedCLI) install(t *testing.T) {
	t.Helper()
	certify.SetCLIRunFunc(s.run)
	t.Cleanup(func() { certify.SetCLIRunFunc(countedCertifyCLIRun) })
}

func (s *scriptedCLI) run(args []string, stdout, stderr io.Writer) int {
	args, root, ok := s.withoutRoot(args)
	if !ok {
		_, _ = fmt.Fprintln(stderr, "scripted certify CLI: missing or inconsistent --root")
		return 1
	}
	s.calls = append(s.calls, append([]string(nil), args...))

	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, "scripted certify CLI: empty command")
		return 1
	}

	switch {
	case exact(args, "init", "--json"):
		s.seen["init"]++
		return writeScriptedEnvelope(stdout, "InitResult", nil)
	case exact(args, "connectors", "list", "--json"):
		s.seen["connectors_list"]++
		return writeScriptedEnvelope(stdout, "ConnectorList", map[string]any{
			"connectors": []map[string]any{{"name": "sample"}},
		})
	case prefix(args, "connectors", "inspect") && hasJSON(args) && len(args) == 4 && args[2] == "sample":
		s.seen["connectors_inspect"]++
		return writeScriptedEnvelope(stdout, "Connector", map[string]any{"connector": "sample"})
	case prefix(args, "credentials", "add") && hasJSON(args):
		if !hasFlag(args, "--connector") {
			return s.protocolError(stderr, "credentials add requires --connector")
		}
		s.seen["credentials_add"]++
		return writeScriptedEnvelope(stdout, "Credential", nil)
	case exact(args, "credentials", "test", "cert-source", "--json"):
		s.seen["credentials_test"]++
		return writeScriptedEnvelope(stdout, "CredentialTest", nil)
	case prefix(args, "connections", "create") && hasJSON(args):
		if len(args) < 3 || !hasFlag(args, "--source") || !hasFlag(args, "--destination") || !hasFlag(args, "--stream") || !hasFlag(args, "--table") || !hasFlag(args, "--sync-mode") {
			return s.protocolError(stderr, "connections create requires source, destination, stream, table, and sync mode")
		}
		s.connections[args[2]] = flagValue(args, "--sync-mode")
		s.seen["connections_create"]++
		return writeScriptedEnvelope(stdout, "Connection", nil)
	case prefix(args, "catalog", "refresh") && hasJSON(args) && hasFlag(args, "--connection"):
		s.seen["catalog_refresh"]++
		return writeScriptedEnvelope(stdout, "Catalog", map[string]any{
			"catalog": map[string]any{"catalog": map[string]any{"streams": []map[string]any{
				{"name": "customers", "primary_key": []string{"id"}, "cursor_fields": []string{"updated_at"}},
				{"name": "events", "primary_key": []string{"id"}, "cursor_fields": []string{"updated_at"}},
			}}},
		})
	case prefix(args, "etl", "run") && hasJSON(args) && hasFlag(args, "--connection") && hasFlag(args, "--stream"):
		s.seen["etl_run"]++
		if isTypedCompatibilityRefusal(s.connections[flagValue(args, "--connection")]) {
			return writeScriptedEnvelopeWithCode(stdout, "ETLRun", map[string]any{
				"run": map[string]any{
					"id":           "run_certification_refusal",
					"status":       "failed",
					"completed_at": "2026-08-21T00:00:00Z",
					"error":        "sync mode is not executable",
				},
			}, 1)
		}
		return writeScriptedEnvelope(stdout, "ETLRun", map[string]any{"run": map[string]any{
			"records_read":      2,
			"records_succeeded": 1,
			"records_failed":    0,
			"checkpoint":        map[string]any{"cursor": "2026-08-06T00:00:00Z"},
		}})
	case prefix(args, "query", "run") && hasJSON(args) && hasFlag(args, "--table"):
		table := flagValue(args, "--table")
		s.seen["query_run"]++
		if strings.HasPrefix(table, "cert_flow_query_") {
			return writeScriptedEnvelopeWithCode(stdout, "Error", map[string]any{"error": "synthetic query table is absent"}, 2)
		}
		return writeScriptedEnvelope(stdout, "QueryResult", map[string]any{
			"count": 2,
			"rows":  []map[string]any{{"id": "synthetic-1"}, {"id": "synthetic-2"}},
		})
	case prefix(args, "flow", "plan") && hasJSON(args) && hasFlag(args, "--file"):
		s.seen["flow_plan"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"status": "ok", "order": []string{"cert_sync", "cert_query"}})
	case prefix(args, "flow", "preview") && hasJSON(args) && hasFlag(args, "--file"):
		s.seen["flow_preview"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"status": "dry_run"})
	case prefix(args, "flow", "run") && hasJSON(args) && hasFlag(args, "--file") && hasFlag(args, "--flows-dir"):
		s.seen["flow_run"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"status": "ok", "steps": []map[string]any{
			{"id": "cert_sync", "status": "ok"}, {"id": "cert_query", "status": "ok"},
		}})
	case prefix(args, "flow", "status") && hasJSON(args) && hasFlag(args, "--flows-dir"):
		s.seen["flow_status"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"steps": []map[string]any{
			{"id": "cert_sync", "status": "success"}, {"id": "cert_query", "status": "success"},
		}})
	case prefix(args, "schedule", "create") && hasJSON(args) && hasFlag(args, "--name") && hasFlag(args, "--cron") && hasFlag(args, "--flow"):
		if hasFlag(args, "--authorization") {
			return s.protocolError(stderr, "schedule create must inherit flow approval without an authorization carrier")
		}
		s.schedule = flagValue(args, "--name")
		s.scheduleFlow = flagValue(args, "--flow")
		s.seen["schedule_create"]++
		return writeScriptedEnvelope(stdout, "Schedule", map[string]any{"name": s.schedule})
	case exact(args, "schedule", "list", "--json"):
		if s.schedule == "" {
			return s.protocolError(stderr, "schedule list ran before schedule create")
		}
		s.seen["schedule_list"]++
		return writeScriptedEnvelope(stdout, "ScheduleList", map[string]any{"schedules": []map[string]any{{"name": s.schedule}}})
	case prefix(args, "schedule", "install") && hasJSON(args) && hasArg(args, "--crontab"):
		if err := s.writeCrontabSentinel(); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["schedule_install"]++
		return writeScriptedEnvelope(stdout, "ScheduleInstall", map[string]any{"backend": "crontab", "schedule": map[string]any{"name": s.schedule}})
	case prefix(args, "schedule", "fire") && hasJSON(args) && len(args) == 4 && args[2] == s.schedule:
		s.seen["schedule_fire"]++
		return writeScriptedEnvelope(stdout, "ScheduleFire", map[string]any{
			"schedule": map[string]any{"name": s.schedule},
			"status":   map[string]any{"status": "succeeded"},
			"flow":     map[string]any{"name": s.scheduleFlow, "status": "ok"},
		})
	case prefix(args, "schedule", "remove") && hasJSON(args) && hasArg(args, "--crontab"):
		if err := s.clearCrontab(); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["schedule_remove"]++
		return writeScriptedEnvelope(stdout, "ScheduleRemove", map[string]any{"name": s.schedule})
	case prefix(args, "reverse", "plan") && !hasJSON(args):
		if !hasFlag(args, "--source-table") || !hasFlag(args, "--destination") || !hasFlag(args, "--action") {
			return s.protocolError(stderr, "reverse plan requires source table, destination, and action")
		}
		action := flagValue(args, "--action")
		if action != "create" && action != "delete" {
			return s.protocolError(stderr, "sample write lifecycle requires create or delete action")
		}
		s.nextPlan++
		planID := fmt.Sprintf("scripted-plan-%d", s.nextPlan)
		approvalToken := fmt.Sprintf("scripted-approval-%d", s.nextPlan)
		s.plans[planID] = scriptedPlan{action: action, approvalToken: approvalToken}
		s.seen["reverse_plan"]++
		if action == "delete" {
			_, _ = fmt.Fprintf(stdout, "Created reverse plan %s\nPreview required before an approval token is issued.\n", planID)
			return 0
		}
		_, _ = fmt.Fprintf(stdout, "Created reverse plan %s\nApproval token: %s\n", planID, approvalToken)
		return 0
	case prefix(args, "reverse", "preview") && hasJSON(args) && len(args) == 4:
		if _, ok := s.plans[args[2]]; !ok {
			return s.protocolError(stderr, "reverse preview received an unknown plan")
		}
		s.previewed[args[2]] = true
		s.seen["reverse_preview"]++
		return writeScriptedEnvelope(stdout, "ReversePlanPreview", map[string]any{"plan": map[string]any{"id": args[2], "records": 1}})
	case prefix(args, "reverse", "preview") && !hasJSON(args) && len(args) == 3:
		plan, ok := s.plans[args[2]]
		if !ok {
			return s.protocolError(stderr, "reverse preview received an unknown plan")
		}
		s.previewed[args[2]] = true
		s.seen["reverse_preview"]++
		_, _ = fmt.Fprintf(stdout, "Reverse plan %s previewed\nApproval token: %s\n", args[2], plan.approvalToken)
		return 0
	case prefix(args, "reverse", "run") && hasJSON(args) && hasFlag(args, "--approval-token-stdin"):
		return s.runReverse(args, root, stdout, stderr)
	default:
		return s.protocolError(stderr, "unexpected command: "+strings.Join(args, " "))
	}
}

func isTypedCompatibilityRefusal(mode string) bool {
	return mode == "full_refresh_overwrite_deduped" || mode == "incremental_append_deduped"
}

func (s *scriptedCLI) runReverse(args []string, root string, stdout, stderr io.Writer) int {
	if len(args) < 3 {
		return s.protocolError(stderr, "reverse run missing plan id")
	}
	planID := args[2]
	plan, ok := s.plans[planID]
	if !ok {
		return s.protocolError(stderr, "reverse run received an unknown plan")
	}
	if !s.previewed[planID] {
		return s.protocolError(stderr, "reverse run occurred before its preview")
	}
	if plan.action == "delete" && flagValue(args, "--confirm") != "destructive" {
		return s.protocolError(stderr, "destructive cleanup requires --confirm destructive")
	}
	if s.consumed[planID] {
		s.replays[planID]++
		if s.replays[planID] > 1 {
			return s.protocolError(stderr, "reverse run retried a consumed approval more than once")
		}
		s.seen["reverse_replay"]++
		return writeScriptedEnvelopeWithCode(stdout, "Error", map[string]any{"error": "synthetic consumed approval"}, 2)
	}
	if err := s.appendOutboxAction(root, plan.action); err != nil {
		return s.protocolError(stderr, err.Error())
	}
	s.consumed[planID] = true
	s.seen["reverse_run"]++
	return writeScriptedEnvelope(stdout, "ReverseRun", map[string]any{"run": map[string]any{"records_succeeded": 1, "records_failed": 0}})
}

func (s *scriptedCLI) withoutRoot(args []string) ([]string, string, bool) {
	var root string
	clean := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" && i+1 < len(args) {
			root = args[i+1]
			i++
			continue
		}
		clean = append(clean, args[i])
	}
	if root == "" {
		return nil, "", false
	}
	if s.root == "" {
		s.root = root
	} else if root != s.root {
		return nil, "", false
	}
	return clean, root, true
}

func (s *scriptedCLI) appendOutboxAction(root, action string) error {
	seed := filepath.Join(root, "cert_write_seed_sample_seed.jsonl")
	raw, err := os.ReadFile(seed)
	if err != nil {
		return fmt.Errorf("read scripted write seed: %w", err)
	}
	var row map[string]any
	if err := json.Unmarshal(bytesTrimSpace(raw), &row); err != nil {
		return fmt.Errorf("parse scripted write seed: %w", err)
	}
	tag, _ := row["tag"].(string)
	if tag == "" {
		return fmt.Errorf("scripted write seed did not contain tag")
	}
	path := filepath.Join(root, ".polymetrics", "outbox", "cert_write_selftest.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create scripted outbox: %w", err)
	}
	encoded, err := json.Marshal(map[string]any{"tag": tag, "_outbox_action": action})
	if err != nil {
		return fmt.Errorf("encode scripted outbox action: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open scripted outbox: %w", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(encoded, '\n')); err != nil {
		return fmt.Errorf("append scripted outbox: %w", err)
	}
	return nil
}

func (s *scriptedCLI) writeCrontabSentinel() error {
	path := os.Getenv("PM_CRONTAB_FILE")
	if path == "" || s.schedule == "" || s.scheduleFlow == "" || s.root == "" {
		return fmt.Errorf("scripted schedule install missing crontab path, schedule name, flow, or root")
	}
	line := "0 3 * * *  pm --root " + s.root + " flow run " + s.scheduleFlow + " --json  # pm-schedule-" + s.schedule + "\n"
	return os.WriteFile(path, []byte(line), 0o600)
}

func (s *scriptedCLI) clearCrontab() error {
	path := os.Getenv("PM_CRONTAB_FILE")
	if path == "" {
		return fmt.Errorf("scripted schedule remove missing crontab path")
	}
	return os.WriteFile(path, nil, 0o600)
}

func (s *scriptedCLI) protocolError(stderr io.Writer, message string) int {
	_, _ = fmt.Fprintln(stderr, "scripted certify CLI:", message)
	return 1
}

func (s *scriptedCLI) assertProtocol(t *testing.T) {
	t.Helper()
	if s.root == "" {
		t.Fatal("scripted CLI did not receive a harness-injected root")
	}
	for _, protocol := range s.protocols {
		if s.seen[protocol] == 0 {
			t.Errorf("scripted CLI never received required %s command; calls=%v", protocol, s.calls)
		}
	}
	if containsProtocol(s.protocols, "reverse_plan") {
		if got := s.seen["reverse_plan"]; got != 2 {
			t.Errorf("scripted write lifecycle planned %d times, want exactly create+cleanup plans", got)
		}
		if got := s.seen["reverse_run"]; got != 2 {
			t.Errorf("scripted write lifecycle executed %d fresh plans, want exactly create+cleanup", got)
		}
		if got := s.seen["reverse_preview"]; got != 2 {
			t.Errorf("scripted write lifecycle previewed %d plans, want exactly create+cleanup", got)
		}
		if got := s.seen["reverse_replay"]; got != 1 {
			t.Errorf("scripted write lifecycle replayed consumed approval %d times, want exactly one negative replay", got)
		}
	}
}

func containsProtocol(protocols []string, want string) bool {
	for _, protocol := range protocols {
		if protocol == want {
			return true
		}
	}
	return false
}

func writeScriptedEnvelope(stdout io.Writer, kind string, fields map[string]any) int {
	return writeScriptedEnvelopeWithCode(stdout, kind, fields, 0)
}

func writeScriptedEnvelopeWithCode(stdout io.Writer, kind string, fields map[string]any, code int) int {
	envelope := make(map[string]any, len(fields)+1)
	for key, value := range fields {
		envelope[key] = value
	}
	envelope["kind"] = kind
	if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
		return 1
	}
	return code
}

func writeScriptedEnvelopeWithoutKind(stdout io.Writer, fields map[string]any) int {
	if err := json.NewEncoder(stdout).Encode(fields); err != nil {
		return 1
	}
	return 0
}

func exact(args []string, want ...string) bool {
	if len(args) != len(want) {
		return false
	}
	for i := range want {
		if args[i] != want[i] {
			return false
		}
	}
	return true
}

func prefix(args []string, want ...string) bool {
	if len(args) < len(want) {
		return false
	}
	for i := range want {
		if args[i] != want[i] {
			return false
		}
	}
	return true
}

func hasJSON(args []string) bool { return hasArg(args, "--json") }

func hasArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func hasFlag(args []string, want string) bool { return flagValue(args, want) != "" }

func flagValue(args []string, want string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == want {
			return args[i+1]
		}
	}
	return ""
}

func bytesTrimSpace(raw []byte) []byte { return []byte(strings.TrimSpace(string(raw))) }

func TestScriptedCLIDriverRejectsProtocolDrift(t *testing.T) {
	driver := newScriptedCLI(t)
	var stdout, stderr strings.Builder
	if code := driver.run([]string{"connections", "create", "missing-required-flags", "--json", "--root", t.TempDir()}, &stdout, &stderr); code == 0 {
		t.Fatal("scripted driver accepted a connection command missing required flags")
	}
	if !strings.Contains(stderr.String(), "requires source") {
		t.Fatalf("protocol error = %q, want required-argument diagnostic", stderr.String())
	}
}
