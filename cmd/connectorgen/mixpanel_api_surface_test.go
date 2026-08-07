package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

// mixpanelDocumentedOperations is the operation total re-derived on 2026-08-07
// from Mixpanel's 13 real OpenAPI documents at
// https://docs.mixpanel.com/openapi/{annotations,data-pipelines,experiments,
// export,feature-flags-management,feature-flags,gdpr,identity,ingestion,
// lexicon-schemas,query,service-accounts,warehouse-connectors}.openapi.yaml
// (392,835 combined bytes; mixed OAS 3.0.2/3.0.3/3.1.0). Raw operationId count
// summed across the 13 files is 105, an exact match to the provider-artifact
// ledger's carried-forward total. The -1 delta is one genuine cross-file path
// collision: POST /import is documented twice — "identity-merge" in
// identity.yaml and "import-events" in ingestion.yaml — which the counting
// policy's method+path dedup rule collapses into a single row. See
// .planning/phases/mixpanel-parity-sweep-r1/PLAN.md.
const (
	mixpanelDocumentedOperations = 104
	mixpanelDocumentedGET        = 41
	mixpanelDocumentedPOST       = 44
	mixpanelDocumentedPUT        = 5
	mixpanelDocumentedPATCH      = 4
	mixpanelDocumentedDELETE     = 10
)

func TestMixpanelAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/mixpanel/api_surface.json")
	if err != nil {
		t.Fatalf("read mixpanel api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Excluded  map[string]any `json:"excluded"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				Risk             string `json:"risk"`
				BlockedByDefault bool   `json:"blocked_by_default"`
				Reason           string `json:"reason"`
				SourceURL        string `json:"source_url"`
				Notes            string `json:"notes"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal mixpanel api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion == 0 {
		t.Errorf("operation_ledger_version = %d, want it set (nonzero)", surface.OperationLedgerVersion)
	}

	if got := len(surface.Endpoints); got != mixpanelDocumentedOperations {
		t.Fatalf("endpoints = %d, want %d", got, mixpanelDocumentedOperations)
	}

	byMethod := map[string]int{}
	seen := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded := 0, 0, 0

	for i, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint %q", key)
		}
		seen[key] = true
		byMethod[ep.Method]++

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			if !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Errorf("endpoint %d (%s): operation row must be blocked_by_default with status blocked", i, key)
			}
			if strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Errorf("endpoint %d (%s): operation row is missing reason", i, key)
			}
			if strings.TrimSpace(ep.Operation.SourceURL) == "" {
				t.Errorf("endpoint %d (%s): blocked row has no source citation", i, key)
			}
			if !strings.HasPrefix(ep.Operation.Notes, "named_dependency=") {
				t.Errorf("endpoint %d (%s): blocked row must name its dependency", i, key)
			}
		}
		if len(ep.Excluded) > 0 {
			dispositions++
			legacyExcluded++
		}
		if dispositions == 0 {
			blank = append(blank, key)
		}
		if dispositions > 1 {
			t.Errorf("endpoint %d (%s): carries %d dispositions, want exactly 1", i, key, dispositions)
		}
	}

	if len(blank) > 0 {
		t.Errorf("%d endpoint(s) carry no disposition, want none: %s", len(blank), strings.Join(blank, ", "))
	}
	if legacyExcluded > 0 {
		t.Errorf("%d legacy excluded row(s) remain; operation_ledger_version mode requires operation rows, never excluded", legacyExcluded)
	}
	if covered+blocked != mixpanelDocumentedOperations {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d rows", covered, blocked, covered+blocked, mixpanelDocumentedOperations)
	}

	wantByMethod := map[string]int{
		"GET":    mixpanelDocumentedGET,
		"POST":   mixpanelDocumentedPOST,
		"PUT":    mixpanelDocumentedPUT,
		"PATCH":  mixpanelDocumentedPATCH,
		"DELETE": mixpanelDocumentedDELETE,
	}
	if !reflect.DeepEqual(byMethod, wantByMethod) {
		t.Errorf("byMethod = %+v, want %+v", byMethod, wantByMethod)
	}

	// The path-collapse hazard this sweep flagged up front: POST /import is
	// documented twice (identity-merge in identity.yaml, import-events in
	// ingestion.yaml) but must appear as exactly ONE row, so the dedup can
	// never silently regress back to 105 or fork into two competing rows.
	importCount := 0
	for _, ep := range surface.Endpoints {
		if ep.Method == "POST" && ep.Path == "/import" {
			importCount++
		}
	}
	if importCount != 1 {
		t.Errorf("POST /import appears %d times, want exactly 1 (identity-merge and import-events collapse to one row)", importCount)
	}
}
