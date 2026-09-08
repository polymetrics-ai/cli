package synctransport

import (
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// DefinitionFactory binds one exact definition-owned executor reference to its
// provider-specific adapter constructor. It deliberately names no connector:
// one explicit factory can be reused by a second connector that declares the
// same closed role.
type DefinitionFactory struct {
	Reference        connectors.TransportExecutorReference
	BuildSource      func(connectors.Connector) (SourceExecutor, error)
	BuildDestination func(connectors.Connector) (DestinationExecutor, error)
}

// DefinitionFactoryProvider lets a connector-local package expose adapters for
// the exact transport roles its bundle may declare. The production composition
// root reads this optional interface generically and never imports a native
// connector package to discover its factories.
type DefinitionFactoryProvider interface {
	SyncTransportDefinitionFactories() []DefinitionFactory
}

type declaredSource struct {
	connector connectors.Connector
	factory   DefinitionFactory
}

type declaredDestination struct {
	connector connectors.Connector
	factory   DefinitionFactory
}

// DefinitionFactoriesFromRegistry gathers factory providers from the concrete
// connectors already admitted to a registry. It does not inspect connector
// names, capabilities, or transport declarations; declaration validation stays
// in RegisterDeclaredTransports.
func DefinitionFactoriesFromRegistry(connectorRegistry *connectors.Registry) ([]DefinitionFactory, error) {
	if connectorRegistry == nil {
		return nil, fmt.Errorf("connector registry is required")
	}
	factories := make([]DefinitionFactory, 0)
	for _, metadata := range connectorRegistry.List() {
		connector, ok := connectorRegistry.Get(metadata.Name)
		if !ok {
			return nil, fmt.Errorf("declared connector %q disappeared from registry", metadata.Name)
		}
		provider, ok := connector.(DefinitionFactoryProvider)
		if !ok {
			continue
		}
		factories = append(factories, provider.SyncTransportDefinitionFactories()...)
	}
	return factories, nil
}

// RegisterDeclaredTransports validates every declaration and selected factory
// before any executor is built or registered. This makes unknown and malformed
// bundle declarations fail closed with zero registry mutation.
func RegisterDeclaredTransports(connectorRegistry *connectors.Registry, transportRegistry *Registry, factories []DefinitionFactory) error {
	if connectorRegistry == nil {
		return fmt.Errorf("connector registry is required")
	}
	if transportRegistry == nil {
		return fmt.Errorf("transport registry is required")
	}
	factoryByReference, err := definitionFactoryIndex(factories)
	if err != nil {
		return err
	}

	sources := make([]declaredSource, 0)
	destinations := make([]declaredDestination, 0)
	seenSourceReferences := make(map[connectors.TransportExecutorReference]struct{})
	seenDestinationReferences := make(map[connectors.TransportExecutorReference]struct{})
	for _, metadata := range connectorRegistry.List() {
		connector, ok := connectorRegistry.Get(metadata.Name)
		if !ok {
			return fmt.Errorf("declared connector %q disappeared from registry", metadata.Name)
		}
		descriptor, declared := connectors.SyncTransportDescriptorOf(connector)
		if !declared {
			continue
		}
		if err := descriptor.Validate(); err != nil {
			return fmt.Errorf("declared transport for connector %q: %w", connector.Name(), err)
		}
		if descriptor.Source != nil {
			factory, exists := factoryByReference[descriptor.Source.Executor]
			if !exists || factory.BuildSource == nil {
				return fmt.Errorf("declared source transport executor %q is not registered", descriptor.Source.Executor.ID)
			}
			if err := connectors.ValidateTransportExecutorFamily(connector.Metadata().IntegrationType, descriptor.Source.Executor); err != nil {
				return fmt.Errorf("declared source transport for connector %q: %w", connector.Name(), err)
			}
			if _, seen := seenSourceReferences[descriptor.Source.Executor]; !seen {
				sources = append(sources, declaredSource{connector: connector, factory: factory})
				seenSourceReferences[descriptor.Source.Executor] = struct{}{}
			}
		}
		if descriptor.Destination != nil {
			factory, exists := factoryByReference[descriptor.Destination.Executor]
			if !exists || factory.BuildDestination == nil {
				return fmt.Errorf("declared destination transport executor %q is not registered", descriptor.Destination.Executor.ID)
			}
			if err := connectors.ValidateTransportExecutorFamily(connector.Metadata().IntegrationType, descriptor.Destination.Executor); err != nil {
				return fmt.Errorf("declared destination transport for connector %q: %w", connector.Name(), err)
			}
			if _, seen := seenDestinationReferences[descriptor.Destination.Executor]; !seen {
				destinations = append(destinations, declaredDestination{connector: connector, factory: factory})
				seenDestinationReferences[descriptor.Destination.Executor] = struct{}{}
			}
		}
	}

	sourceExecutors := make([]SourceExecutor, 0, len(sources))
	for _, declared := range sources {
		executor, err := declared.factory.BuildSource(declared.connector)
		if err != nil {
			return fmt.Errorf("build declared source transport %q: %w", declared.factory.Reference.ID, err)
		}
		if isNilInterface(executor) || executor.TransportExecutorReference() != declared.factory.Reference {
			return fmt.Errorf("declared source transport factory %q returned a mismatched executor", declared.factory.Reference.ID)
		}
		sourceExecutors = append(sourceExecutors, executor)
	}
	destinationExecutors := make([]DestinationExecutor, 0, len(destinations))
	for _, declared := range destinations {
		executor, err := declared.factory.BuildDestination(declared.connector)
		if err != nil {
			return fmt.Errorf("build declared destination transport %q: %w", declared.factory.Reference.ID, err)
		}
		if isNilInterface(executor) || executor.TransportExecutorReference() != declared.factory.Reference {
			return fmt.Errorf("declared destination transport factory %q returned a mismatched executor", declared.factory.Reference.ID)
		}
		destinationExecutors = append(destinationExecutors, executor)
	}
	return transportRegistry.registerDefinitionBatch(sourceExecutors, destinationExecutors)
}

func definitionFactoryIndex(factories []DefinitionFactory) (map[connectors.TransportExecutorReference]DefinitionFactory, error) {
	indexed := make(map[connectors.TransportExecutorReference]DefinitionFactory, len(factories))
	for _, factory := range factories {
		if err := factory.Reference.Validate(); err != nil {
			return nil, fmt.Errorf("definition transport factory reference: %w", err)
		}
		if factory.BuildSource == nil && factory.BuildDestination == nil {
			return nil, fmt.Errorf("definition transport factory %q has no role builder", factory.Reference.ID)
		}
		if _, exists := indexed[factory.Reference]; exists {
			return nil, fmt.Errorf("duplicate definition transport factory for executor %q", factory.Reference.ID)
		}
		indexed[factory.Reference] = factory
	}
	return indexed, nil
}

func (r *Registry) registerDefinitionBatch(sources []SourceExecutor, destinations []DestinationExecutor) error {
	if r == nil {
		return fmt.Errorf("transport registry is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sources == nil {
		r.sources = make(map[connectors.TransportExecutorReference]SourceExecutor)
	}
	if r.destinations == nil {
		r.destinations = make(map[connectors.TransportExecutorReference]DestinationExecutor)
	}
	newSources := make(map[connectors.TransportExecutorReference]SourceExecutor, len(sources))
	for _, executor := range sources {
		if isNilInterface(executor) {
			return fmt.Errorf("source transport executor is required")
		}
		reference := executor.TransportExecutorReference()
		if err := reference.Validate(); err != nil {
			return err
		}
		if _, exists := r.sources[reference]; exists {
			return fmt.Errorf("source transport executor %q is already registered", reference.ID)
		}
		if _, exists := newSources[reference]; exists {
			return fmt.Errorf("source transport executor %q is registered more than once", reference.ID)
		}
		newSources[reference] = executor
	}
	newDestinations := make(map[connectors.TransportExecutorReference]DestinationExecutor, len(destinations))
	for _, executor := range destinations {
		if isNilInterface(executor) {
			return fmt.Errorf("destination transport executor is required")
		}
		reference := executor.TransportExecutorReference()
		if err := reference.Validate(); err != nil {
			return err
		}
		if _, exists := r.destinations[reference]; exists {
			return fmt.Errorf("destination transport executor %q is already registered", reference.ID)
		}
		if _, exists := newDestinations[reference]; exists {
			return fmt.Errorf("destination transport executor %q is registered more than once", reference.ID)
		}
		newDestinations[reference] = executor
	}
	for reference, executor := range newSources {
		r.sources[reference] = executor
	}
	for reference, executor := range newDestinations {
		r.destinations[reference] = executor
	}
	return nil
}
