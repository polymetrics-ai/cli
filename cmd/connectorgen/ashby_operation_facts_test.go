package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAshbyOperationFactsCoverEveryStream(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "internal", "connectors", "defs", "ashby", "source.lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var lock struct {
		ProviderEvidence struct {
			OperationFacts []struct {
				OperationID string         `json:"operation_id"`
				Stream      string         `json:"stream"`
				Path        string         `json:"path"`
				Body        map[string]any `json:"body"`
			} `json:"operation_facts"`
		} `json:"provider_evidence"`
		Operations []struct {
			ID     string `json:"id"`
			Stream *struct {
				Name string         `json:"name"`
				Path string         `json:"path"`
				Body map[string]any `json:"body"`
			} `json:"stream"`
		} `json:"operations"`
	}
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatal(err)
	}
	if len(lock.ProviderEvidence.OperationFacts) != 71 {
		t.Fatalf("Ashby source facts = %d, want 71", len(lock.ProviderEvidence.OperationFacts))
	}
	facts := map[string]struct {
		stream, path string
		body         map[string]any
	}{}
	for _, fact := range lock.ProviderEvidence.OperationFacts {
		facts[fact.OperationID] = struct {
			stream, path string
			body         map[string]any
		}{fact.Stream, fact.Path, fact.Body}
	}
	for _, operation := range lock.Operations {
		if operation.Stream == nil {
			continue
		}
		fact, ok := facts[operation.ID]
		if !ok || fact.stream != operation.Stream.Name || fact.path != operation.Stream.Path || len(fact.body) != len(operation.Stream.Body) {
			t.Fatalf("Ashby operation %q diverges from checked-in source fact", operation.ID)
		}
	}
}
