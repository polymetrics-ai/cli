package decision

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStorePersistsProvenanceAndRendersPRSummary(t *testing.T) {
	t.Parallel()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	record := Record{
		ID: "decision-1", DeliveryID: "issue-380", ExecutionID: "execution-1",
		UnitID: "discuss-milestone/M001", QuestionID: "m001_scope", Question: "What should ship?",
		Answer: "Full safe parity (Recommended)", Actor: ActorShepherd,
		Basis: "explicit user request for Asana connector parity", At: time.Unix(1, 0).UTC(),
	}
	if err := store.Append(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	records, err := Read(t.TempDir())
	if err == nil || records != nil {
		t.Fatal("reading a different ledger unexpectedly succeeded")
	}
	records, err = store.Records()
	if err != nil {
		t.Fatal(err)
	}
	summary := Markdown(records)
	for _, want := range []string{"Shepherd decisions", "Full safe parity", "shepherd", "explicit user request"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q: %s", want, summary)
		}
	}
}

func TestStoreRejectsUnattributedOrUnsafeDecision(t *testing.T) {
	t.Parallel()
	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	base := Record{ID: "d", DeliveryID: "issue-380", ExecutionID: "e", UnitID: "u", QuestionID: "q", Question: "Q", Answer: "A", At: time.Now().UTC()}
	if err := store.Append(context.Background(), base); err == nil {
		t.Fatal("missing actor and basis accepted")
	}
	base.Actor, base.Basis, base.Answer = ActorShepherd, "issue context", "unsafe\nanswer"
	if err := store.Append(context.Background(), base); err == nil {
		t.Fatal("control characters accepted")
	}
	base.Answer = "github_pat_example"
	if err := store.Append(context.Background(), base); err == nil {
		t.Fatal("secret-shaped decision text accepted")
	}
}

func TestStoreRecoversTornTailAndDeduplicatesImmutableDecisionID(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	store, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	record := Record{
		ID: "decision-1", DeliveryID: "issue-389", ExecutionID: "execution-1", UnitID: "unit-1",
		QuestionID: "question-1", Question: "Continue?", Answer: "continue", Actor: ActorHuman,
		Basis: "approved issue context", At: time.Unix(1_700_000_000, 0).UTC(),
	}
	if err := store.Append(context.Background(), record); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(context.Background(), record); err != nil {
		t.Fatalf("exact duplicate was not idempotent: %v", err)
	}
	collision := record
	collision.Answer = "stop"
	if err := store.Append(context.Background(), collision); err == nil {
		t.Fatal("decision ID collision with different content was accepted")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "decisions.jsonl")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(`{"decision_id":"torn"`); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	records, err := reopened.Records()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].ID != record.ID {
		t.Fatalf("recovered records=%+v", records)
	}
}
