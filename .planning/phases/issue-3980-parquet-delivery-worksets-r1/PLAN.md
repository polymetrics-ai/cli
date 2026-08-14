# PLAN — Issue 3980: immutable Parquet delivery worksets

## GSD setup and fallback

- Passed `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check`.
- Resolved sources and generated prompts for `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`.
- This shared foundation is outside a numbered roadmap phase. The canonical
  single-worker contract forbids role spawning, so the generated prompts are
  executed inline; the phase artifacts are the manual fallback record.
- Required skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, and `golang-database`.

## Goal

Create a sealed, StreamID-keyed `ChangeDeliveryWorkset` from a complete
connection-owned source Parquet snapshot, a read-only per-destination baseline,
and explicit tombstones. Its manifest and immutable files must be independently
reopenable and must never identify or reuse a target by a mutable source table
or display name.

## Contract shape

1. Add a concrete `ChangeDeliveryWorksetRequest` and a concrete immutable
   workset to `internal/connectors/database`. The request accepts an asserted
   `ManagedTargetControlRecord`, ordered composite keys, source/baseline Parquet
   paths, explicit `synccontract.Tombstone` values, and a caller-owned artifact
   root. It accepts no raw SQL, destination table, display name, credential, or
   direct target handle.
2. Derive `ManagedTargetDeliveryLedgerKey` from the control record and use it,
   its schema version/fingerprint, ordered key fingerprint, source/baseline
   content versions, and canonical tombstone representation to build a domain
   separated SHA-256 identity. The manifest carries these bindings and output
   counts/content hash; accessors return copies only.
3. Materialize a content-addressed workset directory atomically: complete
   projection Parquet, keyed insert/update delta Parquet, deterministic
   explicit-tombstone artifact, and candidate baseline Parquet. Source/baseline
   inputs are never modified. If a matching directory exists, reopen only after
   validating its manifest and every content hash.
4. Use DuckDB over the real Parquet inputs for the keyed delta. Validate keys,
   reject duplicate or null key slots before publish, and emit source rows that
   are absent from the baseline or whose same-key JSON-shaped row is different.
   Never derive a tombstone from a missing source row.
5. Bound each copy/hash/query stream with fixed buffers, check cancellation
   before and during each phase, write to a unique temporary directory, and
   remove it on cancellation/failure. This foundation deliberately does not
   persist a receipt or promote the candidate baseline.

## TDD slices

1. **Red — immutable target identity:** Add a real Parquet fixture test that
   calls the missing derivation twice, compares identity bytes and manifest
   binding, and varies destination/schema/key. It must fail to compile before
   the type/constructor exist.
2. **Green — sealed artifact and reuse:** Implement validated request/manifest,
   deterministic content identity, atomic artifact publication/reopen, and
   defensive immutable accessors. Re-run the red test to prove byte-identical
   identity and that source-table/provenance mutation cannot select another
   workset.
3. **Red/green — actual DuckDB delta:** Use source/baseline Parquet fixtures
   with composite keys, unchanged records, inserts, updates, null payloads and
   type-sensitive values. Assert the reopened delta rows/counts, full projection
   rows, and candidate baseline—not merely a successful command.
4. **Red/green — deletion and mutation safety:** Provide explicit tombstones
   and a source snapshot with a physically absent baseline row. Assert exactly
   the explicit tombstone is emitted; replace the source Parquet after
   derivation and prove the original artifact/hash remains unchanged.
5. **Red/green — refusal and cleanup:** Give duplicate/null keys, canceled
   contexts, and stale/corrupt artifact manifests. Assert zero published
   worksets/partial files, zero baseline-byte change, and no reuse of invalid
   artifacts.

## Guardrails

- Touch only `internal/connectors/database`, narrow warehouse helpers only if
  necessary, their tests, and this issue’s planning evidence.
- Do not add a dependency, driver/DDL/SQL renderer, target write session,
  receipt/ledger persistence, baseline promotion, source checkpoint, CLI/docs
  surface, connector registration, or capability claim.
- Treat all filesystem paths as data: generate fixed child names only; quote
  DuckDB literals through the repository helper and validate keys before they
  become internally generated identifier references.
- Do not change `WarehouseWorkset` or use it as a substitute for this contract.
- Every refusal test must prove no workset publication and no supplied-baseline
  mutation, not only that an error returned.

## Checkpoints

1. Commit the GSD/TDD plan artifacts.
2. Commit preserved red-test output.
3. Commit the green workset implementation and focused real-Parquet tests.
4. Commit only verified review/gap fixes, then push the green slice and open the
   explicitly based PR.
