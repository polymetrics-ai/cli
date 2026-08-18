package certify

import (
	"encoding/json"
	"os"
	"testing"
)

func githubSourceLockedRESTCount(t *testing.T) int {
	t.Helper()
	raw, err := os.ReadFile("../defs/github/sources/github-operation-source-lock.json")
	if err != nil {
		t.Fatalf("read GitHub source lock: %v", err)
	}
	var lock struct {
		Counts struct {
			REST int `json:"rest"`
		} `json:"counts"`
	}
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("unmarshal GitHub source lock: %v", err)
	}
	if lock.Counts.REST <= 0 {
		t.Fatalf("GitHub source lock REST count = %d, want positive source-derived count", lock.Counts.REST)
	}
	return lock.Counts.REST
}

func githubLegacyGraphQLBindingCount(t *testing.T) int {
	t.Helper()
	raw, err := os.ReadFile("../defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read GitHub api surface: %v", err)
	}
	var surface struct {
		Endpoints []struct {
			Method string `json:"method"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal GitHub api surface: %v", err)
	}
	count := 0
	for _, endpoint := range surface.Endpoints {
		if endpoint.Method == "GRAPHQL" {
			count++
		}
	}
	if count == 0 {
		t.Fatal("GitHub api surface has no legacy fixed GraphQL binding")
	}
	return count
}

// githubSourceLockedGraphQLRootCount is the authoritative generated-operation
// denominator.  The fixed POST /graphql transport is one physical endpoint,
// but it must enumerate every pinned Query and Mutation root rather than
// preserve the former four-operation compatibility count.
func githubSourceLockedGraphQLRootCount(t *testing.T) int {
	t.Helper()
	raw, err := os.ReadFile("../defs/github/sources/github-operation-source-lock.json")
	if err != nil {
		t.Fatalf("read GitHub source lock: %v", err)
	}
	var lock struct {
		Counts struct {
			GraphQLQuery    int `json:"graphql_query"`
			GraphQLMutation int `json:"graphql_mutation"`
		} `json:"counts"`
	}
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("unmarshal GitHub source lock: %v", err)
	}
	total := lock.Counts.GraphQLQuery + lock.Counts.GraphQLMutation
	if total <= 0 {
		t.Fatalf("GitHub source lock GraphQL root count = %d, want positive source-derived count", total)
	}
	return total
}

// githubFixedGraphQLTransportCoverage reports the physical shared transport
// and the exact fixed operation bindings it carries. Legacy GRAPHQL
// pseudo-endpoints are intentionally excluded: they remain compatibility
// metadata, not the executable POST transport. Callers retain the operation
// IDs so source-generated roots can be distinguished from purpose-built fixed
// documents that reuse an existing root.
func githubFixedGraphQLTransportCoverage(t *testing.T) (endpoints int, operations []string) {
	t.Helper()
	raw, err := os.ReadFile("../defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read GitHub api surface: %v", err)
	}
	var surface struct {
		Endpoints []struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			CoveredBy struct {
				Operations []string `json:"operations"`
			} `json:"covered_by"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal GitHub api surface: %v", err)
	}
	seen := make(map[string]struct{})
	for _, endpoint := range surface.Endpoints {
		if endpoint.Method != "POST" || endpoint.Path != "/graphql" || len(endpoint.CoveredBy.Operations) == 0 {
			continue
		}
		endpoints++
		for _, operation := range endpoint.CoveredBy.Operations {
			if _, duplicate := seen[operation]; duplicate {
				t.Fatalf("GitHub fixed GraphQL transport duplicates operation binding %q", operation)
			}
			seen[operation] = struct{}{}
			operations = append(operations, operation)
		}
	}
	return endpoints, operations
}
