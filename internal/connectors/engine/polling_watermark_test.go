package engine

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

type scriptedPollingWatermarkSource struct {
	pages            []PollingWatermarkPage
	calls            []PollingWatermarkPageRequest
	callErr          error
	onFetch          func()
	nextPage         int
	dataRequests     int
	deletionRequests int
}

func (s *scriptedPollingWatermarkSource) FetchPollingWatermarkPage(ctx context.Context, req PollingWatermarkPageRequest) (PollingWatermarkPage, error) {
	if err := ctx.Err(); err != nil {
		return PollingWatermarkPage{}, err
	}
	if err := req.RequestBudget.Consume(ctx); err != nil {
		return PollingWatermarkPage{}, err
	}
	s.dataRequests++
	s.calls = append(s.calls, req)
	if s.onFetch != nil {
		s.onFetch()
	}
	if s.callErr != nil {
		return PollingWatermarkPage{}, s.callErr
	}
	if req.DeletionEndpoint != nil {
		if err := req.RequestBudget.Consume(ctx); err != nil {
			return PollingWatermarkPage{}, err
		}
		s.deletionRequests++
	}
	if s.nextPage >= len(s.pages) {
		return PollingWatermarkPage{}, nil
	}
	page := s.pages[s.nextPage]
	s.nextPage++
	return page, nil
}

type pollingWatermarkPageSourceFunc func(context.Context, PollingWatermarkPageRequest) (PollingWatermarkPage, error)

func (f pollingWatermarkPageSourceFunc) FetchPollingWatermarkPage(ctx context.Context, req PollingWatermarkPageRequest) (PollingWatermarkPage, error) {
	return f(ctx, req)
}

type recordingChangefeedCheckpointCommitter struct {
	states  []map[string]string
	failAt  int
	commits int
}

func (c *recordingChangefeedCheckpointCommitter) CommitChangefeedCheckpoint(_ context.Context, state map[string]string) error {
	c.commits++
	if c.failAt > 0 && c.commits == c.failAt {
		return errors.New("simulated checkpoint persistence crash")
	}
	copy := make(map[string]string, len(state))
	for key, value := range state {
		copy[key] = value
	}
	c.states = append(c.states, copy)
	return nil
}

func pollingWatermarkScanFromState(t *testing.T, state map[string]string) pollingWatermarkScanCheckpoint {
	t.Helper()
	raw, found := state[pollingWatermarkScanStateKey]
	if !found {
		t.Fatalf("checkpoint state = %+v, want scan cursor", state)
	}
	var scan pollingWatermarkScanCheckpoint
	if err := json.Unmarshal([]byte(raw), &scan); err != nil {
		t.Fatalf("unmarshal scan cursor %q: %v", raw, err)
	}
	return scan
}

func pollingWatermarkFrontiersFromState(t *testing.T, state map[string]string) pollingWatermarkFrontiersCheckpoint {
	t.Helper()
	raw, found := state[pollingWatermarkFrontiersStateKey]
	if !found {
		t.Fatalf("checkpoint state = %+v, want source frontiers", state)
	}
	var frontiers pollingWatermarkFrontiersCheckpoint
	if err := json.Unmarshal([]byte(raw), &frontiers); err != nil {
		t.Fatalf("unmarshal source frontiers %q: %v", raw, err)
	}
	return frontiers
}

func loadPollingWatermarkTestBundle(t *testing.T) Bundle {
	t.Helper()
	bundle, err := Load(os.DirFS("testdata/bundles"), "polling-watermark-demo")
	if err != nil {
		t.Fatalf("Load polling test bundle: %v", err)
	}
	return bundle
}

func newPollingWatermarkTestConnector(t *testing.T, source PollingWatermarkPageSource) *PollingWatermarkConnector {
	t.Helper()
	connector, err := NewPollingWatermarkConnector(loadPollingWatermarkTestBundle(t), nil, source)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	return connector
}

func TestPollingWatermarkTestBundlePromotesCDCOnlyWithRegisteredExecutor(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	legacy := New(bundle, nil)
	if got := connectors.HasImplementedChangefeed(legacy, bundle.Changefeed); got {
		t.Fatal("declaration-only engine connector advertised CDC")
	}

	source := &scriptedPollingWatermarkSource{}
	connector := newPollingWatermarkTestConnector(t, source)
	if got := connectors.HasImplementedChangefeed(connector, bundle.Changefeed); !got {
		t.Fatal("matching polling-watermark executor did not promote the test bundle to CDC")
	}
	definition, ok := connectors.DefinitionOf(connector)
	if !ok || !definition.Capabilities.CDC {
		t.Fatalf("DefinitionOf() = %+v, %t; want CDC capability", definition, ok)
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(connector)
	if got := registry.List()[0].Capabilities.CDC; !got {
		t.Fatal("registry list did not derive CDC from the matching executor")
	}
	if got := registry.CatalogEntries()[0].Capabilities.CDC; !got {
		t.Fatal("registry catalog did not derive CDC from the matching executor")
	}
	if got := connectors.ManifestOf(connector).Metadata.Capabilities.CDC; !got {
		t.Fatal("manifest did not derive CDC from the matching executor")
	}
}

func TestPollingWatermarkBundleRejectsEveryMissingCheckpointField(t *testing.T) {
	for _, missing := range []string{"kind", "keys", "commit_after", "on_invalid"} {
		t.Run(missing, func(t *testing.T) {
			bundle := fullValidBundleFS("acme")
			var declaration map[string]any
			if err := json.Unmarshal([]byte(pollingWatermarkChangefeedJSON), &declaration); err != nil {
				t.Fatalf("unmarshal declaration: %v", err)
			}
			checkpoint := declaration["checkpoint"].(map[string]any)
			delete(checkpoint, missing)
			raw, err := json.Marshal(declaration)
			if err != nil {
				t.Fatalf("marshal declaration: %v", err)
			}
			bundle["acme/changefeed.json"] = &fstest.MapFile{Data: raw}
			if _, err := Load(bundle, "acme"); err == nil {
				t.Fatalf("Load accepted polling changefeed missing checkpoint %s", missing)
			}
		})
	}
}

func TestPollingWatermarkConnectorRejectsUndeclaredStream(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.Streams = []string{"missing"}
	_, err := NewPollingWatermarkConnector(bundle, nil, &scriptedPollingWatermarkSource{})
	if err == nil || !strings.Contains(err.Error(), "stream \"missing\" not found") {
		t.Fatalf("NewPollingWatermarkConnector error = %v, want missing declared stream rejection", err)
	}
}

func TestPollingWatermarkExecutorReplaysTieAtPageBoundary(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{
		{Records: []connectors.Record{
			{"id": "a", "updated_at": "2026-08-06T09:59:00Z"},
			{"id": "b", "updated_at": "2026-08-06T10:00:00Z"},
		}, More: true},
		{Records: []connectors.Record{
			{"id": "b", "updated_at": "2026-08-06T10:00:00Z"},
			{"id": "c", "updated_at": "2026-08-06T10:00:00Z"},
		}, More: false},
	}}
	connector := newPollingWatermarkTestConnector(t, source)
	committer := &recordingChangefeedCheckpointCommitter{}
	var got []string
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		got = append(got, event.Record["id"].(string))
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if joined := strings.Join(got, ","); joined != "a,b,b,c" {
		t.Fatalf("records = %s, want inclusive replay at tie boundary", joined)
	}
	if len(source.calls) != 2 || !source.calls[1].Inclusive || source.calls[1].After == nil || source.calls[1].After.TieBreaker != "b" {
		t.Fatalf("second source request = %+v, want inclusive tuple checkpoint after b", source.calls)
	}
	if source.calls[0].PageSize != 2 {
		t.Fatalf("declared page size = %d, want 2", source.calls[0].PageSize)
	}
	if len(committer.states) != 2 || committer.states[1]["updated_at"] != "2026-08-06T10:00:00Z" || committer.states[1]["id"] != "c" {
		t.Fatalf("committed checkpoints = %+v, want final tuple", committer.states)
	}
}

func TestPollingWatermarkExecutorOrdersNumericTieBreakers(t *testing.T) {
	t.Run("ascending numeric values", func(t *testing.T) {
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{Records: []connectors.Record{
			{"id": 2, "updated_at": "2026-08-06T10:00:00Z"},
			{"id": json.Number("10"), "updated_at": "2026-08-06T10:00:00Z"},
		}}}}
		connector := newPollingWatermarkTestConnector(t, source)
		committer := &recordingChangefeedCheckpointCommitter{}
		var emitted int
		err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: committer,
		}, func(connectors.CDCEvent) error {
			emitted++
			return nil
		})
		if err != nil {
			t.Fatalf("ReadCDC: %v", err)
		}
		if emitted != 2 || len(committer.states) != 1 || committer.states[0]["id"] != "10" {
			t.Fatalf("emitted=%d checkpoints=%+v, want ordered numeric tie breakers through 10", emitted, committer.states)
		}
	})

	t.Run("descending numeric values do not emit the invalid record", func(t *testing.T) {
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{Records: []connectors.Record{
			{"id": json.Number("10"), "updated_at": "2026-08-06T10:00:00Z"},
			{"id": 2, "updated_at": "2026-08-06T10:00:00Z"},
		}}}}
		connector := newPollingWatermarkTestConnector(t, source)
		committer := &recordingChangefeedCheckpointCommitter{}
		var emitted int
		err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: committer,
		}, func(connectors.CDCEvent) error {
			emitted++
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), "not ordered") {
			t.Fatalf("ReadCDC error = %v, want numeric ordering rejection", err)
		}
		if emitted != 0 || len(committer.states) != 0 {
			t.Fatalf("emitted=%d checkpoints=%+v, want no delivery or checkpoint from an unordered page", emitted, committer.states)
		}
	})
}

func TestPollingWatermarkExecutorAppliesDeclaredSafetyLag(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{}}}
	connector := newPollingWatermarkTestConnector(t, source)
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               map[string]string{"updated_at": "2026-08-06T09:59:00Z", "id": "already-committed"},
		CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
	}, func(connectors.CDCEvent) error { return nil })
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if len(source.calls) != 1 || source.calls[0].After == nil {
		t.Fatalf("source calls = %+v, want a lagged lower-bound request", source.calls)
	}
	if got := source.calls[0].After.Watermark; got != "2026-08-06T09:57:00Z" {
		t.Fatalf("lagged watermark = %q, want 2026-08-06T09:57:00Z", got)
	}
	if got := source.calls[0].After.TieBreaker; got != "" {
		t.Fatalf("lagged request tie breaker = %q, want empty so late records are replayed", got)
	}
}

func TestPollingWatermarkExecutorSafetyLagNeverRegressesCheckpoint(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{
		{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T09:59:00Z"}}, More: true},
		{Records: []connectors.Record{{"id": "next", "updated_at": "2026-08-06T10:01:00Z"}}},
	}}
	connector := newPollingWatermarkTestConnector(t, source)
	committer := &recordingChangefeedCheckpointCommitter{}
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               map[string]string{"updated_at": "2026-08-06T10:00:00Z", "id": "committed"},
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error { return nil })
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if len(source.calls) != 2 || source.calls[1].After == nil || source.calls[1].After.Watermark != "2026-08-06T09:59:00Z" || source.calls[1].After.TieBreaker != "a" {
		t.Fatalf("source calls = %+v, want query cursor to advance through the safety-lag overlap", source.calls)
	}
	if len(committer.states) != 2 || committer.states[0]["updated_at"] != "2026-08-06T10:00:00Z" || committer.states[0]["id"] != "committed" || committer.states[1]["updated_at"] != "2026-08-06T10:01:00Z" || committer.states[1]["id"] != "next" {
		t.Fatalf("committed checkpoints = %+v, want scan progress without durable regression", committer.states)
	}
	if scan := pollingWatermarkScanFromState(t, committer.states[0]); !scan.Active || scan.Watermark != "2026-08-06T09:59:00Z" || scan.TieBreaker != "a" {
		t.Fatalf("first scan cursor = %+v, want active overlap cursor through a", scan)
	}
	if scan := pollingWatermarkScanFromState(t, committer.states[1]); scan.Active {
		t.Fatalf("final scan cursor = %+v, want inactive cursor after overlap completion", scan)
	}
}

func TestPollingWatermarkExecutorPersistsAndResetsSafetyLagScanCursor(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.SafetyLagSeconds = 120
	bundle.Changefeed.PollingWatermark.MaxPages = 1
	committer := &recordingChangefeedCheckpointCommitter{}
	state := map[string]string{"updated_at": "2026-08-06T10:00:00Z", "id": "committed"}

	firstSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T09:59:00Z"}},
		More:    true,
	}}}
	first, err := NewPollingWatermarkConnector(bundle, nil, firstSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	err = first.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               state,
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error { return nil })
	var stop *PollingWatermarkResumableStopError
	if !errors.As(err, &stop) || stop.Reason != PollingWatermarkStopReasonMaxPages {
		t.Fatalf("first ReadCDC error = %v, want max-pages resumable stop", err)
	}
	if len(firstSource.calls) != 1 || firstSource.calls[0].After == nil || firstSource.calls[0].After.Watermark != "2026-08-06T09:58:00Z" {
		t.Fatalf("first source calls = %+v, want safety-lag boundary", firstSource.calls)
	}
	if len(committer.states) != 1 || committer.states[0]["updated_at"] != "2026-08-06T10:00:00Z" || committer.states[0]["id"] != "committed" {
		t.Fatalf("first checkpoint = %+v, want unchanged durable high-water", committer.states)
	}
	if scan := pollingWatermarkScanFromState(t, committer.states[0]); !scan.Active || scan.Watermark != "2026-08-06T09:59:00Z" || scan.TieBreaker != "a" {
		t.Fatalf("first scan cursor = %+v, want persisted active cursor", scan)
	}

	secondSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		Records: []connectors.Record{
			{"id": "a", "updated_at": "2026-08-06T09:59:00Z"},
			{"id": "b", "updated_at": "2026-08-06T09:59:00Z"},
		},
	}}}
	second, err := NewPollingWatermarkConnector(bundle, nil, secondSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	err = second.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               committer.states[0],
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error { return nil })
	if err != nil {
		t.Fatalf("second ReadCDC: %v", err)
	}
	if len(secondSource.calls) != 1 || secondSource.calls[0].After == nil || secondSource.calls[0].After.Watermark != "2026-08-06T09:59:00Z" || secondSource.calls[0].After.TieBreaker != "a" {
		t.Fatalf("second source calls = %+v, want resumed scan cursor", secondSource.calls)
	}
	if len(committer.states) != 2 || committer.states[1]["updated_at"] != "2026-08-06T10:00:00Z" || committer.states[1]["id"] != "committed" {
		t.Fatalf("second checkpoint = %+v, want unchanged durable high-water", committer.states)
	}
	if scan := pollingWatermarkScanFromState(t, committer.states[1]); scan.Active {
		t.Fatalf("second scan cursor = %+v, want completed overlap cursor", scan)
	}

	thirdSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		Records: []connectors.Record{{"id": "late", "updated_at": "2026-08-06T09:58:30Z"}},
	}}}
	third, err := NewPollingWatermarkConnector(bundle, nil, thirdSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	var delivered []string
	err = third.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               committer.states[1],
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		delivered = append(delivered, event.Record["id"].(string))
		return nil
	})
	if err != nil {
		t.Fatalf("third ReadCDC: %v", err)
	}
	if len(thirdSource.calls) != 1 || thirdSource.calls[0].After == nil || thirdSource.calls[0].After.Watermark != "2026-08-06T09:58:00Z" || thirdSource.calls[0].After.TieBreaker != "" {
		t.Fatalf("third source calls = %+v, want reset safety-lag boundary", thirdSource.calls)
	}
	if strings.Join(delivered, ",") != "late" {
		t.Fatalf("delivered = %q, want late overlap record", delivered)
	}
}

func TestPollingWatermarkExecutorPreservesDeletionSafetyLagStartBeforeFirstTombstone(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.SoftDelete = nil
	bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{
		Path:        "/widgets/deletions",
		RecordsPath: "data",
	}
	bundle.Changefeed.PollingWatermark.MaxPages = 1
	bundle.Changefeed.PollingWatermark.RequestBudget = 2
	state := map[string]string{"updated_at": "2026-08-06T10:00:00Z", "id": "committed"}
	committer := &recordingChangefeedCheckpointCommitter{}

	firstSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		Records:     []connectors.Record{{"id": "primary", "updated_at": "2026-08-06T09:59:00Z"}},
		PrimaryMore: true,
	}}}
	first, err := NewPollingWatermarkConnector(bundle, nil, firstSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	err = first.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               state,
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error { return nil })
	var stop *PollingWatermarkResumableStopError
	if !errors.As(err, &stop) || stop.Reason != PollingWatermarkStopReasonMaxPages {
		t.Fatalf("first ReadCDC error = %v, want max-pages resumable stop", err)
	}
	if len(committer.states) != 1 {
		t.Fatalf("committed checkpoints = %+v, want primary-only checkpoint", committer.states)
	}
	frontiers := pollingWatermarkFrontiersFromState(t, committer.states[0])
	if frontiers.Deletion == nil || !frontiers.Deletion.QuerySet || frontiers.Deletion.Query == nil || frontiers.Deletion.Query.Watermark != "2026-08-06T09:58:00Z" || frontiers.Deletion.Query.TieBreaker != "" {
		t.Fatalf("source frontiers = %+v, want persisted deletion safety-lag start", frontiers)
	}
	if frontiers.Deletion.Watermark != "" || frontiers.Deletion.TieBreaker != "" {
		t.Fatalf("source frontiers = %+v, want no durable deletion frontier", frontiers)
	}

	rejectedSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		DeletionRecords: []connectors.Record{{"id": "late", "updated_at": "2026-08-06T09:58:30Z"}},
	}}}
	rejected, err := NewPollingWatermarkConnector(bundle, nil, rejectedSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	var rejectedEvents []string
	err = rejected.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               committer.states[0],
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		rejectedEvents = append(rejectedEvents, event.Operation+":"+event.Record["id"].(string))
		return errors.New("destination rejected late deletion")
	})
	if err == nil || !strings.Contains(err.Error(), "durably accept") {
		t.Fatalf("restart ReadCDC error = %v, want destination acknowledgement failure", err)
	}
	if strings.Join(rejectedEvents, ",") != "delete:late" {
		t.Fatalf("rejected events = %q, want late deletion inside the safety lag", rejectedEvents)
	}
	if len(committer.states) != 1 {
		t.Fatalf("committed checkpoints = %+v, want no checkpoint before late deletion acknowledgement", committer.states)
	}
	if len(rejectedSource.calls) != 1 || rejectedSource.calls[0].After == nil || rejectedSource.calls[0].After.Watermark != "2026-08-06T09:59:00Z" || rejectedSource.calls[0].DeletionAfter == nil || rejectedSource.calls[0].DeletionAfter.Watermark != "2026-08-06T09:58:00Z" || rejectedSource.calls[0].DeletionAfter.TieBreaker != "" {
		t.Fatalf("restart source calls = %+v, want independent primary and deletion queries", rejectedSource.calls)
	}

	retrySource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		DeletionRecords: []connectors.Record{{"id": "late", "updated_at": "2026-08-06T09:58:30Z"}},
	}}}
	retry, err := NewPollingWatermarkConnector(bundle, nil, retrySource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	var delivered []string
	err = retry.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               committer.states[0],
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		delivered = append(delivered, event.Operation+":"+event.Record["id"].(string))
		return nil
	})
	if err != nil {
		t.Fatalf("retry ReadCDC: %v", err)
	}
	if strings.Join(delivered, ",") != "delete:late" {
		t.Fatalf("delivered = %q, want replayed late deletion", delivered)
	}
	if len(retrySource.calls) != 1 || retrySource.calls[0].DeletionAfter == nil || retrySource.calls[0].DeletionAfter.Watermark != "2026-08-06T09:58:00Z" || retrySource.calls[0].DeletionAfter.TieBreaker != "" {
		t.Fatalf("retry source calls = %+v, want replay-safe deletion start", retrySource.calls)
	}
	if len(committer.states) != 2 {
		t.Fatalf("committed checkpoints = %+v, want accepted late deletion checkpoint", committer.states)
	}
	if frontiers := pollingWatermarkFrontiersFromState(t, committer.states[1]); frontiers.Deletion == nil || frontiers.Deletion.Watermark != "2026-08-06T09:58:30Z" || frontiers.Deletion.TieBreaker != "late" {
		t.Fatalf("source frontiers = %+v, want durable late deletion frontier", frontiers)
	}
}

func TestPollingWatermarkExecutorReappliesDeletionSafetyLagAfterDurableTombstone(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.SoftDelete = nil
	bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{
		Path:        "/widgets/deletions",
		RecordsPath: "data",
	}
	bundle.Changefeed.PollingWatermark.RequestBudget = 2
	committer := &recordingChangefeedCheckpointCommitter{}

	firstSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		DeletionRecords: []connectors.Record{{"id": "committed", "updated_at": "2026-08-06T10:00:00Z"}},
	}}}
	first, err := NewPollingWatermarkConnector(bundle, nil, firstSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	err = first.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error { return nil })
	if err != nil {
		t.Fatalf("first ReadCDC: %v", err)
	}
	if len(committer.states) != 1 {
		t.Fatalf("committed checkpoints = %+v, want durable deletion checkpoint", committer.states)
	}
	frontiers := pollingWatermarkFrontiersFromState(t, committer.states[0])
	if frontiers.Deletion == nil || frontiers.Deletion.Watermark != "2026-08-06T10:00:00Z" || frontiers.Deletion.TieBreaker != "committed" {
		t.Fatalf("source frontiers = %+v, want durable deletion frontier", frontiers)
	}
	if frontiers.Deletion.QuerySet || frontiers.Deletion.Query != nil {
		t.Fatalf("source frontiers = %+v, want durable deletion state without a query bypass", frontiers)
	}
	frontiers.Deletion.QuerySet = true
	frontiers.Deletion.Query = &pollingWatermarkPositionCheckpoint{Watermark: "2026-08-06T10:00:00Z", TieBreaker: "committed"}
	raw, err := json.Marshal(frontiers)
	if err != nil {
		t.Fatalf("marshal legacy deletion frontier: %v", err)
	}
	legacyState := clonePollingWatermarkState(committer.states[0])
	legacyState[pollingWatermarkFrontiersStateKey] = string(raw)

	secondSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		DeletionRecords: []connectors.Record{{"id": "late", "updated_at": "2026-08-06T09:59:00Z"}},
	}}}
	second, err := NewPollingWatermarkConnector(bundle, nil, secondSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	var delivered []string
	err = second.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               legacyState,
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		delivered = append(delivered, event.Operation+":"+event.Record["id"].(string))
		return nil
	})
	if err != nil {
		t.Fatalf("second ReadCDC: %v", err)
	}
	if len(secondSource.calls) != 1 || secondSource.calls[0].DeletionAfter == nil || secondSource.calls[0].DeletionAfter.Watermark != "2026-08-06T09:58:00Z" || secondSource.calls[0].DeletionAfter.TieBreaker != "" {
		t.Fatalf("restart source calls = %+v, want deletion safety-lag boundary", secondSource.calls)
	}
	if strings.Join(delivered, ",") != "delete:late" {
		t.Fatalf("delivered = %q, want late deletion inside the safety lag", delivered)
	}
}

func TestPollingWatermarkExecutorPersistsNumericScanCursorOrderingMetadata(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.MaxPages = 1
	committer := &recordingChangefeedCheckpointCommitter{}
	firstSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		Records: []connectors.Record{{"id": json.Number("2"), "updated_at": "2026-08-06T10:00:00Z"}},
		More:    true,
	}}}
	first, err := NewPollingWatermarkConnector(bundle, nil, firstSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	err = first.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               map[string]string{"updated_at": "2026-08-06T10:00:00Z", "id": "10"},
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error { return nil })
	var stop *PollingWatermarkResumableStopError
	if !errors.As(err, &stop) || stop.Reason != PollingWatermarkStopReasonMaxPages {
		t.Fatalf("first ReadCDC error = %v, want max-pages resumable stop", err)
	}
	if len(committer.states) != 1 {
		t.Fatalf("committed checkpoints = %+v, want active numeric scan cursor", committer.states)
	}
	if scan := pollingWatermarkScanFromState(t, committer.states[0]); !scan.Active || scan.TieBreaker != "2" || !scan.TieBreakerNumber {
		t.Fatalf("scan cursor = %+v, want exact numeric provider text with ordering metadata", scan)
	}

	secondSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{}}}
	second, err := NewPollingWatermarkConnector(bundle, nil, secondSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	err = second.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               committer.states[0],
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error { return nil })
	if err != nil {
		t.Fatalf("second ReadCDC: %v", err)
	}
	if len(secondSource.calls) != 1 || secondSource.calls[0].After == nil || secondSource.calls[0].After.TieBreaker != "2" || !secondSource.calls[0].After.tieBreakerNumber {
		t.Fatalf("second source calls = %+v, want numeric scan cursor resumption", secondSource.calls)
	}
}

func TestPollingWatermarkExecutorLeavesInitialSnapshotBoundaryToSource(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{}}}
	connector := newPollingWatermarkTestConnector(t, source)
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
	}, func(connectors.CDCEvent) error { return nil })
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if len(source.calls) != 1 || source.calls[0].After != nil {
		t.Fatalf("initial source request = %+v, want no implicit current-time boundary", source.calls)
	}
}

func TestPollingWatermarkExecutorOnlyEmitsDeclaredSoftDeletes(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{Records: []connectors.Record{
		{"id": "live", "updated_at": "2026-08-06T10:00:00Z", "deleted": false},
		{"id": "gone", "updated_at": "2026-08-06T10:01:00Z", "deleted": true},
	}}}}
	connector := newPollingWatermarkTestConnector(t, source)
	var operations []string
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
	}, func(event connectors.CDCEvent) error {
		operations = append(operations, event.Operation)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if joined := strings.Join(operations, ","); joined != "upsert,delete" {
		t.Fatalf("operations = %q, want soft-delete tombstone only", joined)
	}
}

func TestPollingWatermarkExecutorTreatsDecimalZeroSoftDeleteAsLive(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{Records: []connectors.Record{
		{"id": "live", "updated_at": "2026-08-06T10:00:00Z", "deleted": json.Number("0.0")},
		{"id": "gone", "updated_at": "2026-08-06T10:01:00Z", "deleted": json.Number("1.0")},
	}}}}
	connector := newPollingWatermarkTestConnector(t, source)
	var operations []string
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
	}, func(event connectors.CDCEvent) error {
		operations = append(operations, event.Operation)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if joined := strings.Join(operations, ","); joined != "upsert,delete" {
		t.Fatalf("operations = %q, want decimal zero to remain live", joined)
	}
}

func TestPollingWatermarkExecutorUsesDeclaredDeletionEndpoint(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.SoftDelete = nil
	bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{
		Path:        "/widgets/deletions",
		RecordsPath: "data",
	}
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{DeletionRecords: []connectors.Record{
		{"id": "gone", "updated_at": "2026-08-06T10:01:00Z"},
	}}}}
	connector, err := NewPollingWatermarkConnector(bundle, nil, source)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	var operation string
	err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
	}, func(event connectors.CDCEvent) error {
		operation = event.Operation
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if operation != "delete" {
		t.Fatalf("operation = %q, want delete from declared deletion endpoint", operation)
	}
	if len(source.calls) != 1 || source.calls[0].DeletionEndpoint == nil || source.calls[0].DeletionEndpoint.Path != "/widgets/deletions" {
		t.Fatalf("source request = %+v, want declared deletion endpoint", source.calls)
	}
}

func TestPollingWatermarkExecutorMergesPrimaryAndDeletionRecordsByPosition(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.SoftDelete = nil
	bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{
		Path:        "/widgets/deletions",
		RecordsPath: "data",
	}
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		Records: []connectors.Record{
			{"id": "a", "updated_at": "2026-08-06T10:00:00Z"},
			{"id": "c", "updated_at": "2026-08-06T10:02:00Z"},
		},
		DeletionRecords: []connectors.Record{{"id": "gone", "updated_at": "2026-08-06T10:01:00Z"}},
	}}}
	connector, err := NewPollingWatermarkConnector(bundle, nil, source)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	committer := &recordingChangefeedCheckpointCommitter{}
	var delivered []string
	err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		delivered = append(delivered, event.Operation+":"+event.Record["id"].(string))
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if joined := strings.Join(delivered, ","); joined != "upsert:a,delete:gone,upsert:c" {
		t.Fatalf("delivery order = %q, want merged tuple order", joined)
	}
	if len(committer.states) != 1 || committer.states[0]["updated_at"] != "2026-08-06T10:02:00Z" || committer.states[0]["id"] != "c" {
		t.Fatalf("checkpoints = %+v, want final merged tuple", committer.states)
	}
}

func TestPollingWatermarkExecutorMaintainsIndependentDeletionFrontier(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.SoftDelete = nil
	bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{
		Path:        "/widgets/deletions",
		RecordsPath: "data",
	}
	bundle.Changefeed.PollingWatermark.MaxPages = 2
	bundle.Changefeed.PollingWatermark.RequestBudget = 4
	var requests []PollingWatermarkPageRequest
	call := 0
	source := pollingWatermarkPageSourceFunc(func(ctx context.Context, req PollingWatermarkPageRequest) (PollingWatermarkPage, error) {
		if err := req.RequestBudget.Consume(ctx); err != nil {
			return PollingWatermarkPage{}, err
		}
		if err := req.RequestBudget.Consume(ctx); err != nil {
			return PollingWatermarkPage{}, err
		}
		requests = append(requests, req)
		call++
		switch call {
		case 1:
			return PollingWatermarkPage{
				Records: []connectors.Record{
					{"id": "p10", "updated_at": "2026-08-06T10:00:00Z"},
					{"id": "p100", "updated_at": "2026-08-06T11:30:00Z"},
				},
				DeletionRecords: []connectors.Record{
					{"id": "d11", "updated_at": "2026-08-06T10:01:00Z"},
					{"id": "d12", "updated_at": "2026-08-06T10:02:00Z"},
				},
				DeletionMore: true,
			}, nil
		case 2:
			return PollingWatermarkPage{
				Records: []connectors.Record{
					{"id": "p10", "updated_at": "2026-08-06T10:00:00Z"},
					{"id": "p100", "updated_at": "2026-08-06T11:30:00Z"},
				},
				DeletionRecords: []connectors.Record{
					{"id": "d12", "updated_at": "2026-08-06T10:02:00Z"},
					{"id": "d13", "updated_at": "2026-08-06T10:03:00Z"},
				},
			}, nil
		default:
			return PollingWatermarkPage{}, errors.New("unexpected polling request")
		}
	})
	connector, err := NewPollingWatermarkConnector(bundle, nil, source)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	committer := &recordingChangefeedCheckpointCommitter{}
	var delivered []string
	err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		delivered = append(delivered, event.Record["id"].(string))
		return nil
	})
	if err != nil {
		t.Fatalf("ReadCDC: %v", err)
	}
	if len(requests) != 2 || requests[1].After == nil || requests[1].DeletionAfter == nil || requests[1].After.TieBreaker != "p10" || requests[1].DeletionAfter.TieBreaker != "d12" {
		t.Fatalf("source requests = %+v, want independent primary and deletion frontiers", requests)
	}
	if !strings.Contains(strings.Join(delivered, ","), "d13") {
		t.Fatalf("delivered = %q, want deletion past the primary physical page", delivered)
	}
	if len(committer.states) != 2 || committer.states[1]["id"] != "p100" {
		t.Fatalf("committed checkpoints = %+v, want primary frontier through p100", committer.states)
	}
	if frontiers := pollingWatermarkFrontiersFromState(t, committer.states[1]); frontiers.Deletion == nil || frontiers.Deletion.TieBreaker != "d13" {
		t.Fatalf("source frontiers = %+v, want deletion frontier through d13", frontiers)
	}
}

func TestPollingWatermarkExecutorKeepsOpaqueDeletionCursorIndependent(t *testing.T) {
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.Checkpoint.Keys = []string{"cursor", "id"}
	bundle.Changefeed.PollingWatermark.Watermark = connectors.PollingWatermarkValue{Kind: "opaque_cursor", Path: "cursor"}
	bundle.Changefeed.PollingWatermark.SafetyLagSeconds = 0
	bundle.Changefeed.PollingWatermark.SoftDelete = nil
	bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{
		Path:        "/widgets/deletions",
		RecordsPath: "data",
	}
	bundle.Changefeed.PollingWatermark.RequestBudget = 2
	firstSource := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{
		Records:         []connectors.Record{{"id": "primary", "cursor": "z-primary"}},
		DeletionRecords: []connectors.Record{{"id": "deletion", "cursor": "a-deletion"}},
	}}}
	first, err := NewPollingWatermarkConnector(bundle, nil, firstSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	committer := &recordingChangefeedCheckpointCommitter{}
	var delivered []string
	err = first.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: committer,
	}, func(event connectors.CDCEvent) error {
		delivered = append(delivered, event.Operation+":"+event.Record["id"].(string))
		return nil
	})
	if err != nil {
		t.Fatalf("first ReadCDC: %v", err)
	}
	if got := strings.Join(delivered, ","); got != "upsert:primary,delete:deletion" {
		t.Fatalf("delivery = %q, want source-local opaque delivery", got)
	}
	if len(committer.states) != 1 {
		t.Fatalf("committed checkpoints = %+v, want independent opaque frontiers", committer.states)
	}
	if frontiers := pollingWatermarkFrontiersFromState(t, committer.states[0]); frontiers.Deletion == nil || frontiers.Deletion.Watermark != "a-deletion" || frontiers.Deletion.TieBreaker != "deletion" {
		t.Fatalf("source frontiers = %+v, want verbatim opaque deletion cursor", frontiers)
	}

	secondSource := pollingWatermarkPageSourceFunc(func(ctx context.Context, req PollingWatermarkPageRequest) (PollingWatermarkPage, error) {
		if err := req.RequestBudget.Consume(ctx); err != nil {
			return PollingWatermarkPage{}, err
		}
		if err := req.RequestBudget.Consume(ctx); err != nil {
			return PollingWatermarkPage{}, err
		}
		if req.After == nil || req.After.Watermark != "z-primary" || req.DeletionAfter == nil || req.DeletionAfter.Watermark != "a-deletion" {
			return PollingWatermarkPage{}, errors.New("opaque source frontier was not restored verbatim")
		}
		return PollingWatermarkPage{}, nil
	})
	second, err := NewPollingWatermarkConnector(bundle, nil, secondSource)
	if err != nil {
		t.Fatalf("NewPollingWatermarkConnector: %v", err)
	}
	err = second.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		State:               committer.states[0],
		CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
	}, func(connectors.CDCEvent) error { return nil })
	if err != nil {
		t.Fatalf("second ReadCDC: %v", err)
	}
}

func TestPollingWatermarkExecutorDoesNotSeedOpaqueDeletionFromPrimaryCheckpoint(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		state map[string]string
	}{
		{
			name:  "missing deletion state",
			state: map[string]string{"cursor": "primary-cursor", "id": "primary"},
		},
		{
			name: "legacy primary-derived query",
			state: map[string]string{
				"cursor":                          "primary-cursor",
				"id":                              "primary",
				pollingWatermarkFrontiersStateKey: `{"version":1,"deletion":{"query_set":true,"query":{"watermark":"primary-cursor","tie_breaker":"primary"}}}`,
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			bundle := loadPollingWatermarkTestBundle(t)
			bundle.Changefeed.Checkpoint.Keys = []string{"cursor", "id"}
			bundle.Changefeed.PollingWatermark.Watermark = connectors.PollingWatermarkValue{Kind: "opaque_cursor", Path: "cursor"}
			bundle.Changefeed.PollingWatermark.SafetyLagSeconds = 0
			bundle.Changefeed.PollingWatermark.SoftDelete = nil
			bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{
				Path:        "/widgets/deletions",
				RecordsPath: "data",
			}
			bundle.Changefeed.PollingWatermark.RequestBudget = 2
			source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{}}}
			connector, err := NewPollingWatermarkConnector(bundle, nil, source)
			if err != nil {
				t.Fatalf("NewPollingWatermarkConnector: %v", err)
			}
			err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
				Stream:              "widgets",
				Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
				State:               testCase.state,
				CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
			}, func(connectors.CDCEvent) error { return nil })
			if err != nil {
				t.Fatalf("ReadCDC: %v", err)
			}
			if len(source.calls) != 1 || source.calls[0].After == nil || source.calls[0].After.Watermark != "primary-cursor" || source.calls[0].DeletionAfter != nil {
				t.Fatalf("source calls = %+v, want nil opaque deletion start", source.calls)
			}
		})
	}
}

func TestPollingWatermarkExecutorRejectsOversizedPhysicalPages(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		bundle func(*Bundle)
		page   PollingWatermarkPage
		source string
	}{
		{
			name:   "primary",
			bundle: func(*Bundle) {},
			page: PollingWatermarkPage{Records: []connectors.Record{
				{"id": "a", "updated_at": "2026-08-06T10:00:00Z"},
				{"id": "b", "updated_at": "2026-08-06T10:01:00Z"},
				{"id": "c", "updated_at": "2026-08-06T10:02:00Z"},
			}},
			source: "primary",
		},
		{
			name: "deletion",
			bundle: func(bundle *Bundle) {
				bundle.Changefeed.PollingWatermark.SoftDelete = nil
				bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{Path: "/widgets/deletions", RecordsPath: "data"}
			},
			page: PollingWatermarkPage{DeletionRecords: []connectors.Record{
				{"id": "a", "updated_at": "2026-08-06T10:00:00Z"},
				{"id": "b", "updated_at": "2026-08-06T10:01:00Z"},
				{"id": "c", "updated_at": "2026-08-06T10:02:00Z"},
			}},
			source: "deletion",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			bundle := loadPollingWatermarkTestBundle(t)
			testCase.bundle(&bundle)
			source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{testCase.page}}
			connector, err := NewPollingWatermarkConnector(bundle, nil, source)
			if err != nil {
				t.Fatalf("NewPollingWatermarkConnector: %v", err)
			}
			committer := &recordingChangefeedCheckpointCommitter{}
			var delivered int
			err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
				Stream:              "widgets",
				Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
				CheckpointCommitter: committer,
			}, func(connectors.CDCEvent) error {
				delivered++
				return nil
			})
			var refusal *PollingWatermarkNonAdvancingError
			if !errors.As(err, &refusal) || refusal.Reason != PollingWatermarkNonAdvancingReasonPageSize || refusal.Source != testCase.source || !refusal.NonAdvancing() {
				t.Fatalf("ReadCDC error = %v, want typed %s physical-page refusal", err, testCase.source)
			}
			if delivered != 0 || len(committer.states) != 0 {
				t.Fatalf("delivered=%d checkpoints=%+v, want non-advancing refusal", delivered, committer.states)
			}
		})
	}
}

func TestPollingWatermarkExecutorRejectsUndeclaredDeletionRecords(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{DeletionRecords: []connectors.Record{
		{"id": "gone", "updated_at": "2026-08-06T10:01:00Z"},
	}}}}
	connector := newPollingWatermarkTestConnector(t, source)
	committer := &recordingChangefeedCheckpointCommitter{}
	var emitted int
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error {
		emitted++
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "without a declared deletion_endpoint") {
		t.Fatalf("ReadCDC error = %v, want undeclared deletion record rejection", err)
	}
	if emitted != 0 || len(committer.states) != 0 {
		t.Fatalf("emitted=%d committed=%+v, want no undeclared tombstone or checkpoint", emitted, committer.states)
	}
}

func TestPollingWatermarkDeclarationRejectsHardDeleteTombstoneClaim(t *testing.T) {
	bundle := fullValidBundleFS("acme")
	withoutSoftDelete := strings.Replace(pollingWatermarkChangefeedJSON, `,"soft_delete":{"path":"deleted"}`, "", 1)
	bundle["acme/changefeed.json"] = &fstest.MapFile{Data: []byte(withoutSoftDelete)}
	if _, err := Load(bundle, "acme"); err == nil || !strings.Contains(err.Error(), "hard deletes") {
		t.Fatalf("Load error = %v, want hard-delete observability rejection", err)
	}
}

func TestPollingWatermarkDeclarationRejectsDurationOverflowSafetyLag(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("int cannot represent an unsafe duration-sized safety lag")
	}
	bundle := loadPollingWatermarkTestBundle(t)
	bundle.Changefeed.PollingWatermark.SafetyLagSeconds = int(^uint(0) >> 1)
	_, err := NewPollingWatermarkConnector(bundle, nil, &scriptedPollingWatermarkSource{})
	if err == nil || !strings.Contains(err.Error(), "maximum duration-safe") {
		t.Fatalf("NewPollingWatermarkConnector error = %v, want unsafe safety lag rejection", err)
	}
}

func TestPollingWatermarkExecutorReplaysAfterCheckpointPersistenceCrash(t *testing.T) {
	page := PollingWatermarkPage{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}}
	committer := &recordingChangefeedCheckpointCommitter{failAt: 1}
	var delivered []string
	for run := 0; run < 2; run++ {
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{page}}
		connector := newPollingWatermarkTestConnector(t, source)
		err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: committer,
		}, func(event connectors.CDCEvent) error {
			delivered = append(delivered, event.Record["id"].(string))
			return nil
		})
		if run == 0 && err == nil {
			t.Fatal("first ReadCDC succeeded despite simulated persistence crash")
		}
		if run == 1 && err != nil {
			t.Fatalf("replay ReadCDC: %v", err)
		}
	}
	if joined := strings.Join(delivered, ","); joined != "a,a" {
		t.Fatalf("delivered = %q, want replay after uncommitted accepted page", joined)
	}
	if len(committer.states) != 1 || committer.states[0]["id"] != "a" {
		t.Fatalf("committed states = %+v, want only retry state", committer.states)
	}
}

func TestPollingWatermarkExecutorDoesNotCommitWhenDestinationRejectsPage(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}}}}
	connector := newPollingWatermarkTestConnector(t, source)
	committer := &recordingChangefeedCheckpointCommitter{}
	err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
		Stream:              "widgets",
		Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
		CheckpointCommitter: committer,
	}, func(connectors.CDCEvent) error {
		return errors.New("destination durability acknowledgement failed")
	})
	if err == nil || !strings.Contains(err.Error(), "durably accept") {
		t.Fatalf("ReadCDC error = %v, want destination acknowledgement failure", err)
	}
	if len(committer.states) != 0 {
		t.Fatalf("committed states = %+v, want none after destination rejection", committer.states)
	}
}

func TestPollingWatermarkExecutorHonorsRequestBudgetAndCancellation(t *testing.T) {
	t.Run("request budget", func(t *testing.T) {
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{
			{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}, More: true},
			{Records: []connectors.Record{{"id": "b", "updated_at": "2026-08-06T10:01:00Z"}}, More: true},
			{Records: []connectors.Record{{"id": "c", "updated_at": "2026-08-06T10:02:00Z"}}, More: true},
			{Records: []connectors.Record{{"id": "d", "updated_at": "2026-08-06T10:03:00Z"}}, More: true},
		}}
		bundle := loadPollingWatermarkTestBundle(t)
		bundle.Changefeed.PollingWatermark.MaxPages = 4
		connector, err := NewPollingWatermarkConnector(bundle, nil, source)
		if err != nil {
			t.Fatalf("NewPollingWatermarkConnector: %v", err)
		}
		err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
		}, func(connectors.CDCEvent) error { return nil })
		var stop *PollingWatermarkResumableStopError
		if !errors.As(err, &stop) {
			t.Fatalf("ReadCDC error = %v, want typed resumable request-budget stop", err)
		}
		if stop.Reason != PollingWatermarkStopReasonRequestBudget || stop.LastDurablePosition == nil || stop.LastDurablePosition.TieBreaker != "c" {
			t.Fatalf("resumable stop = %+v, want request budget with durable c checkpoint", stop)
		}
		if got := len(source.calls); got != 3 {
			t.Fatalf("source calls = %d, want declared request budget 3", got)
		}
		if source.dataRequests != 3 || len(source.calls) != 3 {
			t.Fatalf("source requests = data:%d total:%d, want exactly three physical requests", source.dataRequests, len(source.calls))
		}
	})

	t.Run("maximum pages", func(t *testing.T) {
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{
			{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}, More: true},
			{Records: []connectors.Record{{"id": "b", "updated_at": "2026-08-06T10:01:00Z"}}, More: true},
			{Records: []connectors.Record{{"id": "c", "updated_at": "2026-08-06T10:02:00Z"}}, More: true},
			{Records: []connectors.Record{{"id": "d", "updated_at": "2026-08-06T10:03:00Z"}}, More: true},
		}}
		bundle := loadPollingWatermarkTestBundle(t)
		bundle.Changefeed.PollingWatermark.RequestBudget = 4
		connector, err := NewPollingWatermarkConnector(bundle, nil, source)
		if err != nil {
			t.Fatalf("NewPollingWatermarkConnector: %v", err)
		}
		err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
		}, func(connectors.CDCEvent) error { return nil })
		var stop *PollingWatermarkResumableStopError
		if !errors.As(err, &stop) {
			t.Fatalf("ReadCDC error = %v, want typed resumable max-pages stop", err)
		}
		if stop.Reason != PollingWatermarkStopReasonMaxPages || stop.LastDurablePosition == nil || stop.LastDurablePosition.TieBreaker != "c" {
			t.Fatalf("resumable stop = %+v, want max pages with durable c checkpoint", stop)
		}
		if got := len(source.calls); got != 3 {
			t.Fatalf("source calls = %d, want declared maximum pages 3", got)
		}
	})

	t.Run("source must consume the shared budget", func(t *testing.T) {
		source := pollingWatermarkPageSourceFunc(func(context.Context, PollingWatermarkPageRequest) (PollingWatermarkPage, error) {
			return PollingWatermarkPage{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}}, nil
		})
		connector := newPollingWatermarkTestConnector(t, source)
		committer := &recordingChangefeedCheckpointCommitter{}
		var emitted int
		err := connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: committer,
		}, func(connectors.CDCEvent) error {
			emitted++
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), "without consuming the request budget") {
			t.Fatalf("ReadCDC error = %v, want source budget contract rejection", err)
		}
		if emitted != 0 || len(committer.states) != 0 {
			t.Fatalf("emitted=%d checkpoints=%+v, want no delivery or checkpoint", emitted, committer.states)
		}
	})

	t.Run("deletion endpoint requires two request-budget tokens", func(t *testing.T) {
		bundle := loadPollingWatermarkTestBundle(t)
		bundle.Changefeed.PollingWatermark.SoftDelete = nil
		bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{Path: "/widgets/deletions", RecordsPath: "data"}
		bundle.Changefeed.PollingWatermark.RequestBudget = 1
		_, err := NewPollingWatermarkConnector(bundle, nil, &scriptedPollingWatermarkSource{})
		if err == nil || !strings.Contains(err.Error(), "request_budget of at least 2") {
			t.Fatalf("NewPollingWatermarkConnector error = %v, want deletion budget validation", err)
		}
	})

	t.Run("declared deletion endpoint must consume a budget token", func(t *testing.T) {
		bundle := loadPollingWatermarkTestBundle(t)
		bundle.Changefeed.PollingWatermark.SoftDelete = nil
		bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{Path: "/widgets/deletions", RecordsPath: "data"}
		source := pollingWatermarkPageSourceFunc(func(ctx context.Context, req PollingWatermarkPageRequest) (PollingWatermarkPage, error) {
			if err := req.RequestBudget.Consume(ctx); err != nil {
				return PollingWatermarkPage{}, err
			}
			return PollingWatermarkPage{DeletionRecords: []connectors.Record{{"id": "gone", "updated_at": "2026-08-06T10:00:00Z"}}}, nil
		})
		connector, err := NewPollingWatermarkConnector(bundle, nil, source)
		if err != nil {
			t.Fatalf("NewPollingWatermarkConnector: %v", err)
		}
		committer := &recordingChangefeedCheckpointCommitter{}
		var emitted int
		err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: committer,
		}, func(connectors.CDCEvent) error {
			emitted++
			return nil
		})
		if err == nil || !strings.Contains(err.Error(), "every declared-source request") {
			t.Fatalf("ReadCDC error = %v, want deletion request budget contract rejection", err)
		}
		if emitted != 0 || len(committer.states) != 0 {
			t.Fatalf("emitted=%d checkpoints=%+v, want no delivery or checkpoint", emitted, committer.states)
		}
	})

	t.Run("deletion endpoint preflights shared request budget", func(t *testing.T) {
		bundle := loadPollingWatermarkTestBundle(t)
		bundle.Changefeed.PollingWatermark.SoftDelete = nil
		bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{Path: "/widgets/deletions", RecordsPath: "data"}
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{
			{Records: []connectors.Record{{"id": "later", "updated_at": "2026-08-06T10:02:00Z"}}, DeletionRecords: []connectors.Record{{"id": "gone", "updated_at": "2026-08-06T10:01:00Z"}}, PrimaryMore: true},
			{Records: []connectors.Record{{"id": "later", "updated_at": "2026-08-06T10:02:00Z"}}, More: true},
		}}
		connector, err := NewPollingWatermarkConnector(bundle, nil, source)
		if err != nil {
			t.Fatalf("NewPollingWatermarkConnector: %v", err)
		}
		committer := &recordingChangefeedCheckpointCommitter{}
		err = connector.ReadCDC(context.Background(), connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: committer,
		}, func(connectors.CDCEvent) error { return nil })
		var stop *PollingWatermarkResumableStopError
		if !errors.As(err, &stop) {
			t.Fatalf("ReadCDC error = %v, want typed resumable request-budget stop", err)
		}
		if stop.Reason != PollingWatermarkStopReasonRequestBudget || stop.LastDurablePosition == nil || stop.LastDurablePosition.TieBreaker != "later" {
			t.Fatalf("resumable stop = %+v, want only the first fully accepted page durable", stop)
		}
		if source.dataRequests+source.deletionRequests != 2 || source.dataRequests != 1 || source.deletionRequests != 1 || len(source.calls) != 1 {
			t.Fatalf("physical requests = data:%d deletions:%d calls:%d, want no partial second page", source.dataRequests, source.deletionRequests, len(source.calls))
		}
		if len(committer.states) != 1 || committer.states[0]["id"] != "later" {
			t.Fatalf("committed checkpoints = %+v, want only the first accepted page", committer.states)
		}
	})

	t.Run("cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}, More: true}}, onFetch: cancel}
		connector := newPollingWatermarkTestConnector(t, source)
		err := connector.ReadCDC(ctx, connectors.CDCReadRequest{
			Stream:              "widgets",
			Config:              connectors.RuntimeConfig{Config: map[string]string{"base_url": "https://example.test"}},
			CheckpointCommitter: &recordingChangefeedCheckpointCommitter{},
		}, func(connectors.CDCEvent) error { return nil })
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ReadCDC error = %v, want context cancellation", err)
		}
		if got := len(source.calls); got != 1 {
			t.Fatalf("source calls = %d, want cancellation before a second fetch", got)
		}
	})
}

const pollingWatermarkChangefeedJSON = `{
  "status":"implemented",
  "mechanism":"polling_watermark",
  "source":{"artifact_url":"https://example.test/polling","artifact_version":"v1","retrieved_at":"2026-08-06"},
  "executor":{"kind":"engine","id":"polling_watermark"},
  "checkpoint":{"kind":"watermark_tuple","keys":["updated_at","id"],"commit_after":"downstream_ack","on_invalid":"resnapshot_required"},
  "delivery":{"ordering":"provider_declared_stable_order","duplicates":"at_least_once","deletes":"tombstone","dedupe_key":["id","updated_at"]},
  "streams":["widgets"],
  "polling_watermark":{"watermark":{"kind":"timestamp","path":"updated_at"},"tie_breaker":{"path":"id"},"boundary":"inclusive","safety_lag_seconds":0,"page_size":2,"max_pages":3,"request_budget":3,"soft_delete":{"path":"deleted"}}
}`
