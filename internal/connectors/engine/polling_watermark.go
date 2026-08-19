package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
)

// PollingWatermarkPosition is the complete provider ordering tuple persisted
// after a page is durably accepted. Watermark and TieBreaker retain the
// provider's lossless textual representation; they are not float-coerced.
type PollingWatermarkPosition struct {
	Watermark        string
	TieBreaker       string
	tieBreakerNumber bool
}

const pollingWatermarkScanStateKey = "$polymetrics.polling_watermark.scan.v1"
const pollingWatermarkFrontiersStateKey = "$polymetrics.polling_watermark.frontiers.v1"

type pollingWatermarkScanCheckpoint struct {
	Version          int    `json:"version"`
	Watermark        string `json:"watermark"`
	TieBreaker       string `json:"tie_breaker"`
	TieBreakerNumber bool   `json:"tie_breaker_numeric"`
	Active           bool   `json:"active"`
}

type pollingWatermarkPositionCheckpoint struct {
	Watermark        string `json:"watermark"`
	TieBreaker       string `json:"tie_breaker"`
	TieBreakerNumber bool   `json:"tie_breaker_numeric"`
}

type pollingWatermarkSourceCheckpoint struct {
	Watermark        string                              `json:"watermark"`
	TieBreaker       string                              `json:"tie_breaker"`
	TieBreakerNumber bool                                `json:"tie_breaker_numeric"`
	QuerySet         bool                                `json:"query_set,omitempty"`
	Query            *pollingWatermarkPositionCheckpoint `json:"query,omitempty"`
	Scan             *pollingWatermarkScanCheckpoint     `json:"scan,omitempty"`
}

type pollingWatermarkFrontiersCheckpoint struct {
	Version                 int                               `json:"version"`
	PrimaryTieBreakerNumber bool                              `json:"primary_tie_breaker_numeric"`
	Deletion                *pollingWatermarkSourceCheckpoint `json:"deletion,omitempty"`
}

type pollingWatermarkSourceFrontier struct {
	durable  *PollingWatermarkPosition
	query    *PollingWatermarkPosition
	querySet bool
	scan     *pollingWatermarkScanCheckpoint
}

type pollingWatermarkFrontiers struct {
	primary  pollingWatermarkSourceFrontier
	deletion pollingWatermarkSourceFrontier
}

// PollingWatermarkPageRequest is the closed request passed to a connector
// transport adapter. After and DeletionAfter are inclusive by contract: the
// edge record is read again so timestamp ties cannot be silently skipped.
// DeletionEndpoint is copied from the declaration when a provider exposes a
// separate deletion feed; transport code never receives caller-authored
// endpoint structure.
type PollingWatermarkPageRequest struct {
	Stream           string
	Config           connectors.RuntimeConfig
	After            *PollingWatermarkPosition
	DeletionAfter    *PollingWatermarkPosition
	Inclusive        bool
	PageSize         int
	DeletionEndpoint *connectors.PollingWatermarkDeletionEndpoint
	RequestBudget    PollingWatermarkRequestBudget
}

// PollingWatermarkPage is one bounded source response. DeletionRecords are
// supplied only for a declared deletion endpoint; all other hard deletes are
// intentionally invisible to polling. More is a lockstep continuation signal
// for every configured source. A source with independently paginated primary
// and deletion reads must set PrimaryMore and DeletionMore instead.
type PollingWatermarkPage struct {
	Records         []connectors.Record
	DeletionRecords []connectors.Record
	More            bool
	PrimaryMore     bool
	DeletionMore    bool
}

type PollingWatermarkRequestBudget interface {
	Consume(context.Context) error
	Remaining() int
}

type pollingWatermarkRequestBudget struct {
	mu    sync.Mutex
	limit int
	used  int
}

type pollingWatermarkRequestBudgetExhaustedError struct {
	Limit int
	Used  int
}

func (e *pollingWatermarkRequestBudgetExhaustedError) Error() string {
	return fmt.Sprintf("polling watermark request budget exhausted after %d of %d requests", e.Used, e.Limit)
}

func newPollingWatermarkRequestBudget(limit int) *pollingWatermarkRequestBudget {
	return &pollingWatermarkRequestBudget{limit: limit}
}

func (b *pollingWatermarkRequestBudget) Consume(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.used >= b.limit {
		return &pollingWatermarkRequestBudgetExhaustedError{Limit: b.limit, Used: b.used}
	}
	b.used++
	return nil
}

func (b *pollingWatermarkRequestBudget) Remaining() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.limit - b.used
}

func (b *pollingWatermarkRequestBudget) Used() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.used
}

type PollingWatermarkStopReason string

const (
	PollingWatermarkStopReasonMaxPages      PollingWatermarkStopReason = "max_pages"
	PollingWatermarkStopReasonRequestBudget PollingWatermarkStopReason = "request_budget"
)

type PollingWatermarkResumableStopError struct {
	Reason              PollingWatermarkStopReason
	LastDurablePosition *PollingWatermarkPosition
	PagesRead           int
	RequestsUsed        int
}

func (e *PollingWatermarkResumableStopError) Error() string {
	return fmt.Sprintf("polling watermark %s limit reached; resume from the last durable checkpoint", e.Reason)
}

func (*PollingWatermarkResumableStopError) Resumable() bool {
	return true
}

type PollingWatermarkNonAdvancingReason string

const (
	PollingWatermarkNonAdvancingReasonPageSize PollingWatermarkNonAdvancingReason = "page_size"
	PollingWatermarkNonAdvancingReasonCursor   PollingWatermarkNonAdvancingReason = "cursor"
)

type PollingWatermarkNonAdvancingError struct {
	Reason   PollingWatermarkNonAdvancingReason
	Page     int
	Source   string
	Limit    int
	Returned int
}

func (e *PollingWatermarkNonAdvancingError) Error() string {
	if e.Reason == PollingWatermarkNonAdvancingReasonPageSize {
		return fmt.Sprintf("polling watermark %s page %d returned %d records exceeding declared page size %d without advancing state", e.Source, e.Page, e.Returned, e.Limit)
	}
	return fmt.Sprintf("polling watermark page %d did not advance the source cursor", e.Page)
}

func (*PollingWatermarkNonAdvancingError) NonAdvancing() bool {
	return true
}

// PollingWatermarkPageSource is the transport seam for the shared executor.
// It must implement only declaration-selected read operations; it never
// receives raw SQL, arbitrary HTTP, shell, or caller-selected paths. It must
// consume req.RequestBudget before every physical declared-source request.
type PollingWatermarkPageSource interface {
	FetchPollingWatermarkPage(ctx context.Context, req PollingWatermarkPageRequest) (PollingWatermarkPage, error)
}

// PollingWatermarkExecutor executes the shared, declaration-driven polling
// contract. It is intentionally a consumer of checkpoint persistence: the
// caller supplies the committed state and durable committer, so #3810's
// versioned checkpoint envelope can replace that adapter without a transport
// or delivery rewrite.
type PollingWatermarkExecutor struct {
	descriptor *connectors.ChangefeedDescriptor
	source     PollingWatermarkPageSource
}

var _ connectors.ChangefeedExecutor = (*PollingWatermarkExecutor)(nil)

// NewPollingWatermarkExecutor constructs the shared executor only for a fully
// validated polling-watermark declaration and a closed source adapter.
func NewPollingWatermarkExecutor(descriptor *connectors.ChangefeedDescriptor, source PollingWatermarkPageSource) (*PollingWatermarkExecutor, error) {
	if descriptor == nil {
		return nil, errors.New("polling watermark executor requires a changefeed declaration")
	}
	if err := descriptor.Validate(); err != nil {
		return nil, fmt.Errorf("polling watermark declaration: %w", err)
	}
	if descriptor.Mechanism != connectors.ChangefeedMechanismPollingWatermark || !descriptor.IsImplemented() {
		return nil, errors.New("polling watermark executor requires an implemented polling_watermark declaration")
	}
	if source == nil {
		return nil, errors.New("polling watermark executor requires a page source")
	}
	return &PollingWatermarkExecutor{descriptor: descriptor.Clone(), source: source}, nil
}

// ChangefeedExecutorDescriptor supplies the runtime half of the existing
// fail-closed promotion gate. Provider evidence remains declaration-owned.
func (e *PollingWatermarkExecutor) ChangefeedExecutorDescriptor() connectors.ChangefeedExecutorDescriptor {
	if e == nil || e.descriptor == nil || e.descriptor.Executor == nil || e.descriptor.Checkpoint == nil {
		return connectors.ChangefeedExecutorDescriptor{}
	}
	checkpoint := *e.descriptor.Checkpoint
	checkpoint.Keys = append([]string(nil), e.descriptor.Checkpoint.Keys...)
	return connectors.ChangefeedExecutorDescriptor{
		Status:     connectors.ChangefeedStatusImplemented,
		Mechanism:  connectors.ChangefeedMechanismPollingWatermark,
		Executor:   *e.descriptor.Executor,
		Checkpoint: checkpoint,
	}
}

// ReadCDC reads bounded pages, waits for every emitted event callback to
// succeed, and only then commits the page's ordering tuple. A failure after
// destination acknowledgement but before CommitChangefeedCheckpoint returns an
// error with the prior state still committed, so the next run replays rather
// than skips the page.
func (e *PollingWatermarkExecutor) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	if e == nil || e.descriptor == nil || e.descriptor.PollingWatermark == nil {
		return errors.New("polling watermark executor is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if req.CheckpointCommitter == nil {
		return errors.New("polling watermark changefeed requires a checkpoint committer")
	}
	if emit == nil {
		return errors.New("polling watermark changefeed requires an event callback")
	}
	if !containsDeclaredStream(e.descriptor.Streams, req.Stream) {
		return fmt.Errorf("polling watermark stream %q is not declared", req.Stream)
	}

	polling := e.descriptor.PollingWatermark
	frontiers, err := e.initialFrontiers(req.State)
	if err != nil {
		return err
	}
	requestBudget := newPollingWatermarkRequestBudget(polling.RequestBudget)
	checkpointState := clonePollingWatermarkState(req.State)
	requestsPerPage := 1
	if polling.DeletionEndpoint != nil {
		requestsPerPage++
	}

	for pageNumber := 0; pageNumber < polling.MaxPages; pageNumber++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if requestBudget.Remaining() < requestsPerPage {
			return newPollingWatermarkResumableStop(PollingWatermarkStopReasonRequestBudget, frontiers.primary.durable, pageNumber, requestBudget.Used())
		}
		requestsBefore := requestBudget.Used()
		page, err := e.source.FetchPollingWatermarkPage(ctx, PollingWatermarkPageRequest{
			Stream:           req.Stream,
			Config:           req.Config,
			After:            clonePollingWatermarkPosition(frontiers.primary.query),
			DeletionAfter:    clonePollingWatermarkPosition(frontiers.deletion.query),
			Inclusive:        true,
			PageSize:         polling.PageSize,
			DeletionEndpoint: clonePollingWatermarkDeletionEndpoint(polling.DeletionEndpoint),
			RequestBudget:    requestBudget,
		})
		if err != nil {
			var exhausted *pollingWatermarkRequestBudgetExhaustedError
			if errors.As(err, &exhausted) {
				return newPollingWatermarkResumableStop(PollingWatermarkStopReasonRequestBudget, frontiers.primary.durable, pageNumber, requestBudget.Used())
			}
			return fmt.Errorf("polling watermark fetch page %d: %w", pageNumber+1, err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := e.validatePhysicalPage(pageNumber+1, page); err != nil {
			return err
		}
		if len(page.DeletionRecords) > 0 && polling.DeletionEndpoint == nil {
			return errors.New("polling watermark source returned deletion records without a declared deletion_endpoint")
		}
		if requestBudget.Used()-requestsBefore < requestsPerPage {
			return errors.New("polling watermark source returned without consuming the request budget for every declared-source request")
		}
		primaryMore, deletionMore := page.continuation(polling.DeletionEndpoint != nil)
		primary, err := e.positionedPollingWatermarkRecords(page.Records, false)
		if err != nil {
			return fmt.Errorf("polling watermark page %d: %w", pageNumber+1, err)
		}
		deletions, err := e.positionedPollingWatermarkRecords(page.DeletionRecords, true)
		if err != nil {
			return fmt.Errorf("polling watermark page %d: %w", pageNumber+1, err)
		}
		if err := e.requirePollingWatermarkSourceProgress(frontiers.primary.query, primary, primaryMore, pageNumber+1, "primary"); err != nil {
			return fmt.Errorf("polling watermark page %d: %w", pageNumber+1, err)
		}
		if err := e.requirePollingWatermarkSourceProgress(frontiers.deletion.query, deletions, deletionMore, pageNumber+1, "deletion"); err != nil {
			return fmt.Errorf("polling watermark page %d: %w", pageNumber+1, err)
		}
		positioned, err := e.mergePollingWatermarkPageRecords(primary, deletions, primaryMore, deletionMore)
		if err != nil {
			return fmt.Errorf("polling watermark page %d: %w", pageNumber+1, err)
		}
		for _, record := range positioned {
			if err := e.emitPollingRecord(ctx, record.record, record.deletionRecord, emit); err != nil {
				return err
			}
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		primaryAccepted, deletionAccepted := splitPollingWatermarkRecords(positioned)
		nextPrimary, primaryChanged, err := e.advancePollingWatermarkSource(frontiers.primary, primaryAccepted, primaryMore, len(primary))
		if err != nil {
			return fmt.Errorf("polling watermark page %d checkpoint: %w", pageNumber+1, err)
		}
		nextDeletion, deletionChanged, err := e.advancePollingWatermarkSource(frontiers.deletion, deletionAccepted, deletionMore, len(deletions))
		if err != nil {
			return fmt.Errorf("polling watermark page %d checkpoint: %w", pageNumber+1, err)
		}
		nextFrontiers := pollingWatermarkFrontiers{primary: nextPrimary, deletion: nextDeletion}
		if len(positioned) > 0 || primaryChanged || deletionChanged {
			nextState, err := e.checkpointState(checkpointState, nextFrontiers)
			if err != nil {
				return fmt.Errorf("polling watermark page %d checkpoint: %w", pageNumber+1, err)
			}
			if err := req.CheckpointCommitter.CommitChangefeedCheckpoint(ctx, nextState); err != nil {
				return fmt.Errorf("polling watermark commit checkpoint after destination acknowledgement: %w", err)
			}
			checkpointState = nextState
		}
		frontiers = nextFrontiers
		if !primaryMore && !deletionMore {
			return nil
		}
	}
	return newPollingWatermarkResumableStop(PollingWatermarkStopReasonMaxPages, frontiers.primary.durable, polling.MaxPages, requestBudget.Used())
}

func newPollingWatermarkResumableStop(reason PollingWatermarkStopReason, durable *PollingWatermarkPosition, pagesRead, requestsUsed int) *PollingWatermarkResumableStopError {
	return &PollingWatermarkResumableStopError{
		Reason:              reason,
		LastDurablePosition: clonePollingWatermarkPosition(durable),
		PagesRead:           pagesRead,
		RequestsUsed:        requestsUsed,
	}
}

func (e *PollingWatermarkExecutor) initialFrontiers(state map[string]string) (pollingWatermarkFrontiers, error) {
	keys := e.descriptor.Checkpoint.Keys
	if len(keys) != 2 {
		return pollingWatermarkFrontiers{}, errors.New("polling watermark executor requires exactly two checkpoint keys")
	}
	scanCheckpoint, err := e.scanCheckpointFromState(state)
	if err != nil {
		return pollingWatermarkFrontiers{}, err
	}
	persisted, err := e.frontiersCheckpointFromState(state)
	if err != nil {
		return pollingWatermarkFrontiers{}, err
	}
	watermark, hasWatermark := state[keys[0]]
	tieBreaker, hasTieBreaker := state[keys[1]]
	if hasWatermark != hasTieBreaker {
		return pollingWatermarkFrontiers{}, errors.New("polling watermark checkpoint is incomplete")
	}
	var primaryDurable *PollingWatermarkPosition
	if hasWatermark {
		primaryDurable = &PollingWatermarkPosition{
			Watermark:        watermark,
			TieBreaker:       tieBreaker,
			tieBreakerNumber: persisted.PrimaryTieBreakerNumber,
		}
		if err := e.validatePollingWatermarkPosition(*primaryDurable, "checkpoint"); err != nil {
			return pollingWatermarkFrontiers{}, err
		}
	} else if persisted.PrimaryTieBreakerNumber {
		return pollingWatermarkFrontiers{}, errors.New("polling watermark primary ordering metadata requires a durable checkpoint")
	}
	primaryQuery, err := e.initialPollingWatermarkSourceQuery(primaryDurable, scanCheckpoint)
	if err != nil {
		return pollingWatermarkFrontiers{}, err
	}
	frontiers := pollingWatermarkFrontiers{
		primary: pollingWatermarkSourceFrontier{
			durable: clonePollingWatermarkPosition(primaryDurable),
			query:   primaryQuery,
			scan:    clonePollingWatermarkScanCheckpoint(scanCheckpoint),
		},
	}
	if e.descriptor.PollingWatermark.DeletionEndpoint == nil {
		if persisted.Deletion != nil {
			return pollingWatermarkFrontiers{}, errors.New("polling watermark deletion checkpoint requires a declared deletion_endpoint")
		}
		return frontiers, nil
	}
	if persisted.Deletion == nil {
		var deletionQuery *PollingWatermarkPosition
		if e.descriptor.PollingWatermark.Watermark.Kind != "opaque_cursor" {
			deletionQuery, err = e.initialPollingWatermarkSourceQuery(primaryDurable, nil)
			if err != nil {
				return pollingWatermarkFrontiers{}, fmt.Errorf("initialize polling watermark deletion frontier: %w", err)
			}
		}
		frontiers.deletion = pollingWatermarkSourceFrontier{
			query:    deletionQuery,
			querySet: true,
		}
		return frontiers, nil
	}
	deletionDurable := pollingWatermarkPositionFromSourceCheckpoint(persisted.Deletion)
	var deletionQuery *PollingWatermarkPosition
	if deletionDurable == nil && e.descriptor.PollingWatermark.Watermark.Kind != "opaque_cursor" {
		deletionQuery = pollingWatermarkPositionFromCheckpoint(persisted.Deletion.Query)
	} else if deletionDurable != nil {
		deletionQuery, err = e.initialPollingWatermarkSourceQuery(deletionDurable, persisted.Deletion.Scan)
		if err != nil {
			return pollingWatermarkFrontiers{}, fmt.Errorf("polling watermark deletion checkpoint: %w", err)
		}
	}
	frontiers.deletion = pollingWatermarkSourceFrontier{
		durable:  deletionDurable,
		query:    deletionQuery,
		querySet: deletionDurable == nil && persisted.Deletion.QuerySet,
		scan:     clonePollingWatermarkScanCheckpoint(persisted.Deletion.Scan),
	}
	return frontiers, nil
}

func (e *PollingWatermarkExecutor) initialPollingWatermarkSourceQuery(durable *PollingWatermarkPosition, scan *pollingWatermarkScanCheckpoint) (*PollingWatermarkPosition, error) {
	if durable == nil {
		if scan != nil && scan.Active {
			return nil, errors.New("polling watermark scan cursor requires a durable checkpoint")
		}
		return nil, nil
	}
	if !e.usesPollingWatermarkScanCursor() {
		return clonePollingWatermarkPosition(durable), nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, durable.Watermark)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp polling checkpoint: %w", err)
	}
	queryAfter := &PollingWatermarkPosition{Watermark: parsed.UTC().Add(-time.Duration(e.descriptor.PollingWatermark.SafetyLagSeconds) * time.Second).Format(time.RFC3339Nano)}
	if scan == nil || !scan.Active {
		return queryAfter, nil
	}
	scanPosition := pollingWatermarkPositionFromScanCheckpoint(scan)
	comparison, err := comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, scanPosition, *durable)
	if err != nil {
		return nil, fmt.Errorf("compare polling watermark scan cursor: %w", err)
	}
	if comparison > 0 {
		return nil, errors.New("polling watermark scan cursor exceeds the durable checkpoint")
	}
	comparison, err = comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, scanPosition, *queryAfter)
	if err != nil {
		return nil, fmt.Errorf("compare polling watermark scan cursor: %w", err)
	}
	if comparison >= 0 {
		return &scanPosition, nil
	}
	return queryAfter, nil
}

func (e *PollingWatermarkExecutor) usesPollingWatermarkScanCursor() bool {
	polling := e.descriptor.PollingWatermark
	return polling.Watermark.Kind == "timestamp" && polling.SafetyLagSeconds > 0
}

func (e *PollingWatermarkExecutor) scanCheckpointFromState(state map[string]string) (*pollingWatermarkScanCheckpoint, error) {
	raw, found := state[pollingWatermarkScanStateKey]
	if !found {
		return nil, nil
	}
	var scan pollingWatermarkScanCheckpoint
	if err := json.Unmarshal([]byte(raw), &scan); err != nil {
		return nil, fmt.Errorf("invalid polling watermark scan cursor: %w", err)
	}
	if err := e.validatePollingWatermarkScanCheckpoint(&scan); err != nil {
		return nil, err
	}
	return &scan, nil
}

func (e *PollingWatermarkExecutor) frontiersCheckpointFromState(state map[string]string) (pollingWatermarkFrontiersCheckpoint, error) {
	raw, found := state[pollingWatermarkFrontiersStateKey]
	if !found {
		return pollingWatermarkFrontiersCheckpoint{}, nil
	}
	var checkpoint pollingWatermarkFrontiersCheckpoint
	if err := json.Unmarshal([]byte(raw), &checkpoint); err != nil {
		return pollingWatermarkFrontiersCheckpoint{}, fmt.Errorf("invalid polling watermark source frontiers: %w", err)
	}
	if checkpoint.Version != 1 {
		return pollingWatermarkFrontiersCheckpoint{}, fmt.Errorf("unsupported polling watermark source frontiers version %d", checkpoint.Version)
	}
	if checkpoint.Deletion != nil {
		if err := e.validatePollingWatermarkSourceCheckpoint(checkpoint.Deletion); err != nil {
			return pollingWatermarkFrontiersCheckpoint{}, err
		}
	}
	return checkpoint, nil
}

func (e *PollingWatermarkExecutor) validatePollingWatermarkSourceCheckpoint(checkpoint *pollingWatermarkSourceCheckpoint) error {
	if checkpoint == nil {
		return nil
	}
	durable := pollingWatermarkPositionFromSourceCheckpoint(checkpoint)
	if durable != nil {
		if err := e.validatePollingWatermarkPosition(*durable, "deletion checkpoint"); err != nil {
			return err
		}
	} else if checkpoint.TieBreakerNumber {
		return errors.New("polling watermark deletion ordering metadata requires a durable checkpoint")
	}
	if durable == nil {
		if !checkpoint.QuerySet {
			return errors.New("polling watermark deletion checkpoint requires a query frontier")
		}
		if checkpoint.Query != nil && e.descriptor.PollingWatermark.Watermark.Kind != "opaque_cursor" {
			query := pollingWatermarkPositionFromCheckpoint(checkpoint.Query)
			if err := e.validatePollingWatermarkQueryPosition(*query, "deletion query"); err != nil {
				return err
			}
		}
	}
	if checkpoint.Scan != nil {
		if durable == nil {
			return errors.New("polling watermark deletion scan cursor requires a durable checkpoint")
		}
		if err := e.validatePollingWatermarkScanCheckpoint(checkpoint.Scan); err != nil {
			return fmt.Errorf("invalid polling watermark deletion scan cursor: %w", err)
		}
	}
	return nil
}

func (e *PollingWatermarkExecutor) validatePollingWatermarkScanCheckpoint(scan *pollingWatermarkScanCheckpoint) error {
	if scan == nil {
		return nil
	}
	if scan.Version != 1 {
		return fmt.Errorf("unsupported polling watermark scan cursor version %d", scan.Version)
	}
	if err := e.validatePollingWatermarkPosition(pollingWatermarkPositionFromScanCheckpoint(scan), "scan cursor"); err != nil {
		return fmt.Errorf("invalid polling watermark scan cursor: %w", err)
	}
	return nil
}

func (e *PollingWatermarkExecutor) validatePollingWatermarkPosition(position PollingWatermarkPosition, name string) error {
	if err := validatePollingWatermarkValue(e.descriptor.PollingWatermark.Watermark.Kind, position.Watermark); err != nil {
		return fmt.Errorf("invalid polling watermark %s: %w", name, err)
	}
	if strings.TrimSpace(position.TieBreaker) == "" {
		return fmt.Errorf("polling watermark %s has an empty tie breaker", name)
	}
	if position.tieBreakerNumber {
		if _, ok := new(big.Rat).SetString(position.TieBreaker); !ok {
			return fmt.Errorf("polling watermark %s has invalid numeric tie breaker %q", name, position.TieBreaker)
		}
	}
	return nil
}

func (e *PollingWatermarkExecutor) validatePollingWatermarkQueryPosition(position PollingWatermarkPosition, name string) error {
	if position.TieBreaker == "" {
		if !e.usesPollingWatermarkScanCursor() {
			return fmt.Errorf("polling watermark %s has an empty tie breaker", name)
		}
		if position.tieBreakerNumber {
			return fmt.Errorf("polling watermark %s has numeric metadata without a tie breaker", name)
		}
		if err := validatePollingWatermarkValue(e.descriptor.PollingWatermark.Watermark.Kind, position.Watermark); err != nil {
			return fmt.Errorf("invalid polling watermark %s: %w", name, err)
		}
		return nil
	}
	return e.validatePollingWatermarkPosition(position, name)
}

func (e *PollingWatermarkExecutor) nextPollingWatermarkScanCheckpoint(position PollingWatermarkPosition, active bool) *pollingWatermarkScanCheckpoint {
	if !e.usesPollingWatermarkScanCursor() {
		return nil
	}
	return &pollingWatermarkScanCheckpoint{
		Version:          1,
		Watermark:        position.Watermark,
		TieBreaker:       position.TieBreaker,
		TieBreakerNumber: position.tieBreakerNumber,
		Active:           active,
	}
}

func (e *PollingWatermarkExecutor) advancePollingWatermarkSource(frontier pollingWatermarkSourceFrontier, accepted []positionedPollingWatermarkRecord, more bool, sourceRecordCount int) (pollingWatermarkSourceFrontier, bool, error) {
	next := clonePollingWatermarkSourceFrontier(frontier)
	changed := false
	if len(accepted) == 0 {
		if e.usesPollingWatermarkScanCursor() && next.scan != nil && next.scan.Active && !more && sourceRecordCount == 0 {
			inactive := *next.scan
			inactive.Active = false
			next.scan = &inactive
			changed = true
		}
		return next, changed, nil
	}
	last := accepted[len(accepted)-1].position
	durable, durableChanged, err := e.nextDurablePosition(next.durable, last)
	if err != nil {
		return pollingWatermarkSourceFrontier{}, false, err
	}
	next.durable = &durable
	next.query = clonePollingWatermarkPosition(&last)
	changed = durableChanged
	if e.usesPollingWatermarkScanCursor() {
		scan := e.nextPollingWatermarkScanCheckpoint(last, more || len(accepted) < sourceRecordCount)
		if !samePollingWatermarkScanCheckpoint(next.scan, scan) {
			changed = true
		}
		next.scan = scan
	} else if next.scan != nil {
		next.scan = nil
		changed = true
	}
	return next, changed, nil
}

func (e *PollingWatermarkExecutor) nextDurablePosition(durable *PollingWatermarkPosition, last PollingWatermarkPosition) (PollingWatermarkPosition, bool, error) {
	if durable == nil {
		return last, true, nil
	}
	if e.descriptor.PollingWatermark.Watermark.Kind == "opaque_cursor" {
		if samePollingWatermarkTuple(*durable, last) {
			next := *durable
			next.tieBreakerNumber = durable.tieBreakerNumber || last.tieBreakerNumber
			return next, next.tieBreakerNumber != durable.tieBreakerNumber, nil
		}
		return last, true, nil
	}
	comparison, err := comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, *durable, last)
	if err != nil {
		return PollingWatermarkPosition{}, false, err
	}
	if comparison > 0 {
		return *durable, false, nil
	}
	if comparison == 0 {
		next := *durable
		next.tieBreakerNumber = durable.tieBreakerNumber || last.tieBreakerNumber
		return next, next.tieBreakerNumber != durable.tieBreakerNumber, nil
	}
	return last, true, nil
}

func (e *PollingWatermarkExecutor) emitPollingRecord(ctx context.Context, record connectors.Record, deletionRecord bool, emit func(connectors.CDCEvent) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	operation := "upsert"
	if deletionRecord || (e.descriptor.PollingWatermark.SoftDelete != nil && pollingWatermarkTruthy(recordValueAt(record, e.descriptor.PollingWatermark.SoftDelete.Path))) {
		operation = "delete"
	}
	if err := emit(connectors.CDCEvent{Operation: operation, Record: clonePollingWatermarkRecord(record)}); err != nil {
		return fmt.Errorf("durably accept polling watermark event: %w", err)
	}
	return nil
}

func (e *PollingWatermarkExecutor) positionFromRecord(record connectors.Record) (PollingWatermarkPosition, error) {
	polling := e.descriptor.PollingWatermark
	watermark, found := recordValueAt(record, polling.Watermark.Path)
	if !found {
		return PollingWatermarkPosition{}, fmt.Errorf("record is missing polling watermark path %q", polling.Watermark.Path)
	}
	watermarkValue, err := pollingWatermarkScalar(watermark)
	if err != nil {
		return PollingWatermarkPosition{}, fmt.Errorf("record polling watermark %q: %w", polling.Watermark.Path, err)
	}
	if err := validatePollingWatermarkValue(polling.Watermark.Kind, watermarkValue); err != nil {
		return PollingWatermarkPosition{}, err
	}
	tieBreaker, found := recordValueAt(record, polling.TieBreaker.Path)
	if !found {
		return PollingWatermarkPosition{}, fmt.Errorf("record is missing polling tie_breaker path %q", polling.TieBreaker.Path)
	}
	tieBreakerValue, tieBreakerNumber, err := pollingWatermarkTieBreakerScalar(tieBreaker)
	if err != nil {
		return PollingWatermarkPosition{}, fmt.Errorf("record polling tie_breaker %q: %w", polling.TieBreaker.Path, err)
	}
	if tieBreakerValue == "" {
		return PollingWatermarkPosition{}, errors.New("record polling tie_breaker cannot be empty")
	}
	return PollingWatermarkPosition{Watermark: watermarkValue, TieBreaker: tieBreakerValue, tieBreakerNumber: tieBreakerNumber}, nil
}

func (e *PollingWatermarkExecutor) requireOrderedPosition(previous *PollingWatermarkPosition, current PollingWatermarkPosition) error {
	if previous == nil {
		return nil
	}
	if e.descriptor.PollingWatermark.Watermark.Kind == "opaque_cursor" {
		return errors.New("opaque polling watermark cursors are not orderable")
	}
	comparison, err := comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, *previous, current)
	if err != nil {
		return err
	}
	if comparison > 0 {
		return fmt.Errorf("source page is not ordered by %s then tie_breaker", e.descriptor.PollingWatermark.Watermark.Path)
	}
	return nil
}

type positionedPollingWatermarkRecord struct {
	record         connectors.Record
	deletionRecord bool
	position       PollingWatermarkPosition
}

func (e *PollingWatermarkExecutor) requirePollingWatermarkSourceProgress(after *PollingWatermarkPosition, records []positionedPollingWatermarkRecord, more bool, page int, source string) error {
	if len(records) == 0 {
		if more {
			return &PollingWatermarkNonAdvancingError{Reason: PollingWatermarkNonAdvancingReasonCursor, Page: page, Source: source}
		}
		return nil
	}
	if e.descriptor.PollingWatermark.Watermark.Kind != "opaque_cursor" {
		if err := e.requireOrderedPosition(after, records[0].position); err != nil {
			return err
		}
	}
	if !more || after == nil {
		return nil
	}
	last := records[len(records)-1].position
	if e.descriptor.PollingWatermark.Watermark.Kind == "opaque_cursor" {
		if samePollingWatermarkTuple(*after, last) {
			return &PollingWatermarkNonAdvancingError{Reason: PollingWatermarkNonAdvancingReasonCursor, Page: page, Source: source}
		}
		return nil
	}
	comparison, err := comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, *after, last)
	if err != nil {
		return err
	}
	if comparison >= 0 {
		return &PollingWatermarkNonAdvancingError{Reason: PollingWatermarkNonAdvancingReasonCursor, Page: page, Source: source}
	}
	return nil
}

func (e *PollingWatermarkExecutor) validatePhysicalPage(pageNumber int, page PollingWatermarkPage) error {
	limit := e.descriptor.PollingWatermark.PageSize
	if len(page.Records) > limit {
		return &PollingWatermarkNonAdvancingError{
			Reason:   PollingWatermarkNonAdvancingReasonPageSize,
			Page:     pageNumber,
			Source:   "primary",
			Limit:    limit,
			Returned: len(page.Records),
		}
	}
	if e.descriptor.PollingWatermark.DeletionEndpoint != nil && len(page.DeletionRecords) > limit {
		return &PollingWatermarkNonAdvancingError{
			Reason:   PollingWatermarkNonAdvancingReasonPageSize,
			Page:     pageNumber,
			Source:   "deletion",
			Limit:    limit,
			Returned: len(page.DeletionRecords),
		}
	}
	return nil
}

func (e *PollingWatermarkExecutor) mergePollingWatermarkPageRecords(primary, deletions []positionedPollingWatermarkRecord, primaryMore, deletionMore bool) ([]positionedPollingWatermarkRecord, error) {
	if e.descriptor.PollingWatermark.Watermark.Kind == "opaque_cursor" {
		merged := make([]positionedPollingWatermarkRecord, 0, len(primary)+len(deletions))
		merged = append(merged, primary...)
		return append(merged, deletions...), nil
	}
	merged := make([]positionedPollingWatermarkRecord, 0, len(primary)+len(deletions))
	remainingPrimary := primary
	remainingDeletions := deletions
	for len(remainingPrimary) > 0 && len(remainingDeletions) > 0 {
		comparison, err := comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, remainingPrimary[0].position, remainingDeletions[0].position)
		if err != nil {
			return nil, err
		}
		if comparison <= 0 {
			merged = append(merged, remainingPrimary[0])
			remainingPrimary = remainingPrimary[1:]
			continue
		}
		merged = append(merged, remainingDeletions[0])
		remainingDeletions = remainingDeletions[1:]
	}
	merged = append(merged, remainingPrimary...)
	merged = append(merged, remainingDeletions...)
	if !primaryMore && !deletionMore {
		return merged, nil
	}
	var boundary *PollingWatermarkPosition
	if primaryMore {
		position := primary[len(primary)-1].position
		boundary = &position
	}
	if deletionMore {
		position := deletions[len(deletions)-1].position
		if boundary == nil {
			boundary = &position
		} else {
			comparison, err := comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, position, *boundary)
			if err != nil {
				return nil, err
			}
			if comparison < 0 {
				boundary = &position
			}
		}
	}
	accepted := make([]positionedPollingWatermarkRecord, 0, len(merged))
	for _, record := range merged {
		comparison, err := comparePollingWatermarkPositions(e.descriptor.PollingWatermark.Watermark.Kind, record.position, *boundary)
		if err != nil {
			return nil, err
		}
		if comparison <= 0 {
			accepted = append(accepted, record)
		}
	}
	return accepted, nil
}

func (e *PollingWatermarkExecutor) positionedPollingWatermarkRecords(records []connectors.Record, deletionRecord bool) ([]positionedPollingWatermarkRecord, error) {
	positioned := make([]positionedPollingWatermarkRecord, 0, len(records))
	var previous *PollingWatermarkPosition
	for _, record := range records {
		position, err := e.positionFromRecord(record)
		if err != nil {
			return nil, err
		}
		if e.descriptor.PollingWatermark.Watermark.Kind != "opaque_cursor" {
			if err := e.requireOrderedPosition(previous, position); err != nil {
				return nil, err
			}
		}
		positioned = append(positioned, positionedPollingWatermarkRecord{record: record, deletionRecord: deletionRecord, position: position})
		previous = &positioned[len(positioned)-1].position
	}
	return positioned, nil
}

func (e *PollingWatermarkExecutor) checkpointState(current map[string]string, frontiers pollingWatermarkFrontiers) (map[string]string, error) {
	state := clonePollingWatermarkState(current)
	if frontiers.primary.durable == nil {
		delete(state, e.descriptor.Checkpoint.Keys[0])
		delete(state, e.descriptor.Checkpoint.Keys[1])
	} else {
		state[e.descriptor.Checkpoint.Keys[0]] = frontiers.primary.durable.Watermark
		state[e.descriptor.Checkpoint.Keys[1]] = frontiers.primary.durable.TieBreaker
	}
	if !e.usesPollingWatermarkScanCursor() || frontiers.primary.scan == nil {
		delete(state, pollingWatermarkScanStateKey)
	} else {
		raw, err := json.Marshal(frontiers.primary.scan)
		if err != nil {
			return nil, fmt.Errorf("encode polling watermark scan cursor: %w", err)
		}
		state[pollingWatermarkScanStateKey] = string(raw)
	}
	persisted := pollingWatermarkFrontiersCheckpoint{Version: 1}
	if frontiers.primary.durable != nil {
		persisted.PrimaryTieBreakerNumber = frontiers.primary.durable.tieBreakerNumber
	}
	persisted.Deletion = pollingWatermarkSourceCheckpointFromFrontier(frontiers.deletion)
	if !persisted.PrimaryTieBreakerNumber && persisted.Deletion == nil {
		delete(state, pollingWatermarkFrontiersStateKey)
		return state, nil
	}
	raw, err := json.Marshal(persisted)
	if err != nil {
		return nil, fmt.Errorf("encode polling watermark source frontiers: %w", err)
	}
	state[pollingWatermarkFrontiersStateKey] = string(raw)
	return state, nil
}

func clonePollingWatermarkState(state map[string]string) map[string]string {
	copy := make(map[string]string, len(state))
	for key, value := range state {
		copy[key] = value
	}
	return copy
}

func (page PollingWatermarkPage) continuation(deletionConfigured bool) (bool, bool) {
	primaryMore := page.More || page.PrimaryMore
	if !deletionConfigured {
		return primaryMore, false
	}
	return primaryMore, page.More || page.DeletionMore
}

func splitPollingWatermarkRecords(records []positionedPollingWatermarkRecord) ([]positionedPollingWatermarkRecord, []positionedPollingWatermarkRecord) {
	primary := make([]positionedPollingWatermarkRecord, 0, len(records))
	deletions := make([]positionedPollingWatermarkRecord, 0, len(records))
	for _, record := range records {
		if record.deletionRecord {
			deletions = append(deletions, record)
			continue
		}
		primary = append(primary, record)
	}
	return primary, deletions
}

func pollingWatermarkPositionFromScanCheckpoint(checkpoint *pollingWatermarkScanCheckpoint) PollingWatermarkPosition {
	return PollingWatermarkPosition{
		Watermark:        checkpoint.Watermark,
		TieBreaker:       checkpoint.TieBreaker,
		tieBreakerNumber: checkpoint.TieBreakerNumber,
	}
}

func pollingWatermarkPositionFromSourceCheckpoint(checkpoint *pollingWatermarkSourceCheckpoint) *PollingWatermarkPosition {
	if checkpoint == nil || (checkpoint.Watermark == "" && checkpoint.TieBreaker == "") {
		return nil
	}
	return &PollingWatermarkPosition{
		Watermark:        checkpoint.Watermark,
		TieBreaker:       checkpoint.TieBreaker,
		tieBreakerNumber: checkpoint.TieBreakerNumber,
	}
}

func pollingWatermarkPositionFromCheckpoint(checkpoint *pollingWatermarkPositionCheckpoint) *PollingWatermarkPosition {
	if checkpoint == nil {
		return nil
	}
	return &PollingWatermarkPosition{
		Watermark:        checkpoint.Watermark,
		TieBreaker:       checkpoint.TieBreaker,
		tieBreakerNumber: checkpoint.TieBreakerNumber,
	}
}

func pollingWatermarkPositionCheckpointFromPosition(position *PollingWatermarkPosition) *pollingWatermarkPositionCheckpoint {
	if position == nil {
		return nil
	}
	return &pollingWatermarkPositionCheckpoint{
		Watermark:        position.Watermark,
		TieBreaker:       position.TieBreaker,
		TieBreakerNumber: position.tieBreakerNumber,
	}
}

func pollingWatermarkSourceCheckpointFromFrontier(frontier pollingWatermarkSourceFrontier) *pollingWatermarkSourceCheckpoint {
	if frontier.durable != nil {
		return &pollingWatermarkSourceCheckpoint{
			Watermark:        frontier.durable.Watermark,
			TieBreaker:       frontier.durable.TieBreaker,
			TieBreakerNumber: frontier.durable.tieBreakerNumber,
			Scan:             clonePollingWatermarkScanCheckpoint(frontier.scan),
		}
	}
	if !frontier.querySet && frontier.scan == nil {
		return nil
	}
	return &pollingWatermarkSourceCheckpoint{
		QuerySet: frontier.querySet,
		Query:    pollingWatermarkPositionCheckpointFromPosition(frontier.query),
		Scan:     clonePollingWatermarkScanCheckpoint(frontier.scan),
	}
}

func clonePollingWatermarkSourceFrontier(frontier pollingWatermarkSourceFrontier) pollingWatermarkSourceFrontier {
	return pollingWatermarkSourceFrontier{
		durable:  clonePollingWatermarkPosition(frontier.durable),
		query:    clonePollingWatermarkPosition(frontier.query),
		querySet: frontier.querySet,
		scan:     clonePollingWatermarkScanCheckpoint(frontier.scan),
	}
}

func clonePollingWatermarkScanCheckpoint(checkpoint *pollingWatermarkScanCheckpoint) *pollingWatermarkScanCheckpoint {
	if checkpoint == nil {
		return nil
	}
	copy := *checkpoint
	return &copy
}

func samePollingWatermarkTuple(left, right PollingWatermarkPosition) bool {
	return left.Watermark == right.Watermark && left.TieBreaker == right.TieBreaker
}

func samePollingWatermarkScanCheckpoint(left, right *pollingWatermarkScanCheckpoint) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Version == right.Version && left.Watermark == right.Watermark && left.TieBreaker == right.TieBreaker && left.TieBreakerNumber == right.TieBreakerNumber && left.Active == right.Active
}

// PollingWatermarkConnector joins the standard declarative connector view to
// the shared executor without provider-specific Go. It is the runtime value
// registered for an implemented polling-watermark bundle.
type PollingWatermarkConnector struct {
	*Connector
	executor *PollingWatermarkExecutor
}

var _ connectors.ChangefeedExecutor = (*PollingWatermarkConnector)(nil)

// NewPollingWatermarkConnector creates a connector whose changefeed behavior
// is entirely selected by bundle.Changefeed and the closed page-source port.
func NewPollingWatermarkConnector(bundle Bundle, hooks Hooks, source PollingWatermarkPageSource) (*PollingWatermarkConnector, error) {
	executor, err := NewPollingWatermarkExecutor(bundle.Changefeed, source)
	if err != nil {
		return nil, err
	}
	for _, streamName := range bundle.Changefeed.Streams {
		if _, err := findStream(bundle, streamName); err != nil {
			return nil, fmt.Errorf("polling watermark declaration: %w", err)
		}
	}
	return &PollingWatermarkConnector{Connector: New(bundle, hooks), executor: executor}, nil
}

func (c *PollingWatermarkConnector) ChangefeedExecutorDescriptor() connectors.ChangefeedExecutorDescriptor {
	if c == nil || c.executor == nil {
		return connectors.ChangefeedExecutorDescriptor{}
	}
	return c.executor.ChangefeedExecutorDescriptor()
}

func (c *PollingWatermarkConnector) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	if c == nil || c.executor == nil {
		return errors.New("polling watermark connector is not configured")
	}
	return c.executor.ReadCDC(ctx, req, emit)
}

func containsDeclaredStream(streams []string, wanted string) bool {
	for _, stream := range streams {
		if stream == wanted {
			return true
		}
	}
	return false
}

func clonePollingWatermarkPosition(position *PollingWatermarkPosition) *PollingWatermarkPosition {
	if position == nil {
		return nil
	}
	copy := *position
	return &copy
}

func clonePollingWatermarkDeletionEndpoint(endpoint *connectors.PollingWatermarkDeletionEndpoint) *connectors.PollingWatermarkDeletionEndpoint {
	if endpoint == nil {
		return nil
	}
	copy := *endpoint
	return &copy
}

func clonePollingWatermarkRecord(record connectors.Record) connectors.Record {
	copy := make(connectors.Record, len(record))
	for key, value := range record {
		copy[key] = value
	}
	return copy
}

func recordValueAt(record connectors.Record, path string) (any, bool) {
	var value any = record
	for _, segment := range strings.Split(path, ".") {
		switch object := value.(type) {
		case connectors.Record:
			next, found := object[segment]
			if !found {
				return nil, false
			}
			value = next
		case map[string]any:
			next, found := object[segment]
			if !found {
				return nil, false
			}
			value = next
		default:
			return nil, false
		}
	}
	return value, value != nil
}

func pollingWatermarkScalar(value any) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	case int:
		return strconv.Itoa(typed), nil
	case int8:
		return strconv.FormatInt(int64(typed), 10), nil
	case int16:
		return strconv.FormatInt(int64(typed), 10), nil
	case int32:
		return strconv.FormatInt(int64(typed), 10), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case uint:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint64:
		return strconv.FormatUint(typed, 10), nil
	default:
		return "", fmt.Errorf("requires a lossless string, json.Number, or integer value, got %T", value)
	}
}

func pollingWatermarkTieBreakerScalar(value any) (string, bool, error) {
	scalar, err := pollingWatermarkScalar(value)
	if err != nil {
		return "", false, err
	}
	if !pollingWatermarkNumericValue(value) {
		return scalar, false, nil
	}
	if _, ok := new(big.Rat).SetString(scalar); !ok {
		return "", false, fmt.Errorf("numeric value %q is invalid", scalar)
	}
	return scalar, true, nil
}

func pollingWatermarkNumericValue(value any) bool {
	switch value.(type) {
	case json.Number, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	default:
		return false
	}
}

func validatePollingWatermarkValue(kind, value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New("polling watermark value cannot be empty")
	}
	switch kind {
	case "timestamp":
		if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
			return fmt.Errorf("timestamp watermark must be RFC3339Nano: %w", err)
		}
	case "monotonic_sequence":
		if _, ok := new(big.Int).SetString(value, 10); !ok {
			return fmt.Errorf("monotonic sequence watermark %q is not an integer", value)
		}
	case "opaque_cursor":
		// Opaque cursor ordering is provider-owned. Preserve it exactly and only
		// use equality to detect a non-progressing inclusive page.
	default:
		return fmt.Errorf("unsupported polling watermark kind %q", kind)
	}
	return nil
}

func comparePollingWatermarkPositions(kind string, left, right PollingWatermarkPosition) (int, error) {
	switch kind {
	case "timestamp":
		leftTime, err := time.Parse(time.RFC3339Nano, left.Watermark)
		if err != nil {
			return 0, err
		}
		rightTime, err := time.Parse(time.RFC3339Nano, right.Watermark)
		if err != nil {
			return 0, err
		}
		if leftTime.Before(rightTime) {
			return -1, nil
		}
		if leftTime.After(rightTime) {
			return 1, nil
		}
	case "monotonic_sequence":
		leftNumber, ok := new(big.Int).SetString(left.Watermark, 10)
		if !ok {
			return 0, fmt.Errorf("invalid monotonic sequence %q", left.Watermark)
		}
		rightNumber, ok := new(big.Int).SetString(right.Watermark, 10)
		if !ok {
			return 0, fmt.Errorf("invalid monotonic sequence %q", right.Watermark)
		}
		if comparison := leftNumber.Cmp(rightNumber); comparison != 0 {
			return comparison, nil
		}
	case "opaque_cursor":
		return 0, errors.New("opaque polling watermark cursors are not orderable")
	default:
		return 0, fmt.Errorf("unsupported polling watermark kind %q", kind)
	}
	return comparePollingWatermarkTieBreakers(left, right)
}

func comparePollingWatermarkTieBreakers(left, right PollingWatermarkPosition) (int, error) {
	if left.TieBreaker == "" || right.TieBreaker == "" {
		return strings.Compare(left.TieBreaker, right.TieBreaker), nil
	}
	if left.tieBreakerNumber || right.tieBreakerNumber {
		leftNumber, leftOK := new(big.Rat).SetString(left.TieBreaker)
		rightNumber, rightOK := new(big.Rat).SetString(right.TieBreaker)
		if !leftOK || !rightOK {
			return 0, errors.New("numeric polling watermark tie breaker cannot be compared with non-numeric text")
		}
		return leftNumber.Cmp(rightNumber), nil
	}
	return strings.Compare(left.TieBreaker, right.TieBreaker), nil
}

func pollingWatermarkTruthy(value any, found bool) bool {
	if !found || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		value := strings.TrimSpace(strings.ToLower(typed))
		return value != "" && value != "false" && value != "0" && value != "null"
	case json.Number:
		number, ok := new(big.Rat).SetString(typed.String())
		return ok && number.Sign() != 0
	case float32:
		return typed != 0
	case float64:
		return typed != 0
	case int:
		return typed != 0
	case int8:
		return typed != 0
	case int16:
		return typed != 0
	case int32:
		return typed != 0
	case int64:
		return typed != 0
	case uint:
		return typed != 0
	case uint8:
		return typed != 0
	case uint16:
		return typed != 0
	case uint32:
		return typed != 0
	case uint64:
		return typed != 0
	default:
		return false
	}
}
