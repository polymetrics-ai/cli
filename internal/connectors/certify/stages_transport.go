package certify

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/app"
)

const declaredTransportPairStage = "declared_transport_pair"

type certificationTransportPairProof = app.DeclaredTransportCertificationProof

// stageDeclaredTransportPair closes the gap between a declaration being
// syntactically valid and App actually resolving and executing it. The stage
// is connector-neutral: App reports not-applicable when the registered
// connector has no executable certification adapter for its declared pair.
func stageDeclaredTransportPair(rc *runContext, rep *Report) error {
	if !rc.opts.Full {
		skipStage(rc, rep, declaredTransportPairStage, "skipped: --full not set (declared transport proof is full-certificate only)")
		return nil
	}

	probe := rc.transportPairProbe
	if probe == nil {
		probe = func(ctx context.Context, root string) (certificationTransportPairProof, error) {
			return app.ProbeDeclaredTransportForCertification(ctx, root, rc.opts.Connector)
		}
	}
	proof, err := probe(rc.ctx, rc.root)
	cli := CLIStageInfo{
		ArgvRedacted: fmt.Sprintf("definition:%s.sync_transport %s -> %s", rc.opts.Connector, proof.SourceReference, proof.DestinationReference),
		Kind:         "DeclaredTransportPairProof",
	}
	if err != nil {
		recordStage(rc, rep, declaredTransportPairStage, 2, func() (bool, CLIStageInfo, string) {
			cli.ExitCode = 2
			return false, cli, err.Error()
		})
		return nil
	}
	if !proof.Applicable {
		skipStage(rc, rep, declaredTransportPairStage, "skipped: connector has no executable certification adapter for its declared transport pair")
		return nil
	}
	recordStage(rc, rep, declaredTransportPairStage, 2, func() (bool, CLIStageInfo, string) {
		if proof.SourceReference == "" || proof.DestinationReference == "" {
			cli.ExitCode = 2
			return false, cli, "resolved transport pair omitted a source or destination executor reference"
		}
		if proof.ProviderReads < 2 || proof.ProviderWrites != 1 || proof.RecordsRead != 1 || proof.RecordsLoaded != 1 || !proof.CheckpointCommitted || proof.WarehouseManifests != 1 || proof.WarehouseParquet != 1 {
			cli.ExitCode = 2
			return false, cli, fmt.Sprintf("declared transport pair did not execute durably: reads=%d writes=%d records_read=%d records_loaded=%d checkpoint_committed=%t warehouse_manifests=%d warehouse_parquet=%d", proof.ProviderReads, proof.ProviderWrites, proof.RecordsRead, proof.RecordsLoaded, proof.CheckpointCommitted, proof.WarehouseManifests, proof.WarehouseParquet)
		}
		return true, cli, ""
	})
	return nil
}
