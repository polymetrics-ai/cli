package postgres

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// ReadCDC is a DOCUMENTED STUB. Full PostgreSQL change data capture uses
// logical replication, which requires the pglogrepl dependency — a gated
// add not present in go.mod. Until that dependency is approved, CDC is
// unsupported.
//
// Recorded CDC plan (for the future pglogrepl implementation, ported
// verbatim from legacy internal/connectors/postgres/postgres.go ReadCDC):
//   - Server prerequisite: wal_level=logical.
//   - Create a logical replication slot (pgoutput plugin), e.g.
//     SELECT pg_create_logical_replication_slot('pm_<connector>', 'pgoutput').
//   - Create a PUBLICATION for the target tables:
//     CREATE PUBLICATION pm_pub FOR TABLE schema.table[, ...].
//   - Start replication from the stored LSN (confirmed_flush_lsn) via the
//     START_REPLICATION protocol; decode pgoutput Insert/Update/Delete
//     messages into connectors.CDCEvent{Operation, Record}.
//   - Persist the last committed LSN in CDCEvent.State (e.g.
//     {"lsn": "0/1A2B3C"}) so the next run resumes after the last flushed
//     change.
func (c Connector) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	return fmt.Errorf("postgres CDC requires the gated pglogrepl dependency (wal_level=logical, replication slot, publication, lsn state): %w", connectors.ErrUnsupportedOperation)
}
