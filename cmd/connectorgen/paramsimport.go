package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
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
//   - paging parameters. Those come from the connector's declared pagination
//     spec and are answered by --page / --page-cursor. Importing them would let
//     a caller bypass the paging contract by setting the raw parameter, and
//     would double-declare a value the engine already controls. The exclusion
//     is by MEANING, not by the names one connector happens to declare: a
//     provider's `after`/`before` cursor is a paging parameter even when the
//     bundle's own spec calls its cursor `page`.
//   - a path variable the connection already supplies. github's owner/repo are
//     interpolated into the operation's own path template from config, so
//     turning them into flags would make every command demand a value the
//     connection already knows. It is deliberately NOT every config key:
//     github's `since` is read only by the ETL incremental path and reaches no
//     rest_read request, so skipping it removed a real --since filter.
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
		Type      string   `json:"type"`
		Enum      []string `json:"enum"`
		Pattern   string   `json:"pattern"`
		MinLength int      `json:"minLength"`
		MaxLength int      `json:"maxLength"`
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

	declaredPaging, configProps, err := paramsImportSkipSets(bundleDir)
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
		skip := map[string]bool{}
		for name := range declaredPaging {
			skip[name] = true
		}
		// Only a config key this operation's own path template interpolates is
		// already supplied by the connection. Every other config key is a key
		// some other code path reads, and skipping it drops a parameter the
		// caller has no other way to set.
		for _, name := range pathTemplateVariables(path) {
			if configProps[name] {
				skip[name] = true
			}
		}
		want := importedParameters(doc, parsed.Parameters, skip)
		if !sameImportedParameters(rest, want) {
			changed++
			setImportedParameters(rest, want)
		}
		if paginationRaw, ok := rest.get("pagination"); ok {
			if pagination, ok := paginationRaw.(*orderedObject); ok {
				wantPaging := importedOperationPaginationParameters(doc, parsed.Parameters, operationPaginationParameterNames(pagination))
				if !sameImportedField(rest, "pagination_parameters", wantPaging) {
					changed++
					setImportedField(rest, "pagination_parameters", wantPaging)
				}
			}
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

// paramsImportSkipSets reads the two bundle-declared name sets the import
// consults: the paging parameters this connector's own pagination spec names,
// and its config schema's property names.
//
// They are returned apart because they are applied differently. A declared
// paging parameter is excluded everywhere; a config property is excluded only
// where the operation's own path template actually interpolates it.
func paramsImportSkipSets(bundleDir string) (paging, configProps map[string]bool, err error) {
	paging = map[string]bool{}
	configProps = map[string]bool{}

	streamsRaw, err := os.ReadFile(filepath.Join(bundleDir, "streams.json"))
	if err == nil {
		var streams struct {
			Base struct {
				Pagination map[string]any `json:"pagination"`
			} `json:"base"`
		}
		if err := json.Unmarshal(streamsRaw, &streams); err != nil {
			return nil, nil, fmt.Errorf("parse streams.json: %w", err)
		}
		for _, key := range []string{"page_param", "size_param", "cursor_param", "limit_param", "offset_param", "count_param", "start_index_param"} {
			if value, ok := streams.Base.Pagination[key].(string); ok && strings.TrimSpace(value) != "" {
				paging[value] = true
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, nil, err
	}

	specRaw, err := os.ReadFile(filepath.Join(bundleDir, "spec.json"))
	if err == nil {
		var spec struct {
			Properties map[string]json.RawMessage `json:"properties"`
		}
		if err := json.Unmarshal(specRaw, &spec); err != nil {
			return nil, nil, fmt.Errorf("parse spec.json: %w", err)
		}
		for name := range spec.Properties {
			configProps[name] = true
		}
	} else if !os.IsNotExist(err) {
		return nil, nil, err
	}
	return paging, configProps, nil
}

// pathTemplateVariables lists the {name} placeholders an operation's REST path
// interpolates — the only config keys the connection genuinely supplies to that
// request.
func pathTemplateVariables(path string) []string {
	var out []string
	for {
		open := strings.IndexByte(path, '{')
		if open < 0 {
			return out
		}
		path = path[open+1:]
		close := strings.IndexByte(path, '}')
		if close < 0 {
			return out
		}
		if name := strings.TrimSpace(path[:close]); name != "" {
			out = append(out, name)
		}
		path = path[close+1:]
	}
}

// providerPagingParameterNames are the parameter names that are paging
// mechanics in any specification that uses them, whatever the connector's own
// pagination spec happens to call its cursor. A connector declaring
// `cursor_param: "page"` does not make a provider's `offset` a filter.
//
// Names that are only SOMETIMES paging — `count`, `size`, `per`, `start`,
// `after`, `before` — are deliberately absent: github's `per` is a time frame
// and its `before` on /notifications is an ISO 8601 timestamp filter. Those are
// classified from the specification's own description instead, which is what
// distinguishes that `before` from the `before` whose description opens "A
// cursor, as given in the Link header".
var providerPagingParameterNames = map[string]bool{
	"page": true, "perpage": true, "pagesize": true, "pagenumber": true,
	"pagenum": true, "pageindex": true, "pagelimit": true, "pageoffset": true,
	"pagetoken": true, "pagecursor": true, "nextpagetoken": true,
	"cursor": true, "nextcursor": true, "nexttoken": true, "nextpage": true,
	"continuationtoken": true, "scrollid": true,
	"offset": true, "startindex": true, "startcursor": true,
	"startingafter": true, "startingbefore": true, "endingbefore": true,
	"limit": true, "maxresults": true, "resultsperpage": true,
}

// isProviderPagingParameter reports whether the specification is describing a
// paging mechanism, by name or by its own words.
//
// The contract this enforces is that there is exactly ONE way to page — the
// connector's declared pagination spec, reached through --page/--page-cursor.
// A derived flag for a provider cursor would be a second, unchecked channel
// that bypasses the completeness contract entirely.
func isProviderPagingParameter(p openAPIParameter) bool {
	if p.In != "query" {
		return false
	}
	normalized := strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(p.Name)))
	if providerPagingParameterNames[normalized] {
		return true
	}
	description := strings.ToLower(p.Description)
	for _, marker := range []string{"cursor", "used for pagination", "pagination token", "page token"} {
		if strings.Contains(description, marker) {
			return true
		}
	}
	return false
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
		parameterKey := p.In + "\x00" + name
		if name == "" || seen[parameterKey] || skip[name] {
			continue
		}
		if p.In != "query" && p.In != "path" && p.In != "header" {
			continue
		}
		if isProviderPagingParameter(p) {
			continue
		}
		seen[parameterKey] = true
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
		if p.In == "header" {
			schema, maxBytes, ok := importedHeaderSchema(p)
			if !ok {
				// A header with no bounded string contract is not a safe CLI
				// input. Leave it unavailable rather than inventing a generic
				// header path; a source contract can be improved and re-imported.
				continue
			}
			entry["type"] = "string"
			entry["schema"] = schema
			entry["max_bytes"] = maxBytes
		}
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i]["name"].(string) < out[j]["name"].(string)
	})
	return out
}

// importedHeaderSchema turns the bounded string subset of an OpenAPI header
// parameter into the engine's schema dialect. The byte cap is conservative for
// a provider's character maxLength (UTF-8 can use four bytes per code point),
// or exact for a closed enum. Unbounded or non-string headers are intentionally
// not imported: generated commands must never create a generic header escape
// hatch.
func importedHeaderSchema(p openAPIParameter) (map[string]any, int, bool) {
	typ := firstNonEmpty(p.Schema.Type, p.Type)
	if typ != "string" {
		return nil, 0, false
	}
	values := firstNonEmptySlice(p.Schema.Enum, p.Enum)
	schema := map[string]any{"type": "string"}
	if len(values) > 0 {
		schema["enum"] = values
	}
	if p.Schema.Pattern != "" {
		schema["pattern"] = p.Schema.Pattern
	}
	if p.Schema.MinLength > 0 {
		schema["minLength"] = p.Schema.MinLength
	}
	if p.Schema.MaxLength > 0 {
		schema["maxLength"] = p.Schema.MaxLength
	}
	maxBytes := 0
	if p.Schema.MaxLength > 0 && p.Schema.MaxLength <= 4096 {
		maxBytes = p.Schema.MaxLength * 4
	}
	for _, value := range values {
		if len(value) > maxBytes {
			maxBytes = len(value)
		}
	}
	if maxBytes <= 0 || maxBytes > 16<<10 {
		return nil, 0, false
	}
	return schema, maxBytes, true
}

// importedOperationPaginationParameters records the exact source parameters a
// per-operation pager consumes, but never exposes them as CLI flags. The
// declaration remains authoritative for the pagination shape; the artifact
// proves its query names are actually documented by this endpoint.
func importedOperationPaginationParameters(doc openAPIDoc, params []openAPIParameter, names map[string]bool) []map[string]any {
	out := make([]map[string]any, 0, len(names))
	for _, raw := range params {
		p, ok := resolveOpenAPIParameter(doc, raw)
		if !ok || p.In != "query" || !names[p.Name] {
			continue
		}
		entry := map[string]any{"name": p.Name, "in": p.In}
		if typ := firstNonEmpty(p.Schema.Type, p.Type); typ != "" {
			entry["type"] = typ
		}
		if p.Required {
			entry["required"] = true
		}
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i]["name"].(string) < out[j]["name"].(string) })
	return out
}

func operationPaginationParameterNames(pagination *orderedObject) map[string]bool {
	names := map[string]bool{}
	add := func(name string) {
		if strings.TrimSpace(name) != "" {
			names[name] = true
		}
	}
	switch stringField(pagination, "type") {
	case "page_number":
		add(stringField(pagination, "page_param"))
		add(stringField(pagination, "size_param"))
	case "offset_limit":
		add(stringField(pagination, "offset_param"))
		add(stringField(pagination, "limit_param"))
		add(stringField(pagination, "size_param"))
	case "cursor":
		add(stringField(pagination, "cursor_param"))
		add(stringField(pagination, "size_param"))
	case "start_index":
		start := stringField(pagination, "start_index_param")
		if start == "" {
			start = "startIndex"
		}
		add(start)
		count := stringField(pagination, "count_param")
		if count == "" {
			count = "count"
		}
		add(count)
		add(stringField(pagination, "size_param"))
	}
	return names
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
	return sameImportedField(rest, "parameters", want)
}

func sameImportedField(rest *orderedObject, field string, want []map[string]any) bool {
	existing := arrayField(rest, field)
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
	wantSchema, schemaDeclared := want["schema"]
	gotSchema, gotSchemaDeclared := got.get("schema")
	if schemaDeclared != gotSchemaDeclared || (schemaDeclared && !sameImportedJSONValue(gotSchema, wantSchema)) {
		return false
	}
	wantMaxBytes, maxBytesDeclared := want["max_bytes"].(int)
	gotMaxBytes, gotMaxBytesDeclared := got.get("max_bytes")
	if maxBytesDeclared != gotMaxBytesDeclared {
		return false
	}
	if maxBytesDeclared {
		gotNumber, ok := gotMaxBytes.(json.Number)
		if !ok || gotNumber.String() != strconv.Itoa(wantMaxBytes) {
			return false
		}
	}
	return true
}

// sameImportedJSONValue compares the ordered document representation with the
// map/slice values constructed from the OpenAPI artifact. Re-marshalling keeps
// the comparison semantic rather than treating orderedObject and map as two
// different schemas.
func sameImportedJSONValue(got, want any) bool {
	gotRaw, gotErr := json.Marshal(got)
	wantRaw, wantErr := json.Marshal(want)
	if gotErr != nil || wantErr != nil {
		return false
	}
	var gotNormalized, wantNormalized any
	if json.Unmarshal(gotRaw, &gotNormalized) != nil || json.Unmarshal(wantRaw, &wantNormalized) != nil {
		return false
	}
	return reflect.DeepEqual(gotNormalized, wantNormalized)
}

// setImportedParameters replaces the operation's parameter list wholesale: the
// artifact is the only source of truth, so a stale entry is drift rather than a
// local choice worth preserving.
func setImportedParameters(rest *orderedObject, want []map[string]any) {
	setImportedField(rest, "parameters", want)
}

func setImportedField(rest *orderedObject, field string, want []map[string]any) {
	if len(want) == 0 {
		rest.remove(field)
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
		if schema, ok := param["schema"].(map[string]any); ok {
			entry.set("schema", schema)
		}
		if maxBytes, ok := param["max_bytes"].(int); ok && maxBytes > 0 {
			entry.set("max_bytes", json.Number(strconv.Itoa(maxBytes)))
		}
		entries = append(entries, entry)
	}
	rest.set(field, entries)
}
