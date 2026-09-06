package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func sourceInventoryFixture(t *testing.T, ids []string, total int) (string, sourceLaneCohort) {
	t.Helper()
	root := t.TempDir()
	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		rows = append(rows, map[string]any{"id": id, "protocol": "rest", "method": "get", "path": "/items", "source_location": "paths.items.get", "source_operation": map[string]any{"responses": map[string]any{}}})
	}
	doc := map[string]any{"schema_version": 2, "connector": "fixture", "rest": map[string]any{"operations": rows}, "counts": map[string]any{"total": total}}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	return root, sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "fixture", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: hex.EncodeToString(hash[:]), ExpectedIDs: []string{"source.a", "source.b"}, ExpectedCount: 2, Artifacts: []sourceArtifactPin{}}}}
}

func retainedSourceIDs(rows []retainedSourceOperation) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Key.ID)
	}
	sort.Strings(ids)
	return ids
}

func TestSourceInventoryExactSet(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.b", "source.a"}, 2)
	got := loadRetainedSourceInventory(context.Background(), root, cohort)
	if ids := retainedSourceIDs(got.Operations); !reflect.DeepEqual(ids, []string{"source.a", "source.b"}) {
		t.Fatalf("retained provider IDs = %v; want independently specified [source.a source.b] before availability", ids)
	}
	if len(got.Diagnostics) != 0 {
		t.Fatalf("valid retained source rejected: %+v", got.Diagnostics)
	}
	for _, row := range got.Operations {
		if !row.Observed || len(row.Node) == 0 {
			t.Fatalf("source %s lost its observed provider node", row.Key.ID)
		}
	}
}

func TestSourceInventoryMembershipCounterexample(t *testing.T) {
	for _, tc := range []struct {
		name string
		ids  []string
		code string
	}{
		{"same count substitution", []string{"source.a", "source.other"}, "source_missing"},
		{"duplicate", []string{"source.a", "source.a"}, "source_duplicate"},
		{"removed", []string{"source.a"}, "source_missing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort := sourceInventoryFixture(t, tc.ids, len(tc.ids))
			got := loadRetainedSourceInventory(context.Background(), root, cohort)
			if ids := retainedSourceIDs(got.Operations); !reflect.DeepEqual(ids, []string{"source.a", "source.b"}) {
				t.Fatalf("failure suppressed anchored rows: %v", ids)
			}
			found := false
			for _, d := range got.Diagnostics {
				if d.Code == tc.code && d.Key.Connector == "fixture" {
					found = true
				}
			}
			if !found {
				t.Fatalf("readable incorrect source accepted; want %s diagnostic, got %+v", tc.code, got.Diagnostics)
			}
		})
	}
}

func TestSourceInventoryCounts(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 99)
	got := loadRetainedSourceInventory(context.Background(), root, cohort)
	for _, d := range got.Diagnostics {
		if d.Code == "source_count_mismatch" {
			return
		}
	}
	t.Fatalf("stale retained total99 accepted despite exact expected2: %+v", got.Diagnostics)
}

func TestSourceInventoryPrimarySupplement(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	other := cohort.Inventories[0]
	other.Inventory = "docs"
	other.Class = "supplement"
	cohort.Inventories = append(cohort.Inventories, other)
	got := loadRetainedSourceInventory(context.Background(), root, cohort)
	if len(got.Operations) != 4 || len(got.Diagnostics) != 0 {
		t.Fatalf("classes were collapsed or rejected: %+v", got)
	}
	classes := map[string]int{}
	for _, row := range got.Operations {
		classes[row.Class]++
	}
	if classes["primary"] != 2 || classes["supplement"] != 2 {
		t.Fatalf("wrong class counts: %v", classes)
	}
}

func TestSourceInventoryMissingSourceNoExecutionFallback(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	if err := os.Remove(filepath.Join(root, "source.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.lock.json"), []byte(`{"operations":[{"id":"unrelated.execution"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	got := loadRetainedSourceInventory(context.Background(), root, cohort)
	if !reflect.DeepEqual(retainedSourceIDs(got.Operations), []string{"source.a", "source.b"}) {
		t.Fatal("missing source changed membership")
	}
	for _, row := range got.Operations {
		if row.Observed || len(row.Diagnostics) == 0 || row.Diagnostics[0].Code != "source_unavailable" {
			t.Fatalf("missing source hidden by execution artifact: %+v", row)
		}
	}
}

func TestSourceInventoryRawArtifactIntegrity(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	data := []byte("retained evidence")
	if err := os.WriteFile(filepath.Join(root, "raw.artifact"), data, 0600); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(data)
	cohort.Inventories[0].Artifacts = []sourceArtifactPin{{Path: "raw.artifact", SHA256: hex.EncodeToString(hash[:]), Bytes: int64(len(data))}}
	valid := loadRetainedSourceInventory(context.Background(), root, cohort)
	if len(valid.Diagnostics) != 0 {
		t.Fatalf("positive raw pin rejected: %+v", valid.Diagnostics)
	}
	if err := os.WriteFile(filepath.Join(root, "raw.artifact"), []byte("wrong but readable"), 0600); err != nil {
		t.Fatal(err)
	}
	got := loadRetainedSourceInventory(context.Background(), root, cohort)
	if len(got.Operations) != 2 {
		t.Fatal("artifact failure suppressed source")
	}
	for _, row := range got.Operations {
		if len(row.Diagnostics) == 0 || row.Diagnostics[0].Code != "source_artifact_invalid" {
			t.Fatalf("invalid raw evidence accepted: %+v", row)
		}
	}
}

func TestSourceInventoryHistoricalRestoration(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "data/connector-canon/batch1-source-lane-cohort.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cohort sourceLaneCohort
	if err := decodeStrictJSON(data, &cohort); err != nil {
		t.Fatal(err)
	}
	expectedData, err := os.ReadFile("testdata/source_lanes/batch1-expected-ids.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected struct {
		Keys []sourceOperationKey `json:"keys"`
	}
	if err := json.Unmarshal(expectedData, &expected); err != nil {
		t.Fatal(err)
	}
	got := loadRetainedSourceInventory(context.Background(), root, cohort)
	keys := make([]sourceOperationKey, 0, len(got.Operations))
	observed := map[string]int{}
	for _, row := range got.Operations {
		keys = append(keys, row.Key)
		if row.Observed {
			observed[row.Key.Connector+":"+row.Class]++
		}
	}
	sort.Slice(expected.Keys, func(i, j int) bool { return sourceKeyLess(expected.Keys[i], expected.Keys[j]) })
	if !reflect.DeepEqual(keys, expected.Keys) {
		t.Fatal("full retained source set differs from independent anchored fixture")
	}
	if len(got.Diagnostics) != 0 {
		t.Fatalf("historical source unavailable or invalid: first=%+v; total diagnostics=%d; observed Asana=%d GitLab=%d", got.Diagnostics[0], len(got.Diagnostics), observed["asana:primary"], observed["gitlab:primary"])
	}
	want := map[string]int{"asana:primary": 249, "gitlab:primary": 1752, "gitlab:supplement": 2, "bitbucket:primary": 297, "circleci:primary": 111, "dockerhub:primary": 54, "jira:primary": 617, "notion:primary": 49, "sentry:primary": 223, "stripe:primary": 589, "vercel:primary": 400}
	if !reflect.DeepEqual(observed, want) {
		t.Fatalf("observed source counts %v want independently stipulated %v", observed, want)
	}
}
