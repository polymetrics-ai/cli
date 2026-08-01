package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type bambooHRAPISurface struct {
	OperationLedgerVersion int `json:"operation_ledger_version"`
	Endpoints              []struct {
		Method    string          `json:"method"`
		Path      string          `json:"path"`
		CoveredBy json.RawMessage `json:"covered_by"`
		Excluded  json.RawMessage `json:"excluded"`
		Operation json.RawMessage `json:"operation"`
	} `json:"endpoints"`
}

func TestBambooHRSurfaceTracksCurrentOfficialInventory(t *testing.T) {
	api := loadBambooHRJSON[bambooHRAPISurface](t, "../../internal/connectors/defs/bamboo-hr/api_surface.json")
	if api.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", api.OperationLedgerVersion)
	}
	if got, want := len(api.Endpoints), 316; got != want {
		t.Fatalf("api_surface endpoints = %d, want %d", got, want)
	}
	seen := map[string]bool{}
	coverage := map[string]int{}
	for _, ep := range api.Endpoints {
		key := strings.ToUpper(ep.Method) + " " + ep.Path
		if seen[key] {
			t.Fatalf("duplicate api_surface endpoint %s", key)
		}
		seen[key] = true
		if len(ep.Excluded) != 0 {
			t.Fatalf("endpoint %s still uses legacy excluded classifier", key)
		}
		classifiers := 0
		if len(ep.CoveredBy) != 0 {
			classifiers++
			var covered map[string]any
			if err := json.Unmarshal(ep.CoveredBy, &covered); err != nil {
				t.Fatalf("covered_by for %s: %v", key, err)
			}
			for name := range covered {
				coverage[name]++
			}
		}
		if len(ep.Operation) != 0 {
			classifiers++
			coverage["operation"]++
			var operation struct {
				Model string `json:"model"`
			}
			if err := json.Unmarshal(ep.Operation, &operation); err != nil {
				t.Fatalf("operation for %s: %v", key, err)
			}
			coverage["operation:"+operation.Model]++
		}
		if classifiers != 1 {
			t.Fatalf("endpoint %s classifiers = %d, want exactly one", key, classifiers)
		}
	}

	for _, key := range []string{
		"GET /api/v1/time-tracking/clock-entries",
		"POST /api/v1/time-tracking/clock-entries",
		"GET /api/v1/custom-reports/legacy-id-map",
		"POST webhook:employee.updated",
	} {
		if !seen[key] {
			t.Fatalf("missing current official endpoint %s", key)
		}
	}
	for _, key := range []string{
		"GET /api/v1/benefitgroups",
		"POST /api/v1/benefit/company_benefit",
	} {
		if seen[key] {
			t.Fatalf("stale non-current endpoint %s is still present", key)
		}
	}
	wantCoverage := map[string]int{
		"stream":                      138,
		"direct_read":                 9,
		"write":                       149,
		"operation":                   20,
		"operation:binary_read":       6,
		"operation:admin_reverse_etl": 7,
		"operation:disallowed":        1,
		"operation:local_workflow":    6,
	}
	for key, want := range wantCoverage {
		if got := coverage[key]; got != want {
			t.Fatalf("coverage[%q] = %d, want %d (coverage=%+v)", key, got, want, coverage)
		}
	}
}

func loadBambooHRJSON[T any](t *testing.T, path string) T {
	t.Helper()
	var out T
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}
