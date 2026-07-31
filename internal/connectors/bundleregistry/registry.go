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

var (
	bundlesOnce sync.Once
	bundles     []engine.Bundle
	bundlesErr  error
)

func New() *connectors.Registry {
	loaded, err := loadBundles()
	if err != nil {
		panic("load connector definition bundles: " + err.Error())
	}

	registry := connectors.NewEmptyRegistry()
	registry.RegisterBuiltins()
	for _, bundle := range loaded {
		registry.Register(engine.New(bundle, engine.HooksFor(bundle.Name)))
	}
	nativeset.RegisterInto(registry)
	return registry
}

func loadBundles() ([]engine.Bundle, error) {
	bundlesOnce.Do(func() {
		bundles, bundlesErr = engine.LoadAll(defs.FS)
	})
	return bundles, bundlesErr
}
