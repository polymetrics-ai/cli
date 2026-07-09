package bundleregistry

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
)

func TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides(t *testing.T) {
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs): %v", err)
	}
	if len(bundles) != 547 {
		t.Fatalf("bundle count = %d, want 547", len(bundles))
	}

	registry := New()

	for _, b := range bundles {
		if _, ok := registry.Get(b.Name); !ok {
			t.Fatalf("registry missing bundle connector %q", b.Name)
		}
	}
	for _, legacySlug := range []string{"source-github", "destination-postgres"} {
		if _, ok := registry.Get(legacySlug); ok {
			t.Fatalf("registry contains legacy slug %q; want bare names only", legacySlug)
		}
	}

	akeneo, ok := registry.Get("akeneo")
	if !ok {
		t.Fatal("registry missing akeneo")
	}
	if _, ok := akeneo.(*engine.Connector); !ok {
		t.Fatalf("akeneo registry type = %T, want engine-backed connector", akeneo)
	}
	if engine.HooksFor("github") == nil {
		t.Fatal("hookset side effects were not loaded; github hook is missing")
	}

	postgresConnector, ok := registry.Get("postgres")
	if !ok {
		t.Fatal("registry missing postgres")
	}
	if _, ok := postgresConnector.(nativepostgres.Connector); !ok {
		t.Fatalf("postgres registry type = %T, want Tier-3 native override", postgresConnector)
	}
}

func TestRegistryCatalogEntriesComeFromDefinitions(t *testing.T) {
	registry := New()
	entries := registry.CatalogEntries()
	if len(entries) < 547 {
		t.Fatalf("CatalogEntries() count = %d, want at least 547 bundle/native definitions", len(entries))
	}

	var github connectors.Definition
	foundGithub := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name, "source-") || strings.HasPrefix(entry.Name, "destination-") {
			t.Fatalf("CatalogEntries() contains legacy slug %q", entry.Name)
		}
		if entry.Name == "github" {
			github = entry
			foundGithub = true
		}
	}
	if !foundGithub {
		t.Fatal("CatalogEntries() missing github")
	}
	if !github.Capabilities.Read || len(github.Streams) == 0 {
		t.Fatalf("github definition not sourced from bundle metadata/schemas: %+v", github)
	}
	if len(github.WriteActions) == 0 || !github.Capabilities.Write {
		t.Fatalf("github definition missing bundle write capability/actions: %+v", github)
	}
}

func TestGitHubGuideIncludesCLISurfaceHelp(t *testing.T) {
	registry := New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatalf("github connector not found")
	}

	manual := connectors.RenderConnectorManual(connector)
	for _, want := range []string{
		"COMMAND SURFACE",
		"Usage: pm github <command> <subcommand> [flags]",
		"Core Commands",
		"issue list - List issues",
		"intent=etl availability=implemented stream=issues",
		"issue create - Create an issue",
		"intent=reverse_etl availability=implemented write=create_issue",
		"approval: reverse ETL writes require plan, preview, approval, execute",
		"unsupported local workflow",
		"--json (boolean): Write machine-readable JSON output.",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("GitHub manual missing %q:\n%s", want, manual)
		}
	}
}

func TestMondayGuideIncludesCLISurfaceHelp(t *testing.T) {
	registry := New()
	connector, ok := registry.Get("monday")
	if !ok {
		t.Fatalf("monday connector not found")
	}

	manual := connectors.RenderConnectorManual(connector)
	for _, want := range []string{
		"COMMAND SURFACE",
		"Usage: pm monday <command> <subcommand> [flags]",
		"Core Commands",
		"board list - List board metadata",
		"intent=etl availability=implemented stream=boards",
		"webhook create - Create a webhook",
		"approval: reverse ETL writes require plan, preview, approval, execute plus typed confirmation for external callback targets",
		"--connection (string): Use a saved monday connector credential.",
	} {
		if !strings.Contains(manual, want) {
			t.Fatalf("Monday manual missing %q:\n%s", want, manual)
		}
	}
	if got := strings.Count(manual, "board list - List board metadata"); got != 1 {
		t.Fatalf("Monday manual rendered board list %d times, want exactly once:\n%s", got, manual)
	}
}
