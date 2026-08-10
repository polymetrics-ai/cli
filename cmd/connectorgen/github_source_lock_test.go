package main

import (
	"encoding/json"
	"os"
	"testing"
)

type githubSourceLock struct {
	Counts struct {
		REST            int `json:"rest"`
		GraphQLQuery    int `json:"graphql_query"`
		GraphQLMutation int `json:"graphql_mutation"`
		Total           int `json:"total"`
	} `json:"counts"`
	REST struct {
		Operations []struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		} `json:"operations"`
	} `json:"rest"`
	GraphQL struct {
		QueryFields []struct {
			Name string `json:"name"`
		} `json:"query_fields"`
		MutationFields []struct {
			Name string `json:"name"`
		} `json:"mutation_fields"`
	} `json:"graphql"`
}

func loadGitHubSourceLock(t *testing.T) githubSourceLock {
	t.Helper()
	raw, err := os.ReadFile("../../internal/connectors/defs/github/sources/github-operation-source-lock.json")
	if err != nil {
		t.Fatalf("read GitHub source lock: %v", err)
	}
	var lock githubSourceLock
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("unmarshal GitHub source lock: %v", err)
	}
	if lock.Counts.REST != len(lock.REST.Operations) {
		t.Fatalf("GitHub source lock REST count = %d, operations = %d", lock.Counts.REST, len(lock.REST.Operations))
	}
	if lock.Counts.GraphQLQuery != len(lock.GraphQL.QueryFields) || lock.Counts.GraphQLMutation != len(lock.GraphQL.MutationFields) {
		t.Fatalf("GitHub source lock GraphQL counts = query %d mutation %d, fields = query %d mutation %d", lock.Counts.GraphQLQuery, lock.Counts.GraphQLMutation, len(lock.GraphQL.QueryFields), len(lock.GraphQL.MutationFields))
	}
	if lock.Counts.Total != lock.Counts.REST+lock.Counts.GraphQLQuery+lock.Counts.GraphQLMutation {
		t.Fatalf("GitHub source lock total = %d, want REST + GraphQL roots", lock.Counts.Total)
	}
	for _, field := range lock.GraphQL.MutationFields {
		if field.Name == "createEnterpriseOrganization" {
			return lock
		}
	}
	t.Fatal("GitHub source lock is missing createEnterpriseOrganization mutation canary")
	return githubSourceLock{}
}

func githubRESTMethodSplit(lock githubSourceLock) map[string]int {
	counts := map[string]int{}
	for _, operation := range lock.REST.Operations {
		counts[operation.Method]++
	}
	return counts
}

func githubRESTOperationKeys(lock githubSourceLock) map[string]bool {
	keys := make(map[string]bool, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		keys[operation.Method+" "+operation.Path] = true
	}
	return keys
}

// githubGeneratedGraphQLTransport is the one physical GraphQL endpoint that
// maps every fixed generated root operation. It is not a REST OpenAPI row, so
// REST source-count assertions must keep it in the GraphQL column even though
// its wire method is POST.
func githubGeneratedGraphQLTransport(method, path string, coveredBy map[string]any) bool {
	if method != "POST" || path != "/graphql" {
		return false
	}
	operations, ok := coveredBy["operations"].([]any)
	return ok && len(operations) > 0
}
