package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

const (
	postgresPollingCertificationExecutor = "postgres_polling_watermark"
	postgresManagedCertificationExecutor = "postgres_managed_target"
)

// DeclaredTransportCertificationRequest supplies only structural
// certification context. Credentials are resolved from the existing
// root-scoped ephemeral session and are never copied into this request or a
// resulting proof.
type DeclaredTransportCertificationRequest struct {
	CertificationRoot       string
	Connector               string
	SourceCredential        string
	Stream                  string
	PrimaryKey              string
	CursorField             string
	AllowManagedTargetWrite bool
}

// ProbeDeclaredTransportForCertificationRequest executes PostgreSQL's
// definition-selected native-database pair when the definition declares that
// exact pair. Other connector families retain the existing bounded issue-label
// adapter. This is deliberately a closed PostgreSQL pair adapter, not a
// generic database framework: a future database must add its own declared
// executable proof rather than inheriting PostgreSQL assumptions.
func ProbeDeclaredTransportForCertificationRequest(ctx context.Context, request DeclaredTransportCertificationRequest) (DeclaredTransportCertificationProof, error) {
	if strings.TrimSpace(request.CertificationRoot) == "" || strings.TrimSpace(request.Connector) == "" {
		return DeclaredTransportCertificationProof{}, errors.New("declared transport certification requires root and connector")
	}
	application, err := Open(request.CertificationRoot)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("open declared transport certification project: %w", err)
	}
	registered, ok := application.registry.Get(request.Connector)
	if !ok {
		return DeclaredTransportCertificationProof{}, nil
	}
	descriptor, ok := connectors.SyncTransportDescriptorOf(registered)
	if !ok || descriptor.Source == nil || descriptor.Destination == nil {
		return ProbeDeclaredTransportForCertification(ctx, request.CertificationRoot, request.Connector)
	}
	if descriptor.Source.Executor.ID != postgresPollingCertificationExecutor || descriptor.Destination.Executor.ID != postgresManagedCertificationExecutor {
		return ProbeDeclaredTransportForCertification(ctx, request.CertificationRoot, request.Connector)
	}
	return probePostgresManagedTransportForCertification(ctx, application, *descriptor, request)
}

// postgresManagedTargetContractMode disambiguates the legacy
// incremental_append spelling only for PostgreSQL's exact declared pair. The
// public spelling maps to the contract mode when the destination is the
// managed target; ordinary local-warehouse PostgreSQL connections retain their
// legacy behavior. Other connector families cannot opt in through this path.
func postgresManagedTargetContractMode(source, destination connectors.Connector, rawMode string) bool {
	if strings.TrimSpace(rawMode) != string(synccontract.ModeIncrementalAppend) {
		return false
	}
	sourceDescriptor, sourceOK := connectors.SyncTransportDescriptorOf(source)
	destinationDescriptor, destinationOK := connectors.SyncTransportDescriptorOf(destination)
	return sourceOK && destinationOK && sourceDescriptor.Source != nil && destinationDescriptor.Destination != nil &&
		sourceDescriptor.Source.Executor.ID == postgresPollingCertificationExecutor &&
		destinationDescriptor.Destination.Executor.ID == postgresManagedCertificationExecutor
}

func probePostgresManagedTransportForCertification(ctx context.Context, application *App, descriptor connectors.SyncTransportDescriptor, request DeclaredTransportCertificationRequest) (DeclaredTransportCertificationProof, error) {
	proof := DeclaredTransportCertificationProof{
		Declared:             true,
		Applicable:           true,
		SourceReference:      descriptor.Source.Executor.ID,
		DestinationReference: descriptor.Destination.Executor.ID,
	}
	if !request.AllowManagedTargetWrite {
		proof.SkipReason = "write testing disabled; pass --write to execute the declared PostgreSQL managed target"
		return proof, nil
	}
	if strings.TrimSpace(request.SourceCredential) == "" || strings.TrimSpace(request.Stream) == "" || strings.TrimSpace(request.PrimaryKey) == "" || strings.TrimSpace(request.CursorField) == "" {
		return DeclaredTransportCertificationProof{}, errors.New("PostgreSQL transport certification requires source credential, stream, primary key, and cursor field")
	}

	strategies := make(map[synccontract.Mode]string, len(descriptor.Destination.ApplyStrategies))
	for _, strategy := range descriptor.Destination.ApplyStrategies {
		strategies[strategy.Mode] = string(strategy.Strategy)
	}
	for _, mode := range descriptor.Source.Modes {
		strategy, ok := strategies[mode]
		if !ok {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("PostgreSQL declared transport mode %q has no destination apply strategy", mode)
		}
		connection, err := application.CreateConnection(ctx, CreateConnectionRequest{
			Name:        "cert_transport_" + string(mode),
			Source:      EndpointConfig{Connector: request.Connector, Credential: request.SourceCredential},
			Destination: EndpointConfig{Connector: request.Connector, Credential: request.SourceCredential},
			Streams: map[string]StreamConfig{
				request.Stream: {
					SyncMode:         string(mode),
					PrimaryKey:       []string{request.PrimaryKey},
					CursorField:      request.CursorField,
					DestinationTable: "cert_transport_" + string(mode),
				},
			},
		})
		if err != nil {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("create PostgreSQL certification connection for %s: %w", mode, err)
		}
		plan, err := application.PlanPostgresManagedTargetTransport(ctx, connection.Name, request.Stream)
		if err != nil {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("plan PostgreSQL certification transport for %s: %w", mode, err)
		}
		plan, preview, err := application.PreviewPostgresManagedTargetTransport(ctx, plan.ID)
		if err != nil {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("preview PostgreSQL certification transport for %s: %w", mode, err)
		}
		if strings.TrimSpace(plan.ApprovalToken) == "" || strings.TrimSpace(preview.Digest) == "" {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("PostgreSQL certification transport %s did not produce approval-bound preview", mode)
		}
		run, err := application.RunETL(ctx, RunETLRequest{
			Connection: connection.Name,
			Stream:     request.Stream,
			BatchSize:  100,
			DestinationApproval: synctransport.DestinationApproval{
				PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
				Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
			},
		})
		if err != nil {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("execute PostgreSQL certification transport for %s: %w", mode, err)
		}
		if run.Status != "completed" {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("PostgreSQL certification transport %s run status = %q, want completed", mode, run.Status)
		}
		stored, err := application.GetRun(run.ID)
		if err != nil {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("read PostgreSQL certification transport run for %s: %w", mode, err)
		}
		target, err := postgresCertificationTarget(application, connection, request.Stream)
		if err != nil {
			return DeclaredTransportCertificationProof{}, fmt.Errorf("derive PostgreSQL certification target for %s: %w", mode, err)
		}
		proof.ProviderReads += stored.RecordsRead
		proof.ProviderWrites += stored.RecordsLoaded
		proof.RecordsRead += stored.RecordsRead
		proof.RecordsLoaded += stored.RecordsLoaded
		proof.CheckpointCommitted = proof.CheckpointCommitted || len(stored.Checkpoint) > 0
		proof.Modes = append(proof.Modes, DeclaredTransportModeProof{
			Mode: string(mode), ApplyStrategy: strategy, RecordsRead: stored.RecordsRead,
			RecordsLoaded: stored.RecordsLoaded, CheckpointCommitted: len(stored.Checkpoint) > 0,
			TargetNamespace: target.Namespace(), TargetRelation: target.Relation(),
		})
	}
	manifests, parquet, err := countCertificationTransportStageArtifacts(request.CertificationRoot)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("inspect PostgreSQL certification transport artifacts: %w", err)
	}
	proof.WarehouseManifests = manifests
	proof.WarehouseParquet = parquet
	return proof, nil
}

func postgresCertificationTarget(application *App, connection Connection, streamName string) (database.ManagedTargetRef, error) {
	stream, ok := connection.Streams[streamName]
	if !ok || strings.TrimSpace(stream.StreamID) == "" {
		return database.ManagedTargetRef{}, errors.New("certification transport stream has no structural identity")
	}
	identity := warehouse.ArtifactIdentity{
		WorkspaceID: application.state.WorkspaceID, ConnectorID: connection.Source.Connector, ConnectionID: connection.ID,
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		return database.ManagedTargetRef{}, err
	}
	artifact, err := warehouse.NewArtifactRef(identity, stream.StreamID)
	if err != nil {
		return database.ManagedTargetRef{}, err
	}
	return database.NewManagedTargetRef(owner, artifact, stream.StreamID)
}
