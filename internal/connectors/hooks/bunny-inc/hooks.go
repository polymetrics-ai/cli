// Package bunnyinc bridges the bunny-inc quarantine bundle to the legacy connector.
package bunnyinc

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	legacy "polymetrics.ai/internal/connectors/bunny-inc"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("bunny-inc", func() engine.Hooks { return New() })
}

// Hooks implements CheckHook and StreamHook by delegating to the legacy connector.
type Hooks struct {
	Connector connectors.Connector
}

// New returns a hook set backed by the legacy bunny-inc connector.
func New() engine.Hooks { return Hooks{Connector: legacy.New()} }

func (h Hooks) ConnectorName() string { return "bunny-inc" }

var streamAliases = map[string]string{}

func (h Hooks) connector() connectors.Connector {
	if h.Connector != nil {
		return h.Connector
	}
	return legacy.New()
}

// Check delegates to the legacy connector's Check implementation.
func (h Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	return true, h.connector().Check(ctx, cfg)
}

// ReadStream delegates to the legacy connector's Read implementation.
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if req.Stream == "" {
		req.Stream = stream.Name
	}
	if legacyName, ok := streamAliases[req.Stream]; ok {
		req.Stream = legacyName
	}
	if req.Stream == "" {
		return true, fmt.Errorf("bunny-inc" + " stream name is required")
	}
	return true, h.connector().Read(ctx, req, emit)
}
