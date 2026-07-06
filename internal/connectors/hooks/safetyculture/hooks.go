// Package safetyculture bridges the safetyculture quarantine bundle to the native connector.
package safetyculture

import (
	"context"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/safetyculture"
)

func init() {
	engine.RegisterHooks("safetyculture", func() engine.Hooks { return New() })
}

// Hooks implements CheckHook and StreamHook by delegating to the native connector.
type Hooks struct {
	Connector connectors.Connector
}

// New returns a hook set backed by the native safetyculture connector.
func New() engine.Hooks { return Hooks{Connector: native.New()} }

func (h Hooks) ConnectorName() string { return "safetyculture" }

var handledStreams = map[string]struct{}{
	"audits":    {},
	"templates": {},
	"users":     {},
}

var streamAliases = map[string]string{}

func (h Hooks) connector() connectors.Connector {
	if h.Connector != nil {
		return h.Connector
	}
	return native.New()
}

// Check delegates to the native connector's Check implementation.
func (h Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	return true, h.connector().Check(ctx, cfg)
}

// ReadStream delegates to the native connector's Read implementation.
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if req.Stream == "" {
		req.Stream = stream.Name
	}
	if legacyName, ok := streamAliases[req.Stream]; ok {
		req.Stream = legacyName
	}
	if req.Stream == "" {
		req.Stream = "audits"
	}
	if _, ok := handledStreams[req.Stream]; !ok {
		return false, nil
	}
	return true, h.connector().Read(ctx, req, emit)
}
