package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// This tests the real authoring reader, not runtime certification. The immutable
// inventories retain independent primary/supplement custody even in one render.
func TestSourceProjectionGitLabClass250(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "connectors", "defs", "gitlab")
	directory, err := vNextPublicationOpenDirectory(root, "GitLab retained sources")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := directory.Close(); err != nil {
			t.Error(err)
		}
	})
	inputs := &vNextSourceProjectionInputs{ctx: context.Background(), directory: directory, retained: map[string][]byte{}}
	projection := &vNextSourceProjection{Version: 1}
	var primaryIDs []string
	for _, item := range []struct{ id, path, hash, class string }{
		{"primary", "sources/gitlab-operation-source-lock.json", "de7010cc2a2088b8da33f50312db4d5129fd167b53ef1e028ccf72cae5c72b8d", ""},
		{"binary-docs", "sources/gitlab-binary-operation-source-lock.json", "8ccd5184448121eafe1062b76578c733798c93a3e5e3251c63ce3adae167b924", "supplement"},
	} {
		data, err := os.ReadFile(filepath.Join(root, item.path))
		if err != nil {
			t.Fatal(err)
		}
		if sourceBytesHash(data) != item.hash {
			t.Fatalf("immutable source pin changed: %s", item.path)
		}
		projection.Inventories = append(projection.Inventories, vNextSourceProjectionInventory{ID: item.id, Class: sourceProjectionInventoryClass(item.class), Path: item.path, SHA256: item.hash, Bytes: int64(len(data))})
		if item.id == "primary" {
			var retained struct {
				Rest struct {
					Operations []struct {
						ID string `json:"id"`
					} `json:"operations"`
				} `json:"rest"`
			}
			if err := json.Unmarshal(data, &retained); err != nil {
				t.Fatal(err)
			}
			for _, op := range retained.Rest.Operations {
				primaryIDs = append(primaryIDs, op.ID)
			}
		}
	}
	inventory, err := inputs.inventory(vNextSourceLock{SchemaVersion: 4, Connector: "gitlab", SourceProjection: projection})
	if err != nil {
		t.Fatal(err)
	}
	t.Run("primary_exact_identity_control", func(t *testing.T) {
		var got []string
		for _, op := range inventory.Operations {
			if op.Key.Inventory == "primary" {
				if op.Class != "primary" {
					t.Errorf("primary class=%s", op.Class)
				}
				got = append(got, op.Key.ID)
			}
		}
		sort.Strings(got)
		sort.Strings(primaryIDs)
		if len(primaryIDs) != 1752 || !reflect.DeepEqual(got, primaryIDs) {
			t.Fatal("primary source identity set changed")
		}
	})
	t.Run("separate_supplement_identity", func(t *testing.T) {
		var got []string
		for _, op := range inventory.Operations {
			if op.Key.Inventory == "binary-docs" {
				got = append(got, op.Key.ID)
				if op.Class != "supplement" {
					t.Errorf("source %s class=%q want supplement; separate inventory ID must not erase supplemental class", op.Key.ID, op.Class)
				}
			}
		}
		sort.Strings(got)
		want := []string{"gitlab.docs.generic_packages.upload_file", "gitlab.docs.repository_files.raw_download"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("supplement identities=%v want=%v", got, want)
		}
	})
	if err := inputs.revalidate(); err != nil {
		t.Fatal(err)
	}
}

func TestSourceProjectionClassGrammar250(t *testing.T) {
	for _, tc := range []struct {
		name, member string
		valid        bool
	}{
		{"legacy_omitted", "", true}, {"primary", `,"class":"primary"`, true}, {"supplement", `,"class":"supplement"`, true},
		{"null", `,"class":null`, false}, {"empty", `,"class":""`, false}, {"unknown", `,"class":"auxiliary"`, false}, {"numeric", `,"class":1`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			legacy, err := json.Marshal(minimalVNextLockForTest())
			if err != nil {
				t.Fatal(err)
			}
			var root map[string]json.RawMessage
			if err := json.Unmarshal(legacy, &root); err != nil {
				t.Fatal(err)
			}
			delete(root, "operations")
			delete(root, "schemas")
			root["source_projection"] = json.RawMessage(`{"version":1,"inventories":[{"id":"inventory","path":"sources/operations.json","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","bytes":1024` + tc.member + `}]}`)
			raw, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeVNextSourceLock(raw)
			if (err == nil) != tc.valid {
				t.Fatalf("class grammar valid=%t err=%v", tc.valid, err)
			}
		})
	}
}
