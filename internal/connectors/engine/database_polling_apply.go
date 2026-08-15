package engine

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

// DatabasePollingApplyConfig binds a registered native polling target to its
// already-managed database target. It deliberately accepts a closed mapping,
// target control record, and shared write executor rather than SQL, a relation
// name, or a caller-selected write operation.
type DatabasePollingApplyConfig struct {
	Reference  connectors.TransportExecutorReference
	Evidence   PollingWatermarkConformanceEvidence
	Write      *database.DatabaseWriteExecutor
	Definition database.Definition
	Control    database.ManagedTargetControlRecord
	Mapping    database.MappingContractV1
	BatchSize  int
}

// DatabasePollingApplyExecutor bridges the sealed polling page boundary to
// the sealed native database session boundary. It owns neither a source
// reader nor checkpoint persistence.
type DatabasePollingApplyExecutor struct {
	reference  connectors.TransportExecutorReference
	evidence   PollingWatermarkConformanceEvidence
	write      *database.DatabaseWriteExecutor
	definition database.Definition
	control    database.ManagedTargetControlRecord
	mapping    database.MappingContractV1
	batchSize  int
}

// NewDatabasePollingApplyExecutor constructs a registered-target candidate.
// The first concrete page supplies only bounded records and explicit
// tombstones; its selected mode and strategy remain declaration-derived.
func NewDatabasePollingApplyExecutor(config DatabasePollingApplyConfig) (*DatabasePollingApplyExecutor, error) {
	if config.Reference.Validate() != nil || config.Reference.Family != connectors.TransportExecutorFamilyNativeDatabase || !config.Evidence.matchesRequired() || config.Write == nil || config.Definition.Validate() != nil || len(config.Mapping.Columns()) == 0 || config.BatchSize <= 0 {
		return nil, fmt.Errorf("database polling target configuration is invalid")
	}
	return &DatabasePollingApplyExecutor{
		reference:  config.Reference,
		evidence:   config.Evidence,
		write:      config.Write,
		definition: config.Definition,
		control:    config.Control,
		mapping:    config.Mapping,
		batchSize:  config.BatchSize,
	}, nil
}

func (e *DatabasePollingApplyExecutor) PollingApplyExecutorReference() connectors.TransportExecutorReference {
	if e == nil {
		return connectors.TransportExecutorReference{}
	}
	return e.reference
}

func (e *DatabasePollingApplyExecutor) PollingApplyConformanceEvidence() PollingWatermarkConformanceEvidence {
	if e == nil {
		return PollingWatermarkConformanceEvidence{}
	}
	return e.evidence
}

// ApplyPollingPage creates a new count-bound database plan from the
// preflight-resolved descriptor, writes the cloned page through one approved
// native transaction, and returns only the durable acknowledgement produced
// after the target receipt is recorded.
func (e *DatabasePollingApplyExecutor) ApplyPollingPage(ctx context.Context, resolved ResolvedPollingWatermark, page PollingApplyPage) (synccontract.DownstreamAcknowledgement, error) {
	if ctx == nil || e == nil || e.write == nil || resolved.Declaration == nil || resolved.Apply == nil || resolved.Apply.PollingApplyExecutorReference() != e.reference || resolved.Declaration.Target.Executor != e.reference {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("database polling target is not the resolved registered executor")
	}
	if err := ctx.Err(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := resolved.Declaration.Validate(); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("database polling target declaration: %w", err)
	}
	if err := validatePollingApplyMode(resolved.Declaration.Target, resolved.Mode); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if e.batchSize > resolved.Declaration.Target.MaxBatchRecords {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("database polling target batch size exceeds the resolved target limit")
	}
	clone, err := clonePollingApplyPage(page, resolved.Declaration.Target)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	strategy, supported := database.CanonicalDatabaseWriteStrategy(resolved.Mode)
	if !supported {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("database polling target does not support sync mode %q", resolved.Mode)
	}
	keys := databasePollingTargetKeys(resolved.Mode, resolved.Declaration.Target.StableKeyMapping)
	plan, err := database.NewDatabaseWritePlan(ctx, database.DatabaseWritePlanRequest{
		Definition:            e.definition,
		Control:               e.control,
		Mode:                  resolved.Mode,
		Strategy:              strategy,
		Mapping:               e.mapping,
		Keys:                  keys,
		RecordCount:           len(clone.Records),
		TombstoneCount:        len(clone.Tombstones),
		BatchSize:             e.batchSize,
		Destructive:           resolved.Mode == synccontract.ModeFullOverwrite,
		ConditionalOrderFence: resolved.Declaration.Target.ConditionalOrderFence,
	})
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("database polling target plan: %w", err)
	}
	records := make([]database.OrderedRecord, len(clone.Records))
	for index, record := range clone.Records {
		records[index] = database.OrderedRecord{Record: record.Record, Position: record.Position}
	}
	tombstones, err := database.NewTombstoneEnvelope(clone.Tombstones)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("database polling target tombstones: %w", err)
	}
	input, err := database.NewOrderedDatabaseWriteInput(records, tombstones)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("database polling target input: %w", err)
	}
	preview, err := e.write.Preview(ctx, plan)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	result, err := e.write.ExecuteInput(ctx, plan, approval, input)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	return result.DownstreamAcknowledgement()
}

func databasePollingTargetKeys(mode synccontract.Mode, stableKeys []string) []string {
	switch mode {
	case synccontract.ModeIncrementalUpsert, synccontract.ModeIncrementalDedupe, synccontract.ModeIncrementalDedupeHistory:
		return append([]string(nil), stableKeys...)
	default:
		return nil
	}
}

var _ PollingApplyExecutor = (*DatabasePollingApplyExecutor)(nil)
