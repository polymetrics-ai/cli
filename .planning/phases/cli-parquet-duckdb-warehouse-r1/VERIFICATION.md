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

## Open item — the release pipeline

`.goreleaser.yaml` now sets `CGO_ENABLED=1` and drops `windows/arm64`, and
`scripts/verify-release-assets.sh` expects five archives instead of six. **The release jobs
themselves still build all targets from one `ubuntu-latest` host**, which worked when
`CGO_ENABLED=0` and cannot work now: cross-compiling cgo to darwin and windows from Linux needs
toolchains the runner does not have.

The obvious fix — a native build matrix feeding `goreleaser`'s `prebuilt` builder — is **not
available**: `prebuilt` is GoReleaser **Pro** only. That leaves four routes, with genuinely
different costs:

| route | cost | risk |
|---|---|---|
| **A.** zig cc as a cross toolchain on the existing single Linux host | ~15 lines; the whole signed pipeline (goreleaser OSS, cosign, SLSA, nfpm, checksums) unchanged | unknown whether zig can link go-duckdb's prebuilt static libraries, especially the Mach-O ones for darwin and whichever ABI the windows library uses |
| **B.** GoReleaser Pro + native build matrix + `prebuilt` builder | a paid licence | low — this is the shape the tool is designed for |
| **C.** Native matrix + hand-rolled archives and `nfpm` CLI | no licence | highest — rewrites a signed, attested release path by hand |
| **D.** Publish a smaller matrix until one of the above lands | none | a product decision about which platforms ship |

### The probe answered: route A is dead

CI ran the probe. **All five targets failed, including `linux/amd64` — the runner's own
architecture**, which is the finding that settles it:

```
RESULT linux/amd64:   FAILED
RESULT linux/arm64:   FAILED    ld.lld: undefined symbol: typeinfo for std::bad_weak_ptr
RESULT windows/amd64: FAILED    lld-link: undefined symbol: std::__cxx11::basic_string<...>::reserve()
RESULT darwin/amd64:  FAILED    unable to find dynamic system library 'resolv'
RESULT darwin/arm64:  FAILED    unable to find dynamic system library 'resolv'
```

A native-architecture build failing proves the blocker is **not cross-compilation**. `libduckdb.a`
is a C++ library, and the link needs a matching C++ standard library — `-lc++` on the darwin link
line, libstdc++ symbols on the others. zig does not supply one compatible with a prebuilt archive
built by someone else's toolchain. No flag fixes that; it is the wrong tool for linking against a
third-party C++ static library.

**That kills zig, not route A** — a distinction this record got wrong on the first pass and the
captain corrected. A real GCC cross image ships precisely what was missing: matching libstdc++ for
the linux targets, mingw-w64 for windows, and osxcross (clang + macOS SDK + libc++) for darwin. That
is a different failure mode, so it is probed the same way before any money is spent.

**Probe 2 — `ghcr.io/goreleaser/goreleaser-cross:v1.25.9`**, same five targets, same
`RESULT <target>: OK|FAILED` format, still non-blocking. It additionally prints which cross
compilers the image actually ships, so a *missing compiler* can never be mistaken for a *link
failure* — the distinction that made probe 1 conclusive.

If it passes, the release runs `goreleaser` inside that container and the signed pipeline shape is
unchanged. If it fails too, no free cross route remains and B (a paid GoReleaser Pro licence) is
settled — a purchasing decision that is not this agent's to make.

Route C (hand-rolled archives and `nfpm`) was **ruled out by the captain**: it puts the risk on the
signed and attested path, which is the worst place for it. That ruling stands regardless of the
probe.

Until B lands, `release / package-check` fails on the PR. Deliberately visible.

## Gates run

`gofmt -l cmd internal` (clean) · `go vet ./...` (clean) · `go build ./cmd/pm` ·
`go test -timeout 20m` on `internal/app`, `internal/warehouse`, `internal/connectors`,
`internal/cli` (all ok) · `make tidy-check` · `make agent-contract-check` ·
`make connectorgen-validate` · `make connectorgen-surface-sync` · `make connector-boundary` ·
`make release-workflow-check` · `make docs-check` · `make smoke` · `make lint`.
