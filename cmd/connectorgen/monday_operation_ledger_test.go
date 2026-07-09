package main

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestMondayOperationLedgerInventoryAndSafety(t *testing.T) {
	findings, _ := validateBundleDir(os.DirFS("../../internal/connectors/defs"), "monday")
	if len(findings) != 0 {
		t.Fatalf("monday validate findings = %+v, want none", findings)
	}

	b, err := engine.Load(os.DirFS("../../internal/connectors/defs"), "monday")
	if err != nil {
		t.Fatalf("Load(defs.FS, monday): %v", err)
	}
	if b.Surface == nil {
		t.Fatal("monday api_surface.json is not loaded")
	}
	if b.Surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", b.Surface.OperationLedgerVersion)
	}
	if got := len(b.Surface.Endpoints); got != 367 {
		t.Fatalf("api_surface endpoints = %d, want 367", got)
	}

	methodCounts := map[string]int{}
	coveredStreams := map[string]bool{}
	operationRows := 0
	for i, ep := range b.Surface.Endpoints {
		methodCounts[strings.ToUpper(ep.Method)]++
		if ep.Excluded != nil {
			t.Fatalf("endpoint %d uses legacy excluded in operation ledger mode: %+v", i, ep)
		}
		if ep.CoveredBy != nil {
			if ep.CoveredBy.Write != "" {
				t.Fatalf("endpoint %d exposes executable write coverage: %+v", i, ep.CoveredBy)
			}
			if ep.CoveredBy.Stream != "" {
				coveredStreams[ep.CoveredBy.Stream] = true
			}
			continue
		}
		operationRows++
		if ep.Operation == nil {
			t.Fatalf("endpoint %d has neither covered_by nor operation: %+v", i, ep)
		}
		if ep.Operation.Status != "blocked" || !ep.Operation.BlockedByDefault {
			t.Fatalf("endpoint %d operation is not blocked by default: %+v", i, ep.Operation)
		}
	}
	if methodCounts["GET"] != 87 || methodCounts["POST"] != 280 {
		t.Fatalf("method counts = %+v, want GET=87 POST=280", methodCounts)
	}
	if operationRows != 360 {
		t.Fatalf("blocked operation rows = %d, want 360", operationRows)
	}
	for _, stream := range []string{"boards", "items", "users", "teams", "tags"} {
		if !coveredStreams[stream] {
			t.Fatalf("covered streams = %+v, missing %q", coveredStreams, stream)
		}
	}
	if len(coveredStreams) != 5 {
		t.Fatalf("covered streams = %+v, want exactly implemented Monday streams", coveredStreams)
	}

	kindCounts := map[string]int{}
	for _, op := range b.Operations {
		kindCounts[op.Kind]++
		if op.Kind == "graphql_mutation" && !strings.Contains(op.Approval, "plan, preview, approval, execute") {
			t.Fatalf("mutation operation %q approval = %q, want reverse ETL approval language", op.ID, op.Approval)
		}
	}
	wantKinds := map[string]int{"stream_etl": 53, "graphql_query": 34, "graphql_mutation": 280}
	if len(b.Operations) != 367 || kindCounts["stream_etl"] != wantKinds["stream_etl"] || kindCounts["graphql_query"] != wantKinds["graphql_query"] || kindCounts["graphql_mutation"] != wantKinds["graphql_mutation"] {
		t.Fatalf("operation kinds = %+v (total %d), want %+v total 367", kindCounts, len(b.Operations), wantKinds)
	}
}
