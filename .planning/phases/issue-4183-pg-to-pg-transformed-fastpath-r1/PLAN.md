# Plan — issue #4183 PostgreSQL transformed fast path R1

## Scope and skills

Target connector: PostgreSQL. The only connector-specific code is `internal/connectors/native/postgres`: a bounded range extractor behind the neutral extraction port and a binary-COPY shadow publisher behind the neutral bulk-apply port. Shared packages may change only to supply connector-neutral contracts named by the captain: `internal/synctransport` (controller/byte credits/timing/checkpoint sequencing), `internal/warehouse` (versioned segments and atomic manifest), and `internal/connectors/database` (closed transform and receipt contracts).

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-performance`, `golang-benchmark`, and `golang-observability`.

GSD uses the inline/manual fallback because this worker has no compatible Pi runtime for the generated role commands. Commands resolved: `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.

## CI fix round 1

This narrow follow-up preserves the delivered slice and corrects CI-only drift: make the App transport test destination implement the already-required run-scoped `full_overwrite` port; regenerate only verified `--transform-file` transcript and website-data outputs; and upgrade the existing direct Arrow dependency or its already-transitive Thrift dependency to remove the Dependency Review finding. The target remains PostgreSQL; no production fast-path semantics or recorded 111.78/111.85 MB/s benchmark values change. The inline GSD fallback is revalidated with the same five resolved commands. Required skills additionally used for this repair: `golang-dependency-management`, `golang-lint`, and `golang-troubleshooting`. The double repair revealed that six legacy all-mode cases assert a per-page checkpoint before final completion; this conflicts with the atomic full-overwrite receipt-before-checkpoint contract and requires a recorded decision before their assertions can change.

### Decision implementation: acknowledgement-contract split

The accepted `full-overwrite-app-test-semantics` decision narrows the test-harness repair without altering production transport behavior. First restore ordinary-mode `afterApply` to a post-acknowledgement hook; it must not run from the shared page-apply helper before an acknowledgement is constructed. Then keep the inherently per-page source-failure and page-one/page-two stale-writer scenarios only for per-page acknowledgement modes, adding run-scoped inverse/final-CAS proofs. Retain `full_overwrite` in stale-writer and completion/missing-run coverage, but synchronize it at truthful boundaries: post-read-back before the final checkpoint for the stale writer, and after the durable final checkpoint before completion for the two terminal rebase cases. Cancellation-before-publish is the only admitted full-overwrite cancellation contract; the post-final-checkpoint cancellation/failure scenarios remain explicitly inapplicable because the controller has no observation point there. The test double gains distinct page-apply, post-publish, post-read-back, and abort observation rather than overloading an acknowledgement callback. A locker around the state store's durable replacement supplies the terminal-completion boundary without production hooks. The known stale-writer CI timeout is treated as a flake to eliminate through deterministic synchronization, never a reason to raise the 20-second package timeout.

### CodeQL follow-up: bounded capacity and filesystem arithmetic

CodeQL reports two alerts newly introduced by this slice. The Postgres Arrow range compiler receives a valid immutable `TransformPlanV1`: connection creation admits only a regular `--transform-file` of at most 64 KiB (`internal/cli/transform_plan.go`), the closed parser rejects invalid/unknown syntax, and the native package cannot construct its unexported representation. Thus its source-field count is bounded by that 64 KiB input, far below integer overflow; nevertheless a map's capacity is only an allocation hint, so the repair removes `len(fields)+2` rather than depending on the bound or adding a validation subsystem. The filesystem probe removes its impossible unsigned `Bavail < 0` branch while retaining the signed block-size rejection and the existing multiplication-overflow clamp. Neither change affects production fast-path lifecycle or benchmark values.

## Red / Green delivery slices

1. **R1 regression characterization:** Add a two-plus-page real PostgreSQL `full_overwrite` test through production construction. Its exact published content assertion must fail if current truncate/publish occurs per page. Record the red result; if confirmed, change only the run-scoped lifecycle to one shadow target and one publish, then rerun Green without weakening the assertion.
2. **R2 neutral contracts:** Write failing happy/bad/edge tests for a typed, connector-neutral extraction batch, byte-credit admission, versioned segment manifest, and bulk-applier receipt sequence. Implement the smallest ports with bounded resources and no PostgreSQL/pgx/SQL leakage.
3. **R3 closed transform:** Write failing happy/bad/edge tests for deterministic normalized `TransformPlanV1`, plan hash persistence/binding, pre-I/O validation, and the connection-create `--transform-file` surface. Implement the closed language and all applicable CLI help/manual/website/generated-doc parity.
4. **R4 PostgreSQL adapters:** Write failing live/container tests for range extraction, Arrow/DuckDB transformation, Parquet segment durability, binary COPY to logged shadows, one publish/receipt transaction, and checkpoint-after-reconciliation. Implement through the neutral ports, preserving old interfaces and modes.
5. **R5 durability/measurement:** Write failing success/failure/replay/credit/cancellation tests that observe durable phase counters, no run deadline, per-unit deadlines, receipt-before-checkpoint ordering, and exact typed refusals before I/O. Implement payload-free phase timing/counters and machine-readable MB/s/MiB/s output.
6. **R6 proof harness:** Add the opt-in two-container binary correctness test and the tagged 5 GB external-host harness (≤25.0 s at the 200 MB/s gate). Include transformed-versus-identity reporting, measured logical-byte preflight, peak-disk reporting, a machine-readable report before cleanup, and an explicit skip when the opt-in is absent. The harness refuses to start below 3 GiB free disk; the operator reclaims only enumerated dangling Docker images before invoking it. The source stays present for the identity/realistic comparison and the harness container cleanup then removes it. Add MySQL only if its extractor implements the existing neutral port with no substrate change.

## Non-goals

No object-store source, generic SQL or HTTP writes, non-overwrite fast modes, CDC optimization, unlogged staging, custom pgwire encoder, automatic PostgreSQL settings changes, legacy local-warehouse table/WAL changes, target-table adoption, or petabyte wall-clock claim.

## Checkpoints

- Commit/push the planning evidence.
- Commit/push each red test checkpoint when useful, then each coherent Green slice.
- Run targeted package tests after every Green slice; record commands/results in `TDD-LEDGER.md` and `VERIFICATION.md`.
- Before PR: run scoped test packages and `internal/cli` separately with `-timeout 20m`, vet, build, and individual verify gates; run the tagged DB test only with explicit safe opt-in; complete GSD `verify-work`, then code review.

## CLI parity checklist

- [x] `pm connections create --help` and `pm help connections` document `--transform-file` and its closed, no-arbitrary-SQL contract.
- [x] Bare namespace behavior remains available; invalid-action behavior is unchanged because this slice adds a create flag, not a namespace/action.
- [x] `docs/cli/**`, website docs, generated help/manual/index artifacts, and transcript tests match the runtime help.
- [ ] PR body records runtime help, docs/website grep/generator evidence, and any true not-applicable surface.
