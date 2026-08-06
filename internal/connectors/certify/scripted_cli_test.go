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

	root                  string
	calls                 [][]string
	seen                  map[string]int
	plans                 map[string]scriptedReversePlan
	credentials           map[string]scriptedCredential
	connections           map[string]scriptedConnection
	tables                map[string]scriptedTable
	previewed             map[string]bool
	consumed              map[string]bool
	replays               map[string]int
	nextPlan              int
	schedule              string
	scheduleCron          string
	scheduleFlow          string
	installedScheduleCron string
	installedScheduleFlow string
	installedScheduleGap  string
	installedScheduleRoot string
	installedScheduleBin  string
	installed             bool
	planFlow              string
	previewFlow           string
	flowRunSteps          []any
	flowStatusSteps       []any
	statusFlow            string
	reversePlanOutput     string
	previewOutputLeak     string
	protocols             []string
}

type scriptedReversePlan struct {
	action        string
	approvalToken string
	sourceTable   string
	destination   scriptedEndpoint
}

type scriptedCredential struct {
	connector string
	config    map[string]string
}

type scriptedEndpoint struct {
	connector  string
	credential string
}

type scriptedConnection struct {
	source      scriptedEndpoint
	destination scriptedEndpoint
	stream      string
	syncMode    string
	table       string
}

type scriptedTable struct {
	rows []map[string]any
}

const (
	scriptedWritePlanName    = "cert_write_selftest"
	scriptedWriteSourceTable = "cert_write_seed_sample"
	scriptedWriteDestination = "outbox:cert-outbox"
	scriptedScheduleName     = "cert-schedule-sample"
	scriptedScheduleCron     = "0 3 * * *"
	scriptedFlowName         = "cert_flow_sample"
)

func newScriptedCLI(t *testing.T, protocols ...string) *scriptedCLI {
	t.Helper()
	return &scriptedCLI{
		t:                     t,
		seen:                  make(map[string]int),
		plans:                 make(map[string]scriptedReversePlan),
		credentials:           make(map[string]scriptedCredential),
		connections:           make(map[string]scriptedConnection),
		tables:                make(map[string]scriptedTable),
		previewed:             make(map[string]bool),
		consumed:              make(map[string]bool),
		replays:               make(map[string]int),
		scheduleCron:          scriptedScheduleCron,
		scheduleFlow:          scriptedFlowName,
		installedScheduleCron: scriptedScheduleCron,
		installedScheduleFlow: scriptedFlowName,
		installedScheduleGap:  "  ",
		planFlow:              scriptedFlowName,
		previewFlow:           scriptedFlowName,
		flowRunSteps: []any{
			map[string]any{"id": "cert_sync", "status": "ok"},
			map[string]any{"id": "cert_query", "status": "ok"},
		},
		flowStatusSteps: []any{
			map[string]any{"id": "cert_sync", "status": "success"},
			map[string]any{"id": "cert_query", "status": "success"},
		},
		statusFlow: scriptedFlowName,
		protocols:  protocols,
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
		"credentials_add",
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
		if err := s.addCredential(args); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["credentials_add"]++
		return writeScriptedEnvelope(stdout, "Credential", nil)
	case exact(args, "credentials", "test", "cert-source", "--json"):
		if err := s.testCredential("cert-source"); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["credentials_test"]++
		return writeScriptedEnvelope(stdout, "CredentialTest", nil)
	case prefix(args, "connections", "create") && hasJSON(args):
		if err := s.createConnection(args); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["connections_create"]++
		return writeScriptedEnvelope(stdout, "Connection", nil)
	case prefix(args, "catalog", "refresh") && hasJSON(args):
		if err := s.refreshCatalog(args); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["catalog_refresh"]++
		return writeScriptedEnvelope(stdout, "Catalog", map[string]any{
			"catalog": map[string]any{"catalog": map[string]any{"streams": []map[string]any{
				{"name": "customers", "primary_key": []string{"id"}, "cursor_fields": []string{"updated_at"}},
				{"name": "events", "primary_key": []string{"id"}, "cursor_fields": []string{"updated_at"}},
			}}},
		})
	case prefix(args, "etl", "run") && hasJSON(args):
		if err := s.runETL(args); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["etl_run"]++
		return writeScriptedEnvelope(stdout, "ETLRun", map[string]any{"run": map[string]any{
			"records_read":      2,
			"records_succeeded": 1,
			"records_failed":    0,
			"checkpoint":        map[string]any{"cursor": "2026-08-06T00:00:00Z"},
		}})
	case prefix(args, "query", "run") && hasJSON(args):
		table, err := s.queryTable(args)
		s.seen["query_run"]++
		if err != nil {
			return writeScriptedEnvelopeWithCode(stdout, "Error", map[string]any{"error": "synthetic query table is absent"}, 2)
		}
		return writeScriptedEnvelope(stdout, "QueryResult", map[string]any{
			"count": len(table.rows),
			"rows":  table.rows,
		})
	case prefix(args, "flow", "plan"):
		if err := validateScriptedFlowManifest(args, root, "plan"); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["flow_plan"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"status": "ok", "flow": s.planFlow, "order": []string{"cert_sync", "cert_query"}})
	case prefix(args, "flow", "preview"):
		if err := validateScriptedFlowManifest(args, root, "preview"); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.seen["flow_preview"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"status": "dry_run", "flow": s.previewFlow})
	case prefix(args, "flow", "run") && hasJSON(args) && hasFlag(args, "--file") && hasFlag(args, "--flows-dir"):
		s.seen["flow_run"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"status": "ok", "steps": s.flowRunSteps})
	case prefix(args, "flow", "status"):
		if len(args) != 6 || args[2] != scriptedFlowName || args[3] != "--flows-dir" || args[4] == "" || args[5] != "--json" {
			return s.protocolError(stderr, "flow status requires the sample flow name, --flows-dir, and --json")
		}
		s.seen["flow_status"]++
		return writeScriptedEnvelopeWithoutKind(stdout, map[string]any{"flow": s.statusFlow, "steps": s.flowStatusSteps})
	case exact(args, scriptedScheduleCreateArgs()...):
		if s.schedule != "" {
			return s.protocolError(stderr, "schedule create repeated")
		}
		s.schedule = scriptedScheduleName
		s.seen["schedule_create"]++
		return writeScriptedEnvelope(stdout, "Schedule", nil)
	case exact(args, "schedule", "list", "--json"):
		if s.schedule != scriptedScheduleName {
			return s.protocolError(stderr, "schedule list ran before schedule create")
		}
		s.seen["schedule_list"]++
		return writeScriptedEnvelope(stdout, "ScheduleList", map[string]any{"schedules": []map[string]any{{
			"name": s.schedule,
			"cron": s.scheduleCron,
			"flow": s.scheduleFlow,
		}}})
	case exact(args, scriptedScheduleInstallArgs()...):
		if s.schedule != scriptedScheduleName || s.seen["schedule_list"] == 0 {
			return s.protocolError(stderr, "schedule install ran before schedule create and list")
		}
		if s.installed {
			return s.protocolError(stderr, "schedule install repeated")
		}
		if err := s.writeCrontabSentinel(); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.installed = true
		s.seen["schedule_install"]++
		return writeScriptedEnvelope(stdout, "ScheduleInstall", map[string]any{"backend": "crontab"})
	case exact(args, scriptedScheduleRemoveArgs()...):
		if s.schedule != scriptedScheduleName || !s.installed {
			return s.protocolError(stderr, "schedule remove ran before schedule install")
		}
		if err := s.clearCrontab(); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.installed = false
		s.seen["schedule_remove"]++
		return writeScriptedEnvelope(stdout, "ScheduleRemove", nil)
	case prefix(args, "reverse", "plan"):
		expectedAction, err := nextScriptedReversePlanAction(s.nextPlan)
		if err != nil {
			return s.protocolError(stderr, err.Error())
		}
		plan, err := parseScriptedReversePlan(args, expectedAction)
		if err != nil {
			return s.protocolError(stderr, err.Error())
		}
		if err := s.validateReversePlanDependencies(plan); err != nil {
			return s.protocolError(stderr, err.Error())
		}
		s.nextPlan++
		planID := fmt.Sprintf("scripted-plan-%d", s.nextPlan)
		plan.approvalToken = fmt.Sprintf("scripted-approval-%d", s.nextPlan)
		s.plans[planID] = plan
		s.seen["reverse_plan"]++
		if s.reversePlanOutput != "" {
			_, _ = io.WriteString(stdout, s.reversePlanOutput)
			return 0
		}
		_, _ = fmt.Fprintf(stdout, "Created reverse plan %s\nApproval token: %s\n", planID, plan.approvalToken)
		return 0
	case prefix(args, "reverse", "preview") && hasJSON(args) && len(args) == 4:
		if _, ok := s.plans[args[2]]; !ok {
			return s.protocolError(stderr, "reverse preview received an unknown plan")
		}
		s.previewed[args[2]] = true
		s.seen["reverse_preview"]++
		fields := map[string]any{"plan": map[string]any{"id": args[2], "records": 1}}
		if s.previewOutputLeak != "" {
			fields["preview_token"] = s.previewOutputLeak
		}
		return writeScriptedEnvelope(stdout, "ReversePlanPreview", fields)
	case prefix(args, "reverse", "run"):
		return s.runReverse(args, root, stdout, stderr)
	default:
		return s.protocolError(stderr, "unexpected command: "+strings.Join(args, " "))
	}
}

func (s *scriptedCLI) runReverse(args []string, root string, stdout, stderr io.Writer) int {
	if len(args) != 6 || args[3] != "--approve" || args[5] != "--json" {
		return s.protocolError(stderr, "reverse run requires plan id, its approval token, and --json")
	}
	planID := args[2]
	plan, ok := s.plans[planID]
	if !ok {
		return s.protocolError(stderr, "reverse run received an unknown plan")
	}
	if args[4] != plan.approvalToken {
		return s.protocolError(stderr, "reverse run approval token did not match its plan")
	}
	if plan.action == "create" && !s.previewed[planID] {
		return s.protocolError(stderr, "reverse run occurred before its preview")
	}
	if err := s.validateReversePlanDependencies(plan); err != nil {
		return s.protocolError(stderr, err.Error())
	}
	if s.consumed[planID] {
		s.replays[planID]++
		if s.replays[planID] > 1 {
			return s.protocolError(stderr, "reverse run retried a consumed approval more than once")
		}
		s.seen["reverse_replay"]++
		return writeScriptedEnvelopeWithCode(stdout, "Error", map[string]any{"error": "synthetic consumed approval"}, 2)
	}
	if err := s.appendOutboxAction(root, plan); err != nil {
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

func (s *scriptedCLI) addCredential(args []string) error {
	if len(args) < 3 || args[2] == "" {
		return fmt.Errorf("credentials add requires a credential name")
	}
	connector := flagValue(args, "--connector")
	if connector == "" {
		return fmt.Errorf("credentials add requires --connector")
	}
	if _, exists := s.credentials[args[2]]; exists {
		return fmt.Errorf("scripted credential %q already exists", args[2])
	}
	config, err := scriptedConfig(args)
	if err != nil {
		return err
	}
	s.credentials[args[2]] = scriptedCredential{connector: connector, config: config}
	return nil
}

func (s *scriptedCLI) testCredential(name string) error {
	credential, ok := s.credentials[name]
	if !ok {
		return fmt.Errorf("scripted credential %q does not exist", name)
	}
	if credential.connector != "sample" {
		return fmt.Errorf("scripted credential %q has connector %q, want sample", name, credential.connector)
	}
	return nil
}

func (s *scriptedCLI) createConnection(args []string) error {
	if len(args) < 3 || args[2] == "" {
		return fmt.Errorf("connections create requires a connection name")
	}
	if _, exists := s.connections[args[2]]; exists {
		return fmt.Errorf("scripted connection %q already exists", args[2])
	}
	sourceValue := flagValue(args, "--source")
	destinationValue := flagValue(args, "--destination")
	stream := flagValue(args, "--stream")
	syncMode := flagValue(args, "--sync-mode")
	table := flagValue(args, "--table")
	if sourceValue == "" || destinationValue == "" || stream == "" || syncMode == "" || table == "" {
		return fmt.Errorf("connections create requires source, destination, stream, table, and sync mode")
	}
	expectedMode, ok := expectedScriptedConnectionMode(args[2])
	if !ok {
		return fmt.Errorf("scripted connection %q is not a certification connection", args[2])
	}
	if syncMode != expectedMode {
		return fmt.Errorf("scripted connection %q has sync mode %q, want %q", args[2], syncMode, expectedMode)
	}
	source, err := parseScriptedEndpoint(sourceValue)
	if err != nil {
		return fmt.Errorf("connections create source: %w", err)
	}
	destination, err := parseScriptedEndpoint(destinationValue)
	if err != nil {
		return fmt.Errorf("connections create destination: %w", err)
	}
	if err := s.validateEndpoint(source); err != nil {
		return fmt.Errorf("connections create source: %w", err)
	}
	if err := s.validateEndpoint(destination); err != nil {
		return fmt.Errorf("connections create destination: %w", err)
	}
	s.connections[args[2]] = scriptedConnection{
		source:      source,
		destination: destination,
		stream:      stream,
		syncMode:    syncMode,
		table:       table,
	}
	return nil
}

func expectedScriptedConnectionMode(name string) (string, bool) {
	switch {
	case name == "cert_live" || strings.HasPrefix(name, "cert_live_"):
		return "full_refresh_append", true
	case name == "cert_incremental" || strings.HasPrefix(name, "cert_incremental_"):
		return "incremental_append", true
	case name == "cert_capture_full_refresh_overwrite_deduped" || strings.HasPrefix(name, "cert_capture_full_refresh_overwrite_deduped_"):
		return "full_refresh_overwrite_deduped", true
	case name == "cert_capture_full_refresh_overwrite" || strings.HasPrefix(name, "cert_capture_full_refresh_overwrite_"):
		return "full_refresh_overwrite", true
	case name == "cert_capture_incremental_append_deduped" || strings.HasPrefix(name, "cert_capture_incremental_append_deduped_"):
		return "incremental_append_deduped", true
	case strings.HasPrefix(name, "cert_flow_conn_"):
		return "full_refresh_append", true
	case name == "cert_write_verify" || strings.HasPrefix(name, "cert_write_seed_conn_") || name == "cert_sweep_seed_conn":
		return "full_refresh_overwrite", true
	default:
		return "", false
	}
}

func (s *scriptedCLI) refreshCatalog(args []string) error {
	connectionName := flagValue(args, "--connection")
	if connectionName == "" {
		return fmt.Errorf("catalog refresh requires --connection")
	}
	connection, ok := s.connections[connectionName]
	if !ok {
		return fmt.Errorf("scripted catalog connection %q does not exist", connectionName)
	}
	return s.validateConnection(connection)
}

func (s *scriptedCLI) runETL(args []string) error {
	connectionName := flagValue(args, "--connection")
	stream := flagValue(args, "--stream")
	if connectionName == "" || stream == "" {
		return fmt.Errorf("etl run requires --connection and --stream")
	}
	connection, ok := s.connections[connectionName]
	if !ok {
		return fmt.Errorf("scripted etl connection %q does not exist", connectionName)
	}
	if connection.stream != stream {
		return fmt.Errorf("scripted etl stream %q is not configured on connection %q", stream, connectionName)
	}
	expectedMode, ok := expectedScriptedConnectionMode(connectionName)
	if !ok || connection.syncMode != expectedMode {
		return fmt.Errorf("scripted etl connection %q has an invalid sync mode", connectionName)
	}
	if err := s.validateConnection(connection); err != nil {
		return fmt.Errorf("scripted etl connection %q: %w", connectionName, err)
	}
	if connection.destination.connector != "warehouse" {
		return fmt.Errorf("scripted etl connection %q must target warehouse", connectionName)
	}
	rows, err := s.sourceRows(connection.source)
	if err != nil {
		return fmt.Errorf("scripted etl source: %w", err)
	}
	s.tables[connection.table] = scriptedTable{rows: rows}
	return nil
}

func (s *scriptedCLI) queryTable(args []string) (scriptedTable, error) {
	tableName := flagValue(args, "--table")
	if tableName == "" {
		return scriptedTable{}, fmt.Errorf("query run requires --table")
	}
	table, ok := s.tables[tableName]
	if !ok {
		return scriptedTable{}, fmt.Errorf("scripted table %q is not materialized", tableName)
	}
	return table, nil
}

func (s *scriptedCLI) validateConnection(connection scriptedConnection) error {
	if err := s.validateEndpoint(connection.source); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if err := s.validateEndpoint(connection.destination); err != nil {
		return fmt.Errorf("destination: %w", err)
	}
	return nil
}

func (s *scriptedCLI) validateEndpoint(endpoint scriptedEndpoint) error {
	credential, ok := s.credentials[endpoint.credential]
	if !ok {
		return fmt.Errorf("credential %q does not exist", endpoint.credential)
	}
	if credential.connector != endpoint.connector {
		return fmt.Errorf("credential %q has connector %q, want %q", endpoint.credential, credential.connector, endpoint.connector)
	}
	return nil
}

func (s *scriptedCLI) sourceRows(endpoint scriptedEndpoint) ([]map[string]any, error) {
	if endpoint.connector != "file" {
		return []map[string]any{
			{"id": "synthetic-1", "updated_at": "2026-08-06T00:00:00Z"},
			{"id": "synthetic-2", "updated_at": "2026-08-06T00:00:01Z"},
		}, nil
	}
	credential := s.credentials[endpoint.credential]
	path := credential.config["path"]
	if path == "" {
		return nil, fmt.Errorf("file credential %q has no path", endpoint.credential)
	}
	return readScriptedJSONL(path)
}

func (s *scriptedCLI) validateReversePlanDependencies(plan scriptedReversePlan) error {
	if _, ok := s.tables[plan.sourceTable]; !ok {
		return fmt.Errorf("reverse plan source table %q is not materialized", plan.sourceTable)
	}
	if err := s.validateEndpoint(plan.destination); err != nil {
		return fmt.Errorf("reverse plan destination: %w", err)
	}
	return nil
}

func (s *scriptedCLI) appendOutboxAction(root string, plan scriptedReversePlan) error {
	table, ok := s.tables[plan.sourceTable]
	if !ok {
		return fmt.Errorf("scripted reverse source table %q is not materialized", plan.sourceTable)
	}
	tag := ""
	for _, row := range table.rows {
		if value, _ := row["tag"].(string); value != "" {
			tag = value
			break
		}
	}
	if tag == "" {
		return fmt.Errorf("scripted reverse source table %q did not contain tag", plan.sourceTable)
	}
	credential, ok := s.credentials[plan.destination.credential]
	if !ok {
		return fmt.Errorf("scripted outbox credential %q does not exist", plan.destination.credential)
	}
	outboxDir := credential.config["path"]
	if outboxDir != filepath.Join(root, ".polymetrics", "outbox") {
		return fmt.Errorf("scripted outbox credential path does not match the certification root")
	}
	path := filepath.Join(outboxDir, scriptedWritePlanName+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create scripted outbox: %w", err)
	}
	encoded, err := json.Marshal(map[string]any{"tag": tag, "_outbox_action": plan.action})
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

func scriptedConfig(args []string) (map[string]string, error) {
	config := make(map[string]string)
	for i := 0; i < len(args); i++ {
		if args[i] != "--config" {
			continue
		}
		if i+1 >= len(args) {
			return nil, fmt.Errorf("credentials add has config flag without value")
		}
		key, value, ok := strings.Cut(args[i+1], "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("credentials add has invalid config")
		}
		config[key] = value
		i++
	}
	return config, nil
}

func parseScriptedEndpoint(value string) (scriptedEndpoint, error) {
	connector, credential, ok := strings.Cut(value, ":")
	if !ok || connector == "" || credential == "" || strings.Contains(credential, ":") {
		return scriptedEndpoint{}, fmt.Errorf("invalid endpoint %q", value)
	}
	return scriptedEndpoint{connector: connector, credential: credential}, nil
}

func readScriptedJSONL(path string) ([]map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scripted file source: %w", err)
	}
	rows := make([]map[string]any, 0)
	for index, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("parse scripted file record %d: %w", index+1, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *scriptedCLI) writeCrontabSentinel() error {
	path := os.Getenv("PM_CRONTAB_FILE")
	if path == "" || s.schedule == "" {
		return fmt.Errorf("scripted schedule install missing crontab path or schedule name")
	}
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve scripted executable: %w", err)
	}
	if s.installedScheduleBin != "" {
		executable = s.installedScheduleBin
	}
	root := s.root
	if s.installedScheduleRoot != "" {
		root = s.installedScheduleRoot
	}
	line := fmt.Sprintf("%s  %s --root %s flow run %s --json%s# pm-schedule-%s\n", s.installedScheduleCron, scriptedScheduleArg(executable), scriptedScheduleArg(root), scriptedScheduleArg(s.installedScheduleFlow), s.installedScheduleGap, s.schedule)
	return os.WriteFile(path, []byte(line), 0o600)
}

func scriptedScheduleArg(s string) string {
	if s == "" {
		return "''"
	}
	if strings.IndexFunc(s, func(r rune) bool {
		return !(r >= 'A' && r <= 'Z') &&
			!(r >= 'a' && r <= 'z') &&
			!(r >= '0' && r <= '9') &&
			!strings.ContainsRune("/._:-", r)
	}) == -1 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
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

func validateScriptedFlowManifest(args []string, root, action string) error {
	want := []string{
		"flow", action,
		"--file", filepath.Join(root, ".polymetrics", "flows", scriptedFlowName+".json"),
		"--json",
	}
	if !exact(args, want...) {
		return fmt.Errorf("flow %s requires the sample manifest and --json", action)
	}
	return nil
}

func flagValue(args []string, want string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == want {
			return args[i+1]
		}
	}
	return ""
}

func scriptedReversePlanArgs(action string) []string {
	return []string{
		"reverse", "plan", scriptedWritePlanName,
		"--source-table", scriptedWriteSourceTable,
		"--destination", scriptedWriteDestination,
		"--map", "id:external_id",
		"--map", "tag:tag",
		"--action", action,
	}
}

func scriptedScheduleCreateArgs() []string {
	return []string{
		"schedule", "create",
		"--name", scriptedScheduleName,
		"--cron", scriptedScheduleCron,
		"--flow", scriptedFlowName,
		"--json",
	}
}

func scriptedScheduleInstallArgs() []string {
	return []string{"schedule", "install", scriptedScheduleName, "--crontab", "--json"}
}

func scriptedScheduleRemoveArgs() []string {
	return []string{"schedule", "remove", scriptedScheduleName, "--crontab", "--json"}
}

func nextScriptedReversePlanAction(planCount int) (string, error) {
	switch planCount {
	case 0:
		return "create", nil
	case 1:
		return "delete", nil
	default:
		return "", fmt.Errorf("sample write lifecycle permits exactly create then delete plans")
	}
}

func parseScriptedReversePlan(args []string, expectedAction string) (scriptedReversePlan, error) {
	want := scriptedReversePlanArgs(expectedAction)
	if len(args) != len(want) {
		return scriptedReversePlan{}, fmt.Errorf("reverse plan requires the exact sample write lifecycle arguments")
	}
	for i := range want {
		if args[i] != want[i] {
			return scriptedReversePlan{}, fmt.Errorf("reverse plan does not match the sample write lifecycle")
		}
	}
	return scriptedReversePlan{
		action:      expectedAction,
		sourceTable: scriptedWriteSourceTable,
		destination: scriptedEndpoint{connector: "outbox", credential: "cert-outbox"},
	}, nil
}

func mustRunScriptedCommand(t *testing.T, driver *scriptedCLI, root string, args ...string) {
	t.Helper()
	args = append(args, "--root", root)
	var stdout, stderr strings.Builder
	if code := driver.run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("scripted driver rejected %v: stdout=%s stderr=%s", args, stdout.String(), stderr.String())
	}
}

func prepareScriptedSeedTable(t *testing.T, driver *scriptedCLI, root string) {
	t.Helper()
	seedPath := filepath.Join(root, scriptedWriteSourceTable+"_seed.jsonl")
	seed, err := json.Marshal(map[string]any{"id": "scripted-seed", "tag": "scripted-seed"})
	if err != nil {
		t.Fatalf("marshal scripted seed: %v", err)
	}
	if err := os.WriteFile(seedPath, append(seed, '\n'), 0o600); err != nil {
		t.Fatalf("write scripted seed: %v", err)
	}
	mustRunScriptedCommand(t, driver, root,
		"credentials", "add", "cert-warehouse", "--connector", "warehouse",
		"--config", "path="+filepath.Join(root, ".polymetrics", "warehouse"), "--json")
	mustRunScriptedCommand(t, driver, root,
		"credentials", "add", "cert-write-seed-file-"+scriptedWriteSourceTable, "--connector", "file",
		"--config", "path="+seedPath, "--json")
	mustRunScriptedCommand(t, driver, root,
		"connections", "create", "cert_write_seed_conn_"+scriptedWriteSourceTable,
		"--source", "file:cert-write-seed-file-"+scriptedWriteSourceTable,
		"--destination", "warehouse:cert-warehouse",
		"--stream", scriptedWriteSourceTable+"_seed",
		"--primary-key", "id",
		"--sync-mode", "full_refresh_overwrite",
		"--table", scriptedWriteSourceTable,
		"--json")
	mustRunScriptedCommand(t, driver, root,
		"etl", "run", "--connection", "cert_write_seed_conn_"+scriptedWriteSourceTable,
		"--stream", scriptedWriteSourceTable+"_seed", "--json")
}

func prepareScriptedWriteDependencies(t *testing.T, driver *scriptedCLI, root string) {
	t.Helper()
	prepareScriptedSeedTable(t, driver, root)
	mustRunScriptedCommand(t, driver, root,
		"credentials", "add", "cert-outbox", "--connector", "outbox",
		"--config", "path="+filepath.Join(root, ".polymetrics", "outbox"), "--json")
}

func TestExpectedScriptedConnectionMode(t *testing.T) {
	for _, tt := range []struct {
		name string
		mode string
		ok   bool
	}{
		{name: "cert_live", mode: "full_refresh_append", ok: true},
		{name: "cert_live_events", mode: "full_refresh_append", ok: true},
		{name: "cert_capture_full_refresh_overwrite", mode: "full_refresh_overwrite", ok: true},
		{name: "cert_capture_full_refresh_overwrite_deduped_events", mode: "full_refresh_overwrite_deduped", ok: true},
		{name: "cert_capture_incremental_append_deduped", mode: "incremental_append_deduped", ok: true},
		{name: "cert_incremental", mode: "incremental_append", ok: true},
		{name: "cert_incremental_events", mode: "incremental_append", ok: true},
		{name: "cert_flow_conn_sample", mode: "full_refresh_append", ok: true},
		{name: "cert_flow_conn_sample_events", mode: "full_refresh_append", ok: true},
		{name: "cert_write_seed_conn_cert_write_seed_sample", mode: "full_refresh_overwrite", ok: true},
		{name: "cert_write_verify", mode: "full_refresh_overwrite", ok: true},
		{name: "cert_sweep_seed_conn", mode: "full_refresh_overwrite", ok: true},
		{name: "not-a-certification-connection", ok: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := expectedScriptedConnectionMode(tt.name)
			if got != tt.mode || ok != tt.ok {
				t.Fatalf("expectedScriptedConnectionMode(%q) = (%q, %t), want (%q, %t)", tt.name, got, ok, tt.mode, tt.ok)
			}
		})
	}
}

func TestScriptedCLIDriverRejectsProtocolDrift(t *testing.T) {
	t.Run("connection flags", func(t *testing.T) {
		driver := newScriptedCLI(t)
		var stdout, stderr strings.Builder
		if code := driver.run([]string{"connections", "create", "missing-required-flags", "--json", "--root", t.TempDir()}, &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted a connection command missing required flags")
		}
		if !strings.Contains(stderr.String(), "requires source") {
			t.Fatalf("protocol error = %q, want required-argument diagnostic", stderr.String())
		}
	})

	t.Run("resource dependencies", func(t *testing.T) {
		driver := newScriptedCLI(t)
		root := t.TempDir()
		mustRunScriptedCommand(t, driver, root, "credentials", "add", "cert-source", "--connector", "sample", "--json")
		mustRunScriptedCommand(t, driver, root,
			"credentials", "add", "cert-warehouse", "--connector", "warehouse",
			"--config", "path="+filepath.Join(root, ".polymetrics", "warehouse"), "--json")
		var stdout, stderr strings.Builder
		if code := driver.run([]string{
			"connections", "create", "cert_live",
			"--source", "sample:cert-source",
			"--destination", "warehouse:cert-warehouse",
			"--stream", "events",
			"--sync-mode", "full_refresh_overwrite",
			"--table", "cert_live_events",
			"--json", "--root", root,
		}, &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted a live connection with the wrong sync mode")
		}
		stdout.Reset()
		stderr.Reset()
		mustRunScriptedCommand(t, driver, root,
			"connections", "create", "cert_live",
			"--source", "sample:cert-source",
			"--destination", "warehouse:cert-warehouse",
			"--stream", "events",
			"--sync-mode", "full_refresh_append",
			"--table", "cert_live_events",
			"--json")

		if code := driver.run([]string{"etl", "run", "--connection", "cert_live", "--stream", "customers", "--json", "--root", root}, &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted an ETL stream not configured on the connection")
		}
		stdout.Reset()
		stderr.Reset()
		mustRunScriptedCommand(t, driver, root, "etl", "run", "--connection", "cert_live", "--stream", "events", "--json")
		if code := driver.run([]string{"query", "run", "--table", "cert_live_customers", "--json", "--root", root}, &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted a query for an unmaterialized table")
		}
	})

	t.Run("reverse plan dependencies", func(t *testing.T) {
		driver := newScriptedCLI(t)
		root := t.TempDir()
		var stdout, stderr strings.Builder
		if code := driver.run(append(scriptedReversePlanArgs("create"), "--root", root), &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted a reverse plan without a materialized source table")
		}
		stdout.Reset()
		stderr.Reset()
		prepareScriptedSeedTable(t, driver, root)
		if code := driver.run(append(scriptedReversePlanArgs("create"), "--root", root), &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted a reverse plan without its destination credential")
		}
		stdout.Reset()
		stderr.Reset()
		mustRunScriptedCommand(t, driver, root,
			"credentials", "add", "cert-outbox", "--connector", "outbox",
			"--config", "path="+filepath.Join(root, ".polymetrics", "outbox"), "--json")
		mustRunScriptedCommand(t, driver, root, scriptedReversePlanArgs("create")...)
	})

	for _, tt := range []struct {
		name   string
		mutate func([]string)
	}{
		{
			name: "source table",
			mutate: func(args []string) {
				args[4] = "wrong_source"
			},
		},
		{
			name: "destination",
			mutate: func(args []string) {
				args[6] = "outbox:wrong-destination"
			},
		},
		{
			name: "mapping",
			mutate: func(args []string) {
				args[10] = "tag:wrong_tag"
			},
		},
	} {
		t.Run("reverse plan "+tt.name, func(t *testing.T) {
			driver := newScriptedCLI(t)
			args := scriptedReversePlanArgs("create")
			tt.mutate(args)
			args = append(args, "--root", t.TempDir())
			var stdout, stderr strings.Builder
			if code := driver.run(args, &stdout, &stderr); code == 0 {
				t.Fatalf("scripted driver accepted reverse plan with wrong %s", tt.name)
			}
		})
	}

	t.Run("flow status target", func(t *testing.T) {
		driver := newScriptedCLI(t)
		root := t.TempDir()
		var stdout, stderr strings.Builder
		if code := driver.run([]string{"flow", "status", "wrong-flow", "--flows-dir", root, "--json", "--root", root}, &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted flow status for the wrong flow")
		}
	})

	t.Run("flow plan and preview manifests", func(t *testing.T) {
		for _, action := range []string{"plan", "preview"} {
			t.Run(action, func(t *testing.T) {
				driver := newScriptedCLI(t)
				root := t.TempDir()
				var stdout, stderr strings.Builder
				args := []string{"flow", action, "--file", filepath.Join(root, "wrong-flow.json"), "--json", "--root", root}
				if code := driver.run(args, &stdout, &stderr); code == 0 {
					t.Fatalf("scripted driver accepted %s with the wrong manifest", action)
				}
			})
		}
	})

	t.Run("schedule contract", func(t *testing.T) {
		for _, tt := range []struct {
			name   string
			mutate func([]string)
		}{
			{
				name: "name",
				mutate: func(args []string) {
					args[3] = "wrong-schedule"
				},
			},
			{
				name: "cron",
				mutate: func(args []string) {
					args[5] = "0 4 * * *"
				},
			},
			{
				name: "flow",
				mutate: func(args []string) {
					args[7] = "wrong-flow"
				},
			},
		} {
			t.Run("create "+tt.name, func(t *testing.T) {
				driver := newScriptedCLI(t)
				args := scriptedScheduleCreateArgs()
				tt.mutate(args)
				args = append(args, "--root", t.TempDir())
				var stdout, stderr strings.Builder
				if code := driver.run(args, &stdout, &stderr); code == 0 {
					t.Fatalf("scripted driver accepted schedule create with wrong %s", tt.name)
				}
			})
		}

		driver := newScriptedCLI(t)
		root := t.TempDir()
		t.Setenv("PM_CRONTAB_FILE", filepath.Join(root, "crontab"))
		var stdout, stderr strings.Builder
		for _, args := range [][]string{scriptedScheduleCreateArgs(), {"schedule", "list", "--json"}} {
			args = append(args, "--root", root)
			if code := driver.run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("scripted driver rejected valid schedule setup: stderr=%s", stderr.String())
			}
			stdout.Reset()
			stderr.Reset()
		}

		install := scriptedScheduleInstallArgs()
		install[2] = "wrong-schedule"
		if code := driver.run(append(install, "--root", root), &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted schedule install for the wrong schedule")
		}
		stdout.Reset()
		stderr.Reset()
		if code := driver.run(append(scriptedScheduleInstallArgs(), "--root", root), &stdout, &stderr); code != 0 {
			t.Fatalf("scripted driver rejected valid schedule install: stderr=%s", stderr.String())
		}
		stdout.Reset()
		stderr.Reset()

		remove := scriptedScheduleRemoveArgs()
		remove[2] = "wrong-schedule"
		if code := driver.run(append(remove, "--root", root), &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted schedule remove for the wrong schedule")
		}
	})

	t.Run("reverse plan lifecycle actions", func(t *testing.T) {
		t.Run("first plan must create", func(t *testing.T) {
			driver := newScriptedCLI(t)
			var stdout, stderr strings.Builder
			if code := driver.run(append(scriptedReversePlanArgs("delete"), "--root", t.TempDir()), &stdout, &stderr); code == 0 {
				t.Fatal("scripted driver accepted delete as the first lifecycle plan")
			}
		})

		t.Run("cleanup plan must delete", func(t *testing.T) {
			driver := newScriptedCLI(t)
			root := t.TempDir()
			prepareScriptedWriteDependencies(t, driver, root)
			var stdout, stderr strings.Builder
			if code := driver.run(append(scriptedReversePlanArgs("create"), "--root", root), &stdout, &stderr); code != 0 {
				t.Fatalf("scripted driver rejected valid create plan: stderr=%s", stderr.String())
			}
			stdout.Reset()
			stderr.Reset()
			if code := driver.run(append(scriptedReversePlanArgs("create"), "--root", root), &stdout, &stderr); code == 0 {
				t.Fatal("scripted driver accepted create as the cleanup lifecycle plan")
			}
		})
	})

	t.Run("per-plan approval token", func(t *testing.T) {
		driver := newScriptedCLI(t)
		root := t.TempDir()
		prepareScriptedWriteDependencies(t, driver, root)
		planArgs := append(scriptedReversePlanArgs("create"), "--root", root)
		var stdout, stderr strings.Builder
		if code := driver.run(planArgs, &stdout, &stderr); code != 0 {
			t.Fatalf("scripted driver rejected valid reverse plan: stderr=%s", stderr.String())
		}
		stdout.Reset()
		stderr.Reset()
		if code := driver.run([]string{"reverse", "run", "scripted-plan-1", "--approve", "wrong-approval", "--json", "--root", root}, &stdout, &stderr); code == 0 {
			t.Fatal("scripted driver accepted a mismatched approval token")
		}
		if !strings.Contains(stderr.String(), "did not match") {
			t.Fatalf("protocol error = %q, want approval mismatch diagnostic", stderr.String())
		}
	})
}
