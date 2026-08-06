package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

// PollingWatermarkClock supplies time for timestamp-watermark safety lag. It
// is an explicit dependency so polling behavior has no hidden wall-clock or
// sleep dependency in tests.
type PollingWatermarkClock interface {
	Now() time.Time
}

type wallPollingWatermarkClock struct{}

func (wallPollingWatermarkClock) Now() time.Time { return time.Now().UTC() }

// PollingWatermarkPosition is the complete provider ordering tuple persisted
// after a page is durably accepted. Watermark and TieBreaker retain the
// provider's lossless textual representation; they are not float-coerced.
type PollingWatermarkPosition struct {
	Watermark  string
	TieBreaker string
}

// PollingWatermarkPageRequest is the closed request passed to a connector
// transport adapter. After is inclusive by contract: the edge record is read
// again so timestamp ties cannot be silently skipped. DeletionEndpoint is
// copied from the declaration when a provider exposes a separate deletion
// feed; transport code never receives caller-authored endpoint structure.
type PollingWatermarkPageRequest struct {
	Stream           string
	Config           connectors.RuntimeConfig
	After            *PollingWatermarkPosition
	Inclusive        bool
	PageSize         int
	DeletionEndpoint *connectors.PollingWatermarkDeletionEndpoint
}

// PollingWatermarkPage is one bounded source response. DeletionRecords are
// supplied only for a declared deletion endpoint; all other hard deletes are
// intentionally invisible to polling.
type PollingWatermarkPage struct {
	Records         []connectors.Record
	DeletionRecords []connectors.Record
	More            bool
}

// PollingWatermarkPageSource is the transport seam for the shared executor.
// It must implement only declaration-selected read operations; it never
// receives raw SQL, arbitrary HTTP, shell, or caller-selected paths.
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
	clock      PollingWatermarkClock
}

var _ connectors.ChangefeedExecutor = (*PollingWatermarkExecutor)(nil)

// NewPollingWatermarkExecutor constructs the shared executor only for a fully
// validated polling-watermark declaration and a closed source adapter.
func NewPollingWatermarkExecutor(descriptor *connectors.ChangefeedDescriptor, source PollingWatermarkPageSource, clock PollingWatermarkClock) (*PollingWatermarkExecutor, error) {
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
	if clock == nil {
		clock = wallPollingWatermarkClock{}
	}
	return &PollingWatermarkExecutor{descriptor: descriptor.Clone(), source: source, clock: clock}, nil
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
	after, err := e.initialPosition(req.State)
	if err != nil {
		return err
	}
	requestLimit := polling.MaxPages
	if polling.RequestBudget < requestLimit {
		requestLimit = polling.RequestBudget
	}

	for pageNumber := 0; pageNumber < requestLimit; pageNumber++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		page, err := e.source.FetchPollingWatermarkPage(ctx, PollingWatermarkPageRequest{
			Stream:           req.Stream,
			Config:           req.Config,
			After:            clonePollingWatermarkPosition(after),
			Inclusive:        true,
			PageSize:         polling.PageSize,
			DeletionEndpoint: clonePollingWatermarkDeletionEndpoint(polling.DeletionEndpoint),
		})
		if err != nil {
			return fmt.Errorf("polling watermark fetch page %d: %w", pageNumber+1, err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(page.Records) == 0 && len(page.DeletionRecords) == 0 {
			return nil
		}

		var last *PollingWatermarkPosition
		var previous *PollingWatermarkPosition
		for _, record := range page.Records {
			position, err := e.emitPollingRecord(ctx, record, false, emit)
			if err != nil {
				return err
			}
			if err := e.requireOrderedPosition(previous, position); err != nil {
				return fmt.Errorf("polling watermark page %d: %w", pageNumber+1, err)
			}
			previous = &position
			last = &position
		}
		for _, record := range page.DeletionRecords {
			position, err := e.emitPollingRecord(ctx, record, true, emit)
			if err != nil {
				return err
			}
			if err := e.requireOrderedPosition(previous, position); err != nil {
				return fmt.Errorf("polling watermark page %d: %w", pageNumber+1, err)
			}
			previous = &position
			last = &position
		}
		if last == nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := req.CheckpointCommitter.CommitChangefeedCheckpoint(ctx, e.checkpointState(*last)); err != nil {
			return fmt.Errorf("polling watermark commit checkpoint after destination acknowledgement: %w", err)
		}
		after = last
		if !page.More {
			return nil
		}
	}
	return nil
}

func (e *PollingWatermarkExecutor) initialPosition(state map[string]string) (*PollingWatermarkPosition, error) {
	polling := e.descriptor.PollingWatermark
	keys := e.descriptor.Checkpoint.Keys
	if len(keys) != 2 {
		return nil, errors.New("polling watermark executor requires exactly two checkpoint keys")
	}
	watermark, hasWatermark := state[keys[0]]
	tieBreaker, hasTieBreaker := state[keys[1]]
	if hasWatermark != hasTieBreaker {
		return nil, errors.New("polling watermark checkpoint is incomplete")
	}
	if !hasWatermark {
		if polling.Watermark.Kind != "timestamp" {
			return nil, nil
		}
		return &PollingWatermarkPosition{Watermark: e.clock.Now().UTC().Add(-time.Duration(polling.SafetyLagSeconds) * time.Second).Format(time.RFC3339Nano)}, nil
	}
	committed := PollingWatermarkPosition{Watermark: watermark, TieBreaker: tieBreaker}
	if err := validatePollingWatermarkValue(polling.Watermark.Kind, committed.Watermark); err != nil {
		return nil, fmt.Errorf("invalid polling watermark checkpoint: %w", err)
	}
	if strings.TrimSpace(committed.TieBreaker) == "" {
		return nil, errors.New("polling watermark checkpoint has an empty tie breaker")
	}
	if polling.Watermark.Kind != "timestamp" || polling.SafetyLagSeconds == 0 {
		return &committed, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, committed.Watermark)
	if err != nil {
		return nil, fmt.Errorf("parse timestamp polling checkpoint: %w", err)
	}
	return &PollingWatermarkPosition{Watermark: parsed.UTC().Add(-time.Duration(polling.SafetyLagSeconds) * time.Second).Format(time.RFC3339Nano)}, nil
}

func (e *PollingWatermarkExecutor) emitPollingRecord(ctx context.Context, record connectors.Record, deletionRecord bool, emit func(connectors.CDCEvent) error) (PollingWatermarkPosition, error) {
	if err := ctx.Err(); err != nil {
		return PollingWatermarkPosition{}, err
	}
	position, err := e.positionFromRecord(record)
	if err != nil {
		return PollingWatermarkPosition{}, err
	}
	operation := "upsert"
	if deletionRecord || (e.descriptor.PollingWatermark.SoftDelete != nil && pollingWatermarkTruthy(recordValueAt(record, e.descriptor.PollingWatermark.SoftDelete.Path))) {
		operation = "delete"
	}
	if err := emit(connectors.CDCEvent{Operation: operation, Record: clonePollingWatermarkRecord(record)}); err != nil {
		return PollingWatermarkPosition{}, fmt.Errorf("durably accept polling watermark event: %w", err)
	}
	return position, nil
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
	tieBreakerValue, err := pollingWatermarkScalar(tieBreaker)
	if err != nil {
		return PollingWatermarkPosition{}, fmt.Errorf("record polling tie_breaker %q: %w", polling.TieBreaker.Path, err)
	}
	if tieBreakerValue == "" {
		return PollingWatermarkPosition{}, errors.New("record polling tie_breaker cannot be empty")
	}
	return PollingWatermarkPosition{Watermark: watermarkValue, TieBreaker: tieBreakerValue}, nil
}

func (e *PollingWatermarkExecutor) requireOrderedPosition(previous *PollingWatermarkPosition, current PollingWatermarkPosition) error {
	if previous == nil {
		return nil
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

func (e *PollingWatermarkExecutor) checkpointState(position PollingWatermarkPosition) map[string]string {
	return map[string]string{
		e.descriptor.Checkpoint.Keys[0]: position.Watermark,
		e.descriptor.Checkpoint.Keys[1]: position.TieBreaker,
	}
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
func NewPollingWatermarkConnector(bundle Bundle, hooks Hooks, source PollingWatermarkPageSource, clock PollingWatermarkClock) (*PollingWatermarkConnector, error) {
	executor, err := NewPollingWatermarkExecutor(bundle.Changefeed, source, clock)
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
		if left.Watermark != right.Watermark {
			return -1, nil
		}
	default:
		return 0, fmt.Errorf("unsupported polling watermark kind %q", kind)
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
		return typed.String() != "0"
	case float32:
		return typed != 0
	case float64:
		return typed != 0
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(typed) != "0"
	default:
		return false
	}
}
