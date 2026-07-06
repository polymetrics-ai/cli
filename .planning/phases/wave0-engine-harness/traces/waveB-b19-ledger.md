# T-19 / B-19 TDD ledger — cmd/inventorygen

Executed by: gsd-loop-backend (sonnet), task T/B-19 (floated: depends only on the LEGACY
registry/catalog, disjoint from `internal/connectors/engine`).

## T-19 (bucketization + loc counting + catalog join + JSON shape)

Status: red-confirmed
Timestamp: 2026-07-02 (session)

Command: `go test ./cmd/inventorygen/... -v`

Output:
```
# polymetrics.ai/cmd/inventorygen [polymetrics.ai/cmd/inventorygen.test]
cmd/inventorygen/main_test.go:45:11: undefined: bucketForLOC
cmd/inventorygen/main_test.go:58:10: undefined: bucketForLOC
cmd/inventorygen/main_test.go:60:10: undefined: bucketForLOC
cmd/inventorygen/main_test.go:75:14: undefined: countGoLOC
cmd/inventorygen/main_test.go:81:22: undefined: countGoLOCExcluding
cmd/inventorygen/main_test.go:95:14: undefined: countGoLOC
cmd/inventorygen/main_test.go:115:14: undefined: countGoLOC
cmd/inventorygen/main_test.go:132:9: undefined: catalogSlugsForName
cmd/inventorygen/main_test.go:142:9: undefined: catalogSlugsForName
cmd/inventorygen/main_test.go:151:9: undefined: catalogSlugsForName
cmd/inventorygen/main_test.go:151:9: too many errors
FAIL	polymetrics.ai/cmd/inventorygen [build failed]
FAIL
```

Test file (`cmd/inventorygen/main_test.go`) authored per PLAN.md T-19/B-19 + DECISIONS.md #3 +
orchestration-plan.md §Calibration/§Artifacts, covering:

- `TestBucketForLOC`: table-driven boundary cases at 299/300/699/700/899/900 plus 0 and the max
  observed (3865, github) — asserts S<300, M 300-699, L 700-899, XL>=900.
- `TestBucketForLOC_Monotonic`: sweeps loc 0..4000 to guard against an off-by-one that would make
  a higher loc map to a smaller bucket.
- `TestCountGoLOC_IncludesTestFiles` / `_BlankAndCommentLinesExcluded` / `_IgnoresSubdirectories`:
  documents and locks in the DECISIONS.md #3 convention — loc counts ALL top-level `.go` files
  including `_test.go`, non-blank/non-full-line-comment lines only, no recursion into
  subdirectories (matters for connectors with `fixtures/` or similar nested dirs).
- `TestCatalogSlugsForName_*` / `TestDocumentationURLForName_*` / `TestRuntimeKindForName_*`:
  catalog join keyed first by `PMConnectorName` (exact match against connector dir/registry name),
  falling back to `connectors.BareName(slug)` when no entry sets `pm_connector_name` (true for
  ~556 of 558 dirs per calibration — only stripe/github have it set today).
- `TestConnectorDirs_*`: scans a fixture tree, skips `connsdk`/`httpsource`/`registryset`/
  `_quarantine` and any dir with zero `.go` files.
- `TestBuildInventory_ShapeAndSort` / `_EveryEntryHasBucket` / `_Deterministic` /
  `TestInventory_JSONShape`: end-to-end shape assertions — sorted by name, every entry has a
  bucket, output is byte-identical across two builds from identical inputs (no wall-clock
  `generated_at`; uses a static `generated_note` string instead per dispatch-prompt instruction,
  which supersedes DATA-MODEL.md's illustrative `generated_at` field for diff-stability), and the
  marshaled JSON has exactly the keys `name/path/loc/bucket/runtime_kind/catalog_slugs/
  documentation_url/stream_count` per connector plus top-level `generated_note`/`connectors`.

## B-19 (implementation)

Status: green-confirmed
Timestamp: 2026-07-02 (session, immediately after RED)

Command: `go test ./cmd/inventorygen/... -v`

Output (tail):
```
--- PASS: TestBucketForLOC (…)
--- PASS: TestBucketForLOC_Monotonic (…)
--- PASS: TestCountGoLOC_IncludesTestFiles (…)
--- PASS: TestCountGoLOC_BlankAndCommentLinesExcluded (…)
--- PASS: TestCountGoLOC_IgnoresSubdirectories (…)
--- PASS: TestCatalogSlugsForName_PrefersPMConnectorName (…)
--- PASS: TestCatalogSlugsForName_FallsBackToBareName (…)
--- PASS: TestCatalogSlugsForName_NoMatch (…)
--- PASS: TestDocumentationURLForName_PrefersPMConnectorNameMatch (…)
--- PASS: TestDocumentationURLForName_FallsBackToBareNameMatch (…)
--- PASS: TestDocumentationURLForName_NoMatchIsEmpty (…)
--- PASS: TestRuntimeKindForName_PrefersPMConnectorNameMatch (…)
--- PASS: TestRuntimeKindForName_NoMatchIsEmpty (…)
--- PASS: TestConnectorDirs_SkipsNonConnectorDirs (…)
--- PASS: TestConnectorDirs_SkipsDirsWithoutGoFiles (…)
--- PASS: TestBuildInventory_ShapeAndSort (…)
--- PASS: TestBuildInventory_EveryEntryHasBucket (…)
--- PASS: TestBuildInventory_Deterministic (…)
--- PASS: TestInventory_JSONShape (…)
PASS
ok  	polymetrics.ai/cmd/inventorygen	...
```

`go build ./cmd/inventorygen && go vet ./cmd/inventorygen` clean; `gofmt -l cmd/inventorygen` empty.

## Real-run evidence

Command: `go run ./cmd/inventorygen` (writes `docs/migration/inventory.json`), then
`python3 -c "import json;d=json.load(open('docs/migration/inventory.json'));print(len(d['connectors']))"`.

Result: `557` connector entries (assert `>500` passes). Bucket histogram:
`{'M': 388, 'L': 31, 'S': 137, 'XL': 1}`. Every entry has a non-zero `loc` and a non-empty
`bucket`. `documentation_url` empty only for `openrouter` (self-registered pm-native connector with
no catalog row at all) and `searxng` (pm-native via `RegisterNativeLive`, also no catalog row) —
both are correct joins, not bugs.

## Mid-implementation correction (self-caught, before final GREEN)

First real run produced 560 entries and a bucket histogram far off calibration (M 389/L 31/S
138/XL 2, with `engine` at 2720 loc showing up as an "XL connector"). Root cause: the initial
`nonConnectorDirs` skip set only carried cmd/registrygen's CURRENT skip map
(`connsdk/httpsource/registryset/_quarantine`), but this wave0 build has already created sibling
package dirs under `internal/connectors/` for the engine harness itself — `engine`, `hooks`,
`defs`, `certify` (each with real top-level `.go` files) — which are shared infrastructure, not
per-system connectors. PLAN.md's B-16 task explicitly enumerates the FUTURE registrygen skip-map
addition as `defs, engine, hooks, native, conformance, certify` — confirming these are recognized
non-connector dirs. Added the same six names to inventorygen's `nonConnectorDirs` (matching B-16's
list verbatim, including `native`/`conformance` which don't exist as dirs yet but will), added a
regression test (`TestConnectorDirs_SkipsEngineHarnessInfraDirs`), reran RED->GREEN (test passed on
first try since it exercises `connectorDirs` directly), rebuilt, regenerated inventory.json. Final
count 557 / histogram above. This was a self-caught data-quality fix within the same task, not an
escalation — no scope change to the files list.

## stream_count caveat (documented, not a bug)

554 of 557 connectors report `stream_count: 0`. Root cause: `connectors.ManifestOf` only returns a
populated `Streams` list when the target `Connector` implements the optional `ManifestProvider`
interface (`Manifest() Manifest`); today only 3 packages do (`freshdesk`, `github`, `stripe` —
verified via `grep -rln "func.*Manifest() connectors.Manifest" internal/connectors/*/*.go`). All
other legacy connectors only implement the base `Connector` interface (`Catalog()`, which requires
a live network Check and isn't safe/meaningful to call from a static inventory tool). This matches
PLAN.md's literal instruction ("stream_count from `connectors.ManifestOf` on
`registryset.NewStaged()`") — the field is accurately reporting today's static-manifest coverage,
which will grow as Pass A migration lands `engine.Base`-backed bundles with `Manifest()`.

## Blockers

None. `internal/connectors/registryset.NewStaged()` + `connectors.ManifestOf` worked exactly as
the plan assumed (registryset already blank-imports every connector package via the generated
`registry_gen.go`; `ManifestOf(c).Streams` gives the per-connector stream list keyed by the same
name the factory registered under, which equals the directory name). No `SCHEMA_AMBIGUOUS` or other
typed blocker needed.

## Path-guard self-check

`git status --porcelain` after this task touches only: `cmd/inventorygen/main.go`,
`cmd/inventorygen/main_test.go`, `docs/migration/inventory.json`, and this ledger file. A stray
`./inventorygen` binary left in the repo root by an intermediate `go build ./cmd/inventorygen`
invocation (Go names the output after the last import-path component when no `-o` is given) was
removed before finishing, since it is not one of the task's listed files.
