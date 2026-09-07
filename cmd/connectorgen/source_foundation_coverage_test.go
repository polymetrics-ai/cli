package main

import (
	"reflect"
	"testing"
)

func TestSourceFoundationCoverageExactMembership(t *testing.T) {
	repo, universe, _ := sourceFoundationAssessmentFixture(t)
	observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	coverage, err := buildSourceFoundationCoverage(observed)
	if err != nil {
		t.Fatal(err)
	}
	wantAssessed := []sourceFoundationCell{{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "source.a"}, Lane: "sync_transport"}}
	wantComplement := []sourceFoundationCell{}
	for _, id := range []string{"source.a", "source.b"} {
		for _, lane := range []string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"} {
			if id != "source.a" || lane != "sync_transport" {
				wantComplement = append(wantComplement, sourceFoundationCell{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: id}, Lane: lane})
			}
		}
	}
	if coverage.UniverseCount != 14 || !reflect.DeepEqual(coverage.Assessed, wantAssessed) ||
		!reflect.DeepEqual(coverage.Unassessed, wantComplement) {
		t.Fatalf("exact source/lane accounting differs: %+v", coverage)
	}
	if err := validateSourceFoundationCoverage(coverage, observed); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"same count substituted key", "missing complement", "extra complement", "duplicate complement",
		"assessed moved to complement", "wrong digest", "forged universe count", "reversed complement", "null assessed", "forged projected universe"} {
		t.Run(scenario, func(t *testing.T) {
			candidate, err := buildSourceFoundationCoverage(observed)
			if err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "same count substituted key":
				candidate.Unassessed[0].Key.ID = "source.other"
			case "missing complement":
				candidate.Unassessed = candidate.Unassessed[1:]
			case "extra complement":
				candidate.Unassessed = append(candidate.Unassessed, sourceFoundationCell{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "source.other"}, Lane: "sync_transport"})
			case "duplicate complement":
				candidate.Unassessed[1] = candidate.Unassessed[0]
			case "assessed moved to complement":
				candidate.Unassessed = append(candidate.Unassessed, candidate.Assessed[0])
				candidate.Assessed = []sourceFoundationCell{}
			case "wrong digest":
				candidate.UniverseSHA256 = sourceBytesHash([]byte("wrong universe"))
			case "forged universe count":
				candidate.UniverseCount--
			case "reversed complement":
				candidate.Unassessed[0], candidate.Unassessed[1] = candidate.Unassessed[1], candidate.Unassessed[0]
			case "null assessed":
				candidate.Assessed = nil
			case "forged projected universe":
				forged := observed
				forged.universe.cells = append([]sourceFoundationCell{}, observed.universe.cells...)
				forged.universe.cells[0].Key.ID = "source.other"
				candidate, err = buildSourceFoundationCoverage(forged)
				if err != nil {
					t.Fatal(err)
				}
			}
			// Keep summaries and both changed arrays coherently hashed. The
			// independent original input set, not self-consistency, must win.
			candidate.AssessedSHA256, err = sourceFoundationCellsHash(candidate.Assessed)
			if err != nil {
				t.Fatal(err)
			}
			candidate.UnassessedSHA256, err = sourceFoundationCellsHash(candidate.Unassessed)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateSourceFoundationCoverage(candidate, observed); err == nil {
				t.Fatalf("coherent but false accounting accepted: %+v", candidate)
			}
		})
	}
}
