package bundleregistry

import (
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	_ "polymetrics.ai/internal/connectors/hooks/hookset"
	"polymetrics.ai/internal/connectors/native/nativeset"
)

var (
	bundleLoadOnce sync.Once
	loadedBundles  []engine.Bundle
	loadBundlesErr error
)

func init() {
	connectors.RegisterDefaultRegistryBuilder(New)
}

func New() *connectors.Registry {
	bundles, err := loadBundles()
	if err != nil {
		panic("load connector definition bundles: " + err.Error())
	}

	registry := connectors.NewEmptyRegistry()
	registry.RegisterBuiltins()
	for _, bundle := range bundles {
		registry.Register(engine.New(bundle, engine.HooksFor(bundle.Name)))
	}
	nativeset.RegisterInto(registry)
	return registry
}

func loadBundles() ([]engine.Bundle, error) {
	bundleLoadOnce.Do(func() {
		loadedBundles, loadBundlesErr = engine.LoadAll(defs.FS)
	})
	return loadedBundles, loadBundlesErr
}
