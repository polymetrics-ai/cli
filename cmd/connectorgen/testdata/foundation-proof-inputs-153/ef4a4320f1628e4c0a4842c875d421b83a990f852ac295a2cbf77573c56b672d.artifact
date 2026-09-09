package warehouse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// TransportSegmentVersionV1 is the immutable Parquet segment contract for
	// connector-neutral high-throughput transports. It does not alter the
	// legacy WAL or one-file local warehouse table format.
	TransportSegmentVersionV1 = 1
	transportSegmentsDirName  = "segments"
	maxSegmentManifestBytes   = 64 << 10
)

var ErrTransportSegmentManifestInvalid = errors.New("warehouse transport segment manifest is invalid")

// SegmentManifestV1 describes one immutable typed Parquet segment. It stores
// content identities and counts only: connector implementation, SQL, pgwire,
// credentials, source records, and a mutable filesystem path are deliberately
// absent.
type SegmentManifestV1 struct {
	Version            int    `json:"version"`
	ID                 string `json:"id"`
	SchemaHash         string `json:"schema_hash"`
	TransformPlanHash  string `json:"transform_plan_hash"`
	SourceLogicalBytes int64  `json:"source_logical_bytes"`
	SourceRows         int64  `json:"source_rows"`
	TransformedRows    int64  `json:"transformed_rows"`
	TransformedBytes   int64  `json:"transformed_bytes"`
	ParquetFile        string `json:"parquet_file"`
	ParquetBytes       int64  `json:"parquet_bytes"`
	ParquetSHA256      string `json:"parquet_sha256"`
	ContentSHA256      string `json:"content_sha256"`
}

// NewSegmentManifestV1 seals a segment's immutable accounting and Parquet
// identity. Callers provide a local filename only, never an arbitrary path.
func NewSegmentManifestV1(manifest SegmentManifestV1) (SegmentManifestV1, error) {
	manifest.Version = TransportSegmentVersionV1
	manifest.ContentSHA256 = ""
	if err := manifest.validate(false); err != nil {
		return SegmentManifestV1{}, ErrTransportSegmentManifestInvalid
	}
	content, err := manifest.contentHash()
	if err != nil {
		return SegmentManifestV1{}, ErrTransportSegmentManifestInvalid
	}
	manifest.ContentSHA256 = content
	if err := manifest.validate(true); err != nil {
		return SegmentManifestV1{}, ErrTransportSegmentManifestInvalid
	}
	return manifest, nil
}

// SegmentDir returns the connection-owned location for new transport segments.
// It is independent of the legacy WAL and single-file table locations.
func (l Location) SegmentDir(stream string) (string, error) {
	component, err := PathComponent("stream", stream)
	if err != nil {
		return "", err
	}
	return filepath.Join(l.ConnectionDir, transportSegmentsDirName, component), nil
}

func (l Location) segmentManifestPath(stream, id string) (string, error) {
	if !SafePathPart(id) {
		return "", ErrTransportSegmentManifestInvalid
	}
	dir, err := l.SegmentDir(stream)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "manifest-"+id+".json"), nil
}

// WriteSegmentManifest atomically publishes a previously closed manifest and
// fsyncs its directory chain before reporting the receipt-visible path. It
// refuses an existing identity rather than replacing evidence from another
// attempt or process.
func WriteSegmentManifest(ctx context.Context, location Location, stream string, manifest SegmentManifestV1) (string, error) {
	if ctx == nil || ctx.Err() != nil || manifest.validate(true) != nil {
		return "", ErrTransportSegmentManifestInvalid
	}
	if err := location.EnsureOwnership(); err != nil {
		return "", err
	}
	path, err := location.segmentManifestPath(stream, manifest.ID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	if _, err := os.Lstat(path); err == nil {
		return "", ErrTransportSegmentManifestInvalid
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	encoded, err := json.Marshal(manifest)
	if err != nil || len(encoded) > maxSegmentManifestBytes {
		return "", ErrTransportSegmentManifestInvalid
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".segment-manifest-")
	if err != nil {
		return "", err
	}
	temporaryName := temporary.Name()
	cleanup := func() { _ = os.Remove(temporaryName) }
	defer cleanup()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return "", err
	}
	if _, err := temporary.Write(encoded); err != nil {
		_ = temporary.Close()
		return "", err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	// Link is atomic and refuses an existing destination, unlike Rename which
	// could replace a concurrently published receipt.
	if err := os.Link(temporaryName, path); err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", ErrTransportSegmentManifestInvalid
		}
		return "", err
	}
	if err := syncTransportSegmentDirectoryChain(filepath.Dir(path), location.ConnectionDir); err != nil {
		return "", err
	}
	return path, nil
}

// OpenSegmentManifest independently reopens and re-hashes durable segment
// evidence. Corruption or a forged filename/hash is refused before a bulk
// destination can use it.
func OpenSegmentManifest(ctx context.Context, location Location, stream, id string) (SegmentManifestV1, error) {
	if ctx == nil || ctx.Err() != nil {
		return SegmentManifestV1{}, ErrTransportSegmentManifestInvalid
	}
	if err := location.EnsureOwnership(); err != nil {
		return SegmentManifestV1{}, err
	}
	path, err := location.segmentManifestPath(stream, id)
	if err != nil {
		return SegmentManifestV1{}, err
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxSegmentManifestBytes {
		return SegmentManifestV1{}, ErrTransportSegmentManifestInvalid
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return SegmentManifestV1{}, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	var manifest SegmentManifestV1
	if err := decoder.Decode(&manifest); err != nil || decoder.More() || manifest.ID != id || manifest.validate(true) != nil {
		return SegmentManifestV1{}, ErrTransportSegmentManifestInvalid
	}
	return manifest, nil
}

func (m SegmentManifestV1) validate(requireContentHash bool) error {
	if m.Version != TransportSegmentVersionV1 || !SafePathPart(m.ID) || !validSegmentHash(m.SchemaHash) || (m.TransformPlanHash != "" && !validSegmentHash(m.TransformPlanHash)) || m.SourceLogicalBytes < 0 || m.SourceRows < 0 || m.TransformedRows < 0 || m.TransformedBytes < 0 || !SafePathPart(m.ParquetFile) || filepath.Ext(m.ParquetFile) != ".parquet" || m.ParquetBytes < 0 || !validSegmentHash(m.ParquetSHA256) {
		return ErrTransportSegmentManifestInvalid
	}
	if requireContentHash {
		content, err := m.contentHash()
		if err != nil || !validSegmentHash(m.ContentSHA256) || content != m.ContentSHA256 {
			return ErrTransportSegmentManifestInvalid
		}
	}
	return nil
}

func (m SegmentManifestV1) contentHash() (string, error) {
	clone := m
	clone.ContentSHA256 = ""
	encoded, err := json.Marshal(clone)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func validSegmentHash(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func syncTransportSegmentDirectoryChain(dir, boundary string) error {
	current, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	stop, err := filepath.Abs(boundary)
	if err != nil {
		return err
	}
	for {
		handle, err := os.Open(current)
		if err != nil {
			return err
		}
		syncErr := handle.Sync()
		closeErr := handle.Close()
		if syncErr != nil {
			return syncErr
		}
		if closeErr != nil {
			return closeErr
		}
		if current == stop {
			return nil
		}
		parent := filepath.Dir(current)
		if parent == current || !strings.HasPrefix(current, stop+string(filepath.Separator)) {
			return ErrTransportSegmentManifestInvalid
		}
		current = parent
	}
}

func (m SegmentManifestV1) String() string {
	return fmt.Sprintf("segment-v%d:%s", m.Version, m.ContentSHA256)
}
