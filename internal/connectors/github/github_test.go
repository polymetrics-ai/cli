package github_test

import (
	"context"
	"testing"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/github"
)

// TestNewContract is the red-first test for the migrated per-system GitHub
// package. It is red until internal/connectors/github exists with New().
func TestNewContract(t *testing.T) {
	c := github.New()
	if c.Name() != "github" {
		t.Fatalf("Name() = %q, want github", c.Name())
	}
	caps := c.Metadata().Capabilities
	if !caps.Read || !caps.Write {
		t.Fatalf("capabilities = %+v, want Read && Write (reverse-ETL)", caps)
	}
}

// TestCatalogStreams asserts the unified connector still advertises core streams.
func TestCatalogStreams(t *testing.T) {
	c := github.New()
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{
		Config: map[string]string{"repository": "octocat/hello-world"},
	})
	if err != nil {
		t.Fatalf("Catalog() error = %v", err)
	}
	want := map[string]bool{"issues": false, "pull_requests": false, "repository": false}
	for _, s := range cat.Streams {
		if _, ok := want[s.Name]; ok {
			want[s.Name] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("catalog missing expected stream %q", name)
		}
	}
}

// TestRegisteredInRegistry asserts the package self-registers so NewRegistry
// resolves both the bare name and the legacy slug to the live connector.
func TestRegisteredInRegistry(t *testing.T) {
	_ = github.New() // ensure package init ran
	r := connectors.NewRegistry()
	for _, name := range []string{"github", "source-github"} {
		c, ok := r.Get(name)
		if !ok {
			t.Fatalf("registry did not resolve %q", name)
		}
		if !c.Metadata().Capabilities.Write {
			t.Errorf("%q capabilities missing Write", name)
		}
	}
}
