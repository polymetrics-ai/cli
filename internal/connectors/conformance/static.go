package conformance

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

// staticCheckNames is the fixed, ordered design §E.2 static check list.
var staticCheckNames = []string{
	"spec_schema_valid",
	"stream_schemas_valid",
	"pk_fields_exist",
	"cursor_fields_exist",
	"interpolations_resolve",
	"write_schemas_valid",
	"surface_complete",
	"docs_present",
	"secret_redaction",
	"fixtures_present",
}

// runStaticChecks runs every static check against an already-loaded Bundle
// (a bundle that reached here already passed engine.Load's own structural +
// meta-schema validation, which is what backs spec_schema_valid and
// stream_schemas_valid when they pass — see classifyLoadError for the
// failure path when Load itself errors).
func runStaticChecks(b engine.Bundle) []CheckResult {
	var checks []CheckResult
	checks = addCheck(checks, "spec_schema_valid", checkSpecSchemaValid(b))
	checks = addCheck(checks, "stream_schemas_valid", checkStreamSchemasValid(b))
	checks = addCheck(checks, "pk_fields_exist", checkPKFieldsExist(b))
	checks = addCheck(checks, "cursor_fields_exist", checkCursorFieldsExist(b))
	checks = addCheck(checks, "interpolations_resolve", checkInterpolationsResolve(b))
	checks = addCheck(checks, "write_schemas_valid", checkWriteSchemasValid(b))
	checks = addCheck(checks, "surface_complete", checkSurfaceComplete(b))
	checks = addCheck(checks, "docs_present", checkDocsPresent(b))
	checks = addCheck(checks, "secret_redaction", checkSecretRedaction(b))
	checks = addCheck(checks, "fixtures_present", checkFixturesPresent(b))
	return checks
}

// requiredDocHeadings are the fixed docs.md headings (design §F.6).
var requiredDocHeadings = []string{
	"Overview",
	"Auth setup",
	"Streams notes",
	"Write actions & risks",
	"Known limits",
}

// surfaceCategories is the closed exclusion vocabulary (design §E.1 rule 3).
// Defense-in-depth: the engine loader's meta-schema already enforces this at
// Load time via an enum on api_surface.schema.json.
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

// secretLiteralPattern flags secret-shaped literals accidentally committed
// to fixtures or docs: a Bearer-scheme header value, a long opaque token
// following an auth-flavored key, or a recognizable vendor secret prefix
// (e.g. Stripe's sk_live_/sk_test_). Fixtures/docs must only ever carry
// synthetic data (THREAT-MODEL §4).
var secretLiteralPattern = regexp.MustCompile(`(?i)(bearer\s+[a-z0-9_\-\.]{16,}|(api[_-]?key|access[_-]?token|secret|password)["' ]*[:=]\s*["']?[a-z0-9_\-\.]{16,}|\bsk_(live|test)_[a-z0-9]{10,}\b)`)

// checkSpecSchemaValid reports whether the bundle's compiled spec.json is
// present. A bundle reaching runStaticChecks already had spec.json validated
// against the meta-schema AND compiled by engine.Load, so this check passes
// unconditionally for any successfully loaded bundle — its failing case is
// exercised via ReportFromLoadError/classifyLoadError when Load itself
// rejects spec.json (see TestReportFromLoadError_ClassifiesMetaSchemaFailure).
func checkSpecSchemaValid(b engine.Bundle) error {
	if b.Spec == nil {
		return fmt.Errorf("spec.json did not compile to a schema")
	}
	return nil
}

// checkStreamSchemasValid mirrors checkSpecSchemaValid for stream schemas:
// engine.Load already compiled every stream's schema file before this check
// runs; its failing case is exercised via ReportFromLoadError when Load
// itself rejects a stream schema file.
func checkStreamSchemasValid(b engine.Bundle) error {
	for _, s := range b.Streams {
		if _, ok := b.Schemas[s.Name]; !ok {
			return fmt.Errorf("stream %q has no compiled schema", s.Name)
		}
	}
	return nil
}

// checkPKFieldsExist enforces that every x-primary-key field a stream's
// schema declares actually exists among that schema's own properties.
func checkPKFieldsExist(b engine.Bundle) error {
	for _, s := range b.Streams {
		sch, ok := b.Schemas[s.Name]
		if !ok {
			continue
		}
		props := propertySet(sch.Properties())
		for _, pk := range sch.PrimaryKey {
			if !props[pk] {
				return fmt.Errorf("stream %q x-primary-key field %q not found in schema %q properties", s.Name, pk, s.SchemaRef)
			}
		}
	}
	return nil
}

// checkCursorFieldsExist enforces that every stream's incremental.cursor_field
// (streams.json) actually exists among that stream's schema properties.
func checkCursorFieldsExist(b engine.Bundle) error {
	for _, s := range b.Streams {
		if s.Incremental == nil || s.Incremental.CursorField == "" {
			continue
		}
		sch, ok := b.Schemas[s.Name]
		if !ok {
			continue
		}
		props := propertySet(sch.Properties())
		if !props[s.Incremental.CursorField] {
			return fmt.Errorf("stream %q incremental.cursor_field %q not found in schema %q properties", s.Name, s.Incremental.CursorField, s.SchemaRef)
		}
	}
	return nil
}

// checkInterpolationsResolve statically resolves every {{ }} template found
// in the bundle's streams.json/writes.json against spec.json's declared
// property set, via engine.ResolveCheck. AuthSpec's `when` field is checked
// via engine.ResolveCheckWhen instead (S3 engine mini-wave item 2): a `when`
// clause is the one field here that legitimately uses the `==`/`in`/
// truthiness when-grammar rather than plain `{{ }}` interpolation, and
// ResolveCheck's bare-namespace.key-only parsing would otherwise hard-fail
// any `==`/`in`-shaped when clause even when its referenced key IS
// spec-declared (github's auth_type in [...] restoration, ledger G14).
func checkInterpolationsResolve(b engine.Bundle) error {
	specKeys := map[string]bool{}
	if b.Spec != nil {
		for _, k := range b.Spec.Properties() {
			specKeys[k] = true
		}
	}

	check := func(template string) error {
		if template == "" {
			return nil
		}
		return engine.ResolveCheck(template, specKeys)
	}
	checkWhen := func(template string) error {
		if template == "" {
			return nil
		}
		return engine.ResolveCheckWhen(template, specKeys)
	}

	if err := check(b.HTTP.URL); err != nil {
		return err
	}
	for _, h := range b.HTTP.Headers {
		if err := check(h); err != nil {
			return err
		}
	}
	for _, a := range b.HTTP.Auth {
		if err := check(a.Token); err != nil {
			return err
		}
		if err := check(a.Value); err != nil {
			return err
		}
		if err := checkWhen(a.When); err != nil {
			return err
		}
	}
	for _, s := range b.Streams {
		if err := check(s.Path); err != nil {
			return fmt.Errorf("stream %q: %w", s.Name, err)
		}
		for _, v := range s.Query {
			// engine.StreamSpec.Query values are engine.QueryParam (gap-loop
			// item 3: optional-query dialect) as of this change, not bare
			// strings; check() only ever needed the template text, which is
			// QueryParam.Template regardless of whether the JSON source was a
			// plain string or the new {template, omit_when_absent, default}
			// object form.
			if err := check(v.Template); err != nil {
				return fmt.Errorf("stream %q: %w", s.Name, err)
			}
		}
		if s.GraphQL != nil {
			if err := engine.ResolveCheckGraphQLVariables(s.GraphQL.Variables, specKeys); err != nil {
				return fmt.Errorf("stream %q: %w", s.Name, err)
			}
		}
		for _, v := range s.ComputedFields {
			if err := check(v); err != nil {
				return fmt.Errorf("stream %q: %w", s.Name, err)
			}
		}
	}
	for _, w := range b.Writes {
		if err := check(w.Path); err != nil {
			return fmt.Errorf("write action %q: %w", w.Name, err)
		}
		if w.GraphQL != nil {
			if err := engine.ResolveCheckGraphQLVariables(w.GraphQL.Variables, specKeys); err != nil {
				return fmt.Errorf("write action %q: %w", w.Name, err)
			}
		}
	}
	return nil
}

// checkWriteSchemasValid compiles every write action's record_schema. A
// write action that never declares one (hook-driven body) is skipped.
func checkWriteSchemasValid(b engine.Bundle) error {
	for _, w := range b.Writes {
		if len(w.RecordSchema) == 0 {
			continue
		}
		if _, err := engine.CompileSchema(w.RecordSchema); err != nil {
			return fmt.Errorf("write action %q: record_schema: %w", w.Name, err)
		}
	}
	return nil
}

// checkSurfaceComplete enforces design §E.1 rules 1-4: every endpoint has
// exactly one executable covered_by row or an explicit blocked non-executable
// classifier; covered_by resolves to a declared stream/action/direct-read
// command and every declared stream/action appears in the surface;
// excluded/operation classifiers use closed vocabularies; capabilities.write/read
// == false is only legal when the surface has no executable mutation/GET
// endpoint respectively.
func checkSurfaceComplete(b engine.Bundle) error {
	if b.Surface == nil {
		return fmt.Errorf("api_surface.json did not load")
	}

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
			return fmt.Errorf("endpoint %d (%s %s) uses legacy excluded in operation_ledger_version mode", i, ep.Method, ep.Path)
		}
		if !ledgerMode && hasOperation {
			return fmt.Errorf("endpoint %d (%s %s) uses operation without operation_ledger_version", i, ep.Method, ep.Path)
		}

		switch {
		case hasCovered && (hasExcluded || hasOperation):
			return fmt.Errorf("endpoint %d (%s %s) has covered_by plus another classifier", i, ep.Method, ep.Path)
		case !hasCovered && !hasExcluded && !hasOperation:
			return fmt.Errorf("endpoint %d (%s %s) has no classifier", i, ep.Method, ep.Path)
		case ledgerMode && hasOperation && hasExcluded:
			return fmt.Errorf("endpoint %d (%s %s) has both operation and excluded", i, ep.Method, ep.Path)
		case hasCovered:
			if ep.CoveredBy.Stream != "" {
				if !streams[ep.CoveredBy.Stream] {
					return fmt.Errorf("endpoint %d (%s %s) covered_by.stream %q is not a declared stream", i, ep.Method, ep.Path, ep.CoveredBy.Stream)
				}
				coveredStreams[ep.CoveredBy.Stream] = true
			}
			if ep.CoveredBy.Write != "" {
				if !writes[ep.CoveredBy.Write] {
					return fmt.Errorf("endpoint %d (%s %s) covered_by.write %q is not a declared write action", i, ep.Method, ep.Path, ep.CoveredBy.Write)
				}
				coveredWrites[ep.CoveredBy.Write] = true
			}
			for _, directRead := range coveredDirectReadTargets(ep.CoveredBy) {
				if !directReads[directRead] {
					return fmt.Errorf("endpoint %d (%s %s) covered_by.direct_read %q is not an implemented direct_read command", i, ep.Method, ep.Path, directRead)
				}
				if !strings.EqualFold(ep.Method, "GET") {
					return fmt.Errorf("endpoint %d (%s %s) covered_by.direct_read must use GET", i, ep.Method, ep.Path)
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
				return fmt.Errorf("endpoint %d (%s %s) excluded.category %q is not in the closed vocabulary", i, ep.Method, ep.Path, ep.Excluded.Category)
			}
		case hasOperation:
			if err := checkSurfaceOperation(i, ep); err != nil {
				return err
			}
		}
	}

	for name := range streams {
		if !coveredStreams[name] {
			return fmt.Errorf("stream %q has no covered_by entry in api_surface.json", name)
		}
	}
	for name := range writes {
		if !coveredWrites[name] {
			return fmt.Errorf("write action %q has no covered_by entry in api_surface.json", name)
		}
	}

	if !b.Metadata.Capabilities.Write && hasNonExcludedMutation {
		return fmt.Errorf("capabilities.write is false but api_surface.json has a non-excluded POST/PUT/PATCH/DELETE endpoint")
	}
	if !b.Metadata.Capabilities.Read && hasNonExcludedGET {
		return fmt.Errorf("capabilities.read is false but api_surface.json has a non-excluded GET endpoint")
	}

	return nil
}

func checkSurfaceOperation(i int, ep engine.SurfaceEndpoint) error {
	op := ep.Operation
	if op == nil {
		return nil
	}
	prefix := fmt.Sprintf("endpoint %d (%s %s)", i, ep.Method, ep.Path)
	if !surfaceOperationModels[op.Model] {
		return fmt.Errorf("%s operation.model %q is not in the closed vocabulary", prefix, op.Model)
	}
	if !surfaceOperationStatuses[op.Status] {
		return fmt.Errorf("%s operation.status %q is not in the closed vocabulary", prefix, op.Status)
	}
	if !surfaceOperationRisks[op.Risk] {
		return fmt.Errorf("%s operation.risk %q is not in the closed vocabulary", prefix, op.Risk)
	}
	if !op.BlockedByDefault {
		return fmt.Errorf("%s operation.blocked_by_default must be true", prefix)
	}
	if strings.TrimSpace(op.Reason) == "" {
		return fmt.Errorf("%s operation.reason is required", prefix)
	}
	if op.Model == "duplicate" && strings.TrimSpace(op.DuplicateOf) == "" {
		return fmt.Errorf("%s operation.duplicate_of is required for duplicate rows", prefix)
	}
	if sourceRequiredOperationModels[op.Model] &&
		strings.TrimSpace(op.SourceURL) == "" &&
		strings.TrimSpace(op.Notes) == "" {
		return fmt.Errorf("%s operation.source_url or operation.notes is required for sensitive/admin/destructive/disallowed rows", prefix)
	}
	return nil
}

// checkDocsPresent enforces the fixed docs.md heading set (design §F.6).
// Headings are matched as Markdown "#"/"##" lines by exact (trimmed) text.
func checkDocsPresent(b engine.Bundle) error {
	present := map[string]bool{}
	for _, line := range strings.Split(b.Docs, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimLeft(trimmed, "#")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed != "" {
			present[trimmed] = true
		}
	}
	for _, h := range requiredDocHeadings {
		if !present[h] {
			return fmt.Errorf("docs.md missing required heading %q", h)
		}
	}
	return nil
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

// checkSecretRedaction scans docs.md and every fixture file's raw bytes for
// secret-shaped literals. Fixtures/docs must only ever contain synthetic
// data (THREAT-MODEL §4); a planted real-looking token is a hard failure.
func checkSecretRedaction(b engine.Bundle) error {
	if secretLiteralPattern.MatchString(b.Docs) {
		return fmt.Errorf("docs.md contains a secret-shaped literal")
	}
	if b.Fixtures == nil {
		return nil
	}
	var found string
	_ = fs.WalkDir(b.Fixtures, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || found != "" {
			return nil
		}
		raw, ferr := fs.ReadFile(b.Fixtures, path)
		if ferr != nil {
			return nil
		}
		if secretLiteralPattern.Match(raw) {
			found = path
		}
		return nil
	})
	if found != "" {
		return fmt.Errorf("fixtures/%s contains a secret-shaped literal", found)
	}
	return nil
}

// checkFixturesPresent enforces that the bundle's FIRST declared stream has
// at least one fixture page (design §E.2: "first stream mandatory"). A
// bundle with zero streams (e.g. dynamic_schema Tier-3 natives) trivially
// passes — there is no "first stream" to require fixtures for.
func checkFixturesPresent(b engine.Bundle) error {
	if len(b.Streams) == 0 {
		return nil
	}
	first := b.Streams[0]
	if b.Fixtures == nil {
		return fmt.Errorf("bundle declares stream %q but has no fixtures/ directory at all", first.Name)
	}
	pages, err := loadFixturePages(b.Fixtures, first.Name)
	if err != nil {
		return fmt.Errorf("stream %q: %w", first.Name, err)
	}
	if len(pages) == 0 {
		return fmt.Errorf("stream %q (first stream) has zero fixture pages under fixtures/streams/%s/", first.Name, first.Name)
	}
	return nil
}

func propertySet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, n := range names {
		out[n] = true
	}
	return out
}

// classifyLoadError maps an engine.Load error message to the specific
// static check name it corresponds to, mirroring
// cmd/connectorgen/validate.go's loadErrorFinding classification (kept as
// this package's own independent copy: PLAN.md's corpus-split note says
// B-11's and B-13's seeded corpora are not cross-package-shared, and the
// classification logic travels with its own corpus for the same reason —
// this package must be self-contained).
func classifyLoadError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "missing required file streams.json"):
		return "stream_schemas_valid"
	// loadStreamSchemas' error shape: "...: stream X: schema Y: ...", for
	// both a missing schema file AND a schema compile failure (e.g. an
	// unknown keyword) — both are stream-schema defects.
	case strings.Contains(msg, ": stream ") && strings.Contains(msg, ": schema "):
		return "stream_schemas_valid"
	case strings.Contains(msg, "spec.json"):
		return "spec_schema_valid"
	case strings.Contains(msg, "streams.json"):
		return "stream_schemas_valid"
	case strings.Contains(msg, "writes.json"):
		return "write_schemas_valid"
	case strings.Contains(msg, "api_surface.json"):
		return "surface_complete"
	case strings.Contains(msg, "docs.md"):
		return "docs_present"
	case strings.Contains(msg, "metadata.json"), strings.Contains(msg, "missing required file"), strings.Contains(msg, "does not match"):
		return "spec_schema_valid"
	default:
		return "spec_schema_valid"
	}
}
