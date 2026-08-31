# Plan — issue 4293 whole-cohort validation R2

## Task delivery header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator.
- Base: `fm/cli-top100-declaration-batch-r1` at
  `0a708dea5e0024a173b19959d2c43f2bf5a6e0f2`.
- Delivery: one scoped candidate branch; independent review before any parent
  integration. No PR, main merge, or issue mutation in this task.
- Branch: `codex/4293-whole-cohort-validation-r1`.
- Scope: `cmd/connectorgen` cohort validation and directly related tests plus
  this issue-local evidence only.

## Goal and boundary

The prior cohort command proves the frozen source-lock denominator and the
canonical *path spelling* for the ten matrices, but does not open those
matrices. This slice makes the public `--check` compose the existing
source-backed deferred-visibility matrix reader with validation of the
explicit connector-local artifact-link dialects.

It remains mapping-only. It does not change source locks, source-lane
matrices, connector definitions, descriptor admission, source import,
materialization, runtime execution, transport, credentials, certification, or
the Foundation Atlas. It emits no file and creates no command, operation,
stream, write, transport, or executable declaration.

## Contract

1. Preserve `sourceOperationMappingCohortPathCheck` as the immutable
   denominator helper. `deferred-visibility` continues to use it directly, so
   the new full public check must not introduce recursive validation.
2. `source-operation-mapping-cohort <manifest> --check` first validates the
   frozen primary lock cohort, then opens every owned matrix through the
   existing source/citation/continuation/hidden-row validator.
3. Support only the explicit observed artifact link shapes:
   `source_id` *or* `source_operation_id`, and `lane` *or* `lanes`.
   Unknown, absent, duplicate, or ambiguous forms fail with connector and
   source identity. Provider `record` labels remain opaque.
4. A declared artifact target must be a canonical existing regular file under
   that connector definition; a declared source/lane reference must resolve to
   an already validated matrix row and one of its seven cells. Links are not
   inferred from HTTP methods, schemas, command names, or provider naming.
5. A `mapped_unproven` or `missing_foundation` cell without a declared link is
   retained and reported as the typed
   `deferred_without_declared_artifact_link` projection deficit. It is never
   rejected merely for missing an artifact and never becomes executable.
6. The report states derived primary-source, source-row, cell, deferred,
   explicit-link, and zero-executable counts. The implementation contains no
   connector-specific branches or provider-specific counts.

## Red-green-refactor

1. **Red:** require the public cohort check to publish real matrix/link
   evidence. The unchanged command only reported `10 / 4,341 / 0 findings`.
2. **Green:** compose the two existing generic source readers, add strict
   explicit-link containment/resolution, and add an internal typed deficit
   report for unlinked deferred cells.
3. **Regression:** use a copied immutable ten-connector fixture to prove
   failures for missing and malformed matrices, citation drift, a hidden
   source row, a missing artifact target, an orphan link, ambiguous link
   identity dialect, and an unsafe path. The real cohort remains the positive
   acceptance input.
4. **Refactor/gates:** deterministic sorting, `gofmt`, `git diff --check`,
   focused Go tests, `go vet`, JSON parsing, agent-contract check, and direct
   cohort check as capacity permits. No broad/race suite is required for this
   mapping-only slice.

## Atlas and runtime disposition

The implementation reuses the already present `source.projection-admission.v1`
mapping capability through `deferred-visibility`. No new foundation demand or
runtime behavior was identified; therefore the Atlas is intentionally
unchanged. Any requirement to execute a retained source cell remains a later,
captain-gated foundation/projection task.

## GSD and skills

`scripts/gsd doctor` and the `discuss-phase`, `plan-phase`, `execute-phase`,
`verify-work`, and `code-review` source/prompt paths were resolved. The
repository contract does not permit spawning the role agents in this lane, so
the lifecycle is recorded and performed inline. Loaded guidance: issue-first
delivery, connector-lane build order, `go-engineering`, `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-error-handling`, and
`golang-security`. CodeGraph is absent in this repository.
