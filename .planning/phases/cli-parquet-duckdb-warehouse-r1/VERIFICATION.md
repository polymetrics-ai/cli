# Verification — Parquet tables, DuckDB engine

Phase: `cli-parquet-duckdb-warehouse-r1`

## Definition of done, item by item

| Required | Status | Evidence |
|---|---|---|
| Tables materialize as Parquet under the #3901 layout | done | `traces/e2e-run.log`: `tables/sample_customers.parquet` beside `owner.json` and `wal/customers.jsonl`; `file(1)` says Apache Parquet |
| DuckDB serves queries over them | done | `pm query run --sql` answers GROUP BY, WHERE + aggregate, CTE + window; writes refused |
| ETL proven working by running the binary | done | `pm etl run` + `make smoke` |
| Reverse ETL proven working by running the binary | done | plan → preview → run, records asserted in the outbox |
| **Proven against the LIVE GitHub connector, read-only** | done | `traces/live-github-proof.md` |
| Table-as-file vs table-as-directory settled with evidence | done | `PLAN.md` Decision 1; `traces/layout-decision-experiment.log` |
| Platform matrix change measured and justified | done | Decision 3 below; CI job added to verify before dropping |
| Existing-warehouse behaviour stated plainly | done | Decision 4 below |
| GSD evidence recorded as part of the work | done | `PLAN.md` written before implementation, `TDD-LEDGER.md`, this file |
| **Release pipeline builds the new matrix** | **blocked — needs a decision** | see "Open item" |

## Live GitHub proof

Run against the live GitHub REST API on the public `golang/example` repository, with a token from
`gh auth token` that was never printed or logged. The token carries `repo` scope, so the exercise
was strictly read-only: every `pm` call against GitHub was a source read, and the reverse-ETL
destination was the **local outbox**, not GitHub. Confirmed afterwards by re-reading the repository —
labels 11 → 11, pull requests 74 → 74, open issues 11 → 11, unchanged. No `untested-because-write-scoped`
gap remains: the round trip was provable end to end without any mutation.

| assertion | result |
|---|---|
| `pull_requests` landed in Parquet | 74 read, 74 loaded, 0 failed |
| `labels` landed in Parquet | 11 read, 11 loaded, 0 failed |
| Field values vs the live API | **IDENTICAL** on all 74 PRs (node_id, number, updated_at, state) and all 11 labels (name, color) |
| Reverse ETL out of the Parquet table | 74 staged, 74 succeeded, 0 failed |
| Round trip vs the live API | **IDENTICAL** — 0 value disagreements, 0 records missing |
| Append mode over 3 live runs | 11 → 22 → 33 rows, 11 distinct names — append semantics preserved by the WAL replay |
| Deduped mode over repeated live runs | 74 rows, **0** duplicate primary keys |

The `issues` stream synced 0 records, which is correct: GitHub's issues endpoint returns pull
requests too and the connector filters them out — all 11 open "issues" on this repository are pull
requests. It also exercised the zero-row path against live data.

Full transcript: `traces/live-github-proof.md`.

## Two defects found after the first green, both by looking rather than by luck

| found by | defect | fix |
|---|---|---|
| driving the binary through the migration case | the **read** path refused a pre-Parquet warehouse, the **write** path did not — `pm etl run` reported `records_loaded: 3`, `status: completed`, exit 0 into a warehouse where every read was refused | `CheckLegacyTableFormat` on the sync path, `Warehouse.Write`, and `Warehouse.ValidateWrite` |
| reviewing the write path for Windows portability | JSON schema inference sampled a **bounded 20,480-row prefix**, so a table larger than that with a late-appearing or late-widening field **could not be materialized at all** | `sample_size=-1` |

Neither was reachable by the original tests: the first needed the real binary, the second needed
more rows than any test or the live sync had. Both now have regression tests.

A third, caught before CI could: `File.Sync` on a read-only handle is `FlushFileBuffers` on Windows,
which refuses a handle without write access. The table is now opened read-write purely so the sync
is portable.

## The layout decision, in one line

A table is **one Parquet file**. Parts bought 1.5% on a 1 M-row whole-row read (600 ms vs 609 ms —
DuckDB already parallelises across the 9 row groups inside the single file) and cost 14.5× on write
(2.517 s vs 173 ms). And a directory cannot be renamed into place: `rename(tmp, existing-dir)` fails
with `file exists`, and the two-rename workaround leaves a window where `stat` on the table returns
`no such file or directory` while the rows are intact on disk. Full transcript in
`traces/layout-decision-experiment.log`.

Partitioning is not foreclosed — `read_parquet` takes a glob and file-vs-directory is one `os.Stat`,
so a future partitioned form is detectable rather than guessed.

## Binary size, measured

darwin/arm64, release ldflags (`-s -w` plus version stamps):

| build | bytes | |
|---|---:|---|
| `CGO_ENABLED=0`, no tag (previous release build) | 79,040,146 | 75.4 MiB |
| `CGO_ENABLED=1`, DuckDB embedded | 111,529,074 | 106.4 MiB |
| **delta** | **+32,488,928** | **+31.0 MiB, +41.1%** |

With `-trimpath -ldflags="-s -w"` (the profile the website quotes): **111,479,570 bytes**, up from
the 59,172,752 recorded there. Unstripped local build: **137,779,042 bytes**. Website table updated.

This matches the research's independently measured +31.0 MB / +42%.

## Platform matrix

`go-duckdb@v1.8.5` bundles a static library for `darwin/{amd64,arm64}`, `linux/{amd64,arm64}`,
`freebsd/amd64` and `windows/amd64`. There is none for `windows/arm64`, and the module's production
build constraint says the same. **That is the only target dropped.**

This machine has no mingw-w64 and no zig, so windows/amd64 cannot be built here — the same toolchain
gap that produced a wrong answer in the research's earlier round, so it is not asserted from a local
result. Two CI jobs were added to `verify.yml` instead:

- `native-build` — builds `pm` with cgo on `ubuntu-latest`, `macos-latest` and **`windows-latest`**,
  and runs the materialization, query and reverse-ETL tests on each. This is the verification the
  brief asked for, and it runs on this PR.
- `windows-arm64-unsupported` — asserts go-duckdb still ships no `windows_arm64` library, so the
  single exclusion stays justified by a fact and a future go-duckdb that adds it fails loudly
  asking for the target back.

If `native-build (windows-amd64)` fails, the casualty is all of Windows rather than Windows-on-ARM,
which is a materially different trade and reopens the decision.

## Existing warehouses

**No released `pm` has ever written a nested-layout table.** The last release is `v0.1.1`
(`.release-please-manifest.json`, `CHANGELOG.md`); #3901 introduced the nesting and carries no tag.

- **Released warehouses** are the flat legacy layout, already refused by `CheckLegacyLayout` with a
  rebuild instruction. **This change compounds nothing**: the single rebuild #3901's release note
  already asks for now produces Parquet tables directly. The release note needs no second
  instruction.
- **Unreleased-`main` warehouses** are nested with JSONL tables. Those are detected and refused by
  name on read *and* on write, never read and never deleted. The WAL is untouched, so re-running the
  sync rebuilds each table losslessly.

Verified against the binary: a planted `tables/*.jsonl` makes reads refuse and name it, the file is
byte-identical afterwards, and removing it restores reads.

## The release pipeline — decided, and why

Embedding DuckDB makes every release target a cgo build, and the previous pipeline produced all of
them from one Linux host. Three probes settled how to replace it.

**Probe 1, zig cc: 0 of 5.** Every target failed, including `linux/amd64` — the runner's own
architecture — which is what proved the blocker was not cross-compilation. `libduckdb.a` is C++ and
the link needs a C++ standard library ABI-compatible with an archive built by someone else's
toolchain; zig has none (`undefined symbol: typeinfo for std::bad_weak_ptr`, and on darwin a missing
libresolv). **That ruled out zig, not cross-compiling** — a distinction this record got wrong on the
first pass and the captain corrected.

**Probe 2, `goreleaser-cross` (real GCC + osxcross): 4 of 5.**

```
RESULT linux/amd64:   OK
RESULT linux/arm64:   OK
RESULT darwin/amd64:  OK      osxcross links go-duckdb's Mach-O archives fine
RESULT darwin/arm64:  OK
RESULT windows/amd64: FAILED  collect2: error: ld returned 1 exit status
```

Its first run reported 0 of 5 and was a **false negative**: every target died on
`error obtaining VCS status: exit status 128` — git refusing the bind-mounted workspace as dubious
ownership, before any compiler ran. The compiler-path listing printed by the probe is the only
reason that was visible rather than being read as a toolchain verdict. Fixed with
`safe.directory` + `-buildvcs=false`, and the real answer is above.

So even the best free cross image cannot produce windows/amd64, and a native Windows runner is
required regardless. **A matrix is unavoidable, which is what option E already is.**

**Decision (captain): option E — a GitHub Actions native matrix.** Each target builds on a runner of
its own OS, only on release tags so pull requests stay cheap; one publish job assembles the release
and the existing trust steps run unchanged. No GoReleaser Pro licence.

**Cost: zero.** The repository is public, and standard GitHub-hosted runners are free for public
repositories on every OS. The one exclusion is *larger* runners, which are always billed — this
matrix uses only standard labels (`ubuntu-latest`, `ubuntu-24.04-arm`, `macos-13`, `macos-latest`,
`windows-latest`), so nothing here is billable. Pull requests never touch macOS or Windows runners.

Route C, hand-rolled packaging, was ruled out by the captain and is not what this is: deb and rpm
are still built by **nfpm, the same tool GoReleaser embeds**, from a config that mirrors the old
`nfpms` block. Only archive and checksum assembly is done directly, and it is guarded — see below.

## Gates run

`gofmt -l cmd internal` (clean) · `go vet ./...` (clean) · `go build ./cmd/pm` ·
`go test -timeout 20m` on `internal/app`, `internal/warehouse`, `internal/connectors`,
`internal/cli` (all ok) · `make tidy-check` · `make agent-contract-check` ·
`make connectorgen-validate` · `make connectorgen-surface-sync` · `make connector-boundary` ·
`make release-workflow-check` · `make docs-check` · `make smoke` · `make lint`.

## Full-tree verification — done, and green

`CGO_ENABLED=1 go test -timeout 20m ./...` against `111a8cdcb`, output written to a file and
grepped whole rather than tailed:

```
EXITCODE=0
161 packages ok, 3 with no test files, 0 FAIL lines
internal/cli                              ok  665.437s
internal/app                              ok  219.625s
internal/connectors/certify               ok   59.614s
internal/connectors/defs/zendesk-support  ok   14.689s
internal/warehouse                        ok    2.744s
```

Seven packages came back `(cached)` on the first pass, so they were re-run with `-count=1` — all
seven pass fresh. No result in the statement above rests on a cached one.

**Nothing outstanding.** The section below is kept because the mistakes it records are the reason
this took three CI rounds, and they are worth not repeating.

## What caused three CI rounds

Every failure since the first green has been the **same fixture defect** — a test hand-writing a
root-level warehouse table as JSONL, a format `pm` now refuses — and each round found more of them
because my local sweeps were not authoritative:

1. `go test ./internal/connectors/` does **not** cover `./internal/connectors/certify/` or
   `./internal/connectors/defs/zendesk-support/`. Package paths without `/...` match one package.
2. I read a sweep's output through `tail -12` and treated the visible failures as the complete set.
   `TestQueryRunAgentModeSummaryProjectsFields` and `TestQueryRunAgentModeStreamProjectsNDJSON`
   were truncated out of view in exactly that way.

**The rule that follows: run `CGO_ENABLED=1 go test -timeout 20m ./...` to a file and grep the file
for `^(FAIL|--- FAIL)`. Never pipe the run through `tail`, and never scope a sweep with a package
path that lacks `/...`.**

### If more fixtures ever turn up

They will look identical. The fix is always the same: replace the hand-written JSONL with
`warehouse.WriteTable(ctx, <path>+warehouse.TableFileExt, []warehouse.Row{...})`. Already converted:
`internal/app` (4 helpers), `internal/warehouse/layout_test.go`, `internal/connectors/warehouse_test.go`,
`internal/connectors/certify/stages_write_test.go`, `internal/cli/reverse_cli_test.go`,
`internal/cli/agentmode_query_cli_test.go`,
`internal/connectors/defs/zendesk-support/reverse_etl_execute_test.go`.

Legitimate remaining `.jsonl` writes that must **not** be converted: the WAL
(`wal/<stream>.jsonl`), the outbox, and the deliberate legacy-layout fixtures in
`warehouse_connection_isolation_test.go` and `warehouse_parquet_test.go`, which exist to prove the
refusal fires.
