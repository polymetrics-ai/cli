package postgres

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pglogrepl"

	"polymetrics.ai/internal/synccontract"
)

const postgresBootstrapBarrierVersion uint = 1

// postgresBootstrapBarrier is sealed into the opaque snapshot-barrier token.
// The token remains provider-owned state, while this private shape makes a
// resumed slot prove it still describes the same relation/schema observation.
type postgresBootstrapBarrier struct {
	Version           uint   `json:"version"`
	InitialLSN        string `json:"initial_lsn"`
	SystemID          string `json:"system_id"`
	Timeline          int32  `json:"timeline"`
	Publication       string `json:"publication"`
	Relation          string `json:"relation"`
	SchemaFingerprint string `json:"schema_fingerprint"`
}

func newPostgresBootstrapBarrier(source postgresCDCSource, initial pglogrepl.LSN) (postgresBootstrapBarrier, error) {
	barrier := postgresBootstrapBarrier{
		Version:           postgresBootstrapBarrierVersion,
		InitialLSN:        initial.String(),
		SystemID:          source.system.SystemID,
		Timeline:          source.system.Timeline,
		Publication:       source.publication,
		Relation:          source.identity.ObjectScope,
		SchemaFingerprint: source.schemaFingerprint,
	}
	if err := barrier.validate(); err != nil {
		return postgresBootstrapBarrier{}, err
	}
	if source.identity.Validate() != nil || source.identity.AccountOrCluster != source.system.SystemID+":"+source.system.DBName || source.identity.ObjectScope != barrier.Relation {
		return postgresBootstrapBarrier{}, errors.New("postgres bootstrap barrier source identity is incomplete")
	}
	return barrier, nil
}

func (b postgresBootstrapBarrier) validate() error {
	if b.Version != postgresBootstrapBarrierVersion || strings.TrimSpace(b.InitialLSN) == "" || strings.TrimSpace(b.SystemID) == "" || b.Timeline <= 0 || strings.TrimSpace(b.Publication) == "" || strings.TrimSpace(b.Relation) == "" || strings.TrimSpace(b.SchemaFingerprint) == "" {
		return errors.New("postgres bootstrap barrier is incomplete")
	}
	if _, err := pglogrepl.ParseLSN(b.InitialLSN); err != nil {
		return errors.New("postgres bootstrap barrier has an invalid LSN")
	}
	if err := validateIdentifier(b.Publication); err != nil {
		return errors.New("postgres bootstrap barrier has an invalid publication")
	}
	if _, err := canonicalCDCStream("", b.Relation); err != nil {
		return errors.New("postgres bootstrap barrier has an invalid relation")
	}
	if strings.ContainsAny(b.SchemaFingerprint, "\x00\r\n") || len(b.SchemaFingerprint) > 512 {
		return errors.New("postgres bootstrap barrier has an invalid schema fingerprint")
	}
	return nil
}

func (b postgresBootstrapBarrier) token() synccontract.OpaqueToken {
	encoded, err := json.Marshal(b)
	if err != nil {
		panic("postgres bootstrap barrier is not serializable: " + err.Error())
	}
	return append(synccontract.OpaqueToken(nil), encoded...)
}

func postgresBootstrapBarrierFromCheckpoint(checkpoint *synccontract.CheckpointEnvelope) (postgresBootstrapBarrier, bool, error) {
	if checkpoint == nil || checkpoint.SnapshotBarrier == nil || len(checkpoint.SnapshotBarrier.Token) == 0 {
		return postgresBootstrapBarrier{}, false, nil
	}
	token := bytes.TrimSpace(checkpoint.SnapshotBarrier.Token)
	if len(token) == 0 || token[0] != '{' {
		return postgresBootstrapBarrier{}, false, nil
	}
	var barrier postgresBootstrapBarrier
	if err := json.Unmarshal(token, &barrier); err != nil || barrier.validate() != nil {
		return postgresBootstrapBarrier{}, false, fmt.Errorf("decode PostgreSQL bootstrap barrier: %w", errors.New("invalid barrier"))
	}
	return barrier, true, nil
}
