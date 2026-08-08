# Context — MySQL container harness R1

## Intent

Deliver one reusable, opt-in Podman integration-test harness and prove it with exactly one new
Tier-3 native database connector: MySQL. The harness starts one isolated, pinned-image container;
seeds deterministic multi-page data; exercises check, catalog, snapshot, incremental, and binary-log
CDC paths; and reclaims its container, named volume, and run-owned image references on every exit
path.

## Scope and constraints

- The tagged test is sequential and opt-in (`POLYMETRICS_DATABASE_INTEGRATION=1`). A missing
  opt-in or `POLYMETRICS_PODMAN_CONNECTION` produces a visible skip; once opted in, startup or
  reachability failure is red, never a green no-op.
- Every container command has the caller-supplied `--connection`; the harness never uses or mutates
  the Podman global default. Machine trim uses the explicit configured machine name.
- The host port is Podman-assigned on loopback and is rejected if it equals the
  engine default. The connector receives host and port as separate fields; it does not construct a
  logged endpoint.
- The MySQL source image is pinned at `docker.io/library/mysql:8.4.11`. Each run creates a unique
  local tag, then cleanup attempts container → volume → run tag → pulled source image even after an
  earlier cleanup error. `POLYMETRICS_DATABASE_KEEP_IMAGE=1` is the only retention opt-in.
- Cleanup runs two explicit `fstrim` passes against the configured machine after Podman cleanup.
  This is required on macOS VM-backed storage to return freed guest blocks to the host sparse disk.
  It is no longer an opt-in, but it runs only against a machine this process created through
  `dbtest.NewMachine` and still holds an ownership record for. A matching name is not ownership:
  `fstrim -av` reaches every filesystem on a machine, so a caller-supplied, pre-existing, shared, or
  remote machine is reported with its still-reclaimable byte count and never trimmed. The scoped
  connection and a live `podman machine inspect` follow as defence in depth.
- `POLYMETRICS_DATABASE_OWN_MACHINE=1` is how the live proof creates, uses, and deletes its own
  machine; that is the mode in which the reclaim is exercised at all.
- The machine ownership record is taken before `podman machine init` runs, matching the container
  path's claim-before-create rule: init writes a multi-GiB disk image before it can fail, and a
  cancelled context kills it mid-write. A failed or cancelled init tears its own machine down on a
  deadline of its own, against that exact generated name.
- The test database uses only its isolated ephemeral server configuration. No credential or
  connection string is printed, logged, or stored.
- MySQL is a dynamic-schema Tier-3 native connector. Its binary-log mechanism is declared through
  the closed `binlog_replication` vocabulary only because a matching live executor exists.

## Post-rebase reconciliation — 2026-08-08

- Rebased the branch onto current `origin/main`, including the warehouse layout, Parquet/DuckDB,
  direct-read page-context, and derived-command-parameter changes. `connectorgen surface-sync`
  was rerun and reported zero fields to fill or correct; docs/catalog and website artifacts were
  regenerated rather than hand-merged.
- #3902's `DirectReadPage` contract governs one-page HTTP/API exploration. MySQL declares no REST
  operation or direct-read command and does not implement `DirectReader`; its `Read` is ETL bulk
  extraction that drains deterministic keyset SQL pages subject to `page_size` and `read_limit`.
  `TestReadIsETLNotPagewiseDirectRead` makes that boundary explicit.
- The captain's shared SQL TLS ruling still required a narrow PostgreSQL adjustment. PostgreSQL now
  resolves the same `sslmode` / `sslrootcert` / `sslservername` options and routes Check, Catalog,
  and Read through one pool constructor. This touches no PostgreSQL write path.
- Recovery of the prior pipeline head found a generated production-native wiring issue. `dbtest`
  and shared `sqltls` are support libraries under `native/`, not registrations; the generator now
  excludes them by source policy and regenerates the native set, with red/green test coverage.

## Inline GSD fallback

`scripts/gsd doctor`, source resolution and generated prompts for `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`, plus
`go run ./cmd/agentcontractgen check`, were completed. This firstmate task has no compatible
isolated GSD worker and the delivery contract forbids role spawning, so the lifecycle and TDD
evidence are maintained inline in this phase directory.

## Required skills loaded

`golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`,
`golang-concurrency`, `golang-database`, `golang-documentation`, `golang-dependency-management`,
and `golang-pkg-go-dev`.
