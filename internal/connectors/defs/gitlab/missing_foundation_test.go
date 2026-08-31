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
	if regex == nil || stringAt(regex, "atlas_capability") != "runtime.json-schema-surrogate-regex.v1" {
		t.Fatalf("GitLab surrogate-regex gap = %#v, want cited Atlas candidate", regex)
	}
	if got := gitLabFoundationSourceIDs(t, regex); !gitLabSameStringSet(got, map[string]bool{
		"gitlab.rest.postApiV4VulnerabilitiesVulnerabilityIdFlagsAiDetection": true,
	}) {
		t.Fatalf("GitLab surrogate-regex source ids = %v", got)
	}

	alias := byID["runtime-provider-parameter-alias-investigating-r1"]
	if alias == nil || stringAt(alias, "atlas_capability") != "runtime.provider-parameter-alias.v1" {
		t.Fatalf("GitLab provider-parameter-alias gap = %#v, want cited Atlas candidate", alias)
	}
	if got := gitLabFoundationSourceIDs(t, alias); !gitLabSameStringSet(got, map[string]bool{
		"gitlab.rest.getApiV4ProjectsProjectIdPackagesNugetV2Packages": true,
	}) {
		t.Fatalf("GitLab provider-parameter-alias source ids = %v, want only the unsupported $filter source", got)
	}
	if len(byID) != 3 {
		t.Fatalf("GitLab missing-foundation ids = %v, want exactly the three concrete named gaps", byID)
	}
}

// TestGitLabProviderAliasArtifactDispositionTracksCurrentAliasCapability keeps
// the connector-local artifacts honest after the reusable typed alias
// foundation landed. A safe alias is not itself an executable command: these
// source rows stay mapped_unproven until a declaration-owned direct-read
// artifact is materialized. The sole dollar-prefixed key remains a real
// missing foundation because it is outside the closed alias grammar.
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
	statusOutputSources := map[string]bool{
		"getApiV4GroupsIdIssuesStatistics":   true,
		"getApiV4IssuesStatistics":           true,
		"getApiV4ProjectsIdIssuesStatistics": true,
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
	for sourceID := range safeSourceIDs {
		operation := gitLabDescriptorOperation(t, descriptor, sourceID)
		queryKeys := gitLabDescriptorQueryKeys(t, operation)
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

		apiOperation := gitLabAPISurfaceOperationForSourceID(t, api, sourceID)
		if stringAt(apiOperation, "model") != "direct_read" || stringAt(apiOperation, "status") != "blocked" || apiOperation["blocked_by_default"] != true {
			t.Fatalf("GitLab safe source %q API disposition = %#v, want blocked mapped-unproven direct-read", sourceID, apiOperation)
		}
		reason := stringAt(apiOperation, "reason")
		if !strings.Contains(reason, "source_disposition=mapped_unproven") || !strings.Contains(reason, "typed_provider_query_alias=available") || strings.Contains(reason, "missing_foundation=") {
			t.Fatalf("GitLab safe source %q API reason = %q, want mapped-unproven typed-alias disposition without missing-foundation claim", sourceID, reason)
		}
		notes := stringAt(apiOperation, "notes")
		if !strings.Contains(notes, "historical_descriptor_alias_gap") || !strings.Contains(notes, "no declaration-owned direct-read artifact") {
			t.Fatalf("GitLab safe source %q API notes = %q, want retained historical-gap and non-executable boundary", sourceID, notes)
		}
		if statusOutputSources[sourceID] && !strings.Contains(notes, "retained_success_output=status") {
			t.Fatalf("GitLab status-output safe source %q API notes = %q, want retained status-output boundary", sourceID, notes)
		}
		if gitLabCLISurfaceHasSourceOperation(cli, sourceID) {
			t.Fatalf("GitLab safe source %q incorrectly gained a declaration-owned CLI command", sourceID)
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
