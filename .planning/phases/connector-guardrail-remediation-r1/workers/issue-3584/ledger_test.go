package issue3584

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

type remediationLedger struct {
	Issue       int                `json:"issue"`
	ParentIssue int                `json:"parent_issue"`
	Spawn       string             `json:"spawn_decision"`
	Safety      ledgerSafety       `json:"safety"`
	PRs         []ledgerPR         `json:"prs"`
	Entries     []remediationEntry `json:"entries"`
}

type ledgerSafety struct {
	ForwardOnly      bool `json:"forward_only"`
	NoHistoryRewrite bool `json:"no_history_rewrite"`
	NoForcePush      bool `json:"no_force_push"`
	NoBlanketRevert  bool `json:"no_blanket_revert"`
	NoSecrets        bool `json:"no_secrets"`
}

type ledgerPR struct {
	Number   int    `json:"number"`
	URL      string `json:"url"`
	MergeSHA string `json:"merge_sha"`
}

type remediationEntry struct {
	PRNumber             int      `json:"pr_number"`
	PRURL                string   `json:"pr_url"`
	MergeSHA             string   `json:"merge_sha"`
	Path                 string   `json:"path"`
	PathClass            string   `json:"path_class"`
	ConnectorLaneAllowed bool     `json:"connector_lane_allowed"`
	ConnectorLaneVerdict string   `json:"connector_lane_verdict"`
	Disposition          string   `json:"disposition"`
	Owner                string   `json:"owner"`
	Evidence             []string `json:"evidence"`
}

func TestRemediationLedgerCoversHubSpotAndBitbucketSharedPaths(t *testing.T) {
	ledgerBytes, err := os.ReadFile("remediation-ledger.json")
	if err != nil {
		t.Fatalf("read remediation-ledger.json: %v", err)
	}
	var ledger remediationLedger
	if err := json.Unmarshal(ledgerBytes, &ledger); err != nil {
		t.Fatalf("decode remediation-ledger.json: %v", err)
	}
	if ledger.Issue != 3584 || ledger.ParentIssue != 3579 || ledger.Spawn != "spawned" {
		t.Fatalf("unexpected ledger identity: issue=%d parent=%d spawn=%q", ledger.Issue, ledger.ParentIssue, ledger.Spawn)
	}
	if !ledger.Safety.ForwardOnly || !ledger.Safety.NoHistoryRewrite || !ledger.Safety.NoForcePush || !ledger.Safety.NoBlanketRevert || !ledger.Safety.NoSecrets {
		t.Fatalf("ledger safety flags must all be true: %+v", ledger.Safety)
	}

	prs := map[int]ledgerPR{}
	for _, pr := range ledger.PRs {
		prs[pr.Number] = pr
	}
	for number, want := range map[int]ledgerPR{
		3529: {URL: "https://github.com/polymetrics-ai/cli/pull/3529", MergeSHA: "41a00398a88db809b4e799a59fea381ace5cc06e"},
		3531: {URL: "https://github.com/polymetrics-ai/cli/pull/3531", MergeSHA: "bfe785464d04fd73dba0c4a70f36e23dd84da3d0"},
	} {
		got, ok := prs[number]
		if !ok {
			t.Fatalf("ledger missing PR #%d", number)
		}
		if got.URL != want.URL || got.MergeSHA != want.MergeSHA {
			t.Fatalf("PR #%d metadata = %+v, want URL %q SHA %q", number, got, want.URL, want.MergeSHA)
		}
	}

	entries := map[string]remediationEntry{}
	for _, entry := range ledger.Entries {
		key := entryKey(entry.PRNumber, entry.Path)
		if _, exists := entries[key]; exists {
			t.Fatalf("duplicate ledger entry for %s", key)
		}
		entries[key] = entry
	}
	for _, expected := range expectedSharedPaths() {
		entry, ok := entries[entryKey(expected.prNumber, expected.path)]
		if !ok {
			t.Fatalf("missing disposition for PR #%d path %s", expected.prNumber, expected.path)
		}
		assertDispositionComplete(t, entry, expected)
	}
	if len(entries) < len(expectedSharedPaths()) {
		t.Fatalf("ledger entries = %d, want at least %d", len(entries), len(expectedSharedPaths()))
	}
}

func assertDispositionComplete(t *testing.T, entry remediationEntry, expected expectedPath) {
	t.Helper()
	if entry.PRURL != expected.prURL || entry.MergeSHA != expected.mergeSHA {
		t.Fatalf("%s PR metadata mismatch: %+v", entry.Path, entry)
	}
	if entry.PathClass != expected.pathClass {
		t.Fatalf("%s path_class = %q, want %q", entry.Path, entry.PathClass, expected.pathClass)
	}
	if entry.ConnectorLaneAllowed {
		t.Fatalf("%s must not be allowed in connector-lane scope", entry.Path)
	}
	if entry.ConnectorLaneVerdict == "" {
		t.Fatalf("%s missing connector_lane_verdict", entry.Path)
	}
	if entry.Disposition == "" {
		t.Fatalf("%s missing disposition", entry.Path)
	}
	if entry.Owner == "" {
		t.Fatalf("%s missing owner", entry.Path)
	}
	if len(entry.Evidence) == 0 {
		t.Fatalf("%s missing evidence", entry.Path)
	}
}

type expectedPath struct {
	prNumber  int
	prURL     string
	mergeSHA  string
	path      string
	pathClass string
}

func expectedSharedPaths() []expectedPath {
	const hubspotURL = "https://github.com/polymetrics-ai/cli/pull/3529"
	const hubspotSHA = "41a00398a88db809b4e799a59fea381ace5cc06e"
	const bitbucketURL = "https://github.com/polymetrics-ai/cli/pull/3531"
	const bitbucketSHA = "bfe785464d04fd73dba0c4a70f36e23dd84da3d0"
	out := []expectedPath{
		{3529, hubspotURL, hubspotSHA, "internal/cli/agent_image_cli.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/agentmode_query_cli_test.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/certify_cli.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/cli.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/errors.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/extract_cli.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/flow_cli.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/rlm_cli.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/rlm_cli_test.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/runtime_helpers.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/schedule.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/version.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/cli/worker_cli.go", "shared runtime/tooling"},
		{3529, hubspotURL, hubspotSHA, "internal/connectors/bundleregistry/registry_test.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/app/app.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/app/app_test.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/app/types.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/app/util.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/cli/cli.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/connectors/bundleregistry/registry_test.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/connectors/engine/bundle.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/connectors/engine/connector.go", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/connectors/engine/schema/cli_surface.schema.json", "shared runtime/tooling"},
		{3531, bitbucketURL, bitbucketSHA, "internal/connectors/manifest.go", "shared runtime/tooling"},
	}
	return out
}

func entryKey(prNumber int, path string) string {
	return strconv.Itoa(prNumber) + "\x00" + path
}
