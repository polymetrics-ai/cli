package app

import (
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synctransport"
)

// ensureTransportRegistry composes declaration-owned transports only when a
// saved transport operation needs them. App opening and metadata inspection do
// not require decoding every connector definition.
func (a *App) ensureTransportRegistry() error {
	if a == nil || a.registry == nil {
		return fmt.Errorf("definition-owned transport composition requires an app registry")
	}
	a.transportMu.Lock()
	defer a.transportMu.Unlock()
	if a.transports != nil {
		return nil
	}
	registry := a.transportRegistry
	if registry == nil {
		registry = a.registry
	}
	return a.composeTransportRegistryLocked(registry)
}

// composeTransportRegistry rebuilds production transports from the current
// registry. Tests that replace registry contents use this explicit operation.
func (a *App) composeTransportRegistry() error {
	if a == nil || a.registry == nil {
		return fmt.Errorf("definition-owned transport composition requires an app registry")
	}
	a.transportMu.Lock()
	defer a.transportMu.Unlock()
	return a.composeTransportRegistryLocked(a.registry)
}

func (a *App) composeTransportRegistryLocked(registry *connectors.Registry) error {
	if registry == nil {
		return fmt.Errorf("definition-owned transport composition requires an app registry")
	}
	factories, err := definitionTransportDefinitionFactories(a, registry)
	if err != nil {
		return err
	}
	factories = append(factories, localWarehouseTransportDefinitionFactories(a)...)
	connectorFactories, err := synctransport.DefinitionFactoriesFromRegistry(registry)
	if err != nil {
		return err
	}
	factories = append(factories, connectorFactories...)
	transports := synctransport.NewRegistry()
	if err := synctransport.RegisterDeclaredTransports(registry, transports, factories); err != nil {
		return err
	}
	a.transports = transports
	a.transportStage = newConnectionWarehouseStage(a)
	return nil
}
