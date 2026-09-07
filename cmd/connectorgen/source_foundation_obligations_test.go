package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sourceFoundationObligationsFixture(t *testing.T, repo string, universe sourceFoundationUniverse) (sourceFoundationObligationDocument, sourceArtifactPin) {
	t.Helper()
	row := universe.manifest.SourceOperations[0]
	cell := sourceFoundationCell{Key: row.Source.Key, Lane: "sync_transport"}
	document := sourceFoundationObligationDocument{SchemaVersion: 1, Kind: "foundation_known_obligations",
		BaselineManifest:             sourceArtifactPin{Path: sourceLaneManifestPath, SHA256: sourceBytesHash([]byte("independently retained fixture manifest")), Bytes: 39},
		SourceEvidenceSHA256:         sourceBytesHash([]byte("independent fixture requirement")),
		Obligations:                  []sourceFoundationKnownObligation{{Identity: cell, Source: row.Source, SourceRefs: []sourceFactRef{row.Facts.Refs["source_operation"]}}},
		HistoricalReceiverCandidates: []sourceFoundationCell{cell}, SentryRegistration: []sourceFoundationCell{}}
	var err error
	document.BaselineUniverseSHA256, err = sourceFoundationCellsHash(universe.cells)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range universe.manifest.Documents {
		document.Obligations[0].SourceDocumentPins = append(document.Obligations[0].SourceDocumentPins,
			sourceArtifactPin{Path: source.Path, SHA256: source.RetainedFileSHA256, Bytes: source.Bytes})
	}
	return document, writeSourceFoundationObligationsFixture(t, repo, document)
}

func writeSourceFoundationObligationsFixture(t *testing.T, repo string, document sourceFoundationObligationDocument) sourceArtifactPin {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, sourceFoundationObligationsPath), raw, 0600); err != nil {
		t.Fatal(err)
	}
	return sourceArtifactPin{Path: sourceFoundationObligationsPath, SHA256: sourceBytesHash(raw), Bytes: int64(len(raw))}
}

func TestSourceFoundationObligationsObservation(t *testing.T) {
	repo, universe, _ := sourceFoundationAssessmentFixture(t)
	document, pin := sourceFoundationObligationsFixture(t, repo, universe)
	got, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, pin)
	if err != nil {
		t.Fatal(err)
	}
	// Node is a private parsed-source cache, deliberately absent from JSON.
	document.Obligations[0].Source.Node = nil
	if !reflect.DeepEqual(got.baseline, document) || got.baselinePin != pin || len(got.assessments.authored) != 1 {
		t.Fatal("actual independent baseline/source/assessment observation lost")
	}
}

func TestSourceFoundationObligationsIndependentInputPins(t *testing.T) {
	for _, scenario := range []string{"changed uncited source bytes", "same-count unassessed key replacement"} {
		t.Run(scenario, func(t *testing.T) {
			repo, cohort := sourceFoundationUniverseFixture(t)
			universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			writeSourceFoundationAssessmentFixture(t, repo, sourceFoundationAssessmentDocumentFixture(t, universe))
			baseline, pin := sourceFoundationObligationsFixture(t, repo, universe)
			if _, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, pin); err != nil {
				t.Fatalf("original independently bound source/control: %v", err)
			}
			path := filepath.Join(repo, "source.json")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var source map[string]any
			if err := json.Unmarshal(raw, &source); err != nil {
				t.Fatal(err)
			}
			if scenario == "changed uncited source bytes" {
				source["x-cp13-uncited-note"] = "changed retained metadata"
			} else {
				for _, value := range source["rest"].(map[string]any)["operations"].([]any) {
					operation := value.(map[string]any)
					if operation["id"] == "source.b" {
						operation["id"] = "source.other"
					}
				}
				cohort.Inventories[0].ExpectedIDs = []string{"source.a", "source.other"}
			}
			raw, err = json.Marshal(source)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
			current, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
			if err != nil {
				t.Fatalf("current source producer must succeed before independent baseline check: %v", err)
			}
			if scenario == "same-count unassessed key replacement" {
				// Rebind the file in the independent fixture policy to isolate
				// the original full-universe identity invariant from file custody.
				baseline.Obligations[0].SourceDocumentPins[0].SHA256 = sourceBytesHash(raw)
				baseline.Obligations[0].SourceDocumentPins[0].Bytes = int64(len(raw))
				pin = writeSourceFoundationObligationsFixture(t, repo, baseline)
			}
			if got, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, current, pin); err == nil || len(got.assessments.authored) != 0 {
				t.Errorf("actual demand input consumer accepted %s", scenario)
			}
		})
	}
}

func TestSourceFoundationObligationsAdmission(t *testing.T) {
	for _, scenario := range []string{"omitted independent cell", "duplicate obligation", "wrong source identity", "wrong source pointer", "wrong citation", "missing citations", "missing historical member", "missing sentry member", "rewritten seed without policy"} {
		t.Run(scenario, func(t *testing.T) {
			repo, universe, _ := sourceFoundationAssessmentFixture(t)
			document, pin := sourceFoundationObligationsFixture(t, repo, universe)
			if _, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, pin); err != nil {
				t.Fatalf("complete positive control: %v", err)
			}
			second := universe.manifest.SourceOperations[1]
			secondCell := sourceFoundationCell{Key: second.Source.Key, Lane: "sync_transport"}
			switch scenario {
			case "omitted independent cell":
				// This real retained source is deliberately not selected by the
				// current registration heuristic. Historical policy still owns it.
				document.Obligations = append(document.Obligations, sourceFoundationKnownObligation{Identity: secondCell, Source: second.Source,
					SourceRefs: []sourceFactRef{second.Facts.Refs["source_operation"]}})
			case "duplicate obligation":
				document.Obligations = append(document.Obligations, document.Obligations[0])
			case "wrong source identity":
				document.Obligations[0].Source.Key = second.Source.Key
			case "wrong source pointer":
				document.Obligations[0].Source.Pointer = second.Source.Pointer
			case "wrong citation":
				document.Obligations[0].SourceRefs[0] = second.Facts.Refs["source_operation"]
			case "missing citations":
				document.Obligations[0].SourceRefs = []sourceFactRef{}
			case "missing historical member":
				document.HistoricalReceiverCandidates = append(document.HistoricalReceiverCandidates, secondCell)
			case "missing sentry member":
				document.SentryRegistration = append(document.SentryRegistration, secondCell)
			case "rewritten seed without policy":
				document.Obligations = []sourceFoundationKnownObligation{}
			}
			updated := writeSourceFoundationObligationsFixture(t, repo, document)
			if scenario != "rewritten seed without policy" {
				pin = updated // independent test policy, never authored assessment authority
			}
			if _, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, pin); err == nil {
				t.Errorf("real demand input reader accepted %s", scenario)
			}
		})
	}
}

func TestSourceFoundationObligationsCurrentCorpus(t *testing.T) {
	repo, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := readSourceInput(root, sourceLaneCohortPath, 64<<20)
	var cohort sourceLaneCohort
	if err != nil || decodeStrictJSON(raw, &cohort) != nil {
		t.Fatalf("actual cohort: %v", err)
	}
	raw, err = readSourceInput(root, sourceLaneAnnotationsPath, 64<<20)
	var annotations struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}
	if err != nil || decodeStrictJSON(raw, &annotations) != nil {
		t.Fatalf("actual annotations: %v", err)
	}
	universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, annotations.Annotations)
	if err != nil {
		t.Fatal(err)
	}
	got, err := observeSourceFoundationDemandInputs(t.Context(), repo, sourceFoundationAssessmentsPath, universe, sourceFoundationBaselinePin())
	if err != nil {
		t.Fatal(err)
	}
	coverage, err := buildSourceFoundationCoverage(got.assessments)
	if err != nil || validateSourceFoundationCoverage(coverage, got.assessments) != nil {
		t.Fatalf("actual current corpus membership: %v", err)
	}
	if coverage.UniverseCount != 30401 || len(universe.manifest.SourceOperations) != 4343 || len(coverage.Assessed) != 26 ||
		len(coverage.Unassessed) != 30375 || len(got.baseline.HistoricalReceiverCandidates) != 12 || len(got.baseline.SentryRegistration) != 1 {
		t.Fatal("retained cohort/baseline accounting changed without explicit reconciliation")
	}
	for _, row := range universe.manifest.SourceOperations {
		for _, lane := range row.Lanes {
			if lane.State == "implemented" || len(lane.ProofRefs) != 0 {
				t.Fatal("authoring foundation reconciliation promoted runtime authority")
			}
		}
	}
	requirements, err := buildSourceFoundationRequirements(t.Context(), repo, got.assessments)
	if err != nil || len(requirements) != 26 {
		t.Fatalf("current retained source requirements: %v", err)
	}
	byCell := map[sourceFoundationCell]sourceFoundationRequirementResult{}
	for _, requirement := range requirements {
		byCell[requirement.Identity] = requirement
	}
	historicalUnproven, historicalGap := 0, 0
	for _, cell := range got.baseline.HistoricalReceiverCandidates {
		row := byCell[cell]
		switch row.SourceState {
		case "mapped_unproven":
			historicalUnproven++
		case "missing_foundation":
			historicalGap++
		default:
			t.Fatal("historical source receiver candidate lost its distinct state")
		}
	}
	vercel := byCell[sourceFoundationCell{Key: sourceOperationKey{Connector: "vercel", Inventory: "primary", ID: "vercel.rest.createWebhook"}, Lane: "sync_transport"}]
	sentry := byCell[got.baseline.SentryRegistration[0]]
	if historicalUnproven != 11 || historicalGap != 1 || vercel.Status != "absent_shared_foundation" ||
		vercel.SourceState != "missing_foundation" || len(vercel.GapRefs) != 1 || vercel.GapRefs[0] != "cli-webhook-event-surface-foundation-r1" ||
		len(vercel.DecisionRefs) != 2 || len(vercel.RetainedDecisionOwners) != 2 ||
		sentry.Status != "unresolved" || sentry.SourceState != "mapped_unproven" || len(sentry.GapRefs) != 0 {
		t.Fatal("current Vercel authority/historical12/Sentry distinction drifted")
	}
}
