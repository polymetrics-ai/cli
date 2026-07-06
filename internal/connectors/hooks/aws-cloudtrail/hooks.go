// Package awscloudtrail bridges the aws-cloudtrail quarantine bundle to the native connector.
package awscloudtrail

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	native "polymetrics.ai/internal/connectors/native/aws-cloudtrail"
)

func init() {
	engine.RegisterHooks("aws-cloudtrail", func() engine.Hooks { return New() })
}

// Hooks implements CheckHook and StreamHook by delegating to the native connector.
type Hooks struct {
	Connector connectors.Connector
}

// New returns a hook set backed by the native aws-cloudtrail connector.
func New() engine.Hooks { return Hooks{Connector: native.New()} }

func (h Hooks) ConnectorName() string { return "aws-cloudtrail" }

var streamAliases = map[string]string{}

var legacyStreams = map[string]struct{}{
	"management_events": {},
	"read_only_events":  {},
	"write_only_events": {},
	"console_logins":    {},
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
		return true, fmt.Errorf("aws-cloudtrail" + " stream name is required")
	}
	if _, ok := legacyStreams[req.Stream]; !ok {
		return false, nil
	}
	return true, h.connector().Read(ctx, req, emit)
}
