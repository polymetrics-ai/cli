package bundleregistry

import (
	"fmt"

	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/hooks/hookset"
)

type hookFactory struct {
	connector string
	new       func() engine.Hooks
}

type hookFactories struct {
	factories   map[string]hookFactory
	byConnector map[string]string
}

func newHookFactories(factories []hookset.Factory) (hookFactories, error) {
	out := hookFactories{
		factories:   make(map[string]hookFactory, len(factories)),
		byConnector: make(map[string]string, len(factories)),
	}
	for _, factory := range factories {
		if factory.ID == "" || factory.Connector == "" || factory.New == nil {
			return hookFactories{}, fmt.Errorf("generated hook factory is incomplete")
		}
		if _, exists := out.factories[factory.ID]; exists {
			return hookFactories{}, fmt.Errorf("duplicate generated hook extension %q", factory.ID)
		}
		if _, exists := out.byConnector[factory.Connector]; exists {
			return hookFactories{}, fmt.Errorf("duplicate generated hook connector %q", factory.Connector)
		}
		out.factories[factory.ID] = hookFactory{connector: factory.Connector, new: factory.New}
		out.byConnector[factory.Connector] = factory.ID
	}
	return out, nil
}

func (f hookFactories) validate(id, connector string) error {
	if id == "" {
		return nil
	}
	factory, ok := f.factories[id]
	if !ok {
		return fmt.Errorf("unknown generated hook extension %q for %q", id, connector)
	}
	if factory.connector != connector {
		return fmt.Errorf("generated hook extension %q belongs to %q, not %q", id, factory.connector, connector)
	}
	return nil
}

func (f hookFactories) construct(id, connector string) (engine.Hooks, error) {
	if err := f.validate(id, connector); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, nil
	}
	hooks := f.factories[id].new()
	if hooks == nil {
		return nil, fmt.Errorf("generated hook extension %q returned nil hooks", id)
	}
	if hooks.ConnectorName() != connector {
		return nil, fmt.Errorf("generated hook extension %q returned %q for %q", id, hooks.ConnectorName(), connector)
	}
	return hooks, nil
}
