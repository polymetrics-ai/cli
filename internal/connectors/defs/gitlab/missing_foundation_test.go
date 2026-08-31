package gitlab

import "testing"

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
	wantAlias := gitLabDescriptorGapSourceIDs(t, "runtime-provider-parameter-alias-investigating-r1")
	if got := gitLabFoundationSourceIDs(t, alias); !gitLabSameStringSet(got, wantAlias) {
		t.Fatalf("GitLab provider-parameter-alias source ids = %v, want %v", got, wantAlias)
	}
	if len(byID) != 3 {
		t.Fatalf("GitLab missing-foundation ids = %v, want exactly the three concrete named gaps", byID)
	}
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
