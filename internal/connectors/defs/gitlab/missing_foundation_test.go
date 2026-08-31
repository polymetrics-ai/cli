package gitlab

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

const gitLabMissingFoundationPath = "missing-foundation.json"

// TestGitLabMissingFoundationLedgerStaysSourceBound verifies that the
// connector-local gap ledger only repeats concrete, already-recorded gaps. It
// is authoring evidence, never a runtime capability switch.
func TestGitLabMissingFoundationLedgerStaysSourceBound(t *testing.T) {
	report := loadGitLabObject(t, gitLabMissingFoundationPath)
	if stringAt(report, "connector") != "gitlab" || numberAt(report, "schema_version") != 1 {
		t.Fatalf("GitLab missing-foundation identity = %#v", report)
	}
	lock := objectAt(report, "source_lock")
	lockedSource := objectAt(loadGitLabObject(t, gitLabSourceLockPath), "rest")
	if stringAt(lock, "path") != gitLabSourceLockPath ||
		stringAt(lock, "source_url") != stringAt(lockedSource, "source_url") ||
		stringAt(lock, "sha256") != stringAt(lockedSource, "sha256") {
		t.Fatalf("GitLab missing-foundation source lock = %#v, want cited current lock", lock)
	}

	foundations := mustGitLabArray(t, report["foundations"])
	byID := make(map[string]map[string]any, len(foundations))
	for _, raw := range foundations {
		foundation := mustGitLabObject(t, raw)
		id := stringAt(foundation, "id")
		if id == "" {
			t.Fatal("GitLab missing-foundation entry has no id")
		}
		if _, duplicate := byID[id]; duplicate {
			t.Fatalf("GitLab missing-foundation duplicates %q", id)
		}
		if stringAt(foundation, "state") == "implemented" {
			t.Fatalf("GitLab missing-foundation %q incorrectly claims implementation", id)
		}
		byID[id] = foundation
	}

	inbound := byID["gitlab-inbound-webhook-source-executor-r1"]
	if inbound == nil || stringAt(inbound, "atlas_capability") != "transport.sync-contract.v1" {
		t.Fatalf("GitLab inbound webhook gap = %#v, want concrete transport-gap record", inbound)
	}
	wantInbound := map[string]bool{
		"gitlab.rest.postApiV4GroupsIdHooks":   true,
		"gitlab.rest.postApiV4Hooks":           true,
		"gitlab.rest.postApiV4ProjectsIdHooks": true,
	}
	if got := gitLabFoundationSourceIDs(t, inbound); !gitLabSameStringSet(got, wantInbound) {
		t.Fatalf("GitLab inbound webhook gap source ids = %v, want %v", got, wantInbound)
	}

	regex := byID["cli-request-schema-surrogate-regex-foundation-r1"]
	if regex == nil || stringAt(regex, "state") != "resolved_exact_pattern_mapping_with_remaining_source_contract_gaps" || stringAt(regex, "atlas_capability") != "runtime.json-schema-surrogate-regex.v1" {
		t.Fatalf("GitLab surrogate-regex ledger = %#v, want resolved exact-pattern component and cited Atlas capability", regex)
	}
	if got := gitLabFoundationSourceIDs(t, regex); !gitLabSameStringSet(got, map[string]bool{
		"gitlab.rest.postApiV4VulnerabilitiesVulnerabilityIdFlagsAiDetection": true,
	}) {
		t.Fatalf("GitLab surrogate-regex source ids = %v", got)
	}
	if strings.Contains(strings.ToLower(stringAt(regex, "reason")), "cannot compile") {
		t.Fatalf("GitLab surrogate-regex ledger retains obsolete compiler blocker: %#v", regex)
	}
	remainingRegexGaps := map[string]bool{}
	for _, raw := range mustGitLabArray(t, regex["remaining_source_contract_gaps"]) {
		remainingRegexGaps[stringAt(mustGitLabObject(t, raw), "foundation")] = true
	}
	if !gitLabSameStringSet(remainingRegexGaps, map[string]bool{
		"cli-request-schema-foundation-r1":                   true,
		"source-cited-non-executable-mutation-foundation-r1": true,
	}) {
		t.Fatalf("GitLab surrogate-regex remaining source-contract gaps = %v, want dynamic-root and nonexecutable-mutation causes", remainingRegexGaps)
	}

	alias := byID["runtime-provider-parameter-alias-investigating-r1"]
	if alias == nil || stringAt(alias, "state") != "mapped_unproven_connector_local_typed_alias_grammar" || stringAt(alias, "classification") != "connector_local_mapping_restriction" || stringAt(alias, "consulted_atlas_capability") != "runtime.provider-parameter-alias.v1" {
		t.Fatalf("GitLab provider-parameter-alias ledger = %#v, want connector-local typed-alias grammar restriction", alias)
	}
	if got := gitLabFoundationSourceIDs(t, alias); !gitLabSameStringSet(got, map[string]bool{
		"gitlab.rest.getApiV4ProjectsProjectIdPackagesNugetV2Packages": true,
	}) {
		t.Fatalf("GitLab provider-parameter-alias source ids = %v, want only the unsupported $filter source", got)
	}
	if len(byID) != 3 {
		t.Fatalf("GitLab missing-foundation ids = %v, want exactly the three concrete named gaps", byID)
	}

	debts := mustGitLabArray(t, report["mapping_contract_debts"])
	if len(debts) != 1 {
		t.Fatalf("GitLab mapping-contract debts = %#v, want one semantic-method partition debt", debts)
	}
	debt := mustGitLabObject(t, debts[0])
	if stringAt(debt, "id") != "gitlab-semantic-lane-method-partition-reconciliation-r1" || stringAt(debt, "state") != "resolved_exact_source_id_partition" || !strings.Contains(stringAt(debt, "reason"), "multi-selector") {
		t.Fatalf("GitLab mapping-contract debt = %#v, want explicit non-runtime semantic/method reconciliation record", debt)
	}
}

// TestGitLabProviderAliasArtifactDispositionTracksCurrentAliasCapability keeps
// the connector-local artifacts honest after the reusable typed alias
// foundation landed. The safe bracketed-key cohort is materialized only after
// each row has an exact direct-read operation and CLI binding; the sole
// dollar-prefixed key remains connector-local mapped_unproven because it is
// outside the closed typed alias grammar.
func TestGitLabProviderAliasArtifactDispositionTracksCurrentAliasCapability(t *testing.T) {
	const aliasFoundation = "runtime-provider-parameter-alias-investigating-r1"
	const unsupportedSourceID = "getApiV4ProjectsProjectIdPackagesNugetV2Packages"

	safeSourceIDs := map[string]bool{
		"getApiV4AnalyticsCodeReview":                            true,
		"getApiV4GroupsIdDashEpics":                              true,
		"getApiV4GroupsIdEpics":                                  true,
		"getApiV4GroupsIdIssues":                                 true,
		"getApiV4GroupsIdIssuesStatistics":                       true,
		"getApiV4GroupsIdMergeRequests":                          true,
		"getApiV4IssuesStatistics":                               true,
		"getApiV4MergeRequests":                                  true,
		"getApiV4ProjectsIdDeploymentsDeploymentIdMergeRequests": true,
		"getApiV4ProjectsIdIssues":                               true,
		"getApiV4ProjectsIdIssuesStatistics":                     true,
		"getApiV4ProjectsIdMergeRequests":                        true,
		"getApiV4ProjectsIdRepositoryFilesFilePathBlame":         true,
		"getApiV4ProjectsIdVariablesKey":                         true,
	}
	descriptor := loadGitLabObject(t, gitLabDescriptorPath)
	historicalGaps := gitLabDescriptorGapSourceIDs(t, aliasFoundation)
	wantHistoricalGaps := map[string]bool{"gitlab.rest." + unsupportedSourceID: true}
	for sourceID := range safeSourceIDs {
		wantHistoricalGaps["gitlab.rest."+sourceID] = true
	}
	if !gitLabSameStringSet(historicalGaps, wantHistoricalGaps) {
		t.Fatalf("GitLab retained descriptor alias-gap history = %v, want exact safe-plus-unsupported cohort %v", historicalGaps, wantHistoricalGaps)
	}

	report := loadGitLabObject(t, gitLabMissingFoundationPath)
	var alias map[string]any
	for _, raw := range mustGitLabArray(t, report["foundations"]) {
		foundation := mustGitLabObject(t, raw)
		if stringAt(foundation, "id") == aliasFoundation {
			alias = foundation
			break
		}
	}
	if alias == nil {
		t.Fatalf("GitLab missing-foundation omits %q", aliasFoundation)
	}
	if got := gitLabFoundationSourceIDs(t, alias); !gitLabSameStringSet(got, map[string]bool{"gitlab.rest." + unsupportedSourceID: true}) {
		t.Fatalf("GitLab current alias foundation IDs = %v, want only %q", got, unsupportedSourceID)
	}

	api := loadGitLabObject(t, "api_surface.json")
	cli := loadGitLabObject(t, "cli_surface.json")
	operations := loadGitLabObject(t, "operations.json")
	for sourceID := range safeSourceIDs {
		descriptorOperation := gitLabDescriptorOperation(t, descriptor, sourceID)
		queryKeys := gitLabDescriptorQueryKeys(t, descriptorOperation)
		bracketed := false
		for _, key := range queryKeys {
			if strings.Contains(key, "[") {
				bracketed = true
			}
			if alias, ok := engine.ProviderQueryParameterCLIName(key); !ok || alias == "" {
				t.Fatalf("GitLab safe source %q query key %q alias = (%q, %t), want closed typed alias", sourceID, key, alias, ok)
			}
		}
		if !bracketed {
			t.Fatalf("GitLab safe alias source %q has no bracketed provider key", sourceID)
		}
		if err := engine.ValidateProviderQueryParameterCLINames(queryKeys); err != nil {
			t.Fatalf("GitLab safe source %q alias collision: %v", sourceID, err)
		}

		command := gitLabCommandBySourceIDAndIntent(t, cli, sourceID, "direct_read")
		if stringAt(command, "availability") != "implemented" {
			t.Fatalf("GitLab safe source %q command = %#v, want implemented direct-read artifact", sourceID, command)
		}
		operation := gitLabOperationBySourceID(t, operations, sourceID)
		if stringAt(command, "operation") != stringAt(operation, "id") {
			t.Fatalf("GitLab safe source %q command operation=%q, want %q", sourceID, stringAt(command, "operation"), stringAt(operation, "id"))
		}
		rest := objectAt(operation, "rest")
		endpoint := gitLabAPISurfaceEndpoint(t, api, stringAt(rest, "method"), stringAt(rest, "path"))
		if got := stringAt(objectAt(endpoint, "covered_by"), "direct_read"); got != stringAt(command, "path") {
			t.Fatalf("GitLab safe source %q API coverage=%q, want command %q", sourceID, got, stringAt(command, "path"))
		}
	}

	unsupported := gitLabDescriptorOperation(t, descriptor, unsupportedSourceID)
	for _, key := range gitLabDescriptorQueryKeys(t, unsupported) {
		if key == "$filter" {
			if alias, ok := engine.ProviderQueryParameterCLIName(key); ok || alias != "" {
				t.Fatalf("GitLab unsupported source %q key %q alias = (%q, %t), want no typed alias", unsupportedSourceID, key, alias, ok)
			}
			apiOperation := gitLabAPISurfaceOperationForSourceID(t, api, unsupportedSourceID)
			if !strings.Contains(stringAt(apiOperation, "reason"), "missing_foundation="+aliasFoundation) {
				t.Fatalf("GitLab unsupported source %q API reason = %q, want named alias foundation", unsupportedSourceID, stringAt(apiOperation, "reason"))
			}
			return
		}
	}
	t.Fatalf("GitLab unsupported source %q has no retained $filter key", unsupportedSourceID)
}

// TestGitLabMappingContractDebtIsSourceBound preserves the former
// method-partition discrepancy as historical evidence, while requiring the
// current enabled contract to reconcile semantic lanes by exact source IDs.
func TestGitLabMappingContractDebtIsSourceBound(t *testing.T) {
	matrix := loadGitLabObject(t, gitLabSourceLaneMatrixPath)
	lock := loadGitLabObject(t, gitLabSourceLockPath)
	contract := loadGitLabObject(t, "enabled_connector_contract.json")
	report := loadGitLabObject(t, gitLabMissingFoundationPath)

	primary := map[string]bool{}
	for _, raw := range mustGitLabArray(t, objectAt(lock, "rest")["operations"]) {
		primary["gitlab.rest."+stringAt(mustGitLabObject(t, raw), "id")] = true
	}
	semantic := map[string]int{"direct_read": 0, "direct_write": 0, "reverse_etl": 0}
	semanticPostReads := 0
	for _, raw := range mustGitLabArray(t, matrix["source_operations"]) {
		row := mustGitLabObject(t, raw)
		if !primary[stringAt(row, "source_id")] {
			continue
		}
		lanes := objectAt(row, "lanes")
		for lane := range semantic {
			if stringAt(objectAt(lanes, lane), "applicability") != "not_applicable" {
				semantic[lane]++
			}
		}
		facts := objectAt(row, "source_facts")
		if stringAt(objectAt(lanes, "direct_read"), "applicability") != "not_applicable" && strings.EqualFold(stringAt(facts, "method"), "POST") {
			semanticPostReads++
		}
	}
	if semantic["direct_read"] != 762 || semantic["direct_write"] != 990 || semantic["reverse_etl"] != 990 || semanticPostReads != 13 {
		t.Fatalf("GitLab semantic primary lanes = %+v post_reads=%d, want direct_read=762 direct_write=990 reverse_etl=990 post_reads=13", semantic, semanticPostReads)
	}

	exactPartition := map[string]map[string]any{}
	var directWriteOverlay map[string]any
	for _, raw := range mustGitLabArray(t, contract["lanes"]) {
		lane := mustGitLabObject(t, raw)
		name := stringAt(lane, "name")
		if name == "direct_read" || name == "reverse_etl" {
			exactPartition[name] = objectAt(lane, "source")
		}
		if name == "direct_write" {
			directWriteOverlay = objectAt(lane, "source")
		}
	}
	if numberAt(exactPartition["direct_read"], "expected") != semantic["direct_read"] || numberAt(exactPartition["reverse_etl"], "expected") != semantic["reverse_etl"] {
		t.Fatalf("GitLab exact source-ID contract denominators = %#v, want semantic direct_read=%d reverse_etl=%d", exactPartition, semantic["direct_read"], semantic["reverse_etl"])
	}
	for lane, source := range exactPartition {
		methods, methodsDeclared := source["methods"]
		if source["partition"] != true || len(mustGitLabArray(t, source["operation_ids"])) != numberAt(source, "expected") || (methodsDeclared && len(mustGitLabArray(t, methods)) != 0) {
			t.Fatalf("GitLab %s selector = %#v, want exact operation-ID partition without method approximation", lane, source)
		}
	}
	if directWriteOverlay == nil || numberAt(directWriteOverlay, "expected") != 382 || directWriteOverlay["partition"] != false {
		t.Fatalf("GitLab direct-write execution overlay = %#v, want non-partition declared-action expected=382", directWriteOverlay)
	}

	debts := mustGitLabArray(t, report["mapping_contract_debts"])
	debt := mustGitLabObject(t, debts[0])
	if stringAt(debt, "state") != "resolved_exact_source_id_partition" {
		t.Fatalf("GitLab mapping-contract debt state=%q, want resolved exact-source-ID partition", stringAt(debt, "state"))
	}
	semanticDebt := objectAt(debt, "semantic_primary_applicable_cells")
	if numberAt(semanticDebt, "direct_read") != semantic["direct_read"] || numberAt(semanticDebt, "direct_write") != semantic["direct_write"] || numberAt(semanticDebt, "reverse_etl") != semantic["reverse_etl"] || numberAt(debt, "semantic_post_direct_reads") != semanticPostReads {
		t.Fatalf("GitLab semantic denominator debt = %#v, want exact source-matrix facts", debt)
	}
	methodDebt := objectAt(debt, "legacy_method_partition_expected")
	if numberAt(methodDebt, "direct_read") != 749 || numberAt(methodDebt, "reverse_etl") != 1003 {
		t.Fatalf("GitLab legacy method denominator history = %#v, want direct_read=749 reverse_etl=1003", debt)
	}
	exactDebt := objectAt(debt, "exact_source_id_partition_expected")
	if numberAt(exactDebt, "direct_read") != numberAt(exactPartition["direct_read"], "expected") || numberAt(exactDebt, "reverse_etl") != numberAt(exactPartition["reverse_etl"], "expected") {
		t.Fatalf("GitLab exact source-ID denominator record = %#v, want current enabled-contract facts", debt)
	}
	if numberAt(debt, "direct_write_execution_overlay_expected") != numberAt(directWriteOverlay, "expected") {
		t.Fatalf("GitLab direct-write overlay debt = %#v, want current enabled-contract overlay", debt)
	}
}

func gitLabDescriptorOperation(t *testing.T, descriptor map[string]any, sourceID string) map[string]any {
	t.Helper()
	for _, raw := range mustGitLabArray(t, descriptor["operations"]) {
		operation := mustGitLabObject(t, raw)
		if stringAt(operation, "source_id") == sourceID {
			return operation
		}
	}
	t.Fatalf("GitLab descriptor has no source operation %q", sourceID)
	return nil
}

func gitLabDescriptorQueryKeys(t *testing.T, operation map[string]any) []string {
	t.Helper()
	keys := make([]string, 0)
	for _, raw := range mustGitLabArray(t, objectAt(operation, "request")["query"]) {
		key := stringAt(mustGitLabObject(t, raw), "name")
		if key == "" {
			t.Fatalf("GitLab descriptor source %q has empty query key", stringAt(operation, "source_id"))
		}
		keys = append(keys, key)
	}
	return keys
}

func gitLabAPISurfaceOperationForSourceID(t *testing.T, surface map[string]any, sourceID string) map[string]any {
	t.Helper()
	needle := "source_operation=" + sourceID
	for _, raw := range mustGitLabArray(t, surface["endpoints"]) {
		endpoint := mustGitLabObject(t, raw)
		operation, ok := endpoint["operation"].(map[string]any)
		if !ok {
			continue
		}
		if strings.Contains(stringAt(operation, "reason"), needle) || strings.Contains(stringAt(operation, "notes"), needle) {
			return operation
		}
	}
	t.Fatalf("GitLab API surface has no blocked source operation %q", sourceID)
	return nil
}

func gitLabCLISurfaceHasSourceOperation(surface map[string]any, sourceID string) bool {
	commands, ok := surface["commands"].([]any)
	if !ok {
		panic("GitLab CLI surface commands are not an array")
	}
	for _, raw := range commands {
		command := mustGitLabObjectNoTest(raw)
		if stringAt(command, "source_operation") == sourceID {
			return true
		}
	}
	return false
}

func gitLabFoundationSourceIDs(t *testing.T, foundation map[string]any) map[string]bool {
	t.Helper()
	result := map[string]bool{}
	for _, raw := range mustGitLabArray(t, foundation["source_ids"]) {
		result[stringAt(mustGitLabObject(t, raw), "id")] = true
	}
	return result
}

func gitLabDescriptorGapSourceIDs(t *testing.T, foundationID string) map[string]bool {
	t.Helper()
	descriptor := loadGitLabObject(t, gitLabDescriptorPath)
	result := map[string]bool{}
	for _, raw := range mustGitLabArray(t, descriptor["operations"]) {
		operation := mustGitLabObject(t, raw)
		rawGaps, ok := objectAt(operation, "runtime")["gaps"].([]any)
		if !ok {
			continue
		}
		for _, rawGap := range rawGaps {
			if stringAt(mustGitLabObject(t, rawGap), "foundation") == foundationID {
				result["gitlab.rest."+stringAt(operation, "source_id")] = true
			}
		}
	}
	return result
}

func gitLabSameStringSet(got, want map[string]bool) bool {
	if len(got) != len(want) {
		return false
	}
	for value := range want {
		if !got[value] {
			return false
		}
	}
	return true
}
