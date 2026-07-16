package decision

import (
	"strings"
	"testing"
	"time"
)

func TestSnapshotRecordsIsDeliveryBoundDeterministicAndMonotonic(t *testing.T) {
	t.Parallel()
	at := time.Unix(1_700_000_000, 0).UTC()
	records := []Record{
		{ID: "b", DeliveryID: "issue-389", ExecutionID: "e2", UnitID: "u", QuestionID: "q2", Question: "second", Answer: "stop", Actor: ActorHuman, Basis: "human", At: at},
		{ID: "foreign", DeliveryID: "issue-390", ExecutionID: "e", UnitID: "u", QuestionID: "q", Question: "foreign", Answer: "ignore", Actor: ActorHuman, Basis: "human", At: at},
		{ID: "a", DeliveryID: "issue-389", ExecutionID: "e1", UnitID: "u", QuestionID: "q1", Question: "first", Answer: "retry", Actor: ActorShepherd, Basis: "policy", At: at},
	}
	first, err := SnapshotRecords(records, "issue-389")
	if err != nil {
		t.Fatal(err)
	}
	second, err := SnapshotRecords([]Record{records[2], records[0], records[1]}, "issue-389")
	if err != nil {
		t.Fatal(err)
	}
	if first.Revision != 2 || first.LedgerHash != second.LedgerHash || first.Summary != second.Summary {
		t.Fatalf("snapshot is not deterministic: first=%+v second=%+v", first, second)
	}
	if strings.Contains(first.Summary, "foreign") || strings.Index(first.Summary, "first") > strings.Index(first.Summary, "second") {
		t.Fatalf("snapshot ordering/binding is wrong: %s", first.Summary)
	}
	third, err := SnapshotRecords(append(records, Record{ID: "c", DeliveryID: "issue-389", ExecutionID: "e3", UnitID: "u", QuestionID: "q3", Question: "third", Answer: "continue", Actor: ActorHuman, Basis: "human", At: at.Add(time.Second)}), "issue-389")
	if err != nil {
		t.Fatal(err)
	}
	if third.Revision != first.Revision+1 || third.LedgerHash == first.LedgerHash {
		t.Fatalf("new ledger record did not advance snapshot: before=%+v after=%+v", first, third)
	}
}
