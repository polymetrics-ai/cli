package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSourceFoundationFacetObservation(t *testing.T) {
	repo, universe, _ := sourceFoundationAssessmentFixture(t)
	observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := buildSourceFoundationFacets(t.Context(), observed)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]int{"mime": 2, "body": 2, "auth": 5, "paging": 2}
	if len(rows) != len(expected) {
		t.Fatal("lost one of the four independently required facet dimensions")
	}
	for _, row := range rows {
		count, exists := expected[row.Kind]
		if !exists || len(row.CandidateOwners) != count || row.Identity.Key.ID != "source.a" || row.Identity.Lane != "sync_transport" ||
			row.Status != "unresolved" || len(row.SourceRefs) == 0 || row.Limitation == "" {
			t.Fatal("actual source facet observation lost scope, owner candidates or uncertainty")
		}
		delete(expected, row.Kind)
		for _, owner := range row.CandidateOwners {
			if owner.Owner != "polymetrics.ai/internal/connectors/engine" || len(owner.Symbols) == 0 || len(owner.Constraints) == 0 ||
				!sourceProofDigest(owner.Contract.ValueSHA256) {
				t.Fatal("facet lookup lost actual Atlas owner/constraint/contract identity")
			}
		}
	}
}

func TestSourceFoundationFacetSourceSemantics(t *testing.T) {
	for _, kind := range []string{"mime", "auth", "body"} {
		t.Run(kind, func(t *testing.T) {
			repo, cohort := sourceFoundationUniverseFixture(t)
			path := filepath.Join(repo, "source.json")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var retained map[string]any
			if err := json.Unmarshal(raw, &retained); err != nil {
				t.Fatal(err)
			}
			for _, value := range retained["rest"].(map[string]any)["operations"].([]any) {
				operation := value.(map[string]any)
				if operation["id"] != "source.a" {
					continue
				}
				var contract any
				if err := json.Unmarshal([]byte(`{"requestBody":{"required":true,"content":{"application/vnd.fixture+json":{"schema":{"type":"object"}}}},"security":[{"token":["read:one","read:two"]},{}],"responses":{"200":{"content":{"application/json":{"schema":{"type":"object"}}}},"2XX":{"content":{"application/octet-stream":{"schema":{"type":"string","format":"binary"}}}},"2bad":{"content":{"application/xml":{"schema":{"type":"string"}}}},"400":{"content":{"application/problem+json":{"schema":{"type":"object"}}}}}}`), &contract); err != nil {
					t.Fatal(err)
				}
				operation["source_operation"] = contract
			}
			raw, err = json.Marshal(retained)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
			universe, err := buildSourceFoundationUniverse(t.Context(), repo, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			writeSourceFoundationAssessmentFixture(t, repo, sourceFoundationAssessmentDocumentFixture(t, universe))
			observed, err := observeSourceFoundationAssessments(t.Context(), repo, sourceFoundationAssessmentsPath, universe)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := buildSourceFoundationFacets(t.Context(), observed)
			if err != nil {
				t.Fatal(err)
			}
			expected := map[string][]string{
				"mime": {"request_media:application/vnd.fixture+json", "response_media:200:application/json", "response_media:2XX:application/octet-stream", "response_status_unresolved:2bad"},
				"auth": {"security_alternative_0_scheme:token", "security_alternative_0_scheme:token_scope:read:one", "security_alternative_0_scheme:token_scope:read:two", "security_alternative_1_anonymous"},
				"body": {"request_media:application/vnd.fixture+json", "request_required:true"},
			}
			found := false
			for _, row := range rows {
				if row.Kind == kind {
					found = true
					if !reflect.DeepEqual(row.ObservedValues, expected[kind]) || row.Status != "unresolved" {
						t.Errorf("actual retained %s semantics=%v; want %v with no automatic capability claim", kind, row.ObservedValues, expected[kind])
					}
				}
			}
			if !found {
				t.Fatal("selected facet omitted")
			}
		})
	}
}
