package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

const (
	changeDeliveryWorksetFormatVersion = 1
	changeDeliveryWorksetDomain        = "polymetrics-change-delivery-workset-v1"
	changeDeliveryWorksetBufferSize    = 32 * 1024
	changeDeliveryWorksetHardMaxBytes  = int64(1 << 30)
	changeDeliveryManifestMaxBytes     = int64(1 << 20)

	changeDeliveryProjectionFile = "projection.parquet"
	changeDeliveryDeltaFile      = "delta.parquet"
	changeDeliveryBaselineFile   = "candidate-baseline.parquet"
	changeDeliveryTombstoneFile  = "tombstones.jsonl"
	changeDeliveryManifestFile   = "manifest.json"
)

var (
	// ErrChangeDeliveryWorksetInvalid refuses a request that cannot prove its
	// immutable target binding, finite explicit key contract, or source/baseline
	// input. It is returned before any final workset can be published.
	ErrChangeDeliveryWorksetInvalid = errors.New("change delivery workset is invalid")
	// ErrChangeDeliveryWorksetUnavailable hides filesystem and DuckDB detail
	// from callers. A corrupt or partial artifact is never reusable evidence.
	ErrChangeDeliveryWorksetUnavailable = errors.New("change delivery workset is unavailable")
)

// ChangeDeliveryWorksetRequest carries only the typed target assertion and
// immutable-source inputs needed to derive an outbound workset. It deliberately
// has no destination table text, target handle, receipt, checkpoint, or write
// session: those belong to the later target-delivery slice.
type ChangeDeliveryWorksetRequest struct {
	Control         ManagedTargetControlRecord
	Keys            []string
	SourceParquet   string
	BaselineParquet string
	Tombstones      []synccontract.Tombstone
	Root            string
	// MaxArtifactBytes is a required finite ceiling for every input/output
	// artifact. A workset cannot turn a source snapshot into an unbounded local
	// spool merely because its caller omitted resource policy.
	MaxArtifactBytes int64
}

// ChangeDeliveryWorkset is a sealed on-disk Parquet workset. Its artifact
// directory and manifest are private so callers cannot replace a source path
// or mutate a candidate baseline through the value after derivation.
type ChangeDeliveryWorkset struct {
	dir      string
	manifest changeDeliveryWorksetManifest
}

// Identity returns the deterministic SHA-256 address of the sealed workset.
// It is bound to immutable target identity, schema, ordered key mapping, and
// source/baseline/tombstone versions; it never contains a display or table
// name.
func (w ChangeDeliveryWorkset) Identity() string { return w.manifest.Identity }

// ContentSHA256 returns the SHA-256 digest over the sealed output artifacts.
func (w ChangeDeliveryWorkset) ContentSHA256() string { return w.manifest.ContentSHA256 }

// Records returns the count in the complete immutable staged workset.
func (w ChangeDeliveryWorkset) Records() int64 { return w.manifest.ProjectionRecords }

// Changes returns the count in the keyed insert/update delta projection.
func (w ChangeDeliveryWorkset) Changes() int64 { return w.manifest.DeltaRecords }

// TombstoneCount returns the number of explicitly supplied tombstones. A
// physical source-row absence never changes this count.
func (w ChangeDeliveryWorkset) TombstoneCount() int64 { return w.manifest.TombstoneRecords }

// ReadProjection opens the sealed complete source snapshot without exposing
// its artifact path for mutation.
func (w ChangeDeliveryWorkset) ReadProjection(ctx context.Context, emit func(warehouse.Row) error) error {
	return w.readParquet(ctx, changeDeliveryProjectionFile, emit)
}

// ReadDelta opens the sealed keyed insert/update projection without exposing
// its artifact path for mutation.
func (w ChangeDeliveryWorkset) ReadDelta(ctx context.Context, emit func(warehouse.Row) error) error {
	return w.readParquet(ctx, changeDeliveryDeltaFile, emit)
}

// ReadCandidateBaseline opens the unpromoted baseline candidate. Derivation
// never changes the caller's existing baseline; a later receipt-bound target
// delivery operation decides whether this candidate becomes durable state.
func (w ChangeDeliveryWorkset) ReadCandidateBaseline(ctx context.Context, emit func(warehouse.Row) error) error {
	return w.readParquet(ctx, changeDeliveryBaselineFile, emit)
}

func (w ChangeDeliveryWorkset) readParquet(ctx context.Context, name string, emit func(warehouse.Row) error) error {
	if ctx == nil || w.manifest.validate() != nil || w.dir == "" || emit == nil {
		return ErrChangeDeliveryWorksetInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := requireChangeDeliveryRegularFile(filepath.Join(w.dir, name), w.manifest.MaxArtifactBytes); err != nil {
		return err
	}
	if err := warehouse.ReadTable(ctx, filepath.Join(w.dir, name), emit); err != nil {
		return ErrChangeDeliveryWorksetUnavailable
	}
	return nil
}

// Tombstones returns independent copies of only the explicit tombstones sealed
// in this workset. It never derives a delete from baseline/source comparison.
func (w ChangeDeliveryWorkset) Tombstones(ctx context.Context) ([]synccontract.Tombstone, error) {
	if ctx == nil || w.manifest.validate() != nil || w.dir == "" {
		return nil, ErrChangeDeliveryWorksetInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	tombstones, err := readChangeDeliveryTombstones(ctx, filepath.Join(w.dir, changeDeliveryTombstoneFile), w.manifest.MaxArtifactBytes)
	if err != nil {
		return nil, err
	}
	return tombstones, nil
}

// DeriveChangeDeliveryWorkset seals a complete staged input, a keyed
// insert/update delta, explicit tombstones, and an unpromoted candidate
// baseline. All artifacts are written to a temporary directory and atomically
// published under their content-addressed identity only after every file and
// manifest validates.
func DeriveChangeDeliveryWorkset(ctx context.Context, request ChangeDeliveryWorksetRequest) (ChangeDeliveryWorkset, error) {
	if ctx == nil {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetInvalid
	}
	if err := ctx.Err(); err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	key, keys, tombstones, err := validateChangeDeliveryWorksetRequest(request)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	if err := os.MkdirAll(request.Root, 0o700); err != nil {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	staging, err := os.MkdirTemp(request.Root, ".change-delivery-workset-")
	if err != nil {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(staging)
		}
	}()

	projectionPath := filepath.Join(staging, changeDeliveryProjectionFile)
	sourceSHA256, err := copyChangeDeliveryFile(ctx, request.SourceParquet, projectionPath, request.MaxArtifactBytes)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	baselineSHA256, err := changeDeliveryBaselineSHA256(ctx, request.BaselineParquet, request.MaxArtifactBytes)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	baselineForDelta, err := changeDeliveryParquetInput(request.BaselineParquet)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	tombstonePath := filepath.Join(staging, changeDeliveryTombstoneFile)
	tombstoneSHA256, tombstoneRecords, err := writeChangeDeliveryTombstones(ctx, tombstonePath, tombstones, request.MaxArtifactBytes)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}

	binding := changeDeliveryBindingFromLedgerKey(key)
	keySHA256 := changeDeliveryKeySHA256(keys)
	identity := changeDeliveryIdentity(binding, request.Control.Schema(), keys, sourceSHA256, baselineSHA256, tombstoneSHA256)
	finalDir := filepath.Join(request.Root, identity)
	if _, err := os.Stat(finalDir); err == nil {
		workset, err := openChangeDeliveryWorkset(ctx, finalDir, identity, request.MaxArtifactBytes)
		if err != nil {
			return ChangeDeliveryWorkset{}, err
		}
		return workset, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}

	if err := ctx.Err(); err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	deltaPath := filepath.Join(staging, changeDeliveryDeltaFile)
	if err := deriveChangeDeliveryDelta(ctx, projectionPath, baselineForDelta, keys, deltaPath, request.MaxArtifactBytes); err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	deltaSHA256, err := hashChangeDeliveryFile(ctx, deltaPath, request.MaxArtifactBytes)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	candidateBaselinePath := filepath.Join(staging, changeDeliveryBaselineFile)
	candidateBaselineSHA256, err := copyChangeDeliveryFile(ctx, projectionPath, candidateBaselinePath, request.MaxArtifactBytes)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	projectionRecords, err := countChangeDeliveryParquet(ctx, projectionPath)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	deltaRecords, err := countChangeDeliveryParquet(ctx, deltaPath)
	if err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	manifest := changeDeliveryWorksetManifest{
		Version:                 changeDeliveryWorksetFormatVersion,
		Identity:                identity,
		Binding:                 binding,
		SchemaVersion:           request.Control.Schema().Version(),
		SchemaFingerprint:       request.Control.Schema().Fingerprint().String(),
		Keys:                    append([]string(nil), keys...),
		KeySHA256:               keySHA256,
		SourceSHA256:            sourceSHA256,
		BaselineSHA256:          baselineSHA256,
		ProjectionSHA256:        sourceSHA256,
		DeltaSHA256:             deltaSHA256,
		TombstoneSHA256:         tombstoneSHA256,
		CandidateBaselineSHA256: candidateBaselineSHA256,
		ProjectionRecords:       projectionRecords,
		DeltaRecords:            deltaRecords,
		TombstoneRecords:        tombstoneRecords,
		MaxArtifactBytes:        request.MaxArtifactBytes,
	}
	manifest.ContentSHA256 = changeDeliveryContentSHA256(manifest.ProjectionSHA256, manifest.DeltaSHA256, manifest.TombstoneSHA256, manifest.CandidateBaselineSHA256)
	if manifest.validate() != nil {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetInvalid
	}
	if err := writeChangeDeliveryManifest(ctx, filepath.Join(staging, changeDeliveryManifestFile), manifest); err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	if err := ctx.Err(); err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	if err := os.Rename(staging, finalDir); err != nil {
		if errors.Is(err, os.ErrExist) {
			return openChangeDeliveryWorkset(ctx, finalDir, identity, request.MaxArtifactBytes)
		}
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	published = true
	return openChangeDeliveryWorkset(ctx, finalDir, identity, request.MaxArtifactBytes)
}

func validateChangeDeliveryWorksetRequest(request ChangeDeliveryWorksetRequest) (ManagedTargetDeliveryLedgerKey, []string, []synccontract.Tombstone, error) {
	if request.Control.validate() != nil || request.Root == "" || request.SourceParquet == "" || !validChangeDeliveryArtifactLimit(request.MaxArtifactBytes) {
		return ManagedTargetDeliveryLedgerKey{}, nil, nil, ErrChangeDeliveryWorksetInvalid
	}
	if err := requireChangeDeliveryRegularFile(request.SourceParquet, request.MaxArtifactBytes); err != nil {
		return ManagedTargetDeliveryLedgerKey{}, nil, nil, err
	}
	if request.BaselineParquet != "" {
		if err := requireChangeDeliveryRegularFile(request.BaselineParquet, request.MaxArtifactBytes); err != nil {
			return ManagedTargetDeliveryLedgerKey{}, nil, nil, err
		}
	}
	keys, err := normalizeDatabaseWriteKeys(request.Keys)
	if err != nil || len(keys) == 0 {
		return ManagedTargetDeliveryLedgerKey{}, nil, nil, ErrChangeDeliveryWorksetInvalid
	}
	clonedTombstones := make([]synccontract.Tombstone, len(request.Tombstones))
	for index, tombstone := range request.Tombstones {
		if err := tombstone.Validate(); err != nil {
			return ManagedTargetDeliveryLedgerKey{}, nil, nil, ErrChangeDeliveryWorksetInvalid
		}
		clonedTombstones[index] = tombstone.Clone()
	}
	key, err := NewManagedTargetDeliveryLedgerKey(request.Control)
	if err != nil {
		return ManagedTargetDeliveryLedgerKey{}, nil, nil, ErrChangeDeliveryWorksetInvalid
	}
	return key, keys, clonedTombstones, nil
}

func requireChangeDeliveryRegularFile(path string, maximum int64) error {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maximum {
		return ErrChangeDeliveryWorksetInvalid
	}
	return nil
}

// changeDeliveryParquetInput maps the warehouse's documented zero-byte empty
// table representation to the absence of baseline rows. DuckDB's read_parquet
// cannot open that representation directly, while warehouse.ReadTable treats
// it as an empty table by contract.
func changeDeliveryParquetInput(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", ErrChangeDeliveryWorksetUnavailable
	}
	if info.Size() == 0 {
		return "", nil
	}
	return path, nil
}

func copyChangeDeliveryFile(ctx context.Context, source, destination string, maximum int64) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	in, err := os.Open(source)
	if err != nil {
		return "", ErrChangeDeliveryWorksetUnavailable
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", ErrChangeDeliveryWorksetUnavailable
	}
	hash := sha256.New()
	buffer := make([]byte, changeDeliveryWorksetBufferSize)
	var copied int64
	for {
		if err := ctx.Err(); err != nil {
			_ = out.Close()
			return "", err
		}
		read, readErr := in.Read(buffer)
		if read > 0 {
			copied += int64(read)
			if copied > maximum {
				_ = out.Close()
				return "", ErrChangeDeliveryWorksetInvalid
			}
			if _, err := out.Write(buffer[:read]); err != nil {
				_ = out.Close()
				return "", ErrChangeDeliveryWorksetUnavailable
			}
			if _, err := hash.Write(buffer[:read]); err != nil {
				_ = out.Close()
				return "", ErrChangeDeliveryWorksetUnavailable
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			_ = out.Close()
			return "", ErrChangeDeliveryWorksetUnavailable
		}
	}
	if err := out.Close(); err != nil {
		return "", ErrChangeDeliveryWorksetUnavailable
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func hashChangeDeliveryFile(ctx context.Context, path string, maximum int64) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	file, err := os.Open(path)
	if err != nil {
		return "", ErrChangeDeliveryWorksetUnavailable
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	buffer := make([]byte, changeDeliveryWorksetBufferSize)
	var readTotal int64
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		read, readErr := file.Read(buffer)
		if read > 0 {
			readTotal += int64(read)
			if readTotal > maximum {
				return "", ErrChangeDeliveryWorksetInvalid
			}
			if _, err := hash.Write(buffer[:read]); err != nil {
				return "", ErrChangeDeliveryWorksetUnavailable
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", ErrChangeDeliveryWorksetUnavailable
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func changeDeliveryBaselineSHA256(ctx context.Context, path string, maximum int64) (string, error) {
	if path == "" {
		hash := sha256.Sum256([]byte(changeDeliveryWorksetDomain + "\x00no-baseline"))
		return hex.EncodeToString(hash[:]), nil
	}
	return hashChangeDeliveryFile(ctx, path, maximum)
}

func writeChangeDeliveryTombstones(ctx context.Context, path string, tombstones []synccontract.Tombstone, maximum int64) (string, int64, error) {
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", 0, ErrChangeDeliveryWorksetUnavailable
	}
	hash := sha256.New()
	limited := &changeDeliveryLimitedWriter{writer: io.MultiWriter(file, hash), maximum: maximum}
	encoder := json.NewEncoder(limited)
	for index := range tombstones {
		if err := ctx.Err(); err != nil {
			_ = file.Close()
			return "", 0, err
		}
		if err := encoder.Encode(tombstones[index]); err != nil {
			_ = file.Close()
			if errors.Is(err, ErrChangeDeliveryWorksetInvalid) {
				return "", 0, ErrChangeDeliveryWorksetInvalid
			}
			return "", 0, ErrChangeDeliveryWorksetUnavailable
		}
	}
	if err := file.Close(); err != nil {
		return "", 0, ErrChangeDeliveryWorksetUnavailable
	}
	return hex.EncodeToString(hash.Sum(nil)), int64(len(tombstones)), nil
}

type changeDeliveryLimitedWriter struct {
	writer  io.Writer
	maximum int64
	written int64
}

func (w *changeDeliveryLimitedWriter) Write(bytes []byte) (int, error) {
	if int64(len(bytes)) > w.maximum-w.written {
		return 0, ErrChangeDeliveryWorksetInvalid
	}
	written, err := w.writer.Write(bytes)
	w.written += int64(written)
	return written, err
}

func readChangeDeliveryTombstones(ctx context.Context, path string, maximum int64) ([]synccontract.Tombstone, error) {
	if err := requireChangeDeliveryRegularFile(path, maximum); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, ErrChangeDeliveryWorksetUnavailable
	}
	defer func() { _ = file.Close() }()
	decoder := json.NewDecoder(file)
	var tombstones []synccontract.Tombstone
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var tombstone synccontract.Tombstone
		err := decoder.Decode(&tombstone)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil || tombstone.Validate() != nil {
			return nil, ErrChangeDeliveryWorksetUnavailable
		}
		tombstones = append(tombstones, tombstone.Clone())
	}
	return tombstones, nil
}

func deriveChangeDeliveryDelta(ctx context.Context, source, baseline string, keys []string, destination string, maximum int64) error {
	source, err := changeDeliveryParquetInput(source)
	if err != nil {
		return err
	}
	if source == "" {
		return writeChangeDeliveryEmptyFile(ctx, destination)
	}
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return ErrChangeDeliveryWorksetUnavailable
	}
	defer func() { _ = db.Close() }()
	if err := assertChangeDeliveryKeys(ctx, db, source, keys); err != nil {
		return err
	}
	if baseline != "" {
		if err := assertChangeDeliveryKeys(ctx, db, baseline, keys); err != nil {
			return err
		}
	}
	query := changeDeliveryDeltaQuery(source, baseline, keys)
	statement := "COPY (" + query + ") TO '" + warehouse.EscapeSQLLiteral(destination) + "' (FORMAT parquet, COMPRESSION " + warehouse.ParquetCompression + ")"
	if _, err := db.ExecContext(ctx, statement); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return ErrChangeDeliveryWorksetUnavailable
	}
	if _, err := hashChangeDeliveryFile(ctx, destination, maximum); err != nil {
		return err
	}
	return nil
}

func writeChangeDeliveryEmptyFile(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return ErrChangeDeliveryWorksetUnavailable
	}
	if err := file.Close(); err != nil {
		return ErrChangeDeliveryWorksetUnavailable
	}
	return nil
}

func assertChangeDeliveryKeys(ctx context.Context, db *sql.DB, path string, keys []string) error {
	identifierList := changeDeliveryIdentifierList("row", keys)
	nullChecks := make([]string, len(keys))
	for index, key := range keys {
		nullChecks[index] = "row." + quoteChangeDeliveryIdentifier(key) + " IS NULL"
	}
	for _, query := range []string{
		"SELECT 1 FROM read_parquet('" + warehouse.EscapeSQLLiteral(path) + "') AS row WHERE " + strings.Join(nullChecks, " OR ") + " LIMIT 1",
		"SELECT 1 FROM read_parquet('" + warehouse.EscapeSQLLiteral(path) + "') AS row GROUP BY " + identifierList + " HAVING COUNT(*) > 1 LIMIT 1",
	} {
		var found int
		err := db.QueryRowContext(ctx, query).Scan(&found)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			continue
		case err == nil:
			return ErrChangeDeliveryWorksetInvalid
		case ctx.Err() != nil:
			return ctx.Err()
		default:
			return ErrChangeDeliveryWorksetUnavailable
		}
	}
	return nil
}

func changeDeliveryDeltaQuery(source, baseline string, keys []string) string {
	order := " ORDER BY " + changeDeliveryIdentifierList("source", keys)
	if baseline == "" {
		return "SELECT source.* FROM read_parquet('" + warehouse.EscapeSQLLiteral(source) + "') AS source" + order
	}
	joins := make([]string, len(keys))
	for index, key := range keys {
		quoted := quoteChangeDeliveryIdentifier(key)
		joins[index] = "source." + quoted + " IS NOT DISTINCT FROM baseline." + quoted
	}
	return "SELECT source.* FROM read_parquet('" + warehouse.EscapeSQLLiteral(source) + "') AS source " +
		"WHERE NOT EXISTS (SELECT 1 FROM read_parquet('" + warehouse.EscapeSQLLiteral(baseline) + "') AS baseline WHERE " + strings.Join(joins, " AND ") + " AND to_json(source) = to_json(baseline))" + order
}

func countChangeDeliveryParquet(ctx context.Context, path string) (int64, error) {
	input, err := changeDeliveryParquetInput(path)
	if err != nil {
		return 0, err
	}
	if input == "" {
		return 0, nil
	}
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return 0, ErrChangeDeliveryWorksetUnavailable
	}
	defer func() { _ = db.Close() }()
	var count int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM read_parquet('"+warehouse.EscapeSQLLiteral(input)+"')").Scan(&count); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, ctxErr
		}
		return 0, ErrChangeDeliveryWorksetUnavailable
	}
	return count, nil
}

func quoteChangeDeliveryIdentifier(identifier string) string {
	return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
}

func changeDeliveryIdentifierList(alias string, keys []string) string {
	identifiers := make([]string, len(keys))
	for index, key := range keys {
		identifiers[index] = alias + "." + quoteChangeDeliveryIdentifier(key)
	}
	return strings.Join(identifiers, ", ")
}

type changeDeliveryWorksetBinding struct {
	WorkspaceID        string `json:"workspace_id"`
	ConnectorID        string `json:"connector_id"`
	ConnectionID       string `json:"connection_id"`
	TargetDatabaseKind string `json:"target_database_kind"`
	TargetDatabaseID   string `json:"target_database_id"`
	StreamID           string `json:"stream_id"`
	Namespace          string `json:"namespace"`
	Relation           string `json:"relation"`
}

func changeDeliveryBindingFromLedgerKey(key ManagedTargetDeliveryLedgerKey) changeDeliveryWorksetBinding {
	owner := key.Owner().Identity()
	targetDatabase := key.TargetDatabase()
	return changeDeliveryWorksetBinding{
		WorkspaceID:        owner.WorkspaceID,
		ConnectorID:        owner.ConnectorID,
		ConnectionID:       owner.ConnectionID,
		TargetDatabaseKind: targetDatabase.Kind(),
		TargetDatabaseID:   targetDatabase.Value(),
		StreamID:           key.StreamID(),
		Namespace:          key.Namespace(),
		Relation:           key.Relation(),
	}
}

func (b changeDeliveryWorksetBinding) validate() error {
	owner := TargetOwner{identity: ConnectionIdentity{WorkspaceID: b.WorkspaceID, ConnectorID: b.ConnectorID, ConnectionID: b.ConnectionID}}
	targetDatabase := TargetDatabaseIdentity{kind: b.TargetDatabaseKind, value: b.TargetDatabaseID}
	if owner.validate() != nil || targetDatabase.validate() != nil || !validOpaqueIdentityComponent(b.StreamID) {
		return ErrChangeDeliveryWorksetInvalid
	}
	wantNamespace := deriveManagedTargetName("namespace", b.WorkspaceID, b.ConnectorID, b.ConnectionID)
	wantRelation := deriveManagedTargetName("relation", b.WorkspaceID, b.ConnectorID, b.ConnectionID, b.StreamID)
	if b.Namespace != wantNamespace || b.Relation != wantRelation || validateIdentifierComponent(b.Namespace) != nil || validateIdentifierComponent(b.Relation) != nil {
		return ErrChangeDeliveryWorksetInvalid
	}
	return nil
}

type changeDeliveryWorksetManifest struct {
	Version                 int                          `json:"version"`
	Identity                string                       `json:"identity"`
	Binding                 changeDeliveryWorksetBinding `json:"binding"`
	SchemaVersion           uint                         `json:"schema_version"`
	SchemaFingerprint       string                       `json:"schema_fingerprint"`
	Keys                    []string                     `json:"keys"`
	KeySHA256               string                       `json:"key_sha256"`
	SourceSHA256            string                       `json:"source_sha256"`
	BaselineSHA256          string                       `json:"baseline_sha256"`
	ProjectionSHA256        string                       `json:"projection_sha256"`
	DeltaSHA256             string                       `json:"delta_sha256"`
	TombstoneSHA256         string                       `json:"tombstone_sha256"`
	CandidateBaselineSHA256 string                       `json:"candidate_baseline_sha256"`
	ContentSHA256           string                       `json:"content_sha256"`
	ProjectionRecords       int64                        `json:"projection_records"`
	DeltaRecords            int64                        `json:"delta_records"`
	TombstoneRecords        int64                        `json:"tombstone_records"`
	MaxArtifactBytes        int64                        `json:"max_artifact_bytes"`
}

func (m changeDeliveryWorksetManifest) validate() error {
	if m.Version != changeDeliveryWorksetFormatVersion || m.Binding.validate() != nil || m.SchemaVersion == 0 || !validChangeDeliveryArtifactLimit(m.MaxArtifactBytes) || !validChangeDeliverySHA256(m.SchemaFingerprint) || !validChangeDeliverySHA256(m.Identity) || !validChangeDeliverySHA256(m.KeySHA256) || !validChangeDeliverySHA256(m.SourceSHA256) || !validChangeDeliverySHA256(m.BaselineSHA256) || !validChangeDeliverySHA256(m.ProjectionSHA256) || !validChangeDeliverySHA256(m.DeltaSHA256) || !validChangeDeliverySHA256(m.TombstoneSHA256) || !validChangeDeliverySHA256(m.CandidateBaselineSHA256) || !validChangeDeliverySHA256(m.ContentSHA256) || m.ProjectionRecords < 0 || m.DeltaRecords < 0 || m.TombstoneRecords < 0 {
		return ErrChangeDeliveryWorksetInvalid
	}
	keys, err := normalizeDatabaseWriteKeys(m.Keys)
	if err != nil || len(keys) == 0 || len(keys) != len(m.Keys) {
		return ErrChangeDeliveryWorksetInvalid
	}
	for index := range keys {
		if keys[index] != m.Keys[index] {
			return ErrChangeDeliveryWorksetInvalid
		}
	}
	if m.KeySHA256 != changeDeliveryKeySHA256(m.Keys) || m.Identity != changeDeliveryIdentityFromManifest(m) || m.ContentSHA256 != changeDeliveryContentSHA256(m.ProjectionSHA256, m.DeltaSHA256, m.TombstoneSHA256, m.CandidateBaselineSHA256) || m.SourceSHA256 != m.ProjectionSHA256 || m.ProjectionSHA256 != m.CandidateBaselineSHA256 {
		return ErrChangeDeliveryWorksetInvalid
	}
	return nil
}

func writeChangeDeliveryManifest(ctx context.Context, path string, manifest changeDeliveryWorksetManifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	bytes, err := json.Marshal(manifest)
	if err != nil {
		return ErrChangeDeliveryWorksetUnavailable
	}
	if err := os.WriteFile(path, bytes, 0o600); err != nil {
		return ErrChangeDeliveryWorksetUnavailable
	}
	return nil
}

func openChangeDeliveryWorkset(ctx context.Context, dir, identity string, maximum int64) (ChangeDeliveryWorkset, error) {
	if err := ctx.Err(); err != nil {
		return ChangeDeliveryWorkset{}, err
	}
	manifestPath := filepath.Join(dir, changeDeliveryManifestFile)
	manifestInfo, err := os.Stat(manifestPath)
	if err != nil || !manifestInfo.Mode().IsRegular() || manifestInfo.Size() > changeDeliveryManifestMaxBytes {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	bytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	var manifest changeDeliveryWorksetManifest
	if err := json.Unmarshal(bytes, &manifest); err != nil || manifest.validate() != nil || manifest.Identity != identity {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	if !validChangeDeliveryArtifactLimit(maximum) {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetInvalid
	}
	projectionSHA256, err := hashChangeDeliveryFile(ctx, filepath.Join(dir, changeDeliveryProjectionFile), maximum)
	if err != nil || projectionSHA256 != manifest.ProjectionSHA256 {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	deltaSHA256, err := hashChangeDeliveryFile(ctx, filepath.Join(dir, changeDeliveryDeltaFile), maximum)
	if err != nil || deltaSHA256 != manifest.DeltaSHA256 {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	tombstoneSHA256, err := hashChangeDeliveryFile(ctx, filepath.Join(dir, changeDeliveryTombstoneFile), maximum)
	if err != nil || tombstoneSHA256 != manifest.TombstoneSHA256 {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	baselineSHA256, err := hashChangeDeliveryFile(ctx, filepath.Join(dir, changeDeliveryBaselineFile), maximum)
	if err != nil || baselineSHA256 != manifest.CandidateBaselineSHA256 {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	projectionRecords, err := countChangeDeliveryParquet(ctx, filepath.Join(dir, changeDeliveryProjectionFile))
	if err != nil || projectionRecords != manifest.ProjectionRecords {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	deltaRecords, err := countChangeDeliveryParquet(ctx, filepath.Join(dir, changeDeliveryDeltaFile))
	if err != nil || deltaRecords != manifest.DeltaRecords {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	tombstones, err := readChangeDeliveryTombstones(ctx, filepath.Join(dir, changeDeliveryTombstoneFile), maximum)
	if err != nil || int64(len(tombstones)) != manifest.TombstoneRecords {
		return ChangeDeliveryWorkset{}, ErrChangeDeliveryWorksetUnavailable
	}
	return ChangeDeliveryWorkset{dir: dir, manifest: manifest}, nil
}

func changeDeliveryKeySHA256(keys []string) string {
	return changeDeliveryHash(changeDeliveryWorksetDomain+"\x00keys", keys...)
}

func changeDeliveryIdentity(binding changeDeliveryWorksetBinding, schema ManagedTargetSchema, keys []string, sourceSHA256, baselineSHA256, tombstoneSHA256 string) string {
	return changeDeliveryIdentityFields(binding, schema.Version(), schema.Fingerprint().String(), keys, sourceSHA256, baselineSHA256, tombstoneSHA256)
}

func changeDeliveryIdentityFromManifest(manifest changeDeliveryWorksetManifest) string {
	return changeDeliveryIdentityFields(manifest.Binding, manifest.SchemaVersion, manifest.SchemaFingerprint, manifest.Keys, manifest.SourceSHA256, manifest.BaselineSHA256, manifest.TombstoneSHA256)
}

func changeDeliveryIdentityFields(binding changeDeliveryWorksetBinding, schemaVersion uint, schemaFingerprint string, keys []string, sourceSHA256, baselineSHA256, tombstoneSHA256 string) string {
	components := []string{
		binding.WorkspaceID,
		binding.ConnectorID,
		binding.ConnectionID,
		binding.TargetDatabaseKind,
		binding.TargetDatabaseID,
		binding.StreamID,
		binding.Namespace,
		binding.Relation,
		strconv.FormatUint(uint64(schemaVersion), 10),
		schemaFingerprint,
		sourceSHA256,
		baselineSHA256,
		tombstoneSHA256,
	}
	components = append(components, keys...)
	return changeDeliveryHash(changeDeliveryWorksetDomain+"\x00identity", components...)
}

func changeDeliveryContentSHA256(projection, delta, tombstone, candidateBaseline string) string {
	return changeDeliveryHash(changeDeliveryWorksetDomain+"\x00content", projection, delta, tombstone, candidateBaseline)
}

func changeDeliveryHash(domain string, components ...string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(domain))
	_, _ = hash.Write([]byte{0})
	for _, component := range components {
		writeManagedTargetHashComponent(hash, component)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func validChangeDeliverySHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validChangeDeliveryArtifactLimit(value int64) bool {
	return value > 0 && value <= changeDeliveryWorksetHardMaxBytes
}
