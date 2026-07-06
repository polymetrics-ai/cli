// Package fastbill bridges the fastbill quarantine bundle to the native connector.
package fastbill

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/fastbill"
)

func init() {
	engine.RegisterHooks("fastbill", func() engine.Hooks { return New() })
}

// Hooks implements CheckHook and StreamHook by delegating to the native connector.
type Hooks struct {
	Connector connectors.Connector
}

// New returns a hook set backed by the native fastbill connector.
func New() engine.Hooks { return Hooks{Connector: native.New()} }

func (h Hooks) ConnectorName() string { return "fastbill" }

var streamAliases = map[string]string{}

var legacyStreams = map[string]struct{}{
	"customers":          {},
	"invoices":           {},
	"products":           {},
	"recurring_invoices": {},
	"revenues":           {},
}

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
		return true, fmt.Errorf("fastbill" + " stream name is required")
	}
	if _, ok := legacyStreams[req.Stream]; !ok {
		return false, nil
	}
	return true, h.connector().Read(ctx, req, emit)
}
