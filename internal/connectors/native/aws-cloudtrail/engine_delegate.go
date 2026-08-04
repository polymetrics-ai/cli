package awscloudtrail

import (
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

var (
	cloudTrailBundleOnce  sync.Once
	cloudTrailBundleValue engine.Bundle
	cloudTrailBundleErr   error
)

func cloudTrailBundle() engine.Bundle {
	cloudTrailBundleOnce.Do(func() {
		cloudTrailBundleValue, cloudTrailBundleErr = engine.Load(defs.FS, connectorName)
	})
	if cloudTrailBundleErr != nil {
		panic("native/aws-cloudtrail: failed to load defs/aws-cloudtrail bundle: " + cloudTrailBundleErr.Error())
	}
	return cloudTrailBundleValue
}

func cloudTrailEngineConnector() *engine.Connector {
	return engine.New(cloudTrailBundle(), nil)
}

// CommandSurface exposes the generated, typed CLI surface for AWS CloudTrail.
func (c Connector) CommandSurface() *connectors.CommandSurface {
	return cloudTrailEngineConnector().CommandSurface()
}
