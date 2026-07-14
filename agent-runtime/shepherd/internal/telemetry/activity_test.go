package telemetry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestActivityStoreRejectsRawPayloadAndDeduplicates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "activity"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	bad := Activity{ID: "e1", RunID: "r1", Kind: "message", At: time.Now(), Detail: "raw prompt"}
	if _, err := store.Append(ctx, bad); err == nil {
		t.Fatal("expected unrestricted detail to be rejected")
	}

	event := Activity{ID: "e2", RunID: "r1", UnitID: "u1", Kind: "heartbeat", Status: "alive", At: time.Unix(1, 0).UTC()}
	inserted, err := store.Append(ctx, event)
	if err != nil || !inserted {
		t.Fatalf("append: inserted=%v err=%v", inserted, err)
	}
	inserted, err = store.Append(ctx, event)
	if err != nil || inserted {
		t.Fatalf("duplicate append: inserted=%v err=%v", inserted, err)
	}
}

func TestActivityStoreAcceptsTypedHumanDecisionAudit(t *testing.T) {
	store, err := Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	event := Activity{ID: "decision-1", RunID: "issue-380", UnitID: "new-milestone", Kind: "human.decision", Status: "approved_depth", Tool: "depth_verification_M001-abc123_confirm", At: time.Unix(1, 0).UTC()}
	if appended, err := store.Append(context.Background(), event); err != nil || !appended {
		t.Fatalf("appended=%t err=%v", appended, err)
	}
}

func TestActivityStoreAcceptsBoundedPolicyViolation(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	event := Activity{
		ID: "policy-1", RunID: "issue-380", UnitID: "execute-task/M001/S02/T04",
		Kind: "policy", Status: "write_scope_violation", At: time.Unix(1, 0).UTC(),
	}
	if appended, err := store.Append(context.Background(), event); err != nil || !appended {
		t.Fatalf("append policy activity: appended=%t err=%v", appended, err)
	}
}

func TestActivityStoreRecoversTornTail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "activity")
	store, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	event := Activity{ID: "e1", RunID: "r1", Kind: "heartbeat", At: time.Unix(1, 0).UTC()}
	if _, err := store.Append(ctx, event); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	segment := filepath.Join(dir, segmentName)
	file, err := os.OpenFile(segment, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open segment: %v", err)
	}
	_, _ = file.WriteString(`{"event_id":"torn"`)
	_ = file.Close()
	store, err = Open(ctx, dir)
	if err != nil {
		t.Fatalf("recover: %v", err)
	}
	defer store.Close()
	inserted, err := store.Append(ctx, event)
	if err != nil || inserted {
		t.Fatalf("dedupe after recovery: inserted=%v err=%v", inserted, err)
	}
}
