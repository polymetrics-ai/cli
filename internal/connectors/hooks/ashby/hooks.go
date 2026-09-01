// Package ashby bridges the Ashby bundle to the native connector.
package ashby

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/ashby"
)

func init() {
	engine.RegisterHooks("ashby", func() engine.Hooks { return New() })
}

// Hooks implements CheckHook and StreamHook by delegating to the native connector.
type Hooks struct {
	Connector connectors.Connector
}

// New returns a hook set backed by the native ashby connector.
func New() engine.Hooks { return Hooks{Connector: native.New()} }

func (h Hooks) ConnectorName() string { return "ashby" }

var streamAliases = map[string]string{}

func (h Hooks) connector() connectors.Connector {
	if h.Connector != nil {
		return h.Connector
	}
	return native.New()
}

func hookConfig(cfg connectors.RuntimeConfig, rt *engine.Runtime) connectors.RuntimeConfig {
	if rt == nil || rt.Requester == nil || rt.Requester.BaseURL == "" {
		return cfg
	}
	out := cfg
	out.Config = make(map[string]string, len(cfg.Config)+1)
	for key, value := range cfg.Config {
		out.Config[key] = value
	}
	out.Config["base_url"] = rt.Requester.BaseURL
	return out
}

// Check delegates to the native connector's Check implementation.
func (h Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	return true, h.connector().Check(ctx, hookConfig(cfg, rt))
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
		return true, fmt.Errorf("ashby" + " stream name is required")
	}
	req.Config = hookConfig(req.Config, rt)
	return true, h.connector().Read(ctx, req, emit)
}
