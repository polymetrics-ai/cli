// Package awscloudtrail bridges the aws-cloudtrail declarative bundle to the native CloudTrail executor.
package awscloudtrail

import (
	"context"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/aws-cloudtrail"
)

func init() {
	engine.RegisterHooks("aws-cloudtrail", func() engine.Hooks { return New() })
}

// Hooks delegates AWS JSON-RPC streams, writes, and checks to the native
// SigV4 executor while preserving the declarative bundle for validation,
// docs, CLI metadata, fixtures, and conformance replay.
type Hooks struct {
	Connector connectors.Connector
}

func New() engine.Hooks { return Hooks{Connector: native.New()} }

func (h Hooks) ConnectorName() string { return "aws-cloudtrail" }

func (h Hooks) connector() connectors.Connector {
	if h.Connector != nil {
		return h.Connector
	}
	return native.New()
}

func (h Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	return true, h.connector().Check(ctx, runtimeConfig(cfg, rt))
}

func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if req.Stream == "" {
		req.Stream = stream.Name
	}
	if strings.TrimSpace(req.Stream) == "" {
		return true, fmt.Errorf("aws-cloudtrail stream name is required")
	}
	req.Config = runtimeConfig(req.Config, rt)
	return true, h.connector().Read(ctx, req, emit)
}

func (h Hooks) ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (bool, error) {
	cfg := runtimeConfig(connectors.RuntimeConfig{}, rt)
	_, err := h.connector().Write(ctx, connectors.WriteRequest{Action: action.Name, Config: cfg}, []connectors.Record{rec})
	return true, err
}

func runtimeConfig(cfg connectors.RuntimeConfig, rt *engine.Runtime) connectors.RuntimeConfig {
	if cfg.Config == nil {
		cfg.Config = map[string]string{}
	}
	if cfg.Secrets == nil {
		cfg.Secrets = map[string]string{}
	}
	if rt != nil {
		for k, v := range rt.Config.Config {
			if _, ok := cfg.Config[k]; !ok {
				cfg.Config[k] = v
			}
		}
		for k, v := range rt.Config.Secrets {
			if _, ok := cfg.Secrets[k]; !ok {
				cfg.Secrets[k] = v
			}
		}
		if rt.Requester != nil && isReplayURL(rt.Requester.BaseURL) {
			cfg.Config["base_url"] = rt.Requester.BaseURL
		}
	}
	return cfg
}

func isReplayURL(base string) bool {
	return strings.HasPrefix(base, "http://127.0.0.1:") || strings.HasPrefix(base, "http://localhost:")
}
