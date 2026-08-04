package vault_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/vault"
)

func TestWriteApprovalConsumptionMarkerIsMonotonic(t *testing.T) {
	projectDir := t.TempDir()
	v, err := vault.Init(projectDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	root := v.WriteApprovalRoot()
	approvalID := "rplan_fixture\x00plan-hash\x00connector_command"
	nonce := "fixture-consumption-nonce"
	consumed, err := root.Consumed(approvalID)
	if err != nil || consumed {
		t.Fatalf("Consumed(before) = %v, %v, want false", consumed, err)
	}
	if err := root.Consume(approvalID, nonce, strings.Repeat("a", 64), time.Now().UTC()); err != nil {
		t.Fatalf("Consume(first) error = %v", err)
	}
	consumed, err = root.Consumed(approvalID)
	if err != nil || !consumed {
		t.Fatalf("Consumed(after) = %v, %v, want true", consumed, err)
	}
	if err := root.Consume(approvalID, "different-nonce", strings.Repeat("b", 64), time.Now().UTC()); !errors.Is(err, vault.ErrWriteApprovalConsumed) {
		t.Fatalf("Consume(replay) error = %v, want ErrWriteApprovalConsumed", err)
	}
	markers, err := filepath.Glob(filepath.Join(projectDir, "vault", "write-approval-consumed", "*.used"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("consumption markers = %d, want 1", len(markers))
	}
	payload, err := os.ReadFile(markers[0])
	if err != nil {
		t.Fatalf("ReadFile(marker) error = %v", err)
	}
	if strings.Contains(string(payload), nonce) || strings.Contains(string(payload), approvalID) || !strings.Contains(string(payload), `"nonce_id":`) || !strings.Contains(string(payload), `"mac":`) {
		t.Fatalf("consumption marker does not preserve its authenticated opaque contract: %s", payload)
	}
}
