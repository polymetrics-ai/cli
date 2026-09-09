package app

import (
	"context"
	"errors"
	"polymetrics.ai/internal/connectors"
)

// PreflightConnectorSource inspects safe metadata before opening a project or vault.
// It never constructs an execution connector or selects a command by source ID.
func PreflightConnectorSource(ctx context.Context, registry *connectors.Registry, selection connectors.SourceCellSelection) (connectors.SourceCellPreflight, error) {
	return registry.PreflightSource(ctx, selection)
}
func (a *App) PreflightConnectorSource(ctx context.Context, selection connectors.SourceCellSelection) (connectors.SourceCellPreflight, error) {
	if a == nil {
		return connectors.SourceCellPreflight{}, errors.New("application is required")
	}
	return PreflightConnectorSource(ctx, a.registry, selection)
}
func (a *App) ConnectorSources(ctx context.Context, connector string) (connectors.SourceVisibility, error) {
	if a == nil {
		return connectors.SourceVisibility{}, errors.New("application is required")
	}
	return a.registry.SourceVisibility(ctx, connector)
}
