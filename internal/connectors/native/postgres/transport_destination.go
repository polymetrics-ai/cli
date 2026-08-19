package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

const (
	postgresManagedTargetTransportID   = "postgres_managed_target"
	postgresManagedTargetEvidenceSuite = "postgres_managed_target"
	postgresManagedTargetEvidenceRun   = "warehouse_workset_v1"
	postgresManagedTargetArtifactLimit = int64(1 << 30)
)

var (
	postgresManagedTargetTransportReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyNativeDatabase,
		ID:     postgresManagedTargetTransportID,
	}
	ErrManagedTargetTransportUnavailable      = errors.New("postgres managed target transport is unavailable")
	ErrManagedTargetTransportApprovalRequired = errors.New("postgres managed target transport approval is required")
	ErrManagedTargetTransportBindingInvalid   = errors.New("postgres managed target transport binding is invalid")
	ErrManagedTargetTransportReadBackFailed   = errors.New("postgres managed target transport read-back failed")
)

// ManagedTargetTransportDestination is PostgreSQL's closed warehouse-to-target
// adapter. It is constructed only when the embedded definition selects its
// exact native-database executor reference.
type ManagedTargetTransportDestination struct {
	connector Connector
	now       func() time.Time
}

// managedTargetHistorySourceDefinitionProvider is deliberately narrow: a
// history route must be sealed from the source connector's typed database
// declaration, not inferred from a connector name or runtime configuration.
type managedTargetHistorySourceDefinitionProvider interface {
	managedTargetHistorySourceDefinition() database.Definition
}

func ManagedTargetTransportDefinitionFactory() synctransport.DefinitionFactory {
	return synctransport.DefinitionFactory{
		Reference: postgresManagedTargetTransportReference,
		DestinationEvidence: connectors.ConformanceEvidenceReference{
			Suite: postgresManagedTargetEvidenceSuite,
			RunID: postgresManagedTargetEvidenceRun,
		},
		BuildDestination: func(connector connectors.Connector) (synctransport.DestinationExecutor, error) {
			switch typed := connector.(type) {
			case Connector:
				return &ManagedTargetTransportDestination{connector: typed}, nil
			case *Connector:
				if typed == nil {
					return nil, errors.New("postgres managed target connector is required")
				}
				return &ManagedTargetTransportDestination{connector: *typed}, nil
			default:
				return nil, errors.New("postgres managed target transport requires the native postgres connector")
			}
		},
	}
}

func (*ManagedTargetTransportDestination) TransportExecutorReference() connectors.TransportExecutorReference {
	return postgresManagedTargetTransportReference
}

func (*ManagedTargetTransportDestination) ManagedTargetApprovalDestination() {}

func (d *ManagedTargetTransportDestination) PlanDestination(ctx context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	if ctx == nil || d == nil || request.Connector == nil || request.Connector.Name() != d.connector.Name() {
		return synctransport.DestinationPlan{}, ErrManagedTargetTransportUnavailable
	}
	if err := ctx.Err(); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if err := authorizeManagedTargetTransportApproval(request.Approval); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if err := validateManagedTargetTransportBinding(request.Binding); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if request.BatchSize <= 0 || strings.TrimSpace(request.Stream) == "" {
		return synctransport.DestinationPlan{}, ErrManagedTargetTransportBindingInvalid
	}
	strategy, ok := database.CanonicalDatabaseWriteStrategy(request.Mode)
	if !ok || strategy != request.ApplyStrategy.Strategy {
		return synctransport.DestinationPlan{}, ErrManagedTargetTransportBindingInvalid
	}
	if (strings.TrimSpace(request.TransformPlanJSON) == "") != (strings.TrimSpace(request.TransformPlanHash) == "") {
		return synctransport.DestinationPlan{}, ErrManagedTargetTransportBindingInvalid
	}
	if request.TransformPlanHash != "" {
		plan, err := database.ParseTransformPlanV1([]byte(request.TransformPlanJSON))
		if err != nil || plan.Hash() != request.TransformPlanHash {
			return synctransport.DestinationPlan{}, ErrManagedTargetTransportBindingInvalid
		}
	}
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy, TransformPlanHash: request.TransformPlanHash}, nil
}

func (d *ManagedTargetTransportDestination) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if ctx == nil || d == nil {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportUnavailable
	}
	if err := ctx.Err(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := validateManagedTargetTransportApproval(request.Approval); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := validateManagedTargetTransportApprovalAt(request.Approval, d.currentTime()); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := validateManagedTargetApplyRequest(request); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	historyRoute, err := d.historyRoute(request.Mode, request.Source)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	resolved, err := d.resolveManagedTarget(ctx, request.Source, request.SourceRuntime, request.Runtime, request.Binding, request.Stream)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	defer resolved.close()
	if len(request.Workset.Records) == 0 && len(request.Workset.Tombstones) == 0 {
		return synccontract.NewDurableDownstreamAcknowledgement(d.connector.Name(), request.Workset.CandidateCheckpoint.ObservedAt)
	}
	control, err := resolved.provision(ctx)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	ledger, err := database.NewManagedTargetDeliveryLedger(resolved.driver)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportUnavailable
	}
	write, err := database.NewDatabaseWriteExecutor(resolved.driver, ledger)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, ErrManagedTargetTransportUnavailable
	}

	if request.Mode == synccontract.ModeIncrementalUpsert {
		return d.applyChangeDelivery(ctx, request, resolved, control, write)
	}
	return d.applyDatabaseWrite(ctx, request, resolved, control, write, historyRoute)
}

func (d *ManagedTargetTransportDestination) historyRoute(mode synccontract.Mode, source connectors.Connector) (database.DatabaseWriteHistoryRoute, error) {
	if mode != synccontract.ModeIncrementalDedupeHistory {
		return database.DatabaseWriteHistoryRoute{}, nil
	}
	provider, ok := source.(managedTargetHistorySourceDefinitionProvider)
	if !ok {
		return database.DatabaseWriteHistoryRoute{}, &database.DatabaseWriteHistoryRouteError{Reason: database.DatabaseWriteHistoryRouteSourceUnsupported}
	}
	return database.DatabaseWriteHistoryRoute{
		Source:      provider.managedTargetHistorySourceDefinition().Driver(),
		Destination: d.connector.databaseDefinition.Driver(),
	}, nil
}

func (d *ManagedTargetTransportDestination) currentTime() time.Time {
	if d != nil && d.now != nil {
		return d.now().UTC()
	}
	return time.Now().UTC()
}

func (d *ManagedTargetTransportDestination) applyChangeDelivery(ctx context.Context, request synctransport.DestinationApplyRequest, resolved managedTargetTransportResolution, control database.ManagedTargetControlRecord, write *database.DatabaseWriteExecutor) (synccontract.DownstreamAcknowledgement, error) {
	baselineRoot, err := managedTargetBaselineWindowRoot(resolved.baselineRoot, request.Binding.PrimaryKey, request.Workset)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	baselineStore, err := database.NewFileChangeDeliveryBaselineStore(baselineRoot, postgresManagedTargetArtifactLimit)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	baselinePath, cleanup, err := managedTargetBaselineInput(ctx, baselineStore, control, resolved.worksetRoot)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	defer cleanup()
	workset, err := database.DeriveChangeDeliveryWorkset(ctx, database.ChangeDeliveryWorksetRequest{
		Control:          control,
		Keys:             request.Binding.PrimaryKey,
		SourceParquet:    request.Workset.SourceParquet,
		BaselineParquet:  baselinePath,
		Tombstones:       request.Workset.Tombstones,
		Root:             resolved.worksetRoot,
		MaxArtifactBytes: postgresManagedTargetArtifactLimit,
	})
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if workset.Changes() == 0 && workset.TombstoneCount() == 0 {
		return synccontract.NewDurableDownstreamAcknowledgement(d.connector.Name(), request.Workset.CandidateCheckpoint.ObservedAt)
	}
	delivery, err := database.NewChangeDeliveryExecutor(write, baselineStore)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	plan, err := database.NewChangeDeliveryPlan(ctx, database.ChangeDeliveryPlanRequest{
		Definition: d.connector.databaseDefinition,
		Workset:    workset,
		Control:    control,
		Mapping:    resolved.mapping,
		BatchSize:  request.BatchSize,
	})
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	preview, err := delivery.Preview(ctx, plan)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	approval, err := database.NewChangeDeliveryApproval(preview)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	result, err := delivery.Execute(ctx, plan, approval)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	ack, err := result.DownstreamAcknowledgement()
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	return synccontract.NewDurableDownstreamAcknowledgement(d.connector.Name(), ack.AcknowledgedAt)
}

func (d *ManagedTargetTransportDestination) applyDatabaseWrite(ctx context.Context, request synctransport.DestinationApplyRequest, resolved managedTargetTransportResolution, control database.ManagedTargetControlRecord, write *database.DatabaseWriteExecutor, historyRoute database.DatabaseWriteHistoryRoute) (synccontract.DownstreamAcknowledgement, error) {
	plan, err := database.NewDatabaseWritePlan(ctx, database.DatabaseWritePlanRequest{
		Definition:     d.connector.databaseDefinition,
		Control:        control,
		Mode:           request.Mode,
		Strategy:       request.Plan.ApplyStrategy.Strategy,
		Mapping:        resolved.mapping,
		Keys:           managedTargetWriteKeys(request.Mode, request.Binding.PrimaryKey),
		RecordCount:    len(request.Workset.Records),
		TombstoneCount: len(request.Workset.Tombstones),
		BatchSize:      request.BatchSize,
		Destructive:    request.Mode == synccontract.ModeFullOverwrite,
		HistoryRoute:   historyRoute,
		// The durable workset retains the source page's candidate checkpoint.
		// PostgreSQL polling advances that tuple only after this apply/read-back
		// sequence completes, so it is the source-owned ordering fence for every
		// key observed in this bounded page.
		ConditionalOrderFence: request.Mode == synccontract.ModeIncrementalDedupeHistory,
	})
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	preview, err := write.Preview(ctx, plan)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	tombstones, err := database.NewTombstoneEnvelope(request.Workset.Tombstones)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	input, err := managedTargetDatabaseWriteInput(request, tombstones)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	result, err := write.ExecuteInput(ctx, plan, approval, input)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	ack, err := result.DownstreamAcknowledgement()
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	return synccontract.NewDurableDownstreamAcknowledgement(d.connector.Name(), ack.AcknowledgedAt)
}

func managedTargetDatabaseWriteInput(request synctransport.DestinationApplyRequest, tombstones database.TombstoneEnvelope) (database.DatabaseWriteInput, error) {
	if request.Mode != synccontract.ModeIncrementalDedupeHistory {
		return database.NewDatabaseWriteInput(request.Workset.Records, tombstones)
	}
	position := request.Workset.CandidateCheckpoint.Position.Clone()
	records := make([]database.OrderedRecord, len(request.Workset.Records))
	for index, record := range request.Workset.Records {
		records[index] = database.OrderedRecord{Record: record, Position: position.Clone()}
	}
	return database.NewOrderedDatabaseWriteInput(records, tombstones)
}

func (d *ManagedTargetTransportDestination) ReadBackDestination(ctx context.Context, request synctransport.DestinationReadBackRequest) error {
	if ctx == nil || d == nil || request.Acknowledgement.Sink != d.connector.Name() || request.Acknowledgement.AcknowledgedAt.IsZero() {
		return ErrManagedTargetTransportReadBackFailed
	}
	if len(request.Workset.Records) == 0 && len(request.Workset.Tombstones) == 0 {
		return nil
	}
	resolved, err := d.resolveManagedTarget(ctx, request.Source, request.SourceRuntime, request.Runtime, request.Binding, request.Stream)
	if err != nil {
		return err
	}
	defer resolved.close()
	control, err := resolved.provision(ctx)
	if err != nil {
		return err
	}
	key, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		return ErrManagedTargetTransportReadBackFailed
	}
	ledger, err := database.NewManagedTargetDeliveryLedger(resolved.driver)
	if err != nil {
		return ErrManagedTargetTransportReadBackFailed
	}
	_, found, err := ledger.Lookup(ctx, key)
	if err != nil || !found {
		return ErrManagedTargetTransportReadBackFailed
	}
	if request.Mode == synccontract.ModeIncrementalUpsert {
		baselineRoot, err := managedTargetBaselineWindowRoot(resolved.baselineRoot, request.Binding.PrimaryKey, request.Workset)
		if err != nil {
			return ErrManagedTargetTransportReadBackFailed
		}
		store, err := database.NewFileChangeDeliveryBaselineStore(baselineRoot, postgresManagedTargetArtifactLimit)
		if err != nil {
			return ErrManagedTargetTransportReadBackFailed
		}
		if _, found, err := store.Lookup(ctx, key); err != nil || !found {
			return ErrManagedTargetTransportReadBackFailed
		}
	}
	return nil
}

type managedTargetTransportResolution struct {
	conn         *pgx.Conn
	driver       *DatabaseDriver
	owner        database.TargetOwner
	target       database.ManagedTargetRef
	targetDB     database.TargetDatabaseIdentity
	schema       database.ManagedTargetSchema
	mapping      database.MappingContractV1
	worksetRoot  string
	baselineRoot string
}

func (r managedTargetTransportResolution) close() {
	if r.conn != nil {
		_ = r.conn.Close(context.Background())
	}
}

func (r managedTargetTransportResolution) provision(ctx context.Context) (database.ManagedTargetControlRecord, error) {
	plan, err := database.NewManagedTargetProvisioningPlan(r.owner, r.target, r.targetDB, r.schema, r.mapping)
	if err != nil {
		return database.ManagedTargetControlRecord{}, err
	}
	provisioner, err := database.NewManagedTargetProvisioner(r.driver)
	if err != nil {
		return database.ManagedTargetControlRecord{}, err
	}
	return provisioner.CreateOrAssert(ctx, plan)
}

func (d *ManagedTargetTransportDestination) resolveManagedTarget(ctx context.Context, source connectors.Connector, sourceRuntime, targetRuntime connectors.RuntimeConfig, binding synctransport.DestinationBinding, stream string) (managedTargetTransportResolution, error) {
	return d.resolveManagedTargetWithTransform(ctx, source, sourceRuntime, targetRuntime, binding, stream, nil)
}

// resolveManagedTargetWithTransform provisions the output relation of a
// sealed TransformPlanV1 when the Arrow fast path is selected. The transform
// changes only the managed target's typed schema/mapping; source discovery,
// ownership and all endpoint construction remain unchanged.
func (d *ManagedTargetTransportDestination) resolveManagedTargetWithTransform(ctx context.Context, source connectors.Connector, sourceRuntime, targetRuntime connectors.RuntimeConfig, binding synctransport.DestinationBinding, stream string, transform *database.TransformPlanV1) (managedTargetTransportResolution, error) {
	if err := validateManagedTargetTransportBinding(binding); err != nil {
		return managedTargetTransportResolution{}, err
	}
	catalog, err := database.CatalogForManagedTargetSource(ctx, source, sourceRuntime, stream)
	if err != nil {
		return managedTargetTransportResolution{}, err
	}
	relation := catalog.Relations()[0]
	mapping, err := managedTargetMapping(relation)
	if err != nil {
		return managedTargetTransportResolution{}, err
	}
	schemaRelation := relation
	if transform != nil {
		if err := transform.ValidateAgainstRelation(relation); err != nil {
			return managedTargetTransportResolution{}, database.ErrTransformPlanInvalid
		}
		mapping, err = transform.OutputMapping()
		if err != nil {
			return managedTargetTransportResolution{}, database.ErrTransformPlanInvalid
		}
		schemaRelation, err = transform.OutputRelation(relation)
		if err != nil {
			return managedTargetTransportResolution{}, database.ErrTransformPlanInvalid
		}
	}
	relationCatalog, err := database.NewCatalog(catalog.Ref(), []database.Relation{schemaRelation})
	if err != nil {
		return managedTargetTransportResolution{}, err
	}
	schema, err := database.NewManagedTargetSchema(1, relationCatalog.Fingerprint())
	if err != nil {
		return managedTargetTransportResolution{}, err
	}
	identity := warehouse.ArtifactIdentity{WorkspaceID: binding.WorkspaceID, ConnectorID: binding.SourceConnectorID, ConnectionID: binding.ConnectionID}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		return managedTargetTransportResolution{}, ErrManagedTargetTransportBindingInvalid
	}
	artifact, err := warehouse.NewArtifactRef(identity, binding.StreamID)
	if err != nil {
		return managedTargetTransportResolution{}, ErrManagedTargetTransportBindingInvalid
	}
	target, err := database.NewManagedTargetRef(owner, artifact, binding.StreamID)
	if err != nil {
		return managedTargetTransportResolution{}, ErrManagedTargetTransportBindingInvalid
	}
	config, err := resolveConfig(targetRuntime)
	if err != nil {
		return managedTargetTransportResolution{}, err
	}
	pgxConfig, err := config.dataConfig()
	if err != nil {
		return managedTargetTransportResolution{}, err
	}
	conn, err := pgx.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		return managedTargetTransportResolution{}, fmt.Errorf("%w: connect postgres managed target failed", ErrManagedTargetTransportUnavailable)
	}
	driver, err := NewDatabaseDriver(conn)
	if err != nil {
		_ = conn.Close(context.Background())
		return managedTargetTransportResolution{}, ErrManagedTargetTransportUnavailable
	}
	observation, err := driver.ObserveManagedTarget(ctx, target)
	if err != nil {
		_ = conn.Close(context.Background())
		return managedTargetTransportResolution{}, err
	}
	root := targetRuntime.ProjectDir
	if strings.TrimSpace(root) == "" {
		_ = conn.Close(context.Background())
		return managedTargetTransportResolution{}, ErrManagedTargetTransportBindingInvalid
	}
	return managedTargetTransportResolution{
		conn: conn, driver: driver, owner: owner, target: target,
		targetDB: observation.TargetDatabase, schema: schema, mapping: mapping,
		worksetRoot:  filepath.Join(root, "postgres-worksets"),
		baselineRoot: filepath.Join(root, "postgres-baselines"),
	}, nil
}

func managedTargetMapping(relation database.Relation) (database.MappingContractV1, error) {
	columns := append([]database.Column(nil), relation.Columns...)
	sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
	mapped := make([]database.MappingColumnV1, 0, len(columns))
	for _, column := range columns {
		plan, err := database.CompileTypePlan(column.Type, column.Type)
		if err != nil {
			return database.MappingContractV1{}, err
		}
		mapped = append(mapped, database.MappingColumnV1{Source: column.Ref.Name, Target: column.Ref.Name, Type: plan, Nullable: column.Nullable})
	}
	return database.NewMappingContractV1(mapped)
}

func managedTargetBaselineInput(ctx context.Context, store *database.FileChangeDeliveryBaselineStore, control database.ManagedTargetControlRecord, root string) (string, func(), error) {
	key, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		return "", func() {}, err
	}
	baseline, found, err := store.Lookup(ctx, key)
	if err != nil {
		return "", func() {}, err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return "", func() {}, err
	}
	dir, err := os.MkdirTemp(root, ".baseline-input-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	path := filepath.Join(dir, "baseline.parquet")
	rows := make([]warehouse.Row, 0)
	if found {
		if err := baseline.ReadCandidateBaseline(ctx, func(row warehouse.Row) error {
			rows = append(rows, row)
			return nil
		}); err != nil {
			cleanup()
			return "", func() {}, err
		}
	}
	if err := warehouse.WriteTable(ctx, path, rows); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return path, cleanup, nil
}

func managedTargetWriteKeys(mode synccontract.Mode, keys []string) []string {
	if mode == synccontract.ModeIncrementalUpsert || mode == synccontract.ModeIncrementalDedupe || mode == synccontract.ModeIncrementalDedupeHistory {
		return append([]string(nil), keys...)
	}
	return nil
}

func managedTargetBaselineWindowRoot(root string, primaryKey []string, workset synctransport.WarehouseWorkset) (string, error) {
	if strings.TrimSpace(root) == "" || len(primaryKey) == 0 {
		return "", ErrManagedTargetTransportBindingInvalid
	}
	identities := make([]string, 0, len(workset.Records)+len(workset.Tombstones))
	for _, record := range workset.Records {
		values := make([]any, len(primaryKey))
		for index, key := range primaryKey {
			value, ok := record[key]
			if !ok || value == nil {
				return "", ErrManagedTargetTransportBindingInvalid
			}
			values[index] = value
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			return "", ErrManagedTargetTransportBindingInvalid
		}
		identities = append(identities, "record:"+string(encoded))
	}
	for _, tombstone := range workset.Tombstones {
		var key any
		if err := json.Unmarshal(tombstone.Key, &key); err != nil {
			return "", ErrManagedTargetTransportBindingInvalid
		}
		encoded, err := json.Marshal(key)
		if err != nil {
			return "", ErrManagedTargetTransportBindingInvalid
		}
		identities = append(identities, "tombstone:"+string(encoded))
	}
	sort.Strings(identities)
	hash := sha256.New()
	for _, identity := range identities {
		_, _ = hash.Write([]byte(identity))
		_, _ = hash.Write([]byte{0})
	}
	return filepath.Join(root, "windows", hex.EncodeToString(hash.Sum(nil))), nil
}

func validateManagedTargetTransportApproval(approval synctransport.DestinationApproval) error {
	if strings.TrimSpace(approval.PlanID) == "" || approval.Confirmation.Kind != connectors.ConfirmationKindDestructive || approval.Evidence == nil || strings.TrimSpace(approval.PreviewDigest) == "" {
		return ErrManagedTargetTransportApprovalRequired
	}
	return nil
}

func authorizeManagedTargetTransportApproval(approval synctransport.DestinationApproval) error {
	if err := validateManagedTargetTransportApproval(approval); err != nil {
		return err
	}
	if err := approval.Evidence.Authorize(approval.Target, approval.PreviewDigest, time.Now().UTC()); err != nil {
		return fmt.Errorf("%w: %v", ErrManagedTargetTransportApprovalRequired, err)
	}
	return nil
}

func validateManagedTargetTransportApprovalAt(approval synctransport.DestinationApproval, now time.Time) error {
	if err := validateManagedTargetTransportApproval(approval); err != nil {
		return err
	}
	if err := approval.Evidence.Validate(approval.Target, approval.PreviewDigest, now); err != nil {
		return fmt.Errorf("%w: %v", ErrManagedTargetTransportApprovalRequired, err)
	}
	return nil
}

func validateManagedTargetTransportBinding(binding synctransport.DestinationBinding) error {
	identity := warehouse.ArtifactIdentity{WorkspaceID: binding.WorkspaceID, ConnectorID: binding.SourceConnectorID, ConnectionID: binding.ConnectionID}
	if identity.Validate() != nil || !warehouse.SafePathPart(binding.StreamID) {
		return ErrManagedTargetTransportBindingInvalid
	}
	return nil
}

func validateManagedTargetApplyRequest(request synctransport.DestinationApplyRequest) error {
	if err := validateManagedTargetTransportBinding(request.Binding); err != nil {
		return err
	}
	if request.ConnectionID != request.Binding.ConnectionID || request.Workset.ID != request.Receipt.ID || request.Receipt.Owner != request.Binding.ConnectionID || request.Workset.SourceParquet == "" || request.BatchSize <= 0 || request.Mode != request.Plan.ApplyStrategy.Mode {
		return ErrManagedTargetTransportBindingInvalid
	}
	return nil
}

var _ synctransport.DestinationExecutor = (*ManagedTargetTransportDestination)(nil)
