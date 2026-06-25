package amazonsellerpartner

import (
	"context"

	"polymetrics/internal/connectors"
)

// Write is unsupported: the Amazon Selling Partner connector is read-only. SP-API
// does expose mutating operations (e.g. submitting feeds, confirming shipments),
// but none are safe, idempotent reverse-ETL targets, so the connector advertises
// Capabilities.Write=false and rejects writes here.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
