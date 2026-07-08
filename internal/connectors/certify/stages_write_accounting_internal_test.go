package certify

import "testing"

func TestGithubWriteActionInventoryAccountsForAllDeclaredActions(t *testing.T) {
	items := writeActionInventoryFor("github")
	if len(items) != 231 {
		t.Fatalf("len(items) = %d, want 231", len(items))
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
