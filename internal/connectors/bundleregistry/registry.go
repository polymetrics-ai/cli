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

var registryCache struct {
	once     sync.Once
	registry *connectors.Registry
}

func New() *connectors.Registry {
	registryCache.once.Do(func() {
		registryCache.registry = buildRegistry()
	})
	return cloneRegistry(registryCache.registry)
}

func buildRegistry() *connectors.Registry {
	bundles, err := engine.LoadAll(defs.FS)
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

func cloneRegistry(src *connectors.Registry) *connectors.Registry {
	clone := connectors.NewEmptyRegistry()
	if src == nil {
		return clone
	}
	for _, meta := range src.List() {
		connector, ok := src.Get(meta.Name)
		if !ok {
			continue
		}
		clone.Register(connector)
	}
	return clone
}
