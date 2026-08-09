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
