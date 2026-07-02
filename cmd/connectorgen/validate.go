package main

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

// Rule identifiers named by every validate Finding. Kept as exported-looking
// constants (lowercase, package-private) so tests can assert on them without
// string literals scattered across the corpus.
const (
	ruleMissingFile             = "missing_file"
	ruleMetaSchema              = "meta_schema"
	ruleInterpolationUnresolved = "interpolation_unresolved"
	ruleSchemaRefMissing        = "schema_ref_missing"
	rulePrimaryKeyMissing       = "primary_key_missing"
	ruleCursorFieldMissing      = "cursor_field_missing"
	ruleWritePathFields         = "write_path_fields"
	ruleSurfaceCoverage         = "surface_coverage"
	ruleSurfaceUnknownTarget    = "surface_unknown_target"
	ruleSurfaceIncomplete       = "surface_incomplete"
	ruleSurfaceCategory         = "surface_category"
	ruleSurfaceFailFirstRun     = "surface_fail_first_run"
	ruleNameRegex               = "name_regex"
	ruleSecretLiteral           = "secret_literal"
	ruleDocsHeading             = "docs_heading"
)

// surfaceCategories is the closed exclusion vocabulary (design §E.1 rule 3).
// The engine loader's meta-schema already enforces this via an enum on
// api_surface.schema.json, so an unknown category surfaces as a
// ruleMetaSchema finding at load time; this set is kept here too for
// defense-in-depth documentation of the rule.
var surfaceCategories = map[string]bool{
	"destructive_admin":       true,
	"requires_elevated_scope": true,
	"binary_payload":          true,
	"deprecated":              true,
	"non_data_endpoint":       true,
	"duplicate_of":            true,
	"out_of_scope":            true,
}

// mutationMethods are the HTTP verbs api_surface rule 4 treats as write
// endpoints for the fail-first-run capabilities.write check.
var mutationMethods = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"PATCH":  true,
	"DELETE": true,
}

// requiredDocHeadings are the fixed docs.md headings (design §F.6 /
// DATA-MODEL §1).
var requiredDocHeadings = []string{
	"Overview",
	"Auth setup",
	"Streams notes",
	"Write actions & risks",
	"Known limits",
}

// secretLiteralPattern flags secret-shaped literals accidentally committed to
// fixtures: a Bearer-scheme header value, a long opaque token following an
// auth-flavored key (api_key/access_token/secret/password), or a
// recognizable vendor secret prefix (e.g. Stripe's sk_live_/sk_test_).
// Fixtures must only ever carry synthetic data (THREAT-MODEL §4).
var secretLiteralPattern = regexp.MustCompile(`(?i)(bearer\s+[a-z0-9_\-\.]{16,}|(api[_-]?key|access[_-]?token|secret|password)["' ]*[:=]\s*["']?[a-z0-9_\-\.]{16,}|\bsk_(live|test)_[a-z0-9]{10,}\b)`)

// Finding is one validate defect: which connector/file/rule it belongs to and
// a human-readable message.
type Finding struct {
	Connector string `json:"connector"`
	File      string `json:"file"`
	Rule      string `json:"rule"`
	Message   string `json:"message"`
}

// Report is the aggregate result of validating every bundle in a directory
// tree; it is what both the text and --json output modes render from.
type Report struct {
	Findings          []Finding `json:"findings"`
	ConnectorsChecked int       `json:"connectors_checked"`
}

// validateDir loads and validates every bundle directory at the root of
// fsys, composing the engine loader (structural + meta-schema validation),
// engine.ResolveCheck (template resolution), and the connectorgen-owned
// semantic rules (PK/cursor existence, write path_fields, api_surface rules
// 1-4, naming, docs headings, fixture secret scan).
//
// An empty tree (no bundle directories) is not an error: it returns a Report
// with ConnectorsChecked == 0 and no findings, matching engine.LoadAll's own
// tolerance for defs/ shipping zero bundles before Wave F.
func validateDir(fsys fs.FS) (Report, error) {
	names, err := bundleDirNames(fsys)
	if err != nil {
		return Report{}, err
	}

	// Always non-nil so JSON output renders "findings": [] rather than null
	// on a clean run (the --json contract promises an array).
	findings := []Finding{}
	for _, name := range names {
		findings = append(findings, validateBundleDir(fsys, name)...)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Connector != findings[j].Connector {
			return findings[i].Connector < findings[j].Connector
		}
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		return findings[i].Rule < findings[j].Rule
	})

	return Report{Findings: findings, ConnectorsChecked: len(names)}, nil
}

// bundleDirNames returns the sorted top-level directory names of fsys, the
// same candidate set engine.LoadAll iterates.
func bundleDirNames(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("validate: read root: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// validateBundleDir validates a single candidate bundle directory. It never
// returns a bare error: any structural/loader failure is translated into a
// Finding so a single malformed bundle does not abort validation of its
// siblings (and so `--json` always has a machine-readable shape to render).
func validateBundleDir(fsys fs.FS, name string) []Finding {
	b, err := engine.Load(fsys, name)
	if err != nil {
		return []Finding{loadErrorFinding(name, err)}
	}

	var findings []Finding
	findings = append(findings, checkName(b)...)
	findings = append(findings, checkInterpolations(b)...)
	findings = append(findings, checkSchemaRefs(fsys, b)...)
	findings = append(findings, checkPrimaryKeysAndCursors(b)...)
	findings = append(findings, checkWritePathFields(b)...)
	findings = append(findings, checkAPISurface(b)...)
	findings = append(findings, checkDocsHeadings(b)...)
	findings = append(findings, checkFixtureSecrets(b)...)
	return findings
}

// loadErrorFinding classifies an engine.Load error into the most specific
// rule its message identifies, defaulting to ruleMetaSchema (loader errors
// not otherwise classified are, in practice, meta-schema/compile failures).
func loadErrorFinding(name string, err error) Finding {
	msg := err.Error()
	rule := ruleMetaSchema
	file := "metadata.json"
	switch {
	case strings.Contains(msg, "missing required file"):
		rule = ruleMissingFile
		file = missingFileFromError(msg)
	case strings.Contains(msg, "does not match") && strings.Contains(msg, namePatternDescription):
		rule = ruleNameRegex
		file = "metadata.json"
	case strings.Contains(msg, "directory name") && strings.Contains(msg, "does not match"):
		rule = ruleNameRegex
		file = "metadata.json"
	case strings.Contains(msg, ": schema ") && strings.Contains(msg, "no such file"):
		// loadStreamSchemas' error shape: "...: stream X: schema Y: read Y: ...".
		rule = ruleSchemaRefMissing
		file = "streams.json"
	case strings.Contains(msg, "spec.json"):
		file = "spec.json"
	case strings.Contains(msg, "streams.json"):
		file = "streams.json"
	case strings.Contains(msg, "writes.json"):
		file = "writes.json"
	case strings.Contains(msg, "api_surface.json"):
		file = "api_surface.json"
	}
	return Finding{Connector: name, File: file, Rule: rule, Message: msg}
}

// missingFileFromError extracts the file name named in a "missing required
// file X" loader error message; falls back to metadata.json if it cannot.
func missingFileFromError(msg string) string {
	const marker = "missing required file "
	idx := strings.Index(msg, marker)
	if idx < 0 {
		return "metadata.json"
	}
	rest := strings.TrimSpace(msg[idx+len(marker):])
	// The message may continue with " (required unless ...)"; keep only the
	// leading filename token.
	if sp := strings.IndexAny(rest, " \t"); sp >= 0 {
		rest = rest[:sp]
	}
	return rest
}

const namePatternDescription = "^[a-z0-9][a-z0-9-]*$"

func checkName(b engine.Bundle) []Finding {
	if !namePattern.MatchString(b.Name) {
		return []Finding{{
			Connector: b.Name, File: "metadata.json", Rule: ruleNameRegex,
			Message: fmt.Sprintf("connector name %q does not match %s", b.Name, namePatternDescription),
		}}
	}
	return nil
}

// namePattern mirrors engine's own (unexported) naming rule; connectorgen
// re-validates it defensively even though engine.Load already enforces it,
// since a future loader relaxation should not silently widen what
// `connectorgen validate` accepts.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// checkInterpolations resolves every {{ }} template found in the bundle's
// streams.json (base URL/headers/query/path/pagination knobs are strings
// there) and writes.json (path templates) against spec.json's declared
// property set, using engine.ResolveCheck.
func checkInterpolations(b engine.Bundle) []Finding {
	specKeys := map[string]bool{}
	if b.Spec != nil {
		for _, k := range b.Spec.Properties() {
			specKeys[k] = true
		}
	}

	var findings []Finding
	check := func(file, template string) {
		if template == "" {
			return
		}
		if err := interpolationResolveCheck(template, specKeys); err != nil {
			findings = append(findings, Finding{
				Connector: b.Name, File: file, Rule: ruleInterpolationUnresolved,
				Message: err.Error(),
			})
		}
	}

	check("streams.json", b.HTTP.URL)
	for _, h := range b.HTTP.Headers {
		check("streams.json", h)
	}
	for _, a := range b.HTTP.Auth {
		// engine.ResolveCheckAuthSpec validates EVERY templated AuthSpec
		// field (token/username/password/header/value/token_url/client_id/
		// client_secret/scopes/when, not just token/value/when) against
		// specKeys (F9, REVIEW.md — R1 added this engine-side helper;
		// wiring it here closes the gap connectorgen validate previously
		// left: a typo'd username/password/token_url/client_id/
		// client_secret/scopes template passed validate and only failed at
		// runtime).
		if err := engine.ResolveCheckAuthSpec(a, specKeys); err != nil {
			findings = append(findings, Finding{
				Connector: b.Name, File: "streams.json", Rule: ruleInterpolationUnresolved,
				Message: err.Error(),
			})
		}
	}
	for _, s := range b.Streams {
		check("streams.json", s.Path)
		for _, v := range s.Query {
			check("streams.json", v)
		}
		for _, v := range s.ComputedFields {
			check("streams.json", v)
		}
	}
	for _, w := range b.Writes {
		check("writes.json", w.Path)
	}
	return findings
}

// interpolationResolveCheck delegates to engine.ResolveCheck; kept as its own
// indirection point in case connectorgen ever needs to special-case a
// namespace beyond what engine checks statically.
func interpolationResolveCheck(template string, specKeys map[string]bool) error {
	return engine.ResolveCheck(template, specKeys)
}

// checkSchemaRefs verifies every stream's declared schema file exists. In
// practice engine.Load already fails the whole bundle (surfaced as a
// ruleMetaSchema finding above) when a schema ref is missing, since
// loadStreamSchemas errors out during Load. This function exists so the
// finding is named with the specific ruleSchemaRefMissing rule when we can
// still enumerate the stream (kept independent/defensive: if the loader ever
// becomes lenient about missing schema files, this still catches it).
func checkSchemaRefs(fsys fs.FS, b engine.Bundle) []Finding {
	sub, err := fs.Sub(fsys, b.Name)
	if err != nil {
		return nil
	}
	var findings []Finding
	for _, s := range b.Streams {
		if s.SchemaRef == "" {
			continue
		}
		if _, err := fs.Stat(sub, s.SchemaRef); err != nil {
			findings = append(findings, Finding{
				Connector: b.Name, File: "streams.json", Rule: ruleSchemaRefMissing,
				Message: fmt.Sprintf("stream %q schema ref %q does not exist", s.Name, s.SchemaRef),
			})
		}
	}
	return findings
}

// checkPrimaryKeysAndCursors enforces that every x-primary-key field and
// every incremental.cursor_field named by a stream actually exists among
// that stream's compiled schema properties.
func checkPrimaryKeysAndCursors(b engine.Bundle) []Finding {
	var findings []Finding
	for _, s := range b.Streams {
		sch, ok := b.Schemas[s.Name]
		if !ok {
			continue
		}
		props := map[string]bool{}
		for _, p := range sch.Properties() {
			props[p] = true
		}
		for _, pk := range sch.PrimaryKey {
			if !props[pk] {
				findings = append(findings, Finding{
					Connector: b.Name, File: s.SchemaRef, Rule: rulePrimaryKeyMissing,
					Message: fmt.Sprintf("stream %q x-primary-key field %q not found in schema properties", s.Name, pk),
				})
			}
		}
		if s.Incremental != nil && s.Incremental.CursorField != "" {
			if !props[s.Incremental.CursorField] {
				findings = append(findings, Finding{
					Connector: b.Name, File: "streams.json", Rule: ruleCursorFieldMissing,
					Message: fmt.Sprintf("stream %q incremental.cursor_field %q not found in schema %q properties", s.Name, s.Incremental.CursorField, s.SchemaRef),
				})
			}
		}
	}
	return findings
}

// checkWritePathFields enforces path_fields ⊆ record_schema properties for
// every write action.
func checkWritePathFields(b engine.Bundle) []Finding {
	var findings []Finding
	for _, w := range b.Writes {
		if len(w.PathFields) == 0 {
			continue
		}
		sch, err := engine.CompileSchema(w.RecordSchema)
		if err != nil {
			// Malformed record_schema is reported via the loader's own
			// meta-schema/compile path; skip here to avoid a duplicate,
			// less-specific finding.
			continue
		}
		props := map[string]bool{}
		for _, p := range sch.Properties() {
			props[p] = true
		}
		for _, pf := range w.PathFields {
			if !props[pf] {
				findings = append(findings, Finding{
					Connector: b.Name, File: "writes.json", Rule: ruleWritePathFields,
					Message: fmt.Sprintf("write action %q path_field %q not found in record_schema properties", w.Name, pf),
				})
			}
		}
	}
	return findings
}

// checkAPISurface enforces design §E.1 rules 1-4:
//  1. every endpoint has exactly one of covered_by/excluded.
//  2. covered_by.stream/covered_by.write resolves to a declared stream/action,
//     and every declared stream/action appears in the surface.
//  3. excluded.category is from the closed vocabulary (defense-in-depth; the
//     loader's meta-schema enum already enforces this at load time).
//  4. capabilities.write/read == false is only legal when the surface has
//     zero non-excluded mutation/GET endpoints respectively.
func checkAPISurface(b engine.Bundle) []Finding {
	if b.Surface == nil {
		return nil
	}
	var findings []Finding

	streams := map[string]bool{}
	for _, s := range b.Streams {
		streams[s.Name] = true
	}
	writes := map[string]bool{}
	for _, w := range b.Writes {
		writes[w.Name] = true
	}

	coveredStreams := map[string]bool{}
	coveredWrites := map[string]bool{}
	hasNonExcludedGET := false
	hasNonExcludedMutation := false

	for i, ep := range b.Surface.Endpoints {
		hasCovered := ep.CoveredBy != nil && (ep.CoveredBy.Stream != "" || ep.CoveredBy.Write != "")
		hasExcluded := ep.Excluded != nil

		switch {
		case hasCovered && hasExcluded:
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
				Message: fmt.Sprintf("endpoint %d (%s %s) has both covered_by and excluded", i, ep.Method, ep.Path),
			})
		case !hasCovered && !hasExcluded:
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
				Message: fmt.Sprintf("endpoint %d (%s %s) has neither covered_by nor excluded", i, ep.Method, ep.Path),
			})
		case hasCovered:
			if ep.CoveredBy.Stream != "" {
				if !streams[ep.CoveredBy.Stream] {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceUnknownTarget,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.stream %q is not a declared stream", i, ep.Method, ep.Path, ep.CoveredBy.Stream),
					})
				} else {
					coveredStreams[ep.CoveredBy.Stream] = true
				}
			}
			if ep.CoveredBy.Write != "" {
				if !writes[ep.CoveredBy.Write] {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceUnknownTarget,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.write %q is not a declared write action", i, ep.Method, ep.Path, ep.CoveredBy.Write),
					})
				} else {
					coveredWrites[ep.CoveredBy.Write] = true
				}
			}
			if strings.EqualFold(ep.Method, "GET") {
				hasNonExcludedGET = true
			}
			if mutationMethods[strings.ToUpper(ep.Method)] {
				hasNonExcludedMutation = true
			}
		case hasExcluded:
			if !surfaceCategories[ep.Excluded.Category] {
				findings = append(findings, Finding{
					Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCategory,
					Message: fmt.Sprintf("endpoint %d (%s %s) excluded.category %q is not in the closed vocabulary", i, ep.Method, ep.Path, ep.Excluded.Category),
				})
			}
		}
	}

	for name := range streams {
		if !coveredStreams[name] {
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceIncomplete,
				Message: fmt.Sprintf("stream %q has no covered_by entry in api_surface.json", name),
			})
		}
	}
	for name := range writes {
		if !coveredWrites[name] {
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceIncomplete,
				Message: fmt.Sprintf("write action %q has no covered_by entry in api_surface.json", name),
			})
		}
	}

	if !b.Metadata.Capabilities.Write && hasNonExcludedMutation {
		findings = append(findings, Finding{
			Connector: b.Name, File: "metadata.json", Rule: ruleSurfaceFailFirstRun,
			Message: "capabilities.write is false but api_surface.json has a non-excluded POST/PUT/PATCH/DELETE endpoint",
		})
	}
	if !b.Metadata.Capabilities.Read && hasNonExcludedGET {
		findings = append(findings, Finding{
			Connector: b.Name, File: "metadata.json", Rule: ruleSurfaceFailFirstRun,
			Message: "capabilities.read is false but api_surface.json has a non-excluded GET endpoint",
		})
	}

	return findings
}

// checkDocsHeadings enforces the fixed docs.md heading set (design §F.6).
// Headings are matched as Markdown "# "/"## " lines by exact (trimmed) text,
// so heading LEVEL is not enforced, only presence and text.
func checkDocsHeadings(b engine.Bundle) []Finding {
	present := map[string]bool{}
	for _, line := range strings.Split(b.Docs, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimLeft(trimmed, "#")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed != "" {
			present[trimmed] = true
		}
	}
	var findings []Finding
	for _, h := range requiredDocHeadings {
		if !present[h] {
			findings = append(findings, Finding{
				Connector: b.Name, File: "docs.md", Rule: ruleDocsHeading,
				Message: fmt.Sprintf("docs.md missing required heading %q", h),
			})
		}
	}
	return findings
}

// checkFixtureSecrets scans every fixture file's raw bytes for
// secret-shaped literals. Fixtures must only ever contain synthetic data
// (THREAT-MODEL §4); a planted real-looking token is a hard validate
// failure, not a warning.
func checkFixtureSecrets(b engine.Bundle) []Finding {
	if b.Fixtures == nil {
		return nil
	}
	var findings []Finding
	_ = fs.WalkDir(b.Fixtures, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		raw, ferr := fs.ReadFile(b.Fixtures, path)
		if ferr != nil {
			return nil
		}
		if secretLiteralPattern.Match(raw) {
			findings = append(findings, Finding{
				Connector: b.Name, File: "fixtures/" + path, Rule: ruleSecretLiteral,
				Message: fmt.Sprintf("fixtures/%s contains a secret-shaped literal", path),
			})
		}
		return nil
	})
	return findings
}
