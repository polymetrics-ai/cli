package bundleregistry

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	nativemysql "polymetrics.ai/internal/connectors/native/mysql"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
)

func TestRegistryDirectWriteMetadataUsesEmbeddedOperationSurface(t *testing.T) {
	if _, err := fs.Stat(defs.FS, "github/api_surface.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("fs.Stat(github/api_surface.json) error = %v, want fs.ErrNotExist", err)
	}

	registry := New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("registry missing github")
	}
	provider, ok := connector.(connectors.OperationDirectWriteMetadataProvider)
	if !ok {
		t.Fatalf("github connector = %T, want direct-write metadata provider", connector)
	}
	metadata, err := provider.OperationDirectWriteMetadata("github.repo")
	if err != nil {
		t.Fatalf("OperationDirectWriteMetadata: %v", err)
	}
	if metadata.Operation != "github.repo" {
		t.Fatalf("operation = %q, want github.repo", metadata.Operation)
	}
}

func TestLoadDefinitionsCachesEmbeddedBundleSnapshot(t *testing.T) {
	first, err := loadDefinitions()
	if err != nil {
		t.Fatalf("first loadDefinitions() error = %v", err)
	}
	second, err := loadDefinitions()
	if err != nil {
		t.Fatalf("second loadDefinitions() error = %v", err)
	}
	if len(first) == 0 || len(second) == 0 {
		t.Fatalf("loadDefinitions() returned empty bundles: first=%d second=%d", len(first), len(second))
	}
	if &first[0] != &second[0] {
		t.Fatal("loadDefinitions() returned a separately compiled bundle snapshot")
	}
}

func TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides(t *testing.T) {
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs): %v", err)
	}
	if len(bundles) != 552 {
		t.Fatalf("bundle count = %d, want 552", len(bundles))
	}

	registry := New()
	if err := registry.ValidateIconCoverage(); err != nil {
		t.Fatalf("ValidateIconCoverage(): %v", err)
	}

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
	googleCalendar, ok := registry.Get("google-calendar")
	if !ok {
		t.Fatal("registry missing google-calendar")
	}
	googleCalendarDefinition, ok := connectors.DefinitionOf(googleCalendar)
	if !ok || len(googleCalendarDefinition.WriteActions) != 26 {
		t.Fatalf("google-calendar definition = %+v, want 26 engine-backed write actions", googleCalendarDefinition)
	}
	foundFixtureMode := false
	for _, field := range connectors.ManifestOf(googleCalendar).ConfigFields {
		if field.Name == "mode" {
			foundFixtureMode = true
		}
	}
	if !foundFixtureMode {
		t.Fatal("google-calendar manifest is missing fixture mode configuration")
	}
	googleCalendarSurface, ok := googleCalendar.(connectors.CommandSurfaceProvider)
	if !ok || googleCalendarSurface.CommandSurface() == nil {
		t.Fatalf("google-calendar has no command surface: %T", googleCalendar)
	}
	foundFreeBusy := false
	for _, command := range googleCalendarSurface.CommandSurface().Commands {
		if command.Path == "freebusy query" && command.Availability == "implemented" {
			foundFreeBusy = true
		}
	}
	if !foundFreeBusy {
		t.Fatal("google-calendar command surface is missing implemented freebusy query")
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
	mysqlConnector, ok := registry.Get("mysql")
	if !ok {
		t.Fatal("registry missing mysql")
	}
	if _, ok := mysqlConnector.(nativemysql.Connector); !ok {
		t.Fatalf("mysql registry type = %T, want Tier-3 native override", mysqlConnector)
	}
	mysqlDefinition, ok := connectors.DefinitionOf(mysqlConnector)
	if !ok || mysqlDefinition.Changefeed != nil || mysqlDefinition.Capabilities.CDC {
		t.Fatal("mysql registry connector must keep CDC non-public before a runtime entrypoint exists")
	}
	if _, ok := mysqlConnector.(connectors.ChangefeedExecutor); ok {
		t.Fatal("mysql registry connector must not expose an internal CDC reader as a changefeed executor")
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
