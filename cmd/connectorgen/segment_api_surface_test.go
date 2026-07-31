package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSegmentOfficialAPISurfaceParityCounts(t *testing.T) {
	api := loadSegmentJSON[struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			CoveredBy *struct {
				Stream     string `json:"stream"`
				Write      string `json:"write"`
				DirectRead string `json:"direct_read"`
			} `json:"covered_by"`
			Operation *struct {
				Model string `json:"model"`
			} `json:"operation"`
		} `json:"endpoints"`
	}](t, "../../internal/connectors/defs/segment/api_surface.json")
	streams := loadSegmentJSON[struct {
		Streams []struct {
			Name string `json:"name"`
		} `json:"streams"`
	}](t, "../../internal/connectors/defs/segment/streams.json")
	writes := loadSegmentJSON[struct {
		Actions []struct {
			Name string `json:"name"`
		} `json:"actions"`
	}](t, "../../internal/connectors/defs/segment/writes.json")
	operations := loadSegmentJSON[struct {
		Operations []struct {
			ID   string          `json:"id"`
			Kind string          `json:"kind"`
			REST json.RawMessage `json:"rest"`
		} `json:"operations"`
	}](t, "../../internal/connectors/defs/segment/operations.json")
	cli := loadSegmentJSON[struct {
		Commands []struct {
			Path         string `json:"path"`
			Intent       string `json:"intent"`
			Availability string `json:"availability"`
			Operation    string `json:"operation"`
			Write        string `json:"write"`
			Stream       string `json:"stream"`
		} `json:"commands"`
	}](t, "../../internal/connectors/defs/segment/cli_surface.json")

	if api.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", api.OperationLedgerVersion)
	}
	if got, want := len(api.Endpoints), 197; got != want {
		t.Fatalf("api endpoints = %d, want %d", got, want)
	}
	if got, want := len(operations.Operations), 197; got != want {
		t.Fatalf("operations = %d, want %d", got, want)
	}
	if got, want := len(streams.Streams), 80; got != want {
		t.Fatalf("streams = %d, want %d", got, want)
	}
	if got, want := len(writes.Actions), 96; got != want {
		t.Fatalf("write actions = %d, want %d", got, want)
	}
	if got, want := len(cli.Commands), 28; got != want {
		t.Fatalf("cli commands = %d, want %d", got, want)
	}

	coverage := map[string]int{}
	methodCounts := map[string]int{}
	for _, ep := range api.Endpoints {
		methodCounts[ep.Method]++
		if ep.Path == "/workspaces" {
			t.Fatalf("stale legacy /workspaces endpoint must not appear in official Segment API surface")
		}
		if ep.CoveredBy != nil {
			switch {
			case ep.CoveredBy.Stream != "":
				coverage["stream"]++
			case ep.CoveredBy.Write != "":
				coverage["write"]++
			case ep.CoveredBy.DirectRead != "":
				coverage["direct_read"]++
			default:
				t.Fatalf("endpoint %s %s has empty covered_by", ep.Method, ep.Path)
			}
			continue
		}
		if ep.Operation != nil {
			coverage["operation:"+ep.Operation.Model]++
			continue
		}
		t.Fatalf("endpoint %s %s has no classifier", ep.Method, ep.Path)
	}
	wantMethods := map[string]int{"GET": 97, "POST": 43, "DELETE": 27, "PUT": 7, "PATCH": 23}
	for method, want := range wantMethods {
		if got := methodCounts[method]; got != want {
			t.Fatalf("method count %s = %d, want %d", method, got, want)
		}
	}
	wantCoverage := map[string]int{
		"stream":                80,
		"write":                 96,
		"direct_read":           19,
		"operation:binary_read": 1,
		"operation:disallowed":  1,
	}
	for key, want := range wantCoverage {
		if got := coverage[key]; got != want {
			t.Fatalf("coverage[%s] = %d, want %d (all coverage: %+v)", key, got, want, coverage)
		}
	}

	kindCounts := map[string]int{}
	for _, op := range operations.Operations {
		kindCounts[op.Kind]++
	}
	wantKinds := map[string]int{"stream_etl": 80, "rest_write": 96, "rest_read": 21}
	for kind, want := range wantKinds {
		if got := kindCounts[kind]; got != want {
			t.Fatalf("operation kind %s = %d, want %d (all kinds: %+v)", kind, got, want, kindCounts)
		}
	}
}

func loadSegmentJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return out
}
