package decision

import (
	"context"
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
	defer store.Close()
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
	defer store.Close()
	base := Record{ID: "d", DeliveryID: "issue-380", ExecutionID: "e", UnitID: "u", QuestionID: "q", Question: "Q", Answer: "A", At: time.Now().UTC()}
	if err := store.Append(context.Background(), base); err == nil {
		t.Fatal("missing actor and basis accepted")
	}
	base.Actor, base.Basis, base.Answer = ActorShepherd, "issue context", "unsafe\nanswer"
	if err := store.Append(context.Background(), base); err == nil {
		t.Fatal("control characters accepted")
	}
}
