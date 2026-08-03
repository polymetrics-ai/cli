package awscloudtrail

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func streams(ctx context.Context) ([]connectors.Stream, error) {
	bundle, err := engine.Load(defs.FS, connectorName)
	if err != nil {
		return nil, fmt.Errorf("load aws-cloudtrail bundle catalog: %w", err)
	}
	catalog, err := engine.New(bundle, nil).Catalog(ctx, connectors.RuntimeConfig{})
	if err != nil {
		return nil, err
	}
	byName := make(map[string]connectors.Stream, len(catalog.Streams))
	for _, stream := range catalog.Streams {
		byName[stream.Name] = stream
	}
	out := make([]connectors.Stream, 0, len(cloudTrailPublishedStreams))
	for _, name := range cloudTrailPublishedStreams {
		stream, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("aws-cloudtrail stream %q missing from bundle catalog", name)
		}
		action, ok := cloudTrailStreamActions[name]
		if !ok {
			return nil, fmt.Errorf("aws-cloudtrail stream %q missing native action", name)
		}
		stream.Description = "AWS CloudTrail " + action + " read stream using a fixed signed JSON-RPC request."
		out = append(out, stream)
	}
	return out, nil
}
