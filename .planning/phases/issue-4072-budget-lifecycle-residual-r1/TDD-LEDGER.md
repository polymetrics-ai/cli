# Issue #4072 residual TDD ledger: budget lifecycle

**Base:** `ff6a8710199c10f209d9d47cce87e5c8f7c429e6` on
`integration/4015-mvp-flat-r1`.

**Required skills used:** `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
`gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
`gsd-verify-work`, `gsd-code-review`.

| ID | Truth | Observable assertion | Status |
| --- | --- | --- | --- |
| L1 | A granted token request reserves and completes exactly once | `Decide=1`, `Finish=1`, and provider `send=1` | planned |
| L2 | A refused decision does not have a completion lease | `Decide=1`, `Finish=0`, provider `send=0`, typed refusal | planned |
| L3 | Missing/unreachable `require_shared` remains prior to token I/O | existing typed unavailable errors and `physical_token_mints=0` | planned regression |
| L4 | Token mint still never retries | 500 fixture observes exactly one POST | planned regression |
| L5 | Lifecycle data remains secret-free | fake sees only batch policy fingerprint, opaque scope, declared budgets, and observation; it does not retain request/header/body material | planned |

## RED

- Red: `go test -timeout 20m ./internal/connectors/hooks/github -run '^TestGitHubAppAuthBudgetLifecycle' -count=1`
- Observed on the pre-adapter path: the grant case made one token request with
  `Decide=0`; the refusal case returned nil instead of a typed refusal. This
  is causal evidence that `RuntimeConfig.BudgetCoordinator` had no production
  consumer at the declared token-send boundary.
- Result: expected failure (2026-08-17).

## GREEN

- Pending: wire the engine adapter and rerun the same focused test. Record the
  exact command, result, and commit after it is green.
