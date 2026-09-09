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
	SchemaVersion int                           `json:"schema_version"`
	Purpose       string                        `json:"purpose"`
	Connectors    map[string]s1aMatrixConnector `json:"connectors"`
}

type s1aMatrixConnector struct {
	Auth    []engine.AuthSpec      `json:"auth"`
	Streams []s1aStreamContractRow `json:"streams"`
}

type s1aStreamContractRow struct {
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
	if err := decodeStrictJSON(raw, &matrix); err != nil {
		t.Fatalf("decode source-parent matrix: %v", err)
	}
	if matrix.SchemaVersion != 1 || matrix.Purpose != "source-parent differential contract matrix for S1A execution facts" || len(matrix.Connectors) != len(migratedAPIDocs) || len(matrix.Connectors["ashby"].Streams) != 71 || len(matrix.Connectors["prestashop"].Streams) != 5 {
		t.Fatalf("matrix schema/purpose/cardinality = %d/%q/%d/%d/%d", matrix.SchemaVersion, matrix.Purpose, len(matrix.Connectors), len(matrix.Connectors["ashby"].Streams), len(matrix.Connectors["prestashop"].Streams))
	}
	for _, name := range migratedAPIDocs {
		expected, ok := matrix.Connectors[name]
		if !ok {
			t.Fatalf("matrix omits S1A connector %q", name)
		}
		lockRaw, err := os.ReadFile(filepath.Join(root, "internal", "connectors", "defs", name, "source.lock.json"))
		if err != nil {
			t.Fatalf("read source lock %s: %v", name, err)
		}
		lock, err := decodeVNextSourceLock(lockRaw)
		if err != nil {
			t.Fatalf("decode source lock %s: %v", name, err)
		}
		var sourceHTTP engine.HTTPBase
		if err := json.Unmarshal(lock.HTTP, &sourceHTTP); err != nil {
			t.Fatalf("decode source HTTP %s: %v", name, err)
		}
		if !reflect.DeepEqual(sourceHTTP.Auth, expected.Auth) {
			t.Fatalf("%s source auth diverges from source-parent matrix", name)
		}
		bundle, err := engine.Load(defs.FS, name)
		if err != nil {
			t.Fatalf("load %s: %v", name, err)
		}
		if !reflect.DeepEqual(bundle.HTTP.Auth, expected.Auth) {
			t.Fatalf("%s rendered auth diverges from source-parent matrix", name)
		}
		sourceByOperation := make(map[string]engine.StreamSpec, len(lock.Operations))
		for _, operation := range lock.Operations {
			if len(operation.Stream) == 0 {
				continue
			}
			var stream engine.StreamSpec
			if err := json.Unmarshal(operation.Stream, &stream); err != nil {
				t.Fatalf("%s decode stream %q: %v", name, operation.ID, err)
			}
			if _, duplicate := sourceByOperation[operation.ID]; duplicate {
				t.Fatalf("%s source lock duplicates stream operation %q", name, operation.ID)
			}
			sourceByOperation[operation.ID] = stream
		}
		actualByName := make(map[string]engine.StreamSpec, len(bundle.Streams))
		for _, stream := range bundle.Streams {
			if _, duplicate := actualByName[stream.Name]; duplicate {
				t.Fatalf("%s rendered bundle duplicates stream %q", name, stream.Name)
			}
			actualByName[stream.Name] = stream
		}
		if len(expected.Streams) != len(sourceByOperation) || len(expected.Streams) != len(actualByName) {
			t.Fatalf("%s stream cardinality matrix/source/rendered = %d/%d/%d", name, len(expected.Streams), len(sourceByOperation), len(actualByName))
		}
		seenOperations := make(map[string]struct{}, len(expected.Streams))
		seenStreams := make(map[string]struct{}, len(expected.Streams))
		for _, fact := range expected.Streams {
			if _, duplicate := seenOperations[fact.OperationID]; duplicate {
				t.Fatalf("%s matrix duplicates operation %q", name, fact.OperationID)
			}
			if _, duplicate := seenStreams[fact.StreamName]; duplicate {
				t.Fatalf("%s matrix duplicates stream %q", name, fact.StreamName)
			}
			seenOperations[fact.OperationID] = struct{}{}
			seenStreams[fact.StreamName] = struct{}{}
			source, ok := sourceByOperation[fact.OperationID]
			if !ok {
				t.Fatalf("%s matrix operation %q is absent from source lock", name, fact.OperationID)
			}
			actual, ok := actualByName[fact.StreamName]
			if !ok {
				t.Fatalf("%s matrix stream %q is absent from rendered bundle", name, fact.StreamName)
			}
			if !sameS1AStreamContract(fact, source) {
				t.Fatalf("%s source operation %q diverges from source-parent matrix", name, fact.OperationID)
			}
			if !sameS1AStreamContract(fact, actual) {
				t.Fatalf("%s rendered stream %q diverges from source-parent matrix", name, fact.StreamName)
			}
		}
	}
	for name := range matrix.Connectors {
		found := false
		for _, expected := range migratedAPIDocs {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("matrix contains non-S1A connector %q", name)
		}
	}
}

func sameS1AStreamContract(want s1aStreamContractRow, got engine.StreamSpec) bool {
	if want.StreamName != got.Name || want.Path != got.Path || want.Method != got.Method || want.BodyType != got.BodyType || !equivalentBody(got.Body, want.Body) || len(want.RequiredBodyFields) != len(got.RequiredBodyFields) || !reflect.DeepEqual(want.Records, got.Records) || !reflect.DeepEqual(want.Incremental, got.Incremental) || !reflect.DeepEqual(want.ResponseError, got.ResponseError) {
		return false
	}
	for index, field := range want.RequiredBodyFields {
		if got.RequiredBodyFields[index] != field {
			return false
		}
	}
	return true
}

func equivalentBody(actual, expected map[string]any) bool {
	return (len(actual) == 0 && len(expected) == 0) || reflect.DeepEqual(actual, expected)
}
