package boundary

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// OwnershipScopeKind is the machine-readable contract kind for connector
	// implementation PR scope files.
	OwnershipScopeKind = "ConnectorImplementationScope"
	// OwnershipReportKind is the machine-readable JSON kind for ownership reports.
	OwnershipReportKind = "ConnectorOwnershipReport"
)

const (
	// RuleOwnershipScopeMissing means changed paths require one connector target but
	// neither a scope file nor path inference produced exactly one connector.
	RuleOwnershipScopeMissing = "ownership_scope_missing"
	// RuleOwnershipScopeMismatch means a declared target did not match inferred connector paths.
	RuleOwnershipScopeMismatch = "ownership_scope_mismatch"
	// RuleOwnershipSharedPath rejects shared runtime, docs, tooling, or repo files in a connector lane.
	RuleOwnershipSharedPath = "ownership_shared_path"
	// RuleOwnershipUnrelatedConnector rejects connector-owned implementation paths for another connector.
	RuleOwnershipUnrelatedConnector = "ownership_unrelated_connector"
	// RuleOwnershipUnrelatedGenerated rejects generated docs/website/icon paths for another connector.
	RuleOwnershipUnrelatedGenerated = "ownership_unrelated_generated"
	// RuleOwnershipGateConfigEdit rejects edits to the guardrail implementation or exception/config surfaces.
	RuleOwnershipGateConfigEdit = "ownership_gate_config_edit"
	// RuleOwnershipInvalidPath rejects unsafe changed-path values.
	RuleOwnershipInvalidPath = "ownership_invalid_path"
)

const (
	ownershipClassIgnored              = "ignored"
	ownershipClassPlanningArtifact     = "planning_artifact"
	ownershipClassConnectorDefs        = "connector_defs"
	ownershipClassConnectorHooks       = "connector_hooks"
	ownershipClassConnectorNative      = "connector_native"
	ownershipClassConnectorLegacy      = "connector_legacy"
	ownershipClassConnectorDocs        = "connector_docs"
	ownershipClassConnectorIcon        = "connector_icon"
	ownershipClassConnectorWebsiteIcon = "connector_website_icon"
	ownershipClassConnectorOwnedTest   = "connector_owned_test"
	ownershipClassSharedGenerated      = "shared_generated_index"
	ownershipClassGateConfig           = "gate_config"
	ownershipClassSharedRuntime        = "shared_runtime"
	ownershipClassSharedTooling        = "shared_tooling"
	ownershipClassSharedDocs           = "shared_docs"
	ownershipClassSharedRepo           = "shared_repo"
)

const (
	ownershipDecisionAllowed  = "allowed"
	ownershipDecisionRejected = "rejected"
	ownershipDecisionIgnored  = "ignored"
)

// OwnershipScope is the machine-readable target connector contract consumed by
// connectorgen ownership. Connectors must contain exactly one connector slug.
type OwnershipScope struct {
	APIVersion string   `json:"api_version"`
	Kind       string   `json:"kind"`
	Connectors []string `json:"connectors"`
}

// OwnershipOptions configures changed-path ownership validation.
type OwnershipOptions struct {
	BaseRef      string
	ScopeFile    string
	ChangedPaths []string
	Now          time.Time
}

// OwnershipPath records the classification and decision for one changed path.
type OwnershipPath struct {
	Path      string `json:"path"`
	Class     string `json:"class"`
	Connector string `json:"connector,omitempty"`
	Decision  string `json:"decision"`
	Message   string `json:"message,omitempty"`
}

// OwnershipReport is the deterministic JSON output rendered by connectorgen ownership.
type OwnershipReport struct {
	APIVersion         string          `json:"api_version"`
	Kind               string          `json:"kind"`
	Outcome            string          `json:"outcome"`
	RepoRoot           string          `json:"repo_root"`
	BaseRef            string          `json:"base_ref,omitempty"`
	ScopeFile          string          `json:"scope_file,omitempty"`
	TargetConnector    string          `json:"target_connector,omitempty"`
	InferredConnectors []string        `json:"inferred_connectors"`
	ChangedPaths       []OwnershipPath `json:"changed_paths"`
	Findings           []Finding       `json:"findings"`
	Warnings           []Finding       `json:"warnings"`
}

// ValidateOwnership validates that changed paths belong to exactly one target
// connector implementation lane. It fails closed on shared runtime/tooling,
// unrelated connector paths, unrelated generated connector outputs, and edits
// that could weaken the guardrail itself.
func ValidateOwnership(root string, opts OwnershipOptions) (OwnershipReport, error) {
	absRoot, err := validateRoot(root)
	if err != nil {
		return OwnershipReport{}, err
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}

	lx, err := loadLexicon(absRoot)
	if err != nil {
		return OwnershipReport{}, &ConfigError{Err: err}
	}

	changed, err := ownershipChangedPaths(absRoot, opts)
	if err != nil {
		return OwnershipReport{}, err
	}

	scopeConnector := ""
	scopeFile := ""
	if strings.TrimSpace(opts.ScopeFile) != "" {
		scope, rel, err := loadOwnershipScope(absRoot, opts.ScopeFile, lx)
		if err != nil {
			return OwnershipReport{}, err
		}
		scopeConnector = scope.Connectors[0]
		scopeFile = rel
	}

	rawPaths := make([]ownershipPathClass, 0, len(changed))
	inferredSet := map[string]bool{}
	for _, rel := range changed {
		raw := classifyOwnershipChangedPath(rel, lx)
		rawPaths = append(rawPaths, raw)
		if raw.InferConnector && raw.Connector != "" {
			inferredSet[raw.Connector] = true
		}
	}
	inferred := sortedKeys(inferredSet)

	target := scopeConnector
	var findings []Finding
	if target == "" {
		switch len(inferred) {
		case 1:
			target = inferred[0]
		case 0:
			if hasOwnershipEnforcedChange(rawPaths) {
				findings = append(findings, ownershipFinding(RuleOwnershipScopeMissing, "", "", "", "changed paths require exactly one connector target, but none was declared or inferred", "declare one ConnectorImplementationScope connector or move shared changes to a foundation PR"))
			}
		default:
			findings = append(findings, ownershipFinding(RuleOwnershipScopeMissing, "", "", strings.Join(inferred, ","), fmt.Sprintf("changed paths infer multiple connector targets %v", inferred), "split unrelated connector changes or declare exactly one target and remove unrelated connector paths"))
		}
	}

	changedReport := make([]OwnershipPath, 0, len(rawPaths))
	for _, raw := range rawPaths {
		pathReport, pathFindings := evaluateOwnershipPath(raw, target)
		changedReport = append(changedReport, pathReport)
		findings = append(findings, pathFindings...)
	}

	if scopeConnector != "" {
		for _, inferredConnector := range inferred {
			if inferredConnector != scopeConnector {
				findings = append(findings, ownershipFinding(RuleOwnershipScopeMismatch, inferredConnector, "", inferredConnector, fmt.Sprintf("scope connector %q does not match inferred connector %q", scopeConnector, inferredConnector), "remove unrelated connector paths or update the scope only after paths agree"))
			}
		}
	}

	sortOwnershipPaths(changedReport)
	sortFindings(findings)
	outcome := OutcomeClean
	if len(findings) > 0 {
		outcome = OutcomePolicyViolations
	}

	return OwnershipReport{
		APIVersion:         APIVersion,
		Kind:               OwnershipReportKind,
		Outcome:            outcome,
		RepoRoot:           absRoot,
		BaseRef:            opts.BaseRef,
		ScopeFile:          scopeFile,
		TargetConnector:    target,
		InferredConnectors: nonNilStrings(inferred),
		ChangedPaths:       nonNilOwnershipPaths(changedReport),
		Findings:           nonNilFindings(findings),
		Warnings:           []Finding{},
	}, nil
}

func loadOwnershipScope(root, path string, lx lexicon) (OwnershipScope, string, error) {
	abs, rel, err := resolveRepoRelativePath(root, path)
	if err != nil {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("scope file: %w", err)}
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("read scope file %s: %w", rel, err)}
	}
	var scope OwnershipScope
	if err := json.Unmarshal(b, &scope); err != nil {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("parse scope file %s: %w", rel, err)}
	}
	if strings.TrimSpace(scope.APIVersion) != APIVersion {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("scope file %s: api_version must be %q", rel, APIVersion)}
	}
	if strings.TrimSpace(scope.Kind) != OwnershipScopeKind {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("scope file %s: kind must be %q", rel, OwnershipScopeKind)}
	}
	if len(scope.Connectors) != 1 {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("scope file %s: connectors must contain exactly one connector slug", rel)}
	}
	connector := strings.ToLower(strings.TrimSpace(scope.Connectors[0]))
	if connector == "" {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("scope file %s: connector slug is required", rel)}
	}
	if _, ok := lx.byName[connector]; !ok {
		return OwnershipScope{}, "", &ConfigError{Err: fmt.Errorf("scope file %s: unknown connector slug %q", rel, connector)}
	}
	scope.Connectors[0] = connector
	return scope, rel, nil
}

func ownershipChangedPaths(root string, opts OwnershipOptions) ([]string, error) {
	if len(opts.ChangedPaths) > 0 {
		return normalizeChangedPaths(root, opts.ChangedPaths)
	}
	baseRef := opts.BaseRef
	if baseRef == "" {
		baseRef = "HEAD"
	}
	limit, err := diffLimit(root, baseRef)
	if err != nil {
		return nil, &ConfigError{Err: err}
	}
	return normalizeChangedPaths(root, sortedKeys(limit))
}

func normalizeChangedPaths(root string, paths []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, path := range paths {
		_, rel, err := resolveRepoRelativePath(root, path)
		if err != nil {
			return nil, &ConfigError{Err: fmt.Errorf("changed path %q: %w", path, err)}
		}
		if rel == "" {
			continue
		}
		if !seen[rel] {
			seen[rel] = true
			out = append(out, rel)
		}
	}
	sort.Strings(out)
	return out, nil
}

func resolveRepoRelativePath(root, path string) (string, string, error) {
	if strings.TrimSpace(path) == "" {
		return "", "", fmt.Errorf("path is required")
	}
	if hasControlChar(path) {
		return "", "", fmt.Errorf("path must not contain control characters")
	}
	var abs string
	if filepath.IsAbs(path) {
		abs = filepath.Clean(path)
	} else {
		rel := normalizeRelPath(path)
		if rel == "" || rel == "." {
			return "", "", fmt.Errorf("path is required")
		}
		if rel == ".." || strings.HasPrefix(rel, "../") || strings.Contains(rel, "/../") || strings.HasSuffix(rel, "/..") {
			return "", "", fmt.Errorf("path must stay within the repo")
		}
		abs = filepath.Join(root, filepath.FromSlash(rel))
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", "", fmt.Errorf("resolve path relative to repo: %w", err)
	}
	rel = normalizeRelPath(rel)
	if rel == "" || rel == "." {
		return "", "", fmt.Errorf("path is required")
	}
	if rel == ".." || strings.HasPrefix(rel, "../") || strings.Contains(rel, "/../") || strings.HasSuffix(rel, "/..") || filepath.IsAbs(rel) {
		return "", "", fmt.Errorf("path must stay within the repo")
	}
	return abs, rel, nil
}

func hasControlChar(value string) bool {
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}

type ownershipPathClass struct {
	Path           string
	Class          string
	Connector      string
	InferConnector bool
	Generated      bool
	GateConfig     bool
	Shared         bool
	Ignored        bool
}

func classifyOwnershipChangedPath(rel string, lx lexicon) ownershipPathClass {
	rel = normalizeRelPath(rel)
	if rel == "" {
		return ownershipPathClass{Path: rel, Class: ownershipClassIgnored, Ignored: true}
	}
	if isConnectorPlanningArtifactPath(rel) {
		return ownershipPathClass{Path: rel, Class: ownershipClassPlanningArtifact, Ignored: true}
	}
	if isIgnoredOwnershipPath(rel) {
		return ownershipPathClass{Path: rel, Class: ownershipClassIgnored, Ignored: true}
	}
	if isOwnershipGateConfigPath(rel) {
		return ownershipPathClass{Path: rel, Class: ownershipClassGateConfig, GateConfig: true}
	}
	if isNarrowSharedOwnershipOutput(rel) {
		return ownershipPathClass{Path: rel, Class: ownershipClassSharedGenerated}
	}
	if connector, ok := connectorPathSegment(rel, "internal/connectors/defs/", lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorDefs, Connector: connector, InferConnector: true}
	}
	if connector, ok := connectorPathSegment(rel, "internal/connectors/hooks/", lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorHooks, Connector: connector, InferConnector: true}
	}
	if connector, ok := connectorPathSegment(rel, "internal/connectors/native/", lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorNative, Connector: connector, InferConnector: true}
	}
	if connector, ok := legacyConnectorPathSegment(rel, lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorLegacy, Connector: connector, InferConnector: true}
	}
	if connector, ok := connectorDocsPath(rel, lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorDocs, Connector: connector, InferConnector: true, Generated: true}
	}
	if connector, ok := connectorIconPath(rel, "docs/connectors/icons/", lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorIcon, Connector: connector, InferConnector: true, Generated: true}
	}
	if connector, ok := connectorIconPath(rel, "website/public/connectors/icons/", lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorWebsiteIcon, Connector: connector, InferConnector: true, Generated: true}
	}
	if connector, ok := connectorOwnedSharedPackageTestPath(rel, lx); ok {
		return ownershipPathClass{Path: rel, Class: ownershipClassConnectorOwnedTest, Connector: connector, InferConnector: true}
	}
	if strings.HasPrefix(rel, "cmd/") || strings.HasPrefix(rel, "scripts/") || rel == "Makefile" {
		return ownershipPathClass{Path: rel, Class: ownershipClassSharedTooling, Shared: true}
	}
	if strings.HasPrefix(rel, "internal/") {
		return ownershipPathClass{Path: rel, Class: ownershipClassSharedRuntime, Shared: true}
	}
	if strings.HasPrefix(rel, "docs/") || strings.HasPrefix(rel, "website/") {
		return ownershipPathClass{Path: rel, Class: ownershipClassSharedDocs, Shared: true}
	}
	if strings.HasPrefix(rel, ".planning/") || strings.HasPrefix(rel, dotGHPathPrefix()) || strings.HasPrefix(rel, ".agents/") || rel == "go.mod" || rel == "go.sum" || rel == "README.md" || rel == "AGENTS.md" || rel == "CONTRIBUTING.md" || rel == "CHANGELOG.md" {
		return ownershipPathClass{Path: rel, Class: ownershipClassSharedRepo, Shared: true}
	}
	return ownershipPathClass{Path: rel, Class: ownershipClassIgnored, Ignored: true}
}

func evaluateOwnershipPath(raw ownershipPathClass, target string) (OwnershipPath, []Finding) {
	out := OwnershipPath{Path: raw.Path, Class: raw.Class, Connector: raw.Connector}
	if raw.Ignored {
		out.Decision = ownershipDecisionIgnored
		return out, nil
	}
	if raw.GateConfig {
		out.Decision = ownershipDecisionRejected
		out.Message = "connector implementation lanes cannot edit guardrail tooling, exceptions, or configuration"
		return out, []Finding{ownershipFinding(RuleOwnershipGateConfigEdit, raw.Connector, raw.Path, raw.Path, out.Message, "move guardrail changes to a dedicated foundation PR")}
	}
	if target == "" {
		out.Decision = ownershipDecisionRejected
		out.Message = "no connector target available for ownership validation"
		return out, nil
	}
	if raw.Connector != "" {
		if raw.Connector == target {
			out.Decision = ownershipDecisionAllowed
			return out, nil
		}
		out.Decision = ownershipDecisionRejected
		if raw.Generated {
			out.Message = fmt.Sprintf("generated connector output belongs to %q, not target %q", raw.Connector, target)
			return out, []Finding{ownershipFinding(RuleOwnershipUnrelatedGenerated, raw.Connector, raw.Path, raw.Connector, out.Message, "regenerate only the target connector output or split unrelated generated churn")}
		}
		out.Message = fmt.Sprintf("connector-owned path belongs to %q, not target %q", raw.Connector, target)
		return out, []Finding{ownershipFinding(RuleOwnershipUnrelatedConnector, raw.Connector, raw.Path, raw.Connector, out.Message, "split unrelated connector changes into their own PR")}
	}
	if raw.Class == ownershipClassSharedGenerated {
		out.Decision = ownershipDecisionAllowed
		return out, nil
	}
	if raw.Shared {
		out.Decision = ownershipDecisionRejected
		out.Message = fmt.Sprintf("shared path %q is outside connector target %q", raw.Path, target)
		return out, []Finding{ownershipFinding(RuleOwnershipSharedPath, target, raw.Path, raw.Path, out.Message, "move shared runtime/tooling/docs changes to a separately reviewed foundation PR")}
	}
	out.Decision = ownershipDecisionAllowed
	return out, nil
}

func hasOwnershipEnforcedChange(paths []ownershipPathClass) bool {
	for _, path := range paths {
		if path.Ignored {
			continue
		}
		return true
	}
	return false
}

func connectorPathSegment(rel, prefix string, lx lexicon) (string, bool) {
	if !strings.HasPrefix(rel, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	if !ok || name == "" {
		return "", false
	}
	if _, ok := lx.byName[name]; ok {
		return name, true
	}
	return "", false
}

func legacyConnectorPathSegment(rel string, lx lexicon) (string, bool) {
	const prefix = "internal/connectors/"
	if !strings.HasPrefix(rel, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, ok := strings.Cut(rest, "/")
	if !ok || name == "" {
		return "", false
	}
	if name == "defs" || name == "hooks" || name == "native" {
		return "", false
	}
	if _, ok := lx.byName[name]; ok {
		return name, true
	}
	return "", false
}

func connectorDocsPath(rel string, lx lexicon) (string, bool) {
	const prefix = "docs/connectors/"
	if !strings.HasPrefix(rel, prefix) || strings.HasPrefix(rel, prefix+"icons/") {
		return "", false
	}
	rest := strings.TrimPrefix(rel, prefix)
	name, _, hasSlash := strings.Cut(rest, "/")
	if hasSlash {
		if _, ok := lx.byName[name]; ok {
			return name, true
		}
		return "", false
	}
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if _, ok := lx.byName[base]; ok {
		return base, true
	}
	best := ""
	for connector := range lx.byName {
		if strings.HasPrefix(base, connector+"-") && len(connector) > len(best) {
			best = connector
		}
	}
	if best != "" {
		return best, true
	}
	return "", false
}

func connectorIconPath(rel, prefix string, lx lexicon) (string, bool) {
	if !strings.HasPrefix(rel, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(rel, prefix)
	base := filepath.Base(rest)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if _, ok := lx.byName[name]; ok {
		return name, true
	}
	return "", false
}

func isIgnoredOwnershipPath(rel string) bool {
	return rel == "" ||
		strings.HasPrefix(rel, ".git/") ||
		strings.Contains(rel, "/.git/") ||
		strings.HasPrefix(rel, "vendor/") ||
		strings.Contains(rel, "/node_modules/") ||
		strings.Contains(rel, "/.next/") ||
		strings.HasPrefix(rel, "dist/") ||
		strings.HasPrefix(rel, "build/")
}

// isConnectorPlanningArtifactPath allows a connector implementation lane to
// commit its own GSD plan/TDD/verification artifacts under .planning/phases/,
// which AGENTS.md requires connector lanes to produce. It is intentionally
// unscoped to any single phase directory, matching how the rest of this
// package treats planning evidence as lane-local and non-shared.
func isConnectorPlanningArtifactPath(rel string) bool {
	return strings.HasPrefix(rel, ".planning/phases/")
}

// isOwnershipGateConfigPath rejects edits to the guardrail's own
// implementation, exception ledger, and required-check wiring. It is
// deliberately narrow: cmd/connectorgen/ hosts connector-owned tests
// (see connectorOwnedSharedPackageTestPath) alongside the guard's CLI
// entrypoints, so only the guard's own files are listed here.
func isOwnershipGateConfigPath(rel string) bool {
	switch rel {
	case DefaultExceptionsPath,
		"docs/migration/connector-boundary-guard.md",
		"cmd/connectorgen/ownership.go",
		"cmd/connectorgen/ownership_test.go",
		"cmd/connectorgen/boundary.go":
		return true
	}
	return strings.HasPrefix(rel, "internal/connectors/boundary/") ||
		strings.HasPrefix(rel, dotGHPathPrefix()+"workflows/")
}

func dotGHPathPrefix() string {
	return "." + "git" + "hub/"
}

// isNarrowSharedOwnershipOutput lists the shared generated indexes and goldens
// a connector lane legitimately regenerates. Every entry is a literal path with
// observed evidence in the audited connector merge-commit corpus; no directory
// prefix is allowed, so unrelated generated surfaces under the same directory
// (for example the per-namespace pages of docs/cli/ that pm docs emits for
// agent, rlm, runtime, credentials, query, or schedule) still fail closed as
// shared docs.
func isNarrowSharedOwnershipOutput(rel string) bool {
	switch rel {
	case "internal/connectors/defs/defs.go",
		"internal/connectors/hooks/hookset/hookset_gen.go",
		"internal/connectors/manifestindex/index_gen.go",
		"internal/connectors/native/nativeset/nativeset_gen.go",
		"internal/connectors/icons.go",
		"docs/cli/connectors.md",
		"docs/cli/reverse.md",
		"docs/connectors/README.md",
		"docs/connectors/UNPORTED.md",
		"docs/connectors/catalog/all-connectors.json",
		"docs/connectors/catalog/all-connectors.md",
		"internal/cli/testdata/golden_transcripts.json",
		"website/data/connectors.generated.json",
		"website/lib/connectors.generated.ts",
		"website/lib/connectors.catalog.generated.ts",
		"website/lib/connectors.catalog.data.generated.json",
		"website/lib/docs.generated.ts":
		return true
	default:
		return false
	}
}

// connectorOwnedSharedPackageTestPath recognizes connector-specific test
// files that live directly inside a shared package directory rather than
// under internal/connectors/defs|hooks|native/<connector>/, following the
// established <connector>_..._test.go naming convention (for example
// cmd/connectorgen/xero_api_surface_test.go and
// internal/connectors/engine/xero_operations_test.go). Only files whose name
// is prefixed by a known connector slug qualify, so generic shared test
// files such as main_test.go or write_test.go are unaffected.
func connectorOwnedSharedPackageTestPath(rel string, lx lexicon) (string, bool) {
	for _, dir := range []string{"cmd/connectorgen/", "internal/connectors/engine/"} {
		if !strings.HasPrefix(rel, dir) {
			continue
		}
		rest := strings.TrimPrefix(rel, dir)
		if strings.Contains(rest, "/") || !strings.HasSuffix(rest, "_test.go") {
			continue
		}
		base := strings.TrimSuffix(rest, "_test.go")
		best := ""
		for connector := range lx.byName {
			slug := strings.ReplaceAll(connector, "-", "_")
			if base != slug && !strings.HasPrefix(base, slug+"_") {
				continue
			}
			if len(connector) > len(best) {
				best = connector
			}
		}
		if best != "" {
			return best, true
		}
	}
	return "", false
}

func ownershipFinding(rule, connector, path, match, message, remediation string) Finding {
	return Finding{
		Rule:        rule,
		Severity:    SeverityError,
		Connector:   connector,
		Path:        path,
		Match:       match,
		Message:     message,
		Remediation: remediation,
	}
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortOwnershipPaths(paths []OwnershipPath) {
	sort.SliceStable(paths, func(i, j int) bool { return paths[i].Path < paths[j].Path })
}

func nonNilOwnershipPaths(paths []OwnershipPath) []OwnershipPath {
	if paths == nil {
		return []OwnershipPath{}
	}
	return paths
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
