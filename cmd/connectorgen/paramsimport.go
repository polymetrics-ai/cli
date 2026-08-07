package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// params-import carries a connector's accepted parameters from its own provider
// specification into operations.json, so command flags can be DERIVED from the
// provider's contract instead of hand-authored against it.
//
// It exists because the data was already in our hands and being thrown away:
// the batch materializer parses these same artifacts and skips their
// "parameters" field, which is why a command like `github issue list` exposed
// only --state while the endpoint documents sort, direction, since, labels,
// assignee, creator and milestone.
//
// Two things are deliberately NOT imported:
//
//   - paging parameters (page, per_page, cursor, ...). Those come from the
//     connector's declared pagination spec and are answered by --page /
//     --page-cursor. Importing them would let a caller bypass the paging
//     contract by setting the raw parameter, and would double-declare a value
//     the engine already controls.
//   - anything the connection already supplies. A path parameter named by the
//     connector's own config schema (github's owner/repo) resolves through
//     templating; turning it into a flag would make every command demand a
//     value the connection already knows.
//
// The import is a generator step, not a runtime one: it needs the artifact,
// while `surface-sync` derives flags from the imported result and stays
// hermetic so CI can verify drift without fetching anything.
const paramsImportUsage = `connectorgen params-import <connector> --artifact <path> [--defs <dir>] [--check]

Imports the accepted parameter set for a connector's rest_read operations from
its provider specification (OpenAPI 3 or Swagger 2) into operations.json.

  --artifact <path>  provider specification file (.json)
  --defs <dir>       connector defs root (default internal/connectors/defs)
  --check            report drift and exit non-zero instead of writing`

type paramsImportOptions struct {
	connector string
	artifact  string
	defsDir   string
	check     bool
}

func runParamsImport(args []string, stdout, stderr io.Writer) int {
	opts, err := parseParamsImportOptions(args[1:])
	if err != nil {
		logln(stderr, fmt.Sprintf("connectorgen params-import: %v", err))
		logln(stderr, paramsImportUsage)
		return 2
	}

	changed, total, err := importConnectorParameters(opts)
	if err != nil {
		logln(stderr, fmt.Sprintf("connectorgen params-import: %v", err))
		return 1
	}
	if opts.check && changed > 0 {
		logln(stderr, fmt.Sprintf("connectorgen params-import: %s has drifted; %d operation(s) differ from the artifact, run without --check to update", opts.connector, changed))
		return 1
	}
	logln(stdout, fmt.Sprintf("connectorgen params-import: %s, %d operation(s) scanned, %d updated", opts.connector, total, changed))
	return 0
}

func parseParamsImportOptions(args []string) (paramsImportOptions, error) {
	opts := paramsImportOptions{defsDir: filepath.Join("internal", "connectors", "defs")}
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "--check":
			opts.check = true
		case "--artifact", "--defs":
			if i+1 >= len(args) {
				return paramsImportOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			i++
			if arg == "--artifact" {
				opts.artifact = args[i]
			} else {
				opts.defsDir = args[i]
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return paramsImportOptions{}, fmt.Errorf("unknown flag %q", arg)
			}
			if opts.connector != "" {
				return paramsImportOptions{}, fmt.Errorf("only one connector may be imported at a time")
			}
			opts.connector = arg
		}
	}
	if opts.connector == "" {
		return paramsImportOptions{}, fmt.Errorf("a connector name is required")
	}
	if err := validateBatchConnectorName(opts.connector); err != nil {
		return paramsImportOptions{}, err
	}
	if opts.artifact == "" {
		return paramsImportOptions{}, fmt.Errorf("--artifact is required")
	}
	return opts, nil
}

// openAPIDoc is the bounded slice of an OpenAPI/Swagger document this importer
// reads. Everything else in the artifact is ignored on purpose: this step
// derives a parameter surface, not a full client.
type openAPIDoc struct {
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Parameters map[string]openAPIParameter `json:"parameters"`
	} `json:"components"`
	Parameters map[string]openAPIParameter `json:"parameters"`
}

type openAPIParameter struct {
	Ref         string `json:"$ref"`
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Schema      struct {
		Type string   `json:"type"`
		Enum []string `json:"enum"`
	} `json:"schema"`
	// Swagger 2 puts type/enum directly on the parameter rather than under a
	// schema object.
	Type string   `json:"type"`
	Enum []string `json:"enum"`
}

type openAPIOperation struct {
	Parameters []openAPIParameter `json:"parameters"`
}

func importConnectorParameters(opts paramsImportOptions) (changed, total int, err error) {
	raw, err := os.ReadFile(opts.artifact)
	if err != nil {
		return 0, 0, fmt.Errorf("read artifact: %w", err)
	}
	var doc openAPIDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return 0, 0, fmt.Errorf("parse artifact: %w", err)
	}

	bundleDir := filepath.Join(opts.defsDir, opts.connector)
	opsPath := filepath.Join(bundleDir, "operations.json")
	opsRaw, err := os.ReadFile(opsPath)
	if err != nil {
		return 0, 0, fmt.Errorf("read operations.json: %w", err)
	}
	var ops orderedJSON
	if err := json.Unmarshal(opsRaw, &ops); err != nil {
		return 0, 0, fmt.Errorf("parse operations.json: %w", err)
	}

	skip, err := paramsImportSkipSet(bundleDir)
	if err != nil {
		return 0, 0, err
	}

	for _, entry := range arrayField(ops.root, "operations") {
		op, ok := entry.(*orderedObject)
		if !ok {
			continue
		}
		if stringField(op, "kind") != "rest_read" {
			continue
		}
		restRaw, ok := op.get("rest")
		if !ok {
			continue
		}
		rest, ok := restRaw.(*orderedObject)
		if !ok {
			continue
		}
		total++
		method := strings.ToLower(strings.TrimSpace(stringField(rest, "method")))
		path := stringField(rest, "path")
		item, ok := doc.Paths[path]
		if !ok {
			continue
		}
		rawOp, ok := item[method]
		if !ok {
			continue
		}
		var parsed openAPIOperation
		if err := json.Unmarshal(rawOp, &parsed); err != nil {
			return 0, 0, fmt.Errorf("parse operation %s %s: %w", method, path, err)
		}
		want := importedParameters(doc, parsed.Parameters, skip)
		if !sameImportedParameters(rest, want) {
			changed++
			setImportedParameters(rest, want)
		}
	}

	if opts.check || changed == 0 {
		return changed, total, nil
	}
	if err := writeBundleJSON(opsPath, ops, opsRaw); err != nil {
		return 0, 0, err
	}
	return changed, total, nil
}

// paramsImportSkipSet collects every parameter name that must never become a
// flag for this connector: its declared paging parameters, and every path
// variable the connection already supplies through its config schema.
func paramsImportSkipSet(bundleDir string) (map[string]bool, error) {
	skip := map[string]bool{}

	streamsRaw, err := os.ReadFile(filepath.Join(bundleDir, "streams.json"))
	if err == nil {
		var streams struct {
			Base struct {
				Pagination map[string]any `json:"pagination"`
			} `json:"base"`
		}
		if err := json.Unmarshal(streamsRaw, &streams); err != nil {
			return nil, fmt.Errorf("parse streams.json: %w", err)
		}
		for _, key := range []string{"page_param", "size_param", "cursor_param", "limit_param", "offset_param", "count_param", "start_index_param"} {
			if value, ok := streams.Base.Pagination[key].(string); ok && strings.TrimSpace(value) != "" {
				skip[value] = true
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	specRaw, err := os.ReadFile(filepath.Join(bundleDir, "spec.json"))
	if err == nil {
		var spec struct {
			Properties map[string]json.RawMessage `json:"properties"`
		}
		if err := json.Unmarshal(specRaw, &spec); err != nil {
			return nil, fmt.Errorf("parse spec.json: %w", err)
		}
		for name := range spec.Properties {
			skip[name] = true
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return skip, nil
}

func resolveOpenAPIParameter(doc openAPIDoc, p openAPIParameter) (openAPIParameter, bool) {
	if p.Ref == "" {
		return p, true
	}
	name := p.Ref[strings.LastIndex(p.Ref, "/")+1:]
	if resolved, ok := doc.Components.Parameters[name]; ok {
		return resolved, true
	}
	if resolved, ok := doc.Parameters[name]; ok {
		return resolved, true
	}
	// An unresolvable $ref is dropped rather than guessed: a flag invented from
	// a reference we cannot read would not be derived from anything.
	return openAPIParameter{}, false
}

func importedParameters(doc openAPIDoc, params []openAPIParameter, skip map[string]bool) []map[string]any {
	out := make([]map[string]any, 0, len(params))
	seen := map[string]bool{}
	for _, raw := range params {
		p, ok := resolveOpenAPIParameter(doc, raw)
		if !ok {
			continue
		}
		name := strings.TrimSpace(p.Name)
		if name == "" || seen[name] || skip[name] {
			continue
		}
		if p.In != "query" && p.In != "path" {
			continue
		}
		seen[name] = true
		entry := map[string]any{"name": name, "in": p.In}
		if typ := firstNonEmpty(p.Schema.Type, p.Type); typ != "" {
			entry["type"] = typ
		}
		if p.Required {
			entry["required"] = true
		}
		if values := firstNonEmptySlice(p.Schema.Enum, p.Enum); len(values) > 0 {
			entry["values"] = values
		}
		if summary := strings.TrimSpace(firstLine(p.Description)); summary != "" {
			entry["summary"] = summary
		}
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i]["name"].(string) < out[j]["name"].(string)
	})
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonEmptySlice(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}

func firstLine(text string) string {
	if idx := strings.IndexAny(text, "\r\n"); idx >= 0 {
		return text[:idx]
	}
	return text
}

// sameImportedParameters reports whether the operation already carries exactly
// the parameter set the artifact declares, so --check can distinguish drift
// from a no-op and a rerun does not rewrite an unchanged file.
func sameImportedParameters(rest *orderedObject, want []map[string]any) bool {
	existing := arrayField(rest, "parameters")
	if len(existing) != len(want) {
		return false
	}
	for i, raw := range existing {
		got, ok := raw.(*orderedObject)
		if !ok {
			return false
		}
		if !sameImportedParameter(got, want[i]) {
			return false
		}
	}
	return true
}

func sameImportedParameter(got *orderedObject, want map[string]any) bool {
	if stringField(got, "name") != want["name"] || stringField(got, "in") != want["in"] {
		return false
	}
	wantType, _ := want["type"].(string)
	if stringField(got, "type") != wantType {
		return false
	}
	wantSummary, _ := want["summary"].(string)
	if stringField(got, "summary") != wantSummary {
		return false
	}
	gotRequired, _ := got.get("required")
	required, _ := gotRequired.(bool)
	wantRequired, _ := want["required"].(bool)
	if required != wantRequired {
		return false
	}
	wantValues, _ := want["values"].([]string)
	gotValues := arrayField(got, "values")
	if len(gotValues) != len(wantValues) {
		return false
	}
	for i, raw := range gotValues {
		value, _ := raw.(string)
		if value != wantValues[i] {
			return false
		}
	}
	return true
}

// setImportedParameters replaces the operation's parameter list wholesale: the
// artifact is the only source of truth, so a stale entry is drift rather than a
// local choice worth preserving.
func setImportedParameters(rest *orderedObject, want []map[string]any) {
	if len(want) == 0 {
		rest.remove("parameters")
		return
	}
	entries := make([]any, 0, len(want))
	for _, param := range want {
		entry := newOrderedObject()
		entry.set("name", param["name"])
		entry.set("in", param["in"])
		if typ, ok := param["type"].(string); ok && typ != "" {
			entry.set("type", typ)
		}
		if required, ok := param["required"].(bool); ok && required {
			entry.set("required", true)
		}
		if values, ok := param["values"].([]string); ok && len(values) > 0 {
			list := make([]any, 0, len(values))
			for _, value := range values {
				list = append(list, value)
			}
			entry.set("values", list)
		}
		if summary, ok := param["summary"].(string); ok && summary != "" {
			entry.set("summary", summary)
		}
		entries = append(entries, entry)
	}
	rest.set("parameters", entries)
}
