package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/webhook"
)

func TestWebhookReceiptCommitDeduplicatesAcrossIndependentStores(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	firstApp, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := firstApp.ConfigureWebhookReceiver(ctx, ConfigureWebhookReceiverRequest{
		Name: "receipt-commit",
		Exposure: webhook.ExposureConfig{
			Mode:        webhook.ExposureModeOperatorEndpoint,
			CallbackURL: "https://operator.example.test/receiver",
		},
		ReceiptCapacity: 2,
	}); err != nil {
		t.Fatal(err)
	}
	secondApp, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	insertCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	firstStore, err := firstApp.WebhookReceiptStore("receipt-commit")
	if err != nil {
		t.Fatal(err)
	}
	secondStore, err := secondApp.WebhookReceiptStore("receipt-commit")
	if err != nil {
		t.Fatal(err)
	}

	type insertResult struct {
		body   string
		result webhook.ReceiptInsertResult
		err    error
	}
	start := make(chan struct{})
	results := make(chan insertResult, 2)
	insert := func(store webhook.ReceiptStore, body string) {
		<-start
		result, err := store.Insert(insertCtx, webhook.Receipt{
			Event:      webhook.VerifiedEvent{ID: "same-provider-event"},
			RawBody:    []byte(body),
			ReceivedAt: time.Now().UTC(),
		})
		results <- insertResult{body: body, result: result, err: err}
	}
	go insert(firstStore, `{"source":"first"}`)
	go insert(secondStore, `{"source":"second"}`)
	close(start)

	newBody := ""
	newCount := 0
	duplicateCount := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("Insert() error = %v", result.err)
		}
		switch result.result {
		case webhook.ReceiptInsertNew:
			newCount++
			newBody = result.body
		case webhook.ReceiptInsertDuplicate:
			duplicateCount++
		default:
			t.Fatalf("Insert() result = %q", result.result)
		}
	}
	if newCount != 1 || duplicateCount != 1 {
		t.Fatalf("new=%d duplicate=%d", newCount, duplicateCount)
	}

	recoveredApp, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveredApp.state.WebhookReceipts) != 1 {
		t.Fatalf("durable receipt index count = %d", len(recoveredApp.state.WebhookReceipts))
	}
	for _, receipt := range recoveredApp.state.WebhookReceipts {
		payload, err := recoveredApp.vault.Get(context.Background(), receipt.EncryptedPayloadID)
		if err != nil {
			t.Fatal(err)
		}
		body, err := base64.RawStdEncoding.DecodeString(payload["body_base64"])
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != newBody {
			t.Fatal("duplicate receipt replaced the committed payload")
		}
	}
}

func TestWebhookReceiptCommitKeepsLiveStateCurrent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	instance, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	const receiptCount = 8
	if _, err := instance.ConfigureWebhookReceiver(context.Background(), ConfigureWebhookReceiverRequest{
		Name: "live-state",
		Exposure: webhook.ExposureConfig{
			Mode:        webhook.ExposureModeOperatorEndpoint,
			CallbackURL: "https://operator.example.test/receiver",
		},
		ReceiptCapacity: receiptCount,
	}); err != nil {
		t.Fatal(err)
	}
	store, err := instance.WebhookReceiptStore("live-state")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := make(chan struct{})
	results := make(chan error, receiptCount)
	for index := range receiptCount {
		go func(index int) {
			<-start
			result, err := store.Insert(ctx, webhook.Receipt{
				Event:      webhook.VerifiedEvent{ID: fmt.Sprintf("event-%d", index)},
				RawBody:    []byte(fmt.Sprintf(`{"event":%d}`, index)),
				ReceivedAt: time.Now().UTC(),
			})
			if err == nil && result != webhook.ReceiptInsertNew {
				err = fmt.Errorf("Insert() result = %q", result)
			}
			results <- err
		}(index)
	}
	close(start)
	for range receiptCount {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	instance.stateMu.RLock()
	defer instance.stateMu.RUnlock()
	if got := len(instance.state.WebhookReceipts); got != receiptCount {
		t.Fatalf("live receipt count = %d, want %d", got, receiptCount)
	}
}

func TestWebhookReceiptInsertHonorsStateMutationDeadline(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	instance, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := instance.ConfigureWebhookReceiver(context.Background(), ConfigureWebhookReceiverRequest{
		Name: "deadline-receiver",
		Exposure: webhook.ExposureConfig{
			Mode:        webhook.ExposureModeOperatorEndpoint,
			CallbackURL: "https://operator.example.test/receiver",
		},
		ReceiptCapacity: 1,
	}); err != nil {
		t.Fatal(err)
	}
	store, err := instance.WebhookReceiptStore("deadline-receiver")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	instance.stateMu.Lock()
	result, insertErr := store.Insert(ctx, webhook.Receipt{
		Event:      webhook.VerifiedEvent{ID: "deadline-event"},
		RawBody:    []byte(`{"event":"deadline"}`),
		ReceivedAt: time.Now().UTC(),
	})
	instance.stateMu.Unlock()
	if !errors.Is(insertErr, context.DeadlineExceeded) {
		t.Fatalf("Insert() error = %v, want deadline exceeded", insertErr)
	}
	if result != webhook.ReceiptInsertRejected {
		t.Fatalf("Insert() result = %q, want rejected", result)
	}
	if got := len(instance.snapshotState().WebhookReceipts); got != 0 {
		t.Fatalf("committed receipt count = %d, want 0", got)
	}
	entries, err := os.ReadDir(filepath.Join(root, ".polymetrics", "vault"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "key" {
		t.Fatalf("vault entries after rejected insert = %v", entries)
	}
}

type blockingReceiptSource struct {
	started chan struct{}
	release chan struct{}
}

func (*blockingReceiptSource) Name() string { return "blocking_receipt_source" }

func (s *blockingReceiptSource) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:         s.Name(),
		DisplayName:  "Blocking receipt source",
		Description:  "Test source that pauses an ETL run.",
		Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (*blockingReceiptSource) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (s *blockingReceiptSource) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: s.Name(), Streams: []connectors.Stream{{Name: "records", PrimaryKey: []string{"id"}}}}, nil
}

func (s *blockingReceiptSource) Read(ctx context.Context, _ connectors.ReadRequest, emit func(connectors.Record) error) error {
	select {
	case s.started <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-s.release:
	case <-ctx.Done():
		return ctx.Err()
	}
	return emit(connectors.Record{"id": "1"})
}

func (*blockingReceiptSource) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func TestRunETLAndWebhookReceiptsCommitSharedState(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	instance, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	source := &blockingReceiptSource{started: make(chan struct{}, 1), release: make(chan struct{})}
	instance.registry.Register(source)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := instance.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	if _, err := instance.AddCredential(ctx, AddCredentialRequest{
		Name:      "warehouse",
		Connector: "warehouse",
		Config:    map[string]string{"path": filepath.Join(root, ".polymetrics", "warehouse")},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := instance.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "source-to-warehouse",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
		Streams: map[string]StreamConfig{
			"records": {SyncMode: "full_refresh_overwrite", PrimaryKey: []string{"id"}, DestinationTable: "records"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := instance.ConfigureWebhookReceiver(ctx, ConfigureWebhookReceiverRequest{
		Name: "shared-state",
		Exposure: webhook.ExposureConfig{
			Mode:        webhook.ExposureModeOperatorEndpoint,
			CallbackURL: "https://operator.example.test/receiver",
		},
		ReceiptCapacity: 2,
	}); err != nil {
		t.Fatal(err)
	}
	store, err := instance.WebhookReceiptStore("shared-state")
	if err != nil {
		t.Fatal(err)
	}

	type runResult struct {
		run Run
		err error
	}
	runs := make(chan runResult, 1)
	go func() {
		run, err := instance.RunETL(ctx, RunETLRequest{Connection: "source-to-warehouse", Stream: "records", BatchSize: 1})
		runs <- runResult{run: run, err: err}
	}()
	select {
	case <-source.started:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	start := make(chan struct{})
	receiptErrors := make(chan error, 2)
	for index := range 2 {
		go func(index int) {
			<-start
			result, err := store.Insert(ctx, webhook.Receipt{
				Event:      webhook.VerifiedEvent{ID: fmt.Sprintf("concurrent-event-%d", index)},
				RawBody:    []byte(fmt.Sprintf(`{"event":%d}`, index)),
				ReceivedAt: time.Now().UTC(),
			})
			if err == nil && result != webhook.ReceiptInsertNew {
				err = fmt.Errorf("Insert() result = %q", result)
			}
			receiptErrors <- err
		}(index)
	}
	close(start)
	close(source.release)
	for range 2 {
		if err := <-receiptErrors; err != nil {
			t.Fatal(err)
		}
	}
	result := <-runs
	if result.err != nil {
		t.Fatal(result.err)
	}
	if result.run.Status != "completed" {
		t.Fatalf("run status = %q, want completed", result.run.Status)
	}

	current := instance.snapshotState()
	if got := len(current.WebhookReceipts); got != 2 {
		t.Fatalf("live receipt count = %d, want 2", got)
	}
	if state := current.StreamStates[streamStateKey("source-to-warehouse", "records")]; state.Checkpoint == nil {
		t.Fatal("stream state checkpoint is missing")
	}
	reopened, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.GetRun(result.run.ID); err != nil {
		t.Fatalf("GetRun() after reopen error = %v", err)
	}
	if got := len(reopened.snapshotState().WebhookReceipts); got != 2 {
		t.Fatalf("durable receipt count = %d, want 2", got)
	}
}
