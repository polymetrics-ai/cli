package main

import (
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestGitHubAPISurfaceOperationLedgerMetrics(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string           `json:"method"`
			CoveredBy map[string]any   `json:"covered_by"`
			Excluded  map[string]any   `json:"excluded"`
			Operation *githubOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal github api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	totalByMethod := map[string]int{}
	coveredByMethod := map[string]int{}
	operationByMethod := map[string]int{}
	models := map[string]int{}
	risks := map[string]int{}
	statuses := map[string]int{}
	covered, excluded, operations := 0, 0, 0

	for i, ep := range surface.Endpoints {
		totalByMethod[ep.Method]++
		if len(ep.CoveredBy) > 0 {
			covered++
			coveredByMethod[ep.Method]++
		}
		if len(ep.Excluded) > 0 {
			excluded++
		}
		if ep.Operation != nil {
			operations++
			operationByMethod[ep.Method]++
			models[ep.Operation.Model]++
			risks[ep.Operation.Risk]++
			statuses[ep.Operation.Status]++
			if !ep.Operation.BlockedByDefault {
				t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
			}
			if ep.Operation.Reason == "" {
				t.Fatalf("endpoint %d operation is missing reason: %+v", i, ep.Operation)
			}
			if requiresSourceOrNotes(ep.Operation.Model) && ep.Operation.SourceURL == "" && ep.Operation.Notes == "" {
				t.Fatalf("endpoint %d operation %q is missing source_url or notes", i, ep.Operation.Model)
			}
			if ep.Operation.Model == "duplicate" && ep.Operation.DuplicateOf == "" {
				t.Fatalf("endpoint %d duplicate operation is missing duplicate_of", i)
			}
		}
	}

	if len(surface.Endpoints) != 1604 {
		t.Fatalf("endpoints = %d, want 1604 (1596 official rows plus 8 connector conformance coverage rows)", len(surface.Endpoints))
	}
	if covered != 223 {
		t.Fatalf("covered endpoints = %d, want 223", covered)
	}
	if operations != 1381 {
		t.Fatalf("operation endpoints = %d, want 1381 blocked/planned rows", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	assertStringIntMap(t, "totalByMethod", totalByMethod, map[string]int{
		"DELETE":  187,
		"GET":     634,
		"GRAPHQL": 309,
		"PATCH":   73,
		"POST":    193,
		"PUT":     133,
		"WEBHOOK": 75,
	})
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, map[string]int{
		"DELETE":  17,
		"GET":     155,
		"GRAPHQL": 4,
		"PATCH":   15,
		"POST":    23,
		"PUT":     9,
	})
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{
		"DELETE":  170,
		"GET":     479,
		"GRAPHQL": 305,
		"PATCH":   58,
		"POST":    170,
		"PUT":     124,
		"WEBHOOK": 75,
	})
	assertStringIntMap(t, "models", models, map[string]int{
		"admin_reverse_etl":     494,
		"destructive_action":    224,
		"direct_read":           516,
		"disallowed":            1,
		"duplicate":             53,
		"deprecated":            30,
		"sensitive_reverse_etl": 63,
	})
	assertStringIntMap(t, "risks", risks, map[string]int{
		"critical": 60,
		"high":     729,
		"low":      419,
		"medium":   173,
	})
	assertStringIntMap(t, "statuses", statuses, map[string]int{
		"blocked": 1381,
	})
}

func TestGitHubDestructiveMetadataUsesTypedConfirmation(t *testing.T) {
	fixtureEntries, err := os.ReadDir("../../internal/connectors/defs/github/fixtures/writes")
	if err != nil {
		t.Fatalf("read github write fixtures: %v", err)
	}
	writeFixtures := map[string]bool{}
	for _, entry := range fixtureEntries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		writeFixtures[strings.TrimSuffix(entry.Name(), ".json")] = true
	}

	writesRaw, err := os.ReadFile("../../internal/connectors/defs/github/writes.json")
	if err != nil {
		t.Fatalf("read github writes.json: %v", err)
	}
	var writes struct {
		Actions []githubWriteAction `json:"actions"`
	}
	if err := json.Unmarshal(writesRaw, &writes); err != nil {
		t.Fatalf("unmarshal github writes.json: %v", err)
	}
	destructiveWrites := map[string]bool{}
	for _, action := range writes.Actions {
		if githubWritePathHasLiteralPlaceholder(action.Path) {
			t.Fatalf("write action %q path uses single-brace placeholders: %q", action.Name, action.Path)
		}
		if githubActionRequiresTypedDestructiveConfirmation(action) && action.Confirm != "destructive" {
			t.Fatalf("write action %q confirm = %q, want destructive", action.Name, action.Confirm)
		}
		if !writeFixtures[action.Name] {
			t.Fatalf("write action %q lacks connector-owned write fixture", action.Name)
		}
		if action.RecordSchema.Type != "object" {
			t.Fatalf("write action %q record_schema type = %q, want object", action.Name, action.RecordSchema.Type)
		}
		if len(action.RecordSchema.Properties) == 0 && action.BodyType != "none" {
			t.Fatalf("write action %q lacks typed record_schema properties", action.Name)
		}
		if action.RecordSchema.AdditionalProperties == nil || *action.RecordSchema.AdditionalProperties {
			t.Fatalf("write action %q must disable additional record properties", action.Name)
		}
		if action.Name == "create_repo_ruleset" || action.Name == "update_repo_ruleset" {
			githubAssertRulesetSchemaBlocksUnmodeledFields(t, action)
		}
		sensitiveFields := githubSensitiveSchemaPaths(action.RecordSchema.Properties)
		if len(sensitiveFields) > 0 {
			if action.Confirm != "destructive" {
				t.Fatalf("write action %q has sensitive fields %+v but confirm = %q, want destructive", action.Name, sensitiveFields, action.Confirm)
			}
			for _, field := range sensitiveFields {
				if !githubStringSliceContains(action.RedactFields, field) {
					t.Fatalf("write action %q has sensitive field %q without redact_fields entry", action.Name, field)
				}
			}
		}
		if action.Confirm == "destructive" {
			destructiveWrites[action.Name] = true
		}
		if action.Delete != nil {
			t.Fatalf("write action %q declares missing_ok_status without GitHub already-absent proof: %+v", action.Name, action.Delete)
		}
	}

	cliRaw, err := os.ReadFile("../../internal/connectors/defs/github/cli_surface.json")
	if err != nil {
		t.Fatalf("read github cli_surface.json: %v", err)
	}
	var cli struct {
		Commands []githubCLICommand `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("unmarshal github cli_surface.json: %v", err)
	}
	unsafe := map[string]bool{}
	commandsByPath := map[string]githubCLICommand{}
	for _, cmd := range cli.Commands {
		if _, ok := commandsByPath[cmd.Path]; ok {
			t.Fatalf("duplicate CLI command path %q", cmd.Path)
		}
		commandsByPath[cmd.Path] = cmd
		if cmd.Availability == "unsafe_or_disallowed" {
			unsafe[cmd.Path] = true
		}
		if destructiveWrites[cmd.Write] && !strings.Contains(cmd.Approval, "destructive") {
			t.Fatalf("command %q maps destructive write %q but approval omits typed destructive confirmation: %q", cmd.Path, cmd.Write, cmd.Approval)
		}
	}
	assertStringBoolMap(t, "unsafe_or_disallowed commands", unsafe, map[string]bool{
		"api":        true,
		"auth token": true,
	})
	if _, ok := commandsByPath["repo delete-2"]; ok {
		t.Fatalf("synthetic repo delete-2 command must not be exposed")
	}
	repoDelete, ok := commandsByPath["repo delete"]
	if !ok {
		t.Fatalf("canonical repo delete command is missing")
	}
	if repoDelete.Intent != "reverse_etl" || repoDelete.Availability != "implemented" || repoDelete.Write != "delete_repo" {
		t.Fatalf("repo delete command = %+v, want implemented reverse_etl write delete_repo", repoDelete)
	}
	if !strings.Contains(repoDelete.Approval, "destructive") {
		t.Fatalf("repo delete approval omits destructive confirmation: %q", repoDelete.Approval)
	}
}

func TestGitHubMultiSegmentWritePathsRemainBlocked(t *testing.T) {
	targetWrites := map[string]bool{
		"create_or_update_file": true,
		"delete_file":           true,
		"update_ref":            true,
		"delete_ref":            true,
	}

	writesRaw, err := os.ReadFile("../../internal/connectors/defs/github/writes.json")
	if err != nil {
		t.Fatalf("read github writes.json: %v", err)
	}
	var writes struct {
		Actions []githubWriteAction `json:"actions"`
	}
	if err := json.Unmarshal(writesRaw, &writes); err != nil {
		t.Fatalf("unmarshal github writes.json: %v", err)
	}
	for _, action := range writes.Actions {
		if targetWrites[action.Name] {
			t.Fatalf("write action %q must remain blocked until typed multi-segment write path support exists", action.Name)
		}
	}

	cliRaw, err := os.ReadFile("../../internal/connectors/defs/github/cli_surface.json")
	if err != nil {
		t.Fatalf("read github cli_surface.json: %v", err)
	}
	var cli struct {
		Commands []githubCLICommand `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("unmarshal github cli_surface.json: %v", err)
	}
	for _, cmd := range cli.Commands {
		if targetWrites[cmd.Write] {
			t.Fatalf("command %q still maps blocked multi-segment write %q", cmd.Path, cmd.Write)
		}
	}

	surfaceRaw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
	}
	var surface struct {
		Endpoints []struct {
			Method    string           `json:"method"`
			Path      string           `json:"path"`
			CoveredBy map[string]any   `json:"covered_by"`
			Operation *githubOperation `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(surfaceRaw, &surface); err != nil {
		t.Fatalf("unmarshal github api_surface.json: %v", err)
	}
	targetEndpoints := map[string]bool{
		"PUT /repos/{owner}/{repo}/contents/{path}":    true,
		"DELETE /repos/{owner}/{repo}/contents/{path}": true,
		"PATCH /repos/{owner}/{repo}/git/refs/{ref}":   true,
		"DELETE /repos/{owner}/{repo}/git/refs/{ref}":  true,
	}
	seen := 0
	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if !targetEndpoints[key] {
			continue
		}
		seen++
		if len(ep.CoveredBy) > 0 {
			t.Fatalf("endpoint %s is covered by %+v; want blocked operation", key, ep.CoveredBy)
		}
		if ep.Operation == nil {
			t.Fatalf("endpoint %s has no blocked operation ledger row", key)
		}
		if ep.Operation.Status != "blocked" || ep.Operation.Model != "destructive_action" || !ep.Operation.BlockedByDefault {
			t.Fatalf("endpoint %s operation = %+v, want blocked destructive_action", key, ep.Operation)
		}
		if !strings.Contains(ep.Operation.Reason, "typed allowlisted multi-segment write-path support") {
			t.Fatalf("endpoint %s reason = %q, want typed multi-segment write path dependency", key, ep.Operation.Reason)
		}
	}
	if seen != len(targetEndpoints) {
		t.Fatalf("blocked multi-segment endpoints seen = %d, want %d", seen, len(targetEndpoints))
	}

	cases := []struct {
		name     string
		template string
		record   map[string]any
		want     string
	}{
		{
			name:     "contents path",
			template: "/repos/{{ config.owner }}/{{ config.repo }}/contents/{{ record.path }}",
			record:   map[string]any{"path": "dir/file.txt"},
			want:     "/repos/octocat/hello-world/contents/dir%2Ffile.txt",
		},
		{
			name:     "git ref",
			template: "/repos/{{ config.owner }}/{{ config.repo }}/git/refs/{{ record.ref }}",
			record:   map[string]any{"ref": "heads/main"},
			want:     "/repos/octocat/hello-world/git/refs/heads%2Fmain",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.InterpolatePath(tt.template, engine.Vars{
				Config: map[string]string{"owner": "octocat", "repo": "hello-world"},
				Record: tt.record,
			})
			if err != nil {
				t.Fatalf("InterpolatePath: %v", err)
			}
			if got != tt.want {
				t.Fatalf("InterpolatePath = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGitHubImplementedDirectReadOperationsAreRunnable(t *testing.T) {
	cliRaw, err := os.ReadFile("../../internal/connectors/defs/github/cli_surface.json")
	if err != nil {
		t.Fatalf("read github cli_surface.json: %v", err)
	}
	var cli struct {
		Commands []githubCLICommand `json:"commands"`
	}
	if err := json.Unmarshal(cliRaw, &cli); err != nil {
		t.Fatalf("unmarshal github cli_surface.json: %v", err)
	}

	operationsRaw, err := os.ReadFile("../../internal/connectors/defs/github/operations.json")
	if err != nil {
		t.Fatalf("read github operations.json: %v", err)
	}
	var operations struct {
		Operations []githubOperationSpec `json:"operations"`
	}
	if err := json.Unmarshal(operationsRaw, &operations); err != nil {
		t.Fatalf("unmarshal github operations.json: %v", err)
	}
	operationsByID := map[string]githubOperationSpec{}
	for _, op := range operations.Operations {
		operationsByID[op.ID] = op
	}

	implementedDirectReads := map[string]githubCLICommand{}
	for _, cmd := range cli.Commands {
		if cmd.Intent != "direct_read" || cmd.Availability != "implemented" {
			continue
		}
		implementedDirectReads[cmd.Path] = cmd
		if cmd.Operation == "" {
			continue
		}
		op, ok := operationsByID[cmd.Operation]
		if !ok {
			t.Fatalf("command %q references undeclared operation %q", cmd.Path, cmd.Operation)
		}
		if op.Kind != "rest_read" || op.REST == nil {
			t.Fatalf("command %q operation %q kind = %q, want rest_read with rest metadata", cmd.Path, op.ID, op.Kind)
		}
		if op.REST.MaxBytes <= 0 {
			t.Fatalf("command %q operation %q rest.max_bytes = %d, want positive bound", cmd.Path, op.ID, op.REST.MaxBytes)
		}
		githubAssertOperationDirectReadMetadata(t, cmd, op)
		githubAssertOperationDirectReadPathParamsSupported(t, cmd, op)
		githubAssertDirectReadSensitiveFieldsRedacted(t, cmd, op)
		if !githubSupportedDirectReadOutputPolicy(cmd.OutputPolicy) {
			t.Fatalf("command %q output_policy = %q, want supported direct-read policy", cmd.Path, cmd.OutputPolicy)
		}
		if !githubSupportedDirectReadOutputPolicy(op.OutputPolicy) {
			t.Fatalf("operation %q output_policy = %q, want supported direct-read policy", op.ID, op.OutputPolicy)
		}
	}

	surfaceRaw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
	}
	var surface struct {
		Endpoints []struct {
			CoveredBy *githubSurfaceCoverage `json:"covered_by"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(surfaceRaw, &surface); err != nil {
		t.Fatalf("unmarshal github api_surface.json: %v", err)
	}
	for i, ep := range surface.Endpoints {
		if ep.CoveredBy == nil {
			continue
		}
		directReads := append([]string{}, ep.CoveredBy.DirectReads...)
		if ep.CoveredBy.DirectRead != "" {
			directReads = append(directReads, ep.CoveredBy.DirectRead)
		}
		for _, path := range directReads {
			if _, ok := implementedDirectReads[path]; !ok {
				t.Fatalf("endpoint %d covers direct read %q without an implemented runnable command", i, path)
			}
		}
	}
}

type githubOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url"`
	Notes            string `json:"notes"`
	DuplicateOf      string `json:"duplicate_of"`
}

type githubWriteAction struct {
	Name         string              `json:"name"`
	Kind         string              `json:"kind"`
	Method       string              `json:"method"`
	Path         string              `json:"path"`
	Risk         string              `json:"risk"`
	Confirm      string              `json:"confirm"`
	RedactFields []string            `json:"redact_fields"`
	BodyType     string              `json:"body_type"`
	Delete       *githubDeletePolicy `json:"delete"`
	RecordSchema githubRecordSchema  `json:"record_schema"`
}

type githubRecordSchema struct {
	Type                 string                     `json:"type"`
	Properties           map[string]json.RawMessage `json:"properties"`
	AdditionalProperties *bool                      `json:"additionalProperties"`
}

type githubDeletePolicy struct {
	Idempotent      bool  `json:"idempotent"`
	MissingOKStatus []int `json:"missing_ok_status"`
}

type githubOperationSpec struct {
	ID              string                   `json:"id"`
	Kind            string                   `json:"kind"`
	OutputPolicy    string                   `json:"output_policy"`
	SensitivePolicy *githubSensitivePolicy   `json:"sensitive_policy"`
	REST            *githubOperationRESTSpec `json:"rest"`
}

type githubOperationRESTSpec struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Query    map[string]string `json:"query"`
	MaxBytes int               `json:"max_bytes"`
}

type githubSensitivePolicy struct {
	RedactFields []string `json:"redact_fields"`
}

type githubSurfaceCoverage struct {
	DirectRead  string   `json:"direct_read"`
	DirectReads []string `json:"direct_reads"`
}

type githubCLICommand struct {
	Path         string                     `json:"path"`
	Intent       string                     `json:"intent"`
	Availability string                     `json:"availability"`
	Operation    string                     `json:"operation"`
	APISurface   []githubCLISurfaceEndpoint `json:"api_surface"`
	OutputPolicy string                     `json:"output_policy"`
	RedactFields []string                   `json:"redact_fields"`
	Flags        []githubCLIFlag            `json:"flags"`
	Write        string                     `json:"write"`
	Approval     string                     `json:"approval"`
}

type githubCLISurfaceEndpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type githubCLIFlag struct {
	Name   string `json:"name"`
	MapsTo string `json:"maps_to"`
}

var githubSingleBracePathParamRE = regexp.MustCompile(`(^|[^{])\{[A-Za-z0-9_]+\}([^}]|$)`)

var githubPathParamRE = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)

var githubDestructiveRiskPhrases = []string{
	"irreversibly",
	"repository history",
	"writes a commit",
	"ci/cd",
	"deploy access",
	"protection rules",
	"replaces every",
	"entire topic list",
	"clearing its approval",
	"merge commit",
	"discarding history",
	"suppress a real",
	"deployment automation",
	"ruleset",
	"grants a github user access",
}

func githubWritePathHasLiteralPlaceholder(path string) bool {
	return githubSingleBracePathParamRE.MatchString(path)
}

func githubSensitiveSchemaPaths(properties map[string]json.RawMessage) []string {
	var paths []string
	for name, raw := range properties {
		githubCollectSensitiveSchemaPaths(name, raw, &paths)
	}
	return paths
}

func githubCollectSensitiveSchemaPaths(path string, raw json.RawMessage, paths *[]string) {
	parts := strings.Split(path, ".")
	name := parts[len(parts)-1]
	if githubSensitiveFieldName(name) {
		*paths = append(*paths, path)
	}
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return
	}
	for child, childRaw := range schema.Properties {
		githubCollectSensitiveSchemaPaths(path+"."+child, childRaw, paths)
	}
}

func githubSensitiveFieldName(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	return normalized == "secret" || strings.Contains(normalized, "token") || strings.Contains(normalized, "private_key")
}

func githubStringSliceContains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func githubSupportedDirectReadOutputPolicy(policy string) bool {
	switch policy {
	case "repository_contents_file_metadata", "repository_contents_directory", "json_redacted", "clinical_json_redacted":
		return true
	default:
		return false
	}
}

func githubAssertOperationDirectReadPathParamsSupported(t *testing.T, cmd githubCLICommand, op githubOperationSpec) {
	t.Helper()
	for _, match := range githubPathParamRE.FindAllStringSubmatch(op.REST.Path, -1) {
		if githubUnsupportedDirectReadPathParam(match[1]) {
			t.Fatalf("command %q operation %q path parameter %q requires non-identifier values; keep planned until typed path encoding is supported", cmd.Path, op.ID, match[1])
		}
	}
}

func githubUnsupportedDirectReadPathParam(name string) bool {
	switch name {
	case "basehead", "branch", "concurrency_group_name", "dir", "environment_name", "ref", "subject_digest":
		return true
	default:
		return false
	}
}

func githubAssertDirectReadSensitiveFieldsRedacted(t *testing.T, cmd githubCLICommand, op githubOperationSpec) {
	t.Helper()
	if !githubDirectReadNeedsValueRedaction(op) {
		return
	}
	if githubStringSliceContains(cmd.RedactFields, "value") {
		return
	}
	if op.SensitivePolicy != nil && githubStringSliceContains(op.SensitivePolicy.RedactFields, "value") {
		return
	}
	t.Fatalf("command %q operation %q must redact GitHub variable value fields", cmd.Path, op.ID)
}

func githubDirectReadNeedsValueRedaction(op githubOperationSpec) bool {
	if op.REST == nil {
		return false
	}
	return strings.Contains(op.REST.Path, "/actions/organization-variables") || strings.Contains(op.REST.Path, "/actions/variables")
}

func githubAssertRulesetSchemaBlocksUnmodeledFields(t *testing.T, action githubWriteAction) {
	t.Helper()
	for _, field := range []string{"bypass_actors", "conditions", "rules"} {
		if _, ok := action.RecordSchema.Properties[field]; ok {
			t.Fatalf("write action %q exposes unmodeled ruleset field %q", action.Name, field)
		}
	}
}

func githubAssertOperationDirectReadMetadata(t *testing.T, cmd githubCLICommand, op githubOperationSpec) {
	t.Helper()
	if len(cmd.APISurface) != 1 {
		t.Fatalf("command %q operation direct_read api_surface endpoints = %d, want 1", cmd.Path, len(cmd.APISurface))
	}
	endpoint := cmd.APISurface[0]
	if !strings.EqualFold(endpoint.Method, op.REST.Method) || endpoint.Path != op.REST.Path {
		t.Fatalf("command %q api_surface = %s %s, want %s %s", cmd.Path, endpoint.Method, endpoint.Path, op.REST.Method, op.REST.Path)
	}

	pathParams := map[string]bool{}
	for _, match := range githubPathParamRE.FindAllStringSubmatch(op.REST.Path, -1) {
		pathParams[match[1]] = true
	}
	mappedPathParams := map[string]bool{}
	for _, flag := range cmd.Flags {
		if flag.MapsTo == "" {
			t.Fatalf("command %q flag --%s is missing maps_to", cmd.Path, flag.Name)
		}
		switch {
		case strings.HasPrefix(flag.MapsTo, "path."):
			target := strings.TrimPrefix(flag.MapsTo, "path.")
			if !pathParams[target] {
				t.Fatalf("command %q flag --%s maps to non-path target %q for operation %q", cmd.Path, flag.Name, flag.MapsTo, op.ID)
			}
			mappedPathParams[target] = true
		case strings.HasPrefix(flag.MapsTo, "query."):
			if strings.TrimPrefix(flag.MapsTo, "query.") == "" {
				t.Fatalf("command %q flag --%s maps to empty query target", cmd.Path, flag.Name)
			}
		case strings.HasPrefix(flag.MapsTo, "body."):
			if !strings.EqualFold(op.REST.Method, "POST") || strings.TrimPrefix(flag.MapsTo, "body.") == "" {
				t.Fatalf("command %q flag --%s maps to unsupported body target %q", cmd.Path, flag.Name, flag.MapsTo)
			}
		default:
			t.Fatalf("command %q flag --%s maps to unsupported target %q", cmd.Path, flag.Name, flag.MapsTo)
		}
	}
	for pathParam := range pathParams {
		if pathParam == "owner" || pathParam == "repo" {
			continue
		}
		if !mappedPathParams[pathParam] {
			t.Fatalf("command %q operation %q path parameter %q has no mapped flag", cmd.Path, op.ID, pathParam)
		}
	}
}

func githubActionRequiresTypedDestructiveConfirmation(action githubWriteAction) bool {
	if action.Method == "DELETE" || action.Kind == "delete" {
		return true
	}
	risk := strings.ToLower(action.Risk)
	if risk == "critical" || risk == "high" {
		return true
	}
	for _, phrase := range githubDestructiveRiskPhrases {
		if strings.Contains(risk, phrase) {
			return true
		}
	}
	return false
}

func requiresSourceOrNotes(model string) bool {
	switch model {
	case "sensitive_reverse_etl", "admin_reverse_etl", "destructive_action", "disallowed":
		return true
	default:
		return false
	}
}

func assertStringIntMap(t *testing.T, name string, got, want map[string]int) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}

func assertStringBoolMap(t *testing.T, name string, got, want map[string]bool) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}
