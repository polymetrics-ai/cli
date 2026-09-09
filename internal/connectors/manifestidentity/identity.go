// Package manifestidentity computes the immutable identity of one rendered
// execution generation without parsing connector semantics.
package manifestidentity

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
)

// EmbeddedGeneration identifies the immutable execution generation compiled
// into the current binary. Transactional generation publication may select a
// different declared generation without changing the digest algorithm.
const EmbeddedGeneration = "embedded-v1"

// Identity binds one connector name to its exact closed execution file set.
type Identity struct {
	Connector  string
	Generation string
	Digest     string
	Bytes      int
}

// ForFS computes the execution identity for connector below fsys. It includes
// only runtime execution JSON and schemas; authoring evidence and docs never
// influence the result.
func ForFS(fsys fs.FS, connector, generation string) (Identity, error) {
	if strings.TrimSpace(connector) == "" || strings.TrimSpace(generation) == "" {
		return Identity{}, fmt.Errorf("execution identity requires connector and generation")
	}
	sub, err := fs.Sub(fsys, connector)
	if err != nil {
		return Identity{}, fmt.Errorf("execution identity %q: %w", connector, err)
	}
	files, err := filesFor(sub)
	if err != nil {
		return Identity{}, fmt.Errorf("execution identity %q: %w", connector, err)
	}
	hash := sha256.New()
	bytes := 0
	for _, name := range files {
		if _, err := io.WriteString(hash, name); err != nil {
			return Identity{}, err
		}
		if _, err := hash.Write([]byte{0}); err != nil {
			return Identity{}, err
		}
		data, err := fs.ReadFile(sub, name)
		if err != nil {
			return Identity{}, fmt.Errorf("read %s: %w", name, err)
		}
		bytes += len(data)
		if _, err := hash.Write(data); err != nil {
			return Identity{}, err
		}
		if _, err := hash.Write([]byte{0}); err != nil {
			return Identity{}, err
		}
	}
	return Identity{Connector: connector, Generation: generation, Digest: fmt.Sprintf("sha256:%x", hash.Sum(nil)), Bytes: bytes}, nil
}

func filesFor(fsys fs.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !entry.Type().IsRegular() || !IsExecutionJSONFile(name) {
			return nil
		}
		files = append(files, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// IsExecutionJSONFile reports whether a bundle-relative path contributes to
// the immutable runtime generation identity.
func IsExecutionJSONFile(name string) bool {
	switch name {
	case "metadata.json", "changefeed.json", "polling_watermark.json", "sync_transport.json", "spec.json", "streams.json", "writes.json", "operations.json", "cli_surface.json", "rate_limits.json", "database.json":
		return true
	default:
		return strings.HasPrefix(name, "schemas/") && strings.HasSuffix(name, ".json")
	}
}
