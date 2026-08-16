package app

import (
	"context"
	"fmt"
	"path/filepath"

	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

// connectionArrowSegmentStore binds connector-neutral segment writes to the
// already-owned warehouse directory of one persisted connection. It is not a
// replacement for the legacy local-table WAL: this store is reachable only by
// the transformed Arrow fast path.
type connectionArrowSegmentStore struct {
	app          *App
	connectionID string
	stream       string
}

func (s *connectionArrowSegmentStore) StoreArrowSegment(ctx context.Context, request synctransport.FastSegmentWriteRequest) (synctransport.FastSegmentReceipt, error) {
	if s == nil || s.app == nil || request.ConnectionID != s.connectionID || request.Stream != s.stream || request.Generation <= 0 || request.Record == nil {
		return synctransport.FastSegmentReceipt{}, synctransport.ErrArrowFastPathInvalid
	}
	var conn Connection
	for _, candidate := range s.app.state.Connections {
		if candidate.ID == request.ConnectionID {
			conn = candidate
			break
		}
	}
	if conn.ID == "" {
		return synctransport.FastSegmentReceipt{}, synctransport.ErrArrowFastPathInvalid
	}
	// Stop before any new segment artifact if the warehouse filesystem loses
	// its 3 GiB reserve. The controller records this failure's phase counters
	// before App performs terminal-run cleanup.
	if err := warehouse.RequireMinimumFastPathFreeSpace(s.app.projectDir); err != nil {
		return synctransport.FastSegmentReceipt{}, err
	}
	location, err := s.app.warehouseLocation(filepath.Join(s.app.projectDir, "warehouse"), conn)
	if err != nil {
		return synctransport.FastSegmentReceipt{}, err
	}
	segment, err := warehouse.WriteArrowSegment(ctx, location, request.Stream, warehouse.ArrowSegmentRequest{
		ID: request.SegmentID, TransformPlanHash: request.TransformPlanHash,
		SourceLogicalBytes: request.SourceLogicalBytes, SourceRows: request.SourceRows,
	}, request.Record)
	if err != nil {
		return synctransport.FastSegmentReceipt{}, err
	}
	manifest := segment.Manifest
	if manifest.ID != request.SegmentID || manifest.TransformPlanHash != request.TransformPlanHash {
		return synctransport.FastSegmentReceipt{}, fmt.Errorf("Arrow segment receipt identity changed while publishing")
	}
	return synctransport.FastSegmentReceipt{
		ID: manifest.ID, SchemaHash: manifest.SchemaHash, TransformPlanHash: manifest.TransformPlanHash,
		ContentSHA256: manifest.ContentSHA256, ParquetSHA256: manifest.ParquetSHA256,
		SourceLogicalBytes: manifest.SourceLogicalBytes, SourceRows: manifest.SourceRows,
		TransformedRows: manifest.TransformedRows, TransformedBytes: manifest.TransformedBytes, ParquetBytes: manifest.ParquetBytes,
	}, nil
}

var _ synctransport.FastSegmentStore = (*connectionArrowSegmentStore)(nil)
