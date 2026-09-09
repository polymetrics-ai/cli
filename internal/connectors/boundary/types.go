// Package boundary scans the repository for connector-specific policy that has
// leaked out of connector definition bundles or approved escape-hatch code.
package boundary

import "time"

const (
	// APIVersion is the stable JSON contract version for boundary reports.
	APIVersion = "polymetrics.ai/v1"
	// Kind is the stable JSON kind for boundary reports.
	Kind = "ConnectorBoundaryReport"
)

const (
	// OutcomeClean means no blocking policy findings remain after exact exceptions.
	OutcomeClean = "clean"
	// OutcomePolicyViolations means boundary policy findings were found.
	OutcomePolicyViolations = "policy_violations"
)

const (
	// ModeWholeTree scans every shared production Go file in the repo.
	ModeWholeTree = "whole_tree"
	// ModeBaseDiff scans files changed relative to a git base while still validating
	// exception contracts against the whole tree.
	ModeBaseDiff = "base_diff"
)

const (
	// RuleConnectorLiteral catches exact connector IDs or aliases in shared production Go.
	RuleConnectorLiteral = "connector_literal"
	// RuleConnectorSwitch catches connector IDs in switch cases or equality branches.
	RuleConnectorSwitch = "connector_switch"
	// RuleConnectorImport catches connector-specific hook/native imports outside allowed wiring.
	RuleConnectorImport = "connector_import"
	// RuleProviderPolicy catches provider-prefixed shared policy identifiers.
	RuleProviderPolicy = "provider_policy"
	// RuleDocsExample catches provider-specific examples/resources embedded in shared Go docs/help generators.
	RuleDocsExample = "docs_example"
	// RuleLegacyAlias catches source-/destination-prefixed legacy connector aliases in shared code.
	RuleLegacyAlias = "legacy_alias"
	// RuleExceptionExpired catches expired exception rows.
	RuleExceptionExpired = "exception_expired"
	// RuleExceptionStale catches exception rows whose exact finding stopped matching.
	RuleExceptionStale = "exception_stale"
	// RuleExceptionBroadened catches exception rows whose match count exceeded max_matches.
	RuleExceptionBroadened = "exception_broadened"
)

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// DefaultExceptionsPath is the repo-relative exception ledger location.
const DefaultExceptionsPath = "internal/connectors/boundary/exceptions.json"

// Options configures a boundary scan.
type Options struct {
	BaseRef        string
	ExceptionsPath string
	Now            time.Time
}

// Finding is one boundary policy finding.
type Finding struct {
	Rule        string `json:"rule"`
	Severity    string `json:"severity"`
	Connector   string `json:"connector"`
	Path        string `json:"path"`
	Line        int    `json:"line"`
	Match       string `json:"match"`
	AllowedBy   string `json:"allowed_by"`
	Message     string `json:"message"`
	Remediation string `json:"remediation"`
}

// AppliedException records an exception row that matched and suppressed current findings.
type AppliedException struct {
	ID        string `json:"id"`
	Rule      string `json:"rule"`
	Connector string `json:"connector"`
	Path      string `json:"path"`
	Match     string `json:"match"`
	Matches   int    `json:"matches"`
	IssueURL  string `json:"migration_issue_url"`
	Owner     string `json:"owner"`
	ExpiresOn string `json:"expires_on"`
}

// Report is the deterministic scanner output rendered by connectorgen boundary.
type Report struct {
	APIVersion       string             `json:"api_version"`
	Kind             string             `json:"kind"`
	Outcome          string             `json:"outcome"`
	RepoRoot         string             `json:"repo_root"`
	Mode             string             `json:"mode"`
	BaseRef          string             `json:"base_ref,omitempty"`
	CheckedFiles     int                `json:"checked_files"`
	ConnectorsLoaded int                `json:"connectors_loaded"`
	Findings         []Finding          `json:"findings"`
	Warnings         []Finding          `json:"warnings"`
	Exceptions       []AppliedException `json:"exceptions"`
}

// ConfigError marks invalid invocation or scanner configuration errors. The CLI
// maps these to exit status 2 so they remain distinct from policy violations.
type ConfigError struct {
	Err error
}

func (e *ConfigError) Error() string { return e.Err.Error() }
func (e *ConfigError) Unwrap() error { return e.Err }
