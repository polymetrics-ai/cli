package main

import (
	"encoding/json"
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
	ruleMissingFile              = "missing_file"
	ruleMetaSchema               = "meta_schema"
	ruleInterpolationUnresolved  = "interpolation_unresolved"
	ruleSchemaRefMissing         = "schema_ref_missing"
	rulePrimaryKeyMissing        = "primary_key_missing"
	ruleCursorFieldMissing       = "cursor_field_missing"
	ruleWritePathFields          = "write_path_fields"
	ruleSurfaceCoverage          = "surface_coverage"
	ruleSurfaceUnknownTarget     = "surface_unknown_target"
	ruleSurfaceIncomplete        = "surface_incomplete"
	ruleSurfaceCategory          = "surface_category"
	ruleSurfaceOperation         = "surface_operation"
	ruleSurfaceFailFirstRun      = "surface_fail_first_run"
	ruleCLISurfaceUnknownTarget  = "cli_surface_unknown_target"
	ruleCLISurfaceMissingMapping = "cli_surface_missing_mapping"
	ruleCLISurfaceSafety         = "cli_surface_safety"
	ruleNameRegex                = "name_regex"
	ruleSecretLiteral            = "secret_literal"
	ruleDocsHeading              = "docs_heading"
	ruleStartDateFreeFormString  = "start_date_free_form_string"
	ruleConformanceSkipReason    = "conformance_skip_reason"
	ruleDefaultTypeMismatch      = "default_type_mismatch"
)

// dateShapedParamFormats are the incremental.param_format values whose
// value-parsing path (engine/read.go parseLowerBoundTime, N4/B1) accepts an
// all-digits input as Unix seconds and otherwise requires RFC3339. For these
// two formats specifically (unlike unix_seconds, where digits ARE the
// correct/intended shape), a digit-shaped config value that is NOT actually
// Unix seconds — e.g. a yyyymmdd typo like "20260101" — is silently
// misinterpreted as a 1970s-era lower bound rather than erroring (N2, wave0
// REVIEW.md carried flag).
var dateShapedParamFormats = map[string]bool{
	"date":              true,
	"github_date_range": true,
}

// dateShapedSpecFormats are the JSON Schema "format" annotation values that
// make a start_config_key spec property's shape explicit enough that this
// warning does not apply: an operator filling in a field the spec itself
// declares as a timestamp is not the free-form-string risk N2 describes.
var dateShapedSpecFormats = map[string]bool{
	"date-time": true,
	"date":      true,
}

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

var surfaceOperationModels = map[string]bool{
	"direct_read":           true,
	"binary_read":           true,
	"sensitive_reverse_etl": true,
	"admin_reverse_etl":     true,
	"destructive_action":    true,
	"local_workflow":        true,
	"duplicate":             true,
	"deprecated":            true,
	"disallowed":            true,
}

var surfaceOperationStatuses = map[string]bool{
	"blocked": true,
}

var surfaceOperationRisks = map[string]bool{
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

var directReadOutputPolicies = map[string]bool{
	"github_contents_file_metadata": true,
	"github_contents_directory":     true,
}

var sourceRequiredOperationModels = map[string]bool{
	"sensitive_reverse_etl": true,
	"admin_reverse_etl":     true,
	"destructive_action":    true,
	"disallowed":            true,
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
var secretLiteralPattern = regexp.MustCompile(`(?i)(bearer\s+[a-z0-9_\-\.]{16,}|(api[_-]?key|access[_-]?token|secret|password)["' ]*[:=]\s*["']?[a-z0-9_\-\.]{16,}|\bsk_(live|test)_[a-z0-9]{10,}\b|\bgh[pousr]_[a-z0-9_]{20,}\b|\bgithub_pat_[a-z0-9_]{20,}\b)`)

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
//
// Warnings (N2, wave0 REVIEW.md carried flag) are a SEPARATE, lower-severity
// list, deliberately never merged into Findings: they never affect
// validate's exit code or the "0 findings" self-verify contract goldens
// rely on (`go run ./cmd/connectorgen validate internal/connectors/defs`).
// A warning names a plausibility risk a bundle author should look at, not a
// structural defect connectorgen can prove is wrong.
type Report struct {
	Findings          []Finding `json:"findings"`
	Warnings          []Finding `json:"warnings"`
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

	// Always non-nil so JSON output renders "findings": [] / "warnings": []
	// rather than null on a clean run (the --json contract promises arrays).
	findings := []Finding{}
	warnings := []Finding{}
	for _, name := range names {
		bundleFindings, bundleWarnings := validateBundleDir(fsys, name)
		findings = append(findings, bundleFindings...)
		warnings = append(warnings, bundleWarnings...)
	}
	sortFindings := func(list []Finding) {
		sort.Slice(list, func(i, j int) bool {
			if list[i].Connector != list[j].Connector {
				return list[i].Connector < list[j].Connector
			}
			if list[i].File != list[j].File {
				return list[i].File < list[j].File
			}
			return list[i].Rule < list[j].Rule
		})
	}
	sortFindings(findings)
	sortFindings(warnings)

	return Report{Findings: findings, Warnings: warnings, ConnectorsChecked: len(names)}, nil
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

// validateBundleDir validates a single candidate bundle directory, returning
// (hard findings, warnings) separately. It never returns a bare error: any
// structural/loader failure is translated into a Finding so a single
// malformed bundle does not abort validation of its siblings (and so
// `--json` always has a machine-readable shape to render).
func validateBundleDir(fsys fs.FS, name string) (findings, warnings []Finding) {
	b, err := engine.Load(fsys, name)
	if err != nil {
		return []Finding{loadErrorFinding(name, err)}, nil
	}

	findings = append(findings, checkName(b)...)
	findings = append(findings, checkInterpolations(b)...)
	findings = append(findings, checkSchemaRefs(fsys, b)...)
	findings = append(findings, checkPrimaryKeysAndCursors(b)...)
	findings = append(findings, checkWritePathFields(b)...)
	findings = append(findings, checkAPISurface(b)...)
	findings = append(findings, checkCLISurface(b)...)
	findings = append(findings, checkDocsHeadings(b)...)
	findings = append(findings, checkFixtureSecrets(b)...)
	findings = append(findings, checkCLISurfaceSecrets(b)...)
	findings = append(findings, checkOperationsSecrets(b)...)
	findings = append(findings, checkConformanceSkipReason(b)...)
	findings = append(findings, checkDefaultTypeMismatch(b)...)
	warnings = append(warnings, checkIncrementalStartDateFormat(b)...)
	return findings, warnings
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
	case strings.Contains(msg, "operations.json"):
		file = "operations.json"
	case strings.Contains(msg, "api_surface.json"):
		file = "api_surface.json"
	case strings.Contains(msg, "cli_surface.json"):
		file = "cli_surface.json"
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
	if b.HTTP.Check != nil {
		check("streams.json", b.HTTP.Check.Path)
		// checkquery-ledger.md: base.check.query (RequestSpec.Query) is the
		// SAME QueryParam dialect as stream.Query, so its templates get the
		// SAME static validation stream.Query's entries already get below —
		// an entry templating an undeclared spec key is a
		// ruleInterpolationUnresolved finding, not just a runtime failure.
		for _, v := range b.HTTP.Check.Query {
			check("streams.json", v.Template)
		}
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
			check("streams.json", v.Template)
		}
		for _, v := range s.ComputedFields {
			check("streams.json", v)
		}
		// S4 engine mini-wave item 2: fan_out.ids_from.request.path is a
		// request path template exactly like s.Path — it must get the same
		// static ResolveCheck coverage (an undeclared spec key here would
		// otherwise only fail the first time the stream is actually read).
		if s.FanOut != nil && s.FanOut.IDsFrom.Request != nil {
			check("streams.json", s.FanOut.IDsFrom.Request.Path)
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
//  1. every endpoint has exactly one executable covered_by row or an explicit
//     blocked non-executable classifier. Legacy surfaces use excluded;
//     operation-ledger surfaces use operation.
//  2. covered_by.stream/covered_by.write/covered_by.direct_read resolves to a
//     declared stream/action/implemented direct-read command, and every
//     declared stream/action appears in the surface.
//  3. excluded.category is from the closed vocabulary (defense-in-depth; the
//     loader's meta-schema enum already enforces this at load time), and
//     operation rows use the closed operation vocabulary.
//  4. capabilities.write/read == false is only legal when the surface has
//     zero executable mutation/GET endpoints respectively.
func checkAPISurface(b engine.Bundle) []Finding {
	if b.Surface == nil {
		return []Finding{{
			Connector: b.Name,
			File:      "api_surface.json",
			Rule:      ruleMissingFile,
			Message:   "api_surface.json is required for connector authoring and conformance",
		}}
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
	directReads := map[string]bool{}
	if b.CLISurface != nil {
		for _, cmd := range b.CLISurface.Commands {
			if cmd.Intent == "direct_read" && cmd.Availability == "implemented" {
				directReads[cmd.Path] = true
			}
		}
	}

	coveredStreams := map[string]bool{}
	coveredWrites := map[string]bool{}
	hasNonExcludedGET := false
	hasNonExcludedMutation := false
	ledgerMode := b.Surface.OperationLedgerVersion > 0

	for i, ep := range b.Surface.Endpoints {
		hasCovered := ep.CoveredBy != nil && (ep.CoveredBy.Stream != "" || ep.CoveredBy.Write != "" || len(coveredDirectReadTargets(ep.CoveredBy)) > 0)
		hasExcluded := ep.Excluded != nil
		hasOperation := ep.Operation != nil

		if ledgerMode && hasExcluded {
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceOperation,
				Message: fmt.Sprintf("endpoint %d (%s %s) uses legacy excluded in operation_ledger_version mode", i, ep.Method, ep.Path),
			})
		}
		if !ledgerMode && hasOperation {
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceOperation,
				Message: fmt.Sprintf("endpoint %d (%s %s) uses operation without operation_ledger_version", i, ep.Method, ep.Path),
			})
		}

		switch {
		case hasCovered && (hasExcluded || hasOperation):
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
				Message: fmt.Sprintf("endpoint %d (%s %s) has covered_by plus another classifier", i, ep.Method, ep.Path),
			})
		case !hasCovered && !hasExcluded && !hasOperation:
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
				Message: fmt.Sprintf("endpoint %d (%s %s) has no classifier", i, ep.Method, ep.Path),
			})
		case ledgerMode && hasOperation && hasExcluded:
			findings = append(findings, Finding{
				Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
				Message: fmt.Sprintf("endpoint %d (%s %s) has both operation and excluded", i, ep.Method, ep.Path),
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
			for _, directRead := range coveredDirectReadTargets(ep.CoveredBy) {
				if !directReads[directRead] {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceUnknownTarget,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.direct_read %q is not an implemented direct_read command", i, ep.Method, ep.Path, directRead),
					})
				}
				if !strings.EqualFold(ep.Method, "GET") {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.direct_read must use GET", i, ep.Method, ep.Path),
					})
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
		case hasOperation:
			findings = append(findings, checkAPISurfaceOperation(b, i, ep)...)
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

func checkAPISurfaceOperation(b engine.Bundle, i int, ep engine.SurfaceEndpoint) []Finding {
	op := ep.Operation
	if op == nil {
		return nil
	}

	var findings []Finding
	add := func(message string) {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "api_surface.json",
			Rule:      ruleSurfaceOperation,
			Message:   fmt.Sprintf("endpoint %d (%s %s) %s", i, ep.Method, ep.Path, message),
		})
	}

	if !surfaceOperationModels[op.Model] {
		add(fmt.Sprintf("operation.model %q is not in the closed vocabulary", op.Model))
	}
	if !surfaceOperationStatuses[op.Status] {
		add(fmt.Sprintf("operation.status %q is not in the closed vocabulary", op.Status))
	}
	if !surfaceOperationRisks[op.Risk] {
		add(fmt.Sprintf("operation.risk %q is not in the closed vocabulary", op.Risk))
	}
	if !op.BlockedByDefault {
		add("operation.blocked_by_default must be true")
	}
	if strings.TrimSpace(op.Reason) == "" {
		add("operation.reason is required")
	}
	if op.Model == "duplicate" && strings.TrimSpace(op.DuplicateOf) == "" {
		add("operation.duplicate_of is required for duplicate rows")
	}
	if sourceRequiredOperationModels[op.Model] &&
		strings.TrimSpace(op.SourceURL) == "" &&
		strings.TrimSpace(op.Notes) == "" {
		add("operation.source_url or operation.notes is required for sensitive/admin/destructive/disallowed rows")
	}
	return findings
}

// checkCLISurface validates optional docs-only connector command metadata.
// It deliberately validates references without enabling any command dispatch.
func checkCLISurface(b engine.Bundle) []Finding {
	if b.CLISurface == nil {
		return nil
	}

	streams := map[string]bool{}
	for _, s := range b.Streams {
		streams[s.Name] = true
	}
	writes := map[string]bool{}
	for _, w := range b.Writes {
		writes[w.Name] = true
	}
	operations := map[string]engine.OperationSpec{}
	for _, op := range b.Operations {
		operations[op.ID] = op
	}
	endpoints := cliSurfaceEndpointStates(b.Surface)

	var findings []Finding
	for i, cmd := range b.CLISurface.Commands {
		findings = append(findings, checkCLISurfaceReferences(b, i, cmd, streams, writes, operations)...)
		findings = append(findings, checkCLISurfaceIntent(b, i, cmd)...)
		findings = append(findings, checkCLISurfaceRiskApproval(b, i, cmd)...)
		findings = append(findings, checkCLISurfaceEndpointCoverage(b, i, cmd, endpoints)...)
	}
	return findings
}

func checkCLISurfaceReferences(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	streams map[string]bool,
	writes map[string]bool,
	operations map[string]engine.OperationSpec,
) []Finding {
	var findings []Finding
	mappings := 0
	if cmd.Stream != "" {
		mappings++
	}
	if cmd.Write != "" {
		mappings++
	}
	if cmd.Operation != "" {
		mappings++
	}
	if mappings > 1 {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("command %d (%q) must not reference more than one executable target (stream, write, operation)", i, cmd.Path),
		})
	}
	if cmd.Stream != "" && !streams[cmd.Stream] {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceUnknownTarget,
			Message:   fmt.Sprintf("command %d (%q) references unknown stream %q", i, cmd.Path, cmd.Stream),
		})
	}
	if cmd.Write != "" && !writes[cmd.Write] {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceUnknownTarget,
			Message:   fmt.Sprintf("command %d (%q) references unknown write action %q", i, cmd.Path, cmd.Write),
		})
	}
	if cmd.Operation != "" {
		if _, ok := operations[cmd.Operation]; !ok {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceUnknownTarget,
				Message:   fmt.Sprintf("command %d (%q) references unknown operation %q", i, cmd.Path, cmd.Operation),
			})
		}
	}
	return findings
}

func checkCLISurfaceIntent(b engine.Bundle, i int, cmd engine.CLICommand) []Finding {
	if cmd.Availability != "implemented" {
		return nil
	}

	switch cmd.Intent {
	case "etl":
		if cmd.Stream == "" && cmd.Operation == "" {
			return []Finding{{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceMissingMapping,
				Message:   fmt.Sprintf("implemented ETL command %d (%q) must reference stream", i, cmd.Path),
			}}
		}
	case "reverse_etl":
		if cmd.Write == "" && cmd.Operation == "" {
			return []Finding{{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceMissingMapping,
				Message:   fmt.Sprintf("implemented reverse ETL command %d (%q) must reference write action", i, cmd.Path),
			}}
		}
	case "direct_read":
		if cmd.Operation != "" {
			return nil
		}
		var findings []Finding
		if len(cmd.APISurface) != 1 {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceMissingMapping,
				Message:   fmt.Sprintf("implemented direct read command %d (%q) must reference exactly one api_surface endpoint", i, cmd.Path),
			})
		}
		for _, ep := range cmd.APISurface {
			if strings.ToUpper(strings.TrimSpace(ep.Method)) != "GET" {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented direct read command %d (%q) must reference a GET api_surface endpoint, got %s", i, cmd.Path, strings.ToUpper(ep.Method)),
				})
			}
			if isAbsoluteHTTPURL(ep.Path) {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented direct read command %d (%q) must reference a connector-relative api_surface endpoint", i, cmd.Path),
				})
			}
		}
		if !directReadOutputPolicies[cmd.OutputPolicy] {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented direct read command %d (%q) must declare a supported output_policy", i, cmd.Path),
			})
		}
		if len(findings) > 0 {
			return findings
		}
	case "local_workflow":
		if cmd.Operation != "" {
			return nil
		}
		return []Finding{{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented local workflow command %d (%q) must reference a typed operation", i, cmd.Path),
		}}
	default:
		return []Finding{{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented command %d (%q) has unsupported executable intent %q", i, cmd.Path, cmd.Intent),
		}}
	}
	return nil
}

func checkCLISurfaceRiskApproval(b engine.Bundle, i int, cmd engine.CLICommand) []Finding {
	if (cmd.Availability != "implemented" && cmd.Availability != "partial") || cmd.Intent != "reverse_etl" {
		return nil
	}

	var findings []Finding
	if strings.TrimSpace(cmd.Risk) == "" {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("reverse ETL command %d (%q) must declare risk text", i, cmd.Path),
		})
	}
	if strings.TrimSpace(cmd.Approval) == "" {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("reverse ETL command %d (%q) must declare approval text", i, cmd.Path),
		})
	}
	return findings
}

func checkCLISurfaceEndpointCoverage(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	endpoints map[string]cliSurfaceEndpointState,
) []Finding {
	if b.Surface == nil {
		return nil
	}

	var findings []Finding
	for _, ep := range cmd.APISurface {
		state, ok := endpoints[surfaceEndpointKey(ep.Method, ep.Path)]
		if !ok {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceUnknownTarget,
				Message:   fmt.Sprintf("command %d (%q) references unknown api_surface endpoint %s %s", i, cmd.Path, strings.ToUpper(ep.Method), ep.Path),
			})
			continue
		}
		if state.excluded || state.operation != nil || state.coveredBy == nil || (state.coveredBy.Stream == "" && state.coveredBy.Write == "") {
			if cmd.Intent == "direct_read" && directReadCoverageMatches(state.coveredBy, cmd.Path) {
				continue
			}
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("command %d (%q) references api_surface endpoint %s %s that is not covered by an executable surface", i, cmd.Path, strings.ToUpper(ep.Method), ep.Path),
			})
			continue
		}
		if cmd.Stream != "" && state.coveredBy.Stream != cmd.Stream {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("command %d (%q) references api_surface endpoint %s %s covered by stream %q, want %q", i, cmd.Path, strings.ToUpper(ep.Method), ep.Path, state.coveredBy.Stream, cmd.Stream),
			})
		}
		if cmd.Write != "" && state.coveredBy.Write != cmd.Write {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("command %d (%q) references api_surface endpoint %s %s covered by write %q, want %q", i, cmd.Path, strings.ToUpper(ep.Method), ep.Path, state.coveredBy.Write, cmd.Write),
			})
		}
	}
	return findings
}

func cliSurfaceEndpointStates(surface *engine.APISurface) map[string]cliSurfaceEndpointState {
	endpoints := map[string]cliSurfaceEndpointState{}
	if surface == nil {
		return endpoints
	}
	for _, ep := range surface.Endpoints {
		endpoints[surfaceEndpointKey(ep.Method, ep.Path)] = cliSurfaceEndpointState{
			coveredBy: ep.CoveredBy,
			excluded:  ep.Excluded != nil,
			operation: ep.Operation,
		}
	}
	return endpoints
}

type cliSurfaceEndpointState struct {
	coveredBy *engine.SurfaceCoverage
	excluded  bool
	operation *engine.SurfaceOperation
}

func coveredDirectReadTargets(covered *engine.SurfaceCoverage) []string {
	if covered == nil {
		return nil
	}
	targets := append([]string{}, covered.DirectReads...)
	if covered.DirectRead != "" {
		targets = append(targets, covered.DirectRead)
	}
	return targets
}

func directReadCoverageMatches(covered *engine.SurfaceCoverage, path string) bool {
	for _, target := range coveredDirectReadTargets(covered) {
		if target == path {
			return true
		}
	}
	return false
}

func surfaceEndpointKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func isAbsoluteHTTPURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
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

func checkCLISurfaceSecrets(b engine.Bundle) []Finding {
	if len(b.RawCLISurface) == 0 || !secretLiteralPattern.Match(b.RawCLISurface) {
		return nil
	}
	return []Finding{{
		Connector: b.Name,
		File:      "cli_surface.json",
		Rule:      ruleSecretLiteral,
		Message:   "cli_surface.json contains a secret-shaped literal",
	}}
}

func checkOperationsSecrets(b engine.Bundle) []Finding {
	if len(b.RawOperations) == 0 || !secretLiteralPattern.Match(b.RawOperations) {
		return nil
	}
	return []Finding{{
		Connector: b.Name,
		File:      "operations.json",
		Rule:      ruleSecretLiteral,
		Message:   "operations.json contains a secret-shaped literal",
	}}
}

// checkConformanceSkipReason enforces R3's skip-marker contract (docs/
// migration/conventions.md §4/§6): a bundle-level (metadata.json) or
// stream-level (streams.json) "conformance": {"skip_dynamic": true} marker
// MUST carry a non-empty, non-whitespace-only "reason" — an unreasoned
// skip is indistinguishable from silently hiding a real failure, which
// defeats the whole point of an EXPLICIT marker. A marker with
// skip_dynamic:false (or entirely absent) is never flagged, regardless of
// its reason field.
func checkConformanceSkipReason(b engine.Bundle) []Finding {
	var findings []Finding
	if m := b.Metadata.Conformance; m != nil && m.SkipDynamic && strings.TrimSpace(m.Reason) == "" {
		findings = append(findings, Finding{
			Connector: b.Name, File: "metadata.json", Rule: ruleConformanceSkipReason,
			Message: "metadata.json conformance.skip_dynamic is true but reason is empty",
		})
	}
	for _, s := range b.Streams {
		if s.Conformance == nil || !s.Conformance.SkipDynamic {
			continue
		}
		if strings.TrimSpace(s.Conformance.Reason) == "" {
			findings = append(findings, Finding{
				Connector: b.Name, File: "streams.json", Rule: ruleConformanceSkipReason,
				Message: fmt.Sprintf("stream %q conformance.skip_dynamic is true but reason is empty", s.Name),
			})
		}
	}
	return findings
}

// checkDefaultTypeMismatch is gap-loop cycle-1 item 6's validate rule
// (REVIEW-A.md C3: "Validate rule: default must type-check"). engine's
// `materializeConfigDefaults` (read.go) now fills an absent RuntimeConfig
// config key straight from spec.json's declared "default" value — a default
// whose JSON type mismatches its own property's declared "type" (e.g.
// `"type":"integer","default":"not-a-number"`) would silently materialize a
// wrong-shaped config value into every read/check that hits this bundle. A
// HARD FINDING (not a warning, unlike checkIncrementalStartDateFormat's N2
// plausibility heuristic below): this is a structural defect in the bundle
// author's own spec.json, always fixable by correcting the default, never a
// legitimate authoring choice worth tolerating.
func checkDefaultTypeMismatch(b engine.Bundle) []Finding {
	if b.Spec == nil {
		return nil
	}
	mismatches := b.Spec.DefaultTypeMismatches()
	if len(mismatches) == 0 {
		return nil
	}
	findings := make([]Finding, 0, len(mismatches))
	for _, name := range mismatches {
		findings = append(findings, Finding{
			Connector: b.Name, File: "spec.json", Rule: ruleDefaultTypeMismatch,
			Message: fmt.Sprintf("spec.json property %q declares a \"default\" value that does not type-check against its own declared \"type\"", name),
		})
	}
	return findings
}

// checkIncrementalStartDateFormat is N2's narrow, honest WARNING (wave0
// REVIEW.md carried flag; SPEC.md §4 "promote to a validate-time guard"):
// for every stream whose incremental.param_format is "date" or
// "github_date_range" (the two formats where engine/read.go's
// parseLowerBoundTime treats an all-digits value as Unix seconds and
// anything else as RFC3339, N4/B1) AND which names a start_config_key,
// check whether that spec.json property declares a date-ish JSON Schema
// "format" (date-time/date). If it does not, a digit-shaped config value —
// e.g. an operator typo like "20260101" (yyyymmdd), which is NOT Unix
// seconds — would silently be treated as one instead of erroring, producing
// a bogus 1970s-era lower bound. This is deliberately scoped to ONLY these
// two param_formats: unix_seconds is excluded because there an all-digits
// value IS the correct, intended shape (no misinterpretation risk at all),
// and rfc3339 never attempts digit parsing in the first place (verbatim
// passthrough). Reads spec.json's per-property "format" directly from
// b.RawSpec (F5, REVIEW.md) since the compiled *engine.Schema does not
// expose annotation keywords like "format" through any accessor (schema.go:
// "format" is accepted-but-only-preserved, never structurally enforced).
func checkIncrementalStartDateFormat(b engine.Bundle) []Finding {
	if len(b.RawSpec) == 0 {
		return nil
	}
	var findings []Finding
	seen := map[string]bool{} // de-dupe: multiple streams may share one start_config_key
	for _, s := range b.Streams {
		if s.Incremental == nil || s.Incremental.StartConfigKey == "" {
			continue
		}
		if !dateShapedParamFormats[s.Incremental.ParamFormat] {
			continue
		}
		key := s.Incremental.StartConfigKey
		if seen[key] {
			continue
		}
		if specPropertyHasDateShapedFormat(b.RawSpec, key) {
			continue
		}
		seen[key] = true
		findings = append(findings, Finding{
			Connector: b.Name, File: "spec.json", Rule: ruleStartDateFreeFormString,
			Message: fmt.Sprintf("spec.json property %q is used as stream %q's incremental.start_config_key with param_format %q but declares no date-ish \"format\" (date-time/date) — a digit-shaped value (e.g. a yyyymmdd typo) would be silently misinterpreted as Unix seconds rather than erroring", key, s.Name, s.Incremental.ParamFormat),
		})
	}
	return findings
}

// specPropertyHasDateShapedFormat reports whether rawSpec's top-level
// properties.<key>.format is one of dateShapedSpecFormats. Any parse
// failure or absence is treated as "no date-ish format declared" (the
// warning-worthy case), not an error — spec.json's own structural validity
// is already enforced by the loader's meta-schema check elsewhere.
func specPropertyHasDateShapedFormat(rawSpec []byte, key string) bool {
	var doc struct {
		Properties map[string]struct {
			Format string `json:"format"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(rawSpec, &doc); err != nil {
		return false
	}
	prop, ok := doc.Properties[key]
	if !ok {
		return false
	}
	return dateShapedSpecFormats[prop.Format]
}
