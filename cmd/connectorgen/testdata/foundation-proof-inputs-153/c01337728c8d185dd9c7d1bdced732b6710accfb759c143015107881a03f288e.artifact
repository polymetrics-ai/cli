package warehouse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

// ArrowSegmentRequest supplies bounded, connector-neutral accounting for one
// already transformed Arrow record. It deliberately carries no connector
// configuration, source query, target identifier, or row representation.
type ArrowSegmentRequest struct {
	ID                 string
	TransformPlanHash  string
	SourceLogicalBytes int64
	SourceRows         int64
}

// ArrowSegment is a reopened, integrity-checked local representation of the
// immutable SegmentManifestV1. Path is derived from the owned warehouse
// location and the manifest filename; it is never persisted as user input.
type ArrowSegment struct {
	Manifest SegmentManifestV1
	Path     string
}

// WriteArrowSegment closes and fsyncs a typed Arrow record into a new Parquet
// segment before publishing its immutable manifest. It is entirely separate
// from the legacy JSONL WAL and one-file warehouse table format.
func WriteArrowSegment(ctx context.Context, location Location, stream string, request ArrowSegmentRequest, record arrow.Record) (ArrowSegment, error) {
	if ctx == nil || record == nil || request.SourceLogicalBytes < 0 || request.SourceRows < 0 || request.SourceRows < record.NumRows() || !SafePathPart(request.ID) {
		return ArrowSegment{}, ErrTransportSegmentManifestInvalid
	}
	if err := ctx.Err(); err != nil {
		return ArrowSegment{}, err
	}
	if err := location.EnsureOwnership(); err != nil {
		return ArrowSegment{}, err
	}
	dir, err := location.SegmentDir(stream)
	if err != nil {
		return ArrowSegment{}, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return ArrowSegment{}, err
	}
	filename := request.ID + ".parquet"
	path := filepath.Join(dir, filename)
	if _, err := os.Lstat(path); err == nil {
		return ArrowSegment{}, ErrTransportSegmentManifestInvalid
	} else if !errors.Is(err, os.ErrNotExist) {
		return ArrowSegment{}, err
	}
	temporary, err := os.CreateTemp(dir, ".segment-")
	if err != nil {
		return ArrowSegment{}, err
	}
	temporaryName := temporary.Name()
	cleanup := func() { _ = os.Remove(temporaryName) }
	defer cleanup()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return ArrowSegment{}, err
	}
	writer, err := pqarrow.NewFileWriter(record.Schema(), temporary, parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Zstd)), pqarrow.NewArrowWriterProperties())
	if err != nil {
		_ = temporary.Close()
		return ArrowSegment{}, fmt.Errorf("open Arrow segment writer: %w", err)
	}
	if err := writer.Write(record); err != nil {
		_ = writer.Close()
		_ = temporary.Close()
		return ArrowSegment{}, fmt.Errorf("write Arrow segment: %w", err)
	}
	if err := writer.Close(); err != nil {
		return ArrowSegment{}, fmt.Errorf("close Arrow segment: %w", err)
	}
	// pqarrow closes its io.Writer while sealing the footer. Reopen explicitly
	// for the durability sync instead of Sync'ing a now-closed handle.
	synced, err := os.OpenFile(temporaryName, os.O_RDWR, 0o600)
	if err != nil {
		return ArrowSegment{}, err
	}
	if err := synced.Sync(); err != nil {
		_ = synced.Close()
		return ArrowSegment{}, fmt.Errorf("sync Arrow segment: %w", err)
	}
	if err := synced.Close(); err != nil {
		return ArrowSegment{}, err
	}
	info, err := os.Stat(temporaryName)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 {
		return ArrowSegment{}, ErrTransportSegmentManifestInvalid
	}
	digest, err := digestArrowSegmentFile(temporaryName)
	if err != nil {
		return ArrowSegment{}, err
	}
	if err := os.Link(temporaryName, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return ArrowSegment{}, ErrTransportSegmentManifestInvalid
		}
		return ArrowSegment{}, err
	}
	if err := syncTransportSegmentDirectoryChain(dir, location.ConnectionDir); err != nil {
		return ArrowSegment{}, err
	}
	schemaDigest := sha256.Sum256([]byte(record.Schema().Fingerprint()))
	manifest, err := NewSegmentManifestV1(SegmentManifestV1{
		ID: request.ID, SchemaHash: hex.EncodeToString(schemaDigest[:]), TransformPlanHash: request.TransformPlanHash,
		SourceLogicalBytes: request.SourceLogicalBytes, SourceRows: request.SourceRows,
		TransformedRows: record.NumRows(), TransformedBytes: arrowRecordLogicalBytes(record),
		ParquetFile: filename, ParquetBytes: info.Size(), ParquetSHA256: digest,
	})
	if err != nil {
		return ArrowSegment{}, err
	}
	if _, err := WriteSegmentManifest(ctx, location, stream, manifest); err != nil {
		return ArrowSegment{}, err
	}
	return ArrowSegment{Manifest: manifest, Path: path}, nil
}

// OpenArrowSegment independently verifies the manifest and its sibling
// Parquet file before an adapter can hand it to a bulk apply primitive.
func OpenArrowSegment(ctx context.Context, location Location, stream, id string) (ArrowSegment, error) {
	manifest, err := OpenSegmentManifest(ctx, location, stream, id)
	if err != nil {
		return ArrowSegment{}, err
	}
	dir, err := location.SegmentDir(stream)
	if err != nil {
		return ArrowSegment{}, err
	}
	path := filepath.Join(dir, manifest.ParquetFile)
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() != manifest.ParquetBytes {
		return ArrowSegment{}, ErrTransportSegmentManifestInvalid
	}
	digest, err := digestArrowSegmentFile(path)
	if err != nil || digest != manifest.ParquetSHA256 {
		return ArrowSegment{}, ErrTransportSegmentManifestInvalid
	}
	return ArrowSegment{Manifest: manifest, Path: path}, nil
}

func digestArrowSegmentFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func arrowRecordLogicalBytes(record arrow.Record) int64 {
	var total uint64
	for index := 0; index < int(record.NumCols()); index++ {
		total += record.Column(index).Data().SizeInBytes()
	}
	if total > uint64(^uint64(0)>>1) {
		return int64(^uint64(0) >> 1)
	}
	return int64(total)
}
