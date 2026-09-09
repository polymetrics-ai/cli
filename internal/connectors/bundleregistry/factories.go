package bundleregistry

import (
	"errors"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

var (
	ErrInvalidExecutorFactory   = errors.New("invalid executor factory")
	ErrDuplicateExecutorFactory = errors.New("duplicate executor factory")
	ErrInvalidExecutorSelection = errors.New("invalid executor selection")
	ErrUnknownExecutor          = errors.New("unknown executor")
)

// ExecutorFactory constructs one closed executor kind for an execution bundle.
type ExecutorFactory struct {
	ID        string
	Construct func(engine.Bundle) (connectors.Connector, error)
}

// ExecutorFactories is an immutable set of explicitly supplied constructors.
type ExecutorFactories struct {
	factories map[string]ExecutorFactory
}

// NewExecutorFactories validates all constructor identities before a constructor can run.
func NewExecutorFactories(factories ...ExecutorFactory) (ExecutorFactories, error) {
	out := ExecutorFactories{factories: make(map[string]ExecutorFactory, len(factories))}
	for _, factory := range factories {
		if factory.ID == "" || factory.Construct == nil {
			return ExecutorFactories{}, fmt.Errorf("%w: ID and constructor are required", ErrInvalidExecutorFactory)
		}
		if _, exists := out.factories[factory.ID]; exists {
			return ExecutorFactories{}, fmt.Errorf("%w: %q", ErrDuplicateExecutorFactory, factory.ID)
		}
		out.factories[factory.ID] = factory
	}
	return out, nil
}

// Has reports whether the closed factory set owns executor.
func (f ExecutorFactories) Has(executor string) bool {
	_, ok := f.factories[executor]
	return ok
}

// Construct builds the exact executor selected by entry for bundle.
func (f ExecutorFactories) Construct(entry manifestindex.Entry, bundle engine.Bundle) (connectors.Connector, error) {
	if entry.Connector == "" || entry.Executor == "" || bundle.Name != entry.Connector {
		return nil, fmt.Errorf("%w: connector and executor must match the bundle", ErrInvalidExecutorSelection)
	}
	factory, ok := f.factories[entry.Executor]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownExecutor, entry.Executor)
	}
	connector, err := factory.Construct(bundle)
	if err != nil {
		return nil, fmt.Errorf("construct %q for %q: %w", entry.Executor, entry.Connector, err)
	}
	if connector == nil || connector.Name() != entry.Connector {
		return nil, fmt.Errorf("%w: constructor %q returned the wrong connector", ErrInvalidExecutorSelection, entry.Executor)
	}
	return connector, nil
}
