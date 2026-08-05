package bundleregistry

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

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
	if len(bundles) != 550 {
		t.Fatalf("bundle count = %d, want 550", len(bundles))
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

// TestNewOmitsUnloadableBundlesInsteadOfPanicking proves the kill switch
// disables one connector rather than the whole CLI: a bundle whose metadata
// declares mechanism.disabled_reason fails engine.Load, and New must drop
// just that connector from the catalog while every other bundle still
// registers — never panic, which would abort every pm invocation.
func TestNewOmitsUnloadableBundlesInsteadOfPanicking(t *testing.T) {
	fixture := fstest.MapFS{}
	copyBundleForTest(t, fixture, "github")
	copyBundleForTest(t, fixture, "akeneo")
	disableBundleForTest(t, fixture, "akeneo", "upstream routes rotated; pending re-verification")
	useDefinitionsFSForTest(t, fixture)
	warnings := captureWarningsForTest(t)

	registry := newWithoutPanicking(t)

	if _, ok := registry.Get("akeneo"); ok {
		t.Fatal("registry contains akeneo; a bundle with mechanism.disabled_reason must be omitted from the catalog")
	}
	if _, ok := registry.Get("github"); !ok {
		t.Fatal("registry missing github; bundles that loaded successfully must still register")
	}

	got := warnings.String()
	for _, want := range []string{"akeneo", "upstream routes rotated"} {
		if !strings.Contains(got, want) {
			t.Fatalf("warning output %q does not name %q; an omitted connector must be reported", got, want)
		}
	}
}

func captureWarningsForTest(t *testing.T) *strings.Builder {
	t.Helper()
	var sb strings.Builder
	prev := warnf
	warnf = func(format string, a ...any) { fmt.Fprintf(&sb, format, a...) }
	t.Cleanup(func() { warnf = prev })
	return &sb
}

func newWithoutPanicking(t *testing.T) *connectors.Registry {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("New() panicked on an unloadable bundle: %v", r)
		}
	}()
	return New()
}

func useDefinitionsFSForTest(t *testing.T, fsys fs.FS) {
	t.Helper()
	prev := definitionsFS
	definitionsFS = fsys
	t.Cleanup(func() { definitionsFS = prev })
}

func copyBundleForTest(t *testing.T, dst fstest.MapFS, name string) {
	t.Helper()
	err := fs.WalkDir(defs.FS, name, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		raw, err := fs.ReadFile(defs.FS, path)
		if err != nil {
			return err
		}
		dst[path] = &fstest.MapFile{Data: raw}
		return nil
	})
	if err != nil {
		t.Fatalf("copy bundle %s: %v", name, err)
	}
}

func disableBundleForTest(t *testing.T, fsys fstest.MapFS, name, reason string) {
	t.Helper()
	path := name + "/metadata.json"
	file, ok := fsys[path]
	if !ok {
		t.Fatalf("fixture is missing %s", path)
	}
	var meta map[string]any
	if err := json.Unmarshal(file.Data, &meta); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	meta["mechanism"] = map[string]any{
		"kind":                   "official_api",
		"sanctioned_by_provider": true,
		"disabled_reason":        reason,
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	fsys[path] = &fstest.MapFile{Data: raw}
}
