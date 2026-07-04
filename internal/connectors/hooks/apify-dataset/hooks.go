// Package apifydataset bridges the apify-dataset quarantine bundle to the legacy connector.
package apifydataset

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	legacy "polymetrics.ai/internal/connectors/apify-dataset"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("apify-dataset", func() engine.Hooks { return New() })
}

// Hooks implements CheckHook and StreamHook by delegating to the legacy connector.
type Hooks struct {
	Connector connectors.Connector
}

// New returns a hook set backed by the legacy apify-dataset connector.
func New() engine.Hooks { return Hooks{Connector: legacy.New()} }

func (h Hooks) ConnectorName() string { return "apify-dataset" }

var streamAliases = map[string]string{}

var legacyStreams = map[string]struct{}{
	"item_collection":    {},
	"dataset_collection": {},
	"dataset":            {},
}

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
		return true, fmt.Errorf("apify-dataset" + " stream name is required")
	}
	if _, ok := legacyStreams[req.Stream]; !ok {
		return false, nil
	}
	return true, h.connector().Read(ctx, req, emit)
}
