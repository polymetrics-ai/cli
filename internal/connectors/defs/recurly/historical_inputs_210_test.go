package recurly

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	connectorDefs "polymetrics.ai/internal/connectors/defs"
)

func historicalBytesMatch210(raw []byte, size int, digest string) bool {
	sum := sha256.Sum256(raw)
	return len(raw) == size && hex.EncodeToString(sum[:]) == digest
}

func TestRecurlyHistoricalInputsRemainExactAndTestOnly210(t *testing.T) {
	const root = "testdata/historical-210"
	var provenance struct {
		SourceCommit string `json:"source_commit"`
		Files        []struct {
			Path         string `json:"path"`
			OriginalPath string `json:"original_path"`
			Bytes        int    `json:"bytes"`
			SHA256       string `json:"sha256"`
		} `json:"files"`
	}
	raw, err := os.ReadFile(filepath.Join(root, "provenance.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &provenance); err != nil {
		t.Fatal(err)
	}
	if provenance.SourceCommit != "2008509f40c90de5d041aa944b19fedfecc15533" || len(provenance.Files) != 97 {
		t.Fatal("historical source or required input family changed")
	}
	expected := map[string]bool{"RECURLY-WRITE-RETRY-RESEARCH.json": true, "fixtures/streams/list_sites/page_2.json": true, "fixtures/writes/deactivate_account.json": true, "fixtures/writes/terminate_subscription.json": true}
	for name := range loadRecoveryStreams(t) {
		expected["fixtures/streams/"+name+"/page_1.json"] = true
	}
	for _, input := range provenance.Files {
		t.Run(input.Path, func(t *testing.T) {
			if !expected[input.Path] || input.OriginalPath == "" || !fs.ValidPath(input.Path) {
				t.Fatal("duplicate, unknown or unbound historical input")
			}
			delete(expected, input.Path)
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(input.Path)))
			if err != nil || !historicalBytesMatch210(raw, input.Bytes, input.SHA256) {
				t.Fatal("historical input differs from exact original bytes")
			}
			mutated := append([]byte(nil), raw...)
			mutated[0] ^= 1
			if historicalBytesMatch210(mutated, input.Bytes, input.SHA256) {
				t.Fatal("hash oracle accepted changed readable bytes")
			}
			if _, err := fs.Stat(connectorDefs.FS, "recurly/"+root+"/"+input.Path); !errors.Is(err, fs.ErrNotExist) {
				t.Fatal("historical test input crossed runtime embedding boundary")
			}
		})
	}
	if len(expected) != 0 {
		t.Fatal("required historical fixture membership missing")
	}
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "README.md" || rel == "provenance.json" {
			return nil
		}
		for _, input := range provenance.Files {
			if strings.ReplaceAll(rel, string(filepath.Separator), "/") == input.Path {
				return nil
			}
		}
		t.Error("untracked historical dependency outside exact closure")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
