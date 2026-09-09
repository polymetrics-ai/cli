package cli_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// TestBatch1CP18GitLabG2SourceBounds keeps the four omitted source numeric
// constraints tied to the retained capture. Each invalid value gets an
// independent project whose healthy control has already reached planning, so
// a plan accidentally created by one negative cannot pollute another oracle.
func TestBatch1CP18GitLabG2SourceBounds(t *testing.T) {
	fixture := loadGitLabG2InputFixture(t)
	source := loadGitLabG2SourceSchemas(t, fixture.SourceSHA256)
	bundle, err := engine.Load(defs.FS, "gitlab")
	if err != nil {
		t.Fatal(err)
	}
	writes := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, write := range bundle.Writes {
		writes[write.Name] = write
	}

	for _, want := range []gitLabG2Bound{
		{sourceID: "postApiV4ImportGithub", field: "pagination_limit", minimum: 1, maximum: 100},
		{sourceID: "postApiV4TokenExchange", field: "expires_in", minimum: 1, maximum: 43200},
		{sourceID: "putApiV4ElasticsearchIndexedNamespacesRollback", field: "percentage", minimum: 0, maximum: 100},
		{sourceID: "putApiV4ElasticsearchIndexedNamespacesRollout", field: "percentage", minimum: 0, maximum: 100},
	} {
		input, ok := fixture.input(want.sourceID)
		if !ok {
			t.Fatalf("retained negative fixture lacks %s", want.sourceID)
		}
		schema, ok := source.schema(want.sourceID, want.field)
		if !ok {
			t.Fatalf("retained source capture lacks %s/%s", want.sourceID, want.field)
		}
		if schema.Type != "integer" || schema.Minimum != want.minimum || schema.Maximum != want.maximum {
			t.Fatalf("retained source %s/%s = %+v, want integer %d..%d", want.sourceID, want.field, schema, want.minimum, want.maximum)
		}

		writeName := "source_write_" + hex.EncodeToString([]byte(strings.ToUpper(schema.Method)+" "+strings.TrimPrefix(schema.Path, "/api/v4")))
		write, ok := writes[writeName]
		if !ok {
			t.Fatalf("source-to-write join %s -> %s is absent", want.sourceID, writeName)
		}
		if got := gitLabG2WriteConstraint(t, write, want.field); got.Type != "integer" || got.Nullable != schema.Nullable || got.Minimum != want.minimum || got.Maximum != want.maximum {
			t.Errorf("G2 runtime declaration %s/%s = %+v, want source %+v", want.sourceID, want.field, got, schema)
		}

		for _, value := range []struct {
			name  string
			value int
		}{
			{name: "minimum", value: want.minimum},
			{name: "interior", value: want.minimum + (want.maximum-want.minimum)/2},
			{name: "maximum", value: want.maximum},
		} {
			t.Run(want.sourceID+"/valid_"+value.name, func(t *testing.T) {
				gitLabG2AssertPlanFrontier(t, input, want.field, value.value, true)
			})
		}
		for _, value := range []struct {
			name  string
			value int
		}{
			{name: "below_minimum", value: want.minimum - 1},
			{name: "above_maximum", value: want.maximum + 1},
		} {
			t.Run(want.sourceID+"/"+value.name, func(t *testing.T) {
				gitLabG2AssertPlanFrontier(t, input, want.field, value.value, false)
			})
		}
	}
}

type gitLabG2Bound struct {
	sourceID, field  string
	minimum, maximum int
}

type gitLabG2Fixture struct {
	SourceSHA256 string `json:"source_sha256"`
	Sources      []gitLabG2FixtureInput
}

type gitLabG2FixtureInput struct {
	SourceID   string   `json:"source_id"`
	Command    string   `json:"command"`
	ValidFlags []string `json:"valid_flags"`
}

func (fixture gitLabG2Fixture) input(sourceID string) (gitLabG2FixtureInput, bool) {
	for _, input := range fixture.Sources {
		if input.SourceID == sourceID {
			return input, true
		}
	}
	return gitLabG2FixtureInput{}, false
}

func loadGitLabG2InputFixture(t *testing.T) gitLabG2Fixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "batch1-cp18-gitlab", "remaining252", "input-refusals.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture gitLabG2Fixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SourceSHA256 == "" || len(fixture.Sources) == 0 {
		t.Fatal("G2 retained input fixture is empty")
	}
	return fixture
}

type gitLabG2SourceSchemas struct {
	operations map[string]gitLabG2SourceOperation
	schemas    map[string]gitLabG2SourceObjectSchema
}

type gitLabG2SourceOperation struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Source struct {
		RequestBody struct {
			Content map[string]struct {
				Schema struct {
					Ref string `json:"$ref"`
				} `json:"schema"`
			} `json:"content"`
		} `json:"requestBody"`
	} `json:"source_operation"`
}

type gitLabG2SourceObjectSchema struct {
	Properties map[string]gitLabG2Constraint `json:"properties"`
}

type gitLabG2Constraint struct {
	Type     string `json:"type"`
	Minimum  int    `json:"minimum"`
	Maximum  int    `json:"maximum"`
	Nullable bool   `json:"nullable"`
	Method   string
	Path     string
}

func loadGitLabG2SourceSchemas(t *testing.T, wantSHA256 string) gitLabG2SourceSchemas {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "internal", "connectors", "defs", "gitlab", "sources", "gitlab-operation-source-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != wantSHA256 {
		t.Fatal("G2 retained input fixture source pin does not match capture")
	}
	var source struct {
		REST struct {
			Operations []gitLabG2SourceOperation `json:"operations"`
		} `json:"rest"`
		SourceContract struct {
			Components struct {
				Schemas map[string]gitLabG2SourceObjectSchema `json:"schemas"`
			} `json:"components"`
		} `json:"source_contract"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	out := gitLabG2SourceSchemas{operations: make(map[string]gitLabG2SourceOperation, len(source.REST.Operations)), schemas: source.SourceContract.Components.Schemas}
	for _, operation := range source.REST.Operations {
		if _, exists := out.operations[operation.ID]; exists {
			t.Fatalf("duplicate retained G2 source identity %s", operation.ID)
		}
		out.operations[operation.ID] = operation
	}
	return out
}

func (source gitLabG2SourceSchemas) schema(sourceID, field string) (gitLabG2Constraint, bool) {
	operation, ok := source.operations[sourceID]
	if !ok {
		return gitLabG2Constraint{}, false
	}
	content, ok := operation.Source.RequestBody.Content["application/json"]
	if !ok {
		return gitLabG2Constraint{}, false
	}
	name := strings.TrimPrefix(content.Schema.Ref, "#/components/schemas/")
	schema, ok := source.schemas[name]
	if !ok {
		return gitLabG2Constraint{}, false
	}
	constraint, ok := schema.Properties[field]
	constraint.Method, constraint.Path = operation.Method, operation.Path
	return constraint, ok
}

func gitLabG2WriteConstraint(t *testing.T, write engine.WriteAction, field string) gitLabG2Constraint {
	t.Helper()
	var raw struct {
		Properties map[string]struct {
			Type    json.RawMessage `json:"type"`
			Minimum int             `json:"minimum"`
			Maximum int             `json:"maximum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(write.RecordSchema, &raw); err != nil {
		t.Fatal(err)
	}
	fieldSchema, ok := raw.Properties[field]
	if !ok {
		t.Fatalf("write %s lacks %s record field", write.Name, field)
	}
	constraint := gitLabG2Constraint{Minimum: fieldSchema.Minimum, Maximum: fieldSchema.Maximum}
	var typeSet []string
	if err := json.Unmarshal(fieldSchema.Type, &typeSet); err == nil {
		constraint.Nullable = false
		for _, item := range typeSet {
			if item == "null" {
				constraint.Nullable = true
			}
			if item == "integer" {
				constraint.Type = item
			}
		}
		return constraint
	}
	if err := json.Unmarshal(fieldSchema.Type, &constraint.Type); err != nil {
		t.Fatal(err)
	}
	return constraint
}

func gitLabG2AssertPlanFrontier(t *testing.T, input gitLabG2FixtureInput, field string, value int, valid bool) {
	t.Helper()
	var sends atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		sends.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	root := cp18Lane240Project(t, server.URL)
	invoke := func(flags []string) (string, string, int) {
		args := append([]string{"gitlab"}, strings.Fields(input.Command)...)
		args = append(args, flags...)
		args = append(args, "--credential", "cp18-lane240", "--root", root, "--json")
		return cp18Lane240Run(args)
	}
	interior := value
	if !valid {
		interior = 50
	}
	controlOutput, controlDiag, controlCode := invoke(gitLabG2WithFlag(input.ValidFlags, "--"+strings.ReplaceAll(field, "_", "-"), interior))
	if controlCode != 0 {
		t.Fatalf("healthy planning control source=%s field=%s code=%d output=%s error=%s", input.SourceID, field, controlCode, controlOutput, controlDiag)
	}
	if sends.Load() != 0 {
		t.Fatal("healthy planning control reached provider")
	}
	before := gitLabG2Plans(t, root)
	if valid {
		return
	}
	output, diag, code := invoke(gitLabG2WithFlag(input.ValidFlags, "--"+strings.ReplaceAll(field, "_", "-"), value))
	if code == 0 {
		t.Fatalf("source-invalid input planned successfully: source=%s field=%s value=%d output=%s", input.SourceID, field, value, output)
	}
	if !strings.Contains(output+diag, field) && !strings.Contains(output+diag, strings.ReplaceAll(field, "_", "-")) {
		t.Fatalf("source-invalid refusal loses field %s: output=%s error=%s", field, output, diag)
	}
	if sends.Load() != 0 {
		t.Fatal("source-invalid input reached provider")
	}
	if after := gitLabG2Plans(t, root); after != before {
		t.Fatal("source-invalid input changed retained approved-plan candidates")
	}
}

func gitLabG2WithFlag(flags []string, flag string, value int) []string {
	out := make([]string, 0, len(flags)+2)
	for index := 0; index < len(flags); index++ {
		if flags[index] == flag {
			index++
			continue
		}
		out = append(out, flags[index])
	}
	return append(out, flag, strconv.Itoa(value))
}

func gitLabG2Plans(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]json.RawMessage
	if err := json.Unmarshal(raw, &state); err != nil {
		t.Fatal(err)
	}
	plans, ok := state["reverse_plans"]
	if !ok || len(plans) == 0 {
		t.Fatal("healthy planning control lacks reverse plan state")
	}
	return string(plans)
}
