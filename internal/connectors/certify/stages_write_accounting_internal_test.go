package certify

import "testing"

func TestGithubWriteActionInventoryAccountsForAllDeclaredActions(t *testing.T) {
	items, err := writeActionInventoryFor("github")
	if err != nil {
		t.Fatalf("writeActionInventoryFor(github): %v", err)
	}
	if len(items) != 607 {
		t.Fatalf("len(items) = %d, want 607", len(items))
	}

	byAction := map[string]writeActionInventoryItem{}
	for _, item := range items {
		if item.Action == "" {
			t.Fatalf("inventory item has empty action: %+v", item)
		}
		byAction[item.Action] = item
	}

	for _, action := range []string{"create_issue", "create_label", "create_milestone"} {
		item, ok := byAction[action]
		if !ok {
			t.Fatalf("missing inventory action %q", action)
		}
		if item.Pairing.Create != action || item.Pairing.Cleanup == "" {
			t.Fatalf("%s pairing = %+v, want create+cleanup", action, item.Pairing)
		}
	}

	for _, action := range []string{"update_issue", "delete_release_asset", "merge_pull_request"} {
		item, ok := byAction[action]
		if !ok {
			t.Fatalf("missing inventory action %q", action)
		}
		if item.Pairing.Create != "" {
			t.Fatalf("%s pairing = %+v, want no safe create pairing", action, item.Pairing)
		}
		if item.Reason == "" {
			t.Fatalf("%s reason empty, want blocked/unpaired reason", action)
		}
	}
}

func TestWriteActionInventoryForPropagatesMissingWritesFile(t *testing.T) {
	_, err := writeActionInventoryFor("does-not-exist")
	if err == nil {
		t.Fatal("writeActionInventoryFor(missing) error = nil, want read failure")
	}
}

func TestGithubWriteActionInventoryClassifiesEveryDeferredActionWithConcretePrerequisite(t *testing.T) {
	items, err := writeActionInventoryFor("github")
	if err != nil {
		t.Fatalf("writeActionInventoryFor(github): %v", err)
	}

	wantCounts := map[string]int{
		"repository_wave_ready":                         28,
		"repository_fixture_pending":                    234,
		"gist_fixture_pending":                          9,
		"org_fixture_and_permission_pending":            217,
		"app_or_oauth_pending":                          14,
		"enterprise_trial_and_token_pending":            25,
		"primary_user_fixture_and_permission_pending":   49,
		"secondary_user_fixture_and_permission_pending": 25,
		"notification_token_and_fixture_pending":        5,
		"sacrificial_credential_pending":                1,
	}
	gotCounts := map[string]int{}
	for _, item := range items {
		if item.Classification == "" {
			t.Fatalf("%s (%s) has no classification", item.Action, item.Path)
		}
		if item.Reason == "" {
			t.Fatalf("%s (%s) has no concrete prerequisite", item.Action, item.Path)
		}
		gotCounts[item.Classification]++
	}
	if len(gotCounts) != len(wantCounts) {
		t.Fatalf("classification count = %d, want %d: got %#v", len(gotCounts), len(wantCounts), gotCounts)
	}
	for classification, want := range wantCounts {
		if got := gotCounts[classification]; got != want {
			t.Errorf("%s count = %d, want %d", classification, got, want)
		}
	}
}

func TestWriteSweepDoesNotTreatExistingNonPassResultAsLiveCoverage(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "github", Full: true, Write: true}}
	rep := &Report{Capabilities: Capabilities{WriteActions: map[string]WriteActionResult{
		"create_issue": {Result: "not_live", Reason: "stale prepared-only entry"},
	}}}
	if err := stageWriteSweepAllPairings(rc, rep); err != nil {
		t.Fatalf("stageWriteSweepAllPairings() error = %v", err)
	}
	for _, stage := range rep.Stages {
		if stage.Name != "write_sweep_create_issue" {
			continue
		}
		if stage.Passed {
			t.Fatalf("%s passed for an existing non-pass result", stage.Name)
		}
		if stage.Error == "" {
			t.Fatalf("%s error empty, want not_live reason", stage.Name)
		}
		return
	}
	t.Fatal("write_sweep_create_issue stage was not recorded")
}
