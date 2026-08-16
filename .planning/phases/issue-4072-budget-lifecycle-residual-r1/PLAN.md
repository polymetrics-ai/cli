---
phase: issue-4072-budget-lifecycle-residual-r1
plan: "01"
type: tdd
status: in_progress
base: integration/4015-mvp-flat-r1
requirements:
  - ISSUE-4072-RESIDUAL
required_skills:
  - golang-how-to
  - golang-design-patterns
  - golang-structs-interfaces
  - golang-error-handling
  - golang-security
  - golang-safety
  - golang-testing
  - golang-context
  - golang-concurrency
  - gsd-discuss-phase
  - gsd-plan-phase
  - gsd-execute-phase
  - gsd-verify-work
  - gsd-code-review
files_modified:
  - internal/connectors/connectors.go
  - internal/connectors/connsdk/rate_limits.go
  - internal/connectors/engine/hooks.go
  - internal/connectors/engine/rate_limit_runtime.go
  - internal/connectors/engine/read.go
  - internal/connectors/hooks/github/hooks_test.go
---

# TDD plan: #4072 budget lifecycle residual

## Task Delivery Header

- Issue: Refs #4072 — gate GitHub App token minting through the budget lifecycle.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, with the API-reported base verified after opening and recorded verification green.
- Working branch: `fm/cli-4072-budget-lifecycle-residual-r1`
- Task: Attach the existing secret-free `BudgetCoordinator` to the engine-owned declared token-send boundary. A granted GitHub App token request must call `Decide` once, send once, and call `Finish` once. A coordinator refusal must call `Decide` once, make no send, and call no `Finish`; it must return a typed budget-refusal error.
- Verification: TDD red/green focused GitHub hooks test; engine/hooks, `internal/cli`, and `cmd/connectorgen` tests; `go vet`; build; all independent repository verification gates; docs generation twice for byte stability; diff and secret-boundary review; GitHub API pull-request base read-back.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Granted physical token mint performs one lifecycle pair | fake | A deterministic `BudgetCoordinator` fake is necessary to count exact calls safely. It asserts `decide=1`, `finish=1`, `send=1`, with a generated test-only key and no retained auth material. |
| Refused decision has deliberate completion semantics | fake | The same fake returns `Granted=false` and asserts `decide=1`, `finish=0`, `send=0`, plus the typed pre-I/O budget refusal. No lease exists to finish. |
| Required shared coordination still refuses pre-I/O | fake | Existing recording transport asserts `physical_token_mints=0` and `SharedRateLimitUnavailableError` for missing and unreachable coordination without provider credentials. |
| Existing declared route and no-retry behavior survives | fake | Existing hook tests assert the declared POST route and exactly one physical POST on a 500 fixture response. |

## Execution steps

1. **Red:** Add counter-based grant and decision-refusal cases to the existing
   GitHub App token-mint test file. Run them before any production lifecycle
   change; the grant must fail because the coordinator is not consumed.
2. **Green:** Add the smallest Runtime-configured, engine-owned adapter that
   derives a reservation batch from the already resolved declaration, calls
   `Decide` immediately before the token send, and calls `Finish` exactly once
   after an admitted attempt. Preserve the narrow auth-hook contract, route
   resolution, shared admission, and disabled retries.
3. **Regression:** Run the focused hook/engine tests and all mandatory local
   verification entry points. Record commands and results in
   `TDD-LEDGER.md` and `VERIFICATION.md`.
4. **Review:** Perform the inline code-review fallback over changed source and
   tests, checking error paths, lease completion, context propagation, and
   secret boundaries. Write `REVIEW.md` before requesting remote review.

## GSD execution record

The following command prompts were resolved and executed inline because the
named issue residual is not a numeric roadmap phase and this lane forbids
spawning: `discuss-phase --auto`, `plan-phase --tdd --skip-research`,
`execute-phase --interactive`, `verify-work --auto`, and `code-review` with
the changed engine/hook files. `scripts/gsd doctor`, all five `sources`
lookups, and `go run ./cmd/agentcontractgen check` passed before planning.
