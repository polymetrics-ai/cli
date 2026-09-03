package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

type s1aContractMatrix struct {
	Connectors map[string]struct {
		Auth    []engine.AuthSpec `json:"auth"`
		Streams []struct {
			OperationID        string                    `json:"operation_id"`
			StreamName         string                    `json:"stream_name"`
			Path               string                    `json:"path"`
			Method             string                    `json:"method"`
			Body               map[string]any            `json:"body"`
			BodyType           string                    `json:"body_type"`
			RequiredBodyFields []string                  `json:"required_body_fields"`
			Records            engine.RecordsSpec        `json:"records"`
			Incremental        *engine.IncrementalSpec   `json:"incremental"`
			ResponseError      *engine.ResponseErrorSpec `json:"response_error"`
		} `json:"streams"`
	} `json:"connectors"`
}

func TestS1ASourceParentContractMatrixMatchesExecution(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("repoRoot: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "docs", "connector-canon", "foundations", "s1a-source-parent-contract-matrix.json"))
	if err != nil {
		t.Fatal(err)
	}
	var matrix s1aContractMatrix
	if err := json.Unmarshal(raw, &matrix); err != nil {
		t.Fatal(err)
	}
	if len(matrix.Connectors) != 28 || len(matrix.Connectors["ashby"].Streams) != 71 || len(matrix.Connectors["prestashop"].Streams) != 5 {
		t.Fatalf("matrix cardinality connectors/ashby/prestashop = %d/%d/%d", len(matrix.Connectors), len(matrix.Connectors["ashby"].Streams), len(matrix.Connectors["prestashop"].Streams))
	}
	for name, expected := range matrix.Connectors {
		bundle, err := engine.Load(defs.FS, name)
		if err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
		if !reflect.DeepEqual(bundle.HTTP.Auth, expected.Auth) {
			t.Fatalf("%s rendered auth diverges from source-parent matrix", name)
		}
		actual := map[string]engine.StreamSpec{}
		for _, stream := range bundle.Streams {
			actual[stream.Name] = stream
		}
		for _, fact := range expected.Streams {
			stream, ok := actual[fact.StreamName]
			if !ok || stream.Path != fact.Path || stream.Method != fact.Method || stream.Records.Path != fact.Records.Path {
				t.Fatalf("%s operation %s diverges from source-parent matrix", name, fact.OperationID)
			}
		}
	}
}

func equivalentBody(actual, expected map[string]any) bool {
	return (len(actual) == 0 && len(expected) == 0) || reflect.DeepEqual(actual, expected)
}
