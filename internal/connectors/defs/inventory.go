package defs

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

const githubGraphQLSourceLockPath = "github/sources/github-operation-source-lock.json"

// EmbeddedInventory is a deterministic, attributed accounting of the files
// compiled into FS. It is intentionally derived from the production filesystem
// rather than the repository tree: build-time-only artifacts must not be able
// to appear in this report.
type EmbeddedInventory struct {
	Files      []EmbeddedInventoryFile
	Classes    []EmbeddedInventoryClass
	TotalBytes int64
}

// EmbeddedInventoryFile attributes one shipped definition artifact.
type EmbeddedInventoryFile struct {
	Path  string
	Class string
	Bytes int64
}

// EmbeddedInventoryClass aggregates the shipped files in one artifact class.
type EmbeddedInventoryClass struct {
	Class string
	Files int
	Bytes int64
}

// Inventory returns the complete, sorted production embed inventory. It also
// enforces the negative shipping contract: conformance API surfaces, fixtures,
// and all source locks other than the explicit GitHub certification exception
// are build-time repository artifacts, never runtime bundle content.
func Inventory() (EmbeddedInventory, error) {
	var report EmbeddedInventory
	classes := make(map[string]*EmbeddedInventoryClass)

	err := fs.WalkDir(FS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		class, err := embeddedArtifactClass(path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("embed inventory %q: %w", path, err)
		}

		file := EmbeddedInventoryFile{Path: path, Class: class, Bytes: info.Size()}
		report.Files = append(report.Files, file)
		report.TotalBytes += file.Bytes

		group := classes[class]
		if group == nil {
			group = &EmbeddedInventoryClass{Class: class}
			classes[class] = group
		}
		group.Files++
		group.Bytes += file.Bytes
		return nil
	})
	if err != nil {
		return EmbeddedInventory{}, err
	}

	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].Path < report.Files[j].Path
	})
	for _, group := range classes {
		report.Classes = append(report.Classes, *group)
	}
	sort.Slice(report.Classes, func(i, j int) bool {
		return report.Classes[i].Class < report.Classes[j].Class
	})
	return report, nil
}

func embeddedArtifactClass(path string) (string, error) {
	switch {
	case strings.HasSuffix(path, "/api_surface.json"):
		return "", fmt.Errorf("production embed contains conformance artifact %q", path)
	case strings.Contains("/"+path+"/", "/fixtures/"):
		return "", fmt.Errorf("production embed contains fixture %q", path)
	case strings.Contains("/"+path+"/", "/sources/"):
		if path != githubGraphQLSourceLockPath {
			return "", fmt.Errorf("production embed contains non-exempt source lock %q", path)
		}
		return "certification_source_lock_exception", nil
	case path == "operation_endpoint_ledger.json":
		return "runtime_endpoint_ledger", nil
	case path == "declaration_admission_sources.json":
		return "runtime_declaration_target_ledger", nil
	case path == "circleci/composite_provider_path_identity.json":
		return "composite_provider_path_identity", nil
	case strings.HasSuffix(path, "/schemas.json") || strings.Contains(path, "/schemas/"):
		return "schema", nil
	}

	switch path[strings.LastIndex(path, "/")+1:] {
	case "metadata.json":
		return "metadata", nil
	case "changefeed.json":
		return "changefeed", nil
	case "polling_watermark.json":
		return "polling_watermark", nil
	case "sync_transport.json":
		return "sync_transport", nil
	case "spec.json":
		return "spec", nil
	case "streams.json":
		return "streams", nil
	case "writes.json":
		return "writes", nil
	case "docs.md":
		return "docs", nil
	case "operations.json":
		return "operations", nil
	case "cli_surface.json":
		return "cli_surface", nil
	case "certification.json":
		return "certification", nil
	case "rate_limits.json":
		return "rate_limits", nil
	case "database.json":
		return "database", nil
	default:
		return "", fmt.Errorf("production embed contains unclassified artifact %q", path)
	}
}
