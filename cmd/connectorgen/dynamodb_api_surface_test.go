package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// dynamodbDocumentedOperations is the operation total re-derived on 2026-08-07
// from AWS's own service model for DynamoDB (botocore
// data/dynamodb/2012-08-10/service-2.json, serviceId DynamoDB, targetPrefix
// DynamoDB_20120810), byte-identical at stable tag 1.43.66 and set-equal to the
// operations linked from API_Operations.html.
//
// The provider-artifact sweep carried forward 57 from an older audit. That was
// correct when taken: botocore 1.35.0, 1.38.0 and 1.40.0 each declare 57 and no
// SearchVectors, while current declares 58 with it. The delta is that one
// vector-similarity-search operation, and it classifies as a read.
//
// DynamoDB Streams (4 operations) and DAX (21 operations) are separate AWS
// services with their own targetPrefix and their own API_streams_*/API_dax_*
// doc slugs. They are deliberately excluded. Scraping every API_*.html link on
// the shared index page yields 84 rather than 58; this test exists partly to
// make that miscount impossible to land.
const (
	dynamodbDocumentedOperations = 58
	dynamodbDocumentedReads      = 27
	dynamodbDocumentedWrites     = 31
)

func TestDynamoDBAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/dynamodb/api_surface.json")
	if err != nil {
		t.Fatalf("read dynamodb api_surface.json: %v", err)
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
				Reason           string `json:"reason"`
				SourceURL        string `json:"source_url"`
				Notes            string `json:"notes"`
				BlockedByDefault bool   `json:"blocked_by_default"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal dynamodb api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion == 0 {
		t.Errorf("operation_ledger_version is unset; the v2 provenance ledger is required")
	}

	if got := len(surface.Endpoints); got != dynamodbDocumentedOperations {
		t.Errorf("api_surface declares %d endpoints, want %d documented operations", got, dynamodbDocumentedOperations)
	}

	// Every operation is POST / selected by an X-Amz-Target header, so the row
	// key must carry the target. Rows keyed on the bare path would collapse to
	// one endpoint and silently hide 57 operations.
	seen := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded := 0, 0, 0
	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint row %q", key)
		}
		seen[key] = true

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			if !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Errorf("%s: operation row must be blocked_by_default with status blocked", key)
			}
			if strings.TrimSpace(ep.Operation.SourceURL) == "" {
				t.Errorf("%s: blocked row has no source citation", key)
			}
			if !strings.HasPrefix(ep.Operation.Notes, "named_dependency=") {
				t.Errorf("%s: blocked row must name its dependency", key)
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
			t.Errorf("%s: carries %d dispositions, want exactly 1", key, dispositions)
		}
	}

	if len(blank) > 0 {
		t.Errorf("%d endpoint(s) carry no disposition, want none: %s", len(blank), strings.Join(blank, ", "))
	}
	if legacyExcluded > 0 {
		t.Errorf("%d legacy excluded row(s) remain; operation_ledger_version mode requires operation rows", legacyExcluded)
	}
	if covered+blocked != dynamodbDocumentedOperations {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d", covered, blocked, covered+blocked, dynamodbDocumentedOperations)
	}
	if dynamodbDocumentedReads+dynamodbDocumentedWrites != dynamodbDocumentedOperations {
		t.Fatalf("read/write split %d+%d does not sum to %d", dynamodbDocumentedReads, dynamodbDocumentedWrites, dynamodbDocumentedOperations)
	}
}
