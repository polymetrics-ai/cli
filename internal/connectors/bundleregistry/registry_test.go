package bundleregistry

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestidentity"
	"polymetrics.ai/internal/connectors/manifestindex"
	"polymetrics.ai/internal/connectors/manifeststore"
	bingads "polymetrics.ai/internal/connectors/native/bing-ads"
	nativedynamodb "polymetrics.ai/internal/connectors/native/dynamodb"
	nativefaker "polymetrics.ai/internal/connectors/native/faker"
	nativehubspot "polymetrics.ai/internal/connectors/native/hubspot"
	nativemysql "polymetrics.ai/internal/connectors/native/mysql"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
	tallyprime "polymetrics.ai/internal/connectors/native/tally-prime"
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

func TestConstructionBuildsLazyRegistryWithoutDecodingList(t *testing.T) {
	var loads atomic.Int32
	construction := testConstruction(t, []manifestindex.Entry{
		{Connector: "github", Generation: "g", Digest: "github-digest", Executor: "api_engine.v1", Metadata: connectors.Metadata{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}, Bytes: 1},
		{Connector: "gitlab", Generation: "g", Digest: "gitlab-digest", Executor: "api_engine.v1", Metadata: connectors.Metadata{Name: "gitlab", DisplayName: "GitLab", IntegrationType: "api"}, Bytes: 1},
	}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
		loads.Add(1)
		loaded := loadedBundleForTest(entry)
		loaded.Bundle.Metadata = engine.Metadata{
			Name:            entry.Connector,
			DisplayName:     entry.Connector,
			IntegrationType: "api",
		}
		return loaded, nil
	})

	registry, err := construction.BuildRegistry()
	if err != nil {
		t.Fatalf("BuildRegistry(): %v", err)
	}
	if got := loads.Load(); got != 0 {
		t.Fatalf("BuildRegistry() decoded %d bundles while constructing a metadata registry, want 0", got)
	}
	if got := registry.List(); len(got) != 6 {
		t.Fatalf("List() returned %d metadata rows, want 6 builtins plus indexed entries", len(got))
	}
	listed, ok := func() (connectors.Metadata, bool) {
		for _, metadata := range registry.List() {
			if metadata.Name == "github" {
				return metadata, true
			}
		}
		return connectors.Metadata{}, false
	}()
	if !ok || listed.DisplayName != "GitHub" || listed.IntegrationType != "api" {
		t.Fatalf("List() github metadata = %#v, want generated metadata without a bundle decode", listed)
	}
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("Get(github) did not resolve an indexed connector")
	}
	if connector.Name() != "github" {
		t.Fatalf("Get(github).Name() = %q, want github", connector.Name())
	}
	if got := loads.Load(); got != 1 {
		t.Fatalf("Get(github) decoded %d bundles, want 1 selected bundle", got)
	}
}

func TestConstructionRejectsUnknownExtensionBeforeLoading(t *testing.T) {
	var loads atomic.Int32
	construction := testConstruction(t, []manifestindex.Entry{{
		Connector:  "github",
		Generation: "g",
		Digest:     "github-digest",
		Executor:   "api_engine.v1",
		Extension:  "hook/unknown.v1",
		Metadata:   connectors.Metadata{Name: "github", DisplayName: "GitHub", IntegrationType: "api"},
		Bytes:      1,
	}}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
		loads.Add(1)
		return loadedBundleForTest(entry), nil
	})

	_, err := construction.BuildRegistry()
	if err == nil || !strings.Contains(err.Error(), "hook/unknown.v1") {
		t.Fatalf("BuildRegistry() error = %v, want unknown extension refusal", err)
	}
	if got := loads.Load(); got != 0 {
		t.Fatalf("BuildRegistry() decoded %d bundles before rejecting an unknown extension, want 0", got)
	}
}

func TestConstructionRejectsBuiltinIdentityBeforeLoading(t *testing.T) {
	var loads atomic.Int32
	construction := testConstruction(t, []manifestindex.Entry{{
		Connector:  "sample",
		Generation: "g",
		Digest:     "sample-digest",
		Executor:   "api_engine.v1",
		Metadata:   connectors.Metadata{Name: "sample", DisplayName: "Sample", IntegrationType: "api"},
		Bytes:      1,
	}}, func(_ context.Context, entry manifestindex.Entry) (manifeststore.LoadedBundle, error) {
		loads.Add(1)
		return loadedBundleForTest(entry), nil
	})

	_, err := construction.BuildRegistry()
	if err == nil || !strings.Contains(err.Error(), "reserved builtin") {
		t.Fatalf("BuildRegistry() error = %v, want reserved builtin refusal", err)
	}
	if got := loads.Load(); got != 0 {
		t.Fatalf("BuildRegistry() decoded %d bundles before reserved builtin refusal, want 0", got)
	}
}

func TestConstructionRejectsSameNameLoadedIdentityBeforeFactory(t *testing.T) {
	indexed := manifestindex.Entry{Connector: "github", Generation: "g1", Digest: "sha256:index", Executor: "api_engine.v1", Metadata: connectors.Metadata{Name: "github", IntegrationType: "api"}, Bytes: 1}
	loaded := indexed
	loaded.Digest = "sha256:loaded"
	construction := testConstruction(t, []manifestindex.Entry{indexed}, func(_ context.Context, _ manifestindex.Entry) (manifeststore.LoadedBundle, error) {
		return loadedBundleForTest(loaded), nil
	})
	var factoryCalls atomic.Int32
	factories, err := NewExecutorFactories(ExecutorFactory{
		ID: "api_engine.v1",
		Construct: func(bundle engine.Bundle) (connectors.Connector, error) {
			factoryCalls.Add(1)
			return engine.New(bundle, nil), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	construction.factories = factories
	registry, err := construction.BuildRegistry()
	if err != nil {
		t.Fatal(err)
	}
	connector, err := registry.Resolve(context.Background(), indexed.Connector)
	if !errors.Is(err, manifeststore.ErrBundleIdentityMismatch) {
		t.Fatalf("Resolve() = %T %v, want ErrBundleIdentityMismatch", err, err)
	}
	if connector != nil {
		t.Fatalf("Resolve() returned %T after loaded identity mismatch", connector)
	}
	if got := factoryCalls.Load(); got != 0 {
		t.Fatalf("factory calls = %d, want 0 after loaded identity mismatch", got)
	}
}

func TestConstructionHoldsSelectedGenerationUntilClose(t *testing.T) {
	entry := manifestindex.Entry{
		Connector:  "github",
		Generation: "generation-1",
		Digest:     "sha256:github",
		Executor:   "api_engine.v1",
		Metadata:   connectors.Metadata{Name: "github", DisplayName: "GitHub", IntegrationType: "api"},
		Bytes:      1,
	}
	construction := testConstruction(t, []manifestindex.Entry{entry}, func(_ context.Context, selected manifestindex.Entry) (manifeststore.LoadedBundle, error) {
		loaded := loadedBundleForTest(selected)
		loaded.Bundle.Metadata = engine.Metadata{Name: selected.Connector, IntegrationType: "api"}
		return loaded, nil
	})
	registry, err := construction.BuildRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := registry.Get(entry.Connector); !ok {
		t.Fatal("selected connector was not constructed")
	}
	if !construction.store.GenerationHeld(entry) {
		t.Fatal("selected generation was not held by the constructed registry")
	}
	construction.Close()
	if construction.store.GenerationHeld(entry) {
		t.Fatal("selected generation remained held after construction close")
	}
}

func testConstruction(t *testing.T, entries []manifestindex.Entry, loader manifeststore.BundleLoader) *Construction {
	t.Helper()
	index, err := manifestindex.New(entries, len(entries))
	if err != nil {
		t.Fatal(err)
	}
	store, err := manifeststore.NewBundleStore(index, manifeststore.Limits{Entries: len(entries), Bytes: len(entries)}, loader)
	if err != nil {
		t.Fatal(err)
	}
	factories, err := NewExecutorFactories(ExecutorFactory{
		ID: "api_engine.v1",
		Construct: func(bundle engine.Bundle) (connectors.Connector, error) {
			return engine.New(bundle, nil), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return &Construction{index: index, factories: factories, store: store}
}

func loadedBundleForTest(entry manifestindex.Entry) manifeststore.LoadedBundle {
	identity := manifestidentity.Identity{Connector: entry.Connector, Generation: entry.Generation, Digest: entry.Digest, Bytes: entry.Bytes}
	bundle := &engine.Bundle{Name: entry.Connector, Identity: identity}
	return manifeststore.LoadedBundle{Bundle: bundle, Identity: identity}
}

func TestNewLoadsDeclarativeBundlesWithProtectedNativeDatabases(t *testing.T) {
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs): %v", err)
	}
	if len(bundles) != 553 {
		t.Fatalf("bundle count = %d, want 553", len(bundles))
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
	if _, ok := googleCalendar.(*engine.Connector); !ok {
		t.Fatalf("google-calendar registry type = %T, want engine-backed API connector", googleCalendar)
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
	construction, err := NewConstruction()
	if err != nil {
		t.Fatalf("NewConstruction(): %v", err)
	}
	hooks, err := construction.hooks.construct("hook/github.v1", "github")
	if err != nil || hooks == nil {
		t.Fatalf("explicit production hook factory for github = %T, %v", hooks, err)
	}

	postgresConnector, ok := registry.Get("postgres")
	if !ok {
		t.Fatal("registry missing postgres")
	}
	if _, ok := postgresConnector.(nativepostgres.Connector); !ok {
		t.Fatalf("postgres registry type = %T, want protected native database connector", postgresConnector)
	}
	mysqlConnector, ok := registry.Get("mysql")
	if !ok {
		t.Fatal("registry missing mysql")
	}
	if _, ok := mysqlConnector.(nativemysql.Connector); !ok {
		t.Fatalf("mysql registry type = %T, want protected native database connector", mysqlConnector)
	}
	mysqlDefinition, ok := connectors.DefinitionOf(mysqlConnector)
	if !ok || mysqlDefinition.Changefeed != nil || mysqlDefinition.Capabilities.CDC {
		t.Fatal("mysql registry connector must keep CDC non-public before a runtime entrypoint exists")
	}
	if _, ok := mysqlConnector.(connectors.ChangefeedExecutor); ok {
		t.Fatal("mysql registry connector must not expose an internal CDC reader as a changefeed executor")
	}
}

func TestProtectedNativeDatabasesRemainRegistered(t *testing.T) {
	registry := New()

	dynamoDB, ok := registry.Get("dynamodb")
	if !ok {
		t.Fatal("registry missing dynamodb")
	}
	if _, ok := dynamoDB.(nativedynamodb.Connector); !ok {
		t.Fatalf("dynamodb registry type = %T, want protected native database connector", dynamoDB)
	}

	mysql, ok := registry.Get("mysql")
	if !ok {
		t.Fatal("registry missing mysql")
	}
	if _, ok := mysql.(nativemysql.Connector); !ok {
		t.Fatalf("mysql registry type = %T, want protected native database connector", mysql)
	}

	postgres, ok := registry.Get("postgres")
	if !ok {
		t.Fatal("registry missing postgres")
	}
	if _, ok := postgres.(nativepostgres.Connector); !ok {
		t.Fatalf("postgres registry type = %T, want protected native database connector", postgres)
	}
}

func TestProtectedCompatibilityConnectorsRemainNative(t *testing.T) {
	registry := New()
	for name, want := range map[string]any{
		"bing-ads":    bingads.Connector{},
		"faker":       nativefaker.Connector{},
		"hubspot":     &nativehubspot.Connector{},
		"tally-prime": tallyprime.Connector{},
	} {
		connector, ok := registry.Get(name)
		if !ok {
			t.Fatalf("registry missing protected compatibility connector %q", name)
		}
		switch want.(type) {
		case bingads.Connector:
			if _, ok := connector.(bingads.Connector); !ok {
				t.Fatalf("%s registry type = %T, want historical native compatibility connector", name, connector)
			}
		case nativefaker.Connector:
			if _, ok := connector.(nativefaker.Connector); !ok {
				t.Fatalf("%s registry type = %T, want historical native compatibility connector", name, connector)
			}
		case *nativehubspot.Connector:
			if _, ok := connector.(*nativehubspot.Connector); !ok {
				t.Fatalf("%s registry type = %T, want historical native compatibility connector", name, connector)
			}
		case tallyprime.Connector:
			if _, ok := connector.(tallyprime.Connector); !ok {
				t.Fatalf("%s registry type = %T, want historical native compatibility connector", name, connector)
			}
		}
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
