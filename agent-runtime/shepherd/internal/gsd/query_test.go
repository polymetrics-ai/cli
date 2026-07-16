package gsd

import "testing"

func TestDecodeQuerySnapshot(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"state":{"activeMilestone":{"id":"M001","title":"Build"},"activeSlice":null,"activeTask":null,"phase":"pre-planning","blockers":[],"nextAction":"Plan milestone M001."},"next":{"action":"dispatch","unitType":"discuss-milestone","unitId":"M001"},"cost":{"total":0}}`)
	snapshot, err := DecodeQuery(raw)
	if err != nil {
		t.Fatalf("decode query: %v", err)
	}
	if snapshot.MilestoneID != "M001" || snapshot.Phase != "pre-planning" || snapshot.Next.UnitType != "discuss-milestone" {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestDecodeQueryFailsClosedOnUnknownShape(t *testing.T) {
	t.Parallel()

	for _, raw := range [][]byte{[]byte(`{}`), []byte(`{"state":{"phase":"mystery"},"next":{"action":"teleport"}}`)} {
		if _, err := DecodeQuery(raw); err == nil {
			t.Fatalf("expected query %s to fail", raw)
		}
	}
}

func TestDecodeQueryRejectsUnsafeUnitIdentity(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"state":{"phase":"executing"},"next":{"action":"dispatch","unitType":"execute-task","unitId":"execute-task/M001; gh api"}}`)
	if _, err := DecodeQuery(raw); err == nil {
		t.Fatal("unsafe command-shaped unit identity was accepted")
	}
}

func TestDecodeQueryAcceptsPinnedGSD11Skip(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"state":{"activeMilestone":{"id":"M001"},"phase":"refining","blockers":[],"nextAction":"Skip stale sketch"},"next":{"action":"skip","unitType":"validate-slice","unitId":"S01"}}`)
	snapshot, err := DecodeQuery(raw)
	if err != nil {
		t.Fatalf("decode skip: %v", err)
	}
	if snapshot.Next.Action != "skip" {
		t.Fatalf("action=%q want skip", snapshot.Next.Action)
	}
}
