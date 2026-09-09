package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// TestBatch1CP18GitLabG1SourceScalarInputJoins keeps the G1 repair tied to
// the immutable retained source capture. The existing remaining252 fixture
// names every selected request whose generated CLI surface omitted one scalar
// source query input. The three scope oneOf(string,array) joins are explicitly
// not lowered here: they are the separate structured-query G6 contract.
func TestBatch1CP18GitLabG1SourceScalarInputJoins(t *testing.T) {
	fixture := loadGitLabG1ReadFixture(t)
	source := loadGitLabG1SourceOperations(t, fixture.SourceSHA256)
	bundle, err := engine.Load(defs.FS, "gitlab")
	if err != nil {
		t.Fatal(err)
	}

	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}
	if bundle.CLISurface == nil {
		t.Fatal("GitLab bundle lacks generated CLI surface")
	}
	commands := make(map[string][]engine.CLICommand)
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Operation] = append(commands[command.Operation], command)
	}

	type selectedInput struct {
		sourceID  string
		operation string
		name      string
	}
	selected := map[string]selectedInput{}
	for _, testCase := range fixture.Cases {
		if testCase.CLIInputsJoined {
			continue
		}
		for _, name := range gitLabG1MissingScalarInputs(testCase.Query, testCase.CLIFlags) {
			key := testCase.SourceID + "/" + name
			selected[key] = selectedInput{sourceID: testCase.SourceID, operation: testCase.Operation, name: name}
		}
	}

	counts := map[string]int{}
	keys := make([]string, 0, len(selected))
	for key := range selected {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		want := selected[key]
		counts[want.name]++
		sourceOperation, ok := source[want.sourceID]
		if !ok {
			t.Fatalf("selected source identity %q is absent from retained capture", want.sourceID)
		}
		sourceParameter, ok := sourceOperation.parameter(want.name)
		if !ok {
			t.Fatalf("selected source input %s/%s is absent from retained capture", want.sourceID, want.name)
		}
		if sourceParameter.In != "query" || sourceParameter.Type != "integer" {
			t.Fatalf("G1 input %s/%s = in=%q type=%q, want retained scalar integer query input", want.sourceID, want.name, sourceParameter.In, sourceParameter.Type)
		}

		operation, ok := operations[want.operation]
		if !ok || operation.REST == nil {
			t.Fatalf("source-to-generated operation join %s -> %s is not a REST operation", want.sourceID, want.operation)
		}
		parameter, ok := gitLabG1OperationParameter(operation.REST.Parameters, want.name)
		if !ok {
			t.Errorf("G1 declaration missing %s/%s on generated operation %s", want.sourceID, want.name, want.operation)
			continue
		}
		if parameter.In != "query" || parameter.Type != sourceParameter.Type || parameter.Required != sourceParameter.Required || gitLabG1ExactNumber(parameter.Minimum) != sourceParameter.Minimum || gitLabG1ExactNumber(parameter.Maximum) != sourceParameter.Maximum {
			t.Errorf("G1 declaration mismatch %s/%s: generated=%+v source={in:%s type:%s required:%t minimum:%s maximum:%s}", want.sourceID, want.name, parameter, sourceParameter.In, sourceParameter.Type, sourceParameter.Required, sourceParameter.Minimum, sourceParameter.Maximum)
		}
		wantFlag := gitLabG1FlagName(want.name)
		matched := false
		for _, command := range commands[want.operation] {
			for _, flag := range command.Flags {
				if flag.Name == wantFlag && flag.MapsTo == "query."+want.name && flag.Type == sourceParameter.Type && gitLabG1ExactNumber(flag.Minimum) == sourceParameter.Minimum && gitLabG1ExactNumber(flag.Maximum) == sourceParameter.Maximum {
					matched = true
				}
			}
		}
		if !matched {
			t.Errorf("G1 CLI join missing %s/%s: want --%s mapped to query.%s", want.sourceID, want.name, wantFlag, want.name)
		}
	}
	if len(selected) != 203 || counts["per_page"] != 200 || counts["max_results"] != 2 || counts["limit"] != 1 {
		t.Fatalf("G1 retained selected scalar inputs=%d counts=%v, want 203 with per_page=200 max_results=2 limit=1", len(selected), counts)
	}
}

// TestBatch1CP18GitLabG1GeneratedCLIWire uses three independently retained
// source identities to prove the generated CLI, rather than a parser helper,
// transports each scalar class. A provider-owned `limit` is deliberately
// rendered as provider-limit because PM owns the unqualified control spelling.
func TestBatch1CP18GitLabG1GeneratedCLIWire(t *testing.T) {
	fixture := loadGitLabG1ReadFixture(t)
	for _, sourceID := range []string{
		"getApiV4AdminCiVariables",
		"getApiV4ProjectsIdMlMlflowApi20MlflowMetricsGetHistory",
		"getApiV4ProjectsIdDashSearchSemantic",
	} {
		testCase, ok := fixture.caseFor(sourceID)
		if !ok {
			t.Fatalf("retained G1 source fixture lacks %s", sourceID)
		}
		t.Run(sourceID, func(t *testing.T) {
			var mu sync.Mutex
			var requests []struct{ method, uri, auth string }
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				_, _ = io.ReadAll(io.LimitReader(request.Body, 1024))
				mu.Lock()
				requests = append(requests, struct{ method, uri, auth string }{request.Method, request.URL.RequestURI(), request.Header.Get("Authorization")})
				mu.Unlock()
				w.Header().Set("Content-Type", testCase.ResponseMedia)
				w.WriteHeader(testCase.ResponseStatus)
				_, _ = w.Write(testCase.ResponseBody)
			}))
			t.Cleanup(server.Close)

			root := t.TempDir()
			runCLI(t, []string{"init", "--root", root, "--json"})
			t.Setenv("PM_CP18_G1_TOKEN", "cp18-g1-fake")
			runCLI(t, []string{"credentials", "add", "cp18-g1", "--connector", "gitlab", "--from-env", "access_token=PM_CP18_G1_TOKEN", "--config", "base_url=" + server.URL + "/api/v4", "--root", root, "--json"})
			args := append([]string{"gitlab"}, strings.Fields(testCase.Command)...)
			args = append(args, testCase.CLIFlags...)
			for _, name := range gitLabG1MissingScalarInputs(testCase.Query, testCase.CLIFlags) {
				args = append(args, "--"+gitLabG1FlagName(name), testCase.Query[name])
			}
			args = append(args, "--credential", "cp18-g1", "--root", root, "--json")
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("generated G1 CLI source=%s code=%d output=%s error=%s", sourceID, code, stdout.String(), stderr.String())
			}
			mu.Lock()
			captured := append([]struct{ method, uri, auth string }(nil), requests...)
			mu.Unlock()
			if len(captured) != 1 {
				t.Fatalf("generated G1 CLI source=%s physical sends=%d, want 1", sourceID, len(captured))
			}
			const expectedAuth = "Bearer cp18-g1-fake"
			if got := captured[0]; got.method != testCase.ExpectedMethod || got.uri != testCase.ExpectedURI || got.auth != expectedAuth {
				t.Fatalf("generated G1 CLI source=%s wire=%+v, want method=%s uri=%s auth=%s", sourceID, got, testCase.ExpectedMethod, testCase.ExpectedURI, expectedAuth)
			}
		})
	}
}

type gitLabG1Fixture struct {
	SourceSHA256 string `json:"source_sha256"`
	Cases        []gitLabG1ReadCase
}

type gitLabG1ReadCase struct {
	SourceID        string            `json:"source_id"`
	Command         string            `json:"command"`
	CLIFlags        []string          `json:"cli_flags"`
	CLIInputsJoined bool              `json:"cli_inputs_joined"`
	Operation       string            `json:"operation"`
	Variant         string            `json:"variant"`
	Query           map[string]string `json:"query"`
	ExpectedMethod  string            `json:"expected_method"`
	ExpectedURI     string            `json:"expected_uri"`
	ExpectedAuth    string            `json:"expected_auth"`
	ResponseStatus  int               `json:"response_status"`
	ResponseMedia   string            `json:"response_media"`
	ResponseBody    json.RawMessage   `json:"response_body"`
}

func loadGitLabG1ReadFixture(t *testing.T) gitLabG1Fixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "batch1-cp18-gitlab", "remaining252", "read-contracts.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture gitLabG1Fixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SourceSHA256 == "" || len(fixture.Cases) == 0 {
		t.Fatal("G1 retained fixture has no source pin or cases")
	}
	return fixture
}

func (fixture gitLabG1Fixture) caseFor(sourceID string) (gitLabG1ReadCase, bool) {
	for _, testCase := range fixture.Cases {
		if testCase.SourceID == sourceID && testCase.Variant == "all_scalar_inputs_json200" && !testCase.CLIInputsJoined {
			return testCase, true
		}
	}
	return gitLabG1ReadCase{}, false
}

type gitLabG1SourceOperation struct {
	ID        string `json:"id"`
	Operation struct {
		Parameters []struct {
			Name     string                     `json:"name"`
			In       string                     `json:"in"`
			Required bool                       `json:"required"`
			Schema   map[string]json.RawMessage `json:"schema"`
		} `json:"parameters"`
	} `json:"source_operation"`
}

type gitLabG1SourceParameter struct {
	In, Type         string
	Required         bool
	Minimum, Maximum string
}

func (operation gitLabG1SourceOperation) parameter(name string) (gitLabG1SourceParameter, bool) {
	for _, parameter := range operation.Operation.Parameters {
		if parameter.Name != name {
			continue
		}
		var kind string
		_ = json.Unmarshal(parameter.Schema["type"], &kind)
		return gitLabG1SourceParameter{
			In: parameter.In, Type: kind, Required: parameter.Required,
			Minimum: string(parameter.Schema["minimum"]), Maximum: string(parameter.Schema["maximum"]),
		}, true
	}
	return gitLabG1SourceParameter{}, false
}

func loadGitLabG1SourceOperations(t *testing.T, wantSHA256 string) map[string]gitLabG1SourceOperation {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "gitlab", "sources", "gitlab-operation-source-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != wantSHA256 {
		t.Fatal("G1 source fixture pin does not match retained source capture")
	}
	var source struct {
		REST struct {
			Operations []gitLabG1SourceOperation `json:"operations"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	operations := make(map[string]gitLabG1SourceOperation, len(source.REST.Operations))
	for _, operation := range source.REST.Operations {
		if _, exists := operations[operation.ID]; exists {
			t.Fatalf("duplicate retained G1 source identity %s", operation.ID)
		}
		operations[operation.ID] = operation
	}
	return operations
}

func gitLabG1MissingScalarInputs(query map[string]string, flags []string) []string {
	existing := map[string]bool{}
	for index := 0; index+1 < len(flags); index += 2 {
		existing[strings.TrimPrefix(flags[index], "--")] = true
	}
	var missing []string
	for _, name := range []string{"per_page", "max_results", "limit"} {
		if _, exists := query[name]; !exists || existing[gitLabG1FlagName(name)] {
			continue
		}
		missing = append(missing, name)
	}
	return missing
}

func gitLabG1FlagName(name string) string {
	if name == "limit" {
		return "provider-limit"
	}
	return strings.ReplaceAll(name, "_", "-")
}

func gitLabG1OperationParameter(parameters []engine.OperationParameter, name string) (engine.OperationParameter, bool) {
	for _, parameter := range parameters {
		if parameter.Name == name {
			return parameter, true
		}
	}
	return engine.OperationParameter{}, false
}

func gitLabG1ExactNumber(value *connectors.ExactNumber) string {
	if value == nil {
		return ""
	}
	return value.String()
}
