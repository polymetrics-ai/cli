// Command materialize_cli_rate_limits creates the one provider-cited
// rate_limits.json declaration required for every target in the 426-connector
// artifact sweep. It is intentionally task-scoped: the engine owns schema
// validation and runtime pacing, while this command preserves the official
// source evidence that determines declared versus unknown.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	rateLimitSourceLedgerSchemaVersion = 1
	rateLimitBatchReportSchemaVersion  = 1
	maxRateLimitMaterializeBatchSize   = 40
	oneFlowRateLimitAuthority          = "CAPTAIN-ORDER-426-one-flow-sweep-20260809.md"
)

var rateLimitConnectorNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type rateLimitTargetLedger struct {
	Targets []rateLimitTarget `json:"targets"`
}

type rateLimitTarget struct {
	Connector    string `json:"connector"`
	SourceKind   string `json:"source_kind"`
	SourceURL    string `json:"source_url"`
	RetrievedAt  string `json:"retrieved_at"`
	SourceReason string `json:"source_reason"`
}

type rateLimitResearchLedger struct {
	Records []rateLimitResearchRecord `json:"records"`
}

type rateLimitResearchRecord struct {
	Connector   string `json:"connector"`
	Verdict     string `json:"verdict"`
	SourceURL   string `json:"source_url"`
	RetrievedAt string `json:"retrieved_at"`
	Reason      string `json:"reason"`
}

type rateLimitSourceLedger struct {
	SchemaVersion int                    `json:"schema_version"`
	Authority     string                 `json:"authority"`
	GeneratedAt   string                 `json:"generated_at"`
	TargetTotal   int                    `json:"target_total"`
	Entries       []rateLimitSourceEntry `json:"entries"`
}

type rateLimitSourceEntry struct {
	Connector   string             `json:"connector"`
	Evidence    rateLimitEvidence  `json:"evidence"`
	Declaration connsdk.RateLimits `json:"declaration"`
}

type rateLimitEvidence struct {
	Kind        string `json:"kind"`
	URL         string `json:"url,omitempty"`
	RetrievedAt string `json:"retrieved_at,omitempty"`
}

type rateLimitBatchReport struct {
	SchemaVersion int                     `json:"schema_version"`
	SourceLedger  string                  `json:"source_ledger"`
	Offset        int                     `json:"offset"`
	RequestedSize int                     `json:"requested_size"`
	Entries       []rateLimitBatchOutcome `json:"entries"`
	Counts        rateLimitCounts         `json:"counts"`
}

type rateLimitBatchOutcome struct {
	Connector string `json:"connector"`
	State     string `json:"state"`
	File      string `json:"file"`
	Status    string `json:"status"`
}

type rateLimitCounts struct {
	Declared      int `json:"declared"`
	Unknown       int `json:"unknown"`
	NotApplicable int `json:"not_applicable"`
	FileTotal     int `json:"file_total"`
}

type rateLimitMaterializeOptions struct {
	bootstrap      bool
	targetLedger   string
	researchLedger string
	sourceLedger   string
	generatedAt    string
	defsRoot       string
	reportPath     string
	offset         int
	size           int
}

func main() {
	os.Exit(runRateLimitMaterializer(os.Args[1:], os.Stdout, os.Stderr))
}

func runRateLimitMaterializer(args []string, stdout, stderr io.Writer) int {
	opts, err := parseRateLimitMaterializeOptions(args)
	if err != nil {
		fmt.Fprintf(stderr, "materialize_cli_rate_limits: %v\n", err)
		return 2
	}
	if opts.bootstrap {
		ledger, err := buildRateLimitSourceLedger(opts.targetLedger, opts.researchLedger, opts.generatedAt)
		if err != nil {
			fmt.Fprintf(stderr, "materialize_cli_rate_limits: bootstrap: %v\n", err)
			return 1
		}
		ledger, _, err = reconcileRateLimitSourceLedger(ledger, opts.sourceLedger, opts.defsRoot, true)
		if err != nil {
			fmt.Fprintf(stderr, "materialize_cli_rate_limits: preserve declared source evidence: %v\n", err)
			return 1
		}
		if err := writeJSONAtomically(opts.sourceLedger, ledger); err != nil {
			fmt.Fprintf(stderr, "materialize_cli_rate_limits: write source ledger: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "materialize_cli_rate_limits: bootstrapped %d provider-cited source entries at %s\n", len(ledger.Entries), opts.sourceLedger)
		return 0
	}

	ledger, err := readRateLimitSourceLedger(opts.sourceLedger)
	if err != nil {
		fmt.Fprintf(stderr, "materialize_cli_rate_limits: read source ledger: %v\n", err)
		return 1
	}
	ledger, changed, err := reconcileRateLimitSourceLedger(ledger, opts.sourceLedger, opts.defsRoot, false)
	if err != nil {
		fmt.Fprintf(stderr, "materialize_cli_rate_limits: preserve declared source evidence: %v\n", err)
		return 1
	}
	if changed {
		if err := writeJSONAtomically(opts.sourceLedger, ledger); err != nil {
			fmt.Fprintf(stderr, "materialize_cli_rate_limits: update source ledger: %v\n", err)
			return 1
		}
	}
	selected, err := selectRateLimitEntries(ledger.Entries, opts.offset, opts.size)
	if err != nil {
		fmt.Fprintf(stderr, "materialize_cli_rate_limits: select batch: %v\n", err)
		return 2
	}
	report, err := materializeRateLimitBatch(opts.sourceLedger, opts.defsRoot, opts.offset, opts.size, selected)
	if err != nil {
		fmt.Fprintf(stderr, "materialize_cli_rate_limits: materialize batch: %v\n", err)
		return 1
	}
	if err := writeJSONAtomically(opts.reportPath, report); err != nil {
		fmt.Fprintf(stderr, "materialize_cli_rate_limits: write report: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "materialize_cli_rate_limits: materialized %d declaration(s), report %s\n", len(report.Entries), opts.reportPath)
	return 0
}

func parseRateLimitMaterializeOptions(args []string) (rateLimitMaterializeOptions, error) {
	fs := flag.NewFlagSet("materialize_cli_rate_limits", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	opts := rateLimitMaterializeOptions{}
	fs.BoolVar(&opts.bootstrap, "bootstrap", false, "build the source ledger from immutable inputs")
	fs.StringVar(&opts.targetLedger, "target-ledger", "", "426 target ledger JSON")
	fs.StringVar(&opts.researchLedger, "research-ledger", "", "official rate-limit research ledger JSON")
	fs.StringVar(&opts.sourceLedger, "source-ledger", "", "rate-limit source ledger JSON")
	fs.StringVar(&opts.generatedAt, "generated-at", "", "ISO date for the source-ledger snapshot")
	fs.StringVar(&opts.defsRoot, "defs-root", "", "connector definition root")
	fs.StringVar(&opts.reportPath, "report", "", "compact materialization report")
	fs.IntVar(&opts.offset, "offset", 0, "zero-based deterministic source-ledger offset")
	fs.IntVar(&opts.size, "size", maxRateLimitMaterializeBatchSize, "batch size, 1 through 40")
	if err := fs.Parse(args); err != nil {
		return rateLimitMaterializeOptions{}, err
	}
	if fs.NArg() != 0 {
		return rateLimitMaterializeOptions{}, fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}
	if opts.sourceLedger == "" {
		return rateLimitMaterializeOptions{}, errors.New("--source-ledger is required")
	}
	if opts.bootstrap {
		if opts.targetLedger == "" || opts.researchLedger == "" || opts.generatedAt == "" {
			return rateLimitMaterializeOptions{}, errors.New("--bootstrap requires --target-ledger, --research-ledger, and --generated-at")
		}
		if err := validateRateLimitDate(opts.generatedAt); err != nil {
			return rateLimitMaterializeOptions{}, fmt.Errorf("--generated-at: %w", err)
		}
		if opts.reportPath != "" {
			return rateLimitMaterializeOptions{}, errors.New("--bootstrap does not accept --report")
		}
		if opts.defsRoot == "" {
			opts.defsRoot = "internal/connectors/defs"
		}
		return opts, nil
	}
	if opts.defsRoot == "" || opts.reportPath == "" {
		return rateLimitMaterializeOptions{}, errors.New("materialization requires --defs-root and --report")
	}
	if opts.offset < 0 {
		return rateLimitMaterializeOptions{}, errors.New("--offset must not be negative")
	}
	if opts.size < 1 || opts.size > maxRateLimitMaterializeBatchSize {
		return rateLimitMaterializeOptions{}, fmt.Errorf("--size must be between 1 and %d", maxRateLimitMaterializeBatchSize)
	}
	return opts, nil
}

func buildRateLimitSourceLedger(targetPath, researchPath, generatedAt string) (rateLimitSourceLedger, error) {
	var targets rateLimitTargetLedger
	if err := readRateLimitJSON(targetPath, &targets); err != nil {
		return rateLimitSourceLedger{}, fmt.Errorf("read target ledger: %w", err)
	}
	var research rateLimitResearchLedger
	if err := readRateLimitJSON(researchPath, &research); err != nil {
		return rateLimitSourceLedger{}, fmt.Errorf("read official rate-limit research: %w", err)
	}
	if len(targets.Targets) == 0 {
		return rateLimitSourceLedger{}, errors.New("target ledger has no targets")
	}
	if err := validateRateLimitDate(generatedAt); err != nil {
		return rateLimitSourceLedger{}, fmt.Errorf("generated_at: %w", err)
	}

	targetsByName := make(map[string]rateLimitTarget, len(targets.Targets))
	for _, target := range targets.Targets {
		if !rateLimitConnectorNamePattern.MatchString(target.Connector) {
			return rateLimitSourceLedger{}, fmt.Errorf("target connector %q is invalid", target.Connector)
		}
		if _, exists := targetsByName[target.Connector]; exists {
			return rateLimitSourceLedger{}, fmt.Errorf("target connector %q is duplicated", target.Connector)
		}
		targetsByName[target.Connector] = target
	}

	overrides := make(map[string]rateLimitResearchRecord, len(research.Records))
	for _, record := range research.Records {
		if !rateLimitConnectorNamePattern.MatchString(record.Connector) {
			return rateLimitSourceLedger{}, fmt.Errorf("research connector %q is invalid", record.Connector)
		}
		if _, known := targetsByName[record.Connector]; !known {
			return rateLimitSourceLedger{}, fmt.Errorf("research connector %q is outside the target ledger", record.Connector)
		}
		if _, exists := overrides[record.Connector]; exists {
			return rateLimitSourceLedger{}, fmt.Errorf("research connector %q is duplicated", record.Connector)
		}
		if err := validateRateLimitReference(record.SourceURL, true); err != nil {
			return rateLimitSourceLedger{}, fmt.Errorf("research connector %q source: %w", record.Connector, err)
		}
		if err := validateRateLimitDate(record.RetrievedAt); err != nil {
			return rateLimitSourceLedger{}, fmt.Errorf("research connector %q retrieved_at: %w", record.Connector, err)
		}
		if strings.TrimSpace(record.Reason) == "" {
			return rateLimitSourceLedger{}, fmt.Errorf("research connector %q has no official-source reasoning", record.Connector)
		}
		overrides[record.Connector] = record
	}

	entries := make([]rateLimitSourceEntry, 0, len(targets.Targets))
	for _, target := range targets.Targets {
		entry, err := rateLimitEntryForTarget(target, overrides[target.Connector])
		if err != nil {
			return rateLimitSourceLedger{}, err
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Connector < entries[j].Connector })
	return rateLimitSourceLedger{
		SchemaVersion: rateLimitSourceLedgerSchemaVersion,
		Authority:     oneFlowRateLimitAuthority,
		GeneratedAt:   generatedAt,
		TargetTotal:   len(entries),
		Entries:       entries,
	}, nil
}

func rateLimitEntryForTarget(target rateLimitTarget, override rateLimitResearchRecord) (rateLimitSourceEntry, error) {
	if target.SourceKind == "none" {
		reason := strings.TrimSpace(target.SourceReason)
		if reason == "" {
			reason = "The target ledger records no external provider HTTP/API surface."
		}
		return rateLimitSourceEntry{
			Connector: target.Connector,
			Evidence: rateLimitEvidence{
				Kind:        "no_provider_http_api",
				URL:         target.SourceURL,
				RetrievedAt: target.RetrievedAt,
			},
			Declaration: connsdk.RateLimits{
				SchemaVersion: 1,
				State:         connsdk.RateLimitStateNotApplicable,
				Reason:        fmt.Sprintf("%s Provider HTTP/API rate limiting is not applicable.", reason),
			},
		}, nil
	}

	if override.Connector != "" {
		declaration, err := rateLimitDeclarationFromResearch(override)
		if err != nil {
			return rateLimitSourceEntry{}, err
		}
		return rateLimitSourceEntry{
			Connector: target.Connector,
			Evidence: rateLimitEvidence{
				Kind:        "official_rate_limit_reference",
				URL:         override.SourceURL,
				RetrievedAt: override.RetrievedAt,
			},
			Declaration: declaration,
		}, nil
	}

	if err := validateRateLimitReference(target.SourceURL, false); err != nil {
		return rateLimitSourceEntry{}, fmt.Errorf("target connector %q official operation source: %w", target.Connector, err)
	}
	if err := validateRateLimitDate(target.RetrievedAt); err != nil {
		return rateLimitSourceEntry{}, fmt.Errorf("target connector %q source retrieved_at: %w", target.Connector, err)
	}
	return rateLimitSourceEntry{
		Connector: target.Connector,
		Evidence: rateLimitEvidence{
			Kind:        "official_operation_reference",
			URL:         target.SourceURL,
			RetrievedAt: target.RetrievedAt,
		},
		Declaration: connsdk.RateLimits{
			SchemaVersion: 1,
			State:         connsdk.RateLimitStateUnknown,
			Reason: fmt.Sprintf(
				"Official provider operation source %s (retrieved %s) does not publish a complete enforceable rate-limit policy in the retained source: selector, non-secret scope, budget, and replenishment model are not all documented. No numeric rate or bucket model is declared.",
				target.SourceURL,
				target.RetrievedAt,
			),
		},
	}, nil
}

func rateLimitDeclarationFromResearch(record rateLimitResearchRecord) (connsdk.RateLimits, error) {
	source := connsdk.RateLimitSource{URL: record.SourceURL, RetrievedAt: record.RetrievedAt}
	unknown := connsdk.RateLimits{
		SchemaVersion: 1,
		State:         connsdk.RateLimitStateUnknown,
		Reason:        fmt.Sprintf("Official provider rate-limit source %s (retrieved %s): %s", record.SourceURL, record.RetrievedAt, record.Reason),
	}
	switch record.Verdict {
	case "unknown":
		return unknown, nil
	case "declared":
		switch record.Connector {
		case "harvest":
			return declaredRateLimits([]connsdk.RateLimitPolicy{{
				ID:       "general-account",
				Source:   source,
				Selector: connsdk.RateLimitSelector{All: true},
				Scope:    connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "account_id"},
				Budgets: []connsdk.RateLimitBudget{{
					Model:         connsdk.RateLimitBudgetSlidingWindow,
					Dimension:     connsdk.RateLimitBudgetSustained,
					Unit:          connsdk.RateLimitBudgetRequests,
					Limit:         rateLimitInt(100),
					WindowSeconds: rateLimitInt(15),
				}},
			}}), nil
		case "callrail":
			return declaredRateLimits([]connsdk.RateLimitPolicy{
				{
					ID:       "general-account-hourly",
					Source:   source,
					Selector: connsdk.RateLimitSelector{All: true},
					Scope:    connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "account_id"},
					Budgets: []connsdk.RateLimitBudget{{
						Model:         connsdk.RateLimitBudgetFixedWindow,
						Dimension:     connsdk.RateLimitBudgetSustained,
						Unit:          connsdk.RateLimitBudgetRequests,
						Limit:         rateLimitInt(1000),
						WindowSeconds: rateLimitInt(3600),
					}},
				},
				{
					ID:       "general-account-daily",
					Source:   source,
					Selector: connsdk.RateLimitSelector{All: true},
					Scope:    connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "account_id"},
					Budgets: []connsdk.RateLimitBudget{{
						Model:         connsdk.RateLimitBudgetFixedWindow,
						Dimension:     connsdk.RateLimitBudgetSustained,
						Unit:          connsdk.RateLimitBudgetRequests,
						Limit:         rateLimitInt(10000),
						WindowSeconds: rateLimitInt(86400),
					}},
				},
			}), nil
		case "aha":
			return declaredRateLimits([]connsdk.RateLimitPolicy{
				{
					ID:       "account-per-second",
					Source:   source,
					Selector: connsdk.RateLimitSelector{All: true},
					Scope:    connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "base_url"},
					Budgets: []connsdk.RateLimitBudget{{
						Model:         connsdk.RateLimitBudgetFixedWindow,
						Dimension:     connsdk.RateLimitBudgetSustained,
						Unit:          connsdk.RateLimitBudgetRequests,
						Limit:         rateLimitInt(20),
						WindowSeconds: rateLimitInt(1),
					}},
				},
				{
					ID:       "account-per-minute",
					Source:   source,
					Selector: connsdk.RateLimitSelector{All: true},
					Scope:    connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "base_url"},
					Budgets: []connsdk.RateLimitBudget{{
						Model:         connsdk.RateLimitBudgetFixedWindow,
						Dimension:     connsdk.RateLimitBudgetSustained,
						Unit:          connsdk.RateLimitBudgetRequests,
						Limit:         rateLimitInt(300),
						WindowSeconds: rateLimitInt(60),
					}},
				},
			}), nil
		default:
			return connsdk.RateLimits{}, fmt.Errorf("official research marks %q declared but no policy template is approved", record.Connector)
		}
	default:
		return connsdk.RateLimits{}, fmt.Errorf("official research connector %q has unsupported verdict %q", record.Connector, record.Verdict)
	}
}

func declaredRateLimits(policies []connsdk.RateLimitPolicy) connsdk.RateLimits {
	return connsdk.RateLimits{SchemaVersion: 1, State: connsdk.RateLimitStateDeclared, Policies: policies}
}

func rateLimitInt(value int) *int {
	return &value
}

// reconcileRateLimitSourceLedger is deliberately monotonic for policy
// confidence. A provider-cited declared record already present in either the
// compact ledger or a connector bundle wins over an automatically generated
// unknown fallback. This makes a later rebase of a connector-specific policy
// safe without weakening the unknown-source guard for ordinary bundles.
func reconcileRateLimitSourceLedger(ledger rateLimitSourceLedger, sourceLedgerPath, defsRoot string, includePersisted bool) (rateLimitSourceLedger, bool, error) {
	changed := false
	if includePersisted {
		persisted, found, err := readExistingRateLimitSourceLedger(sourceLedgerPath)
		if err != nil {
			return rateLimitSourceLedger{}, false, err
		}
		if found {
			entries, preserved, err := preserveDeclaredSourceEntries(ledger.Entries, persisted.Entries)
			if err != nil {
				return rateLimitSourceLedger{}, false, err
			}
			ledger.Entries = entries
			changed = changed || preserved
		}
	}

	entries, preserved, err := preserveDeclaredDefinitionEntries(ledger.Entries, defsRoot)
	if err != nil {
		return rateLimitSourceLedger{}, false, err
	}
	ledger.Entries = entries
	return ledger, changed || preserved, nil
}

func readExistingRateLimitSourceLedger(sourceLedgerPath string) (rateLimitSourceLedger, bool, error) {
	if _, err := os.Stat(sourceLedgerPath); errors.Is(err, os.ErrNotExist) {
		return rateLimitSourceLedger{}, false, nil
	} else if err != nil {
		return rateLimitSourceLedger{}, false, err
	}
	ledger, err := readRateLimitSourceLedger(sourceLedgerPath)
	if err != nil {
		return rateLimitSourceLedger{}, false, err
	}
	return ledger, true, nil
}

func preserveDeclaredSourceEntries(generated, existing []rateLimitSourceEntry) ([]rateLimitSourceEntry, bool, error) {
	existingByConnector := make(map[string]rateLimitSourceEntry, len(existing))
	for _, entry := range existing {
		if _, duplicate := existingByConnector[entry.Connector]; duplicate {
			return nil, false, fmt.Errorf("existing source ledger duplicates connector %q", entry.Connector)
		}
		existingByConnector[entry.Connector] = entry
	}
	merged := append([]rateLimitSourceEntry(nil), generated...)
	changed := false
	for index, entry := range merged {
		persisted, found := existingByConnector[entry.Connector]
		if found && persisted.Declaration.State == connsdk.RateLimitStateDeclared && entry.Declaration.State != connsdk.RateLimitStateDeclared {
			merged[index] = persisted
			changed = true
		}
	}
	return merged, changed, nil
}

func preserveDeclaredDefinitionEntries(entries []rateLimitSourceEntry, defsRoot string) ([]rateLimitSourceEntry, bool, error) {
	merged := append([]rateLimitSourceEntry(nil), entries...)
	changed := false
	for index, entry := range merged {
		if entry.Declaration.State == connsdk.RateLimitStateDeclared {
			continue
		}
		destination, _, err := rateLimitDeclarationPath(defsRoot, entry.Connector)
		if err != nil {
			return nil, false, err
		}
		declared, found, err := readExistingDeclaredRateLimitDeclaration(destination)
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", entry.Connector, err)
		}
		if !found {
			continue
		}
		merged[index] = rateLimitSourceEntryForDeclaredBundle(entry.Connector, declared)
		changed = true
	}
	return merged, changed, nil
}

func readExistingDeclaredRateLimitDeclaration(destination string) (connsdk.RateLimits, bool, error) {
	raw, err := os.ReadFile(destination)
	if errors.Is(err, os.ErrNotExist) {
		return connsdk.RateLimits{}, false, nil
	}
	if err != nil {
		return connsdk.RateLimits{}, false, err
	}
	declaration, err := decodeRateLimitDeclaration(raw)
	if err != nil {
		return connsdk.RateLimits{}, false, err
	}
	if declaration.State != connsdk.RateLimitStateDeclared {
		return connsdk.RateLimits{}, false, nil
	}
	return declaration, true, nil
}

func rateLimitSourceEntryForDeclaredBundle(connector string, declaration connsdk.RateLimits) rateLimitSourceEntry {
	policySource := declaration.Policies[0].Source
	return rateLimitSourceEntry{
		Connector: connector,
		Evidence: rateLimitEvidence{
			Kind:        "preserved_declared_bundle_policy",
			URL:         policySource.URL,
			RetrievedAt: policySource.RetrievedAt,
		},
		Declaration: declaration,
	}
}

func readRateLimitSourceLedger(path string) (rateLimitSourceLedger, error) {
	var ledger rateLimitSourceLedger
	if err := readRateLimitJSON(path, &ledger); err != nil {
		return rateLimitSourceLedger{}, err
	}
	if ledger.SchemaVersion != rateLimitSourceLedgerSchemaVersion {
		return rateLimitSourceLedger{}, fmt.Errorf("schema_version = %d, want %d", ledger.SchemaVersion, rateLimitSourceLedgerSchemaVersion)
	}
	if ledger.Authority != oneFlowRateLimitAuthority {
		return rateLimitSourceLedger{}, fmt.Errorf("authority = %q, want %q", ledger.Authority, oneFlowRateLimitAuthority)
	}
	if err := validateRateLimitDate(ledger.GeneratedAt); err != nil {
		return rateLimitSourceLedger{}, fmt.Errorf("generated_at: %w", err)
	}
	if ledger.TargetTotal != len(ledger.Entries) {
		return rateLimitSourceLedger{}, fmt.Errorf("target_total = %d, entries = %d", ledger.TargetTotal, len(ledger.Entries))
	}
	if ledger.TargetTotal != 426 {
		return rateLimitSourceLedger{}, fmt.Errorf("target_total = %d, want 426", ledger.TargetTotal)
	}
	seen := make(map[string]bool, len(ledger.Entries))
	for _, entry := range ledger.Entries {
		if !rateLimitConnectorNamePattern.MatchString(entry.Connector) || seen[entry.Connector] {
			return rateLimitSourceLedger{}, fmt.Errorf("entry connector %q is invalid or duplicated", entry.Connector)
		}
		seen[entry.Connector] = true
		if err := validateRateLimitSourceEntry(entry); err != nil {
			return rateLimitSourceLedger{}, fmt.Errorf("entry %q: %w", entry.Connector, err)
		}
	}
	return ledger, nil
}

func validateRateLimitSourceEntry(entry rateLimitSourceEntry) error {
	if err := validateRateLimitDeclaration(entry.Declaration); err != nil {
		return err
	}
	if entry.Evidence.Kind != "no_provider_http_api" {
		if err := validateRateLimitReference(entry.Evidence.URL, false); err != nil {
			return fmt.Errorf("evidence source: %w", err)
		}
		if err := validateRateLimitDate(entry.Evidence.RetrievedAt); err != nil {
			return fmt.Errorf("evidence retrieved_at: %w", err)
		}
	}
	return nil
}

func validateRateLimitDeclaration(declaration connsdk.RateLimits) error {
	if declaration.SchemaVersion != 1 {
		return errors.New("declaration schema_version must be 1")
	}
	switch declaration.State {
	case connsdk.RateLimitStateDeclared:
		if len(declaration.Policies) == 0 {
			return errors.New("declared entry has no policies")
		}
		for _, policy := range declaration.Policies {
			if err := validateRateLimitReference(policy.Source.URL, true); err != nil {
				return fmt.Errorf("declared policy source: %w", err)
			}
			if err := validateRateLimitDate(policy.Source.RetrievedAt); err != nil {
				return fmt.Errorf("declared policy retrieved_at: %w", err)
			}
		}
	case connsdk.RateLimitStateUnknown, connsdk.RateLimitStateNotApplicable:
		if strings.TrimSpace(declaration.Reason) == "" || len(declaration.Policies) != 0 {
			return fmt.Errorf("state %q must have a nonblank reason and no policies", declaration.State)
		}
	default:
		return fmt.Errorf("unsupported declaration state %q", declaration.State)
	}
	return nil
}

func selectRateLimitEntries(entries []rateLimitSourceEntry, offset, size int) ([]rateLimitSourceEntry, error) {
	if offset < 0 || offset >= len(entries) {
		return nil, fmt.Errorf("offset %d is outside %d source entries", offset, len(entries))
	}
	if size < 1 || size > maxRateLimitMaterializeBatchSize {
		return nil, fmt.Errorf("size %d must be between 1 and %d", size, maxRateLimitMaterializeBatchSize)
	}
	end := offset + size
	if end > len(entries) {
		end = len(entries)
	}
	return append([]rateLimitSourceEntry(nil), entries[offset:end]...), nil
}

func materializeRateLimitBatch(sourceLedgerPath, defsRoot string, offset, size int, entries []rateLimitSourceEntry) (rateLimitBatchReport, error) {
	report := rateLimitBatchReport{
		SchemaVersion: rateLimitBatchReportSchemaVersion,
		SourceLedger:  sourceLedgerPath,
		Offset:        offset,
		RequestedSize: size,
		Entries:       make([]rateLimitBatchOutcome, 0, len(entries)),
	}
	for _, entry := range entries {
		destination, displayPath, err := rateLimitDeclarationPath(defsRoot, entry.Connector)
		if err != nil {
			return rateLimitBatchReport{}, err
		}
		status, err := writeRateLimitDeclaration(destination, entry.Declaration)
		if err != nil {
			return rateLimitBatchReport{}, fmt.Errorf("%s: %w", entry.Connector, err)
		}
		report.Entries = append(report.Entries, rateLimitBatchOutcome{
			Connector: entry.Connector,
			State:     string(entry.Declaration.State),
			File:      displayPath,
			Status:    status,
		})
		report.Counts.add(entry.Declaration.State)
	}
	return report, nil
}

func (counts *rateLimitCounts) add(state connsdk.RateLimitState) {
	counts.FileTotal++
	switch state {
	case connsdk.RateLimitStateDeclared:
		counts.Declared++
	case connsdk.RateLimitStateUnknown:
		counts.Unknown++
	case connsdk.RateLimitStateNotApplicable:
		counts.NotApplicable++
	}
}

func rateLimitDeclarationPath(defsRoot, connector string) (string, string, error) {
	if !rateLimitConnectorNamePattern.MatchString(connector) {
		return "", "", fmt.Errorf("invalid connector name %q", connector)
	}
	root, err := filepath.Abs(defsRoot)
	if err != nil {
		return "", "", fmt.Errorf("resolve defs root: %w", err)
	}
	destination := filepath.Join(root, connector, "rate_limits.json")
	relative, err := filepath.Rel(root, destination)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("rate-limit destination escapes defs root")
	}
	directory := filepath.Dir(destination)
	info, err := os.Lstat(directory)
	if err != nil {
		return "", "", fmt.Errorf("inspect connector directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", "", errors.New("connector directory must be a non-symlink directory")
	}
	return destination, filepath.ToSlash(filepath.Join("internal/connectors/defs", connector, "rate_limits.json")), nil
}

func writeRateLimitDeclaration(destination string, declaration connsdk.RateLimits) (string, error) {
	if raw, err := os.ReadFile(destination); err == nil {
		if jsonEquivalent(raw, declaration) {
			return "already_matches", nil
		}
		existing, err := decodeRateLimitDeclaration(raw)
		if err != nil {
			return "", fmt.Errorf("existing rate_limits.json is not a preservable declaration: %w", err)
		}
		if existing.State == connsdk.RateLimitStateDeclared && declaration.State != connsdk.RateLimitStateDeclared {
			return "preserved_declared", nil
		}
		if existing.State != connsdk.RateLimitStateDeclared && declaration.State == connsdk.RateLimitStateDeclared {
			if err := writeJSONAtomically(destination, declaration); err != nil {
				return "", err
			}
			return "upgraded_to_declared", nil
		}
		return "", errors.New("existing rate_limits.json differs; refusing to overwrite retained provider evidence")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read existing declaration: %w", err)
	}
	if err := writeJSONAtomically(destination, declaration); err != nil {
		return "", err
	}
	return "written", nil
}

func decodeRateLimitDeclaration(raw []byte) (connsdk.RateLimits, error) {
	if !json.Valid(raw) {
		return connsdk.RateLimits{}, errors.New("not valid JSON")
	}
	var declaration connsdk.RateLimits
	if err := json.Unmarshal(raw, &declaration); err != nil {
		return connsdk.RateLimits{}, err
	}
	if err := validateRateLimitDeclaration(declaration); err != nil {
		return connsdk.RateLimits{}, err
	}
	return declaration, nil
}

func jsonEquivalent(existing []byte, expected any) bool {
	var existingValue any
	if err := json.Unmarshal(existing, &existingValue); err != nil {
		return false
	}
	raw, err := json.Marshal(expected)
	if err != nil {
		return false
	}
	var expectedValue any
	if err := json.Unmarshal(raw, &expectedValue); err != nil {
		return false
	}
	return reflect.DeepEqual(existingValue, expectedValue)
}

func writeJSONAtomically(destination string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	raw = append(raw, '\n')
	if !json.Valid(raw) {
		return errors.New("encoded JSON is invalid")
	}
	directory := filepath.Dir(destination)
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("inspect destination directory: %w", err)
	}
	if !info.IsDir() {
		return errors.New("destination parent is not a directory")
	}
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(destination)+".tmp-")
	if err != nil {
		return fmt.Errorf("create temporary JSON: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return fmt.Errorf("set temporary JSON permissions: %w", err)
	}
	if _, err := temporary.Write(raw); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary JSON: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync temporary JSON: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary JSON: %w", err)
	}
	if err := os.Rename(temporaryPath, destination); err != nil {
		return fmt.Errorf("replace JSON atomically: %w", err)
	}
	return nil
}

func readRateLimitJSON(path string, destination any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return nil
}

func validateRateLimitDate(value string) error {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil || parsed.Format(time.DateOnly) != value {
		return errors.New("must be an ISO full date")
	}
	return nil
}

func validateRateLimitReference(raw string, requireHTTPS bool) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return errors.New("must be an absolute provider URL without userinfo")
	}
	if requireHTTPS && parsed.Scheme != "https" {
		return errors.New("must use HTTPS")
	}
	if !requireHTTPS && parsed.Scheme != "https" && parsed.Scheme != "http" {
		return errors.New("must use HTTP or HTTPS")
	}
	if hasCredentialLikeRateLimitReference(parsed) {
		return errors.New("must not carry credential-like query or fragment parameters")
	}
	return nil
}

func hasCredentialLikeRateLimitReference(parsed *url.URL) bool {
	for key := range parsed.Query() {
		if rateLimitCredentialKey(normalizeRateLimitReferenceKey(key)) {
			return true
		}
	}
	for _, part := range strings.FieldsFunc(parsed.Fragment, func(r rune) bool {
		return r == '&' || r == ';' || r == '?'
	}) {
		key, _, hasValue := strings.Cut(part, "=")
		if hasValue && rateLimitCredentialKey(normalizeRateLimitReferenceKey(key)) {
			return true
		}
	}
	return false
}

func normalizeRateLimitReferenceKey(value string) string {
	return strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(value)))
}

func rateLimitCredentialKey(value string) bool {
	switch value {
	case "accesskey", "accesstoken", "apikey", "apitoken", "authorization", "authtoken", "bearertoken", "clientsecret", "credential", "credentials", "idtoken", "key", "password", "privatekey", "refreshtoken", "secret", "secretkey", "sig", "signature", "token":
		return true
	default:
		return false
	}
}
