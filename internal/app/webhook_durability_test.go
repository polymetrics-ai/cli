package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

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
