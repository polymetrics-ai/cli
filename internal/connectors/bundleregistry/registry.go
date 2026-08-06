package bundleregistry

import (
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/hookset"
	"polymetrics.ai/internal/connectors/native/nativeset"
)

func init() {
	connectors.RegisterDefaultRegistryBuilder(New)
}

// cacheBundleLoad loads immutable embedded bundle definitions once per
// process. New still constructs a fresh registry and fresh connector/hooks
// for every caller, so app-scoped execution state is never shared.
func cacheBundleLoad(load func() ([]engine.Bundle, error)) func() ([]engine.Bundle, error) {
	return sync.OnceValues(load)
}

var loadBundledDefinitions = cacheBundleLoad(func() ([]engine.Bundle, error) {
	return engine.LoadAll(defs.FS)
})

func New() *connectors.Registry {
	bundles, err := loadBundledDefinitions()
	if err != nil {
		panic("load connector definition bundles: " + err.Error())
	}

	registry := connectors.NewEmptyRegistry()
	registry.RegisterBuiltins()
	for _, bundle := range bundles {
		registry.Register(engine.New(bundle, engine.HooksFor(bundle.Name)))
	}
	nativeset.RegisterInto(registry)
	registry.MustValidateIconCoverage()
	return registry
}
