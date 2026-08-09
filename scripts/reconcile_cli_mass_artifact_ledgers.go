// Command reconcile_cli_mass_artifact_ledgers deterministically rebuilds the
// artifact-sweep accounting ledgers from durable evidence. It intentionally
// has no network or connector execution path: target inventory, batch reports,
// reconciliation outcomes, and checked-in bundle provenance are its only
// inputs.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	defaultPhaseRoot = ".planning/phases/cli-mass-artifact-materialize-r1"
	defaultDefsRoot  = "internal/connectors/defs"
	authority        = "CAPTAIN-ORDER-unblock-to-mergeable-20260809.md"
)

var recoveredSeven = map[string]bool{
	"chatwoot":     true,
	"gmail":        true,
	"greenhouse":   true,
	"help-scout":   true,
	"jira":         true,
	"lever-hiring": true,
	"workday-rest": true,
}

// TargetLedger is deliberately small: this command treats the target manifest
// as the authority for target names and primary official-source provenance.
type TargetLedger struct {
	SchemaVersion int      `json:"schema_version"`
	Targets       []Target `json:"targets"`
}

type Target struct {
	Connector    string `json:"connector"`
	SourceStatus string `json:"source_status"`
	SourceKind   string `json:"source_kind"`
	SourceURL    string `json:"source_url"`
	RetrievedAt  string `json:"retrieved_at"`
	SourceReason string `json:"source_reason"`
}

type ReconciliationRecord struct {
	Outcomes []ReconciliationOutcome `json:"outcomes"`
}

type ReconciliationOutcome struct {
	Connector         string `json:"connector"`
	State             string `json:"state"`
	Reachability      string `json:"reachability"`
	FoundationPending *bool  `json:"foundation_pending"`
	Evidence          string `json:"evidence"`
}

type Surface struct {
	OperationLedgerVersion int               `json:"operation_ledger_version"`
	Artifacts              []SurfaceArtifact `json:"artifacts"`
	Endpoints              []SurfaceEndpoint `json:"endpoints"`
}

type SurfaceArtifact struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Kind        string `json:"kind"`
	Version     string `json:"version"`
	RetrievedAt string `json:"retrieved_at"`
	SHA256      string `json:"sha256"`
}

type SurfaceEndpoint struct {
	Provenance *SurfaceProvenance `json:"provenance"`
}

type SurfaceProvenance struct {
	SourceURL string `json:"source_url"`
}

type BatchEvent struct {
	Connector         string `json:"connector"`
	State             string `json:"state"`
	Route             string `json:"route,omitempty"`
	Stage             string `json:"stage,omitempty"`
	Reason            string `json:"reason,omitempty"`
	Evidence          string `json:"evidence"`
	Reachability      string `json:"reachability,omitempty"`
	FoundationPending *bool  `json:"foundation_pending,omitempty"`
}

type buildOptions struct {
	PhaseRoot            string
	DefsRoot             string
	ExpectedMaterialized int
}

type buildResult struct {
	Materialization any
	RetryQueue      any
	RunState        any
	Report          any
	Counts          map[string]int
}

func main() {
	phaseRoot := flag.String("phase-root", defaultPhaseRoot, "artifact-sweep phase directory")
	defsRoot := flag.String("defs-root", defaultDefsRoot, "connector definitions directory")
	expectedMaterialized := flag.Int("expected-materialized", -1, "fail unless this many targets are materialized; -1 disables the assertion")
	check := flag.Bool("check", false, "validate reconstruction only; do not replace files")
	flag.Parse()

	result, err := buildLedgers(buildOptions{
		PhaseRoot:            *phaseRoot,
		DefsRoot:             *defsRoot,
		ExpectedMaterialized: *expectedMaterialized,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "reconcile artifact ledgers:", err)
		os.Exit(1)
	}
	if *check {
		fmt.Printf("reconcile artifact ledgers: check passed; %d materialized, %d retry_pending, %d genuinely_blocked, %d targets\n",
			result.Counts["materialized_total"], result.Counts["retry_pending"], result.Counts["genuinely_blocked"], result.Counts["target_total"])
		return
	}

	writes := []stagedJSON{
		{Path: filepath.Join(*phaseRoot, "MATERIALIZATION-LEDGER.json"), Value: result.Materialization},
		{Path: filepath.Join(*phaseRoot, "RETRY-QUEUE.json"), Value: result.RetryQueue},
		{Path: filepath.Join(*phaseRoot, "RUN-STATE.json"), Value: result.RunState},
		{Path: filepath.Join(*phaseRoot, "LEDGER-RECONSTRUCTION-20260809.json"), Value: result.Report},
	}
	if err := stageAndCommitJSON(writes); err != nil {
		fmt.Fprintln(os.Stderr, "reconcile artifact ledgers:", err)
		os.Exit(1)
	}
	fmt.Printf("reconcile artifact ledgers: rebuilt %d targets; %d materialized, %d retry_pending, %d genuinely_blocked\n",
		result.Counts["target_total"], result.Counts["materialized_total"], result.Counts["retry_pending"], result.Counts["genuinely_blocked"])
}

func buildLedgers(opts buildOptions) (buildResult, error) {
	targetPath := filepath.Join(opts.PhaseRoot, "TARGET-LEDGER.json")
	var targets TargetLedger
	if err := readJSON(targetPath, &targets); err != nil {
		return buildResult{}, err
	}
	if len(targets.Targets) == 0 {
		return buildResult{}, errors.New("target manifest has no targets")
	}
	targetByName := make(map[string]Target, len(targets.Targets))
	for _, target := range targets.Targets {
		if strings.TrimSpace(target.Connector) == "" {
			return buildResult{}, errors.New("target manifest has a blank connector")
		}
		if _, exists := targetByName[target.Connector]; exists {
			return buildResult{}, fmt.Errorf("target manifest duplicates connector %q", target.Connector)
		}
		targetByName[target.Connector] = target
	}

	var reconciliation ReconciliationRecord
	if err := readJSON(filepath.Join(opts.PhaseRoot, "reconciled-complete-outcomes.json"), &reconciliation); err != nil {
		return buildResult{}, err
	}
	reconciledByName := make(map[string]ReconciliationOutcome, len(reconciliation.Outcomes))
	for _, outcome := range reconciliation.Outcomes {
		if _, known := targetByName[outcome.Connector]; !known {
			return buildResult{}, fmt.Errorf("reconciliation outcome %q is outside the target manifest", outcome.Connector)
		}
		reconciledByName[outcome.Connector] = outcome
	}

	events, eventFiles, err := collectBatchEvents(filepath.Join(opts.PhaseRoot, "batches"))
	if err != nil {
		return buildResult{}, err
	}
	eventsByName := make(map[string][]BatchEvent, len(targets.Targets))
	for _, event := range events {
		if _, known := targetByName[event.Connector]; known {
			eventsByName[event.Connector] = append(eventsByName[event.Connector], event)
		}
	}

	surfaces := make(map[string]Surface, len(targets.Targets))
	v2 := make(map[string]bool, len(targets.Targets))
	for _, target := range targets.Targets {
		var surface Surface
		surfacePath := filepath.Join(opts.DefsRoot, target.Connector, "api_surface.json")
		if err := readJSON(surfacePath, &surface); err != nil {
			return buildResult{}, fmt.Errorf("read current surface for %q: %w", target.Connector, err)
		}
		surfaces[target.Connector] = surface
		if surface.OperationLedgerVersion == 2 {
			v2[target.Connector] = true
		}
	}
	for connector := range recoveredSeven {
		if _, known := targetByName[connector]; !known {
			return buildResult{}, fmt.Errorf("recovered connector %q is not in the target manifest", connector)
		}
	}

	materialized := make(map[string]bool, len(v2)+len(recoveredSeven))
	for connector := range v2 {
		materialized[connector] = true
	}
	for connector := range recoveredSeven {
		materialized[connector] = true
	}
	if opts.ExpectedMaterialized >= 0 && len(materialized) != opts.ExpectedMaterialized {
		return buildResult{}, fmt.Errorf("materialized target count = %d, want %d", len(materialized), opts.ExpectedMaterialized)
	}

	entries := make([]map[string]any, 0, len(targets.Targets))
	queued := make([]map[string]any, 0, len(targets.Targets)-len(materialized))
	resolved := make([]map[string]any, 0, len(materialized))
	counts := map[string]int{
		"already_complete":   0,
		"recovered_pilot":    0,
		"recovered_seven":    0,
		"newly_materialized": 0,
		"materialized_total": 0,
		"reachable":          0,
		"foundation_pending": 0,
		"genuinely_blocked":  0,
		"retry_pending":      0,
		"remaining":          0,
		"target_total":       len(targets.Targets),
	}

	for _, target := range sortedTargets(targets.Targets) {
		source := targetSource(target)
		connectorEvents := eventsByName[target.Connector]
		reports := uniqueEventEvidence(connectorEvents)
		if materialized[target.Connector] {
			state, reachability, foundationPending, reconciliationEvidence := materializedState(target.Connector, reconciledByName, connectorEvents)
			counts[state]++
			counts["materialized_total"]++
			if reachability == "reachable" {
				counts["reachable"]++
			}
			if foundationPending {
				counts["foundation_pending"]++
			}
			evidence := lastMaterializedEvidence(connectorEvents)
			if evidence == "" {
				evidence = reconciliationEvidence
			}
			if evidence == "" {
				evidence = filepath.ToSlash(filepath.Join("internal/connectors/defs", target.Connector, "api_surface.json"))
			}
			entry := map[string]any{
				"connector":          target.Connector,
				"state":              state,
				"reachability":       reachability,
				"foundation_pending": foundationPending,
				"evidence":           evidence,
				"source":             source,
				"provenance_reports": reports,
			}
			if v2[target.Connector] {
				entry["bundle_provenance"] = surfaceProvenance(target.Connector, surfaces[target.Connector])
			} else {
				entry["bundle_provenance"] = map[string]any{
					"bundle_api_surface":       filepath.ToSlash(filepath.Join("internal/connectors/defs", target.Connector, "api_surface.json")),
					"operation_ledger_version": surfaces[target.Connector].OperationLedgerVersion,
					"recovery_basis":           "cli-sweep-seven-connector-extract-r1 recovered bundle retained by the deterministic 426 target union",
				}
			}
			entries = append(entries, entry)
			resolved = append(resolved, map[string]any{
				"connector":          target.Connector,
				"primary_source":     source,
				"provenance_reports": reports,
				"resolved_by": map[string]any{
					"state":    state,
					"evidence": evidence,
					"route":    "current validated bundle and retained batch/recovery evidence",
				},
			})
			continue
		}

		counts["retry_pending"]++
		attempts := retryAttempts(connectorEvents)
		primaryAttempt := map[string]any{
			"evidence": "TARGET-LEDGER.json",
			"stage":    "not_attempted_in_retained_batch_report",
			"reason":   "No retained per-batch drop was found; official alternative-source routes remain pending.",
		}
		if len(attempts) > 0 {
			primaryAttempt = attempts[len(attempts)-1]
		}
		queueEntry := map[string]any{
			"connector":          target.Connector,
			"primary_source":     source,
			"primary_attempt":    primaryAttempt,
			"official_routes":    retryRoutes(attempts),
			"retry_attempts":     attempts,
			"provenance_reports": reports,
		}
		queued = append(queued, queueEntry)
		entries = append(entries, map[string]any{
			"connector":          target.Connector,
			"state":              "retry_pending",
			"reachability":       "not_checked",
			"foundation_pending": false,
			"evidence":           primaryAttempt["evidence"],
			"stage":              primaryAttempt["stage"],
			"reason":             primaryAttempt["reason"],
			"source":             source,
			"provenance_reports": reports,
		})
	}
	counts["remaining"] = counts["retry_pending"]
	if err := assertCounts(counts); err != nil {
		return buildResult{}, err
	}
	if err := assertPartition(targetByName, entries, queued, resolved); err != nil {
		return buildResult{}, err
	}

	materialization := map[string]any{
		"schema_version":        1,
		"target_manifest":       "TARGET-LEDGER.json",
		"target_total":          len(targets.Targets),
		"reconstruction_report": "LEDGER-RECONSTRUCTION-20260809.json",
		"counts":                withoutTargetTotal(counts),
		"entries":               entries,
	}
	retryQueue := map[string]any{
		"schema_version":        1,
		"authority":             authority,
		"target_manifest":       "TARGET-LEDGER.json",
		"reconstruction_report": "LEDGER-RECONSTRUCTION-20260809.json",
		"queue":                 queued,
		"resolved":              resolved,
	}
	runState := map[string]any{
		"phase":                 "cli-mass-artifact-materialize-r1",
		"authority":             authority,
		"execution_mode":        "inline_manual_gsd_fallback_single_worker",
		"state":                 "materializing",
		"target_total":          len(targets.Targets),
		"counts":                withoutTargetTotal(counts),
		"reconstruction_report": "LEDGER-RECONSTRUCTION-20260809.json",
	}
	report := map[string]any{
		"schema_version": 1,
		"authority":      authority,
		"inputs": map[string]any{
			"target_manifest":         "TARGET-LEDGER.json",
			"reconciliation_outcomes": "reconciled-complete-outcomes.json",
			"batch_report_files":      eventFiles,
			"current_defs_root":       filepath.ToSlash(opts.DefsRoot),
		},
		"v2_bundle_count": len(v2),
		"recovered_seven": sortedRecoveredSeven(),
		"counts":          withoutTargetTotal(counts),
		"assertions": map[string]any{
			"unique_target_names":                          len(targetByName),
			"materialized_plus_retry_pending_plus_blocked": counts["materialized_total"] + counts["retry_pending"] + counts["genuinely_blocked"],
			"target_total":                                 len(targets.Targets),
			"queue_entries":                                len(queued),
			"resolved_entries":                             len(resolved),
			"lost_provenance":                              false,
		},
	}
	return buildResult{Materialization: materialization, RetryQueue: retryQueue, RunState: runState, Report: report, Counts: counts}, nil
}

func materializedState(connector string, reconciled map[string]ReconciliationOutcome, events []BatchEvent) (string, string, bool, string) {
	if outcome, ok := reconciled[connector]; ok {
		foundation := false
		if outcome.FoundationPending != nil {
			foundation = *outcome.FoundationPending
		}
		return outcome.State, firstNonBlank(outcome.Reachability, "pending_static_gate"), foundation, outcome.Evidence
	}
	if recoveredSeven[connector] {
		return "recovered_seven", "reachable", false, "cli-sweep-seven-connector-extract-r1 recovered bundle"
	}
	foundation := true
	// These two Discovery bundles have retained executable foundations; their
	// batch outcomes intentionally omitted foundation_pending rather than
	// claiming the normal migration fallback.
	if connector == "google-search-console" || connector == "google-webfonts" {
		foundation = false
	}
	return "newly_materialized", reachabilityFromEvents(events), foundation, ""
}

func targetSource(target Target) map[string]any {
	return map[string]any{
		"kind":          target.SourceKind,
		"url":           target.SourceURL,
		"retrieved_at":  target.RetrievedAt,
		"source_status": target.SourceStatus,
		"source_reason": target.SourceReason,
	}
}

func surfaceProvenance(connector string, surface Surface) map[string]any {
	sources := make([]string, 0, len(surface.Endpoints))
	seen := map[string]bool{}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Provenance == nil || strings.TrimSpace(endpoint.Provenance.SourceURL) == "" || seen[endpoint.Provenance.SourceURL] {
			continue
		}
		seen[endpoint.Provenance.SourceURL] = true
		sources = append(sources, endpoint.Provenance.SourceURL)
	}
	sort.Strings(sources)
	return map[string]any{
		"bundle_api_surface":       filepath.ToSlash(filepath.Join("internal/connectors/defs", connector, "api_surface.json")),
		"operation_ledger_version": surface.OperationLedgerVersion,
		"endpoint_count":           len(surface.Endpoints),
		"artifacts":                surface.Artifacts,
		"source_urls":              sources,
	}
}

func collectBatchEvents(root string) ([]BatchEvent, []string, error) {
	files, err := filepath.Glob(filepath.Join(root, "*.json"))
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(files)
	events := make([]BatchEvent, 0)
	paths := make([]string, 0, len(files))
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			return nil, nil, err
		}
		var document map[string]json.RawMessage
		if err := json.Unmarshal(raw, &document); err != nil {
			return nil, nil, fmt.Errorf("decode batch evidence %s: %w", file, err)
		}
		evidence := filepath.ToSlash(filepath.Join("batches", filepath.Base(file)))
		paths = append(paths, evidence)
		events = append(events, eventsFromDocument(document, evidence)...)
	}
	return events, paths, nil
}

func eventsFromDocument(document map[string]json.RawMessage, evidence string) []BatchEvent {
	events := make([]BatchEvent, 0)
	for _, key := range []string{"outcomes", "included", "dropped"} {
		raw, ok := document[key]
		if !ok {
			continue
		}
		var values []map[string]json.RawMessage
		if json.Unmarshal(raw, &values) != nil {
			continue
		}
		for _, value := range values {
			connector := rawString(value["connector"])
			if connector == "" {
				continue
			}
			state := rawString(value["state"])
			if state == "" {
				switch key {
				case "included":
					state = "materialized"
				case "dropped":
					state = "retry_pending"
				}
			}
			if state == "" {
				continue
			}
			events = append(events, BatchEvent{
				Connector:         connector,
				State:             state,
				Route:             rawString(value["route"]),
				Stage:             rawString(value["stage"]),
				Reason:            rawString(value["reason"]),
				Evidence:          evidence,
				Reachability:      rawString(value["reachability"]),
				FoundationPending: rawBoolPointer(value["foundation_pending"]),
			})
		}
	}
	return events
}

func retryAttempts(events []BatchEvent) []map[string]any {
	attempts := make([]map[string]any, 0)
	for _, event := range events {
		if event.State != "retry_pending" {
			continue
		}
		attempts = append(attempts, map[string]any{
			"route":    firstNonBlank(event.Route, "retained batch source attempt"),
			"status":   "attempted_failed",
			"evidence": event.Evidence,
			"stage":    firstNonBlank(event.Stage, "artifact_inventory_unknown"),
			"reason":   firstNonBlank(event.Reason, "retained batch report recorded retry_pending"),
		})
	}
	return attempts
}

func retryRoutes(attempts []map[string]any) []map[string]any {
	status := "pending"
	var evidence any
	if len(attempts) > 0 {
		status = "attempted_failed"
		evidence = attempts[len(attempts)-1]["evidence"]
	}
	routes := []map[string]any{{
		"route":  "primary ledger artifact",
		"status": status,
	}}
	if evidence != nil {
		routes[0]["evidence"] = evidence
	}
	routes = append(routes,
		map[string]any{"route": "provider-published OpenAPI/Swagger or versioned API export", "status": "pending"},
		map[string]any{"route": "official Postman collection or SDK endpoint source", "status": "pending"},
		map[string]any{"route": "official API reference traversal", "status": "pending"},
	)
	return routes
}

func uniqueEventEvidence(events []BatchEvent) []string {
	seen := map[string]bool{}
	values := make([]string, 0, len(events))
	for _, event := range events {
		if event.Evidence == "" || seen[event.Evidence] {
			continue
		}
		seen[event.Evidence] = true
		values = append(values, event.Evidence)
	}
	return values
}

func lastMaterializedEvidence(events []BatchEvent) string {
	for index := len(events) - 1; index >= 0; index-- {
		if events[index].State == "materialized" && events[index].Evidence != "" {
			return events[index].Evidence
		}
	}
	return ""
}

func reachabilityFromEvents(events []BatchEvent) string {
	for index := len(events) - 1; index >= 0; index-- {
		if events[index].Reachability == "reachable" {
			return "reachable"
		}
	}
	return "pending_static_gate"
}

func assertCounts(counts map[string]int) error {
	if counts["materialized_total"] != counts["already_complete"]+counts["recovered_pilot"]+counts["recovered_seven"]+counts["newly_materialized"] {
		return errors.New("materialized count components do not sum to materialized_total")
	}
	if counts["materialized_total"]+counts["retry_pending"]+counts["genuinely_blocked"] != counts["target_total"] {
		return fmt.Errorf("target conservation failed: %d materialized + %d retry_pending + %d blocked != %d targets", counts["materialized_total"], counts["retry_pending"], counts["genuinely_blocked"], counts["target_total"])
	}
	if counts["remaining"] != counts["retry_pending"] {
		return errors.New("remaining must equal retry_pending while genuinely_blocked is zero")
	}
	return nil
}

func assertPartition(targets map[string]Target, entries, queued, resolved []map[string]any) error {
	if len(entries) != len(targets) {
		return fmt.Errorf("ledger entries = %d, want %d", len(entries), len(targets))
	}
	seenEntries := map[string]bool{}
	for _, entry := range entries {
		connector, _ := entry["connector"].(string)
		if _, known := targets[connector]; !known {
			return fmt.Errorf("ledger entry %q is not a target", connector)
		}
		if seenEntries[connector] {
			return fmt.Errorf("ledger entry %q is duplicated", connector)
		}
		seenEntries[connector] = true
	}
	if len(queued)+len(resolved) != len(targets) {
		return fmt.Errorf("retry queue partition = %d queue + %d resolved, want %d", len(queued), len(resolved), len(targets))
	}
	partition := map[string]bool{}
	for _, set := range [][]map[string]any{queued, resolved} {
		for _, entry := range set {
			connector, _ := entry["connector"].(string)
			if targets[connector].Connector == "" {
				return fmt.Errorf("retry queue entry %q is not a target", connector)
			}
			if partition[connector] {
				return fmt.Errorf("retry queue partition duplicates %q", connector)
			}
			partition[connector] = true
		}
	}
	return nil
}

func sortedTargets(targets []Target) []Target {
	sorted := append([]Target(nil), targets...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Connector < sorted[j].Connector })
	return sorted
}

func sortedRecoveredSeven() []string {
	values := make([]string, 0, len(recoveredSeven))
	for connector := range recoveredSeven {
		values = append(values, connector)
	}
	sort.Strings(values)
	return values
}

func withoutTargetTotal(counts map[string]int) map[string]int {
	copy := make(map[string]int, len(counts)-1)
	for key, value := range counts {
		if key != "target_total" {
			copy[key] = value
		}
	}
	return copy
}

func readJSON(path string, value any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, value); err != nil {
		return err
	}
	return nil
}

func rawString(raw json.RawMessage) string {
	var value string
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return value
}

func rawBoolPointer(raw json.RawMessage) *bool {
	var value bool
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return &value
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

type stagedJSON struct {
	Path  string
	Value any
}

type stagedFile struct {
	Path string
	Temp string
}

// stageAndCommitJSON guarantees that an invalid or empty candidate never
// truncates an existing destination. Every replacement is written, synced,
// re-read, and JSON-validated before any atomic rename starts.
func stageAndCommitJSON(writes []stagedJSON) error {
	staged := make([]stagedFile, 0, len(writes))
	for _, write := range writes {
		temporary, err := stageValidatedJSON(write.Path, write.Value)
		if err != nil {
			for _, created := range staged {
				_ = os.Remove(created.Temp)
			}
			return err
		}
		staged = append(staged, stagedFile{Path: write.Path, Temp: temporary})
	}
	for _, file := range staged {
		if err := os.Rename(file.Temp, file.Path); err != nil {
			return fmt.Errorf("atomic replace %s: %w", file.Path, err)
		}
	}
	return nil
}

func stageValidatedJSON(destination string, value any) (string, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode %s: %w", destination, err)
	}
	raw = append(raw, '\n')
	return stageValidatedJSONBytes(destination, raw)
}

func stageValidatedJSONBytes(destination string, raw []byte) (string, error) {
	if len(strings.TrimSpace(string(raw))) == 0 || !json.Valid(raw) {
		return "", fmt.Errorf("refuse invalid or empty JSON candidate for %s", destination)
	}
	directory := filepath.Dir(destination)
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(destination)+".tmp-")
	if err != nil {
		return "", err
	}
	cleanup := func(err error) (string, error) {
		_ = temporary.Close()
		_ = os.Remove(temporary.Name())
		return "", err
	}
	if _, err := temporary.Write(raw); err != nil {
		return cleanup(err)
	}
	if err := temporary.Sync(); err != nil {
		return cleanup(err)
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporary.Name())
		return "", err
	}
	verify, err := os.ReadFile(temporary.Name())
	if err != nil {
		_ = os.Remove(temporary.Name())
		return "", err
	}
	if len(strings.TrimSpace(string(verify))) == 0 || !json.Valid(verify) {
		_ = os.Remove(temporary.Name())
		return "", fmt.Errorf("temporary candidate for %s failed JSON verification", destination)
	}
	return temporary.Name(), nil
}
