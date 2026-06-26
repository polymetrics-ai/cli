package dropboxsign

import (
	"context"

	"polymetrics.ai/internal/connectors"
)

// Write satisfies the connectors.Connector interface. Dropbox Sign is exposed as
// a read-only source here: signature requests are created with file uploads and
// signer flows that do not map cleanly onto reverse-ETL record writes, so no
// write actions are offered. The connector deliberately reports Write=false in
// its Metadata and does not implement WriteValidator/DryRunWriter.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
