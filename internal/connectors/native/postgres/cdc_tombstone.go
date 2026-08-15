package postgres

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pglogrepl"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

var errCDCDeleteTombstone = errors.New("postgres CDC delete cannot form an explicit tombstone")

// CDCDeleteTombstone converts one pgoutput delete event into the source-keyed
// delete envelope consumed by a sealed target mapping. The caller supplies the
// committed-transaction ordinal because one LSN can contain multiple deletes
// of the same key; keeping it in the opaque identity prevents either event
// from being silently coalesced before keyed apply.
func CDCDeleteTombstone(event connectors.CDCEvent, sourceKeys []string, ordinal uint64) (synccontract.Tombstone, error) {
	if event.Operation != "delete" || len(event.Record) == 0 || ordinal == 0 || len(sourceKeys) == 0 {
		return synccontract.Tombstone{}, errCDCDeleteTombstone
	}
	lsn, ok := event.State["lsn"].(string)
	lsn = strings.TrimSpace(lsn)
	if !ok || lsn == "" {
		return synccontract.Tombstone{}, errCDCDeleteTombstone
	}
	if _, err := pglogrepl.ParseLSN(lsn); err != nil {
		return synccontract.Tombstone{}, errCDCDeleteTombstone
	}
	keyValues := make(map[string]any, len(sourceKeys))
	for _, sourceKey := range sourceKeys {
		if strings.TrimSpace(sourceKey) == "" {
			return synccontract.Tombstone{}, errCDCDeleteTombstone
		}
		if _, exists := keyValues[sourceKey]; exists {
			return synccontract.Tombstone{}, errCDCDeleteTombstone
		}
		value, found := event.Record[sourceKey]
		if !found || value == nil {
			return synccontract.Tombstone{}, errCDCDeleteTombstone
		}
		keyValues[sourceKey] = value
	}
	key, err := json.Marshal(keyValues)
	if err != nil || !json.Valid(key) {
		return synccontract.Tombstone{}, errCDCDeleteTombstone
	}
	digest := sha256.Sum256([]byte("postgres-cdc-delete-v1\x00" + lsn + "\x00" + string(key) + "\x00" + strconv.FormatUint(ordinal, 10)))
	eventID := synccontract.OpaqueToken(append([]byte(nil), digest[:]...))
	tombstone := synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     eventID,
		Key:         key,
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken([]byte(lsn)),
			TieBreaker: synccontract.OpaqueToken(append([]byte(nil), digest[:]...)),
		},
	}
	if err := tombstone.Validate(); err != nil {
		return synccontract.Tombstone{}, errCDCDeleteTombstone
	}
	return tombstone, nil
}
