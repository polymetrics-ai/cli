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
			return app.ProbeDeclaredTransportForCertificationRequest(ctx, app.DeclaredTransportCertificationRequest{
				CertificationRoot:       root,
				Connector:               rc.opts.Connector,
				SourceCredential:        sourceCredentialName,
				Stream:                  rc.streamName(),
				PrimaryKey:              rc.primaryKey(),
				CursorField:             rc.cursorField(),
				AllowManagedTargetWrite: rc.opts.Write,
			})
		}
	}
	proof, err := probe(rc.ctx, rc.root)
	cli := CLIStageInfo{
		ArgvRedacted: fmt.Sprintf("definition:%s.sync_transport %s -> %s", rc.opts.Connector, proof.SourceReference, proof.DestinationReference),
		Kind:         "DeclaredTransportPairProof",
	}
	if err != nil {
		setDeclaredTransportResult(rep, proof, "fail", err.Error())
		recordStage(rc, rep, declaredTransportPairStage, 2, func() (bool, CLIStageInfo, string) {
			cli.ExitCode = 2
			return false, cli, err.Error()
		})
		return nil
	}
	if !proof.Applicable {
		if !proof.Declared {
			setDeclaredTransportResult(rep, proof, "skipped", "connector has no declared source/destination transport pair")
			skipStage(rc, rep, declaredTransportPairStage, "skipped: connector has no declared source/destination transport pair")
			return nil
		}
		setDeclaredTransportResult(rep, proof, "unexecutable", "connector has no executable certification adapter for its declared transport pair")
		unexecutableStage(rc, rep, declaredTransportPairStage, "connector has no executable certification adapter for its declared transport pair")
		return nil
	}
	if proof.SkipReason != "" {
		setDeclaredTransportResult(rep, proof, "skipped", proof.SkipReason)
		skipStage(rc, rep, declaredTransportPairStage, "skipped: "+proof.SkipReason)
		return nil
	}
	stage := recordStage(rc, rep, declaredTransportPairStage, 2, func() (bool, CLIStageInfo, string) {
		if proof.SourceReference == "" || proof.DestinationReference == "" {
			cli.ExitCode = 2
			return false, cli, "resolved transport pair omitted a source or destination executor reference"
		}
		if len(proof.Modes) > 0 {
			if proof.RecordsRead == 0 || proof.RecordsLoaded == 0 || !proof.CheckpointCommitted {
				cli.ExitCode = 2
				return false, cli, fmt.Sprintf("declared database transport pair did not execute durably: records_read=%d records_loaded=%d checkpoint_committed=%t", proof.RecordsRead, proof.RecordsLoaded, proof.CheckpointCommitted)
			}
			for _, mode := range proof.Modes {
				if mode.Mode == "" || mode.ApplyStrategy == "" || mode.RecordsRead == 0 || mode.RecordsLoaded == 0 || !mode.CheckpointCommitted || mode.TargetNamespace == "" || mode.TargetRelation == "" {
					cli.ExitCode = 2
					return false, cli, fmt.Sprintf("declared database transport mode %q did not produce target/read/checkpoint evidence", mode.Mode)
				}
			}
			return true, cli, ""
		}
		if proof.ProviderReads < 2 || proof.ProviderWrites != 1 || proof.RecordsRead != 1 || proof.RecordsLoaded != 1 || !proof.CheckpointCommitted || proof.WarehouseManifests != 1 || proof.WarehouseParquet != 1 {
			cli.ExitCode = 2
			return false, cli, fmt.Sprintf("declared transport pair did not execute durably: reads=%d writes=%d records_read=%d records_loaded=%d checkpoint_committed=%t warehouse_manifests=%d warehouse_parquet=%d", proof.ProviderReads, proof.ProviderWrites, proof.RecordsRead, proof.RecordsLoaded, proof.CheckpointCommitted, proof.WarehouseManifests, proof.WarehouseParquet)
		}
		return true, cli, ""
	})
	if stage.Passed {
		setDeclaredTransportResult(rep, proof, "pass", "")
	} else {
		setDeclaredTransportResult(rep, proof, "fail", stage.Error)
	}
	return nil
}

func setDeclaredTransportResult(rep *Report, proof certificationTransportPairProof, result, reason string) {
	if rep == nil {
		return
	}
	transport := &DeclaredTransportResult{
		Result: result, SourceExecutor: proof.SourceReference, DestinationExecutor: proof.DestinationReference,
		RecordsRead: proof.RecordsRead, RecordsLoaded: proof.RecordsLoaded,
		CheckpointCommitted: proof.CheckpointCommitted, Reason: reason,
	}
	if len(proof.Modes) > 0 {
		transport.Modes = make([]DeclaredTransportModeResult, 0, len(proof.Modes))
		for _, mode := range proof.Modes {
			transport.Modes = append(transport.Modes, DeclaredTransportModeResult{
				Mode: mode.Mode, ApplyStrategy: mode.ApplyStrategy, RecordsRead: mode.RecordsRead,
				RecordsLoaded: mode.RecordsLoaded, CheckpointCommitted: mode.CheckpointCommitted,
				TargetNamespace: mode.TargetNamespace, TargetRelation: mode.TargetRelation,
			})
		}
	}
	rep.Capabilities.DeclaredTransport = transport
}
