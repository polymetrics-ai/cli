package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"mime"
	"regexp"
	"slices"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

// Rule identifiers named by every validate Finding. Kept as exported-looking
// constants (lowercase, package-private) so tests can assert on them without
// string literals scattered across the corpus.
const (
	ruleMissingFile               = "missing_file"
	ruleMetaSchema                = "meta_schema"
	ruleInterpolationUnresolved   = "interpolation_unresolved"
	ruleSchemaRefMissing          = "schema_ref_missing"
	rulePrimaryKeyMissing         = "primary_key_missing"
	ruleCursorFieldMissing        = "cursor_field_missing"
	ruleIncrementalCursorMismatch = "incremental_cursor_schema_mismatch"
	ruleWritePathFields           = "write_path_fields"
	ruleSurfaceCoverage           = "surface_coverage"
	ruleSurfaceUnknownTarget      = "surface_unknown_target"
	ruleSurfaceIncomplete         = "surface_incomplete"
	ruleSurfaceCategory           = "surface_category"
	ruleSurfaceOperation          = "surface_operation"
	ruleSurfaceProvenance         = "surface_provenance"
	ruleSurfaceFailFirstRun       = "surface_fail_first_run"
	ruleCLISurfaceUnknownTarget   = "cli_surface_unknown_target"
	ruleCLISurfaceMissingMapping  = "cli_surface_missing_mapping"
	ruleCLISurfaceSafety          = "cli_surface_safety"
	ruleNameRegex                 = "name_regex"
	ruleSecretLiteral             = "secret_literal"
	ruleDocsHeading               = "docs_heading"
	ruleStartDateFreeFormString   = "start_date_free_form_string"
	ruleConformanceSkipReason     = "conformance_skip_reason"
	ruleDefaultTypeMismatch       = "default_type_mismatch"
	ruleIncrementalPolicy         = "incremental_policy"
)

var supportedParamFormats = map[string]bool{
	"":             true,
	"rfc3339":      true,
	"rfc3339_utc":  true,
	"unix_seconds": true,
	"date":         true,
}

var allowedOperatorPrefixes = map[string]bool{
	"":   true,
	">=": true,
	">":  true,
	"<=": true,
	"<":  true,
}

// dateShapedParamFormats are the incremental.param_format values whose
// value-parsing path (engine/read.go parseLowerBoundTime, N4/B1) accepts an
// all-digits input as Unix seconds and otherwise requires RFC3339. For these
// formats specifically (unlike unix_seconds, where digits ARE the
// correct/intended shape), a digit-shaped config value that is NOT actually
// Unix seconds — e.g. a yyyymmdd typo like "20260101" — is silently
// misinterpreted as a 1970s-era lower bound rather than erroring (N2, wave0
// REVIEW.md carried flag).
var dateShapedParamFormats = map[string]bool{
	"date":        true,
	"rfc3339_utc": true,
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

const maxCLIRecordPathArrayIndex = 128

var directReadOutputPolicies = map[string]bool{
	"repository_contents_file_metadata": true,
	"repository_contents_directory":     true,
	"json_redacted":                     true,
	"clinical_json_redacted":            true,
	"none":                              true,
	"text":                              true,
}

// directWriteOutputPolicies mirrors engine.validateOperationDirectWriteOutputPolicy
// and commandrunner.isSupportedDirectWriteOutputPolicy. A write command must
// name the operation's exact policy: unlike direct reads, the operation's
// response is part of the preview-bound write contract, not a display choice.
var directWriteOutputPolicies = map[string]bool{
	"none":                        true,
	"json":                        true,
	"json_redacted":               true,
	"write_result_redacted":       true,
	"gong_bounded_input_redacted": true,
}

var repositoryDirectReadOutputPolicies = map[string]bool{
	"repository_contents_file_metadata": true,
	"repository_contents_directory":     true,
}

var operationOnlyDirectReadOutputPolicies = map[string]bool{
	"none": true,
	"text": true,
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
	findings = append(findings, checkCertificationSecrets(b)...)
	findings = append(findings, checkConformanceSkipReason(b)...)
	findings = append(findings, checkDefaultTypeMismatch(b)...)
	findings = append(findings, checkIncrementalPolicies(b)...)
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
	case strings.Contains(msg, "certification.json"):
		file = "certification.json"
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

	checkPath := func(file, what, template string) {
		if template == "" {
			return
		}
		if err := engine.ResolveCheckRequestPath(what, template, specKeys); err != nil {
			findings = append(findings, Finding{
				Connector: b.Name, File: file, Rule: ruleInterpolationUnresolved,
				Message: err.Error(),
			})
		}
	}

	checkPath("streams.json", "base.url", b.HTTP.URL)
	for _, h := range b.HTTP.Headers {
		check("streams.json", h)
	}
	if b.HTTP.Check != nil {
		checkPath("streams.json", "base.check.path", b.HTTP.Check.Path)
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
		checkPath("streams.json", fmt.Sprintf("stream %q path", s.Name), s.Path)
		for _, v := range s.Query {
			check("streams.json", v.Template)
		}
		if s.GraphQL != nil {
			if err := engine.ResolveCheckGraphQLVariables(s.GraphQL.Variables, specKeys); err != nil {
				findings = append(findings, Finding{
					Connector: b.Name, File: "streams.json", Rule: ruleInterpolationUnresolved,
					Message: fmt.Sprintf("stream %q: %v", s.Name, err),
				})
			}
		}
		for _, v := range s.ComputedFields {
			check("streams.json", v)
		}
		// S4 engine mini-wave item 2: fan_out.ids_from.request.path is a
		// request path template exactly like s.Path — it must get the same
		// static ResolveCheck coverage (an undeclared spec key here would
		// otherwise only fail the first time the stream is actually read).
		if s.FanOut != nil && s.FanOut.IDsFrom.Request != nil {
			checkPath("streams.json", fmt.Sprintf("stream %q fan_out.ids_from.request.path", s.Name), s.FanOut.IDsFrom.Request.Path)
		}
	}
	for _, w := range b.Writes {
		check("writes.json", w.Path)
		if w.GraphQL != nil {
			if err := engine.ResolveCheckGraphQLVariables(w.GraphQL.Variables, specKeys); err != nil {
				findings = append(findings, Finding{
					Connector: b.Name, File: "writes.json", Rule: ruleInterpolationUnresolved,
					Message: fmt.Sprintf("write action %q: %v", w.Name, err),
				})
			}
		}
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
			if sch.CursorField != "" && sch.CursorField != s.Incremental.CursorField {
				findings = append(findings, Finding{
					Connector: b.Name, File: s.SchemaRef, Rule: ruleIncrementalCursorMismatch,
					Message: fmt.Sprintf("stream %q incremental.cursor_field %q does not match schema %q x-cursor-field %q", s.Name, s.Incremental.CursorField, s.SchemaRef, sch.CursorField),
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
	for _, issue := range engine.ValidateSurfaceProvenance(b.Surface).Issues {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "api_surface.json",
			Rule:      ruleSurfaceProvenance,
			Message:   issue.Error(),
		})
	}

	streams := map[string]bool{}
	for _, s := range b.Streams {
		streams[s.Name] = true
	}
	writes := map[string]bool{}
	for _, w := range b.Writes {
		writes[w.Name] = true
	}
	operationSpecs := map[string]engine.OperationSpec{}
	for _, operation := range b.Operations {
		operationSpecs[operation.ID] = operation
	}
	directReads := map[string]bool{}
	if b.CLISurface != nil {
		for _, cmd := range b.CLISurface.Commands {
			// binary_download commands consume an api_surface endpoint the same
			// way a direct read does and are tracked by the same covered_by
			// bookkeeping, so they satisfy that coverage too.
			if (cmd.Intent == "direct_read" || cmd.Intent == "binary_download") &&
				cmd.Availability == "implemented" {
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
		hasCovered := ep.CoveredBy != nil && (ep.CoveredBy.Stream != "" || len(ep.CoveredBy.WriteTargets()) > 0 || len(coveredDirectReadTargets(ep.CoveredBy)) > 0 || len(ep.CoveredBy.OperationTargets()) > 0)
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
			// Singular and plural are checked identically: one endpoint may
			// back several distinct write contracts, but every one of them
			// still has to be a write action the bundle actually declares.
			for _, write := range ep.CoveredBy.WriteTargets() {
				if !writes[write] {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceUnknownTarget,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.write %q is not a declared write action", i, ep.Method, ep.Path, write),
					})
				} else {
					coveredWrites[write] = true
				}
			}
			for _, directRead := range coveredDirectReadTargets(ep.CoveredBy) {
				if !directReads[directRead] {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceUnknownTarget,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.direct_read %q is not an implemented direct_read command", i, ep.Method, ep.Path, directRead),
					})
				}
				method := strings.ToUpper(strings.TrimSpace(ep.Method))
				if method != "GET" && method != "POST" {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.direct_read must use GET or POST", i, ep.Method, ep.Path),
					})
				}
			}
			for _, operationID := range ep.CoveredBy.OperationTargets() {
				operation, ok := operationSpecs[operationID]
				if !ok {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceUnknownTarget,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.operations %q is not a declared operation", i, ep.Method, ep.Path, operationID),
					})
					continue
				}
				if operation.Kind != "graphql_query" && operation.Kind != "graphql_mutation" {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.operations %q must be a fixed GraphQL operation, got %q", i, ep.Method, ep.Path, operationID, operation.Kind),
					})
					continue
				}
				if !strings.EqualFold(ep.Method, "POST") || operation.GraphQL == nil || operation.GraphQL.Path != ep.Path {
					findings = append(findings, Finding{
						Connector: b.Name, File: "api_surface.json", Rule: ruleSurfaceCoverage,
						Message: fmt.Sprintf("endpoint %d (%s %s) covered_by.operations %q does not match its declared fixed GraphQL transport", i, ep.Method, ep.Path, operationID),
					})
				}
				switch operation.Kind {
				case "graphql_query":
					// GraphQL queries use the shared POST transport, but they
					// are still executable reads for the capability contract.
					hasNonExcludedGET = true
				case "graphql_mutation":
					hasNonExcludedMutation = true
				}
			}
			if strings.EqualFold(ep.Method, "GET") {
				hasNonExcludedGET = true
			}
			if len(ep.CoveredBy.WriteTargets()) > 0 && mutationMethods[strings.ToUpper(ep.Method)] {
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
			Message: "capabilities.read is false but api_surface.json has a non-excluded executable read surface",
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
	if b.Surface.OperationLedgerVersion < 2 && sourceRequiredOperationModels[op.Model] &&
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
	writes := map[string]engine.WriteAction{}
	for _, w := range b.Writes {
		writes[w.Name] = w
	}
	operations := map[string]engine.OperationSpec{}
	for _, op := range b.Operations {
		operations[op.ID] = op
	}
	endpoints := cliSurfaceEndpointStates(b.Surface)

	var findings []Finding
	for i, cmd := range b.CLISurface.Commands {
		findings = append(findings, checkCLISurfaceReferences(b, i, cmd, streams, writes, operations)...)
		findings = append(findings, checkCLISurfaceOperationSafety(b, i, cmd, operations)...)
		findings = append(findings, checkCLISurfaceIntent(b, i, cmd)...)
		findings = append(findings, checkCLISurfaceRiskApproval(b, i, cmd)...)
		findings = append(findings, checkCLISurfaceValidationDeclarations(b, i, cmd)...)
		findings = append(findings, checkCLISurfaceEnvOnlyFlags(b, i, cmd, operations)...)
		findings = append(findings, checkCLISurfaceStructuredJSONFlags(b, i, cmd)...)
		findings = append(findings, checkCLISurfaceWriteFlags(b, i, cmd, writes)...)
		findings = append(findings, checkCLISurfaceEndpointCoverage(b, i, cmd, endpoints)...)
	}
	return findings
}

// checkCLISurfaceEnvOnlyFlags keeps --from-env a narrow secret-delivery
// channel rather than a second, untyped source of arbitrary command values.
// An env_only declaration is permitted only for the complete JSON input of a
// sensitive fixed GraphQL mutation, where the operation itself declares the
// corresponding redaction and typed-confirmation contract.
func checkCLISurfaceEnvOnlyFlags(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	operations map[string]engine.OperationSpec,
) []Finding {
	var findings []Finding
	for _, flag := range cmd.Flags {
		if !flag.EnvOnly {
			continue
		}
		op, found := operations[cmd.Operation]
		variable, mapsToBody := strings.CutPrefix(strings.TrimSpace(flag.MapsTo), "body.")
		valid := cmd.Availability == "implemented" &&
			cmd.Intent == "direct_write" &&
			flag.Type == "json" &&
			flag.Required &&
			mapsToBody && variable != "" && !strings.Contains(variable, ".") &&
			found && op.Kind == "graphql_mutation" &&
			strings.EqualFold(op.MutationClass, "secret") &&
			op.SensitivePolicy != nil &&
			strings.EqualFold(op.SensitivePolicy.InputMode, "env") &&
			strings.EqualFold(op.SensitivePolicy.ApprovalMode, "typed_confirmation") &&
			slices.Contains(op.SensitivePolicy.RedactFields, "body."+variable)
		if valid {
			continue
		}
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("command %d (%q) env_only flag --%s must be a required top-level JSON input for an implemented secret GraphQL mutation with env redaction and typed confirmation", i, cmd.Path, flag.Name),
		})
	}
	return findings
}

// checkCLISurfaceStructuredJSONFlags keeps the schema vocabulary closed even
// though the bundle meta-schema must recognize `json` before the engine can
// inspect an authored command. Runtime preflight permits it only for a named,
// top-level reverse-ETL record field; this static half rejects every other
// executable placement before it can become an `implemented` claim.
func checkCLISurfaceStructuredJSONFlags(b engine.Bundle, i int, cmd engine.CLICommand) []Finding {
	if cmd.Availability != "implemented" {
		return nil
	}
	operations := make(map[string]engine.OperationSpec, len(b.Operations))
	for _, operation := range b.Operations {
		operations[operation.ID] = operation
	}
	var findings []Finding
	for _, flag := range cmd.Flags {
		if flag.Type != "json" {
			continue
		}
		switch {
		case cmd.Intent == "reverse_etl" && strings.TrimSpace(cmd.Write) != "" && strings.HasPrefix(flag.MapsTo, "record."):
			continue
		case (cmd.Intent == "direct_read" || cmd.Intent == "direct_write") && strings.TrimSpace(cmd.Operation) != "":
			variable, ok := strings.CutPrefix(flag.MapsTo, "body.")
			operation, found := operations[cmd.Operation]
			if ok && variable != "" && !strings.Contains(variable, ".") && found {
				if err := engine.ValidateGraphQLOperationStructuredJSONVariable(operation, variable); err == nil {
					continue
				} else {
					findings = append(findings, Finding{
						Connector: b.Name,
						File:      "cli_surface.json",
						Rule:      ruleCLISurfaceSafety,
						Message:   fmt.Sprintf("implemented command %d (%q) fixed GraphQL variable --%s is not declared safely: %v", i, cmd.Path, flag.Name, err),
					})
					continue
				}
			}
		}
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented command %d (%q) structured JSON flag --%s is allowed only on a declared reverse-ETL record field or fixed GraphQL variable", i, cmd.Path, flag.Name),
		})
	}
	return findings
}

func checkCLISurfaceGraphQLOperationSafety(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	op engine.OperationSpec,
) []Finding {
	var findings []Finding
	if len(cmd.APISurface) != 1 {
		return []Finding{{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented GraphQL command %d (%q) requires exactly one api_surface endpoint", i, cmd.Path),
		}}
	}
	endpoint := cmd.APISurface[0]
	switch op.Kind {
	case "graphql_query":
		if cmd.Intent != "direct_read" {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented GraphQL query command %d (%q) must use direct_read intent", i, cmd.Path),
			})
			return findings
		}
		maxBytes := 1
		if op.GraphQL != nil {
			maxBytes = op.GraphQL.MaxBytes
		}
		if err := engine.PreflightOperationDirectRead(b, cmd.Operation, endpoint.Method, endpoint.Path, maxBytes, cmd.OutputPolicy); err != nil {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented GraphQL query command %d (%q) operation %q is not executable: %v", i, cmd.Path, cmd.Operation, err),
			})
		}
	case "graphql_mutation":
		if cmd.Intent != "direct_write" {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented GraphQL mutation command %d (%q) must use direct_write intent", i, cmd.Path),
			})
			return findings
		}
		if err := engine.PreflightOperationDirectWrite(b, cmd.Operation, endpoint.Method, endpoint.Path, cmd.OutputPolicy); err != nil {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented GraphQL mutation command %d (%q) operation %q is not executable: %v", i, cmd.Path, cmd.Operation, err),
			})
		}
	default:
		return []Finding{{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented GraphQL command %d (%q) operation %q has unsupported kind %q", i, cmd.Path, cmd.Operation, op.Kind),
		}}
	}
	for _, flag := range cmd.Flags {
		target, ok := strings.CutPrefix(strings.TrimSpace(flag.MapsTo), "body.")
		if !ok || target == "" || strings.Contains(target, ".") {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented GraphQL command %d (%q) flag --%s must map to one top-level body.<variable>", i, cmd.Path, flag.Name),
			})
		}
	}
	return findings
}

func checkCLISurfaceReferences(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	streams map[string]bool,
	writes map[string]engine.WriteAction,
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
	if cmd.Write != "" {
		if _, ok := writes[cmd.Write]; !ok {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceUnknownTarget,
				Message:   fmt.Sprintf("command %d (%q) references unknown write action %q", i, cmd.Path, cmd.Write),
			})
		}
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

func checkCLISurfaceOperationSafety(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	operations map[string]engine.OperationSpec,
) []Finding {
	// An empty output_policy is NOT a reason to skip: it is precisely the
	// state the runtime rejects. Skipping here is what gave GitHub's
	// operation-backed commands zero validation from either path.
	if cmd.Availability != "implemented" || cmd.Operation == "" {
		return nil
	}
	op, ok := operations[cmd.Operation]
	if !ok {
		return nil
	}
	if cmd.Intent == "binary_download" {
		return checkCLISurfaceBinaryOperationSafety(b, i, cmd, op)
	}
	if op.Kind == "graphql_query" || op.Kind == "graphql_mutation" {
		return checkCLISurfaceGraphQLOperationSafety(b, i, cmd, op)
	}
	if cmd.Intent == "direct_write" {
		return checkCLISurfaceDirectWriteOperationSafety(b, i, cmd, op)
	}
	if cmd.Intent != "direct_read" {
		return []Finding{{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented command %d (%q) references operation %q, but its intent has no operation executor", i, cmd.Path, cmd.Operation),
		}}
	}
	var findings []Finding
	if op.Kind != "rest_read" || op.REST == nil {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q must be rest_read", i, cmd.Path, cmd.Operation)})
		return findings
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	if method != "GET" && method != "POST" {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q must use GET or POST, got %s", i, cmd.Path, cmd.Operation, method)})
	}
	if isAbsoluteHTTPURL(op.REST.Path) {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q must use connector-relative path", i, cmd.Path, cmd.Operation)})
	}
	if op.REST.MaxBytes <= 0 {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q must declare positive rest.max_bytes", i, cmd.Path, cmd.Operation)})
	}
	contentType := ""
	if method == "POST" {
		mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(op.REST.ContentType))
		if err != nil {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q POST has invalid content_type %q", i, cmd.Path, cmd.Operation, op.REST.ContentType)})
		} else {
			contentType = strings.ToLower(mediaType)
		}
		if len(op.REST.BodySchema) == 0 {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q POST must declare body_schema", i, cmd.Path, cmd.Operation)})
		} else if contentType == "application/json" {
			findings = append(findings, checkCLISurfaceOperationBodyMappings(b, i, cmd, op)...)
		} else if contentType == "text/plain" {
			findings = append(findings, checkCLISurfaceOperationPlainTextBodyMapping(b, i, cmd, op)...)
		} else {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q POST must declare application/json or text/plain content_type", i, cmd.Path, cmd.Operation)})
		}
	}
	if !directReadOutputPolicies[cmd.OutputPolicy] {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q must declare a supported output_policy", i, cmd.Path, cmd.Operation)})
	}
	if repositoryDirectReadOutputPolicies[cmd.OutputPolicy] && !endpointPathHasVariable(op.REST.Path, "path") {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q uses repository output policy %q but endpoint path lacks {path}", i, cmd.Path, cmd.Operation, cmd.OutputPolicy)})
	}
	for _, flag := range cmd.Flags {
		mapsTo := strings.TrimSpace(flag.MapsTo)
		switch {
		case strings.HasPrefix(mapsTo, "path."), strings.HasPrefix(mapsTo, "query."):
			// allowed
		case strings.HasPrefix(mapsTo, "body."):
			if method != "POST" {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) flag --%s maps to body for non-POST operation", i, cmd.Path, flag.Name)})
			} else if contentType == "text/plain" {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) flag --%s maps JSON body fields for text/plain operation", i, cmd.Path, flag.Name)})
			}
		case mapsTo == "body":
			if method != "POST" || contentType != "text/plain" {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) flag --%s maps to raw body without a text/plain POST operation", i, cmd.Path, flag.Name)})
			}
		default:
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) flag --%s maps to unsupported target %q", i, cmd.Path, flag.Name, flag.MapsTo)})
		}
	}
	return findings
}

// checkCLISurfaceDirectWriteOperationSafety validates the operation-specific
// half of a direct_write command. checkCLISurfaceIntent below independently
// mirrors commandrunner's command-shape preflight; this helper makes the
// operation itself safe for the engine's typed no-retry executor.
func checkCLISurfaceDirectWriteOperationSafety(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	op engine.OperationSpec,
) []Finding {
	var findings []Finding
	if op.Kind != "rest_write" || op.REST == nil {
		return []Finding{{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented direct write command %d (%q) operation %q must be rest_write", i, cmd.Path, cmd.Operation),
		}}
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	if !mutationMethods[method] {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) operation %q must use POST, PUT, PATCH, or DELETE, got %s", i, cmd.Path, cmd.Operation, method)})
	}
	if isAbsoluteHTTPURL(op.REST.Path) {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) operation %q must use connector-relative path", i, cmd.Path, cmd.Operation)})
	}
	if op.REST.MaxBytes <= 0 {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) operation %q must declare positive rest.max_bytes", i, cmd.Path, cmd.Operation)})
	}
	if !supportedDirectWriteContentType(op.REST) {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) operation %q must use application/json, application/x-www-form-urlencoded, no content_type, or literal multipart/form-data with rest.multipart", i, cmd.Path, cmd.Operation)})
	}
	if !directWriteOutputPolicies[cmd.OutputPolicy] {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) operation %q must declare a supported output_policy", i, cmd.Path, cmd.Operation)})
	} else if cmd.OutputPolicy != op.OutputPolicy {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) output_policy %q must match operation policy %q", i, cmd.Path, cmd.OutputPolicy, op.OutputPolicy)})
	}

	bodyMapped := false
	for _, flag := range cmd.Flags {
		mapsTo := strings.TrimSpace(flag.MapsTo)
		switch {
		case strings.HasPrefix(mapsTo, "path."), strings.HasPrefix(mapsTo, "query."):
			// Typed path/query bindings are supported by the shared operation
			// shaper and validated again at command runtime.
		case strings.HasPrefix(mapsTo, "body."):
			bodyMapped = true
		default:
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) flag --%s maps to unsupported target %q", i, cmd.Path, flag.Name, flag.MapsTo)})
		}
	}
	if bodyMapped && len(op.REST.BodySchema) == 0 {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct write command %d (%q) maps flags to body but operation %q has no body_schema", i, cmd.Path, op.ID)})
	}
	if len(op.REST.BodySchema) > 0 {
		findings = append(findings, checkCLISurfaceOperationBodyMappingsForIntent(b, i, cmd, op, "direct write")...)
	}
	return findings
}

func supportedDirectWriteContentType(rest *engine.RESTOperationSpec) bool {
	if rest == nil {
		return false
	}
	if rest.Multipart != nil {
		return rest.ContentType == "multipart/form-data"
	}
	raw := strings.TrimSpace(rest.ContentType)
	if raw == "" {
		return true
	}
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return false
	}
	return strings.EqualFold(mediaType, "application/json") ||
		strings.EqualFold(mediaType, "application/x-www-form-urlencoded")
}

func checkCLISurfaceOperationBodyMappings(b engine.Bundle, i int, cmd engine.CLICommand, op engine.OperationSpec) []Finding {
	return checkCLISurfaceOperationBodyMappingsForIntent(b, i, cmd, op, "direct read")
}

// checkCLISurfaceOperationPlainTextBodyMapping admits the one literal body
// shape the operation direct-read executor supports. It deliberately does not
// reuse the dotted JSON mapper: that would make an arbitrary raw request
// shape look declared even though the engine cannot validate or serialize it.
func checkCLISurfaceOperationPlainTextBodyMapping(b engine.Bundle, i int, cmd engine.CLICommand, op engine.OperationSpec) []Finding {
	var rawMappings []engine.CLIFlag
	var findings []Finding
	for _, flag := range cmd.Flags {
		switch {
		case strings.TrimSpace(flag.MapsTo) == "body":
			rawMappings = append(rawMappings, flag)
		case strings.HasPrefix(strings.TrimSpace(flag.MapsTo), "body."):
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q text/plain body must use one raw body flag, not --%s", i, cmd.Path, op.ID, flag.Name)})
		}
	}
	if len(rawMappings) != 1 {
		return append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) operation %q text/plain body requires exactly one flag mapped to body", i, cmd.Path, op.ID)})
	}
	raw := rawMappings[0]
	if raw.Type != "string" {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) raw body flag --%s must have string type", i, cmd.Path, raw.Name)})
	}
	if !raw.Required {
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented direct read command %d (%q) raw body flag --%s must be required", i, cmd.Path, raw.Name)})
	}
	return findings
}

func checkCLISurfaceOperationBodyMappingsForIntent(b engine.Bundle, i int, cmd engine.CLICommand, op engine.OperationSpec, intent string) []Finding {
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return nil
	}
	schema, err := parseCLIRecordSchema(op.REST.BodySchema)
	if err != nil {
		return []Finding{{Connector: b.Name, File: "operations.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("operation %q has invalid body_schema for CLI validation: %v", op.ID, err)}}
	}
	requiredPaths := schema.requiredMappingPaths("")
	if len(requiredPaths) == 0 {
		return nil
	}
	mappedTargets := make([]cliBodyFlagMapping, 0, len(cmd.Flags))
	for _, flag := range cmd.Flags {
		target, ok := strings.CutPrefix(flag.MapsTo, "body.")
		if ok && target != "" {
			mappedTargets = append(mappedTargets, cliBodyFlagMapping{target: target, name: flag.Name, required: flag.Required})
		}
	}
	var findings []Finding
	for _, requiredPath := range requiredPaths {
		if operationStaticBodyProvidesPath(op.REST.Body, requiredPath) {
			continue
		}
		mapping, ok := commandBodyFlagCoveringRequiredPath(schema, mappedTargets, requiredPath)
		if !ok {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented %s command %d (%q) operation %q requires body.%s but no command flag maps to it and rest.body does not provide it", intent, i, cmd.Path, op.ID, requiredPath)})
			continue
		}
		if !mapping.required {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented %s command %d (%q) operation %q requires body.%s but flag --%s is not marked required", intent, i, cmd.Path, op.ID, requiredPath, mapping.name)})
		}
	}
	return findings
}

type cliBodyFlagMapping struct {
	target   string
	name     string
	required bool
}

func operationStaticBodyProvidesPath(body map[string]any, path string) bool {
	if len(body) == 0 || strings.TrimSpace(path) == "" {
		return false
	}
	var current any = body
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return false
		}
		current, ok = object[part]
		if !ok {
			return false
		}
	}
	return true
}

func commandBodyFlagCoveringRequiredPath(schema *cliRecordSchemaNode, mappedTargets []cliBodyFlagMapping, requiredPath string) (cliBodyFlagMapping, bool) {
	requiredNode, err := schema.recordPath(requiredPath)
	var optional cliBodyFlagMapping
	optionalFound := false
	for _, mapping := range mappedTargets {
		covers := mapping.target == requiredPath
		if !covers && err == nil && requiredNode != nil && (requiredNode.isObject() || requiredNode.isArray()) && dottedPathPrefix(requiredPath, mapping.target) {
			covers = true
		}
		if !covers {
			continue
		}
		if mapping.required {
			return mapping, true
		}
		if !optionalFound {
			optional = mapping
			optionalFound = true
		}
	}
	return optional, optionalFound
}

func checkCLISurfaceValidationDeclarations(b engine.Bundle, i int, cmd engine.CLICommand) []Finding {
	mappedTargets := map[string]string{}
	var findings []Finding
	for _, flag := range cmd.Flags {
		if strings.TrimSpace(flag.MapsTo) != "" {
			mappedTargets[flag.MapsTo] = flag.Name
		}
		if flag.Format != "" && flag.Format != "date-time" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) flag --%s declares unsupported format %q", i, cmd.Path, flag.Name, flag.Format)})
		}
		if flag.Format != "" && flag.Type != "string" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) flag --%s format validation requires string type", i, cmd.Path, flag.Name)})
		}
		if flag.AllowEmpty != nil && flag.Type != "string" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) flag --%s allow_empty is supported only for string flags", i, cmd.Path, flag.Name)})
		}
		// The meta-schema dialect has no "minimum", so these bounds are checked
		// here instead of being declarable in cli_surface.schema.json.
		if (flag.MaxItems != 0 || flag.MinItems != 0) && flag.Type != "string_array" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) flag --%s max_items/min_items are supported only for string_array flags", i, cmd.Path, flag.Name)})
		}
		if flag.MaxItems < 0 || flag.MinItems < 0 {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) flag --%s max_items/min_items must not be negative", i, cmd.Path, flag.Name)})
		}
		if flag.MaxItems > 0 && flag.MinItems > flag.MaxItems {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) flag --%s min_items %d exceeds max_items %d", i, cmd.Path, flag.Name, flag.MinItems, flag.MaxItems)})
		}
	}
	for j, constraint := range cmd.Constraints {
		if constraint.Kind != "order" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) constraint %d declares unsupported kind %q", i, cmd.Path, j, constraint.Kind)})
		}
		if constraint.Op != "lt" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) constraint %d declares unsupported op %q", i, cmd.Path, j, constraint.Op)})
		}
		if constraint.ValueType != "date-time" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) constraint %d must declare value_type date-time", i, cmd.Path, j)})
		}
		for _, ref := range []struct {
			role   string
			target string
		}{
			{role: "left", target: constraint.Left},
			{role: "right", target: constraint.Right},
		} {
			if err := validateCLIConstraintMappedTarget(ref.target); err != nil {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) constraint %d %s target %q is invalid: %v", i, cmd.Path, j, ref.role, ref.target, err)})
				continue
			}
			if mappedTargets[ref.target] == "" {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) constraint %d %s target %q is not mapped by a command flag", i, cmd.Path, j, ref.role, ref.target)})
			}
		}
		for _, ref := range []struct {
			role   string
			target string
		}{
			{role: "left_fallback", target: constraint.LeftFallback},
			{role: "right_fallback", target: constraint.RightFallback},
		} {
			if ref.target == "" {
				continue
			}
			if err := validateCLIConstraintConfigFallback(ref.target); err != nil {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) constraint %d %s %q is invalid: %v", i, cmd.Path, j, ref.role, ref.target, err)})
			}
		}
		if strings.ContainsAny(constraint.Message, "\r\n") {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("command %d (%q) constraint %d message must be a single line", i, cmd.Path, j)})
		}
	}
	return findings
}

func validateCLIConstraintMappedTarget(target string) error {
	switch {
	case strings.HasPrefix(target, "query."):
		return validateCLIConstraintPath(strings.TrimPrefix(target, "query."))
	case strings.HasPrefix(target, "body."):
		return validateCLIConstraintPath(strings.TrimPrefix(target, "body."))
	default:
		return fmt.Errorf("must use query. or body. target")
	}
}

func validateCLIConstraintConfigFallback(target string) error {
	if !strings.HasPrefix(target, "config.") {
		return fmt.Errorf("must use config. fallback")
	}
	return validateCLIConstraintPath(strings.TrimPrefix(target, "config."))
}

func validateCLIConstraintPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path is required")
	}
	for _, r := range path {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.':
		default:
			return fmt.Errorf("path contains invalid character %q", r)
		}
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path must not contain empty segments")
	}
	return nil
}

func checkCLISurfaceWriteFlags(
	b engine.Bundle,
	i int,
	cmd engine.CLICommand,
	writes map[string]engine.WriteAction,
) []Finding {
	if cmd.Availability != "implemented" || cmd.Intent != "reverse_etl" || cmd.Write == "" {
		return nil
	}

	action, ok := writes[cmd.Write]
	if !ok {
		return nil
	}

	var schema *cliRecordSchemaNode
	if len(action.RecordSchema) > 0 {
		var err error
		schema, err = parseCLIRecordSchema(action.RecordSchema)
		if err != nil {
			return []Finding{{Connector: b.Name, File: "writes.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("write %q has invalid record_schema for CLI validation: %v", action.Name, err)}}
		}
	}

	mapped := map[string]engine.CLIFlag{}
	mappedByFlag := map[string]string{}
	var findings []Finding
	for _, flag := range cmd.Flags {
		target, ok := strings.CutPrefix(flag.MapsTo, "record.")
		if !ok || target == "" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented reverse ETL command %d (%q) flag --%s maps to unsupported target %q", i, cmd.Path, flag.Name, flag.MapsTo)})
			continue
		}
		if prior := mappedByFlag[target]; prior != "" {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented reverse ETL command %d (%q) flags --%s and --%s both map to record.%s", i, cmd.Path, prior, flag.Name, target)})
			continue
		}
		for existing, prior := range mappedByFlag {
			if dottedPathPrefix(existing, target) || dottedPathPrefix(target, existing) {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented reverse ETL command %d (%q) flags --%s and --%s have conflicting record mappings", i, cmd.Path, prior, flag.Name)})
			}
		}
		mappedByFlag[target] = flag.Name
		mapped[target] = flag
		if schema == nil {
			continue
		}
		leaf, err := schema.recordPath(target)
		if err != nil {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented reverse ETL command %d (%q) flag --%s maps outside write %q schema: %v", i, cmd.Path, flag.Name, cmd.Write, err)})
			continue
		}
		if flag.Type == "json" {
			// Runtime preflight delegates this exact check to the engine owning
			// the write's raw record schema. Keep the static authoring gate on
			// that shared rule: a hand-copied object/array test here would drift
			// as soon as the declarative schema contract changes.
			if err := engine.ValidateStructuredJSONRecordField(action.RecordSchema, target); err != nil {
				findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented reverse ETL command %d (%q) structured JSON flag --%s is not declared safely: %v", i, cmd.Path, flag.Name, err)})
			}
			continue
		}
		if !cliFlagTypeMatchesSchema(flag.Type, leaf) {
			findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceSafety, Message: fmt.Sprintf("implemented reverse ETL command %d (%q) flag --%s type %q is incompatible with record.%s schema type %s", i, cmd.Path, flag.Name, flag.Type, target, strings.Join(leaf.effectiveTypes(), ","))})
		}
	}

	required := map[string]bool{}
	if schema != nil {
		for _, field := range schema.requiredMappingPaths("") {
			required[field] = true
		}
	}
	for _, field := range action.PathFields {
		required[field] = true
	}
	if len(required) == 0 {
		return findings
	}

	missing := make([]string, 0, len(required))
	for field := range required {
		if !mappedRecordPathSatisfies(mapped, field) {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		findings = append(findings, Finding{Connector: b.Name, File: "cli_surface.json", Rule: ruleCLISurfaceMissingMapping, Message: fmt.Sprintf("implemented reverse ETL command %d (%q) for write %q lacks flag mappings for required record fields: %s", i, cmd.Path, cmd.Write, strings.Join(missing, ", "))})
	}
	return findings
}

func mappedRecordPathSatisfies(mapped map[string]engine.CLIFlag, required string) bool {
	for target, flag := range mapped {
		// A scalar/nested mapping satisfies its exact declared leaf (and the
		// long-standing child-field case that constructs a required container).
		// A structured JSON mapping has one additional, narrowly preflighted
		// meaning: its top-level object/array value supplies every required
		// descendant of that same declared container. Treating it as only an
		// exact leaf is what made a runtime-valid `--payload` JSON flag fail the
		// static command-surface validator.
		if target == required || dottedPathPrefix(required, target) ||
			(flag.Type == "json" && dottedPathPrefix(target, required)) {
			return true
		}
	}
	return false
}

func dottedPathPrefix(parent, child string) bool {
	return strings.HasPrefix(child, parent+".")
}

type cliRecordSchemaNode struct {
	types                []string
	required             []string
	properties           map[string]*cliRecordSchemaNode
	items                *cliRecordSchemaNode
	additionalProperties bool
	hasAdditionalProps   bool
}

func parseCLIRecordSchema(raw json.RawMessage) (*cliRecordSchemaNode, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return parseCLIRecordSchemaNode(m)
}

func parseCLIRecordSchemaNode(m map[string]json.RawMessage) (*cliRecordSchemaNode, error) {
	n := &cliRecordSchemaNode{additionalProperties: true}
	if raw, ok := m["type"]; ok {
		var single string
		if err := json.Unmarshal(raw, &single); err == nil {
			n.types = []string{single}
		} else if err := json.Unmarshal(raw, &n.types); err != nil {
			return nil, err
		}
	}
	if raw, ok := m["required"]; ok {
		if err := json.Unmarshal(raw, &n.required); err != nil {
			return nil, err
		}
	}
	if raw, ok := m["additionalProperties"]; ok {
		if err := json.Unmarshal(raw, &n.additionalProperties); err != nil {
			return nil, err
		}
		n.hasAdditionalProps = true
	}
	if raw, ok := m["properties"]; ok {
		var props map[string]map[string]json.RawMessage
		if err := json.Unmarshal(raw, &props); err != nil {
			return nil, err
		}
		n.properties = make(map[string]*cliRecordSchemaNode, len(props))
		for name, prop := range props {
			child, err := parseCLIRecordSchemaNode(prop)
			if err != nil {
				return nil, err
			}
			n.properties[name] = child
		}
	}
	if raw, ok := m["items"]; ok {
		var item map[string]json.RawMessage
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		child, err := parseCLIRecordSchemaNode(item)
		if err != nil {
			return nil, err
		}
		n.items = child
	}
	return n, nil
}

func (n *cliRecordSchemaNode) recordPath(path string) (*cliRecordSchemaNode, error) {
	parts := strings.Split(path, ".")
	cur := n
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty path segment")
		}
		if cur.isArray() {
			if err := validateCLIRecordArrayIndex(part); err != nil {
				return nil, err
			}
			if cur.items == nil {
				return nil, fmt.Errorf("array segment %q has no item schema", part)
			}
			cur = cur.items
			continue
		}
		if !cur.isObject() {
			return nil, fmt.Errorf("%q descends into non-object schema", part)
		}
		child := cur.properties[part]
		if child == nil {
			return nil, fmt.Errorf("record field %q is not declared", part)
		}
		cur = child
	}
	return cur, nil
}

func validateCLIRecordArrayIndex(part string) error {
	for _, r := range part {
		if r < '0' || r > '9' {
			return fmt.Errorf("array segment %q is not a numeric index", part)
		}
	}
	if len(part) > 1 && strings.HasPrefix(part, "0") {
		return fmt.Errorf("array index %q must not have leading zeroes", part)
	}
	if len(part) > 3 || (len(part) == 3 && part > "128") {
		return fmt.Errorf("array index %q exceeds max %d", part, maxCLIRecordPathArrayIndex)
	}
	return nil
}

func (n *cliRecordSchemaNode) requiredMappingPaths(prefix string) []string {
	if n == nil {
		return nil
	}
	var out []string
	for _, req := range n.required {
		child := n.properties[req]
		path := joinSchemaPath(prefix, req)
		childPaths := child.requiredNodeMappingPaths(path)
		if len(childPaths) == 0 {
			out = append(out, path)
			continue
		}
		out = append(out, childPaths...)
	}
	return out
}

func (n *cliRecordSchemaNode) requiredNodeMappingPaths(prefix string) []string {
	if n == nil {
		return nil
	}
	if n.isArray() {
		if n.items == nil {
			return nil
		}
		itemPrefix := joinSchemaPath(prefix, "0")
		paths := n.items.requiredNodeMappingPaths(itemPrefix)
		if len(paths) == 0 {
			return []string{prefix}
		}
		return paths
	}
	if n.isObject() {
		return n.requiredMappingPaths(prefix)
	}
	return nil
}

func joinSchemaPath(prefix, part string) string {
	if prefix == "" {
		return part
	}
	return prefix + "." + part
}

func (n *cliRecordSchemaNode) isArray() bool {
	for _, typ := range n.types {
		if typ == "array" {
			return true
		}
	}
	return false
}

func (n *cliRecordSchemaNode) isObject() bool {
	if len(n.properties) > 0 {
		return true
	}
	for _, typ := range n.types {
		if typ == "object" {
			return true
		}
	}
	return len(n.types) == 0
}

func (n *cliRecordSchemaNode) effectiveTypes() []string {
	if len(n.types) > 0 {
		return n.types
	}
	if n.isArray() {
		return []string{"array"}
	}
	if n.isObject() {
		return []string{"object"}
	}
	return []string{"any"}
}

func cliFlagTypeMatchesSchema(flagType string, node *cliRecordSchemaNode) bool {
	schemaTypes := map[string]bool{}
	for _, typ := range node.effectiveTypes() {
		schemaTypes[typ] = true
	}
	switch flagType {
	case "", "string", "enum":
		return schemaTypes["string"] || schemaTypes["any"]
	case "integer":
		return schemaTypes["integer"] || schemaTypes["number"] || schemaTypes["any"]
	case "number":
		return schemaTypes["number"] || schemaTypes["any"]
	case "boolean":
		return schemaTypes["boolean"] || schemaTypes["any"]
	case "string_array":
		return schemaTypes["array"] || schemaTypes["any"]
	default:
		return false
	}
}

// directReadMethodRequirement names the methods a direct read command may
// reference, matching commandrunner: operation-backed commands may POST for
// bounded read-queries, endpoint-backed commands are GET-only.
func directReadMethodRequirement(cmd engine.CLICommand) string {
	if cmd.Operation != "" {
		return "GET or POST"
	}
	return "GET"
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
		// Operation-backed commands are NOT exempt. The runtime
		// (commandrunner.validateOperationDirectReadCommand) enforces exactly
		// these rules on them, so exempting them here is what let 174 commands
		// ship as "implemented" while blocking on every invocation.
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
			// Operation-backed direct reads may use POST for bounded
			// read-queries; endpoint-backed ones stay GET-only. This mirrors
			// commandrunner exactly.
			method := strings.ToUpper(strings.TrimSpace(ep.Method))
			if method != "GET" && (cmd.Operation == "" || method != "POST") {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented direct read command %d (%q) must reference a %s api_surface endpoint, got %s", i, cmd.Path, directReadMethodRequirement(cmd), method),
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
			if repositoryDirectReadOutputPolicies[cmd.OutputPolicy] && !endpointPathHasVariable(ep.Path, "path") {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented direct read command %d (%q) uses repository output policy %q but endpoint path lacks {path}", i, cmd.Path, cmd.OutputPolicy),
				})
			}
		}
		// Asserted unconditionally: an empty output_policy is the finding, not
		// a reason to skip. The runtime rejects both empty and unsupported.
		if !directReadOutputPolicies[cmd.OutputPolicy] {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented direct read command %d (%q) must declare a supported output_policy", i, cmd.Path),
			})
		}
		if cmd.Operation == "" && operationOnlyDirectReadOutputPolicies[cmd.OutputPolicy] {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented direct read command %d (%q) output_policy %q requires an operation", i, cmd.Path, cmd.OutputPolicy),
			})
		}
		if len(findings) > 0 {
			return findings
		}
	case "direct_write":
		// Mirrors commandrunner.validateOperationDirectWriteCommand. Direct
		// execution still remains blocked there; satisfying this contract only
		// makes the command eligible for App's plan -> preview -> approval
		// lifecycle.
		var findings []Finding
		if cmd.Operation == "" {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceMissingMapping,
				Message:   fmt.Sprintf("implemented direct write command %d (%q) must reference an operation", i, cmd.Path),
			})
		}
		if len(cmd.APISurface) != 1 {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceMissingMapping,
				Message:   fmt.Sprintf("implemented direct write command %d (%q) must reference exactly one api_surface endpoint", i, cmd.Path),
			})
		}
		for _, ep := range cmd.APISurface {
			method := strings.ToUpper(strings.TrimSpace(ep.Method))
			if !mutationMethods[method] {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented direct write command %d (%q) must reference a POST, PUT, PATCH, or DELETE api_surface endpoint, got %s", i, cmd.Path, method),
				})
			}
			if isAbsoluteHTTPURL(ep.Path) {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented direct write command %d (%q) must reference a connector-relative api_surface endpoint", i, cmd.Path),
				})
			}
		}
		if !directWriteOutputPolicies[cmd.OutputPolicy] {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("implemented direct write command %d (%q) must declare a supported output_policy", i, cmd.Path),
			})
		}
		if len(findings) > 0 {
			return findings
		}
	case "binary_download":
		// Mirrors commandrunner.validateBinaryDownloadCommand. No output_policy
		// applies: the response becomes a file, not a JSON body.
		var findings []Finding
		if cmd.Operation == "" {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceMissingMapping,
				Message:   fmt.Sprintf("implemented binary download command %d (%q) must reference an operation", i, cmd.Path),
			})
		}
		if len(cmd.APISurface) != 1 {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceMissingMapping,
				Message:   fmt.Sprintf("implemented binary download command %d (%q) must reference exactly one api_surface endpoint", i, cmd.Path),
			})
		}
		for _, ep := range cmd.APISurface {
			if strings.ToUpper(strings.TrimSpace(ep.Method)) != "GET" {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented binary download command %d (%q) must reference a GET api_surface endpoint, got %s", i, cmd.Path, strings.ToUpper(ep.Method)),
				})
			}
			if isAbsoluteHTTPURL(ep.Path) {
				findings = append(findings, Finding{
					Connector: b.Name,
					File:      "cli_surface.json",
					Rule:      ruleCLISurfaceSafety,
					Message:   fmt.Sprintf("implemented binary download command %d (%q) must reference a connector-relative api_surface endpoint", i, cmd.Path),
				})
			}
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
	if (cmd.Availability != "implemented" && cmd.Availability != "partial") ||
		(cmd.Intent != "reverse_etl" && cmd.Intent != "direct_write") {
		return nil
	}
	label := "reverse ETL"
	if cmd.Intent == "direct_write" {
		label = "direct write"
	}

	var findings []Finding
	if strings.TrimSpace(cmd.Risk) == "" {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("%s command %d (%q) must declare risk text", label, i, cmd.Path),
		})
	}
	if strings.TrimSpace(cmd.Approval) == "" {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("%s command %d (%q) must declare approval text", label, i, cmd.Path),
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
		if cmd.Operation != "" && slices.Contains(state.coveredBy.OperationTargets(), cmd.Operation) {
			continue
		}
		if state.excluded || state.operation != nil || state.coveredBy == nil || (state.coveredBy.Stream == "" && len(state.coveredBy.WriteTargets()) == 0) {
			if cmd.Operation != "" && state.operation != nil {
				continue
			}
			// binary_download shares direct_read's coverage bookkeeping: an
			// api_surface row records the command that consumes the endpoint,
			// and which executor runs it does not change who covers it.
			if (cmd.Intent == "direct_read" || cmd.Intent == "binary_download") &&
				directReadCoverageMatches(state.coveredBy, cmd.Path) {
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
		// The endpoint may back several write actions; the command has to be
		// one of them, not necessarily the first one listed.
		if cmd.Write != "" && !slices.Contains(state.coveredBy.WriteTargets(), cmd.Write) {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "cli_surface.json",
				Rule:      ruleCLISurfaceSafety,
				Message:   fmt.Sprintf("command %d (%q) references api_surface endpoint %s %s covered by write %v, want %q", i, cmd.Path, strings.ToUpper(ep.Method), ep.Path, state.coveredBy.WriteTargets(), cmd.Write),
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

func endpointPathHasVariable(path, name string) bool {
	return strings.Contains(path, "{"+name+"}")
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

func checkCertificationSecrets(b engine.Bundle) []Finding {
	if len(b.RawCertification) == 0 || !secretLiteralPattern.Match(b.RawCertification) {
		return nil
	}
	return []Finding{{
		Connector: b.Name,
		File:      "certification.json",
		Rule:      ruleSecretLiteral,
		Message:   "certification.json contains a secret-shaped literal",
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

func checkIncrementalPolicies(b engine.Bundle) []Finding {
	var findings []Finding
	for _, stream := range b.Streams {
		if stream.Incremental == nil {
			continue
		}
		format := strings.TrimSpace(stream.Incremental.ParamFormat)
		if !supportedParamFormats[format] || format != stream.Incremental.ParamFormat {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "streams.json",
				Rule:      ruleIncrementalPolicy,
				Message:   fmt.Sprintf("stream %q declares unsupported incremental.param_format %q", stream.Name, stream.Incremental.ParamFormat),
			})
		}
		prefix := strings.TrimSpace(stream.Incremental.OperatorPrefix)
		if !allowedOperatorPrefixes[prefix] || prefix != stream.Incremental.OperatorPrefix {
			findings = append(findings, Finding{
				Connector: b.Name,
				File:      "streams.json",
				Rule:      ruleIncrementalPolicy,
				Message:   fmt.Sprintf("stream %q declares unsupported incremental.operator_prefix %q", stream.Name, stream.Incremental.OperatorPrefix),
			})
		}
	}
	return findings
}

// checkIncrementalStartDateFormat is N2's narrow, honest WARNING (wave0
// REVIEW.md carried flag; SPEC.md §4 "promote to a validate-time guard"):
// for every stream whose incremental.param_format parses timestamp input
// through engine/read.go's parseLowerBoundTime and which names a
// start_config_key, check whether that spec.json property declares a
// date-ish JSON Schema
// "format" (date-time/date). If it does not, a digit-shaped config value —
// e.g. an operator typo like "20260101" (yyyymmdd), which is NOT Unix
// seconds — would silently be treated as one instead of erroring, producing
// a bogus 1970s-era lower bound. This is deliberately scoped to ONLY these
// timestamp-parsing formats: unix_seconds is excluded because there an
// all-digits value IS the correct, intended shape (no misinterpretation risk
// at all), and rfc3339 never attempts digit parsing in the first place
// (verbatim passthrough). Reads spec.json's per-property "format" directly from
// b.RawSpec (F5, REVIEW.md) since the compiled *engine.Schema does not expose
// annotation keywords like "format" through any accessor. Schema validation
// enforces format:uri, but date and date-time remain annotations, so this
// validator still has to inspect the raw property declaration.
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

// checkCLISurfaceBinaryOperationSafety enforces, against the operation a
// binary_download command references, exactly what engine.OperationBinaryDownload
// enforces at execution time.
//
// extract_archives is checked here rather than only at execution because a
// command backed by an extracting operation can never succeed: the executor
// refuses it outright, since archive extraction is zip-slip and
// decompression-bomb territory and is a separate capability, never a flag.
func checkCLISurfaceBinaryOperationSafety(b engine.Bundle, i int, cmd engine.CLICommand, op engine.OperationSpec) []Finding {
	var findings []Finding
	if op.Kind != "binary_download" || op.Binary == nil {
		return append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented binary download command %d (%q) operation %q must be binary_download", i, cmd.Path, cmd.Operation),
		})
	}
	if method := strings.ToUpper(strings.TrimSpace(op.Binary.Method)); method != "GET" {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented binary download command %d (%q) operation %q must use GET, got %s", i, cmd.Path, cmd.Operation, method),
		})
	}
	if isAbsoluteHTTPURL(op.Binary.Path) {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented binary download command %d (%q) operation %q must use connector-relative path", i, cmd.Path, cmd.Operation),
		})
	}
	if op.Binary.MaxBytes <= 0 {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented binary download command %d (%q) operation %q must declare positive binary.max_bytes", i, cmd.Path, cmd.Operation),
		})
	}
	if op.Binary.ExtractArchives {
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented binary download command %d (%q) operation %q declares extract_archives, which the executor refuses; archive extraction is a separate capability", i, cmd.Path, cmd.Operation),
		})
	}
	for _, flag := range cmd.Flags {
		mapsTo := strings.TrimSpace(flag.MapsTo)
		if strings.HasPrefix(mapsTo, "path.") || strings.HasPrefix(mapsTo, "query.") {
			continue
		}
		findings = append(findings, Finding{
			Connector: b.Name,
			File:      "cli_surface.json",
			Rule:      ruleCLISurfaceSafety,
			Message:   fmt.Sprintf("implemented binary download command %d (%q) flag --%s maps to unsupported target %q", i, cmd.Path, flag.Name, flag.MapsTo),
		})
	}
	return findings
}
