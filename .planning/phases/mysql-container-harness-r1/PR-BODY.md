## Intent

Land exactly one new engine, MySQL, on a reusable containerised database-test harness. The harness
must prove real connector behaviour and release all of its disk-bearing resources on success,
assertion failure, and interrupt. It is intentionally configuration-driven so MariaDB, PostgreSQL,
SQL Server, and Oracle can follow without copying harness code.

This PR is task-backed rather than issue-backed. It carries the full inline GSD/TDD delivery record
in `.planning/phases/mysql-container-harness-r1/`.

## What Changed

- Added `internal/connectors/native/dbtest`: one sequential-by-default, explicit-Podman-connection
  harness with pinned image enforcement, unique container/volume/run-image ownership, dynamic
  loopback port allocation, before/after disk reporting, unconditional cleanup, and optional
  two-pass machine trim.
- Added the native dynamic-schema MySQL source connector and bundle: check, catalog, bounded full
  and incremental reads, and row-based binary-log CDC. `integration_type` is `database`; `cdc: true`
  is exposed only with its matching executable `binlog_replication` reader.
- Seeded/verified real multi-page data. The tagged proof asserts returned records for check,
  discovery, five-row full read with `page_size=2`, incremental read, TLS modes, and
  insert/update/delete CDC — not merely an exit status.
- Added one shared SQL TLS option shape: `sslmode`, `sslrootcert`, `sslservername`.
  MySQL applies it to normal and replication clients. PostgreSQL now enforces the same shape through
  its pool constructor for Check, Catalog, and Read.
- Recovered the prior pipeline review fix in generator source: production native wiring excludes
  the `dbtest` harness and shared `sqltls` library, then regenerates `nativeset_gen.go`. A
  red/green generator test retains the boundary.
- Regenerated derived catalog and website artifacts from current source; `connectorgen surface-sync`
  is clean. The docs generator also refreshed a stale current-main warehouse catalog description.

### #3902 direct-read page context

No exception was taken to #3902. Its `DirectReadPage` contract is for a one-page HTTP/API
exploration result. MySQL declares no REST operation or direct-read command and does not implement
`DirectReader`; its `Read` is ETL bulk extraction that drains deterministic keyset SQL pages under
`page_size` and `read_limit`. `TestReadIsETLNotPagewiseDirectRead` locks that boundary, and the live
five-row / size-two assertion proves more than one SQL page is consumed.

### PostgreSQL connection change

The PostgreSQL TLS change still applies after rebase. It changes only its connection adapter,
catalog/read pool creation, definition/docs, and tests; **no PostgreSQL write-path file is touched**.
A subsequent PostgreSQL writer should use the shared `conn.openPool` path rather than introduce a
separate `pgxpool.New` construction that could ignore `sslservername`.

### Podman ownership and disk

The documented command requires an explicit `POLYMETRICS_PODMAN_CONNECTION`; no bare container
command is used. A fresh task-owned `fm-cli-db-harness-r1-20260808` Podman machine (2 CPU, 2 GiB,
8 GiB) was created solely for the live proof. After the test, scoped container/volume/image listings
were empty, then that exact machine was stopped and removed. No other machine, container, or volume
was touched.

The harness emits disk-free bytes before and after each engine. With
`POLYMETRICS_DATABASE_RECLAIM_DISK=1`, its live assertion allows only ordinary build noise after
teardown; the fresh proof passed that assertion. Retaining the source image is opt-in only via
`POLYMETRICS_DATABASE_KEEP_IMAGE=1`.

## Dependency Evidence — `github.com/go-mysql-org/go-mysql v1.16.0`

This is the sole new direct module, used for both MySQL client protocol and binary-log replication.

- **Maintenance / release:** the upstream [v1.16.0 release](https://github.com/go-mysql-org/go-mysql/releases/tag/v1.16.0)
  is verified and dated 2026-07-15; the public [upstream repository](https://github.com/go-mysql-org/go-mysql)
  shows active issues, pull requests, and a maintained MySQL/MariaDB replication/client codebase.
- **Licence:** direct module `LICENSE` is MIT. No additional direct module is introduced.
- **Known CVEs:** fresh `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reports
  `No vulnerabilities found.`
- **Transitive footprint:** nine newly listed indirect modules
  (`filippo.io/edwards25519`, `coreos/go-semver`, `pingcap/errors`, `pingcap/log`,
  `pingcap/tidb/pkg/parser`, `shopspring/decimal`, `go.uber.org/multierr`, `go.uber.org/zap`,
  `natefinch/lumberjack.v2`) plus bumps to `goccy/go-json`, `klauspost/compress`, and
  `stretchr/objx`. The TiDB parser is the material size contributor.
- **Binary-size delta:** clean darwin/arm64 `go build -trimpath ./cmd/pm` at dependency addition:

  | build | bytes | MiB |
  | --- | ---: | ---: |
  | baseline | 94,191,266 | 89.83 |
  | MySQL branch | 100,101,650 | 95.46 |
  | delta | +5,910,384 | +5.64 (+6.27%) |

The rebased branch also builds successfully. The increase is recorded explicitly because it should
inform the dependency decision before adding multiple further replication engines.

## Testing

```text
go test -count=1 -timeout 20m \
  ./internal/connectors/native/dbtest \
  ./internal/connectors/native/mysql \
  ./internal/connectors/native/postgres \
  ./internal/connectors/native/sqltls \
  ./internal/connectors/defs \
  ./internal/connectors/engine \
  ./internal/connectors/commandrunner
go test -count=1 -timeout 20m ./internal/cli
go test -count=1 -timeout 20m ./cmd/connectorgen
go vet ./...
go build ./cmd/pm
make tidy-check docs-check-no-build smoke-no-build lint agent-contract-check \
  connectorgen-validate connectorgen-surface-sync connector-boundary \
  release-workflow-check
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

All passed. The tagged live proof also passed with `-timeout 20m`, explicit connection/machine,
and host reclaim enabled. `TestGoldenTranscripts` was covered by `internal/cli` and passed without
regeneration.

## TDD / GSD Delivery Record

- **Red → Green:** initial harness/binlog tests failed before implementation; focused tests passed
  after implementation. After rebase, PostgreSQL definition validation rejected canonical TLS mode
  `disabled`; the added definition/pool tests passed after the shared TLS fix. During recovery of
  the old pipeline custody, generator tests first proved `dbtest`/`sqltls` support libraries were
  incorrectly wired into production, then passed after source-driven regeneration.
- **Lifecycle:** resolved and executed inline prompts for `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed.
  The delivery contract forbids spawning the unavailable compatible GSD role, so the fallback and
  evidence are documented in the phase files.
- **Skills:** `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-database`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`,
  `golang-dependency-management`, and `golang-pkg-go-dev`.
- **Commits/pushes:** coherent slices were pushed after rebase, TLS reconciliation, and direct-read
  boundary review. The current live-start diagnostic is included in the final pushed slice.

## Ready Next Engines

| Engine | Harness fit | Remaining blocker |
| --- | --- | --- |
| MariaDB | Same `dbtest.Config` shape and closely related binlog protocol. | Confirm version-specific GTID/checkpoint semantics and its closed mechanism declaration. |
| PostgreSQL | Same harness configuration shape. | Implement logical replication with approved `pglogrepl`; test must remove replication slots to avoid server-side WAL leakage. |
| SQL Server | Same lifecycle/cleanup shape. | Choose/approve driver, review image licence acceptance, and raise memory budget for the image. |
| Oracle | Same harness can own image/container/volume lifecycle. | Image size/licence/architecture and registry-auth constraints; assess LogMiner/XStream feasibility. |

## Pipeline

No human merge is requested. The PR is opened for automatic review and CI; any actionable review
finding will be dispositioned in the PR before handoff. `no-mistakes` delivery validation is the
next PR-stage gate and its result will be added here.
