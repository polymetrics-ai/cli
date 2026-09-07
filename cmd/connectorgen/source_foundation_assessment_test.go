package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sourceFoundationAssessmentFixture(t *testing.T) (string, sourceFoundationUniverse, sourceFoundationAssessmentDocument) {
	t.Helper()
	repo, cohort := sourceFoundationUniverseFixture(t)
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	document := sourceFoundationAssessmentDocumentFixture(t, universe)
	writeSourceFoundationAssessmentFixture(t, repo, document)
	return repo, universe, document
}

func sourceFoundationAssessmentDocumentFixture(t *testing.T, universe sourceFoundationUniverse) sourceFoundationAssessmentDocument {
	t.Helper()
	ref, exists := universe.manifest.SourceOperations[0].Facts.Refs["source_operation"]
	if !exists {
		t.Fatal("actual retained operation citation missing")
	}
	document := sourceFoundationAssessmentDocument{SchemaVersion: 1, Kind: "foundation_cell_assessments", Atlas: universe.atlasPin,
		Assessments: []sourceFoundationCellAssessment{{
			Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "source.a"}, Lane: "sync_transport",
			NextOwner: "polymetrics.ai/internal/synctransport",
			Requirements: []sourceFoundationRequirement{{ID: "transport-fit", SourceRefs: []sourceFactRef{ref},
				Statement: "Assess whether the retained read has a supported warehouse transport mode",
				AtlasLookup: sourceFoundationLookup{Atlas: universe.atlasPin, Candidates: []sourceFoundationLookupCandidate{{
					AtlasID: "transport.sync-contract.v1", Contract: sourceFoundationContractRef{
						Pointer: "/supported_contracts/sync_modes", ValueSHA256: "75525380daccb562ea7c2d8df0f5f6905f89535bc48d329d4f5633f88d32348f"},
					Disposition: "unresolved", Rationale: "Mode vocabulary does not establish the source and executor intersection"}}},
				Assessment: "unresolved", ProofIDs: []string{}, AffectedArtifacts: []string{},
				EvidenceRequirements: []string{"source-backed executor and mode fit"}, DecisionRefs: []sourceFoundationDecision{},
			}},
		}}}
	return document
}

func writeSourceFoundationAssessmentFixture(t *testing.T, repo string, document sourceFoundationAssessmentDocument) []byte {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(repo, sourceFoundationAssessmentsPath)
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestSourceFoundationAssessmentObservation(t *testing.T) {
	repo, universe, document := sourceFoundationAssessmentFixture(t)
	raw := writeSourceFoundationAssessmentFixture(t, repo, document)
	manifestBefore, err := json.Marshal(universe.manifest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.authored, document.Assessments) || len(got.universe.cells) != 14 ||
		got.pin != (sourceArtifactPin{Path: sourceFoundationAssessmentsPath, SHA256: sourceBytesHash(raw), Bytes: int64(len(raw))}) {
		t.Fatalf("actual source/assessment observation lost identity: %+v", got)
	}
	if got.atlas.entries["transport.sync-contract.v1"].Owner.PrimaryPackage != "polymetrics.ai/internal/synctransport" {
		t.Fatal("actual rich Atlas owner not observed")
	}
	manifestAfter, err := json.Marshal(got.universe.manifest)
	if err != nil || !bytes.Equal(manifestBefore, manifestAfter) {
		t.Fatalf("authoring observation altered lane manifest: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(repo, sourceFoundationAssessmentsPath))
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatalf("read-only assessment observation changed input: %v", err)
	}
}

func TestSourceFoundationAssessmentAdmission(t *testing.T) {
	for _, scenario := range []string{
		"unknown source", "unknown lane", "duplicate cell", "empty requirements", "duplicate requirement",
		"missing source citation", "wrong source digest", "sibling source citation", "missing next owner", "unknown next owner",
		"missing lookup", "stale lookup", "unknown atlas owner", "wrong contract occurrence", "wrong contract hash",
		"duplicate candidate", "missing mismatch rationale", "unknown assessment", "missing evidence requirement",
	} {
		t.Run(scenario, func(t *testing.T) {
			repo, universe, document := sourceFoundationAssessmentFixture(t)
			// The unchanged positive input must reach the actual source/Atlas/
			// assessment reader before a single admission invariant is challenged.
			if got, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe); err != nil || len(got.authored) != 1 {
				t.Fatalf("valid reader frontier unavailable: %v", err)
			}
			cell := &document.Assessments[0]
			req := &cell.Requirements[0]
			switch scenario {
			case "unknown source":
				cell.Key.ID = "source.other"
			case "unknown lane":
				cell.Lane = "invented_lane"
			case "duplicate cell":
				document.Assessments = append(document.Assessments, *cell)
			case "empty requirements":
				cell.Requirements = []sourceFoundationRequirement{}
			case "duplicate requirement":
				cell.Requirements = append(cell.Requirements, *req)
			case "missing source citation":
				req.SourceRefs = []sourceFactRef{}
			case "wrong source digest":
				req.SourceRefs[0].ValueSHA256 = sourceBytesHash([]byte("wrong"))
			case "sibling source citation":
				req.SourceRefs[0] = universe.manifest.SourceOperations[1].Facts.Refs["source_operation"]
			case "missing next owner":
				cell.NextOwner = ""
			case "unknown next owner":
				cell.NextOwner = "polymetrics.ai/internal/invented"
			case "missing lookup":
				req.AtlasLookup.Candidates = []sourceFoundationLookupCandidate{}
			case "stale lookup":
				req.AtlasLookup.Atlas.SHA256 = sourceBytesHash([]byte("old atlas"))
			case "unknown atlas owner":
				req.AtlasLookup.Candidates[0].AtlasID = "invented.foundation.v1"
			case "wrong contract occurrence":
				req.AtlasLookup.Candidates[0].Contract.Pointer = "/constraints/0"
			case "wrong contract hash":
				req.AtlasLookup.Candidates[0].Contract.ValueSHA256 = sourceBytesHash([]byte("wrong"))
			case "duplicate candidate":
				req.AtlasLookup.Candidates = append(req.AtlasLookup.Candidates, req.AtlasLookup.Candidates[0])
			case "missing mismatch rationale":
				req.AtlasLookup.Candidates[0].Rationale = ""
			case "unknown assessment":
				req.Assessment = "implemented"
			case "missing evidence requirement":
				req.EvidenceRequirements = []string{}
			}
			writeSourceFoundationAssessmentFixture(t, repo, document)
			got, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err == nil || len(got.authored) != 0 {
				t.Fatalf("invalid assessment reached returned observation: rows=%d error=%v", len(got.authored), err)
			}
		})
	}
}

func TestSourceFoundationAssessmentKnownOmission(t *testing.T) {
	repo, cohort := sourceFoundationUniverseFixture(t)
	name := filepath.Join(repo, "source.json")
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var source map[string]any
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	// A current extra update/header requirement, not a copied historical
	// receiver label, must enter K through the actual retained-source producer.
	rows := source["rest"].(map[string]any)["operations"].([]any)
	for _, value := range rows {
		row := value.(map[string]any)
		if row["id"] == "source.b" {
			row["method"] = "put"
			row["source_operation"].(map[string]any)["summary"] = "Update webhook headers"
		}
	}
	raw, err = json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, raw, 0600); err != nil {
		t.Fatal(err)
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	var required sourceLaneManifestRow
	for _, row := range universe.manifest.SourceOperations {
		if row.Source.Key.ID == "source.b" {
			required = row
		}
	}
	ref, demanded := sourceRegistrationDemand(required.Facts)
	if !demanded || required.Lanes[6].State != "mapped_unproven" {
		t.Fatal("source producer did not establish the intended extra obligation")
	}
	document := sourceFoundationAssessmentDocumentFixture(t, universe)
	writeSourceFoundationAssessmentFixture(t, repo, document)
	got, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err == nil || len(got.authored) != 0 {
		t.Errorf("omitting independently known update/header cell returned %d assessments: %v", len(got.authored), err)
	}
	complete := document.Assessments[0]
	complete.Key = required.Source.Key
	complete.Requirements = append([]sourceFoundationRequirement{}, complete.Requirements...)
	complete.Requirements[0].SourceRefs = []sourceFactRef{ref}
	document.Assessments = append(document.Assessments, complete)
	writeSourceFoundationAssessmentFixture(t, repo, document)
	got, err = observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil || len(got.authored) != 2 {
		t.Fatalf("source-cited unresolved known requirement refused: rows=%d error=%v", len(got.authored), err)
	}
}
