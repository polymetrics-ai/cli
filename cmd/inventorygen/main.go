// Command inventorygen scans internal/connectors/<name>/ directories and writes
// docs/migration/inventory.json — a deterministic per-connector inventory feeding
// the migration orchestration plan (bundle sizing/assignment for wave1+).
//
// Run once at convergence, same posture as cmd/registrygen:
//
//	go run ./cmd/inventorygen
//
// The output is deterministic: no wall-clock timestamp is embedded, so
// regenerating from an unchanged tree produces a byte-identical file (diff-stable
// across reruns, required for reproducible migration planning).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/registryset"
)

const (
	connectorsRel = "internal/connectors"
	outRel        = "docs/migration/inventory.json"

	bucketSmall      = "S"
	bucketMedium     = "M"
	bucketLarge      = "L"
	bucketExtraLarge = "XL"
)

// nonConnectorDirs lists directories under internal/connectors that are not
// per-system connector packages: shared SDK, generated wiring, quarantine, and
// (per PLAN.md B-16's registrygen skip-map, which enumerates the same wave0
// engine-harness infrastructure packages) the engine/connectorgen/conformance/
// certify/native/defs/hooks packages introduced alongside this tool. Migration
// bundle sizing must never target these — they are shared infra, not
// per-system connectors to be migrated.
var nonConnectorDirs = map[string]bool{
	"connsdk":     true,
	"httpsource":  true,
	"registryset": true,
	"_quarantine": true,
	"defs":        true,
	"engine":      true,
	"hooks":       true,
	"native":      true,
	"conformance": true,
	"certify":     true,
}

// connectorInventoryEntry is one row of docs/migration/inventory.json.
type connectorInventoryEntry struct {
	Name             string   `json:"name"`
	Path             string   `json:"path"`
	LOC              int      `json:"loc"`
	Bucket           string   `json:"bucket"`
	RuntimeKind      string   `json:"runtime_kind"`
	CatalogSlugs     []string `json:"catalog_slugs"`
	DocumentationURL string   `json:"documentation_url"`
	StreamCount      int      `json:"stream_count"`
}

// connectorInventory is the top-level docs/migration/inventory.json shape.
// Intentionally has no generated-at timestamp field: regeneration from an
// unchanged source tree must be byte-for-byte stable.
type connectorInventory struct {
	GeneratedNote string                    `json:"generated_note"`
	Connectors    []connectorInventoryEntry `json:"connectors"`
}

func main() {
	root, err := repoRoot()
	if err != nil {
		fail(err)
	}
	catalog := connectors.ConnectorCatalog()
	streamCounts, err := streamCountsByName()
	if err != nil {
		fail(err)
	}
	inv, err := buildInventory(filepath.Join(root, connectorsRel), catalog, streamCounts)
	if err != nil {
		fail(err)
	}
	if err := writeInventory(filepath.Join(root, outRel), inv); err != nil {
		fail(err)
	}
	fmt.Printf("inventorygen: wrote %d connector entries to %s\n", len(inv.Connectors), outRel)
}

// repoRoot finds the module root by walking up to the directory containing go.mod.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from working directory")
		}
		dir = parent
	}
}

// streamCountsByName builds a name -> stream count map from the staged registry
// (registryset.NewStaged), which blank-imports every connector package so all
// self-registered factories are reachable regardless of catalog/live status.
// Keyed by the same name the connector package registered under via
// connectors.RegisterFactory, which is the connector directory name.
func streamCountsByName() (map[string]int, error) {
	registry := registryset.NewStaged()
	counts := make(map[string]int)
	for _, meta := range registry.List() {
		c, ok := registry.Get(meta.Name)
		if !ok {
			continue
		}
		manifest := connectors.ManifestOf(c)
		counts[meta.Name] = len(manifest.Streams)
	}
	return counts, nil
}

// buildInventory scans connectorsRoot for connector directories and joins each
// against the catalog and stream-count data. catalog/streamCounts may be nil,
// producing zero-value joins (used by unit tests that only exercise the
// dir-scan + loc + bucket path).
func buildInventory(connectorsRoot string, catalog []connectors.ConnectorDefinition, streamCounts map[string]int) (connectorInventory, error) {
	names, err := connectorDirs(connectorsRoot)
	if err != nil {
		return connectorInventory{}, err
	}
	entries := make([]connectorInventoryEntry, 0, len(names))
	for _, name := range names {
		loc, err := countGoLOC(filepath.Join(connectorsRoot, name))
		if err != nil {
			return connectorInventory{}, fmt.Errorf("count loc for %s: %w", name, err)
		}
		entries = append(entries, connectorInventoryEntry{
			Name:             name,
			Path:             filepath.ToSlash(filepath.Join(connectorsRel, name)),
			LOC:              loc,
			Bucket:           bucketForLOC(loc),
			RuntimeKind:      runtimeKindForName(name, catalog),
			CatalogSlugs:     catalogSlugsForName(name, catalog),
			DocumentationURL: documentationURLForName(name, catalog),
			StreamCount:      streamCounts[name],
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return connectorInventory{
		GeneratedNote: "Generated by cmd/inventorygen (go run ./cmd/inventorygen). Deterministic: " +
			"no wall-clock timestamp is embedded so reruns over an unchanged tree are byte-identical.",
		Connectors: entries,
	}, nil
}

// connectorDirs returns the sorted list of connector directory names under
// connectorsRoot: any entry that is a directory, is not in nonConnectorDirs, and
// contains at least one top-level *.go file.
func connectorDirs(connectorsRoot string) ([]string, error) {
	dirEntries, err := os.ReadDir(connectorsRoot)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", connectorsRoot, err)
	}
	var names []string
	for _, entry := range dirEntries {
		if !entry.IsDir() || nonConnectorDirs[entry.Name()] {
			continue
		}
		hasGo, err := dirHasGoFiles(filepath.Join(connectorsRoot, entry.Name()))
		if err != nil {
			return nil, err
		}
		if hasGo {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func dirHasGoFiles(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", dir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true, nil
		}
	}
	return false, nil
}

// countGoLOC counts non-blank, non-full-line-comment lines across every
// top-level *.go file directly under dir (no recursion into subdirectories such
// as fixtures/ or testdata/). Per DECISIONS.md #3, this INCLUDES _test.go files
// to match the orchestration-plan.md calibration (558 connectors, ~309k lines).
func countGoLOC(dir string) (int, error) {
	return countGoLOCExcluding(dir)
}

// countGoLOCExcluding is countGoLOC but skips any file whose base name is listed
// in excludeNames. Exists so tests can assert _test.go files change the count
// (see TestCountGoLOC_IncludesTestFiles); production code always calls
// countGoLOC (no exclusions).
func countGoLOCExcluding(dir string, excludeNames ...string) (int, error) {
	excluded := make(map[string]bool, len(excludeNames))
	for _, name := range excludeNames {
		excluded[name] = true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", dir, err)
	}
	total := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || excluded[entry.Name()] {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return 0, fmt.Errorf("read %s: %w", filepath.Join(dir, entry.Name()), err)
		}
		total += countLOCInSource(data)
	}
	return total, nil
}

// countLOCInSource counts non-blank lines that are not entirely a "//" comment.
// This is an intentionally simple heuristic (no block-comment tracking) — good
// enough for size-bucket calibration, not a precise SLOC tool.
func countLOCInSource(src []byte) int {
	count := 0
	for _, line := range strings.Split(string(src), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		count++
	}
	return count
}

// bucketForLOC maps a loc count to a migration size bucket per
// docs/migration/orchestration-plan.md §Calibration: S<300, M 300-699,
// L 700-899, XL>=900.
func bucketForLOC(loc int) string {
	switch {
	case loc < 300:
		return bucketSmall
	case loc < 700:
		return bucketMedium
	case loc < 900:
		return bucketLarge
	default:
		return bucketExtraLarge
	}
}

// catalogSlugsForName returns every catalog slug for the given connector
// directory name, joined first by exact PMConnectorName match, then (if no
// entry sets PMConnectorName for this name) by connectors.BareName(slug).
// Sorted for deterministic output.
func catalogSlugsForName(name string, catalog []connectors.ConnectorDefinition) []string {
	var slugs []string
	for _, def := range catalog {
		if def.PMConnectorName == name {
			slugs = append(slugs, def.Slug)
		}
	}
	if len(slugs) > 0 {
		sort.Strings(slugs)
		return slugs
	}
	for _, def := range catalog {
		if def.PMConnectorName != "" {
			continue
		}
		if connectors.BareName(def.Slug) == name {
			slugs = append(slugs, def.Slug)
		}
	}
	sort.Strings(slugs)
	return slugs
}

// documentationURLForName returns the documentation URL for the first catalog
// match (PMConnectorName exact match preferred, falling back to BareName), or
// empty if there is no catalog entry for this connector.
func documentationURLForName(name string, catalog []connectors.ConnectorDefinition) string {
	if def, ok := catalogEntryForName(name, catalog); ok {
		return def.DocumentationURL
	}
	return ""
}

// runtimeKindForName returns the runtime_kind for the first catalog match, or
// empty if there is no catalog entry for this connector (e.g. pm-native-only
// connectors like searxng that opt in via RegisterNativeLive with no catalog
// row).
func runtimeKindForName(name string, catalog []connectors.ConnectorDefinition) string {
	if def, ok := catalogEntryForName(name, catalog); ok {
		return string(def.RuntimeKind)
	}
	return ""
}

// catalogEntryForName returns the single best catalog entry for name: exact
// PMConnectorName match wins; otherwise the first BareName(slug) match (source
// entries sort before destination entries alphabetically in most unify pairs,
// but the choice is deterministic either way since callers only read fields that
// are stable across a system's source/destination catalog rows).
func catalogEntryForName(name string, catalog []connectors.ConnectorDefinition) (connectors.ConnectorDefinition, bool) {
	for _, def := range catalog {
		if def.PMConnectorName == name {
			return def, true
		}
	}
	var candidates []connectors.ConnectorDefinition
	for _, def := range catalog {
		if def.PMConnectorName != "" {
			continue
		}
		if connectors.BareName(def.Slug) == name {
			candidates = append(candidates, def)
		}
	}
	if len(candidates) == 0 {
		return connectors.ConnectorDefinition{}, false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Slug < candidates[j].Slug })
	return candidates[0], true
}

func writeInventory(path string, inv connectorInventory) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal inventory: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "inventorygen:", err)
	os.Exit(1)
}
