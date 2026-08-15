package app

import (
	"fmt"

	"polymetrics.ai/internal/connectors/native/postgres"
	"polymetrics.ai/internal/synctransport"
)

// composeTransportRegistry builds production transports only from definition
// declarations. Provider-specific constructors are keyed by exact executor
// references; App never selects a connector by name or capability.
func (a *App) composeTransportRegistry() error {
	if a == nil || a.registry == nil {
		return fmt.Errorf("definition-owned transport composition requires an app registry")
	}
	factories := issueLabelTransportDefinitionFactories(a)
	factories = append(factories, postgres.SnapshotTransportDefinitionFactory())
	verifier, err := synctransport.NewDefinitionConformanceVerifier(factories)
	if err != nil {
		return err
	}
	transports := synctransport.NewRegistry(verifier)
	if err := synctransport.RegisterDeclaredTransports(a.registry, transports, factories); err != nil {
		return err
	}
	a.transports = transports
	a.transportStage = newConnectionWarehouseStage(a)
	return nil
}
