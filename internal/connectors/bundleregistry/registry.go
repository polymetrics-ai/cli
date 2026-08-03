package bundleregistry

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/hookset"
	"polymetrics.ai/internal/connectors/native/nativeset"
)

func init() {
	connectors.RegisterDefaultRegistryBuilder(New)
}

// definitionsFS is the bundle root New reads. It is a package variable so a
// test can substitute a fixture tree; production always uses defs.FS.
var definitionsFS fs.FS = defs.FS

// New builds the default connector registry from the embedded definition
// bundles.
//
// A bundle that fails to load is omitted from the catalog and reported on
// stderr; it never stops the bundles that did load from registering. That
// partial-success behavior is what makes mechanism.disabled_reason a usable
// kill switch: shipping a patch release that disables one connector must
// make that one connector disappear, not abort every pm invocation.
func New() *connectors.Registry {
	bundles, err := engine.LoadAll(definitionsFS)
	if err != nil {
		reportBundleLoadFailures(err)
	}

	registry := connectors.NewEmptyRegistry()
	registry.RegisterBuiltins()
	for _, bundle := range bundles {
		registry.Register(engine.New(bundle, engine.HooksFor(bundle.Name)))
	}
	nativeset.RegisterInto(registry)
	return registry
}

// warnf emits New's diagnostics. New takes no writer (its two callers build
// registries far from any command's plumbing), so this writes to os.Stderr
// directly — unprefixed, matching the rest of pm's output rather than the
// stdlib log package's timestamped format. Tests swap it to capture output.
var warnf = func(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func reportBundleLoadFailures(err error) {
	var loadAll *engine.LoadAllError
	if !errors.As(err, &loadAll) {
		warnf("warning: load connector definition bundles: %v\n", err)
		return
	}
	for _, failure := range loadAll.GetFailures() {
		warnf("warning: connector %q omitted from the catalog: %v\n", failure.Name, failure.Err)
	}
}
