package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"strings"
)

const sourceFoundationProofPath = "docs/connector-canon/foundations/proofs.json"
const sourceFoundationReceiptPrefix = "data/connector-canon/proof-receipts/foundation/"

type sourceFoundationContractRef struct {
	Pointer     string `json:"pointer"`
	ValueSHA256 string `json:"value_sha256"`
}

type sourceFoundationProofTest struct {
	File     string `json:"file"`
	Symbol   string `json:"symbol"`
	Selected string `json:"selected"`
	Package  string `json:"package"`
}

type sourceFoundationAssertion struct {
	Statement    string `json:"statement"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
	SourceSHA256 string `json:"source_sha256"`
}

type sourceFoundationProofInput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Role   string `json:"role"`
	Bytes  int64  `json:"bytes"`
}

type sourceFoundationProofRecord struct {
	ID             string                       `json:"id"`
	AtlasID        string                       `json:"atlas_id"`
	Contract       sourceFoundationContractRef  `json:"contract"`
	OwnerSymbols   []sourceFoundationSymbol     `json:"owner_symbols"`
	Test           sourceFoundationProofTest    `json:"test"`
	Assertion      sourceFoundationAssertion    `json:"assertion"`
	Inputs         []sourceFoundationProofInput `json:"inputs"`
	Capture        sourceArtifactPin            `json:"capture"`
	Output         sourceArtifactPin            `json:"output"`
	ExecutionClass string                       `json:"execution_class"`
	Scope          string                       `json:"scope"`
	Limitations    []string                     `json:"limitations"`
}

// Observation precedes admission. This type carries no accepted lane targets
// and cannot be passed to the lane-proof reducer.
type sourceFoundationProofObservation struct {
	record sourceFoundationProofRecord
	atlas  sourceFoundationAtlasEntry
	status string
	issues []sourceFoundationProofIssue
}

type sourceFoundationProofDocument struct {
	SchemaVersion int                           `json:"schema_version"`
	Kind          string                        `json:"kind"`
	Atlas         sourceArtifactPin             `json:"atlas"`
	Records       []sourceFoundationProofRecord `json:"records"`
}

type sourceFoundationProofReview struct {
	ID, RecordSHA256, InputClosureSHA256, AssertionSHA256 string
}

// Only the trusted authoring caller supplies review authorization. None of
// these fields can be supplied by proof JSON or generated demand assessments.
type sourceFoundationProofCatalog struct {
	Reviews []sourceFoundationProofReview
}

// readSourceFoundationProofObservations establishes the real record/Atlas
// projection before assertion admission. A returned observation is not current
// proof: execution, closure, scope and independent review bindings follow it.
func readSourceFoundationProofObservations(ctx context.Context, repo string) ([]sourceFoundationProofObservation, error) {
	return readSourceFoundationProofsReviewed(ctx, repo, reviewedSourceFoundationProofs())
}

func readSourceFoundationProofsReviewed(ctx context.Context, repo string, catalog sourceFoundationProofCatalog) ([]sourceFoundationProofObservation, error) {
	return readSourceFoundationProofsObserved(ctx, repo, catalog, nil)
}

// The observer runs only after real confined reads, for deterministic custody
// evidence. It supplies no record, review authority, result or execution target.
func readSourceFoundationProofsObserved(ctx context.Context, repo string, catalog sourceFoundationProofCatalog, observer func(sourceProofReadEvent)) (result []sourceFoundationProofObservation, resultErr error) {
	// Cancellation may be observed by a shared reader before the next explicit
	// context check. Preserve both that scoped failure and the inspectable cause.
	defer func() {
		if err := ctx.Err(); err != nil {
			result, resultErr = nil, errors.Join(resultErr, err)
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	batch, err := readSourceFoundationProofBatch(ctx, repo, catalog, observer)
	return batch.records, err
}

func sourceFoundationReviewMatches(record sourceFoundationProofRecord, review sourceFoundationProofReview) bool {
	digest, err := sourceFoundationProofRecordHash(record)
	if err != nil || digest != review.RecordSHA256 {
		return false
	}
	for _, part := range []struct {
		value any
		want  string
	}{{record.Inputs, review.InputClosureSHA256}, {record.Assertion, review.AssertionSHA256}} {
		raw, err := json.Marshal(part.value)
		if err != nil {
			return false
		}
		canonical, err := canonicalSourceJSON(raw)
		if err != nil || sourceBytesHash(canonical) != part.want {
			return false
		}
	}
	return true
}

// Shape is only the closed authoring language. It neither authorizes the
// assertion nor establishes that its declared owner/test was actually reached.
func sourceFoundationProofRecordShape(r sourceFoundationProofRecord) bool {
	if !validSourceID(r.ID) || !validSourceID(r.AtlasID) || r.ExecutionClass != "C2" || r.Scope != "foundation_contract" {
		return false
	}
	if !strings.HasPrefix(r.Contract.Pointer, "/supported_contracts/") || !sourceProofDigest(r.Contract.ValueSHA256) {
		return false
	}
	if !sourceProofSafePath(r.Test.File) || !strings.HasSuffix(r.Test.File, "_test.go") || !token.IsIdentifier(r.Test.Symbol) || !strings.HasPrefix(r.Test.Symbol, "Test") || len(r.Test.Symbol) == 4 {
		return false
	}
	if r.Test.Selected != r.Test.Symbol && (!strings.HasPrefix(r.Test.Selected, r.Test.Symbol+"/") || strings.HasSuffix(r.Test.Selected, "/")) {
		return false
	}
	if !strings.HasPrefix(r.Test.Package, "./") || !sourceProofSafePath(strings.TrimPrefix(r.Test.Package, "./")) {
		return false
	}
	if strings.TrimSpace(r.Assertion.Statement) == "" || r.Assertion.StartLine < 1 || r.Assertion.EndLine < r.Assertion.StartLine || !sourceProofDigest(r.Assertion.SourceSHA256) {
		return false
	}
	for _, pin := range []sourceArtifactPin{r.Capture, r.Output} {
		if !sourceProofSafePath(pin.Path) || !strings.HasPrefix(pin.Path, sourceFoundationReceiptPrefix) || !sourceProofDigest(pin.SHA256) || pin.Bytes < 0 {
			return false
		}
	}
	if len(r.OwnerSymbols) == 0 || len(r.Inputs) == 0 || len(r.Limitations) == 0 {
		return false
	}
	owners := map[sourceFoundationSymbol]bool{}
	for _, owner := range r.OwnerSymbols {
		if owners[owner] || !sourceProofSafePath(owner.File) || !strings.HasSuffix(owner.File, ".go") || !validSourceID(owner.Name) {
			return false
		}
		owners[owner] = true
	}
	last := ""
	module, sum, test := false, false, false
	for _, input := range r.Inputs {
		if input.Path <= last || !sourceProofSafePath(input.Path) || !sourceProofDigest(input.SHA256) || input.Bytes < 0 {
			return false
		}
		last = input.Path
		switch input.Role {
		case "code", "test":
			if !strings.HasSuffix(input.Path, ".go") || (!strings.HasPrefix(input.Path, "cmd/") && !strings.HasPrefix(input.Path, "internal/")) || (input.Role == "test") != strings.HasSuffix(input.Path, "_test.go") {
				return false
			}
		case "dependency":
			if input.Path != "go.mod" && input.Path != "go.sum" {
				return false
			}
			module = module || input.Path == "go.mod"
			sum = sum || input.Path == "go.sum"
		case "fixture":
		default:
			return false
		}
		test = test || input.Path == r.Test.File && input.Role == "test" && input.SHA256 == r.Assertion.SourceSHA256
	}
	limits := map[string]bool{}
	for _, limit := range r.Limitations {
		if strings.TrimSpace(limit) == "" || limits[limit] {
			return false
		}
		limits[limit] = true
	}
	return module && sum && test
}

// Canonical record identity is independent of formatting. Review authorization
// must be supplied separately; this hash alone cannot establish authenticity.
func sourceFoundationProofRecordHash(record sourceFoundationProofRecord) (string, error) {
	raw, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("foundation record encoding: %w", err)
	}
	canonical, err := canonicalSourceJSON(raw)
	if err != nil {
		return "", fmt.Errorf("foundation record canonicalization: %w", err)
	}
	return sourceBytesHash(canonical), nil
}
