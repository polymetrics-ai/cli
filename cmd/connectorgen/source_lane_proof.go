package main

import (
	"os"
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
	root, err := os.OpenRoot(repo)
	if err != nil {
		return result
	}
	defer root.Close()
	raw, err := readSourceInput(root, sourceLaneProofPath, 4<<20)
	if err != nil {
		return result
	}
	var document struct {
		SchemaVersion int                     `json:"schema_version"`
		Records       []sourceLaneProofRecord `json:"records"`
	}
	if decodeStrictJSON(raw, &document) == nil {
		result.Records = document.Records
	}
	return result
}

// assessSourceLaneProof preserves source membership and independently derived
// exclusions/gaps. Evidence absence is an explicit deficit, never applicability.
func assessSourceLaneProof(key sourceOperationKey, cells []sourceLaneCell, inputs sourceLaneProofInputs) []sourceLaneCell {
	out := append([]sourceLaneCell(nil), cells...)
	for i := range out {
		out[i].ProofRefs = []string{}
		if out[i].State != "not_applicable" && out[i].State != "missing_foundation" {
			out[i].State = "mapped_unproven"
			out[i].Diagnostics = append(append([]sourceLaneDiagnostic(nil), out[i].Diagnostics...), sourceLaneDiagnostic{Key: key, Lanes: []string{out[i].Lane}, Stage: "proof", Code: "proof_unavailable", Owner: key.Connector, Severity: "deficit"})
		}
	}
	return out
}
