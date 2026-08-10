package cli

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestResolveConnectorCommandEnvironmentOnlyFlags(t *testing.T) {
	const envName = "PM_TEST_GRAPHQL_INPUT"
	const sentinel = `{"sourceAccessToken":"not-for-state"}`
	t.Setenv(envName, sentinel)

	surface := &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path: "graphql mutation start-organization-migration",
		Flags: []connectors.CommandSurfaceFlag{{
			Name: "input", Type: "json", MapsTo: "body.input", Required: true, EnvOnly: true,
		}},
	}}}
	resolved, err := resolveConnectorCommandEnvironmentOnlyFlags(surface, []string{"graphql", "mutation", "start-organization-migration"}, map[string][]string{
		"from-env": {"input=" + envName},
	})
	if err != nil {
		t.Fatalf("resolve environment-only input: %v", err)
	}
	if got := resolved["input"]; len(got) != 1 || got[0] != sentinel {
		t.Fatalf("resolved input = %#v, want environment value", got)
	}
	if _, exists := resolved["from-env"]; exists {
		t.Fatalf("resolved command flags retain --from-env control: %#v", resolved)
	}

	_, err = resolveConnectorCommandEnvironmentOnlyFlags(surface, []string{"graphql", "mutation", "start-organization-migration"}, map[string][]string{
		"input": {"direct-secret-value"},
	})
	if err == nil {
		t.Fatal("direct environment-only input succeeded")
	}
	if strings.Contains(err.Error(), "direct-secret-value") {
		t.Fatalf("direct-value refusal leaked a secret-shaped value: %v", err)
	}
}

func TestResolveReversePlanEnvironmentOnlyFlags(t *testing.T) {
	const envName = "PM_TEST_REVERSE_GRAPHQL_INPUT"
	const sentinel = `{"githubPat":"not-for-state"}`
	t.Setenv(envName, sentinel)

	plan := app.ReversePlan{
		DestinationConnector: "github",
		ConnectorCommand:     "graphql mutation create-migration-source",
		ConnectorCommandPath: []string{"graphql", "mutation", "create-migration-source"},
	}
	resolved, err := resolveReversePlanEnvironmentOnlyFlags(plan, map[string][]string{
		"from-env": {"input=" + envName},
	})
	if err != nil {
		t.Fatalf("resolve reverse-plan environment-only input: %v", err)
	}
	if got := resolved["input"]; len(got) != 1 || got[0] != sentinel {
		t.Fatalf("resolved reverse-plan flags = %#v, want environment value", resolved)
	}
	if _, exists := resolved["from-env"]; exists {
		t.Fatalf("resolved reverse-plan flags retain --from-env control: %#v", resolved)
	}

	_, err = resolveReversePlanEnvironmentOnlyFlags(plan, map[string][]string{
		"input": {"direct-secret-value"},
	})
	if err == nil {
		t.Fatal("direct reverse-plan environment-only input succeeded")
	}
	if strings.Contains(err.Error(), "direct-secret-value") {
		t.Fatalf("direct-value refusal leaked a secret-shaped value: %v", err)
	}
}
