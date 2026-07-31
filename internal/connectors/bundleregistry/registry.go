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

var bundleCache struct {
	once    sync.Once
	bundles []engine.Bundle
	err     error
}

func New() *connectors.Registry {
	bundles, err := cachedBundles()
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

func cachedBundles() ([]engine.Bundle, error) {
	bundleCache.once.Do(func() {
		bundleCache.bundles, bundleCache.err = engine.LoadAll(defs.FS)
	})
	return bundleCache.bundles, bundleCache.err
}
