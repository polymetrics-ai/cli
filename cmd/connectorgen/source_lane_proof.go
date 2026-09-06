package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path"
	"reflect"
	"strings"
)

const sourceLaneProofPath = "data/connector-canon/batch1-source-lane-proofs.json"

type sourceLaneProofInput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Role   string `json:"role"`
}

type sourceLaneProofRecord struct {
	ID                 string                 `json:"id"`
	Key                sourceOperationKey     `json:"key"`
	Lane               string                 `json:"lane"`
	Targets            []sourceLaneTargetRef  `json:"targets"`
	Inputs             []sourceLaneProofInput `json:"inputs"`
	TestPath           string                 `json:"test_path"`
	TestSymbol         string                 `json:"test_symbol"`
	SelectedTest       string                 `json:"selected_test"`
	Package            string                 `json:"package"`
	ExecutionClass     string                 `json:"execution_class"`
	Scope              string                 `json:"scope"`
	ReceiptPath        string                 `json:"receipt_path"`
	ReceiptSHA256      string                 `json:"receipt_sha256"`
	ObservableContract string                 `json:"observable_contract"`
	Limitations        []string               `json:"limitations"`
	ClaimCurrent       bool                   `json:"claim_current"`
}

// Reviews are supplied by the authoring caller, never loaded from proof JSON.
// The owner reviews assertion meaning, complete input coverage and required
// lane targets. Hashes alone cannot establish those semantic facts.
type sourceLaneProofReview struct {
	Record  sourceLaneProofRecord
	Fixture bool
}

type sourceLaneProofInputs struct {
	Records     []sourceLaneProofRecord
	Diagnostics []sourceLaneDiagnostic
	reviews     []sourceLaneProofReview
	accepted    map[string]bool
}

func loadSourceLaneProofs(repo string, reviews []sourceLaneProofReview) sourceLaneProofInputs {
	result := sourceLaneProofInputs{reviews: reviews, accepted: map[string]bool{}}
	add := func(r sourceLaneProofRecord, code, severity string) {
		result.Diagnostics = append(result.Diagnostics, sourceLaneDiagnostic{Key: r.Key, Lanes: []string{r.Lane}, Stage: "proof", Code: code, Pointer: sourceLaneProofPath, Owner: r.Key.Connector, Severity: severity})
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		add(sourceLaneProofRecord{}, "proof_root_unavailable", "error")
		return result
	}
	defer root.Close()
	// Only an absent optional document is absence. Subsequent read/close errors
	// are never reclassified through a potentially joined ErrNotExist cause.
	if _, err := root.Lstat(sourceLaneProofPath); errors.Is(err, fs.ErrNotExist) {
		return result
	}
	raw, err := readSourceInput(root, sourceLaneProofPath, 4<<20)
	if err != nil {
		add(sourceLaneProofRecord{}, "proof_document_invalid", "error")
		return result
	}
	var document struct {
		SchemaVersion int                     `json:"schema_version"`
		Records       []sourceLaneProofRecord `json:"records"`
	}
	if decodeSourceJSON(raw, &document) != nil || decodeStrictJSON(raw, &document) != nil || document.SchemaVersion != 1 || document.Records == nil || len(document.Records) > 4096 {
		add(sourceLaneProofRecord{}, "proof_document_invalid", "error")
		return result
	}
	result.Records = document.Records
	ids := map[string]int{}
	type cellKey struct {
		key  sourceOperationKey
		lane string
	}
	cells := map[cellKey]int{}
	for _, r := range result.Records {
		ids[r.ID]++
		cells[cellKey{r.Key, r.Lane}]++
	}
	budget := int64(64<<20) - int64(len(raw))
	for _, r := range result.Records {
		if ids[r.ID] != 1 || cells[cellKey{r.Key, r.Lane}] != 1 {
			add(r, "proof_duplicate_claim", "error")
			continue
		}
		var review *sourceLaneProofReview
		count := 0
		for i := range reviews {
			if reviews[i].Record.ID == r.ID {
				review = &reviews[i]
				count++
			}
		}
		if count > 1 || count == 1 && !reflect.DeepEqual(review.Record, r) {
			add(r, "proof_review_mismatch", "error")
			continue
		}
		if !sourceLaneProofShape(r) {
			add(r, "proof_record_invalid", "error")
			continue
		}
		if count == 0 {
			severity := "deficit"
			if r.ClaimCurrent {
				severity = "error"
			}
			add(r, "proof_review_unavailable", severity)
			continue
		}
		productionScope := (r.Scope == "provider_live" && r.ExecutionClass == "C1") || (r.Scope == "connector_hermetic" && r.ExecutionClass == "C2")
		if (review.Fixture && (r.Scope != "hermetic_fixture" || r.ExecutionClass != "C2" || r.Key.Inventory != "fixture")) || (!review.Fixture && (!productionScope || r.Key.Inventory == "fixture")) {
			add(r, "proof_scope_unproven", "deficit")
			continue
		}
		code, severity := sourceLaneReadProof(root, r, &budget)
		if code != "" {
			add(r, code, severity)
			continue
		}
		result.accepted[r.ID] = true
	}
	return result
}

func sourceLaneProofSafePath(p string) bool {
	return validSourceID(p) && p != "." && path.Clean(p) == p && !path.IsAbs(p) && !strings.HasPrefix(p, "../") && !strings.Contains(p, "\\")
}

func sourceLaneProofDigest(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == 32 && strings.ToLower(s) == s
}

func sourceLaneProofShape(r sourceLaneProofRecord) bool {
	if !validSourceID(r.ID) || !validSourceID(r.Key.Connector) || !validSourceID(r.Key.Inventory) || !validSourceID(r.Key.ID) || !validSourceID(r.ObservableContract) || len(r.Limitations) == 0 || len(r.Targets) == 0 || len(r.Targets) > 64 || len(r.Inputs) > 4096 {
		return false
	}
	lane := false
	for _, name := range sourceLaneNames() {
		lane = lane || name == r.Lane
	}
	if !lane {
		return false
	}
	if !sourceLaneProofSafePath(r.TestPath) || !strings.HasPrefix(r.TestPath, "cmd/") && !strings.HasPrefix(r.TestPath, "internal/") || !strings.HasSuffix(r.TestPath, "_test.go") || !validSourceID(r.Package) || !strings.HasPrefix(r.TestSymbol, "Test") || strings.Contains(r.TestSymbol, "/") || !strings.HasPrefix(r.SelectedTest, r.TestSymbol+"/") || !validSourceID(r.SelectedTest) {
		return false
	}
	if !sourceLaneProofSafePath(r.ReceiptPath) || !strings.HasPrefix(r.ReceiptPath, "data/connector-canon/proof-receipts/") || !strings.HasSuffix(r.ReceiptPath, ".jsonl") || !sourceLaneProofDigest(r.ReceiptSHA256) {
		return false
	}
	for _, s := range r.Limitations {
		if !validSourceID(s) {
			return false
		}
	}
	seen := map[string]bool{}
	roles := map[string]bool{}
	testBound := false
	for _, in := range r.Inputs {
		if seen[in.Path] || !sourceLaneProofSafePath(in.Path) || !sourceLaneProofDigest(in.SHA256) {
			return false
		}
		seen[in.Path] = true
		switch in.Role {
		case "code", "test":
			if (!strings.HasPrefix(in.Path, "cmd/") && !strings.HasPrefix(in.Path, "internal/")) || !strings.HasSuffix(in.Path, ".go") {
				return false
			}
		case "dependency":
			if in.Path != "go.mod" && in.Path != "go.sum" {
				return false
			}
		default:
			return false
		}
		roles[in.Role] = true
		testBound = testBound || in.Role == "test" && in.Path == r.TestPath
	}
	if !roles["code"] || !roles["test"] || !roles["dependency"] || !testBound || !seen["go.mod"] || !seen["go.sum"] {
		return false
	}
	targets := map[sourceLaneTargetRef]bool{}
	for _, ref := range r.Targets {
		if targets[ref] || ref.Connector != r.Key.Connector || ref.Lane != r.Lane || !validSourceID(ref.ID) || !sourceLaneProofSafePath(ref.Artifact) || !strings.HasPrefix(ref.Artifact, "internal/connectors/defs/"+r.Key.Connector+"/") || !strings.HasSuffix(ref.Artifact, ".json") || !sourceLaneProofDigest(ref.ArtifactSHA256) || !strings.HasPrefix(ref.Pointer, "/") || !validSourceID(ref.CanonicalID) || !strings.HasPrefix(ref.CanonicalPointer, "/") || !sourceLaneProofDigest(ref.Generation) {
			return false
		}
		targets[ref] = true
	}
	return true
}

func sourceLaneReadProof(root *os.Root, r sourceLaneProofRecord, budget *int64) (string, string) {
	read := func(p string) ([]byte, string) {
		if *budget <= 0 {
			return nil, "proof_read_budget_exceeded"
		}
		if _, err := root.Lstat(p); errors.Is(err, fs.ErrNotExist) {
			return nil, "missing"
		}
		raw, err := readSourceInput(root, p, min(*budget, 4<<20))
		if err != nil {
			return nil, "proof_input_invalid"
		}
		*budget -= int64(len(raw))
		return raw, ""
	}
	stale := func(code string) (string, string) {
		if r.ClaimCurrent {
			return "proof_current_claim_invalid", "error"
		}
		return code, "deficit"
	}
	for _, in := range r.Inputs {
		raw, code := read(in.Path)
		if code == "missing" {
			return stale("proof_inputs_outdated")
		}
		if code != "" {
			return code, "error"
		}
		if sourceBytesHash(raw) != in.SHA256 {
			return stale("proof_inputs_outdated")
		}
	}
	for _, ref := range r.Targets {
		raw, code := read(ref.Artifact)
		if code == "missing" {
			return stale("proof_targets_outdated")
		}
		if code != "" {
			return code, "error"
		}
		if sourceBytesHash(raw) != ref.ArtifactSHA256 {
			return stale("proof_targets_outdated")
		}
	}
	raw, code := read(r.ReceiptPath)
	if code == "missing" {
		return stale("proof_receipt_unavailable")
	}
	if code != "" {
		return code, "error"
	}
	if sourceBytesHash(raw) != r.ReceiptSHA256 {
		return "proof_receipt_digest_invalid", "error"
	}
	if !sourceLaneProofSuccessfulResult(raw, r) {
		return "proof_result_invalid", "error"
	}
	return "", ""
}

// The reviewed original receipt supplies execution authenticity; this checks
// actual selected result events, not a PASS string or declaration. No code or
// command from the document is ever executed.
func sourceLaneProofSuccessfulResult(raw []byte, r sourceLaneProofRecord) bool {
	runs, passes, packagePass := 0, 0, 0
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte{'\n'}) {
		var event struct {
			Time        string
			Action      string
			Package     string
			Test        string
			Output      string
			Elapsed     float64
			FailedBuild string
		}
		if decodeStrictJSON(line, &event) != nil || event.Package != r.Package {
			return false
		}
		switch event.Action {
		case "start", "run", "pause", "cont", "output", "pass":
		case "fail", "skip":
			return false
		default:
			return false
		}
		if event.Test == r.SelectedTest {
			if event.Action == "run" {
				runs++
				if passes != 0 {
					return false
				}
			}
			if event.Action == "pass" {
				passes++
				if runs != 1 {
					return false
				}
			}
		}
		if event.Test == "" && event.Action == "pass" {
			packagePass++
			if passes != 1 {
				return false
			}
		}
	}
	return runs == 1 && passes == 1 && packagePass == 1
}

// assessSourceLaneProof preserves source membership and independently derived
// exclusions/gaps. Evidence absence is an explicit deficit, never applicability.
func assessSourceLaneProof(key sourceOperationKey, cells []sourceLaneCell, inputs sourceLaneProofInputs) []sourceLaneCell {
	out := append([]sourceLaneCell(nil), cells...)
	for i := range out {
		out[i].Diagnostics = append([]sourceLaneDiagnostic(nil), out[i].Diagnostics...)
		out[i].ProofRefs = []string{}
		for _, d := range inputs.Diagnostics {
			if d.Key == (sourceOperationKey{}) || d.Key == key {
				if d.Key == (sourceOperationKey{}) {
					d.Key = key
					d.Owner = key.Connector
					d.Lanes = []string{out[i].Lane}
				}
				if len(d.Lanes) == 0 || d.Lanes[0] == "" || d.Lanes[0] == out[i].Lane {
					out[i].Diagnostics = append(out[i].Diagnostics, d)
				}
			}
		}
		if out[i].State != "not_applicable" && out[i].State != "missing_foundation" {
			out[i].State = "mapped_unproven"
			for _, r := range inputs.Records {
				if r.Key != key || r.Lane != out[i].Lane || !inputs.accepted[r.ID] {
					continue
				}
				blocked := out[i].Applicability != "applicable" || len(out[i].References) == 0
				for _, d := range out[i].Diagnostics {
					blocked = blocked || d.Severity == "error" || d.Stage == "reference"
				}
				for _, required := range r.Targets {
					found := false
					for _, ref := range out[i].References {
						found = found || ref == required
					}
					blocked = blocked || !found
				}
				if blocked {
					out[i].Diagnostics = append(out[i].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{out[i].Lane}, Stage: "proof", Code: "proof_prerequisites_unproven", Owner: key.Connector, Severity: "deficit"})
					continue
				}
				out[i].State = "implemented"
				out[i].ProofRefs = []string{r.ID}
				out[i].Reason = sourceLaneReason{Code: "behavior_proven", Text: "Reviewed lane-specific behavior is current within the proof's declared scope."}
			}
			if out[i].State != "implemented" {
				out[i].Diagnostics = append(out[i].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{out[i].Lane}, Stage: "proof", Code: "proof_unavailable", Owner: key.Connector, Severity: "deficit"})
			}
		}
	}
	return out
}
