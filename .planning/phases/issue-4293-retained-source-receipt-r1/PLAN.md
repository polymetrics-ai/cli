# Plan — Issue #4293 retained source-evidence receipt cohort R1

## Task Delivery Header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator.
- Base branch: `fm/cli-top100-declaration-batch-r1` at frozen commit `ceaae873aef0dd19aa23c036b9cb598f9b3eacc8`.
- Merges into: the Batch R1 parent branch only; no parent integration, pull request, or `main` merge is authorized by this slice.
- Delivery: one candidate-only commit and normal push after independent review-ready scoped checks.
- Working branch: `fix/4293-retained-source-receipt-cohort-r1`.
- Task: add a deterministic, read-only cohort validator that re-derives and exact-byte checks the eight existing v2 `retention_only` mapping receipts from their frozen source locks and connector-owned lane matrices.
- Verification: focused red/green `cmd/connectorgen` tests; command help; the Batch R1 receipt check; JSON parse; `gofmt`; focused `go vet`; agent-contract check; and `git diff --check`.

## Accepted design

The eight descriptor-free v2 locks already have one connector-owned retention
receipt each.  The safe bridge is not a descriptor or a runtime artifact: it
is the existing `sources/<connector>-retained-mapping-contract.json` sidecar.
Wave 0 adds a cohort-level **read-only receipt check** that:

1. validates the immutable ten-lock Batch R1 cohort first;
2. derives eligibility from lock facts, not connector names or operation-ID
   allowlists: v2, non-canonical evidence, complete retained REST source
   contract and source-operation objects, and zero GraphQL inventory;
3. rebuilds the existing retention-only contract from that lock plus its owned
   matrix, checks its exact source-ID reconciliation, and exact-byte checks the
   persisted sidecar;
4. reports the eligible connector count, retained source-operation count, and
   zero executable declarations; and
5. rejects missing, malformed, unsafe, or byte-drifted receipts before any
   descriptor, projection, source import, CLI, transport, credential, or
   runtime path is reached.

The frozen cohort currently yields eight eligible receipts and 2,340 retained
source IDs.  That number is a test witness of the immutable cohort, not a
runtime admission rule and not a hardcoded connector/operation exception.

## TDD slices

1. **RED:** add focused tests for a clean eight-receipt/2,340-ID check and CLI
   output, a missing or byte-drifted fixture sidecar, option misuse, CircleCI's
   `operations`/`source_operation_id` matrix form, and zero executable
   declaration claims.  Run the focused command before production changes.
2. **GREEN:** add a narrowly scoped `--check-retention-receipts` option to the
   existing `source-operation-mapping-cohort` developer command.  Reuse the
   existing retained-source builder and sidecar validator; refactor only enough
   to let the cohort checker use a verified cohort entry without re-entering a
   hardcoded per-connector path.
3. **REFACTOR:** keep reports deterministic and source-accounting-only,
   preserve the legacy `--check` output, and add no source lock, matrix,
   connector definition, descriptor, enabled root contract, or engine change.

## Acceptance evidence

| Requirement | Observable proof |
| --- | --- |
| All eight v2 receipts are bound to frozen source evidence | The new check discovers eligible v2 locks from the ten-lock cohort and reports 8 connectors / 2,340 source IDs with no findings. |
| A receipt cannot silently drift | A temporary repository fixture with a missing or byte-drifted sidecar returns a connector-scoped finding. |
| Both retained matrix forms remain supported | The clean cohort proof includes CircleCI's `operations` plus `source_operation_id` form, while existing decoder tests preserve both forms. |
| Zero executable claims | Each rebuilt and persisted contract passes `ValidateRetentionOnly`; the receipt report totals zero executable declarations and tests reject runtime fields/implemented coverage. |
| Normal runtime admission stays closed | No change is made to `source-import`, `source-materialize`, `sourceprojection`, engine, or runtime bundle loading. Existing canonical-evidence guard test remains part of the focused suite. |

## GSD, skills, and CLI parity

- Ran `scripts/gsd doctor`, resolved all five required GSD command sources, and
  generated the `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` prompts. Compatible isolated GSD workers
  are unavailable and repository policy forbids role spawning, so the phases
  are executed inline and recorded here.
- Loaded: `connector-lane-build-order`, `go-engineering`, `golang-how-to`,
  `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, and `golang-documentation`.
- This changes the developer-only `connectorgen` help surface. The focused
  help test and root usage will be updated. `pm` namespace help, `docs/cli`,
  website pages, generated PM manuals, completions, JSON output, credentials,
  and reverse-ETL safety documentation are not applicable: this command
  creates no PM runtime surface and accepts only a local cohort path plus
  check-only flags.

## Scope and safety boundary

- Included: `cmd/connectorgen` receipt/cohort validator, its focused tests,
  developer help, and this GSD evidence.
- Excluded: source locks, source-lane matrices, connector JSON artifacts,
  canonical descriptors, root enabled contracts, operations/writes/streams,
  CLI bindings, sync transport, source import/materialize/projection, runtime,
  engine, certification, Foundation Atlas, and inbound receivers.
- No new dependency, credential access, provider I/O, shell escape, HTTP
  write, SQL write, or executable connector claim is permitted.
