# T/B-11 — cmd/connectorgen (validate | gen | new)

Phase: wave0-engine-harness · Wave: D (task label uses `wave:C` naming collision in
dispatch note, actual PLAN.md section is `Wave D`) · Executor: gsd-loop-backend (sonnet)

## Scope

- `cmd/connectorgen/main.go` — CLI dispatch (`validate`, `gen`, `new`, usage/help)
- `cmd/connectorgen/validate.go` — `Finding`, `Report`, rule constants, `validateDir` and all
  semantic checks (name regex, interpolation resolution, schema refs, PK/cursor existence, write
  path_fields, api_surface rules 1-4, docs.md headings, fixture secret scan)
- `cmd/connectorgen/gen.go` — deterministic hookset/nativeset regeneration
- `cmd/connectorgen/new.go` — bundle scaffolding
- `cmd/connectorgen/main_test.go` — table-driven test suite (RED first)
- `cmd/connectorgen/testdata/valid/goodconn/**` — control bundle (passes validate with 0 findings)
- `cmd/connectorgen/testdata/invalid/**` — 14 seeded-invalid bundles, one per defect class
- `internal/connectors/native/nativeset/nativeset_gen.go` (+ created `native/` dir — did not
  exist before this task; native/ had zero packages, matching the task brief)
- `internal/connectors/hooks/hookset/hookset_gen.go` — regenerated (content equivalent: empty
  import list, `package hookset`, generated-by header; wording differs from the Wave B placeholder
  but the wiring contract is identical)

## RED evidence

Command: `go vet ./cmd/connectorgen` immediately after authoring `main_test.go` (before any
production file existed):

```
# polymetrics.ai/cmd/connectorgen
# [polymetrics.ai/cmd/connectorgen]
vet: cmd/connectorgen/main_test.go:19:17: undefined: validateDir
```

Confirmed RED: the test file referenced `validateDir`, `Finding`, `Report`, rule constants
(`ruleMissingFile` etc.), `genHookset`, `genNativeset`, `runGenAt`, `scaffoldNew`, `runNewAt`,
`run` — none of which existed yet. Compile failure is the correct RED signal for a CLI package
with no prior implementation.

## Implementation notes

- `validateDir(fsys fs.FS) (Report, error)` mirrors `engine.LoadAll`'s contract exactly: fsys root
  is the PARENT of bundle directories (a directory per connector), not a bundle directory itself.
  This matters for both the CLI (`validate [dir]`, default `internal/connectors/defs`) and the test
  harness — a bundle's own `schemas/`/`fixtures/` subdirectories must never be mistaken for sibling
  bundles. Added an `onlyDirFS` test helper wrapper to isolate one seeded-invalid bundle per subtest
  from its 13 siblings under `testdata/invalid/`.
- `validateBundleDir` never returns a bare error: any `engine.Load` structural/meta-schema failure
  is translated into a `Finding` via `loadErrorFinding` (classifying the loader's error text into
  the most specific rule it names — `missing_file`, `schema_ref_missing`, `name_regex`, or a
  `meta_schema` fallback with the file guessed from message content). This keeps `--json` always
  machine-readable and lets one malformed bundle not abort validation of its siblings.
- Composes already-committed engine primitives per the task brief: `engine.Load`/`LoadAll`
  (structural/meta-schema validation), `engine.ResolveCheck` (template resolution against
  `spec.json` properties), `Schema.Properties()`/`PrimaryKeys()`/`CursorFieldName()` (PK/cursor
  existence), `engine.CompileSchema` (write `record_schema` property set for the path_fields
  subset check). Did NOT touch `engine/read.go` or `engine/write.go` (parallel agent's files) —
  confirmed by `grep` before and after, and by watching a transient compile error in those files
  self-resolve mid-session (their WIP, not mine).
- api_surface rules 1-4 (design §E.1) implemented in `checkAPISurface`: exactly-one-of
  covered_by/excluded (`surface_coverage`), bidirectional resolution — covered_by references a
  real stream/write (`surface_unknown_target`) AND every declared stream/write appears in the
  surface (`surface_incomplete`) — closed exclusion vocabulary (`surface_category`, defense in
  depth; the loader's own meta-schema enum already catches an unknown category as a `meta_schema`
  loader-level finding, exercised by the `surface-unknown-category` seeded case), and fail-first-run
  (`surface_fail_first_run`: `capabilities.write`/`read` == false only legal with zero non-excluded
  mutation/GET endpoints).
- `gen` renders both wiring files from one shared `renderWiringFile` helper: fixed header, sorted
  package names, no timestamps — verified byte-stable across 3 consecutive runs
  (`TestGen_ByteStableOnRerun` + manual `diff` across 3 real `go run ./cmd/connectorgen gen`
  invocations against the actual repo hooks/native trees).
- `new <name>` scaffold templates a bundle that passes `validate` outright (one read-only stream,
  schema with both `x-primary-key` and `x-cursor-field`, api_surface fully covering the one stream,
  docs.md with all 5 fixed headings) and rejects both invalid names (regex) and pre-existing
  directories without overwriting.
- Ran `golangci-lint run` (locally installed, v2.12.2 per DECISIONS.md #1) against
  `cmd/connectorgen`, `internal/connectors/native`, `internal/connectors/hooks` even though
  `.golangci.yml` (B-18) is not committed yet; fixed the 9 findings it surfaced (errcheck on
  Fprintln/Fprintf to stdout/stderr via named `logln`/`logf` helpers, one staticcheck SA6003
  `[]rune` range simplification) since they were cheap and load-bearing for the eventual lint gate.
  Final run: `0 issues`.

## Seeded-invalid corpus (defect classes)

14 cases under `cmd/connectorgen/testdata/invalid/`, covering 14 distinct rule IDs (EVAL-PLAN.md §3
requires >=10 seeded / >=8 distinct):

| Dir | Rule | Notes |
|---|---|---|
| `missing-metadata-file` | `missing_file` | metadata.json absent |
| `bad-spec-schema` | `meta_schema` | spec.json has an unknown keyword (draft-07 subset compile failure) |
| `unresolvable-interpolation` | `interpolation_unresolved` | stream path references `{{ config.nope }}` |
| `missing-schema-ref` | `schema_ref_missing` | stream `schema` points at a nonexistent file |
| `pk-not-in-schema` | `primary_key_missing` | `x-primary-key` names a field absent from schema properties |
| `cursor-not-in-schema` | `cursor_field_missing` | `incremental.cursor_field` absent from schema properties |
| `write-path-fields-not-in-schema` | `write_path_fields` | write action `path_fields` not in `record_schema` properties |
| `surface-both-covered-and-excluded` | `surface_coverage` | one endpoint has both `covered_by` and `excluded` |
| `surface-missing-stream` | `surface_incomplete` | declared stream absent from `api_surface.json` |
| `Source-GitHub` | `name_regex` | connector name violates `^[a-z0-9][a-z0-9-]*$` |
| `secret-literal-in-fixture` | `secret_literal` | planted Bearer/`sk_live_`-shaped token in a fixture |
| `docs-missing-heading` | `docs_heading` | docs.md missing the `Write actions & risks` heading |
| `surface-unknown-category` | `meta_schema` | `excluded.category` outside the closed vocabulary (caught by the loader's meta-schema enum before connectorgen's own semantic pass runs) |
| `write-false-with-mutation-endpoint` | `surface_fail_first_run` | `capabilities.write=false` with a non-excluded PATCH endpoint |

Control case: `cmd/connectorgen/testdata/valid/goodconn/` — one stream (`widgets`, incremental,
PK+cursor both present in schema), one write action (`update_widget`), a fully-covering
`api_surface.json`, all 5 docs.md headings, and a clean fixture. `validateDir` on this bundle
returns 0 findings (`TestValidate_AcceptsGoodBundle`).

## GREEN evidence

`go test ./cmd/connectorgen -v` — 23 top-level tests (including the 14-case table-driven
`TestValidate_RejectsSeededInvalidBundles` subtests), all PASS:

```
ok  	polymetrics.ai/cmd/connectorgen	0.374s
```

## Self-verify (all commands from the dispatch, run at the end of the session)

```
$ go build ./...                                            # clean, no output
$ go vet ./cmd/connectorgen                                  # clean, no output
$ go test ./cmd/connectorgen -v 2>&1 | tail -25              # PASS, ok
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 0 connector(s) checked, 0 findings    # exit 0 (defs/ ships no bundles until Wave F)
$ go run ./cmd/connectorgen gen   (x3 in a row)
$ git diff --stat internal/connectors/hooks internal/connectors/native
 internal/connectors/hooks/hookset/hookset_gen.go | 5 +----     # one-time diff vs the Wave B placeholder text; 2nd/3rd runs produce ZERO further diff (byte-stable, confirmed via `diff` on saved copies)
 internal/connectors/native/ is untracked (new dir + nativeset_gen.go; native/ had no packages before this task)
$ gofmt -l cmd/connectorgen internal/connectors/native internal/connectors/hooks
                                                              # empty output = clean
$ golangci-lint run ./cmd/connectorgen/... ./internal/connectors/native/... ./internal/connectors/hooks/...
0 issues.
```

## Notes / deviations

- `internal/connectors/native/nativeset/` did not exist prior to this task (native/ had zero
  packages). Created it per the explicit task instruction ("native/ has nothing (create nativeset
  dir with empty-import file)"). This is the one directory-creation outside the initially-listed
  file paths, but it is exactly the file the task brief names as in-scope
  (`internal/connectors/native/nativeset/nativeset_gen.go`).
- Regenerating `hooks/hookset/hookset_gen.go` changed its comment wording from the Wave B
  placeholder (which had wave0-specific prose about stripe/searxng/postgres needing no hooks) to
  connectorgen's generic, deterministic template comment. The generated CONTRACT is unchanged
  (empty import list, `package hookset`, "Code generated" header) — this is what "regen only —
  content should stay equivalent" in the dispatch scope means; a `gen` tool that could reproduce
  arbitrary hand-written prose verbatim would not be a generator.
- Did not touch `engine/read.go`/`engine/write.go` (confirmed via grep and by observing a transient,
  unrelated compile error in those files self-resolve mid-session — that was the parallel agent's
  own WIP, not a dependency of connectorgen).
- No new go.mod dependencies. No engine API gaps encountered (all needed engine surface —
  `Load`/`LoadAll`, `ResolveCheck`, `Schema.Properties/PrimaryKeys/CursorFieldName`, `CompileSchema`,
  bundle field structs — was already committed and matched API-CONTRACT.md exactly).
- `.golangci.yml` is not committed yet (that's B-18, a separate Wave G task); ran golangci-lint
  locally anyway against only the paths this task owns and fixed all findings preemptively.
