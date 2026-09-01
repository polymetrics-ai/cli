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

var definitionCache struct {
	once    sync.Once
	bundles []engine.Bundle
	err     error
}

// loadDefinitions compiles the immutable embedded definitions once per
// process. New still constructs a fresh registry, connector, and hook set for
// each caller, so callers can safely register test-specific connectors without
// sharing mutable registry state.
func loadDefinitions() ([]engine.Bundle, error) {
	definitionCache.once.Do(func() {
		definitionCache.bundles, definitionCache.err = engine.LoadAll(defs.FS)
	})
	return definitionCache.bundles, definitionCache.err
}

func New() *connectors.Registry {
	bundles, err := loadDefinitions()
	if err != nil && len(bundles) == 0 {
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
