# PLAN — Issue 4001: shared connector failure classification contract

## Goal

Provide one typed, dependency-free failure vocabulary that database validation, engine dispatch,
and certification can share without parsing error text or serializing sensitive internal causes.

## Allowed scope

- `internal/failures/**` — the sole contract owner and tests.
- `internal/connectors/configuration_validation*_test.go` — generic native/configuration
  compatibility guard only; no PostgreSQL driver edits.
- `internal/connectors/engine/configuration_validation*.go` — classify existing declarative
  configuration failures with exact JSON-Pointer paths.
- `internal/connectors/commandrunner/runner.go` plus focused test — carry the common dispatch
  classification through an existing blocked-command error without implementing #3991 analysis.
- `internal/connectors/certify/report.go` plus focused test — expose an optional structured
  `untestable_reason` field for certification output.
- `docs/architecture/connector-certification-design.md` — own the certification report contract
  for the optional safe `untestable_reason` extension.
- `.planning/phases/issue-4001-shared-failure-classification-r1/**` — required evidence.

## Explicit exclusions

- No PostgreSQL driver, database execution, write-session, workset, receipt, CDC, API operation,
  provider budget, call-graph analysis, generated baseline, credential, CLI, website work, or
  documentation beyond the certification report contract above.
- No new dependencies and no credentials or provider calls.
- No change to `synccontract.RecoveryOutcome`; it remains a CDC resume/rebootstrap contract.

## Design

`internal/failures.Classification` is an `error` with constructor-validated, closed domain and
dispatch codes. It exposes stable JSON through explicit marshal/unmarshal methods, verifies all
codes and JSON-Pointer paths, defensively copies references, and refuses control characters or
unbounded reference text. `Error` exposes only the safe field path and user-facing message.
`Unwrap` and `Cause` retain the internal Go cause in memory but exclude it from JSON.
`Retryable` returns true only for the `transient` domain.

Engine configuration validation converts its existing format/enum/pattern failures to this typed
form with stable safe messages while retaining detailed diagnostics as internal causes.
`BlockedCommandError` gains optional classification transport so #3991 can attach its concrete
dispatch findings later. Certification's `CapabilityResult.UntestableReason` uses the same type;
a future matrix producer can emit the stable object without another local enum.

## TDD sequence

1. **Red:** Add the contract and consumer tests before the new package or wiring exists. Record the
   focused failing command output.
2. **Green:** Implement the dependency-free classification package and its tests.
3. **Green:** Adapt declarative configuration validation, blocked command transport, and
   certification report serialization; expose only stable safe configuration messages while
   retaining detailed diagnostics as internal causes.
4. **Regression:** Run focused contract, configuration, commandrunner, and certification tests;
   then formatting, vet/build, and the non-suite repository gates required by `AGENTS.md`.
5. **Verify/review:** Complete the issue verification checklist and a manual cross-file review.

## Consumer evidence

- Generic configuration validation models the future database foundation and demonstrates typed,
  non-retryable field-scoped propagation.
- Commandrunner demonstrates engine-dispatch carriage for each closed dispatch kind.
- Certification serializes `untestable_reason` without the private cause.

## Stacked-delivery topology supplement — 2026-08-11

This supplement records delivery ancestry only. It does not alter the original implementation
scope, TDD guarantees, source patches, or the audit record above.

- Current parent base: `origin/docs/4015-connector-release-certification` at
  `5996a8a2a5e99c8aa8eb5a8603ecb1f6bba21f12` (draft parent PR #4016).
- Preserved source: `origin/fm/cli-cert-shared-foundations-r1` at
  `d517756f0f9bc0cc41d90b4cf717325328352a3f` (direct-to-main PR #4013, left open and unchanged).
- Child branch: `feat/4001-stack-shared-certification-failures`, replaying the seven source
  patches in source order onto the current parent base; its child PR must target only
  `docs/4015-connector-release-certification`.
- Delivery scope remains the original shared classification foundation. Provider operations,
  PostgreSQL work, GitHub connector behavior, scheduling, and new taxonomy remain excluded.

The installed GSD command path was resolved for `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review`. This issue foundation is not a numbered
roadmap phase, and the canonical single-worker contract prohibits role spawning in this isolated
worktree, so the generated workflows are executed as a recorded manual-inline fallback. Required
skills for the current replay review are `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-testing`, `golang-lint`, and `golang-documentation`.
