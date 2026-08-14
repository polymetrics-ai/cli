# Context — Issue 3980: immutable Parquet delivery worksets

## Task Delivery Header

- Issue: Closes #3980 — Postgres Parity: derive immutable Parquet delivery worksets
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with green checks; after opening, verify its API-reported base is exactly `integration/4015-mvp-flat-r1`.
- Working branch: `fm/cli-3980-parquet-worksets-r1`
- Task: Derive immutable, StreamID-keyed Parquet delivery worksets.
- Verification: `go test -timeout 20m ./internal/warehouse/... ./internal/connectors/database/... ./internal/synctransport/...`, then the applicable local verification gates and green CI.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| DuckDB/Parquet fixtures derive deterministic inserts, updates, unchanged rows, and explicit tombstones across composite keys and null/type edge cases. | live | Tests materialize real input/baseline Parquet through `warehouse.WriteTable`, derive twice, reopen the produced Parquet, and assert its rows and counts. |
| Physical absence never emits a tombstone. | live | A source row absent from the complete projection produces no tombstone artifact row; only the explicit tombstone input can produce one. |
| Identical inputs yield the same immutable identity/content hash; changed schema, key, or destination invalidates reuse. | live | Tests compare the two identity encodings byte-for-byte, then vary each binding and assert a distinct identity. |
| Mutating source data after derivation cannot alter an existing workset. | live | The test replaces the source Parquet after deriving, reopens the old workset, and asserts the prior output hash and row content remain unchanged. |
| Candidate baseline remains unadvanced until a later target receipt binds it. | live | Derivation creates a separate candidate baseline artifact and leaves the supplied baseline file byte-identical; this foundation has no target driver, receipt store, or write-session call to mutate it. |
| Temporary resources are bounded and cancellation cleans partial artifacts. | live | A canceled derivation returns the context error and leaves no workset directory or partial Parquet file at the requested root. |

## Decision Record

- This is a shared foundation (F5), not a PostgreSQL driver or connector-lane change. It changes only `internal/connectors/database`, `internal/warehouse` if a narrow generic helper is required, and its tests/evidence.
- `ManagedTargetDeliveryLedgerKey` is the only workset destination identity. It is derived from the asserted `ManagedTargetControlRecord`, retaining source owner, asserted target database identity, immutable StreamID, namespace, and relation. A warehouse `ArtifactRef.Table`, source display name, and target display/table text are provenance only and cannot participate in workset reuse or storage identity.
- The workset owns a complete immutable Parquet projection, a keyed insert/update delta, explicit tombstones, and a separately materialized candidate baseline. It accepts a prior baseline as read-only input and never infers deletion from physical absence.
- The sealed manifest binds delivery-key identity, target schema version/fingerprint, ordered composite key fingerprint, source and baseline content versions, bounded counts, and content hash. Public accessors expose defensive copies or opaque strings only.
- This slice deliberately performs no target DML, durable receipt persistence, source checkpoint advancement, or baseline promotion. #3983 consumes the candidate only after its target-session receipt; #3979 owns snapshot/bootstrap behavior. The observable safety guarantee here is that derivation cannot advance or overwrite the supplied baseline.
- Workset artifacts are created in a caller-owned root under a deterministic, identity-safe content address. A request must declare a finite `MaxArtifactBytes` ceiling (at most 1 GiB); every input/output artifact is checked against it, and staging is removed on cancellation/failure. Existing matching artifacts are reopened only after their manifest and files validate; different schema/key/destination bindings are distinct addresses.

## Required Skills and GSD Fallback

- Loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, and `golang-database`.
- Passed `scripts/gsd doctor`, resolved the five required lifecycle commands, and passed `go run ./cmd/agentcontractgen check`.
- The canonical single-worker contract forbids role spawning in this lane. Generated GSD prompts are executed inline; this context, discussion log, plan, TDD ledger, verification, and review record are the documented manual fallback.
- No CLI surface, connector definition, credentials, target connection, or documentation surface changes; CLI help/manual/website parity is not applicable.
