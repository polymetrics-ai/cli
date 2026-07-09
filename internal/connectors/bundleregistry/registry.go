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
	loadOnce      sync.Once
	loadedBundles []engine.Bundle
	loadErr       error
)

func New() *connectors.Registry {
	bundles := loadDeclarativeBundles()

	registry := connectors.NewEmptyRegistry()
	registry.RegisterBuiltins()
	for _, bundle := range bundles {
		registry.Register(engine.New(bundle, engine.HooksFor(bundle.Name)))
	}
	nativeset.RegisterInto(registry)
	return registry
}

func loadDeclarativeBundles() []engine.Bundle {
	loadOnce.Do(func() {
		loadedBundles, loadErr = engine.LoadAll(defs.FS)
	})
	if loadErr != nil {
		panic("load connector definition bundles: " + loadErr.Error())
	}
	return loadedBundles
}
