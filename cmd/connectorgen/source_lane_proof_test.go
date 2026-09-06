package main

import (
	"context"
	"encoding/json"
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
		in := loadSourceLaneProofs(root, nil)
		if !proofHasDiagnostic(in.Diagnostics, "proof_review_unavailable", "deficit") || assessSourceLaneProof(r.Key, cells, in)[0].State == "implemented" {
			t.Fatalf("unreviewed proof accepted: %+v", in)
		}
	})
	t.Run("fixture_cannot_promote_primary", func(t *testing.T) {
		copy := r
		copy.Key.Inventory = "primary"
		proofDocument(t, root, []sourceLaneProofRecord{copy})
		in := loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
		if !proofHasDiagnostic(in.Diagnostics, "proof_scope_unproven", "deficit") {
			t.Fatalf("fixture became primary proof: %+v", in)
		}
	})
	t.Run("contradictory_same_cell", func(t *testing.T) {
		other := r
		other.ID = "other-proof"
		other.ObservableContract = "Contradictory claim"
		proofDocument(t, root, []sourceLaneProofRecord{r, other})
		in := loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: r, Fixture: true}, {Record: other, Fixture: true}})
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
		in := loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
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
		defer opened.Close()
		budget := int64(0)
		code, severity := sourceLaneReadProof(opened, r, &budget)
		if code != "proof_read_budget_exceeded" || severity != "error" {
			t.Fatalf("budget ignored: %s/%s", code, severity)
		}
	})
	t.Run("unchanged_input_cells", func(t *testing.T) {
		proofDocument(t, root, []sourceLaneProofRecord{r})
		in := loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: r, Fixture: true}})
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
		in := loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: copy, Fixture: true}})
		if !proofHasDiagnostic(in.Diagnostics, "proof_record_invalid", "error") {
			t.Fatalf("ambiguous ID accepted: %+v", in)
		}
	})
	t.Run("absent_document_beneath_symlink", func(t *testing.T) {
		isolated := t.TempDir()
		if err := os.Symlink(t.TempDir(), filepath.Join(isolated, "data")); err != nil {
			t.Fatal(err)
		}
		in := loadSourceLaneProofs(isolated, nil)
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
		got := assessSourceLaneProof(record.Key, cells, loadSourceLaneProofs(root, nil))
		if len(got) != 7 || got[0].State != "mapped_unproven" || !proofHasDiagnostic(got[0].Diagnostics, "proof_unavailable", "deficit") {
			t.Fatalf("absent proof: %+v", got)
		}
		if !reflect.DeepEqual(got[1].Reason, cells[1].Reason) {
			t.Fatal("source exclusion overwritten")
		}
	})
	t.Run("matching_actual_fixture_promotes", func(t *testing.T) {
		proofDocument(t, root, []sourceLaneProofRecord{record})
		got := assessSourceLaneProof(record.Key, cells, loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: record, Fixture: true}}))
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
			inputs := loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: original, Fixture: true}})
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
			inputs := loadSourceLaneProofs(root, []sourceLaneProofReview{{Record: r, Fixture: true}})
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
		in := loadSourceLaneProofs(base, []sourceLaneProofReview{{Record: original, Fixture: true}})
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
			in := loadSourceLaneProofs(base, nil)
			if !proofHasDiagnostic(in.Diagnostics, "proof_document_invalid", "error") {
				t.Fatalf("malformed proof accepted %s: %+v", raw, in)
			}
		}
	})
	t.Run("missing_applicability_and_references", func(t *testing.T) {
		proofDocument(t, base, []sourceLaneProofRecord{original})
		in := loadSourceLaneProofs(base, []sourceLaneProofReview{{Record: original, Fixture: true}})
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
