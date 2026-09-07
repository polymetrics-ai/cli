package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// This is an exact fixture operation, not certification of a provider row.
// The independently specified IDs and wire window falsify status-only proof.
func TestSourceLaneProofFixtureBehavior(t *testing.T) {
	t.Run("bounded_records", func(t *testing.T) {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if r.Method != "GET" || r.URL.Path != "/widgets" || r.URL.Query().Get("limit") != "2" {
				t.Errorf("unexpected wire request: %s %s", r.Method, r.URL)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":"widget-A"},{"id":"widget-B"}]`))
		}))
		defer server.Close()
		bundle := engine.Bundle{Name: "proof-fixture", HTTP: engine.HTTPBase{URL: server.URL, Pagination: &engine.PaginationSpec{Type: "offset_limit", LimitParam: "limit", OffsetParam: "offset", PageSize: 2}}, Operations: []engine.OperationSpec{{ID: "fixture.widgets", Kind: "rest_read", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &engine.RESTOperationSpec{Method: "GET", Path: "/widgets", MaxBytes: 1024}}}}
		result, err := engine.OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{Operation: "fixture.widgets"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(result.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != `[{"id":"widget-A"},{"id":"widget-B"}]` || requests != 1 || result.Page.Records != 2 || result.Page.Complete {
			t.Fatalf("want exact A/B records, one bounded request, incomplete page; got %s requests=%d page=%+v", raw, requests, result.Page)
		}
	})
}

func TestSourceLaneProofBatchFixture(t *testing.T) {
	for _, group := range []struct {
		name  string
		count int
	}{{"six", 6}, {"thirtythree", 33}, {"capacity", 4097}} {
		t.Run(group.name, func(t *testing.T) {
			seen := map[string]bool{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id := strings.TrimPrefix(r.URL.Path, "/widgets/")
				if r.Method != "GET" || r.URL.Path != "/widgets/"+id || r.URL.Query().Get("limit") != "2" || seen[id] {
					t.Errorf("unexpected or duplicate fixture request %s %s", r.Method, r.URL)
				}
				seen[id] = true
				w.Header().Set("Content-Type", "application/json")
				if _, err := fmt.Fprintf(w, `[{"id":%q},{"id":%q}]`, id+"-A", id+"-B"); err != nil {
					t.Errorf("write fixture response: %v", err)
				}
			}))
			defer server.Close()
			for i := 0; i < group.count; i++ {
				id := fmt.Sprintf("fixture-%06d", i)
				b := engine.Bundle{Name: "proof-fixture", HTTP: engine.HTTPBase{URL: server.URL, Pagination: &engine.PaginationSpec{Type: "offset_limit", LimitParam: "limit", OffsetParam: "offset", PageSize: 2}}, Operations: []engine.OperationSpec{{ID: id, Kind: "rest_read", Risk: "low", Approval: "none", OutputPolicy: "json_redacted", REST: &engine.RESTOperationSpec{Method: "GET", Path: "/widgets/" + id, MaxBytes: 1024}}}}
				result, err := engine.OperationDirectRead(context.Background(), b, connectors.OperationDirectReadRequest{Operation: id}, nil)
				if err != nil {
					t.Fatal(err)
				}
				body, err := json.Marshal(result.Body)
				if err != nil {
					t.Fatal(err)
				}
				want := fmt.Sprintf(`[{"id":%q},{"id":%q}]`, id+"-A", id+"-B")
				if string(body) != want || result.Page.Records != 2 || result.Page.Complete || !seen[id] {
					t.Fatalf("operation %s: want %s, two records/incomplete page; got %s %+v", id, want, body, result.Page)
				}
			}
			if len(seen) != group.count {
				t.Fatalf("expected %d exact fixture operations, observed %d", group.count, len(seen))
			}
		})
	}
}

func proofBatch(t *testing.T, group string, count int) (string, []sourceLaneProofRecord, []sourceLaneProofReview) {
	t.Helper()
	root, base, _ := proofFixture(t)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := exec.Command("go", "tool", "test2json", "-p", base.Package, executable, "-test.v", "-test.run=^TestSourceLaneProofBatchFixture$/^"+group+"$", "-test.timeout=20m").CombinedOutput()
	if err != nil {
		t.Fatalf("actual batch fixture: %v\n%s", err, receipt)
	}
	t.Logf("Original actual %s batch fixture receipt SHA256 %s:\n%s", group, sourceBytesHash(receipt), receipt)
	base.TestSymbol = "TestSourceLaneProofBatchFixture"
	base.SelectedTest = base.TestSymbol + "/" + group
	base.ReceiptSHA256 = sourceBytesHash(receipt)
	proofWrite(t, root, base.ReceiptPath, receipt)
	operations := make([]map[string]string, 0, count)
	for i := 0; i < count; i++ {
		operations = append(operations, map[string]string{"id": fmt.Sprintf("fixture-%06d", i)})
	}
	raw, err := json.Marshal(map[string]any{"operations": operations})
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, base.Targets[0].Artifact, raw)
	records := make([]sourceLaneProofRecord, 0, count)
	reviews := make([]sourceLaneProofReview, 0, count)
	for i := 0; i < count; i++ {
		r := base
		r.ID = fmt.Sprintf("proof-%06d", i)
		r.Key.ID = fmt.Sprintf("fixture-%06d", i)
		r.Targets = append([]sourceLaneTargetRef(nil), base.Targets...)
		r.Targets[0].ID = r.Key.ID
		r.Targets[0].Pointer = fmt.Sprintf("/operations/%d", i)
		r.Targets[0].ArtifactSHA256 = sourceBytesHash(raw)
		r.Targets[0].CanonicalID = "operation:" + r.Key.ID
		r.Targets[0].CanonicalPointer = fmt.Sprintf("/operations/%d/operation", i)
		r.ObservableContract = "GET /widgets/" + r.Key.ID + " sends limit=2; exact " + r.Key.ID + "-A and " + r.Key.ID + "-B; one request; incomplete page"
		records = append(records, r)
		reviews = append(reviews, sourceLaneProofReview{Record: r, Fixture: true})
	}
	return root, records, reviews
}

func proofBatchCells(r sourceLaneProofRecord) []sourceLaneCell {
	cells := make([]sourceLaneCell, 0, 7)
	for _, lane := range sourceLaneNames() {
		cells = append(cells, sourceLaneCell{Lane: lane, Applicability: "not_applicable", State: "not_applicable"})
	}
	cells[0].Applicability = "applicable"
	cells[0].State = "mapped_unproven"
	cells[0].References = r.Targets
	return cells
}

func TestSourceLaneProofCapacity(t *testing.T) {
	for _, group := range []struct {
		name  string
		count int
	}{{"six", 6}, {"thirtythree", 33}} {
		t.Run(group.name, func(t *testing.T) {
			root, records, reviews := proofBatch(t, group.name, group.count)
			proofDocument(t, root, records)
			raw, err := os.ReadFile(filepath.Join(root, sourceLaneProofPath))
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("inline document bytes=%d closure pins=%d", len(raw), len(records[0].Inputs))
			in := proofLoadFixture(root, reviews)
			promoted := 0
			for _, r := range records {
				cells := assessSourceLaneProof(r.Key, proofBatchCells(r), in)
				if len(cells) != 7 {
					t.Fatalf("lost cells for %s", r.Key.ID)
				}
				if cells[0].State == "implemented" && reflect.DeepEqual(cells[0].ProofRefs, []string{r.ID}) {
					promoted++
				}
			}
			if promoted != group.count || len(in.Diagnostics) != 0 {
				t.Fatalf("want %d exact promoted fixture records, got %d; diagnostics %+v", group.count, promoted, in.Diagnostics)
			}
		})
	}
	t.Run("4097_shared_in_4343_anchors", func(t *testing.T) {
		root, records, _ := proofBatch(t, "capacity", 4097)
		keys := make([]sourceOperationKey, 0, 4343)
		for i := 0; i < 4343; i++ {
			keys = append(keys, sourceOperationKey{Connector: "proof-fixture", Inventory: "fixture", ID: fmt.Sprintf("fixture-%06d", i)})
		}
		policy, err := newSourceLaneProofPolicy(keys)
		if err != nil {
			t.Fatal(err)
		}
		if policy.MaxRecords != 30401 {
			t.Fatalf("trusted capacity=%d, want 30401", policy.MaxRecords)
		}
		records, catalog := proofSharedDocument(t, root, records)
		witness := []sourceLaneProofReadEvent{}
		policy.afterRead = func(e sourceLaneProofReadEvent) { witness = append(witness, e) }
		in := loadSourceLaneProofs(context.Background(), root, policy, catalog)
		promoted, cellsCount := 0, 0
		for i, key := range keys {
			r := sourceLaneProofRecord{Key: key}
			if i < len(records) {
				r = records[i]
			}
			cells := assessSourceLaneProof(key, proofBatchCells(r), in)
			cellsCount += len(cells)
			if len(cells) != 7 {
				t.Fatalf("lost seven cells at %s", key.ID)
			}
			if cells[0].State == "implemented" {
				promoted++
				if i >= 4097 || !reflect.DeepEqual(cells[0].ProofRefs, []string{fmt.Sprintf("proof-%06d", i)}) {
					t.Fatalf("wrong exact mapping %s: %+v", key.ID, cells[0])
				}
			} else if i < 4097 {
				t.Fatalf("fixture proof %s unproven: %+v", key.ID, cells[0])
			}
		}
		if promoted != 4097 || cellsCount != 30401 || len(in.Diagnostics) != 0 {
			t.Fatalf("capacity result promoted=%d cells=%d diagnostics=%+v", promoted, cellsCount, in.Diagnostics)
		}
		unique := map[string]int64{}
		for _, pin := range catalog.InputSets[0].Inputs {
			info, err := os.Stat(filepath.Join(root, pin.Path))
			if err != nil {
				t.Fatal(err)
			}
			unique[pin.Path] = info.Size()
		}
		for _, name := range []string{sourceLaneProofPath, records[0].ReceiptPath, records[0].Targets[0].Artifact} {
			info, err := os.Stat(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			unique[name] = info.Size()
		}
		var bytes int64
		for _, size := range unique {
			bytes += size
		}
		if in.Stats.UniqueFiles != len(unique) || in.Stats.ReadCalls != 2*len(unique) || in.Stats.FinalReads != len(unique) || in.Stats.UniqueBytes != bytes || in.Stats.PhysicalCharged != 2*bytes {
			t.Fatalf("want two unique content passes for %d files/%d bytes, got %+v", len(unique), bytes, in.Stats)
		}
		if !proofReadWitnessMatches(witness, unique) {
			t.Fatalf("actual post-I/O witness does not match independent unique files")
		}
		t.Logf("4097 records / 4343 anchors / 30401 cells; document=%d unique_files=%d unique_bytes=%d physical_charged=%d read_calls=%d final_reads=%d", unique[sourceLaneProofPath], in.Stats.UniqueFiles, in.Stats.UniqueBytes, in.Stats.PhysicalCharged, in.Stats.ReadCalls, in.Stats.FinalReads)
	})
}

func proofSharedDocument(t *testing.T, root string, records []sourceLaneProofRecord) ([]sourceLaneProofRecord, sourceLaneProofCatalog) {
	t.Helper()
	set := sourceLaneProofInputSet{Inputs: append([]sourceLaneProofInput(nil), records[0].Inputs...)}
	set.SHA256 = sourceLaneProofSetHash(set.Inputs)
	out := append([]sourceLaneProofRecord(nil), records...)
	catalog := sourceLaneProofCatalog{InputSets: []sourceLaneProofInputSet{set}}
	for i := range out {
		out[i].Inputs = nil
		out[i].InputSetSHA256 = set.SHA256
		catalog.Reviews = append(catalog.Reviews, sourceLaneProofReview{Record: out[i], Fixture: true})
	}
	proofSharedRaw(t, root, out, []sourceLaneProofInputSet{set})
	return out, catalog
}

func proofSharedRaw(t *testing.T, root string, records []sourceLaneProofRecord, sets []sourceLaneProofInputSet) {
	t.Helper()
	raw, err := json.Marshal(struct {
		SchemaVersion int                       `json:"schema_version"`
		Records       []sourceLaneProofRecord   `json:"records"`
		InputSets     []sourceLaneProofInputSet `json:"input_sets,omitempty"`
	}{1, records, sets})
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, sourceLaneProofPath, raw)
}

func proofReadWitnessMatches(events []sourceLaneProofReadEvent, expected map[string]int64) bool {
	counts := map[string]map[string]int{}
	for _, e := range events {
		size, ok := expected[e.Path]
		if !ok || !e.Success || e.Bytes != size || e.Before == nil || e.After == nil || !os.SameFile(e.Before, e.After) {
			return false
		}
		if counts[e.Path] == nil {
			counts[e.Path] = map[string]int{}
		}
		counts[e.Path][e.Phase]++
	}
	for name := range expected {
		if counts[name]["initial"] != 1 || counts[name]["final"] != 1 || len(counts[name]) != 2 {
			return false
		}
	}
	return true
}

func TestSourceLaneProofSharedInputs(t *testing.T) {
	root, original, _ := proofBatch(t, "six", 6)
	records, catalog := proofSharedDocument(t, root, original)
	t.Run("valid_cap_plus_one", func(t *testing.T) {
		p := proofFixturePolicy()
		p.MaxRecords = 5
		in := loadSourceLaneProofs(context.Background(), root, p, catalog)
		if !proofHasDiagnostic(in.Diagnostics, "proof_document_invalid", "error") || len(in.accepted) != 0 {
			t.Fatalf("cap+1 admitted: %+v", in)
		}
	})
	t.Run("unknown_anchored_key", func(t *testing.T) {
		p, _ := newSourceLaneProofPolicy([]sourceOperationKey{original[0].Key})
		in := loadSourceLaneProofs(context.Background(), root, p, catalog)
		if !proofHasDiagnostic(in.Diagnostics, "proof_source_unknown", "error") {
			t.Fatalf("unknown source silently accepted: %+v", in.Diagnostics)
		}
	})
	t.Run("tuple_budget", func(t *testing.T) {
		p := proofFixturePolicy()
		p.limits.Tuples = len(catalog.InputSets[0].Inputs) - 1
		in := loadSourceLaneProofs(context.Background(), root, p, catalog)
		if !proofHasDiagnostic(in.Diagnostics, "proof_document_invalid", "error") {
			t.Fatalf("tuple ceiling ignored: %+v", in.Diagnostics)
		}
	})
	for _, kind := range []string{"duplicate_set", "missing_set", "wrong_hash", "duplicate_path", "both_representations", "unreviewed_set"} {
		t.Run(kind, func(t *testing.T) {
			sets := append([]sourceLaneProofInputSet(nil), catalog.InputSets...)
			rs := append([]sourceLaneProofRecord(nil), records...)
			trusted := catalog
			switch kind {
			case "duplicate_set":
				sets = append(sets, sets[0])
			case "missing_set":
				sets = nil
			case "wrong_hash":
				sets[0].SHA256 = strings.Repeat("b", 64)
			case "duplicate_path":
				sets[0].Inputs = append(append([]sourceLaneProofInput(nil), sets[0].Inputs...), sets[0].Inputs[0])
			case "both_representations":
				rs[0].Inputs = original[0].Inputs
				trusted.Reviews = append([]sourceLaneProofReview(nil), catalog.Reviews...)
				trusted.Reviews[0].Record = rs[0]
			case "unreviewed_set":
				trusted.InputSets = nil
			}
			proofSharedRaw(t, root, rs, sets)
			in := loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), trusted)
			if len(in.Diagnostics) == 0 || len(in.accepted) != 0 && kind != "both_representations" {
				t.Fatalf("invalid shared input accepted %s: %+v", kind, in.Diagnostics)
			}
		})
	}
	proofSharedRaw(t, root, records, catalog.InputSets)
	t.Run("set_order_independent_hash", func(t *testing.T) {
		pins := append([]sourceLaneProofInput(nil), catalog.InputSets[0].Inputs...)
		for i, j := 0, len(pins)-1; i < j; i, j = i+1, j-1 {
			pins[i], pins[j] = pins[j], pins[i]
		}
		if sourceLaneProofSetHash(pins) != catalog.InputSets[0].SHA256 {
			t.Fatal("set hash depends on input order")
		}
	})
	t.Run("realistic_four_target_size", func(t *testing.T) {
		model := records[0]
		model.Targets = append([]sourceLaneTargetRef(nil), model.Targets...)
		for _, kind := range []string{"command", "schema", "canonical_operation"} {
			ref := model.Targets[0]
			ref.Kind = kind
			ref.ID = "fixture.realistic.operation.with.long.identity"
			ref.Artifact = "internal/connectors/defs/proof-fixture/cli_surface.json"
			model.Targets = append(model.Targets, ref)
		}
		raw, err := json.Marshal(model)
		if err != nil {
			t.Fatal(err)
		}
		setBytes, err := json.Marshal(catalog.InputSets)
		if err != nil {
			t.Fatal(err)
		}
		size := int64(len(raw)+1)*30401 + int64(len(setBytes)) + 64
		if size >= 128<<20 {
			t.Fatalf("realistic four-target model exceeds policy: %d", size)
		}
		t.Logf("serialized four-target 30401-record model=%d bytes (<128 MiB); sizing model only, not fabricated proof", size)
	})
}

func TestSourceLaneProofReadAccounting(t *testing.T) {
	root := t.TempDir()
	name := "internal/oversized.go"
	raw := []byte(strings.Repeat("x", 65))
	proofWrite(t, root, name, raw)
	opened, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := opened.Close(); err != nil {
			t.Errorf("close fixture root: %v", err)
		}
	}()
	before, err := opened.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	// Independent actual read witnesses prove the expected lookahead boundary.
	for i := 0; i < 2; i++ {
		got, err := readSourceInput(opened, name, 32)
		after, statErr := opened.Stat(name)
		if err == nil || got != nil || statErr != nil || !os.SameFile(before, after) || after.Size() != 65 {
			t.Fatalf("real oversized read witness did not reach expected regular file: %v", err)
		}
	}
	budget := int64(32)
	r := sourceLaneProofRecord{Inputs: []sourceLaneProofInput{{Path: name, SHA256: sourceBytesHash(raw), Role: "code"}}}
	code, _ := proofReadWithBudget(opened, r, &budget)
	if code == "" || budget >= 32 {
		t.Fatalf("failed bounded read must spend attempted allowance, got code=%s remaining=%d", code, budget)
	}
	t.Run("actual_failed_attempts_and_cached_refusal", func(t *testing.T) {
		p := proofFixturePolicy()
		p.limits.UniqueBytes = 64
		p.limits.InputBytes = 16
		events := []sourceLaneProofReadEvent{}
		p.afterRead = func(e sourceLaneProofReadEvent) { events = append(events, e) }
		cache := sourceLaneProofCache{ctx: context.Background(), root: opened, policy: p, files: map[string]*sourceLaneProofFile{}}
		for i := 0; i < 8; i++ {
			_, f := cache.get(name, 16, false)
			if f.code != "proof_input_invalid" {
				t.Fatalf("oversized error lost: %s", f.code)
			}
		}
		if len(events) != 1 || events[0].Success || events[0].Charged != 17 || events[0].Before.Size() != 65 || !os.SameFile(before, events[0].Before) || cache.stats.UniqueBytes != 16 || cache.stats.PhysicalCharged != 17 {
			t.Fatalf("failed-attempt witness/accounting incorrect: events=%+v stats=%+v", events, cache.stats)
		}
		for i := 0; i < 8; i++ {
			next := fmt.Sprintf("internal/oversized-%d.go", i)
			proofWrite(t, root, next, raw)
			_, f := cache.get(next, 16, false)
			if f.code == "" {
				t.Fatal("oversized file accepted")
			}
		}
		if cache.stats.ReadCalls != 4 || cache.stats.UniqueBytes != 64 || cache.stats.PhysicalCharged != 68 {
			t.Fatalf("bounded failed attempts escaped: %+v", cache.stats)
		}
	})
	t.Run("witness_rejects_duplicate_actual_reads_and_wrong_phase", func(t *testing.T) {
		name := "internal/small.go"
		proofWrite(t, root, name, []byte("small"))
		p := proofFixturePolicy()
		events := []sourceLaneProofReadEvent{}
		p.afterRead = func(e sourceLaneProofReadEvent) { events = append(events, e) }
		cache := sourceLaneProofCache{ctx: context.Background(), root: opened, policy: p, files: map[string]*sourceLaneProofFile{}}
		cache.get(name, 16, false)
		cache.finalize()
		expected := map[string]int64{name: 5}
		if !proofReadWitnessMatches(events, expected) {
			t.Fatal("valid actual two-pass witness rejected")
		}
		cache.read(name, "initial", 16)
		if proofReadWitnessMatches(events, expected) {
			t.Fatal("witness accepted a third actual read")
		}
		wrong := append([]sourceLaneProofReadEvent(nil), events[:2]...)
		wrong[1].Phase = "cache_hit"
		if proofReadWitnessMatches(wrong, expected) {
			t.Fatal("witness accepted incorrect phase")
		}
	})
}

func TestSourceLaneProofCacheIdentity(t *testing.T) {
	base, original, _ := proofBatch(t, "six", 6)
	// Each record has an independently bound artifact to make unaffected
	// sibling evidence observable after replacing only the first artifact.
	for i := range original {
		original[i].Targets = append([]sourceLaneTargetRef(nil), original[i].Targets...)
		name := fmt.Sprintf("internal/connectors/defs/proof-fixture/target-%d.json", i)
		bytes, err := os.ReadFile(filepath.Join(base, original[i].Targets[0].Artifact))
		if err != nil {
			t.Fatal(err)
		}
		proofWrite(t, base, name, bytes)
		original[i].Targets[0].Artifact = name
	}
	for _, kind := range []string{"inode_replacement", "same_inode_content", "symlink", "directory"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			proofCopyFixtureFiles(t, base, root, original)
			records, catalog := proofSharedDocument(t, root, original)
			name := records[0].Targets[0].Artifact
			before, err := os.Stat(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				t.Fatal(err)
			}
			p := proofFixturePolicy()
			reached := false
			finalRead := false
			p.afterRead = func(e sourceLaneProofReadEvent) {
				if e.Path != name {
					return
				}
				if e.Phase == "final" {
					finalRead = true
				}
				if e.Phase != "initial" || reached {
					return
				}
				reached = true
				if !e.Success || e.Bytes != int64(len(expected)) || !os.SameFile(before, e.Before) {
					t.Fatal("mutation hook did not reach original completed read")
				}
				filename := filepath.Join(root, name)
				switch kind {
				case "inode_replacement":
					replacement := filename + ".replacement"
					if err := os.WriteFile(replacement, expected, 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.Rename(replacement, filename); err != nil {
						t.Fatal(err)
					}
					after, err := os.Stat(filename)
					if err != nil || os.SameFile(before, after) {
						t.Fatal("replacement did not produce independently different inode")
					}
				case "same_inode_content":
					changed := append([]byte(nil), expected...)
					changed[len(changed)-2] = ' '
					if err := os.WriteFile(filename, changed, 0600); err != nil {
						t.Fatal(err)
					}
					after, err := os.Stat(filename)
					if err != nil || !os.SameFile(before, after) {
						t.Fatal("mutation did not retain inode")
					}
				case "symlink":
					if err := os.Remove(filename); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(filepath.Join(base, name), filename); err != nil {
						t.Fatal(err)
					}
				case "directory":
					if err := os.Remove(filename); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(filename, 0700); err != nil {
						t.Fatal(err)
					}
				}
			}
			in := loadSourceLaneProofs(context.Background(), root, p, catalog)
			if !reached || !proofHasDiagnostic(in.Diagnostics, "proof_file_changed", "error") || in.accepted[records[0].ID] {
				t.Fatalf("changed cached file accepted: reached=%v diagnostics=%+v", reached, in.Diagnostics)
			}
			for _, r := range records[1:] {
				if !in.accepted[r.ID] {
					t.Fatalf("unaffected sibling %s lost: %+v", r.ID, in.Diagnostics)
				}
			}
			if kind == "same_inode_content" && !finalRead {
				t.Fatal("in-place mutation was not checked by actual final content read")
			}
		})
	}
	for _, order := range []string{"stale_then_correct", "correct_then_stale", "current_false"} {
		t.Run(order, func(t *testing.T) {
			root := t.TempDir()
			proofCopyFixtureFiles(t, base, root, original)
			records := append([]sourceLaneProofRecord(nil), original[:2]...)
			bad := 0
			if order == "correct_then_stale" {
				bad = 1
			}
			records[bad].Inputs = append([]sourceLaneProofInput(nil), records[bad].Inputs...)
			records[bad].Inputs[1].SHA256 = strings.Repeat("b", 64)
			records[bad].ClaimCurrent = order == "current_false"
			proofDocument(t, root, records)
			catalog := sourceLaneProofCatalog{}
			for _, r := range records {
				catalog.Reviews = append(catalog.Reviews, sourceLaneProofReview{Record: r, Fixture: true})
			}
			in := loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), catalog)
			code, severity := "proof_inputs_outdated", "deficit"
			if order == "current_false" {
				code, severity = "proof_current_claim_invalid", "error"
			}
			if !proofHasDiagnostic(in.Diagnostics, code, severity) || in.accepted[records[bad].ID] || !in.accepted[records[1-bad].ID] {
				t.Fatalf("claim-specific digest poisoned actual cache: %+v", in.Diagnostics)
			}
		})
	}
	t.Run("cached_bytes_do_not_replace_exact_target_identity", func(t *testing.T) {
		records, catalog := proofSharedDocument(t, base, original)
		in := loadSourceLaneProofs(context.Background(), base, proofFixturePolicy(), catalog)
		cells := proofBatchCells(records[0])
		cells[0].References = append([]sourceLaneTargetRef(nil), records[0].Targets...)
		cells[0].References[0].ID = records[1].Targets[0].ID
		got := assessSourceLaneProof(records[0].Key, cells, in)
		if got[0].State == "implemented" || !proofHasDiagnostic(got[0].Diagnostics, "proof_prerequisites_unproven", "deficit") {
			t.Fatalf("same-byte wrong target accepted: %+v", got[0])
		}
	})
	for _, phase := range []string{"before_read", "after_initial", "after_final", "after_last_final"} {
		t.Run("cancel_"+phase, func(t *testing.T) {
			records, catalog := proofSharedDocument(t, base, original)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := proofFixturePolicy()
			calls := 0
			if phase == "before_read" {
				cancel()
			}
			p.afterRead = func(e sourceLaneProofReadEvent) {
				calls++
				if phase == "after_initial" && e.Phase == "initial" || phase == "after_final" && e.Phase == "final" || phase == "after_last_final" && e.Phase == "final" && e.Path == records[len(records)-1].Targets[0].Artifact {
					cancel()
				}
			}
			in := loadSourceLaneProofs(ctx, base, p, catalog)
			if !proofHasDiagnostic(in.Diagnostics, "proof_cancelled", "error") || len(in.accepted) != 0 {
				t.Fatalf("cancellation hidden: %+v", in.Diagnostics)
			}
			if phase == "before_read" && calls != 0 {
				t.Fatal("read after initial cancellation")
			}
			for _, r := range records {
				if len(assessSourceLaneProof(r.Key, proofBatchCells(r), in)) != 7 {
					t.Fatal("cancel dropped cells")
				}
			}
		})
	}
}

func TestSourceLaneProofFiniteLimits(t *testing.T) {
	root, r, cells := proofFixture(t)
	t.Run("target_reader_above_four_MiB", func(t *testing.T) {
		copy := r
		copy.Targets = append([]sourceLaneTargetRef(nil), r.Targets...)
		copy.Targets[0].Artifact = "internal/connectors/defs/proof-fixture/large-target.json"
		raw := append([]byte(`{"operations":[{"id":"fixture.widgets"}]}`), []byte(strings.Repeat(" ", 4<<20))...)
		proofWrite(t, root, copy.Targets[0].Artifact, raw)
		copy.Targets[0].ArtifactSHA256 = sourceBytesHash(raw)
		proofDocument(t, root, []sourceLaneProofRecord{copy})
		in := proofLoadFixture(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
		local := append([]sourceLaneCell(nil), cells...)
		local[0].References = copy.Targets
		if assessSourceLaneProof(copy.Key, local, in)[0].State != "implemented" {
			t.Fatalf("valid target within 64 MiB refused: %+v", in.Diagnostics)
		}
	})
	t.Run("unique_file_limit", func(t *testing.T) {
		proofDocument(t, root, []sourceLaneProofRecord{r})
		p := proofFixturePolicy()
		p.limits.Files = 1
		in := loadSourceLaneProofs(context.Background(), root, p, sourceLaneProofCatalog{Reviews: []sourceLaneProofReview{{Record: r, Fixture: true}}})
		if !proofHasDiagnostic(in.Diagnostics, "proof_file_limit_exceeded", "error") {
			t.Fatalf("unique file ceiling ignored: %+v", in.Diagnostics)
		}
	})
	t.Run("document_lookahead_charged", func(t *testing.T) {
		proofDocument(t, root, []sourceLaneProofRecord{r})
		p := proofFixturePolicy()
		p.limits.DocumentBytes = 16
		in := loadSourceLaneProofs(context.Background(), root, p, sourceLaneProofCatalog{})
		if !proofHasDiagnostic(in.Diagnostics, "proof_document_invalid", "error") || in.Stats.ReadCalls != 1 || in.Stats.PhysicalCharged != 17 {
			t.Fatalf("document failed read uncharged: %+v %+v", in.Diagnostics, in.Stats)
		}
	})
	t.Run("record_target_bound", func(t *testing.T) {
		copy := r
		copy.Targets = make([]sourceLaneTargetRef, 65)
		for i := range copy.Targets {
			copy.Targets[i] = r.Targets[0]
			copy.Targets[i].ID = fmt.Sprintf("target-%d", i)
		}
		proofDocument(t, root, []sourceLaneProofRecord{copy})
		in := proofLoadFixture(root, nil)
		if !proofHasDiagnostic(in.Diagnostics, "proof_document_invalid", "error") {
			t.Fatalf("target allocation limit ignored: %+v", in.Diagnostics)
		}
	})
	for _, kind := range []string{"duplicate_run", "duplicate_pass", "failed_event", "wrong_package"} {
		t.Run(kind, func(t *testing.T) {
			copy := r
			copy.ReceiptPath = "data/connector-canon/proof-receipts/invalid-" + kind + ".jsonl"
			raw, err := os.ReadFile(filepath.Join(root, r.ReceiptPath))
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "duplicate_run", "duplicate_pass":
				action := "run"
				if kind == "duplicate_pass" {
					action = "pass"
				}
				lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
				for i, line := range lines {
					var event struct{ Action, Test string }
					if json.Unmarshal([]byte(line), &event) == nil && event.Action == action && event.Test == r.SelectedTest {
						lines = append(lines[:i+1], append([]string{line}, lines[i+1:]...)...)
						break
					}
				}
				raw = []byte(strings.Join(lines, "\n") + "\n")
			case "failed_event":
				raw = []byte(strings.ReplaceAll(string(raw), `"Action":"pass"`, `"Action":"fail"`))
			case "wrong_package":
				raw = []byte(strings.ReplaceAll(string(raw), r.Package, "wrong/package"))
			}
			copy.ReceiptSHA256 = sourceBytesHash(raw)
			proofWrite(t, root, copy.ReceiptPath, raw)
			proofDocument(t, root, []sourceLaneProofRecord{copy})
			in := proofLoadFixture(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
			if !proofHasDiagnostic(in.Diagnostics, "proof_result_invalid", "error") || len(in.accepted) != 0 {
				t.Fatalf("false result accepted %s: %+v", kind, in.Diagnostics)
			}
		})
	}
	t.Run("null_inline_plus_shared_is_ambiguous", func(t *testing.T) {
		records, catalog := proofSharedDocument(t, root, []sourceLaneProofRecord{r})
		raw, err := os.ReadFile(filepath.Join(root, sourceLaneProofPath))
		if err != nil {
			t.Fatal(err)
		}
		raw = []byte(strings.Replace(string(raw), `"input_set_sha256":`, `"inputs":null,"input_set_sha256":`, 1))
		proofWrite(t, root, sourceLaneProofPath, raw)
		in := loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), catalog)
		if len(in.accepted) != 0 || len(in.Diagnostics) == 0 {
			t.Fatalf("ambiguous input representations accepted %s: %+v", records[0].ID, in)
		}
	})
}

func proofCopyFixtureFiles(t *testing.T, source, destination string, records []sourceLaneProofRecord) {
	t.Helper()
	names := map[string]bool{}
	for _, r := range records {
		for _, pin := range r.Inputs {
			names[pin.Path] = true
		}
		for _, ref := range r.Targets {
			names[ref.Artifact] = true
		}
		names[r.ReceiptPath] = true
	}
	for name := range names {
		raw, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		proofWrite(t, destination, name, raw)
	}
}

func proofWrite(t *testing.T, root, name string, raw []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, name), raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func proofDocument(t *testing.T, root string, records []sourceLaneProofRecord) {
	t.Helper()
	raw, err := json.Marshal(struct {
		SchemaVersion int                     `json:"schema_version"`
		Records       []sourceLaneProofRecord `json:"records"`
	}{1, records})
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, sourceLaneProofPath, raw)
}

func proofFixture(t *testing.T) (string, sourceLaneProofRecord, []sourceLaneCell) {
	t.Helper()
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "tool", "test2json", "-p", "polymetrics.ai/cmd/connectorgen", executable, "-test.v", "-test.run=^TestSourceLaneProofFixtureBehavior$/^bounded_records$", "-test.timeout=20m")
	receipt, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("actual fixture execution: %v\n%s", err, receipt)
	}
	t.Logf("Original actual hermetic fixture receipt (SHA256 %s):\n%s", sourceBytesHash(receipt), receipt)
	if !strings.Contains(string(receipt), `"Action":"pass","Package":"polymetrics.ai/cmd/connectorgen","Test":"TestSourceLaneProofFixtureBehavior/bounded_records"`) {
		t.Fatalf("selected fixture did not pass: %s", receipt)
	}
	r := sourceLaneProofRecord{ID: "fixture-widgets-read", Key: sourceOperationKey{Connector: "proof-fixture", Inventory: "fixture", ID: "fixture.widgets"}, Lane: "direct_read", TestPath: "cmd/connectorgen/source_lane_proof_test.go", TestSymbol: "TestSourceLaneProofFixtureBehavior", SelectedTest: "TestSourceLaneProofFixtureBehavior/bounded_records", Package: "polymetrics.ai/cmd/connectorgen", ExecutionClass: "C2", Scope: "hermetic_fixture", ReceiptPath: "data/connector-canon/proof-receipts/fixture.jsonl", ReceiptSHA256: sourceBytesHash(receipt), ObservableContract: "GET /widgets sends limit=2; exact widget-A/widget-B records; one request; incomplete page", Limitations: []string{"Fixture-only operation; no provider-live or C1 certification."}}
	proofWrite(t, root, r.ReceiptPath, receipt)
	for _, entry := range []struct{ path, role string }{{"cmd/connectorgen/source_lane_proof_test.go", "test"}, {"internal/connectors/engine/direct_read.go", "code"}, {"go.mod", "dependency"}, {"go.sum", "dependency"}} {
		raw, err := os.ReadFile(filepath.Join("../..", entry.path))
		if err != nil {
			t.Fatal(err)
		}
		proofWrite(t, root, entry.path, raw)
		r.Inputs = append(r.Inputs, sourceLaneProofInput{Path: entry.path, Role: entry.role, SHA256: sourceBytesHash(raw)})
	}
	// Retain the local code dependency closure as well as module pins. The
	// reviewed fixture does not infer completeness from a single engine file.
	for _, directory := range []string{"internal", "cmd/connectorgen"} {
		err := filepath.WalkDir(filepath.Join("../..", directory), func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(filename, ".go") {
				return nil
			}
			rel, err := filepath.Rel("../..", filename)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			for _, in := range r.Inputs {
				if in.Path == rel {
					return nil
				}
			}
			raw, err := os.ReadFile(filename)
			if err != nil {
				return err
			}
			proofWrite(t, root, rel, raw)
			role := "code"
			if strings.HasSuffix(rel, "_test.go") {
				role = "test"
			}
			r.Inputs = append(r.Inputs, sourceLaneProofInput{Path: rel, Role: role, SHA256: sourceBytesHash(raw)})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	raw := []byte(`{"operations":[{"id":"fixture.widgets"}]}`)
	artifact := "internal/connectors/defs/proof-fixture/operations.json"
	proofWrite(t, root, artifact, raw)
	r.Targets = []sourceLaneTargetRef{{Kind: "operation", Connector: r.Key.Connector, ID: r.Key.ID, Lane: r.Lane, Artifact: artifact, Pointer: "/operations/0", ArtifactSHA256: sourceBytesHash(raw), CanonicalID: "operation:fixture.widgets", CanonicalPointer: "/operations/0/operation", Generation: strings.Repeat("a", 64)}}
	cells := make([]sourceLaneCell, 0, 7)
	for _, lane := range sourceLaneNames() {
		cells = append(cells, sourceLaneCell{Lane: lane, Applicability: "not_applicable", State: "not_applicable", RuleID: "fixture_source_exclusion", Reason: sourceLaneReason{Code: "source_exclusion"}, OwnerRefs: []string{r.Key.Connector}})
	}
	cells[0].Applicability = "applicable"
	cells[0].State = "mapped_unproven"
	cells[0].References = append([]sourceLaneTargetRef(nil), r.Targets...)
	return root, r, cells
}

func TestSourceLaneProofAdditionalControls(t *testing.T) {
	root, r, cells := proofFixture(t)
	t.Run("unreviewed_is_not_evidence", func(t *testing.T) {
		proofDocument(t, root, []sourceLaneProofRecord{r})
		in := proofLoadFixture(root, nil)
		if !proofHasDiagnostic(in.Diagnostics, "proof_review_unavailable", "deficit") || assessSourceLaneProof(r.Key, cells, in)[0].State == "implemented" {
			t.Fatalf("unreviewed proof accepted: %+v", in)
		}
	})
	t.Run("fixture_cannot_promote_primary", func(t *testing.T) {
		copy := r
		copy.Key.Inventory = "primary"
		proofDocument(t, root, []sourceLaneProofRecord{copy})
		in := proofLoadFixture(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
		if !proofHasDiagnostic(in.Diagnostics, "proof_scope_unproven", "deficit") {
			t.Fatalf("fixture became primary proof: %+v", in)
		}
	})
	t.Run("contradictory_same_cell", func(t *testing.T) {
		other := r
		other.ID = "other-proof"
		other.ObservableContract = "Contradictory claim"
		proofDocument(t, root, []sourceLaneProofRecord{r, other})
		in := proofLoadFixture(root, []sourceLaneProofReview{{Record: r, Fixture: true}, {Record: other, Fixture: true}})
		if len(in.Diagnostics) != 2 || assessSourceLaneProof(r.Key, cells, in)[0].State == "implemented" {
			t.Fatalf("contradiction accepted: %+v", in)
		}
	})
	t.Run("missing_one_required_target", func(t *testing.T) {
		copy := r
		copy.Targets = append([]sourceLaneTargetRef(nil), r.Targets...)
		additional := r.Targets[0]
		additional.Kind = "command"
		additional.ID = "widgets list"
		copy.Targets = append(copy.Targets, additional)
		proofDocument(t, root, []sourceLaneProofRecord{copy})
		in := proofLoadFixture(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
		got := assessSourceLaneProof(r.Key, cells, in)
		if !proofHasDiagnostic(got[0].Diagnostics, "proof_prerequisites_unproven", "deficit") || got[0].State == "implemented" {
			t.Fatalf("missing required reference accepted: %+v", got)
		}
	})
	t.Run("budget_exhausted", func(t *testing.T) {
		opened, err := os.OpenRoot(root)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := opened.Close(); err != nil {
				t.Errorf("close fixture root: %v", err)
			}
		}()
		budget := int64(0)
		code, severity := proofReadWithBudget(opened, r, &budget)
		if code != "proof_read_budget_exceeded" || severity != "error" {
			t.Fatalf("budget ignored: %s/%s", code, severity)
		}
	})
	t.Run("unchanged_input_cells", func(t *testing.T) {
		proofDocument(t, root, []sourceLaneProofRecord{r})
		in := proofLoadFixture(root, []sourceLaneProofReview{{Record: r, Fixture: true}})
		before, _ := json.Marshal(cells)
		_ = assessSourceLaneProof(r.Key, cells, in)
		after, _ := json.Marshal(cells)
		if string(before) != string(after) {
			t.Fatal("reducer mutated input cells")
		}
	})
	t.Run("control_character_identity", func(t *testing.T) {
		copy := r
		copy.Key.ID = "bad\nidentity"
		proofDocument(t, root, []sourceLaneProofRecord{copy})
		in := proofLoadFixture(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
		if !proofHasDiagnostic(in.Diagnostics, "proof_record_invalid", "error") {
			t.Fatalf("ambiguous ID accepted: %+v", in)
		}
	})
	t.Run("absent_document_beneath_symlink", func(t *testing.T) {
		isolated := t.TempDir()
		if err := os.Symlink(t.TempDir(), filepath.Join(isolated, "data")); err != nil {
			t.Fatal(err)
		}
		in := proofLoadFixture(isolated, nil)
		if !proofHasDiagnostic(in.Diagnostics, "proof_document_invalid", "error") {
			t.Fatalf("unsafe optional path treated as absence: %+v", in)
		}
	})
}

func proofHasDiagnostic(ds []sourceLaneDiagnostic, code, severity string) bool {
	for _, d := range ds {
		if d.Code == code && d.Severity == severity {
			return true
		}
	}
	return false
}

func TestSourceLaneStateReduction(t *testing.T) {
	root, record, cells := proofFixture(t)
	t.Run("absent_evidence_retains_seven", func(t *testing.T) {
		got := assessSourceLaneProof(record.Key, cells, proofLoadFixture(root, nil))
		if len(got) != 7 || got[0].State != "mapped_unproven" || !proofHasDiagnostic(got[0].Diagnostics, "proof_unavailable", "deficit") {
			t.Fatalf("absent proof: %+v", got)
		}
		if !reflect.DeepEqual(got[1].Reason, cells[1].Reason) {
			t.Fatal("source exclusion overwritten")
		}
	})
	t.Run("matching_actual_fixture_promotes", func(t *testing.T) {
		proofDocument(t, root, []sourceLaneProofRecord{record})
		got := assessSourceLaneProof(record.Key, cells, proofLoadFixture(root, []sourceLaneProofReview{{Record: record, Fixture: true}}))
		if len(got) != 7 || got[0].State != "implemented" || !reflect.DeepEqual(got[0].ProofRefs, []string{record.ID}) {
			t.Fatalf("matching actual fixture must promote only direct_read: %+v", got)
		}
		for i := 1; i < 7; i++ {
			if got[i].State != "not_applicable" {
				t.Fatalf("cross-lane promotion: %+v", got[i])
			}
		}
	})
}

func TestSourceLaneProofInvalidClaims(t *testing.T) {
	root, original, cells := proofFixture(t)
	for _, tc := range []struct {
		name   string
		change func(*sourceLaneProofRecord)
		code   string
	}{
		{"wrong_source", func(r *sourceLaneProofRecord) { r.Key.ID = "same-count-other" }, "proof_review_mismatch"},
		{"cross_lane", func(r *sourceLaneProofRecord) { r.Lane = "etl" }, "proof_review_mismatch"},
		{"target_digest", func(r *sourceLaneProofRecord) { r.Targets[0].ArtifactSHA256 = strings.Repeat("b", 64) }, "proof_review_mismatch"},
		{"input_digest", func(r *sourceLaneProofRecord) { r.Inputs[1].SHA256 = strings.Repeat("b", 64) }, "proof_review_mismatch"},
		{"test_digest", func(r *sourceLaneProofRecord) { r.Inputs[0].SHA256 = strings.Repeat("b", 64) }, "proof_review_mismatch"},
		{"dependency_digest", func(r *sourceLaneProofRecord) { r.Inputs[2].SHA256 = strings.Repeat("b", 64) }, "proof_review_mismatch"},
		{"wrong_scope", func(r *sourceLaneProofRecord) { r.Scope = "provider_live" }, "proof_review_mismatch"},
		{"receipt_hash", func(r *sourceLaneProofRecord) { r.ReceiptSHA256 = strings.Repeat("b", 64) }, "proof_review_mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := original
			r.Inputs = append([]sourceLaneProofInput(nil), original.Inputs...)
			r.Targets = append([]sourceLaneTargetRef(nil), original.Targets...)
			tc.change(&r)
			proofDocument(t, root, []sourceLaneProofRecord{r})
			inputs := proofLoadFixture(root, []sourceLaneProofReview{{Record: original, Fixture: true}})
			if !proofHasDiagnostic(inputs.Diagnostics, tc.code, "error") {
				t.Fatalf("invalid asserted claim lost: %+v", inputs.Diagnostics)
			}
			got := assessSourceLaneProof(original.Key, cells, inputs)
			if len(got) != 7 || got[0].State == "implemented" {
				t.Fatalf("proof failure altered membership/promoted: %+v", got)
			}
		})
	}
}

func TestSourceLaneProofBoundaries(t *testing.T) {
	base, original, cells := proofFixture(t)
	for _, tc := range []struct {
		name, code, severity string
		edit                 func(string, *sourceLaneProofRecord)
	}{
		{"stale_input", "proof_inputs_outdated", "deficit", func(root string, r *sourceLaneProofRecord) {
			proofWrite(t, root, r.Inputs[1].Path, []byte("changed code"))
		}},
		{"falsely_current_input", "proof_current_claim_invalid", "error", func(root string, r *sourceLaneProofRecord) {
			r.ClaimCurrent = true
			proofWrite(t, root, r.Inputs[1].Path, []byte("changed code"))
		}},
		{"missing_receipt", "proof_receipt_unavailable", "deficit", func(root string, r *sourceLaneProofRecord) {
			if err := os.Remove(filepath.Join(root, r.ReceiptPath)); err != nil {
				t.Fatal(err)
			}
		}},
		{"unsafe_path", "proof_record_invalid", "error", func(root string, r *sourceLaneProofRecord) { r.Inputs[1].Path = "../outside.go" }},
		{"absolute_path", "proof_record_invalid", "error", func(root string, r *sourceLaneProofRecord) { r.ReceiptPath = "/tmp/secret" }},
		{"symlink_escape", "proof_input_invalid", "error", func(root string, r *sourceLaneProofRecord) {
			name := filepath.Join(root, r.Inputs[1].Path)
			if err := os.Remove(name); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(base, r.Inputs[1].Path), name); err != nil {
				t.Fatal(err)
			}
		}},
		{"preflight_only", "proof_scope_unproven", "deficit", func(_ string, r *sourceLaneProofRecord) { r.Scope = "preflight_only" }},
		{"syntax_only", "proof_scope_unproven", "deficit", func(_ string, r *sourceLaneProofRecord) { r.Scope = "syntax_only" }},
		{"shared_engine", "proof_scope_unproven", "deficit", func(_ string, r *sourceLaneProofRecord) { r.Scope = "shared_engine" }},
		{"C3", "proof_scope_unproven", "deficit", func(_ string, r *sourceLaneProofRecord) { r.ExecutionClass = "C3" }},
		{"C4", "proof_scope_unproven", "deficit", func(_ string, r *sourceLaneProofRecord) { r.ExecutionClass = "C4" }},
		{"fixture_claims_C1", "proof_scope_unproven", "deficit", func(_ string, r *sourceLaneProofRecord) { r.ExecutionClass = "C1" }},
		{"zero_selected", "proof_result_invalid", "error", func(root string, r *sourceLaneProofRecord) {
			raw, err := os.ReadFile(filepath.Join(root, r.ReceiptPath))
			if err != nil {
				t.Fatal(err)
			}
			raw = []byte(strings.ReplaceAll(string(raw), r.SelectedTest, "Other/subtest"))
			r.ReceiptSHA256 = sourceBytesHash(raw)
			proofWrite(t, root, r.ReceiptPath, raw)
		}},
		{"skipped_selected", "proof_result_invalid", "error", func(root string, r *sourceLaneProofRecord) {
			raw, err := os.ReadFile(filepath.Join(root, r.ReceiptPath))
			if err != nil {
				t.Fatal(err)
			}
			raw = []byte(strings.ReplaceAll(string(raw), `"Action":"pass"`, `"Action":"skip"`))
			r.ReceiptSHA256 = sourceBytesHash(raw)
			proofWrite(t, root, r.ReceiptPath, raw)
		}},
		{"forged_receipt_bytes", "proof_receipt_digest_invalid", "error", func(root string, r *sourceLaneProofRecord) { proofWrite(t, root, r.ReceiptPath, []byte("PASS")) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			r := original
			r.Inputs = append([]sourceLaneProofInput(nil), r.Inputs...)
			for _, p := range append(append([]sourceLaneProofInput(nil), r.Inputs...), sourceLaneProofInput{Path: r.ReceiptPath}, sourceLaneProofInput{Path: r.Targets[0].Artifact}) {
				raw, err := os.ReadFile(filepath.Join(base, p.Path))
				if err != nil {
					t.Fatal(err)
				}
				proofWrite(t, root, p.Path, raw)
			}
			tc.edit(root, &r)
			proofDocument(t, root, []sourceLaneProofRecord{r})
			inputs := proofLoadFixture(root, []sourceLaneProofReview{{Record: r, Fixture: true}})
			if !proofHasDiagnostic(inputs.Diagnostics, tc.code, tc.severity) {
				t.Fatalf("want %s/%s, got %+v", tc.code, tc.severity, inputs.Diagnostics)
			}
			got := assessSourceLaneProof(r.Key, cells, inputs)
			if len(got) != 7 || got[0].State == "implemented" {
				t.Fatalf("invalid proof promoted/lost cells: %+v", got)
			}
			if tc.severity == "error" && !proofHasDiagnostic(got[0].Diagnostics, tc.code, "error") {
				t.Fatalf("invalidity suppressed by state: %+v", got[0])
			}
		})
	}
	t.Run("duplicate_claims", func(t *testing.T) {
		proofDocument(t, base, []sourceLaneProofRecord{original, original})
		in := proofLoadFixture(base, []sourceLaneProofReview{{Record: original, Fixture: true}})
		if !proofHasDiagnostic(in.Diagnostics, "proof_duplicate_claim", "error") {
			t.Fatalf("duplicate accepted: %+v", in)
		}
		if assessSourceLaneProof(original.Key, cells, in)[0].State == "implemented" {
			t.Fatal("duplicate promoted")
		}
	})
	t.Run("malformed_closed_document", func(t *testing.T) {
		for _, raw := range []string{`{"schema_version":1,"records":[],"extra":1}`, `{"schema_version":1,"records":[],"records":[]}`, `{"schema_version":1,"records":null}`, `{"schema_version":1,"records":[]} {}`} {
			proofWrite(t, base, sourceLaneProofPath, []byte(raw))
			in := proofLoadFixture(base, nil)
			if !proofHasDiagnostic(in.Diagnostics, "proof_document_invalid", "error") {
				t.Fatalf("malformed proof accepted %s: %+v", raw, in)
			}
		}
	})
	t.Run("missing_applicability_and_references", func(t *testing.T) {
		proofDocument(t, base, []sourceLaneProofRecord{original})
		in := proofLoadFixture(base, []sourceLaneProofReview{{Record: original, Fixture: true}})
		for _, kind := range []string{"undetermined", "reference", "missing_foundation", "exclusion"} {
			t.Run(kind, func(t *testing.T) {
				local := append([]sourceLaneCell(nil), cells...)
				switch kind {
				case "undetermined":
					local[0].Applicability = "undetermined"
				case "reference":
					local[0].References = nil
				case "missing_foundation":
					local[0].State = "missing_foundation"
					local[0].GapRefs = []string{"existing-gap"}
				case "exclusion":
					local[0].Applicability = "not_applicable"
					local[0].State = "not_applicable"
				}
				got := assessSourceLaneProof(original.Key, local, in)
				if got[0].State == "implemented" || len(got) != 7 {
					t.Fatalf("proof overrode source/reference: %+v", got)
				}
				if kind == "missing_foundation" && (got[0].State != "missing_foundation" || !reflect.DeepEqual(got[0].GapRefs, local[0].GapRefs)) {
					t.Fatal("existing gap changed")
				}
			})
		}
	})
}

func proofFixturePolicy() sourceLaneProofPolicy {
	keys := []sourceOperationKey{{Connector: "proof-fixture", Inventory: "fixture", ID: "fixture.widgets"}, {Connector: "proof-fixture", Inventory: "primary", ID: "fixture.widgets"}}
	for i := 0; i < 4343; i++ {
		keys = append(keys, sourceOperationKey{Connector: "proof-fixture", Inventory: "fixture", ID: fmt.Sprintf("fixture-%06d", i)})
	}
	p, err := newSourceLaneProofPolicy(keys)
	if err != nil {
		panic(err)
	}
	return p
}
func proofLoadFixture(root string, reviews []sourceLaneProofReview) sourceLaneProofInputs {
	return loadSourceLaneProofs(context.Background(), root, proofFixturePolicy(), sourceLaneProofCatalog{Reviews: reviews})
}
func proofReadWithBudget(root *os.Root, r sourceLaneProofRecord, budget *int64) (string, string) {
	p := proofFixturePolicy()
	p.limits.UniqueBytes = *budget
	c := sourceLaneProofCache{ctx: context.Background(), root: root, policy: p, files: map[string]*sourceLaneProofFile{}}
	code, severity, _ := assessSourceLaneProofFiles(&c, r, r.Inputs)
	*budget -= c.stats.UniqueBytes
	return code, severity
}

func TestSourceLaneProofSchemaRootShape(t *testing.T) {
	_, record, _ := proofFixture(t)
	for _, tc := range []struct {
		name   string
		mutate func(*sourceLaneTargetRef)
		valid  bool
	}{
		{"existing operation", nil, true},
		{"schema root", func(ref *sourceLaneTargetRef) {
			ref.Kind = "schema"
			ref.ID = "schemas/widgets.json"
			ref.Artifact = "internal/connectors/defs/proof-fixture/schemas/widgets.json"
			ref.Pointer = ""
			ref.CanonicalPointer = "/operations/0/schema_refs/record"
			ref.SchemaRole = "record"
		}, true},
		{"operation empty pointer", func(ref *sourceLaneTargetRef) { ref.Pointer = "" }, false},
		{"unknown kind", func(ref *sourceLaneTargetRef) { ref.Kind = "unknown" }, false},
		{"unknown schema role", func(ref *sourceLaneTargetRef) { ref.SchemaRole = "unknown" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copy := record
			copy.Targets = append([]sourceLaneTargetRef(nil), record.Targets...)
			if tc.mutate != nil {
				tc.mutate(&copy.Targets[0])
			}
			if got := sourceLaneProofShape(copy); got != tc.valid {
				t.Fatalf("actual target shape acceptance=%v; want%v for %+v", got, tc.valid, copy.Targets[0])
			}
		})
	}
}

func TestSourceLaneProofProjectionIdentity(t *testing.T) {
	_, record, cells := proofFixture(t)
	ref := record.Targets[0]
	ref.SchemaRole = sourceLaneSchemaRequest
	ref.SourceSchema = &sourceFactRef{DocumentID: "fixture:retained", Pointer: "/request/schema", ValueSHA256: strings.Repeat("b", 64)}
	pointer := "/parameters/0"
	ref.FieldMappings = []sourceLaneFieldMapping{{Source: sourceFactRef{DocumentID: "fixture:retained", Pointer: "/parameters/0", ValueSHA256: strings.Repeat("c", 64)}, Target: sourceLaneFieldTarget{Kind: sourceLaneFieldParameter, Pointer: &pointer}}}
	record.Targets = []sourceLaneTargetRef{canonicalSourceLaneTargetRef(ref)}
	// This is the required-target reducer boundary: References and reviewed
	// byCell input are supplied explicitly. It does not certify source-binding
	// admission or claim the new projection was exercised by the HTTP fixture.
	inputs := sourceLaneProofInputs{byCell: map[sourceLaneProofCell]sourceLaneProofRecord{{record.Key, record.Lane}: record}}
	for _, tc := range []struct {
		name    string
		change  func(*sourceLaneTargetRef)
		blocked bool
	}{
		{"separately allocated equal", func(*sourceLaneTargetRef) {}, false},
		{"changed source schema", func(r *sourceLaneTargetRef) { r.SourceSchema.ValueSHA256 = strings.Repeat("d", 64) }, true},
		{"changed mapping source", func(r *sourceLaneTargetRef) { r.FieldMappings[0].Source.Pointer = "/parameters/1" }, true},
		{"changed mapping target", func(r *sourceLaneTargetRef) { *r.FieldMappings[0].Target.Pointer = "/parameters/1" }, true},
		{"duplicate invalid mapping", func(r *sourceLaneTargetRef) { r.FieldMappings = append(r.FieldMappings, r.FieldMappings[0]) }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			observed := canonicalSourceLaneTargetRef(ref)
			tc.change(&observed)
			actual := append([]sourceLaneCell(nil), cells...)
			actual[0].References = []sourceLaneTargetRef{observed}
			result := assessSourceLaneProof(record.Key, actual, inputs)
			if got := result[0].State == "implemented"; got == tc.blocked {
				t.Fatalf("projection match promoted=%v blocked=%v diagnostics=%+v", got, tc.blocked, result[0].Diagnostics)
			}
		})
	}
	actual := append([]sourceLaneCell(nil), cells...)
	actual[0].References = []sourceLaneTargetRef{canonicalSourceLaneTargetRef(ref)}
	actual[0].Diagnostics = []sourceLaneDiagnostic{{Key: record.Key, Lanes: []string{record.Lane}, Stage: "reference", Code: "source_projection_unverified", Severity: "deficit"}}
	if result := assessSourceLaneProof(record.Key, actual, inputs); result[0].State == "implemented" {
		t.Fatal("matching projection proof overrode reference deficit")
	}
}
