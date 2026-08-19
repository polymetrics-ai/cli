package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The params mechanism exists so a command's flags are derived from the
// provider's own specification rather than hand-authored against it. These
// tests pin the two rules that make the derivation safe to run across a fleet:
// paging parameters never become flags, and neither does anything the
// connection already supplies.

func writeParamsFixture(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", path, err)
		}
	}
}

const paramsFixtureOperations = `{
  "operations": [
    {
      "id": "acme.list_things",
      "kind": "rest_read",
      "summary": "List things",
      "risk": "low",
      "approval": "none",
      "output_policy": "json_redacted",
      "rest": { "method": "GET", "path": "/orgs/{org}/things", "max_bytes": 1048576 }
    }
  ]
}`

const paramsFixtureStreams = `{
  "base": {
    "url": "https://api.acme.test",
    "pagination": { "type": "page_number", "page_param": "page", "size_param": "per_page", "page_size": 100 }
  },
  "streams": []
}`

const paramsFixtureSpec = `{
  "type": "object",
  "properties": { "org": { "type": "string" }, "token": { "type": "string" } }
}`

const paramsFixtureArtifact = `{
  "openapi": "3.0.3",
  "paths": {
    "/orgs/{org}/things": {
      "get": {
        "parameters": [
          { "name": "org", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "state", "in": "query", "description": "Filter by state.\nSecond line ignored.", "schema": { "type": "string", "enum": ["open", "closed"] } },
          { "name": "since", "in": "query", "required": true, "schema": { "type": "string" } },
          { "name": "count", "in": "query", "schema": { "type": "integer" } },
          { "$ref": "#/components/parameters/page" },
          { "$ref": "#/components/parameters/per-page" },
          { "name": "accept", "in": "header", "schema": { "type": "string" } }
        ]
      }
    }
  },
  "components": {
    "parameters": {
      "page": { "name": "page", "in": "query", "schema": { "type": "integer" } },
      "per-page": { "name": "per_page", "in": "query", "schema": { "type": "integer" } }
    }
  }
}`

func setupParamsFixture(t *testing.T) (defsDir, artifact string) {
	t.Helper()
	root := t.TempDir()
	defsDir = filepath.Join(root, "defs")
	writeParamsFixture(t, filepath.Join(defsDir, "acme"), map[string]string{
		"operations.json": paramsFixtureOperations,
		"streams.json":    paramsFixtureStreams,
		"spec.json":       paramsFixtureSpec,
	})
	artifact = filepath.Join(root, "artifact.json")
	if err := os.WriteFile(artifact, []byte(paramsFixtureArtifact), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	return defsDir, artifact
}

func importedFixtureParameters(t *testing.T, defsDir string) []map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(defsDir, "acme", "operations.json"))
	if err != nil {
		t.Fatalf("read operations.json: %v", err)
	}
	var doc struct {
		Operations []struct {
			REST struct {
				Parameters []map[string]any `json:"parameters"`
			} `json:"rest"`
		} `json:"operations"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse operations.json: %v", err)
	}
	if len(doc.Operations) != 1 {
		t.Fatalf("operations = %d, want 1", len(doc.Operations))
	}
	return doc.Operations[0].REST.Parameters
}

// TestParamsImportExcludesPagingAndConnectionParameters is the load-bearing
// rule: paging comes from the declared pagination spec and the connection
// supplies its own config, so neither may ever reach a command as a flag.
func TestParamsImportExcludesPagingAndConnectionParameters(t *testing.T) {
	defsDir, artifact := setupParamsFixture(t)

	changed, total, err := importConnectorParameters(paramsImportOptions{
		connector: "acme", artifact: artifact, defsDir: defsDir,
	})
	if err != nil {
		t.Fatalf("importConnectorParameters: %v", err)
	}
	if total != 1 || changed != 1 {
		t.Fatalf("scanned/changed = %d/%d, want 1/1", total, changed)
	}

	got := importedFixtureParameters(t, defsDir)
	names := make([]string, 0, len(got))
	for _, param := range got {
		names = append(names, param["name"].(string))
	}
	joined := strings.Join(names, ",")
	if joined != "count,since,state" {
		t.Fatalf("imported parameters = %q, want \"count,since,state\"", joined)
	}
	for _, banned := range []string{"page", "per_page", "org", "accept"} {
		for _, name := range names {
			if name == banned {
				t.Fatalf("imported %q, which must never become a flag", banned)
			}
		}
	}
}

func TestParamsImportCarriesEnumRequirednessAndSummary(t *testing.T) {
	defsDir, artifact := setupParamsFixture(t)
	if _, _, err := importConnectorParameters(paramsImportOptions{
		connector: "acme", artifact: artifact, defsDir: defsDir,
	}); err != nil {
		t.Fatalf("importConnectorParameters: %v", err)
	}

	byName := map[string]map[string]any{}
	for _, param := range importedFixtureParameters(t, defsDir) {
		byName[param["name"].(string)] = param
	}

	state := byName["state"]
	values, _ := state["values"].([]any)
	if len(values) != 2 || values[0] != "open" || values[1] != "closed" {
		t.Fatalf("state values = %v, want [open closed]", values)
	}
	if summary, _ := state["summary"].(string); summary != "Filter by state." {
		t.Fatalf("state summary = %q, want the first description line only", summary)
	}
	if required, ok := state["required"]; ok && required == true {
		t.Fatal("state must not be marked required")
	}
	if required, _ := byName["since"]["required"].(bool); !required {
		t.Fatal("since must be marked required")
	}
	if typ, _ := byName["count"]["type"].(string); typ != "integer" {
		t.Fatalf("count type = %q, want integer", typ)
	}
}

func TestParamsImportImportsOnlyBoundedTypedHeaders(t *testing.T) {
	defsDir, artifact := setupParamsFixture(t)
	raw, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	updated := strings.Replace(string(raw),
		`{ "name": "accept", "in": "header", "schema": { "type": "string" } }`,
		`{ "name": "accept", "in": "header", "schema": { "type": "string" } },
          { "name": "X-Request-Mode", "in": "header", "required": true, "schema": { "type": "string", "enum": ["safe", "full"], "pattern": "^(safe|full)$", "minLength": 4, "maxLength": 4 } }`,
		1,
	)
	if err := os.WriteFile(artifact, []byte(updated), 0o644); err != nil {
		t.Fatalf("write bounded-header artifact: %v", err)
	}
	if _, _, err := importConnectorParameters(paramsImportOptions{connector: "acme", artifact: artifact, defsDir: defsDir}); err != nil {
		t.Fatalf("importConnectorParameters: %v", err)
	}
	byName := map[string]map[string]any{}
	for _, parameter := range importedFixtureParameters(t, defsDir) {
		byName[parameter["name"].(string)] = parameter
	}
	if _, found := byName["accept"]; found {
		t.Fatal("unbounded header was imported")
	}
	header, found := byName["X-Request-Mode"]
	if !found {
		t.Fatalf("parameters = %#v, want bounded header", byName)
	}
	if header["in"] != "header" || header["type"] != "string" || header["max_bytes"] != float64(16) || header["required"] != true {
		t.Fatalf("header import = %#v, want bounded required string header", header)
	}
	schema, ok := header["schema"].(map[string]any)
	if !ok || schema["pattern"] != "^(safe|full)$" || schema["maxLength"] != float64(4) {
		t.Fatalf("header schema = %#v, want imported string constraints", header["schema"])
	}
	changed, _, err := importConnectorParameters(paramsImportOptions{connector: "acme", artifact: artifact, defsDir: defsDir})
	if err != nil || changed != 0 {
		t.Fatalf("bounded header re-import changed/err = %d/%v, want 0/nil", changed, err)
	}
}

// TestParamsImportIsIdempotent guards --check: a rerun against the same
// artifact must report no drift, or CI would fail on an unchanged bundle.
func TestParamsImportIsIdempotent(t *testing.T) {
	defsDir, artifact := setupParamsFixture(t)
	opts := paramsImportOptions{connector: "acme", artifact: artifact, defsDir: defsDir}
	if _, _, err := importConnectorParameters(opts); err != nil {
		t.Fatalf("first import: %v", err)
	}
	changed, _, err := importConnectorParameters(opts)
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if changed != 0 {
		t.Fatalf("second import changed %d operation(s), want 0", changed)
	}
}

// TestDeriveCommandParameterFlagsOnlyAdds pins that the derivation never
// overwrites an authored flag: an author's narrower type or better summary is
// judgement the generator has no basis to replace.
func TestDeriveCommandParameterFlagsOnlyAdds(t *testing.T) {
	var cli orderedJSON
	if err := json.Unmarshal([]byte(`{"commands":[{"path":"things list","flags":[{"name":"state","type":"string","summary":"authored"}]}]}`), &cli); err != nil {
		t.Fatalf("parse cli fixture: %v", err)
	}
	var ops orderedJSON
	if err := json.Unmarshal([]byte(`{"rest":{"parameters":[{"name":"state","in":"query","type":"string","values":["open","closed"]},{"name":"since","in":"query","type":"string","required":true}]}}`), &ops); err != nil {
		t.Fatalf("parse ops fixture: %v", err)
	}
	cmd := arrayField(cli.root, "commands")[0].(*orderedObject)
	restRaw, _ := ops.root.get("rest")
	rest := restRaw.(*orderedObject)

	added := deriveCommandParameterFlags(cmd, rest)
	if added != 1 {
		t.Fatalf("added = %d, want 1 (since only; state was already authored)", added)
	}
	flags := arrayField(cmd, "flags")
	if len(flags) != 2 {
		t.Fatalf("flags = %d, want 2", len(flags))
	}
	state := flags[0].(*orderedObject)
	if stringField(state, "summary") != "authored" || stringField(state, "type") != "string" {
		t.Fatal("authored flag was overwritten; the derivation must only add")
	}
	since := flags[1].(*orderedObject)
	if stringField(since, "name") != "since" || stringField(since, "maps_to") != "query.since" {
		t.Fatalf("derived flag = %v, want since -> query.since", since.values)
	}
	if required, _ := since.get("required"); required != true {
		t.Fatal("derived flag lost its requiredness")
	}
}

func TestDeriveCommandParameterFlagsAddsExactHeaderMapping(t *testing.T) {
	var cli orderedJSON
	if err := json.Unmarshal([]byte(`{"commands":[{"path":"things list"}]}`), &cli); err != nil {
		t.Fatalf("parse cli fixture: %v", err)
	}
	var ops orderedJSON
	if err := json.Unmarshal([]byte(`{"rest":{"parameters":[{"name":"X-Request-Mode","in":"header","type":"string","values":["safe","full"],"required":true,"schema":{"type":"string"},"max_bytes":16}]}}`), &ops); err != nil {
		t.Fatalf("parse operation fixture: %v", err)
	}
	cmd := arrayField(cli.root, "commands")[0].(*orderedObject)
	restRaw, _ := ops.root.get("rest")
	if got := deriveCommandParameterFlags(cmd, restRaw.(*orderedObject)); got != 1 {
		t.Fatalf("derived flags = %d, want 1", got)
	}
	flag := arrayField(cmd, "flags")[0].(*orderedObject)
	if stringField(flag, "name") != "header-x-request-mode" || stringField(flag, "maps_to") != "header.X-Request-Mode" || stringField(flag, "type") != "enum" {
		t.Fatalf("header flag = %#v, want exact header mapping", flag.values)
	}
	if required, _ := flag.get("required"); required != true {
		t.Fatalf("header flag required = %#v, want true", required)
	}
}

func TestDerivedFlagTypeMapsEnumAndScalars(t *testing.T) {
	for _, tc := range []struct {
		name string
		json string
		want string
	}{
		{"enum wins over type", `{"type":"string","values":["a"]}`, "enum"},
		{"integer", `{"type":"integer"}`, "integer"},
		{"boolean", `{"type":"boolean"}`, "boolean"},
		{"array", `{"type":"array"}`, "string_array"},
		{"absent type defaults to string", `{}`, "string"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var doc orderedJSON
			if err := json.Unmarshal([]byte(tc.json), &doc); err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := derivedFlagType(doc.root); got != tc.want {
				t.Fatalf("derivedFlagType = %q, want %q", got, tc.want)
			}
		})
	}
}

// The exclusion has to match the claim: paging parameters are never imported.
// It cannot be limited to the names the connector's OWN pagination spec
// declares — github declares page/per_page and its specification still offers
// `after`/`before` cursors, which became 7 derived flags and a second, unchecked
// way to page. It also must not over-reach: the same `before` is an ISO 8601
// timestamp filter on /notifications, and dropping it would remove a real flag.
const paramsPagingFixtureOperations = `{
  "operations": [
    {
      "id": "acme.list_findings",
      "kind": "rest_read",
      "summary": "List findings",
      "risk": "low",
      "approval": "none",
      "output_policy": "json_redacted",
      "rest": { "method": "GET", "path": "/orgs/{org}/findings", "max_bytes": 1048576 }
    }
  ]
}`

const paramsPagingFixtureSpec = `{
  "type": "object",
  "properties": { "org": { "type": "string" }, "since": { "type": "string" }, "token": { "type": "string" } }
}`

const paramsPagingFixtureArtifact = `{
  "openapi": "3.0.3",
  "paths": {
    "/orgs/{org}/findings": {
      "get": {
        "parameters": [
          { "name": "org", "in": "path", "required": true, "schema": { "type": "string" } },
          { "name": "after", "in": "query", "description": "A cursor, as given in the Link header.", "schema": { "type": "string" } },
          { "name": "before", "in": "query", "description": "Only show results updated before the given time, in ISO 8601 format.", "schema": { "type": "string" } },
          { "name": "offset", "in": "query", "schema": { "type": "integer" } },
          { "name": "since", "in": "query", "description": "Only show results updated after the given time.", "schema": { "type": "string" } },
          { "name": "severity", "in": "query", "schema": { "type": "string" } }
        ]
      }
    }
  }
}`

func importedPagingFixtureNames(t *testing.T) []string {
	t.Helper()
	root := t.TempDir()
	defsDir := filepath.Join(root, "defs")
	writeParamsFixture(t, filepath.Join(defsDir, "acme"), map[string]string{
		"operations.json": paramsPagingFixtureOperations,
		"streams.json":    paramsFixtureStreams,
		"spec.json":       paramsPagingFixtureSpec,
	})
	artifact := filepath.Join(root, "artifact.json")
	if err := os.WriteFile(artifact, []byte(paramsPagingFixtureArtifact), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	if _, _, err := importConnectorParameters(paramsImportOptions{connector: "acme", artifact: artifact, defsDir: defsDir}); err != nil {
		t.Fatalf("importConnectorParameters: %v", err)
	}
	var names []string
	for _, param := range importedFixtureParameters(t, defsDir) {
		names = append(names, param["name"].(string))
	}
	return names
}

func TestParamsImportExcludesProviderPagingParametersByMeaning(t *testing.T) {
	names := strings.Join(importedPagingFixtureNames(t), ",")
	if names != "before,severity,since" {
		t.Fatalf("imported parameters = %q, want \"before,severity,since\"", names)
	}
}

// A config key is "already supplied by the connection" only where the
// operation's own path template interpolates it. github's `since` is read by
// the ETL incremental path alone and reaches no rest_read request, so skipping
// every config key removed a --since filter nothing else could supply.
func TestParamsImportSkipsOnlyPathVariablesTheConnectionSupplies(t *testing.T) {
	names := importedPagingFixtureNames(t)
	var sawOrg, sawSince bool
	for _, name := range names {
		switch name {
		case "org":
			sawOrg = true
		case "since":
			sawSince = true
		}
	}
	if sawOrg {
		t.Fatal("imported org, which the path template interpolates from config")
	}
	if !sawSince {
		t.Fatalf("imported parameters = %v, want since: it is a config key no request template consumes", names)
	}
}

func TestPathTemplateVariables(t *testing.T) {
	for _, tc := range []struct {
		path string
		want string
	}{
		{"/repos/{owner}/{repo}/notifications", "owner,repo"},
		{"/orgs/{org}/things", "org"},
		{"/rate_limit", ""},
		{"/broken/{unterminated", ""},
	} {
		if got := strings.Join(pathTemplateVariables(tc.path), ","); got != tc.want {
			t.Fatalf("pathTemplateVariables(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}
