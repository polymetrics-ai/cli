package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/warehouse"
)

// Independently read both physical formats. Expected records come from the
// carrier's seeded source, never from its result counts or receipt flags.
func transportPhysicalRows173(ctx context.Context, wal, parquet, walHash, parquetHash string, expected []map[string]any) error {
	raw, err := os.ReadFile(wal)
	if err != nil {
		return err
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != walHash {
		return fmt.Errorf("WAL hash differs")
	}
	table, err := os.ReadFile(parquet)
	if err != nil {
		return err
	}
	if fmt.Sprintf("%x", sha256.Sum256(table)) != parquetHash {
		return fmt.Errorf("Parquet hash differs")
	}
	var walRows []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	for {
		var record struct {
			Record map[string]any `json:"record"`
		}
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if record.Record == nil {
			return fmt.Errorf("missing WAL record")
		}
		walRows = append(walRows, record.Record)
	}
	var tableRows []map[string]any
	if err := warehouse.ReadTable(ctx, parquet, func(row warehouse.Row) error { tableRows = append(tableRows, row); return nil }); err != nil {
		return err
	}
	canonical := func(rows []map[string]any) ([]string, error) {
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			raw, err := json.Marshal(row)
			if err != nil {
				return nil, err
			}
			out = append(out, string(raw))
		}
		sort.Strings(out)
		return out, nil
	}
	want, err := canonical(expected)
	if err != nil {
		return err
	}
	gotWAL, err := canonical(walRows)
	if err != nil {
		return err
	}
	gotTable, err := canonical(tableRows)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(want, gotWAL) {
		return fmt.Errorf("WAL source records differ: got %d want %d", len(gotWAL), len(want))
	}
	if !reflect.DeepEqual(want, gotTable) {
		return fmt.Errorf("Parquet source records differ: got %d want %d", len(gotTable), len(want))
	}
	return nil
}

func TestTransportPhysicalOracle173(t *testing.T) {
	expected := []map[string]any{{"id": 1, "label": "original"}, {"id": 2, "label": "second"}}
	for _, fault := range []string{"healthy", "wrong_wal_hash", "wrong_parquet_hash", "same_count_wrong_wal", "same_count_wrong_parquet", "missing_row"} {
		t.Run(fault, func(t *testing.T) {
			dir := t.TempDir()
			wal := filepath.Join(dir, "page.jsonl")
			parquet := filepath.Join(dir, "page.parquet")
			var raw bytes.Buffer
			encoder := json.NewEncoder(&raw)
			for i, row := range expected {
				if fault == "same_count_wrong_wal" && i == 0 {
					row = map[string]any{"id": 1, "label": "other"}
				}
				if err := encoder.Encode(map[string]any{"record": row}); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(wal, raw.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			writer, err := warehouse.NewTableWriter(parquet)
			if err != nil {
				t.Fatal(err)
			}
			defer writer.Abort()
			for i, row := range expected {
				if fault == "missing_row" && i == 1 {
					continue
				}
				if fault == "same_count_wrong_parquet" && i == 0 {
					row = map[string]any{"id": 1, "label": "other"}
				}
				if err := writer.Write(row); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Commit(t.Context()); err != nil {
				t.Fatal(err)
			}
			table, err := os.ReadFile(parquet)
			if err != nil {
				t.Fatal(err)
			}
			wh, ph := fmt.Sprintf("%x", sha256.Sum256(raw.Bytes())), fmt.Sprintf("%x", sha256.Sum256(table))
			if fault == "wrong_wal_hash" {
				wh = ph
			}
			if fault == "wrong_parquet_hash" {
				ph = wh
			}
			err = transportPhysicalRows173(t.Context(), wal, parquet, wh, ph, expected)
			if (err == nil) != (fault == "healthy") {
				t.Fatalf("physical oracle fault=%s error=%v", fault, err)
			}
		})
	}
}

// Receipt retirement is the sole allowed stream-state change on this refusal.
// All unknown/sibling fields remain part of the exact comparison.
func transportRetirement173(before, after string, filesBefore, filesAfter map[string]string, wantReceipts int) error {
	var oldState, newState map[string]map[string]any
	if err := json.Unmarshal([]byte(before), &oldState); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(after), &newState); err != nil {
		return err
	}
	expectedFiles := make(map[string]string, len(filesBefore))
	for path, hash := range filesBefore {
		expectedFiles[path] = hash
	}
	retired := 0
	for _, state := range oldState {
		value, exists := state["committed_transport_receipts"]
		if !exists {
			continue
		}
		receipts, ok := value.([]any)
		if !ok {
			return fmt.Errorf("invalid receipt list")
		}
		if len(receipts) == 0 {
			continue
		}
		if active, ok := state["active_work_id"]; ok && active != "" {
			return fmt.Errorf("live work cannot retire")
		}
		for _, value := range receipts {
			receipt, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid receipt")
			}
			id, _ := receipt["receipt_id"].(string)
			owner, _ := receipt["owner"].(string)
			hash, _ := receipt["manifest_sha256"].(string)
			if id == "" || owner == "" || hash == "" {
				return fmt.Errorf("receipt identity missing")
			}
			matches := 0
			for path, digest := range filesBefore {
				if filepath.Base(path) != id+".json" || filepath.Base(filepath.Dir(path)) != "transport" || filepath.Base(filepath.Dir(filepath.Dir(path))) != owner {
					continue
				}
				if digest != hash {
					return fmt.Errorf("committed manifest hash mismatch")
				}
				base := filepath.Dir(filepath.Dir(path))
				required := []string{path, filepath.Join(base, "wal", "transport-"+id+".jsonl"), filepath.Join(base, "tables", "transport-"+id+".parquet")}
				for _, file := range required {
					if _, ok := expectedFiles[file]; !ok {
						return fmt.Errorf("retired artifact missing or duplicated")
					}
					delete(expectedFiles, file)
				}
				matches++
			}
			if matches != 1 {
				return fmt.Errorf("receipt has %d structural manifests", matches)
			}
			retired++
		}
		delete(state, "committed_transport_receipts")
	}
	if retired != wantReceipts {
		return fmt.Errorf("retired %d receipts want %d", retired, wantReceipts)
	}
	if !reflect.DeepEqual(oldState, newState) {
		return fmt.Errorf("checkpoint or unrelated stream state changed")
	}
	if !reflect.DeepEqual(expectedFiles, filesAfter) {
		return fmt.Errorf("warehouse changed beyond exact committed worksets")
	}
	return nil
}

func transportWarehouseFiles173(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(filepath.Join(root, ".polymetrics", "warehouse"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("nonregular warehouse artifact")
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[path] = fmt.Sprintf("%x", sha256.Sum256(raw))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestTransportRetirementOracle173(t *testing.T) {
	before := `{"stream":{"checkpoint":{"cursor":1001},"active_work_id":"","sibling":"retained","committed_transport_receipts":[{"receipt_id":"stage-a","owner":"conn-a","manifest_sha256":"manifest"}]}}`
	after := `{"stream":{"checkpoint":{"cursor":1001},"active_work_id":"","sibling":"retained"}}`
	initial := map[string]string{"w/postgres/conn-a/transport/stage-a.json": "manifest", "w/postgres/conn-a/wal/transport-stage-a.jsonl": "wal", "w/postgres/conn-a/tables/transport-stage-a.parquet": "parquet", "w/postgres/conn-a/owner.json": "owner"}
	for _, fault := range []string{"healthy", "checkpoint_advanced", "sibling_changed", "retained_workset", "unrelated_deleted", "wrong_manifest", "wrong_owner", "new_artifact"} {
		t.Run(fault, func(t *testing.T) {
			b, a := before, after
			oldFiles := map[string]string{}
			for k, v := range initial {
				oldFiles[k] = v
			}
			newFiles := map[string]string{"w/postgres/conn-a/owner.json": "owner"}
			switch fault {
			case "checkpoint_advanced":
				a = strings.Replace(a, "1001", "1002", 1)
			case "sibling_changed":
				a = strings.Replace(a, "retained", "changed", 1)
			case "retained_workset":
				newFiles["w/postgres/conn-a/wal/transport-stage-a.jsonl"] = "wal"
			case "unrelated_deleted":
				delete(newFiles, "w/postgres/conn-a/owner.json")
			case "wrong_manifest":
				oldFiles["w/postgres/conn-a/transport/stage-a.json"] = "other"
			case "wrong_owner":
				b = strings.Replace(b, "conn-a", "conn-b", 1)
			case "new_artifact":
				newFiles["w/postgres/conn-a/wal/new.jsonl"] = "new"
			}
			err := transportRetirement173(b, a, oldFiles, newFiles, 1)
			if (err == nil) != (fault == "healthy") {
				t.Fatalf("fault=%s error=%v", fault, err)
			}
		})
	}
}

func TestTransportWarehouseFiles173(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".polymetrics", "warehouse", "workspace", "postgres", "connection")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	owner := filepath.Join(dir, "owner.json")
	table := filepath.Join(dir, "rows.parquet")
	if err := os.WriteFile(owner, []byte("owner-a"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(table, []byte("rows-a"), 0600); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{owner: fmt.Sprintf("%x", sha256.Sum256([]byte("owner-a"))), table: fmt.Sprintf("%x", sha256.Sum256([]byte("rows-a")))}
	if got := transportWarehouseFiles173(t, root); !reflect.DeepEqual(got, want) {
		t.Fatalf("exact physical inventory: %v", got)
	}
	if err := os.WriteFile(table, []byte("rows-b"), 0600); err != nil {
		t.Fatal(err)
	}
	changed := transportWarehouseFiles173(t, root)
	if changed[owner] != want[owner] || changed[table] == want[table] {
		t.Fatal("same-sized row mutation or unchanged owner lost")
	}
	extra := filepath.Join(dir, "unrelated")
	if err := os.WriteFile(extra, []byte("extra"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := transportWarehouseFiles173(t, root); len(got) != 3 || got[extra] == "" {
		t.Fatal("unrelated added file not observed")
	}
}
