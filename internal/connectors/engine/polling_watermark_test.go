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
	"time"

	"polymetrics.ai/internal/connectors"
)

type fixedPollingWatermarkClock struct{ now time.Time }

func (c fixedPollingWatermarkClock) Now() time.Time { return c.now }

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

func loadPollingWatermarkTestBundle(t *testing.T) Bundle {
	t.Helper()
	bundle, err := Load(os.DirFS("testdata/bundles"), "polling-watermark-demo")
	if err != nil {
		t.Fatalf("Load polling test bundle: %v", err)
	}
	return bundle
}

func newPollingWatermarkTestConnector(t *testing.T, source PollingWatermarkPageSource, clock PollingWatermarkClock) *PollingWatermarkConnector {
	t.Helper()
	connector, err := NewPollingWatermarkConnector(loadPollingWatermarkTestBundle(t), nil, source, clock)
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
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)})
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
	_, err := NewPollingWatermarkConnector(bundle, nil, &scriptedPollingWatermarkSource{}, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)})
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
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 2, 0, 0, time.UTC)})
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
		connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{})
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
		connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{})
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
		if emitted != 1 || len(committer.states) != 0 {
			t.Fatalf("emitted=%d checkpoints=%+v, want no invalid record or checkpoint", emitted, committer.states)
		}
	})
}

func TestPollingWatermarkExecutorAppliesDeclaredSafetyLagWithInjectedClock(t *testing.T) {
	clock := fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)}
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{}}}
	connector := newPollingWatermarkTestConnector(t, source, clock)
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
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{})
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
	if len(committer.states) != 1 || committer.states[0]["updated_at"] != "2026-08-06T10:01:00Z" || committer.states[0]["id"] != "next" {
		t.Fatalf("committed checkpoints = %+v, want no regression below the durable checkpoint", committer.states)
	}
}

func TestPollingWatermarkExecutorLeavesInitialSnapshotBoundaryToSource(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{}}}
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)})
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
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 2, 0, 0, time.UTC)})
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
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{})
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
	connector, err := NewPollingWatermarkConnector(bundle, nil, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 2, 0, 0, time.UTC)})
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

func TestPollingWatermarkExecutorRejectsUndeclaredDeletionRecords(t *testing.T) {
	source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{DeletionRecords: []connectors.Record{
		{"id": "gone", "updated_at": "2026-08-06T10:01:00Z"},
	}}}}
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 2, 0, 0, time.UTC)})
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
	_, err := NewPollingWatermarkConnector(bundle, nil, &scriptedPollingWatermarkSource{}, fixedPollingWatermarkClock{})
	if err == nil || !strings.Contains(err.Error(), "maximum duration-safe") {
		t.Fatalf("NewPollingWatermarkConnector error = %v, want unsafe safety lag rejection", err)
	}
}

func TestPollingWatermarkExecutorReplaysAfterCheckpointPersistenceCrash(t *testing.T) {
	page := PollingWatermarkPage{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}}
	clock := fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 2, 0, 0, time.UTC)}
	committer := &recordingChangefeedCheckpointCommitter{failAt: 1}
	var delivered []string
	for run := 0; run < 2; run++ {
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{page}}
		connector := newPollingWatermarkTestConnector(t, source, clock)
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
	connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 2, 0, 0, time.UTC)})
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
		connector, err := NewPollingWatermarkConnector(bundle, nil, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 4, 0, 0, time.UTC)})
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
		connector, err := NewPollingWatermarkConnector(bundle, nil, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 4, 0, 0, time.UTC)})
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
		connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{})
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
		connector, err := NewPollingWatermarkConnector(bundle, nil, source, fixedPollingWatermarkClock{})
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

	t.Run("deletion endpoint shares request budget", func(t *testing.T) {
		bundle := loadPollingWatermarkTestBundle(t)
		bundle.Changefeed.PollingWatermark.SoftDelete = nil
		bundle.Changefeed.PollingWatermark.DeletionEndpoint = &connectors.PollingWatermarkDeletionEndpoint{Path: "/widgets/deletions", RecordsPath: "data"}
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{
			{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}, DeletionRecords: []connectors.Record{{"id": "gone", "updated_at": "2026-08-06T10:01:00Z"}}, More: true},
			{Records: []connectors.Record{{"id": "later", "updated_at": "2026-08-06T10:02:00Z"}}, More: true},
		}}
		connector, err := NewPollingWatermarkConnector(bundle, nil, source, fixedPollingWatermarkClock{})
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
		if stop.Reason != PollingWatermarkStopReasonRequestBudget || stop.LastDurablePosition == nil || stop.LastDurablePosition.TieBreaker != "gone" {
			t.Fatalf("resumable stop = %+v, want only the first fully accepted page durable", stop)
		}
		if source.dataRequests+source.deletionRequests != 3 || source.dataRequests != 2 || source.deletionRequests != 1 {
			t.Fatalf("physical requests = data:%d deletions:%d, want shared cap of three", source.dataRequests, source.deletionRequests)
		}
		if len(committer.states) != 1 || committer.states[0]["id"] != "gone" {
			t.Fatalf("committed checkpoints = %+v, want only the first accepted page", committer.states)
		}
	})

	t.Run("cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		source := &scriptedPollingWatermarkSource{pages: []PollingWatermarkPage{{Records: []connectors.Record{{"id": "a", "updated_at": "2026-08-06T10:00:00Z"}}, More: true}}, onFetch: cancel}
		connector := newPollingWatermarkTestConnector(t, source, fixedPollingWatermarkClock{now: time.Date(2026, 8, 6, 10, 4, 0, 0, time.UTC)})
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
