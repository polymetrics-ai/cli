package app

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func (a *App) validateETLReadInputs(ctx context.Context, endpoint EndpointConfig, stream string) error {
	if err := connectors.RejectLegacyConnectorName(endpoint.Connector); err != nil {
		return err
	}
	credential, ok := a.findCredential(endpoint.Credential)
	if !ok {
		return fmt.Errorf("configured source credential not found")
	}
	if endpoint.Connector != "" && endpoint.Connector != credential.Connector {
		return fmt.Errorf("source connector does not match its credential")
	}
	source, err := a.registry.Resolve(ctx, credential.Connector)
	if err != nil {
		return engine.SafeBundleError(err)
	}
	validator, ok := source.(connectors.ReadInputValidator)
	if !ok {
		return nil
	}
	config := cloneStringMap(credential.Config)
	if config == nil {
		config = map[string]string{}
	}
	for name, value := range endpoint.Config {
		config[name] = value
	}
	return validator.ValidateReadInputs(ctx, connectors.ReadInputValidationRequest{Stream: stream, Config: config})
}
