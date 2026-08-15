# TDD ledger — Issue #3976: PostgreSQL dynamic typed catalog discovery

## R2 — resumable source reads (active)

| ID | Class | Guarantee | RED assertion | GREEN proof |
| --- | --- | --- | --- | --- |
| RR1 | Happy path | PostgreSQL incremental/cursor reads are reached through #3858 rather than a private loop. | The real binary/sync construction path does not produce the shared source executor or resume tuple. | It produces exact records and resumes from the committed tuple through the shared executor. |
| RR2 | Bad path | A cursor-required stream cannot silently ignore a stored checkpoint. | An unset cursor enters a source session/query or returns all rows. | A typed missing-stream-cursor refusal occurs before source I/O, delivery, or checkpoint mutation. |
| RR3 | Edge case | Null cursor values cannot silently disappear. | A nullable fixture loses its null-cursor identity under the PostgreSQL source path. | The lossless shared ordering contract returns every identity exactly once, or a typed fail-closed boundary prevents delivery. |
| RR4 | Bad path | Stale/invalid checkpoints do not restart from page one. | An incompatible checkpoint is refused only after the reader starts, or becomes a full read. | The recovery reason is specific and no source/delivery/checkpoint side effect occurs. |
| RR5 | Edge case | Cursor selection belongs to a stream, not a connection. | A second relation is bound to the first relation's cursor column. | Each stream binds and validates its own catalog cursor; a mismatched/missing one rejects specifically. |

| ID | Guarantee | RED assertion | GREEN proof |
| --- | --- | --- | --- |
| R1 | Dynamic variation | One static catalog/table/field model cannot satisfy two materially different live PostgreSQL schema fixtures. | Each fixture produces a distinct typed catalog/fingerprint without code or connector-bundle schema edits. |
| R2 | Structured source identity | Database/schema/table identity is flattened into a string or same-named relations across schemas merge. | Typed catalog values retain configured database, schema, and relation identity separately. |
| R3 | Ordered metadata | Column order, nullable state, primary-key membership/order, or deterministic relation order can be lost. | Live discovery and an independent PostgreSQL catalog oracle agree on ordinal order and metadata. |
| R4 | Supported native/logical mapping | PostgreSQL native identity/modifiers collapse into the old coarse field type. | Supported native types retain native identity/modifiers and map through #4034's logical types. |
| R5 | Fail closed | Enum/domain/composite/unsupported shapes silently become string/object or return a guessed catalog. | A typed, named, secret-safe unsupported-shape error is returned before success. |
| R6 | Runtime connection | #4034's typed catalog foundation remains descriptor-only while production PostgreSQL `Catalog()` uses a disconnected static/coarse model. | The native runtime calls the typed adapter and any legacy projection derives from its result. |
| R7 | Scope / safety | System schemas, views, arbitrary SQL, target DDL/write, Parquet materialization, or CDC execution leak into source discovery. | Configured-schema base-table reads only; ownership tests/changed-path audit prove downstream boundaries stayed untouched. |
| R8 | Resource/cancellation | Queries ignore cancellation, leak rows/pools, or concatenate configured identity into SQL. | `QueryContext`, parameter binding, close/error paths, cancellation tests, and race/vet/lint proof pass. |
| R9 | Live source reads | A container test can exit successfully without proving catalog, full-read, or cursor-advanced records. | A real PostgreSQL harness seeds distinct rows and asserts catalog details, full primary keys, and the post-cursor primary keys/values. |
| R10 | Cursor contract observation | An absent, nonexistent, nullable, or connection-level `cursor_field` silently produces an undocumented claim. | The live proof logs and asserts each present behavior without changing the separately deferred product contract. |

## RED command target

Captured before implementation in `traces/dynamic-catalog-red.txt`:

```sh
go test -tags=databaseintegration -timeout 20m -count=1 \
  ./internal/connectors/native/postgres \
  -run '^TestPostgresDynamicTypedCatalogUsesLiveMetadata$'
```

The regression fails because the shipping connector has no typed catalog
boundary or named unsupported-shape rejection. It is a real PostgreSQL
`dbtest` fixture that creates two materially different schemas and an
unsupported enum type; it never prints a connection string or credential.

## GREEN command targets

```sh
go test -timeout 20m -count=1 ./internal/connectors/native/postgres
go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres
```

GREEN proof so far:

- `go test -timeout 20m -count=1 ./internal/connectors/native/postgres` — pass.
- `go test -race -timeout 20m -count=1 ./internal/connectors/native/postgres` — pass.
- `go vet ./internal/connectors/native/postgres` — pass.
- `go test -timeout 20m -count=1 ./internal/connectors/database` and
  `./internal/connectors/engine` — pass.
- The exact live-harness regression ran against PostgreSQL 16.10 through
  Docker and Colima's explicit Unix endpoint. It discovered the real seeded
  catalog, returned full IDs `1,2,3,4,5`, and after cursor `10` returned only
  `3,4,5`. The command and verbatim output are retained in
  `traces/live-reads-green.txt` and posted on issue #3976.
- Final green gates: focused PostgreSQL/database/engine packages, the
  PostgreSQL race suite, `internal/cli`, `go vet ./...`, `go build ./cmd/pm`,
  and every individually invoked repository gate in `VERIFICATION.md` passed.

Broader checks and their outcomes belong in `VERIFICATION.md`.

## Live-proof resumption red/green

**Red:** `dynamic_catalog_integration_test.go` is extended first to require
the base-owned explicit Docker/Podman harness constructor and live row
assertions. Before that constructor is supplied, the databaseintegration
package must fail to compile; the exact output is retained as
`traces/live-reads-red.txt`.

**Green:** after adding only the focused test-harness wiring and assertions,
run the pinned Docker command against Colima. The test must log real catalog,
full, incremental, and cursor-boundary records; its exact output is retained
as `traces/live-reads-green.txt` and posted verbatim to issue #3976.
