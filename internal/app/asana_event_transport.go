package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	asanaEventSourceExecutorID = "asana_event_token_source"
	asanaEventMechanism        = "asana_events_sync_token"
	asanaEventMaxWindowPages   = 1000
)

var (
	asanaEventSourceReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     asanaEventSourceExecutorID,
	}
)

// asanaEventPage is the source-locked response shared by Asana's resource and
// workspace Events endpoints. The sync token is provider-owned cursor state;
// has_more says whether the returned token must be followed before the current
// token window is complete.
type asanaEventPage struct {
	Data    []asanaEvent `json:"data"`
	Sync    string       `json:"sync"`
	HasMore bool         `json:"has_more"`
}

type asanaEvent struct {
	Action    string        `json:"action"`
	CreatedAt string        `json:"created_at"`
	Resource  asanaResource `json:"resource"`
	Parent    asanaResource `json:"parent"`
}

type asanaResource struct {
	GID          string `json:"gid"`
	ResourceType string `json:"resource_type"`
}

type asanaEventStreamBinding struct {
	ResourceType string
	Operation    string
	PathField    string
}

// asanaEventStreamBindings is connector execution policy over source-backed
// operations. Each resource type is the singular form carried by
// EventResponse.resource.resource_type; each hydration operation and path
// field exists in the pinned Asana source lock and operations projection.
var asanaEventStreamBindings = map[string]asanaEventStreamBinding{
	// The pinned getEvents description explicitly proves that a project
	// subscription contains events for tasks in that project. No equally
	// specific coverage statement exists there for the other eleven ETL
	// resource collections, so they remain full-refresh-only.
	"tasks": {ResourceType: "task", Operation: "get_task", PathField: "task_gid"},
}

// asanaEventSourceExecutor implements the documented token protocol as a
// connector-local transport. It intentionally does not reuse list pagination
// offsets or page-content hashes as incremental cursors.
type asanaEventSourceExecutor struct{}

func (*asanaEventSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return asanaEventSourceReference
}

func (*asanaEventSourceExecutor) AllowEmptySourceResult() {}

func isAsanaEventTransportConnector(connector connectors.Connector) bool {
	descriptor, ok := connectors.SourceTransportDescriptorOf(connector)
	return ok && descriptor.Executor == asanaEventSourceReference
}

func isAsanaFullRefreshWarehouseRoute(source, destination connectors.Connector, mode synccontract.Mode) bool {
	if !isAsanaEventTransportConnector(source) || !isLocalWarehouseDestination(destination) {
		return false
	}
	materializer, ok := destination.(connectors.LocalWarehouseMaterializer)
	if !ok || !materializer.MaterializesLocalWarehouse() {
		return false
	}
	return mode == synccontract.ModeFullOverwrite || mode == synccontract.ModeFullAppend
}

// validateEndpointStreamSyncConfig keeps opaque provider-token checkpoints out
// of the legacy row-cursor field. All other sources retain the established
// cursor/primary-key validation unchanged.
func (a *App) validateEndpointStreamSyncConfig(source, destination connectors.Connector, streamName string, stream StreamConfig, mode SyncMode) error {
	descriptor, declared := connectors.SourceTransportDescriptorOf(source)
	if !declared || descriptor.Executor != asanaEventSourceReference || !transportContainsName(descriptor.EligibleStreams, streamName) || !transportContainsMode(descriptor.Modes, mode.ContractMode) {
		return ValidateStreamSyncConfig(stream)
	}
	if a == nil {
		return fmt.Errorf("Asana event-token transport registry is unavailable")
	}
	if err := a.ensureTransportRegistry(); err != nil {
		return err
	}
	if _, err := a.transports.Preflight(synctransport.PreflightRequest{Source: source, Destination: destination, Stream: streamName, Mode: mode.ContractMode, DestinationAction: stream.DestinationAction}); err != nil {
		return err
	}
	if len(stream.PrimaryKey) != 1 || strings.TrimSpace(stream.PrimaryKey[0]) != "gid" {
		return fmt.Errorf("Asana event-token stream %q requires primary key gid", streamName)
	}
	return nil
}

func (e *asanaEventSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if e == nil || request.Connector == nil {
		return fmt.Errorf("Asana event-token source is unavailable")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(request.Connector)
	if !ok || descriptor.Executor != asanaEventSourceReference || !transportContainsName(descriptor.EligibleStreams, request.Stream) || !transportContainsMode(descriptor.Modes, request.Mode) {
		return fmt.Errorf("Asana event-token source received an undeclared stream or sync mode")
	}
	if request.BatchSize <= 0 || request.BatchSize > issueCollectionTransportMaxRecords {
		return fmt.Errorf("Asana event-token batch size must be between 1 and %d", issueCollectionTransportMaxRecords)
	}
	if err := request.Resume.Source.Validate(); err != nil || len(request.Resume.SourceGeneration) == 0 {
		return fmt.Errorf("Asana event-token source requires a complete resume identity")
	}
	binding, ok := asanaEventStreamBindings[request.Stream]
	if !ok {
		return fmt.Errorf("Asana stream %q has no event hydration binding", request.Stream)
	}
	if len(request.PrimaryKey) != 1 || strings.TrimSpace(request.PrimaryKey[0]) != "gid" {
		return fmt.Errorf("Asana event-token stream %q requires the source-backed gid primary key", request.Stream)
	}
	projectGID := strings.TrimSpace(request.Runtime.Config["project_id"])
	if projectGID == "" || strings.TrimSpace(request.Runtime.Config["workspace_id"]) != "" || strings.TrimSpace(request.Runtime.Config["assignee"]) != "" {
		return fmt.Errorf("Asana event-token sync for stream %q requires an exact project scope: set project_id and leave workspace_id and assignee unset; all twelve streams remain available in full-refresh modes", request.Stream)
	}
	reader, ok := request.Connector.(connectors.OperationDirectReader)
	if !ok {
		return fmt.Errorf("Asana event-token source requires source-bound operation reads")
	}

	if request.Checkpoint == nil {
		bootstrap, err := readAsanaProjectEventPage(ctx, reader, request.Runtime, projectGID, "")
		if err != nil {
			return err
		}
		if !bootstrap.Rebootstrap {
			return fmt.Errorf("Asana initial Events request returned success without the documented 412 bootstrap token")
		}
		checkpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, bootstrap.Page.Sync, bootstrap.Page.Sync, nil)
		if err != nil {
			return err
		}
		return emitAsanaFullSnapshot(ctx, request, checkpoint, emit)
	}
	if err := request.Checkpoint.ValidateResume(request.Resume); err != nil {
		return err
	}
	if request.Checkpoint.Mechanism != asanaEventMechanism {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint was not produced by the Asana Events token protocol")
	}
	previousToken := strings.TrimSpace(string(request.Checkpoint.Position.Primary))
	if previousToken == "" {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "Asana Events checkpoint has no sync token")
	}
	events, finalToken, resetToken, err := readAsanaEventWindow(ctx, reader, request.Runtime, projectGID, previousToken)
	if err != nil {
		return err
	}
	if resetToken != "" {
		checkpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, resetToken, resetToken, nil)
		if err != nil {
			return err
		}
		return emitAsanaFullSnapshot(ctx, request, checkpoint, emit)
	}
	checkpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, previousToken, finalToken, events)
	if err != nil {
		return err
	}
	records, tombstones, err := hydrateAsanaEventWindow(ctx, reader, request, binding, events, checkpoint.Position)
	if err != nil {
		return err
	}
	return emitAsanaEventWorksets(records, tombstones, checkpoint, request.BatchSize, emit)
}

type asanaEventRead struct {
	Page        asanaEventPage
	Rebootstrap bool
}

func readAsanaProjectEventPage(ctx context.Context, reader connectors.OperationDirectReader, runtime connectors.RuntimeConfig, projectGID, token string) (asanaEventRead, error) {
	query := map[string]string{"resource": projectGID}
	if token = strings.TrimSpace(token); token != "" {
		query["sync"] = token
	}
	result, err := reader.OperationDirectRead(ctx, connectors.OperationDirectReadRequest{
		Operation: "get_events",
		Config:    runtime,
		Query:     query,
	})
	if err == nil {
		if result.Status != 0 && result.Status != http.StatusOK {
			return asanaEventRead{}, fmt.Errorf("Asana Events operation returned unexpected status %d", result.Status)
		}
		page, decodeErr := decodeAsanaEventPage(result.Body)
		if decodeErr != nil {
			return asanaEventRead{}, decodeErr
		}
		return asanaEventRead{Page: page}, nil
	}
	if result.Status == http.StatusPreconditionFailed && result.Receipt != nil {
		page, decodeErr := decodeAsanaEventPage(result.Receipt.Body)
		if decodeErr != nil {
			return asanaEventRead{}, fmt.Errorf("decode Asana 412 sync-token receipt: %w", decodeErr)
		}
		return asanaEventRead{Page: page, Rebootstrap: true}, nil
	}
	return asanaEventRead{}, fmt.Errorf("read Asana project Events token: %w", err)
}

func readAsanaEventWindow(ctx context.Context, reader connectors.OperationDirectReader, runtime connectors.RuntimeConfig, projectGID, startToken string) ([]asanaEvent, string, string, error) {
	token := startToken
	var events []asanaEvent
	for pageNumber := 1; pageNumber <= asanaEventMaxWindowPages; pageNumber++ {
		read, err := readAsanaProjectEventPage(ctx, reader, runtime, projectGID, token)
		if err != nil {
			return nil, "", "", err
		}
		if read.Rebootstrap {
			return nil, "", read.Page.Sync, nil
		}
		events = append(events, read.Page.Data...)
		if !read.Page.HasMore {
			return events, read.Page.Sync, "", nil
		}
		if read.Page.Sync == token {
			return nil, "", "", fmt.Errorf("Asana Events has_more page did not advance its sync token")
		}
		token = read.Page.Sync
	}
	return nil, "", "", fmt.Errorf("Asana Events token window exceeded the safety limit of %d pages without exhaustion", asanaEventMaxWindowPages)
}

func emitAsanaFullSnapshot(ctx context.Context, request synctransport.SourceRequest, checkpoint synccontract.CheckpointEnvelope, emit func(synctransport.SourcePage) error) error {
	var pending *synctransport.SourcePage
	records := make([]connectors.Record, 0, request.BatchSize)
	emitBatch := func() error {
		if len(records) == 0 {
			return nil
		}
		page := synctransport.SourcePage{Records: append([]connectors.Record(nil), records...), CandidateCheckpoint: checkpoint, DeferCheckpoint: true}
		records = records[:0]
		if pending != nil {
			if err := emit(*pending); err != nil {
				return err
			}
		}
		pending = &page
		return nil
	}
	err := request.Connector.Read(ctx, connectors.ReadRequest{
		Stream:           request.Stream,
		Config:           request.Runtime,
		MaxPages:         0,
		PageDeadline:     request.UnitDeadline,
		ObservePageFetch: request.RecordExtraction,
	}, func(record connectors.Record) error {
		cloned, err := cloneTransportRecord(record)
		if err != nil {
			return err
		}
		records = append(records, cloned)
		if len(records) == request.BatchSize {
			return emitBatch()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("read Asana bootstrap snapshot: %w", err)
	}
	if err := emitBatch(); err != nil {
		return err
	}
	if pending == nil {
		return emit(synctransport.SourcePage{CandidateCheckpoint: checkpoint})
	}
	pending.DeferCheckpoint = false
	return emit(*pending)
}

type asanaEventTarget struct {
	GID        string
	Events     []asanaEvent
	SawDelete  bool
	OnlyRemove bool
}

func hydrateAsanaEventWindow(ctx context.Context, reader connectors.OperationDirectReader, request synctransport.SourceRequest, binding asanaEventStreamBinding, events []asanaEvent, position synccontract.CheckpointPosition) ([]connectors.Record, []synccontract.Tombstone, error) {
	targets := map[string]*asanaEventTarget{}
	for _, event := range events {
		if event.Resource.ResourceType != binding.ResourceType || event.Resource.GID == "" {
			continue
		}
		target := targets[event.Resource.GID]
		if target == nil {
			target = &asanaEventTarget{GID: event.Resource.GID, OnlyRemove: true}
			targets[event.Resource.GID] = target
		}
		target.Events = append(target.Events, event)
		target.SawDelete = target.SawDelete || asanaEventIsDelete(event)
		target.OnlyRemove = target.OnlyRemove && event.Action == "removed"
	}
	gids := make([]string, 0, len(targets))
	for gid := range targets {
		gids = append(gids, gid)
	}
	sort.Strings(gids)
	records := make([]connectors.Record, 0, len(gids))
	tombstones := make([]synccontract.Tombstone, 0)
	for _, gid := range gids {
		target := targets[gid]
		record, missing, err := hydrateAsanaResource(ctx, reader, request.Runtime, binding, gid)
		if err != nil {
			return nil, nil, err
		}
		if !missing {
			records = append(records, record)
			continue
		}
		if !target.SawDelete {
			if target.OnlyRemove {
				continue
			}
			return nil, nil, fmt.Errorf("Asana event resource %s/%s disappeared without a documented deleted action; refusing to infer a tombstone", binding.ResourceType, gid)
		}
		sortAsanaEvents(target.Events)
		var deleted asanaEvent
		for _, event := range target.Events {
			if asanaEventIsDelete(event) {
				deleted = event
				break
			}
		}
		tombstone, ok, err := asanaEventTombstone(deleted, request.PrimaryKey[0], position)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			tombstones = append(tombstones, tombstone)
		}
	}
	return records, tombstones, nil
}

func hydrateAsanaResource(ctx context.Context, reader connectors.OperationDirectReader, runtime connectors.RuntimeConfig, binding asanaEventStreamBinding, gid string) (connectors.Record, bool, error) {
	result, err := reader.OperationDirectRead(ctx, connectors.OperationDirectReadRequest{
		Operation:  binding.Operation,
		Config:     runtime,
		PathParams: map[string]string{binding.PathField: gid},
	})
	if err != nil {
		if result.Status == http.StatusNotFound {
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("hydrate Asana %s %s: %w", binding.ResourceType, gid, err)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("hydrate Asana %s %s: response body is not an object", binding.ResourceType, gid)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("hydrate Asana %s %s: response has no object data", binding.ResourceType, gid)
	}
	record, err := cloneTransportRecord(connectors.Record(data))
	if err != nil {
		return nil, false, err
	}
	if recordGID, _ := record["gid"].(string); strings.TrimSpace(recordGID) != gid {
		return nil, false, fmt.Errorf("hydrate Asana %s %s: response gid does not match event resource", binding.ResourceType, gid)
	}
	return record, false, nil
}

func emitAsanaEventWorksets(records []connectors.Record, tombstones []synccontract.Tombstone, checkpoint synccontract.CheckpointEnvelope, batchSize int, emit func(synctransport.SourcePage) error) error {
	if len(records) == 0 && len(tombstones) == 0 {
		return emit(synctransport.SourcePage{CandidateCheckpoint: checkpoint})
	}
	total := len(records) + len(tombstones)
	consumed := 0
	for len(records) > 0 || len(tombstones) > 0 {
		page := synctransport.SourcePage{CandidateCheckpoint: checkpoint}
		remaining := batchSize
		if take := min(remaining, len(records)); take > 0 {
			page.Records = append([]connectors.Record(nil), records[:take]...)
			records = records[take:]
			remaining -= take
			consumed += take
		}
		if take := min(remaining, len(tombstones)); take > 0 {
			page.Tombstones = append([]synccontract.Tombstone(nil), tombstones[:take]...)
			tombstones = tombstones[take:]
			consumed += take
		}
		page.DeferCheckpoint = consumed < total
		if err := emit(page); err != nil {
			return err
		}
	}
	return nil
}

func asanaEventCheckpoint(resume synccontract.ResumeExpectation, stream, startToken, endToken string, events []asanaEvent) (synccontract.CheckpointEnvelope, error) {
	startToken = strings.TrimSpace(startToken)
	endToken = strings.TrimSpace(endToken)
	if startToken == "" || endToken == "" {
		return synccontract.CheckpointEnvelope{}, fmt.Errorf("Asana Events checkpoint requires non-empty token-window bounds")
	}
	ordered := append([]asanaEvent(nil), events...)
	sortAsanaEvents(ordered)
	digest, err := hashJSON(struct {
		Stream string       `json:"stream"`
		Start  string       `json:"start"`
		End    string       `json:"end"`
		Events []asanaEvent `json:"events"`
	}{Stream: stream, Start: startToken, End: endToken, Events: ordered})
	if err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	observed := true
	end := synccontract.OpaqueToken(endToken)
	tie := synccontract.OpaqueToken(digest)
	checkpoint := synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           resume.Source,
		Mechanism:        asanaEventMechanism,
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "asana_events_sync", Token: append(synccontract.OpaqueToken(nil), end...)},
		Position:         synccontract.CheckpointPosition{Primary: append(synccontract.OpaqueToken(nil), end...), TieBreaker: append(synccontract.OpaqueToken(nil), tie...)},
		PositionObserved: &observed,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: append(synccontract.OpaqueToken(nil), resume.SourceGeneration...),
		SchemaVersion:    "asana-" + stream + "-events-v1",
		ProtocolVersion:  "asana-events-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "asana_event_window", Value: append(synccontract.OpaqueToken(nil), tie...)},
		DedupeWindow: synccontract.DedupeWindow{
			Kind:  "asana_sync_token_window",
			Start: synccontract.OpaqueToken(startToken),
			End:   append(synccontract.OpaqueToken(nil), end...),
		},
		ObservedAt: time.Now().UTC(),
	}
	if err := checkpoint.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, fmt.Errorf("validate Asana Events checkpoint: %w", err)
	}
	return checkpoint, nil
}

func sortAsanaEvents(events []asanaEvent) {
	sort.Slice(events, func(i, j int) bool {
		left, _ := json.Marshal(events[i])
		right, _ := json.Marshal(events[j])
		return string(left) < string(right)
	})
}

// decodeAsanaEventPage accepts the engine's bounded JSON projection from both
// a successful response and a 412 receipt. It deliberately does not parse an
// error message: the pinned provider contract makes the top-level sync field
// the only bootstrap/recovery authority.
func decodeAsanaEventPage(body any) (asanaEventPage, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return asanaEventPage{}, fmt.Errorf("encode Asana Events response: %w", err)
	}
	var page asanaEventPage
	if err := json.Unmarshal(raw, &page); err != nil {
		return asanaEventPage{}, fmt.Errorf("decode Asana Events response: %w", err)
	}
	page.Sync = strings.TrimSpace(page.Sync)
	if page.Sync == "" {
		return asanaEventPage{}, fmt.Errorf("Asana Events response did not include the required sync token")
	}
	for index := range page.Data {
		page.Data[index].Action = strings.ToLower(strings.TrimSpace(page.Data[index].Action))
		page.Data[index].Resource.GID = strings.TrimSpace(page.Data[index].Resource.GID)
		page.Data[index].Resource.ResourceType = strings.ToLower(strings.TrimSpace(page.Data[index].Resource.ResourceType))
	}
	return page, nil
}

func asanaEventIsDelete(event asanaEvent) bool {
	return event.Action == "deleted"
}

// asanaEventTombstone converts only the provider's documented deleted action.
// In particular, removed describes a relationship transition and must never
// be promoted to a resource deletion. The event identity is a deterministic
// system-owned digest of the complete provider event plus the token window;
// Asana's EventResponse does not expose an event gid.
func asanaEventTombstone(event asanaEvent, primaryKey string, position synccontract.CheckpointPosition) (synccontract.Tombstone, bool, error) {
	if !asanaEventIsDelete(event) {
		return synccontract.Tombstone{}, false, nil
	}
	if strings.TrimSpace(primaryKey) == "" {
		return synccontract.Tombstone{}, false, fmt.Errorf("Asana tombstone primary key field is required")
	}
	if event.Resource.GID == "" {
		return synccontract.Tombstone{}, false, fmt.Errorf("Asana deleted event has no resource gid")
	}
	key, err := json.Marshal(connectors.Record{primaryKey: event.Resource.GID})
	if err != nil {
		return synccontract.Tombstone{}, false, fmt.Errorf("encode Asana tombstone key: %w", err)
	}
	eventBytes, err := json.Marshal(struct {
		Event    asanaEvent                      `json:"event"`
		Position synccontract.CheckpointPosition `json:"position"`
	}{Event: event, Position: position})
	if err != nil {
		return synccontract.Tombstone{}, false, fmt.Errorf("encode Asana tombstone identity: %w", err)
	}
	digest := sha256.Sum256(eventBytes)
	tombstone := synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken(hex.EncodeToString(digest[:])),
		Key:         key,
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position:    position.Clone(),
	}
	if err := tombstone.Validate(); err != nil {
		return synccontract.Tombstone{}, false, fmt.Errorf("validate Asana tombstone: %w", err)
	}
	return tombstone, true, nil
}
