package postgres_test

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/postgres"
)

func TestCDCRejectsFixtureModeRatherThanPretendingToReplicate(t *testing.T) {
	c := native.New()
	cdc, ok := any(c).(connectors.ChangefeedExecutor)
	if !ok {
		t.Fatal("postgres connector must implement a logical-replication ChangefeedExecutor")
	}
	err := cdc.ReadCDC(context.Background(), connectors.CDCReadRequest{Stream: "public.users", Config: fixtureConfig()}, func(connectors.CDCEvent) error {
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "requires a real PostgreSQL source") {
		t.Fatalf("ReadCDC = %v, want fixture-mode refusal", err)
	}
}
