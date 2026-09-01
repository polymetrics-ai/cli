package boundary

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

var fixedNow = time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

func TestScanDetectsSharedProviderSwitch(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/engine/branch.go": `package engine

func providerBranch(connector string) string {
	switch connector {
	case "gong":
		return "fromDateTime"
	default:
		return ""
	}
}
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	finding := requireFinding(t, report, RuleConnectorSwitch, "gong", "internal/connectors/engine/branch.go", "gong")
	if finding.Line == 0 {
		t.Fatalf("finding line was not populated: %+v", finding)
	}
	if report.Outcome != OutcomePolicyViolations {
		t.Fatalf("Outcome = %q, want %q", report.Outcome, OutcomePolicyViolations)
	}
}

func TestScanDetectsProviderPolicyInSharedHelper(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/commandrunner/helper.go": `package commandrunner

const helperParamFormat = "github_date_range"
const helperFallbackFormat = "githubDateRangeFallback"
const helperPolicyFormat = "githubOutputPolicy"
type githubDateRangeFallback struct{}
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/connectors/commandrunner/helper.go", "github_date_range")
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/connectors/commandrunner/helper.go", "githubDateRangeFallback")
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/connectors/commandrunner/helper.go", "githubOutputPolicy")
}

func TestScanDetectsWeakConnectorPolicyIdentifiers(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/defs/box/metadata.json":  `{"name":"box","display_name":"Box","integration_type":"api","docs_url":"https://developer.box.com/reference/","capabilities":{"write":true}}`,
		"internal/connectors/defs/gong/metadata.json": `{"name":"gong","display_name":"Gong","integration_type":"api","docs_url":"https://gong.example/docs","capabilities":{"write":true}}`,
		"internal/connectors/defs/mode/metadata.json": `{"name":"mode","display_name":"Mode","integration_type":"api","docs_url":"https://mode.com/developer/api-reference/","capabilities":{"read":true,"write":false}}`,
		"internal/connectors/commandrunner/weak_policy.go": `package commandrunner

type gongDateRangeFallback struct{}
type GongDateRangeFallback struct{}
type modeReadQueryFallback struct{}
type ModeReadQueryFallback struct{}
type ModeReadQueryPolicyConfig struct{}

const weakBoxLiteral = "box_output_policy"
const weakModeLiteral = "mode_read_query_fallback"
const boxOutputPolicy = "definition-owned"
const modeReadQueryFallbackValue = "definition-owned"
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleProviderPolicy, "gong", "internal/connectors/commandrunner/weak_policy.go", "gongDateRangeFallback")
	requireFinding(t, report, RuleProviderPolicy, "gong", "internal/connectors/commandrunner/weak_policy.go", "GongDateRangeFallback")
	requireFinding(t, report, RuleProviderPolicy, "mode", "internal/connectors/commandrunner/weak_policy.go", "modeReadQueryFallback")
	requireFinding(t, report, RuleProviderPolicy, "mode", "internal/connectors/commandrunner/weak_policy.go", "ModeReadQueryFallback")
	requireFinding(t, report, RuleProviderPolicy, "mode", "internal/connectors/commandrunner/weak_policy.go", "ModeReadQueryPolicyConfig")
	requireFinding(t, report, RuleProviderPolicy, "box", "internal/connectors/commandrunner/weak_policy.go", "box_output_policy")
	requireFinding(t, report, RuleProviderPolicy, "mode", "internal/connectors/commandrunner/weak_policy.go", "mode_read_query_fallback")
	requireFinding(t, report, RuleProviderPolicy, "box", "internal/connectors/commandrunner/weak_policy.go", "boxOutputPolicy")
	requireFinding(t, report, RuleProviderPolicy, "mode", "internal/connectors/commandrunner/weak_policy.go", "modeReadQueryFallbackValue")
}

func TestScanKeepsWeakConnectorMatchesConservative(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/defs/box/metadata.json":   `{"name":"box","display_name":"Box","integration_type":"api","docs_url":"https://developer.box.com/reference/","capabilities":{"write":true}}`,
		"internal/connectors/defs/merge/metadata.json": `{"name":"merge","display_name":"Merge","integration_type":"api","docs_url":"https://docs.merge.dev/api-reference/","capabilities":{"read":true,"write":false}}`,
		"internal/connectors/defs/mode/metadata.json":  `{"name":"mode","display_name":"Mode","integration_type":"api","docs_url":"https://mode.com/developer/api-reference/","capabilities":{"read":true,"write":false}}`,
		"internal/runtime/state.go": `package runtime

const ModeWholeTree = "whole_tree"

func selectedBox(box string, mode string) bool {
	boxSet := map[string]bool{}
	modeSet := map[string]bool{}
	return boxSet[box] || modeSet[mode]
}

func mergeResponseFields() {}

const plainBoxLiteral = "box"
const boxOutputLiteral = "box_output"
const modeReadQueryLiteral = "mode_read_query"
const neutralLiteral = "githubClient"
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected weak connector matches to remain conservative, got %+v", report.Findings)
	}
}

func TestScanDistinguishesAgentProjectionLookupFromHarnessProviderPolicy(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		wantFinding bool
	}{
		{
			name:        "legacy harness policy lookup trips provider policy rule",
			method:      "HarnessPolicyFor",
			wantFinding: true,
		},
		{
			name:   "neutral projection lookup stays outside connector policy",
			method: "ProjectionFor",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := newFixtureRepo(t, map[string]string{
				"internal/connectors/defs/harness/metadata.json": `{"name":"harness","display_name":"Harness"}`,
				"internal/agentcontract/contract.go":             "package agentcontract\n\ntype Contract struct{}\n\nfunc (contract *Contract) " + test.method + "(harness string) {}\n",
			})

			report, err := Scan(root, Options{Now: fixedNow})
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			if test.wantFinding {
				requireFinding(t, report, RuleProviderPolicy, "harness", "internal/agentcontract/contract.go", test.method)
				return
			}
			if len(report.Findings) != 0 {
				t.Fatalf("neutral projection lookup triggered connector policy findings: %+v", report.Findings)
			}
		})
	}
}

func TestScanDetectsLegacyConnectorPackageImport(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/engine/import_legacy.go": `package engine

import _ "polymetrics.ai/internal/connectors/gong"
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleConnectorImport, "gong", "internal/connectors/engine/import_legacy.go", "polymetrics.ai/internal/connectors/gong")
}

func TestScanDetectsDisplayNameAliasesInStringLiterals(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/defs/microsoft-teams/metadata.json":      `{"name":"microsoft-teams","display_name":"Microsoft Teams","integration_type":"api"}`,
		"internal/connectors/defs/free-agent-connector/metadata.json": `{"name":"free-agent-connector","display_name":"FreeAgent","integration_type":"api"}`,
		"internal/connectors/engine/aliases.go": `package engine

var providers = []string{"Microsoft Teams", "FreeAgent"}
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleConnectorLiteral, "microsoft-teams", "internal/connectors/engine/aliases.go", "Microsoft Teams")
	requireFinding(t, report, RuleConnectorLiteral, "free-agent-connector", "internal/connectors/engine/aliases.go", "FreeAgent")
}

func TestScanDetectsConnectorSwitchForContextualMetadataName(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/defs/mode/metadata.json": `{"name":"mode","display_name":"Mode","integration_type":"api"}`,
		"internal/runtime/provider.go": `package runtime

func selected(connector string) bool { return connector == "mode" }
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleConnectorSwitch, "mode", "internal/runtime/provider.go", "mode")
}

func TestScanDetectsProviderPolicyInSharedInternalPackages(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/runtime/policy.go": `package runtime

const helperParamFormat = "github_date_range"
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/runtime/policy.go", "github_date_range")
}

func TestScanRejectsUnsafeBaseRefsBeforeGitDiff(t *testing.T) {
	root := newFixtureRepo(t, nil)
	outputPath := filepath.Join(root, "git-option-output")
	for _, baseRef := range []string{"--output=" + outputPath, "HEAD\n--output=" + outputPath} {
		t.Run(baseRef, func(t *testing.T) {
			_, err := Scan(root, Options{BaseRef: baseRef, Now: fixedNow})
			var cfgErr *ConfigError
			if !errors.As(err, &cfgErr) {
				t.Fatalf("Scan error = %v, want ConfigError", err)
			}
			if _, statErr := os.Stat(outputPath); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("unsafe base ref created output file: %v", statErr)
			}
		})
	}
}

func TestScanAllowsDefinitionsNativeHooksGeneratedTestsAndDocs(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/defs/gong/streams.json": `{"provider":"gong","query":"fromDateTime"}`,
		"internal/connectors/native/gong/native.go": `package gong
const provider = "gong"
`,
		"internal/connectors/gong/legacy.go": `package gong
const provider = "gong"
`,
		"internal/connectors/hooks/gong/hooks.go": `package gong
const provider = "gong"
`,
		"internal/connectors/hooks/hookset/hookset_gen.go": `// Code generated by connectorgen gen. DO NOT EDIT.
package hookset

import _ "polymetrics.ai/internal/connectors/hooks/gong"
`,
		"internal/connectors/engine/branch_test.go": `package engine

const testProvider = "gong"
`,
		"internal/connectors/engine/testdata/shared.go": `package testdata

const fixtureProvider = "gong"
`,
		"docs/connectors/gong/manual.md": "# Gong\n\nProvider-owned docs can say gong.\n",
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected no blocking findings, got %+v", report.Findings)
	}
	if report.Outcome != OutcomeClean {
		t.Fatalf("Outcome = %q, want %q", report.Outcome, OutcomeClean)
	}
}

func TestScanDoesNotAllowUnknownHookOrNativeEscapeHatchDirs(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/hooks/shared/policy.go": `package shared

const outputPolicy = "github_date_range"
`,
		"internal/connectors/native/common/policy.go": `package common

type githubDateRangeFallback struct{}
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/connectors/hooks/shared/policy.go", "github_date_range")
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/connectors/native/common/policy.go", "githubDateRangeFallback")
}

func TestScanDoesNotTreatUnknownConnectorSubpackagesAsNative(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/policy/policy.go": `package policy

const outputPolicy = "github_date_range"
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/connectors/policy/policy.go", "github_date_range")
}

func TestScanDoesNotTrustGeneratedMarkerOutsideKnownOutputs(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/engine/generated_bypass.go": `// Code generated by hand. DO NOT EDIT.
package engine

const outputPolicy = "github_date_range"
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleProviderPolicy, "github", "internal/connectors/engine/generated_bypass.go", "github_date_range")
}

func TestScanAppliesLedgerToConnectorDocsOutput(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/cli/connector_docs.go": "package cli\n\nfunc requiredDocs() []string { return []string{\"icons/github.svg\", \"`github`\"} }\n",
		DefaultExceptionsPath: exceptionsJSON([]map[string]any{{
			"id":                  "github-cli-connector-docs-catalog-validation",
			"rule":                RuleDocsExample,
			"connector":           "github",
			"path":                "internal/cli/connector_docs.go",
			"match":               "github",
			"reason":              "bootstrap until generated connector catalog validation examples are provider-neutral",
			"migration_issue_url": "https://github.com/polymetrics-ai/cli/issues/67",
			"owner":               "connector-architecture-v2",
			"expires_on":          "2026-09-30",
			"max_matches":         2,
		}}),
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected connector docs exception to suppress findings, got %+v", report.Findings)
	}
	if len(report.Exceptions) != 1 || report.Exceptions[0].Matches != 2 {
		t.Fatalf("expected connector docs exception with two matches, got %+v", report.Exceptions)
	}
}

func TestScanFailsClosedWhenConnectorMetadataCannotLoad(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name: "missing defs",
			want: "read connector defs",
		},
		{
			name: "missing metadata",
			files: map[string]string{
				"internal/connectors/defs/github/streams.json": `{}`,
			},
			want: "load connector metadata github",
		},
		{
			name: "invalid metadata",
			files: map[string]string{
				"internal/connectors/defs/github/metadata.json": `{`,
			},
			want: "load connector metadata github",
		},
		{
			name: "invalid cli surface",
			files: map[string]string{
				"internal/connectors/defs/github/metadata.json":    `{"name":"github","display_name":"GitHub"}`,
				"internal/connectors/defs/github/cli_surface.json": `{`,
			},
			want: "load connector cli surface github",
		},
		{
			name: "blank metadata name",
			files: map[string]string{
				"internal/connectors/defs/github/metadata.json": `{"display_name":"GitHub"}`,
			},
			want: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeFixtureFile(t, root, "go.mod", "module polymetrics.ai\n\ngo 1.25\n")
			for path, content := range tt.files {
				writeFixtureFile(t, root, path, content)
			}
			_, err := Scan(root, Options{Now: fixedNow})
			var cfgErr *ConfigError
			if !errors.As(err, &cfgErr) {
				t.Fatalf("Scan error = %v, want ConfigError", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Scan error = %q, want substring %q", err, tt.want)
			}
		})
	}
}

func TestExceptionLedgerSuppressesExactBoundedFinding(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/engine/direct_read.go": `package engine

const outputPolicy = "github_contents_directory"
`,
		DefaultExceptionsPath: exceptionsJSON([]map[string]any{{
			"id":                  "github-direct-read-output-policy",
			"rule":                RuleProviderPolicy,
			"connector":           "github",
			"path":                "internal/connectors/engine/direct_read.go",
			"match":               "github_contents_directory",
			"reason":              "bootstrap until output policies are definition-owned",
			"migration_issue_url": "https://github.com/polymetrics-ai/cli/issues/599",
			"owner":               "connector-architecture-v2",
			"expires_on":          "2026-08-31",
			"max_matches":         1,
		}}),
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected exception to suppress finding, got %+v", report.Findings)
	}
	if len(report.Exceptions) != 1 || report.Exceptions[0].Matches != 1 {
		t.Fatalf("expected one applied exception with one match, got %+v", report.Exceptions)
	}
}

func TestExceptionLedgerFailures(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string
		now       time.Time
		wantRule  string
		wantMatch string
	}{
		{
			name: "expired even with approved_by prose",
			files: map[string]string{
				"internal/connectors/engine/direct_read.go": `package engine

const outputPolicy = "github_contents_directory"
`,
				DefaultExceptionsPath: exceptionsJSON([]map[string]any{{
					"id":                  "expired",
					"rule":                RuleProviderPolicy,
					"connector":           "github",
					"path":                "internal/connectors/engine/direct_read.go",
					"match":               "github_contents_directory",
					"reason":              "bootstrap",
					"migration_issue_url": "https://github.com/polymetrics-ai/cli/issues/599",
					"owner":               "connector-architecture-v2",
					"expires_on":          "2026-07-01",
					"max_matches":         1,
					"approved_by":         "do not treat this prose as approval",
				}}),
			},
			now:       fixedNow,
			wantRule:  RuleExceptionExpired,
			wantMatch: "github_contents_directory",
		},
		{
			name: "stale",
			files: map[string]string{
				DefaultExceptionsPath: exceptionsJSON([]map[string]any{{
					"id":                  "stale",
					"rule":                RuleProviderPolicy,
					"connector":           "github",
					"path":                "internal/connectors/engine/direct_read.go",
					"match":               "github_contents_directory",
					"reason":              "bootstrap",
					"migration_issue_url": "https://github.com/polymetrics-ai/cli/issues/599",
					"owner":               "connector-architecture-v2",
					"expires_on":          "2026-08-31",
					"max_matches":         1,
				}}),
			},
			now:       fixedNow,
			wantRule:  RuleExceptionStale,
			wantMatch: "github_contents_directory",
		},
		{
			name: "broadened",
			files: map[string]string{
				"internal/connectors/engine/direct_read.go": `package engine

const outputPolicy = "github_contents_directory"
const secondOutputPolicy = "github_contents_directory"
`,
				DefaultExceptionsPath: exceptionsJSON([]map[string]any{{
					"id":                  "broadened",
					"rule":                RuleProviderPolicy,
					"connector":           "github",
					"path":                "internal/connectors/engine/direct_read.go",
					"match":               "github_contents_directory",
					"reason":              "bootstrap",
					"migration_issue_url": "https://github.com/polymetrics-ai/cli/issues/599",
					"owner":               "connector-architecture-v2",
					"expires_on":          "2026-08-31",
					"max_matches":         1,
				}}),
			},
			now:       fixedNow,
			wantRule:  RuleExceptionBroadened,
			wantMatch: "github_contents_directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := newFixtureRepo(t, tt.files)
			report, err := Scan(root, Options{Now: tt.now})
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			requireFinding(t, report, tt.wantRule, "github", DefaultExceptionsPath, tt.wantMatch)
		})
	}
}

func TestFindingsAreSortedDeterministically(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/engine/z.go": `package engine
const z = "github_date_range"
`,
		"internal/connectors/engine/a.go": `package engine
func a(connector string) bool { return connector == "gong" }
`,
	})

	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(report.Findings) < 2 {
		t.Fatalf("expected at least two findings, got %+v", report.Findings)
	}
	paths := make([]string, 0, len(report.Findings))
	for _, f := range report.Findings {
		paths = append(paths, f.Path)
	}
	if !slices.IsSorted(paths) {
		t.Fatalf("findings are not sorted by path: %v", paths)
	}
}

func TestReportJSONShapeUsesStableArrays(t *testing.T) {
	root := newFixtureRepo(t, nil)
	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	b, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	jsonText := string(b)
	for _, want := range []string{`"api_version":"polymetrics.ai/v1"`, `"kind":"ConnectorBoundaryReport"`, `"findings":[]`, `"warnings":[]`, `"exceptions":[]`} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("report JSON missing %s: %s", want, jsonText)
		}
	}
}

func TestScanBaseDiffRestrictsPrimaryScan(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/engine/changed.go": `package engine

const neutral = "ok"
`,
		"internal/connectors/engine/unchanged.go": `package engine
const unchanged = "github_date_range"
`,
	})
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "base")
	writeFixtureFile(t, root, "internal/connectors/engine/changed.go", `package engine
func changed(connector string) bool { return connector == "gong" }
`)

	report, err := Scan(root, Options{BaseRef: "HEAD", Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleConnectorSwitch, "gong", "internal/connectors/engine/changed.go", "gong")
	if hasFinding(report, RuleProviderPolicy, "github", "github_date_range") {
		t.Fatalf("base diff scan reported unchanged baseline provider policy: %+v", report.Findings)
	}
}

func TestScanBaseDiffDoesNotMarkUnchangedAppliedExceptionsStale(t *testing.T) {
	root := newFixtureRepo(t, map[string]string{
		"internal/connectors/engine/changed.go": `package engine

const neutral = "ok"
`,
		"internal/connectors/engine/direct_read.go": `package engine

const outputPolicy = "github_contents_directory"
`,
		DefaultExceptionsPath: exceptionsJSON([]map[string]any{{
			"id":                  "github-direct-read-output-policy",
			"rule":                RuleProviderPolicy,
			"connector":           "github",
			"path":                "internal/connectors/engine/direct_read.go",
			"match":               "github_contents_directory",
			"reason":              "bootstrap until output policies are definition-owned",
			"migration_issue_url": "https://github.com/polymetrics-ai/cli/issues/599",
			"owner":               "connector-architecture-v2",
			"expires_on":          "2026-08-31",
			"max_matches":         1,
		}}),
	})
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "base")
	writeFixtureFile(t, root, "internal/connectors/engine/changed.go", `package engine
func changed(connector string) bool { return connector == "gong" }
`)

	report, err := Scan(root, Options{BaseRef: "HEAD", Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	requireFinding(t, report, RuleConnectorSwitch, "gong", "internal/connectors/engine/changed.go", "gong")
	if hasFinding(report, RuleExceptionStale, "github", "github_contents_directory") {
		t.Fatalf("base diff scan reported unchanged applied exception as stale: %+v", report.Findings)
	}
	if len(report.Exceptions) != 1 || report.Exceptions[0].ID != "github-direct-read-output-policy" {
		t.Fatalf("expected unchanged exception to remain applied, got %+v", report.Exceptions)
	}
}

func TestCurrentRepositoryBaselinePasses(t *testing.T) {
	root := findRepoRoot(t)
	report, err := Scan(root, Options{Now: fixedNow})
	if err != nil {
		t.Fatalf("Scan current repository: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("current repository baseline has boundary findings: %+v", report.Findings)
	}
	if hasAppliedExceptionForConnector(report, "gong") {
		t.Fatalf("Gong must remain definition-owned without a shared-code exception: %+v", report.Exceptions)
	}
}

func requireFinding(t *testing.T, report Report, rule, connector, path, match string) Finding {
	t.Helper()
	for _, f := range report.Findings {
		if f.Rule == rule && f.Connector == connector && f.Path == path && f.Match == match {
			return f
		}
	}
	t.Fatalf("missing finding rule=%s connector=%s path=%s match=%s in %+v", rule, connector, path, match, report.Findings)
	return Finding{}
}

func hasFinding(report Report, rule, connector, match string) bool {
	for _, f := range report.Findings {
		if f.Rule == rule && f.Connector == connector && f.Match == match {
			return true
		}
	}
	return false
}

func hasAppliedExceptionForConnector(report Report, connector string) bool {
	for _, ex := range report.Exceptions {
		if ex.Connector == connector {
			return true
		}
	}
	return false
}

func newFixtureRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module polymetrics.ai\n\ngo 1.25\n")
	writeFixtureFile(t, root, "internal/connectors/defs/github/metadata.json", `{"name":"github","display_name":"GitHub"}`)
	writeFixtureFile(t, root, "internal/connectors/defs/gong/metadata.json", `{"name":"gong","display_name":"Gong"}`)
	for path, content := range files {
		writeFixtureFile(t, root, path, content)
	}
	return root
}

func writeFixtureFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func exceptionsJSON(entries []map[string]any) string {
	b, err := json.Marshal(map[string]any{"exceptions": entries})
	if err != nil {
		panic(err)
	}
	return string(b)
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found from %s", dir)
		}
		dir = parent
	}
}
