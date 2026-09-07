package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// CR-158-01: admission independently contributes to K even when the retained
// historical baseline predates that source requirement. Exercise the command,
// not an authored list masquerading as a completed admission observation.
func TestSourceFoundationAdmissionKnownUnion159(t *testing.T) {
	repo, baseline, _ := sourceFoundationCombinedRetained155(t, true)
	control := sourceFoundationAdmissionCommand155(t, repo, baseline)
	var before sourceFoundationRegister
	if err := json.Unmarshal(control, &before); err != nil {
		t.Fatal(err)
	}
	if len(before.Known) != 1 || len(before.SourceAdmission.MissingCLI) != 1 || len(before.SourceAdmission.MissingCLI[0].Affected) != 1 {
		t.Fatal("positive control did not reach exact completed admission and deduplicated known cell")
	}
	expected := before.Known[0].Identity
	if before.SourceAdmission.MissingCLI[0].Affected[0].Identity != expected {
		t.Fatal("fixture baseline and independently observed admission disagree")
	}
	raw, err := os.ReadFile(filepath.Join(repo, baseline.Path))
	if err != nil {
		t.Fatal(err)
	}
	var prior sourceFoundationObligationDocument
	if err := json.Unmarshal(raw, &prior); err != nil {
		t.Fatal(err)
	}
	prior.Obligations = []sourceFoundationKnownObligation{}
	baseline = writeSourceFoundationObligationsFixture(t, repo, prior)
	var out, diag bytes.Buffer
	if code := runSourceDemandsPolicy(t.Context(), []string{"source-demands", "--repo", repo}, &out, &diag, baseline); code != 0 {
		t.Fatalf("actual command refused valid older baseline: %s", diag.String())
	}
	var got sourceFoundationRegister
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Coverage.UniverseCount != 14 || len(got.Coverage.Assessed) != 1 || len(got.Requirements) != 3 || !reflect.DeepEqual(got.SourceAdmission, before.SourceAdmission) {
		t.Fatal("source/admission/requirement controls changed before known-union assertion")
	}
	if len(got.Known) != 1 || got.Known[0].Identity != expected {
		t.Fatalf("completed admission cell omitted or substituted in K: expected %+v, got %+v", expected, got.Known)
	}
	var source sourceLaneManifest
	raw, err = os.ReadFile(filepath.Join(repo, sourceLaneManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	for _, row := range source.SourceOperations {
		if row.Source.Key != expected.Key {
			continue
		}
		for _, lane := range row.Lanes {
			if lane.Lane == expected.Lane && (!reflect.DeepEqual(got.Known[0].SourceRefs, lane.FactRefs) || got.Known[0].SourceState != lane.State || got.Known[0].Applicability != lane.Applicability) {
				t.Fatal("known cell did not retain actual observed lane facts/state")
			}
		}
	}
	if len(got.Known[0].SourceRefs) == 0 || len(got.Known[0].SourceDocumentPins) == 0 {
		t.Fatal("known obligation lost source custody")
	}
	for _, pin := range got.Known[0].SourceDocumentPins {
		data, err := os.ReadFile(filepath.Join(repo, pin.Path))
		if err != nil {
			t.Fatal(err)
		}
		if sourceBytesHash(data) != pin.SHA256 || int64(len(data)) != pin.Bytes {
			t.Fatal("known document pin does not describe actual retained source")
		}
	}
}
