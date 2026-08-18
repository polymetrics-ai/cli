package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	githubRetiredUserDraftRESTPath = "/user/{user_id}/projectsV2/{project_number}/drafts"
	githubUserDraftGraphQLTarget   = "POST /graphql (github.graphql.mutation.add-project-v2-draft-issue)"
	githubUserDraftCLICommand      = "projects create-draft-item-for-authenticated-user"
)

func TestGitHubAPISurfaceOperationLedgerMetrics(t *testing.T) {
	lock := loadGitHubSourceLock(t)
	raw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string           `json:"method"`
			Path      string           `json:"path"`
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
	covered, excluded, operations, restEndpoints, graphQLBindings, generatedGraphQLTransports := 0, 0, 0, 0, 0, 0
	sawRetiredUserDraftRESTRoute := false

	for i, ep := range surface.Endpoints {
		generatedTransport := githubGeneratedGraphQLTransport(ep.Method, ep.Path, ep.CoveredBy)
		if ep.Method == "GRAPHQL" {
			graphQLBindings++
		} else if generatedTransport {
			graphQLBindings++
			generatedGraphQLTransports++
		} else {
			restEndpoints++
		}
		if !generatedTransport {
			totalByMethod[ep.Method]++
		}
		if len(ep.CoveredBy) > 0 && !generatedTransport {
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
		if ep.Method == "POST" && ep.Path == githubRetiredUserDraftRESTPath {
			sawRetiredUserDraftRESTRoute = true
			if len(ep.CoveredBy) != 0 || ep.Operation == nil || ep.Operation.Model != "duplicate" ||
				ep.Operation.Status != "blocked" || ep.Operation.Risk != "low" ||
				!ep.Operation.BlockedByDefault || ep.Operation.DuplicateOf != githubUserDraftGraphQLTarget {
				t.Fatalf("retired user-project draft REST route = covered_by:%#v operation:%+v, want the fixed GraphQL command's blocked duplicate", ep.CoveredBy, ep.Operation)
			}
		}
	}

	// api_surface is the bundle's binding layer. Its REST population must match
	// the pinned source lock; its small GraphQL population consists only of
	// fixed-document bindings, never the GraphQL completeness denominator.
	if restEndpoints != lock.Counts.REST {
		t.Fatalf("REST endpoint bindings = %d, want %d from source lock", restEndpoints, lock.Counts.REST)
	}
	if graphQLBindings == 0 {
		t.Fatal("fixed GraphQL bindings = 0, want at least one bundle binding")
	}
	if generatedGraphQLTransports != 1 {
		t.Fatalf("generated GraphQL transports = %d, want one shared POST /graphql binding", generatedGraphQLTransports)
	}
	wantCovered := restEndpoints + graphQLBindings - generatedGraphQLTransports - 1
	if covered != wantCovered {
		t.Fatalf("covered endpoints = %d, want %d executable bindings plus one blocked duplicate", covered, wantCovered)
	}
	if !sawRetiredUserDraftRESTRoute {
		t.Fatalf("source inventory omits POST %s", githubRetiredUserDraftRESTPath)
	}
	if operations != 1 {
		t.Fatalf("operation endpoints = %d, want the retired user-project draft REST duplicate only", operations)
	}
	if excluded != 0 {
		t.Fatalf("legacy excluded endpoints = %d, want 0", excluded)
	}
	delete(totalByMethod, "GRAPHQL")
	assertStringIntMap(t, "totalByMethod", totalByMethod, githubRESTMethodSplit(lock))
	delete(coveredByMethod, "GRAPHQL")
	wantCoveredByMethod := githubRESTMethodSplit(lock)
	wantCoveredByMethod["POST"]--
	assertStringIntMap(t, "coveredByMethod", coveredByMethod, wantCoveredByMethod)
	assertStringIntMap(t, "operationByMethod", operationByMethod, map[string]int{"POST": 1})
	assertStringIntMap(t, "models", models, map[string]int{"duplicate": 1})
	assertStringIntMap(t, "risks", risks, map[string]int{"low": 1})
	assertStringIntMap(t, "statuses", statuses, map[string]int{"blocked": 1})
}

// TestGitHubStatusAndTextOperationContracts pins the classified endpoints
// which become executable only after the closed none/text/raw-body engine
// contracts exist. This is intentionally a concrete table rather than a
// promotion of every blocked direct_read row: each entry asserts its documented
// method, bounded response policy, input shape, and the real operation
// preflight the command uses.
func TestGitHubStatusAndTextOperationContracts(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load embedded github bundle: %v", err)
	}
	commands := map[string]engine.CLICommand{}
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Path] = command
	}
	operations := map[string]engine.OperationSpec{}
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}

	type flag struct {
		name, kind, mapsTo string
		required           bool
		values             []string
	}
	tests := []struct {
		command, method, endpoint, outputPolicy, contentType, bodyType string
		maxBytes                                                       int
		flags                                                          []flag
	}{
		{"gists star check", "GET", "/gists/{gist_id}/star", "none", "", "", 1024, []flag{{"gist-id", "string", "path.gist_id", true, nil}}},
		{"orgs blocks check", "GET", "/orgs/{org}/blocks/{username}", "none", "", "", 1024, []flag{{"org", "string", "path.org", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"orgs members check", "GET", "/orgs/{org}/members/{username}", "none", "", "", 1024, []flag{{"org", "string", "path.org", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"orgs public-members check", "GET", "/orgs/{org}/public_members/{username}", "none", "", "", 1024, []flag{{"org", "string", "path.org", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"teams members check", "GET", "/teams/{team_id}/members/{username}", "none", "", "", 1024, []flag{{"team-id", "integer", "path.team_id", true, nil}, {"username", "string", "path.username", true, nil}}},
		{"user blocks check", "GET", "/user/blocks/{username}", "none", "", "", 1024, []flag{{"username", "string", "path.username", true, nil}}},
		{"user following check", "GET", "/user/following/{username}", "none", "", "", 1024, []flag{{"username", "string", "path.username", true, nil}}},
		{"user starred check", "GET", "/user/starred/{owner}/{repo}", "none", "", "", 1024, []flag{{"owner", "string", "path.owner", true, nil}, {"repo", "string", "path.repo", true, nil}}},
		{"users following check", "GET", "/users/{username}/following/{target_user}", "none", "", "", 1024, []flag{{"username", "string", "path.username", true, nil}, {"target-user", "string", "path.target_user", true, nil}}},
		{"meta zen view", "GET", "/zen", "text", "", "", 1024, nil},
		{"meta octocat view", "GET", "/octocat", "text", "", "", 16384, []flag{{"s", "string", "query.s", false, nil}}},
		{"markdown render", "POST", "/markdown", "text", "application/json", "object", 1048576, []flag{{"text", "string", "body.text", true, nil}, {"mode", "enum", "body.mode", false, []string{"markdown", "gfm"}}, {"context", "string", "body.context", false, nil}}},
		{"markdown raw render", "POST", "/markdown/raw", "text", "text/plain", "string", 400 * 1024, []flag{{"text", "string", "body", true, nil}}},
	}

	for _, tt := range tests {
		t.Run(strings.ReplaceAll(tt.command, " ", "_"), func(t *testing.T) {
			command, ok := commands[tt.command]
			if !ok {
				t.Fatalf("generated command %q is missing", tt.command)
			}
			if command.Intent != "direct_read" || command.Availability != "implemented" {
				t.Fatalf("command = intent %q availability %q, want implemented direct_read", command.Intent, command.Availability)
			}
			if command.OutputPolicy != tt.outputPolicy {
				t.Fatalf("output_policy = %q, want %q", command.OutputPolicy, tt.outputPolicy)
			}
			op, ok := operations[command.Operation]
			if !ok || op.REST == nil {
				t.Fatalf("command operation %q = %#v, want REST operation", command.Operation, op)
			}
			if op.Kind != "rest_read" || op.REST.Method != tt.method || op.REST.Path != tt.endpoint || op.REST.MaxBytes != tt.maxBytes {
				t.Fatalf("operation = kind=%q method=%q path=%q max_bytes=%d, want rest_read %s %s %d", op.Kind, op.REST.Method, op.REST.Path, op.REST.MaxBytes, tt.method, tt.endpoint, tt.maxBytes)
			}
			if op.REST.ContentType != tt.contentType {
				t.Fatalf("content_type = %q, want %q", op.REST.ContentType, tt.contentType)
			}
			if tt.bodyType != "" {
				var schema map[string]any
				if err := json.Unmarshal(op.REST.BodySchema, &schema); err != nil {
					t.Fatalf("unmarshal body_schema: %v", err)
				}
				if schema["type"] != tt.bodyType {
					t.Fatalf("body_schema type = %#v, want %q", schema["type"], tt.bodyType)
				}
			}
			if err := engine.PreflightOperationDirectRead(bundle, command.Operation, tt.method, tt.endpoint, tt.maxBytes, command.OutputPolicy); err != nil {
				t.Fatalf("operation preflight: %v", err)
			}
			for _, want := range tt.flags {
				got, ok := githubCommandFlag(command, want.name)
				if !ok {
					t.Fatalf("flag --%s is missing from %#v", want.name, command.Flags)
				}
				if got.Type != want.kind || got.MapsTo != want.mapsTo || got.Required != want.required || !reflect.DeepEqual(got.Values, want.values) {
					t.Fatalf("flag --%s = type=%q maps_to=%q required=%t values=%#v, want type=%q maps_to=%q required=%t values=%#v", want.name, got.Type, got.MapsTo, got.Required, got.Values, want.kind, want.mapsTo, want.required, want.values)
				}
			}
		})
	}
}

// TestGitHubOneOfWriteContracts keeps one documented top-level oneOf arm from
// being flattened into a permissive command, or silently omitted because
// several named write actions share one provider endpoint. Every action below
// is a separate command contract: it names the concrete required fields, must
// survive the real runtime preflight, and must be linked from the documented
// endpoint through covered_by.writes.
func TestGitHubOneOfWriteContracts(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load embedded github bundle: %v", err)
	}
	rawSurface, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
	}
	surface, err := engine.ParseAPISurface(rawSurface)
	if err != nil {
		t.Fatalf("parse github api_surface.json: %v", err)
	}

	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector does not expose a command surface")
	}
	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, command := range provider.CommandSurface().Commands {
		commands[command.Path] = command
	}
	writes := map[string]engine.WriteAction{}
	for _, action := range bundle.Writes {
		writes[action.Name] = action
	}

	type contract struct {
		command, action, method, path, confirmation string
		required                                    []string
	}
	const destructive = string(connectors.ConfirmationKindDestructive)
	contracts := []contract{
		{"orgs attestations delete-by-subject-digests", "orgs_attestations_delete_request_by_subject_digests", "POST", "/orgs/{org}/attestations/delete-request", destructive, []string{"org", "subject_digests"}},
		{"orgs attestations delete-by-attestation-ids", "orgs_attestations_delete_request_by_attestation_ids", "POST", "/orgs/{org}/attestations/delete-request", destructive, []string{"attestation_ids", "org"}},
		{"users attestations delete-by-subject-digests", "users_attestations_delete_request_by_subject_digests", "POST", "/users/{username}/attestations/delete-request", destructive, []string{"subject_digests", "username"}},
		{"users attestations delete-by-attestation-ids", "users_attestations_delete_request_by_attestation_ids", "POST", "/users/{username}/attestations/delete-request", destructive, []string{"attestation_ids", "username"}},
		{"orgs campaigns create-code-scanning", "orgs_campaigns_create_code_scanning", "POST", "/orgs/{org}/campaigns", "", []string{"code_scanning_alerts", "description", "ends_at", "name", "org"}},
		{"orgs campaigns create-secret-scanning", "orgs_campaigns_create_secret_scanning", "POST", "/orgs/{org}/campaigns", "", []string{"description", "ends_at", "name", "org", "secret_scanning_alerts"}},
		{"orgs projects fields create-existing-issue-field", "orgs_projectsv2_fields_create_existing_issue_field", "POST", "/orgs/{org}/projectsV2/{project_number}/fields", "", []string{"issue_field_id", "org", "project_number"}},
		{"orgs projects fields create-new-field", "orgs_projectsv2_fields_create_new_field", "POST", "/orgs/{org}/projectsV2/{project_number}/fields", "", []string{"data_type", "name", "org", "project_number"}},
		{"orgs projects fields create-single-select", "orgs_projectsv2_fields_create_single_select", "POST", "/orgs/{org}/projectsV2/{project_number}/fields", "", []string{"data_type", "name", "org", "project_number", "single_select_options"}},
		{"orgs projects fields create-iteration", "orgs_projectsv2_fields_create_iteration", "POST", "/orgs/{org}/projectsV2/{project_number}/fields", "", []string{"data_type", "iteration_configuration", "name", "org", "project_number"}},
		{"users projects fields create-new-field", "users_projectsv2_fields_create_new_field", "POST", "/users/{username}/projectsV2/{project_number}/fields", "", []string{"data_type", "name", "project_number", "username"}},
		{"users projects fields create-single-select", "users_projectsv2_fields_create_single_select", "POST", "/users/{username}/projectsV2/{project_number}/fields", "", []string{"data_type", "name", "project_number", "single_select_options", "username"}},
		{"users projects fields create-iteration", "users_projectsv2_fields_create_iteration", "POST", "/users/{username}/projectsV2/{project_number}/fields", "", []string{"data_type", "iteration_configuration", "name", "project_number", "username"}},
		{"orgs projects items create-by-id", "orgs_projectsv2_items_create_by_id", "POST", "/orgs/{org}/projectsV2/{project_number}/items", "", []string{"id", "org", "project_number", "type"}},
		{"orgs projects items create-by-repo-number", "orgs_projectsv2_items_create_by_repo_number", "POST", "/orgs/{org}/projectsV2/{project_number}/items", "", []string{"number", "org", "owner", "project_number", "repo", "type"}},
		{"users projects items create-by-id", "users_projectsv2_items_create_by_id", "POST", "/users/{username}/projectsV2/{project_number}/items", "", []string{"id", "project_number", "type", "username"}},
		{"users projects items create-by-repo-number", "users_projectsv2_items_create_by_repo_number", "POST", "/users/{username}/projectsV2/{project_number}/items", "", []string{"number", "owner", "project_number", "repo", "type", "username"}},
		{"codespaces create-from-repository", "user_codespaces_create_from_repository", "POST", "/user/codespaces", "", []string{"repository_id"}},
		{"codespaces create-from-pull-request", "user_codespaces_create_from_pull_request", "POST", "/user/codespaces", "", []string{"pull_request"}},
	}

	endpointActions := map[string][]string{}
	for _, want := range contracts {
		t.Run(strings.ReplaceAll(want.command, " ", "_"), func(t *testing.T) {
			command, ok := commands[want.command]
			if !ok {
				t.Fatalf("generated command %q is missing", want.command)
			}
			if command.Intent != "reverse_etl" || command.Availability != "implemented" || command.Write != want.action {
				t.Fatalf("command = intent=%q availability=%q write=%q, want implemented reverse_etl write=%q", command.Intent, command.Availability, command.Write, want.action)
			}
			if got := commandrunner.ConfirmationChallengeForCommand(connector, command); got != want.confirmation {
				t.Fatalf("confirmation = %q, want %q", got, want.confirmation)
			}
			if err := commandrunner.Preflight(connector, strings.Fields(want.command)); err != nil {
				t.Fatalf("runtime preflight: %v", err)
			}

			action, ok := writes[want.action]
			if !ok {
				t.Fatalf("write action %q is missing", want.action)
			}
			wantWritePath := githubEngineWritePath(want.path)
			if action.Method != want.method || action.Path != wantWritePath {
				t.Fatalf("write action = %s %s, want %s %s", action.Method, action.Path, want.method, wantWritePath)
			}
			if err := engine.ValidatePromotableRecordSchema(action.RecordSchema); err != nil {
				t.Fatalf("write record_schema must be a concrete arm: %v", err)
			}

			var schema struct {
				Required []string `json:"required"`
			}
			if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
				t.Fatalf("unmarshal record_schema: %v", err)
			}
			gotRequired := append([]string(nil), schema.Required...)
			wantRequired := append([]string(nil), want.required...)
			sort.Strings(gotRequired)
			sort.Strings(wantRequired)
			if !reflect.DeepEqual(gotRequired, wantRequired) {
				t.Fatalf("record_schema required = %#v, want %#v", gotRequired, wantRequired)
			}
			for _, field := range wantRequired {
				var mapped *connectors.CommandSurfaceFlag
				for i := range command.Flags {
					if command.Flags[i].MapsTo == "record."+field {
						mapped = &command.Flags[i]
						break
					}
				}
				if mapped == nil || !mapped.Required {
					t.Fatalf("required record field %q has no required command flag: %#v", field, command.Flags)
				}
			}

			key := want.method + " " + want.path
			endpointActions[key] = append(endpointActions[key], want.action)
		})
	}

	for _, endpoint := range surface.Endpoints {
		key := endpoint.Method + " " + endpoint.Path
		want, ok := endpointActions[key]
		if !ok {
			continue
		}
		if endpoint.CoveredBy == nil {
			t.Fatalf("%s is not covered by write actions", key)
		}
		got := append([]string(nil), endpoint.CoveredBy.WriteTargets()...)
		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s covered_by.writes = %#v, want %#v", key, got, want)
		}
		delete(endpointActions, key)
	}
	if len(endpointActions) != 0 {
		t.Fatalf("oneOf endpoints absent from api_surface: %#v", endpointActions)
	}
}

func githubCommandFlag(command engine.CLICommand, name string) (engine.CLIFlag, bool) {
	for _, flag := range command.Flags {
		if flag.Name == name {
			return flag, true
		}
	}
	return engine.CLIFlag{}, false
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

func githubEngineWritePath(apiPath string) string {
	var out strings.Builder
	for index := 0; index < len(apiPath); {
		start := strings.IndexByte(apiPath[index:], '{')
		if start < 0 {
			out.WriteString(apiPath[index:])
			break
		}
		start += index
		out.WriteString(apiPath[index:start])
		end := strings.IndexByte(apiPath[start:], '}')
		if end < 0 {
			out.WriteString(apiPath[start:])
			break
		}
		end += start
		name := apiPath[start+1 : end]
		namespace := "record"
		if name == "owner" || name == "repo" {
			namespace = "config"
		}
		out.WriteString("{{ ")
		out.WriteString(namespace)
		out.WriteByte('.')
		out.WriteString(name)
		out.WriteString(" }}")
		index = end + 1
	}
	return out.String()
}
