package postgres_test

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	native "polymetrics.ai/internal/connectors/native/postgres"
)

func TestCDCIsFailClosedUntilStreamedStagingExists(t *testing.T) {
	c := native.New()
	cdc, ok := any(c).(connectors.ChangefeedExecutor)
	if !ok {
		t.Fatal("postgres connector must implement a logical-replication ChangefeedExecutor")
	}
	err := cdc.ReadCDC(context.Background(), connectors.CDCReadRequest{Stream: "public.users", Config: fixtureConfig()}, func(connectors.CDCEvent) error {
		return nil
	})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("ReadCDC = %v, want fail-closed unsupported result", err)
	}
}
