package cli_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type cp18Remaining252Selection struct {
	SourceSHA256 string            `json:"source_sha256"`
	Members      map[string]string `json:"members"`
}

// This validates the frozen test-input selection, not source support. The
// independent retained-source joins and wire assertions remain in each test.
func cp18Remaining252ValidateSelection(raw []byte, expected cp18Remaining252Selection) error {
	var fixture struct {
		SourceSHA256 string           `json:"source_sha256"`
		Cases        []map[string]any `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		return err
	}
	if fixture.SourceSHA256 != expected.SourceSHA256 || expected.SourceSHA256 == "" {
		return fmt.Errorf("fixture source identity differs")
	}
	if len(expected.Members) == 0 || len(fixture.Cases) != len(expected.Members) {
		return fmt.Errorf("fixture selection member count differs")
	}
	seen := map[string]bool{}
	for _, member := range fixture.Cases {
		sourceID, ok := member["source_id"].(string)
		if !ok || sourceID == "" {
			return fmt.Errorf("member lacks source identity")
		}
		variant, ok := member["variant"].(string)
		if !ok || variant == "" {
			return fmt.Errorf("member lacks variant")
		}
		key := sourceID + "/" + variant
		if seen[key] {
			return fmt.Errorf("duplicate member %s", key)
		}
		seen[key] = true
		expectedHash, exists := expected.Members[key]
		if !exists {
			return fmt.Errorf("unexpected member %s", key)
		}
		encoded, err := json.Marshal(member)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(encoded)
		if hex.EncodeToString(digest[:]) != expectedHash {
			return fmt.Errorf("member bytes changed %s", key)
		}
	}
	return nil
}

func cp18Remaining252CheckSelection(t *testing.T, name string, raw []byte) {
	t.Helper()
	path := filepath.Join("testdata", "batch1-cp18-gitlab", "remaining252", name+"-selection.json")
	encoded, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var expected cp18Remaining252Selection
	if err := json.Unmarshal(encoded, &expected); err != nil {
		t.Fatal(err)
	}
	if err := cp18Remaining252ValidateSelection(raw, expected); err != nil {
		t.Fatal(err)
	}
}

func TestBatch1CP18GitLabRemaining252SelectionOracle(t *testing.T) {
	// Literal expected members are independent of the mutated test candidates.
	a := `{"source_id":"source-a","variant":"json","expected_body":{"id":101}}`
	b := `{"source_id":"source-b","variant":"status","response_status":204}`
	hash := func(raw string) string {
		var x any
		if err := json.Unmarshal([]byte(raw), &x); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(x)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(encoded)
		return hex.EncodeToString(digest[:])
	}
	expected := cp18Remaining252Selection{SourceSHA256: "pinned-source", Members: map[string]string{"source-a/json": hash(a), "source-b/status": hash(b)}}
	fixture := func(source string, members ...string) []byte {
		var cases []json.RawMessage
		for _, member := range members {
			cases = append(cases, json.RawMessage(member))
		}
		raw, err := json.Marshal(map[string]any{"source_sha256": source, "cases": cases})
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	for _, tc := range []struct {
		name      string
		raw       []byte
		wantError bool
	}{
		{"healthy", fixture("pinned-source", a, b), false},
		{"allowed_member_order", fixture("pinned-source", b, a), false},
		{"allowed_json_whitespace", fixture("pinned-source", "  "+a+" \n", b), false},
		{"wrong_source", fixture("wrong-source", a, b), true},
		{"omitted_member", fixture("pinned-source", a), true},
		{"duplicate_replaces_member", fixture("pinned-source", a, a), true},
		{"wrong_identity", fixture("pinned-source", `{"source_id":"wrong","variant":"json","expected_body":{"id":101}}`, b), true},
		{"wrong_expected_bytes", fixture("pinned-source", `{"source_id":"source-a","variant":"json","expected_body":{"id":202}}`, b), true},
		{"wrong_status", fixture("pinned-source", a, `{"source_id":"source-b","variant":"status","response_status":200}`), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := cp18Remaining252ValidateSelection(tc.raw, expected)
			if (err != nil) != tc.wantError {
				t.Fatalf("oracle err=%v wantError=%t", err, tc.wantError)
			}
		})
	}
}

// Reads only the approval/receipt fields needed by the ordering oracle. It never
// returns credentials, approval material or unrelated project state to output.
func cp18Remaining252DurableFrontier(root, planID string, terminal bool) ([]byte, error) {
	raw, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		return nil, err
	}
	var state struct {
		Plans []struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			TokenHash string `json:"approval_token_hash"`
			Consumed  string `json:"approval_consumed_at"`
		} `json:"reverse_plans"`
		Runs []json.RawMessage `json:"reverse_runs"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, err
	}
	found := false
	for _, plan := range state.Plans {
		if plan.ID == planID {
			found = true
			consumed, timeErr := time.Parse(time.RFC3339Nano, plan.Consumed)
			if plan.TokenHash != "" || timeErr != nil || consumed.IsZero() {
				return nil, fmt.Errorf("physical send/terminal receipt preceded durable approval consumption")
			}
			if !terminal && plan.Status != "approval_consumption_uncertain" {
				return nil, fmt.Errorf("unexpected in-flight approval status %s", plan.Status)
			}
			if terminal && plan.Status != "executed" && plan.Status != "failed" {
				return nil, fmt.Errorf("terminal receipt lacks terminal plan status")
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("approved plan absent from durable state")
	}
	var matched []json.RawMessage
	for _, rawRun := range state.Runs {
		var run struct {
			PlanID string `json:"plan_id"`
		}
		if err := json.Unmarshal(rawRun, &run); err != nil {
			return nil, err
		}
		if run.PlanID == planID {
			matched = append(matched, rawRun)
		}
	}
	if !terminal && len(matched) != 0 {
		return nil, fmt.Errorf("terminal provider receipt exists before provider response")
	}
	if terminal && len(matched) != 1 {
		return nil, fmt.Errorf("terminal durable receipt count=%d want1", len(matched))
	}
	if !terminal {
		return nil, nil
	}
	var canonical any
	if err := json.Unmarshal(matched[0], &canonical); err != nil {
		return nil, err
	}
	return json.Marshal(canonical)
}

func TestBatch1CP18GitLabRemaining252DurableOracle(t *testing.T) {
	for _, tc := range []struct {
		name      string
		terminal  bool
		mutate    func(map[string]any)
		wantError bool
	}{
		{"healthy_before_send", false, nil, false},
		{"healthy_terminal", true, nil, false},
		{"token_not_consumed", false, func(s map[string]any) {
			s["reverse_plans"].([]any)[0].(map[string]any)["approval_token_hash"] = "still-present"
		}, true},
		{"wrong_plan_identity", false, func(s map[string]any) { s["reverse_plans"].([]any)[0].(map[string]any)["id"] = "different" }, true},
		{"invalid_consumption_time", false, func(s map[string]any) {
			s["reverse_plans"].([]any)[0].(map[string]any)["approval_consumed_at"] = "invalid"
		}, true},
		{"premature_receipt", false, func(s map[string]any) {
			s["reverse_runs"] = []any{map[string]any{"plan_id": "approved", "id": "invented"}}
		}, true},
		{"missing_terminal_receipt", true, func(s map[string]any) { s["reverse_runs"] = []any{} }, true},
		{"duplicate_terminal_receipt", true, func(s map[string]any) {
			s["reverse_runs"] = append(s["reverse_runs"].([]any), s["reverse_runs"].([]any)[0])
		}, true},
		{"wrong_terminal_status", true, func(s map[string]any) { s["reverse_plans"].([]any)[0].(map[string]any)["status"] = "previewed" }, true},
		{"allowed_unrelated_history", true, func(s map[string]any) {
			s["reverse_runs"] = append(s["reverse_runs"].([]any), map[string]any{"plan_id": "unrelated", "id": "old"})
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status := "approval_consumption_uncertain"
			runs := []any{}
			if tc.terminal {
				status = "executed"
				runs = append(runs, map[string]any{"plan_id": "approved", "id": "recorded", "records_succeeded": 1})
			}
			state := map[string]any{"reverse_plans": []any{map[string]any{"id": "approved", "status": status, "approval_consumed_at": "2026-01-02T03:04:05Z", "approval_token_hash": ""}}, "reverse_runs": runs}
			if tc.mutate != nil {
				tc.mutate(state)
			}
			root := t.TempDir()
			dir := filepath.Join(root, ".polymetrics", "state")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(state)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "state.json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			output, err := cp18Remaining252DurableFrontier(root, "approved", tc.terminal)
			if (err != nil) != tc.wantError {
				t.Fatalf("durable oracle error=%v wantError=%t", err, tc.wantError)
			}
			if !tc.wantError && tc.terminal && string(output) != `{"id":"recorded","plan_id":"approved","records_succeeded":1}` {
				t.Fatalf("oracle changed matched durable receipt: %s", output)
			}
		})
	}
}
