package cli

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDynamicConnectorCommandsUseLazyMetadata(t *testing.T) {
	var loads atomic.Int32
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{
		Name:            "github",
		DisplayName:     "GitHub",
		IntegrationType: "api",
	}}, func(_ context.Context, name string) (connectors.Connector, error) {
		loads.Add(1)
		return engine.New(engine.Bundle{
			Name: name,
			Metadata: engine.Metadata{
				Name:            name,
				DisplayName:     "GitHub",
				IntegrationType: "api",
			},
			CLISurface: &engine.CLISurface{
				Usage:   "pm github <command>",
				Tagline: "Work with GitHub.",
			},
		}, nil), nil
	}, connectors.CommandSummary{Connector: "github", Usage: "pm github <command>", Tagline: "Work with GitHub."})
	if err != nil {
		t.Fatal(err)
	}

	section := dynamicConnectorCommandsSection(registry)
	if !strings.Contains(section, "pm github <command> - GitHub: Work with GitHub.") {
		t.Fatalf("dynamic connector command section = %q", section)
	}
	if got := loads.Load(); got != 0 {
		t.Fatalf("dynamicConnectorCommandsSection() resolved %d bundles, want 0", got)
	}
}

// Override one actual embedded artifact; all other bundle files remain real.
type malformedSelectedBundleFS160 struct{ fs.FS }

func (f malformedSelectedBundleFS160) Open(name string) (fs.File, error) {
	if name == "github/operations.json" {
		return fstest.MapFS{name: &fstest.MapFile{Data: []byte(`{"operations":[`)}}.Open(name)
	}
	return f.FS.Open(name)
}

func TestCLICommandPreservesSelectedDataError160(t *testing.T) {
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatal(err)
	}
	valid, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatal(err)
	}
	good := connectors.NewEmptyRegistry()
	if err := good.Register(engine.New(valid, nil)); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := runWithPreflightRegistry([]string{"github", "--help", "--root", root}, &out, &diag, good); code != 0 || !strings.Contains(out.String(), "pm github") {
		t.Fatalf("valid real bundle help control: %s %s", out.String(), diag.String())
	}
	sentinel := errors.New("selected-bundle-sentinel")
	var observed error
	loads := 0
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(ctx context.Context, name string) (connectors.Connector, error) {
		loads++
		_, failure := engine.Load(malformedSelectedBundleFS160{defs.FS}, name)
		if failure == nil || !strings.Contains(failure.Error(), "operations.json") {
			t.Fatalf("malformed actual loader control: %v", failure)
		}
		observed = errors.Join(sentinel, failure)
		return nil, observed
	})
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	diag.Reset()
	err = runMaybeConnectorCommandWithRegistry(t.Context(), root, "github", []string{"label", "delete"}, &out, &diag, false, registry)
	if loads != 1 || !errors.Is(err, sentinel) || !errors.Is(err, observed) {
		t.Fatalf("command selected-data cause lost: loads=%d got=%v want=%v", loads, err, observed)
	}
	out.Reset()
	diag.Reset()
	code := runWithPreflightRegistry([]string{"github", "label", "delete", "--root", root}, &out, &diag, registry)
	if code == 0 || !strings.Contains(diag.String(), "selected-bundle-sentinel") || !strings.Contains(diag.String(), "operations.json") || strings.Contains(diag.String(), "unknown command") || out.Len() != 0 {
		t.Fatalf("public CLI erased selected-data reason: stdout=%q stderr=%q", out.String(), diag.String())
	}
}

func TestCLICommandSelectionBoundaries160(t *testing.T) {
	valid, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatal(err)
	}
	loads := 0
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "github", DisplayName: "GitHub", IntegrationType: "api"}}, func(context.Context, string) (connectors.Connector, error) {
		loads++
		return engine.New(valid, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	err = runMaybeConnectorCommandWithRegistry(t.Context(), t.TempDir(), "absent-connector", nil, &out, &diag, false, registry)
	if err == nil || err.Error() != `unknown command "absent-connector"` || loads != 0 {
		t.Fatalf("unknown selection changed: %v loads=%d", err, loads)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err = runMaybeConnectorCommandWithRegistry(ctx, t.TempDir(), "github", nil, &out, &diag, false, registry)
	if !errors.Is(err, context.Canceled) || loads != 0 {
		t.Fatalf("cancellation crossed selected construction: %v loads=%d", err, loads)
	}
	err = runMaybeConnectorCommandWithRegistry(t.Context(), t.TempDir(), "github", nil, &out, &diag, false, registry)
	if err != nil || !strings.Contains(out.String(), "pm github") || loads != 1 {
		t.Fatalf("valid selection did not reach manual: %v loads=%d", err, loads)
	}
}
