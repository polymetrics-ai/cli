package synctransport

import (
	"context"
	"errors"
	"time"

	"github.com/apache/arrow-go/v18/arrow"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// ErrArrowFastPathInvalid is a typed, pre-I/O refusal for a malformed fast
// path composition. It deliberately does not disclose a provider route,
// target identity, query, or payload.
var ErrArrowFastPathInvalid = errors.New("Arrow transport fast path is invalid")

// ArrowRangeExtractor is the optional source-side fast path. A new connector
// (including S3/Parquet) implements this port to feed the shared controller;
// it does not implement a destination-specific pipeline or expose SQL to it.
// ExtractArrowRanges synchronously transfers each record's ownership to emit:
// emit must finish consuming it before it returns, and the extractor releases
// its record only after that return.
type ArrowRangeExtractor interface {
	SourceExecutor
	ExtractArrowRanges(context.Context, ArrowExtractRequest, func(ArrowSourceBatch) error) error
}

// ArrowExtractRequest is source-neutral extraction input. Range construction,
// provider pagination, and snapshot mechanics remain behind the extractor;
// the controller owns only unit deadlines, credits, segments and checkpoints.
type ArrowExtractRequest struct {
	Connector         connectors.Connector
	Runtime           connectors.RuntimeConfig
	Stream            string
	CursorField       string
	PrimaryKey        []string
	Resume            synccontract.ResumeExpectation
	Checkpoint        *synccontract.CheckpointEnvelope
	BatchSize         int
	UnitDeadline      time.Duration
	TransformPlanJSON string
	TransformHash     string
}

// ArrowSourceBatch is one bounded typed range. SourceLogicalBytes means the
// logical source payload bytes before transform (Arrow buffer bytes in this
// slice), not parquet bytes, pgwire bytes, or target storage bytes.
type ArrowSourceBatch struct {
	Record              arrow.Record
	SourceLogicalBytes  int64
	SourceRows          int64
	ExtractElapsed      time.Duration
	CandidateCheckpoint synccontract.CheckpointEnvelope
}

// FastSegmentStore is the connector-neutral durable segment port. The App
// binds it to warehouse's versioned Parquet manifest implementation; sources
// and destinations see only its record-free receipt and cannot choose a path.
type FastSegmentStore interface {
	StoreArrowSegment(context.Context, FastSegmentWriteRequest) (FastSegmentReceipt, error)
}

type FastSegmentWriteRequest struct {
	ConnectionID       string
	Generation         int64
	Stream             string
	SegmentID          string
	TransformPlanHash  string
	SourceLogicalBytes int64
	SourceRows         int64
	Record             arrow.Record
}

// FastSegmentReceipt is an immutable handle derived by the warehouse after
// Parquet close and fsync. It contains no filesystem path, database handle,
// provider cursor, record, or connector-specific type.
type FastSegmentReceipt struct {
	ID                 string
	SchemaHash         string
	TransformPlanHash  string
	ContentSHA256      string
	ParquetSHA256      string
	SourceLogicalBytes int64
	SourceRows         int64
	TransformedRows    int64
	TransformedBytes   int64
	ParquetBytes       int64
}

// ArrowBulkDestination is the optional run-scoped binary bulk-apply port. A
// future ClickHouse or MongoDB destination implements this port and inherits
// extraction, transform, Parquet durability, byte credits, receipt sequencing
// and checkpoint behavior from synctransport unchanged.
type ArrowBulkDestination interface {
	DestinationExecutor
	BeginArrowFullOverwrite(context.Context, ArrowFullOverwriteRunRequest) (ArrowFullOverwriteRun, error)
}

type ArrowFullOverwriteRunRequest struct {
	ConnectionID      string
	Generation        int64
	Plan              DestinationPlan
	Runtime           connectors.RuntimeConfig
	Source            connectors.Connector
	SourceRuntime     connectors.RuntimeConfig
	Binding           DestinationBinding
	Stream            string
	BatchSize         int
	TransformPlanJSON string
	TransformPlanHash string
	Approval          DestinationApproval `json:"-"`
}

// ArrowFullOverwriteRun is destination-private state for one shadow-and-
// publish lifecycle. ApplyArrowSegment must use the destination's native bulk
// primitive; per-row maps/structs and generic INSERT are outside this port.
type ArrowFullOverwriteRun interface {
	ApplyArrowSegment(context.Context, ArrowBulkApplyRequest) error
	PublishArrowFullOverwrite(context.Context, ArrowFullOverwritePublicationRequest) (synccontract.DownstreamAcknowledgement, error)
	ReadBackArrowFullOverwrite(context.Context, synccontract.DownstreamAcknowledgement) error
	AbortArrowFullOverwrite(context.Context) error
}

// ArrowBulkPhaseReporter is an optional destination-neutral measurement port.
// A destination that creates a private shadow reports schema/index work here;
// the controller keeps the durable run measurement without learning a
// database driver's types or vocabulary.
type ArrowBulkPhaseReporter interface {
	ArrowBulkPhaseMeasurement() ArrowBulkPhaseMeasurement
}

type ArrowBulkPhaseMeasurement struct {
	IndexConstraintBuildElapsed time.Duration
}

type ArrowBulkApplyRequest struct {
	ConnectionID string
	Plan         DestinationPlan
	Segment      FastSegmentReceipt
	// Record is borrowed for the synchronous ApplyArrowSegment call only. The
	// controller releases the transformed record immediately after the call
	// returns, so a destination that needs to retain it must Retain and Release
	// its own reference within its private run lifecycle.
	Record        arrow.Record
	Runtime       connectors.RuntimeConfig
	Source        connectors.Connector
	SourceRuntime connectors.RuntimeConfig
	Binding       DestinationBinding
	Stream        string
	BatchSize     int
	Approval      DestinationApproval `json:"-"`
}

func cloneArrowFullOverwriteRunRequest(request ArrowFullOverwriteRunRequest) ArrowFullOverwriteRunRequest {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.SourceRuntime = cloneRuntimeConfig(request.SourceRuntime)
	clone.Binding.PrimaryKey = append([]string(nil), request.Binding.PrimaryKey...)
	return clone
}

func cloneArrowBulkApplyRequest(request ArrowBulkApplyRequest) ArrowBulkApplyRequest {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.SourceRuntime = cloneRuntimeConfig(request.SourceRuntime)
	clone.Binding.PrimaryKey = append([]string(nil), request.Binding.PrimaryKey...)
	return clone
}

// ArrowFullOverwritePublicationRequest contains aggregate, record-free proof
// after all source ranges were transformed and bulk-applied to a private
// shadow. The generic controller advances the source checkpoint only after
// the returned receipt is read back.
type ArrowFullOverwritePublicationRequest struct {
	LastCheckpoint     *synccontract.CheckpointEnvelope
	Segments           []FastSegmentReceipt
	SourceLogicalBytes int64
	SourceRows         int64
	TransformedRows    int64
	TransformedBytes   int64
}
