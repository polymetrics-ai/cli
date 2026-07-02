// Command inventorygen scans internal/connectors/<name>/ directories and emits
// docs/migration/inventory.json — a deterministic per-connector inventory used by
// the migration orchestration plan to size and route Pass A/B migration agents.
//
// loc counting convention (DECISIONS.md #3): loc for a connector directory is the
// sum of non-blank, non-comment-only lines across ALL top-level *.go files,
// INCLUDING _test.go files. This matches the orchestration-plan.md calibration
// numbers (558 connector packages, ~309k lines of connector Go; size buckets
// S<300/M300-699/L700-899/XL>=900), which were computed over the full .go set,
// not just non-test source.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// --- bucketization -----------------------------------------------------------

func TestBucketForLOC(t *testing.T) {
	cases := []struct {
		name string
		loc  int
		want string
	}{
		{"zero", 0, "S"},
		{"small-mid", 150, "S"},
		{"s-upper-boundary", 299, "S"},
		{"m-lower-boundary", 300, "M"},
		{"m-mid", 500, "M"},
		{"m-upper-boundary", 699, "M"},
		{"l-lower-boundary", 700, "L"},
		{"l-mid", 800, "L"},
		{"l-upper-boundary", 899, "L"},
		{"xl-lower-boundary", 900, "XL"},
		{"xl-large", 3865, "XL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := bucketForLOC(tc.loc)
			if got != tc.want {
				t.Fatalf("bucketForLOC(%d) = %q, want %q", tc.loc, got, tc.want)
			}
		})
	}
}

// TestBucketForLOC_Monotonic guards against a bucket function that is merely
// piecewise-correct on the table above but non-monotonic elsewhere (e.g. an
// off-by-one that would let a higher loc count map to a smaller bucket).
func TestBucketForLOC_Monotonic(t *testing.T) {
	rank := map[string]int{"S": 0, "M": 1, "L": 2, "XL": 3}
	prev := bucketForLOC(0)
	for loc := 1; loc <= 4000; loc++ {
		cur := bucketForLOC(loc)
		if rank[cur] < rank[prev] {
			t.Fatalf("bucketForLOC regressed at loc=%d: prev=%q cur=%q", loc, prev, cur)
		}
		prev = cur
	}
}

// --- loc counting --------------------------------------------------------------

func TestCountGoLOC_IncludesTestFiles(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "connector.go"), "package foo\n\nfunc A() {}\n")
	mustWriteFile(t, filepath.Join(dir, "connector_test.go"), "package foo\n\nfunc TestA(t *testing.T) {}\n")

	got, err := countGoLOC(dir)
	if err != nil {
		t.Fatalf("countGoLOC: %v", err)
	}
	// Both files must contribute; a loc counter that excludes _test.go would
	// undercount here (see the DECISIONS.md #3 comment on countGoLOC).
	nonTestOnly, err := countGoLOCExcluding(dir, "connector_test.go")
	if err != nil {
		t.Fatalf("countGoLOCExcluding: %v", err)
	}
	if got <= nonTestOnly {
		t.Fatalf("countGoLOC(%d) did not include _test.go contribution (non-test-only=%d)", got, nonTestOnly)
	}
}

func TestCountGoLOC_BlankAndCommentLinesExcluded(t *testing.T) {
	dir := t.TempDir()
	src := "package foo\n\n// a comment line\nfunc A() {\n\treturn\n}\n\n"
	mustWriteFile(t, filepath.Join(dir, "a.go"), src)

	got, err := countGoLOC(dir)
	if err != nil {
		t.Fatalf("countGoLOC: %v", err)
	}
	// Non-blank, non-comment-only lines: "package foo", "func A() {", "return", "}"
	want := 4
	if got != want {
		t.Fatalf("countGoLOC = %d, want %d", got, want)
	}
}

func TestCountGoLOC_IgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "a.go"), "package foo\nfunc A() {}\n")
	subdir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	mustWriteFile(t, filepath.Join(subdir, "b.go"), "package sub\nfunc B() {}\nfunc C() {}\nfunc D() {}\n")

	got, err := countGoLOC(dir)
	if err != nil {
		t.Fatalf("countGoLOC: %v", err)
	}
	if got != 2 {
		t.Fatalf("countGoLOC = %d, want 2 (subdir files must not be counted)", got)
	}
}

// --- catalog join --------------------------------------------------------------

func TestCatalogSlugsForName_PrefersPMConnectorName(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-stripe", PMConnectorName: "stripe", DocumentationURL: "https://stripe.com/docs"},
		{Slug: "destination-stripe", PMConnectorName: "stripe", DocumentationURL: "https://stripe.com/docs2"},
		{Slug: "source-unrelated", PMConnectorName: "other", DocumentationURL: "https://example.com"},
	}
	got := catalogSlugsForName("stripe", entries)
	want := []string{"destination-stripe", "source-stripe"}
	assertStringSliceEqual(t, got, want)
}

func TestCatalogSlugsForName_FallsBackToBareName(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-github", DocumentationURL: "https://github.com/docs"},
		{Slug: "destination-github", DocumentationURL: "https://github.com/docs2"},
	}
	got := catalogSlugsForName("github", entries)
	want := []string{"destination-github", "source-github"}
	assertStringSliceEqual(t, got, want)
}

func TestCatalogSlugsForName_NoMatch(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-github", DocumentationURL: "https://github.com/docs"},
	}
	got := catalogSlugsForName("totally-unknown-connector", entries)
	if len(got) != 0 {
		t.Fatalf("catalogSlugsForName = %v, want empty", got)
	}
}

func TestDocumentationURLForName_PrefersPMConnectorNameMatch(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-stripe", PMConnectorName: "stripe", DocumentationURL: "https://stripe.com/docs"},
		{Slug: "source-unrelated", PMConnectorName: "other", DocumentationURL: "https://example.com"},
	}
	got := documentationURLForName("stripe", entries)
	if got != "https://stripe.com/docs" {
		t.Fatalf("documentationURLForName = %q, want %q", got, "https://stripe.com/docs")
	}
}

func TestDocumentationURLForName_FallsBackToBareNameMatch(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-github", DocumentationURL: "https://github.com/docs"},
	}
	got := documentationURLForName("github", entries)
	if got != "https://github.com/docs" {
		t.Fatalf("documentationURLForName = %q, want %q", got, "https://github.com/docs")
	}
}

func TestDocumentationURLForName_NoMatchIsEmpty(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-github", DocumentationURL: "https://github.com/docs"},
	}
	got := documentationURLForName("nope", entries)
	if got != "" {
		t.Fatalf("documentationURLForName = %q, want empty", got)
	}
}

func TestRuntimeKindForName_PrefersPMConnectorNameMatch(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-stripe", PMConnectorName: "stripe", RuntimeKind: connectors.RuntimeDeclarativeHTTPGo},
	}
	got := runtimeKindForName("stripe", entries)
	if got != "declarative_http_go" {
		t.Fatalf("runtimeKindForName = %q, want %q", got, "declarative_http_go")
	}
}

func TestRuntimeKindForName_NoMatchIsEmpty(t *testing.T) {
	entries := []connectors.ConnectorDefinition{
		{Slug: "source-stripe", PMConnectorName: "stripe", RuntimeKind: connectors.RuntimeDeclarativeHTTPGo},
	}
	got := runtimeKindForName("nope", entries)
	if got != "" {
		t.Fatalf("runtimeKindForName = %q, want empty", got)
	}
}

// --- connector directory discovery --------------------------------------------

func TestConnectorDirs_SkipsNonConnectorDirs(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"stripe", "github", "connsdk", "httpsource", "registryset", "_quarantine"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		mustWriteFile(t, filepath.Join(dir, "x.go"), "package x\nfunc X(){}\n")
	}
	got, err := connectorDirs(root)
	if err != nil {
		t.Fatalf("connectorDirs: %v", err)
	}
	want := []string{"github", "stripe"}
	assertStringSliceEqual(t, got, want)
}

// TestConnectorDirs_SkipsEngineHarnessInfraDirs guards against the migration
// inventory accidentally treating wave0 engine-harness infrastructure packages
// (shared SDK/tooling, not per-system connectors) as migration targets. This
// set matches PLAN.md B-16's registrygen skip-map list exactly, since both
// tools must agree on what counts as a "connector directory".
func TestConnectorDirs_SkipsEngineHarnessInfraDirs(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"defs", "engine", "hooks", "native", "conformance", "certify"} {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		mustWriteFile(t, filepath.Join(dir, "x.go"), "package x\nfunc X(){}\n")
	}
	mustWriteFile(t, filepath.Join(root, "stripe", "stripe.go"), samplePackage("stripe", 5))

	got, err := connectorDirs(root)
	if err != nil {
		t.Fatalf("connectorDirs: %v", err)
	}
	want := []string{"stripe"}
	assertStringSliceEqual(t, got, want)
}

func TestConnectorDirs_SkipsDirsWithoutGoFiles(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "real", "a.go"), "package real\nfunc A(){}\n")
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatalf("mkdir empty: %v", err)
	}
	got, err := connectorDirs(root)
	if err != nil {
		t.Fatalf("connectorDirs: %v", err)
	}
	want := []string{"real"}
	assertStringSliceEqual(t, got, want)
}

// --- JSON output shape + determinism -------------------------------------------

func TestBuildInventory_ShapeAndSort(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "stripe", "stripe.go"), samplePackage("stripe", 10))
	mustWriteFile(t, filepath.Join(root, "github", "github.go"), samplePackage("github", 5))

	catalog := []connectors.ConnectorDefinition{
		{Slug: "source-stripe", PMConnectorName: "stripe", DocumentationURL: "https://stripe.com/docs", RuntimeKind: connectors.RuntimeDeclarativeHTTPGo},
		{Slug: "source-github", PMConnectorName: "github", DocumentationURL: "https://github.com/docs", RuntimeKind: connectors.RuntimeNativeGo},
	}
	streamCounts := map[string]int{"stripe": 5, "github": 12}

	inv, err := buildInventory(root, catalog, streamCounts)
	if err != nil {
		t.Fatalf("buildInventory: %v", err)
	}
	if len(inv.Connectors) != 2 {
		t.Fatalf("len(connectors) = %d, want 2", len(inv.Connectors))
	}
	// sorted by name
	if inv.Connectors[0].Name != "github" || inv.Connectors[1].Name != "stripe" {
		t.Fatalf("connectors not sorted by name: %+v", inv.Connectors)
	}
	gh := inv.Connectors[0]
	if gh.Path != filepath.Join("internal/connectors", "github") {
		t.Fatalf("gh.Path = %q", gh.Path)
	}
	if gh.LOC <= 0 {
		t.Fatalf("gh.LOC = %d, want > 0", gh.LOC)
	}
	if gh.Bucket == "" {
		t.Fatalf("gh.Bucket is empty")
	}
	if gh.RuntimeKind != "native_go" {
		t.Fatalf("gh.RuntimeKind = %q, want native_go", gh.RuntimeKind)
	}
	if gh.DocumentationURL != "https://github.com/docs" {
		t.Fatalf("gh.DocumentationURL = %q", gh.DocumentationURL)
	}
	if len(gh.CatalogSlugs) != 1 || gh.CatalogSlugs[0] != "source-github" {
		t.Fatalf("gh.CatalogSlugs = %v", gh.CatalogSlugs)
	}
	if gh.StreamCount != 12 {
		t.Fatalf("gh.StreamCount = %d, want 12", gh.StreamCount)
	}
	if inv.GeneratedNote == "" {
		t.Fatalf("GeneratedNote is empty")
	}
}

func TestBuildInventory_EveryEntryHasBucket(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "a", "a.go"), samplePackage("a", 1))
	mustWriteFile(t, filepath.Join(root, "b", "b.go"), samplePackage("b", 400))
	mustWriteFile(t, filepath.Join(root, "c", "c.go"), samplePackage("c", 750))
	mustWriteFile(t, filepath.Join(root, "d", "d.go"), samplePackage("d", 950))

	inv, err := buildInventory(root, nil, nil)
	if err != nil {
		t.Fatalf("buildInventory: %v", err)
	}
	for _, c := range inv.Connectors {
		if c.Bucket != "S" && c.Bucket != "M" && c.Bucket != "L" && c.Bucket != "XL" {
			t.Fatalf("connector %q has invalid bucket %q", c.Name, c.Bucket)
		}
	}
}

// TestBuildInventory_Deterministic guards the no-timestamp / diff-stability
// requirement: regenerating from identical inputs must byte-for-byte match.
func TestBuildInventory_Deterministic(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "stripe", "stripe.go"), samplePackage("stripe", 10))
	catalog := []connectors.ConnectorDefinition{
		{Slug: "source-stripe", PMConnectorName: "stripe", DocumentationURL: "https://stripe.com/docs", RuntimeKind: connectors.RuntimeDeclarativeHTTPGo},
	}
	streamCounts := map[string]int{"stripe": 5}

	first, err := buildInventory(root, catalog, streamCounts)
	if err != nil {
		t.Fatalf("buildInventory (first): %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	second, err := buildInventory(root, catalog, streamCounts)
	if err != nil {
		t.Fatalf("buildInventory (second): %v", err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("buildInventory not deterministic:\nfirst:  %s\nsecond: %s", firstJSON, secondJSON)
	}
}

func TestInventory_JSONShape(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "stripe", "stripe.go"), samplePackage("stripe", 10))
	catalog := []connectors.ConnectorDefinition{
		{Slug: "source-stripe", PMConnectorName: "stripe", DocumentationURL: "https://stripe.com/docs", RuntimeKind: connectors.RuntimeDeclarativeHTTPGo},
	}
	inv, err := buildInventory(root, catalog, map[string]int{"stripe": 5})
	if err != nil {
		t.Fatalf("buildInventory: %v", err)
	}
	raw, err := json.Marshal(inv)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := generic["generated_note"]; !ok {
		t.Fatalf("top-level JSON missing generated_note key: %s", raw)
	}
	if _, ok := generic["generated_at"]; ok {
		t.Fatalf("top-level JSON must NOT contain a wall-clock generated_at key (diff-stability): %s", raw)
	}
	connectorsRaw, ok := generic["connectors"].([]any)
	if !ok || len(connectorsRaw) == 0 {
		t.Fatalf("top-level JSON missing non-empty connectors array: %s", raw)
	}
	entry, ok := connectorsRaw[0].(map[string]any)
	if !ok {
		t.Fatalf("connectors[0] is not an object: %v", connectorsRaw[0])
	}
	for _, key := range []string{"name", "path", "loc", "bucket", "runtime_kind", "catalog_slugs", "documentation_url", "stream_count"} {
		if _, ok := entry[key]; !ok {
			t.Fatalf("connectors[0] missing key %q: %s", key, raw)
		}
	}
}

// --- test helpers ---------------------------------------------------------------

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func samplePackage(name string, lines int) string {
	out := "package " + name + "\n\n"
	for i := 0; i < lines; i++ {
		out += "var _ = 0\n"
	}
	return out
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sortedWant := append([]string(nil), want...)
	sort.Strings(sortedWant)
	if len(got) != len(sortedWant) {
		t.Fatalf("got %v, want %v", got, sortedWant)
	}
	for i := range got {
		if got[i] != sortedWant[i] {
			t.Fatalf("got %v, want %v", got, sortedWant)
		}
	}
}
