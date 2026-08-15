# Issue #4072 TDD ledger: GitHub App token admission

**Parent:** #3855

**Base:** `7eea99bae` on `integration/4015-mvp-flat-r1` (#4122 / #3754)

**Required skills used:** `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
`gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
`gsd-verify-work`, `gsd-code-review`.

| ID | Truth | Evidence | Result |
|---|---|---|---|
| A1 | Missing shared coordinator does not send a token POST | recording transport and typed error | pass |
| A2 | Unreachable shared coordinator does not send a token POST | real unavailable registry at a refused local endpoint | pass |
| A3 | Shared budget tightens across processes | two child test processes and real local Dragonfly | pass |
| A4 | Admission sees the physical escaped token path | fixture asserts the actual token route | pass |
| A5 | Test evidence does not retain credentials | recorders retain only counts/route, never headers/body | pass |

## RED

- Red: `go test -timeout 20m ./internal/connectors/hooks/github -run '^TestGitHubAppAuthRateAdmissionRequireSharedRefusesBeforeTokenSend$' -count=1`
- Commit: `51c2de835` (`test(4072): prove GitHub App auth admission bypass`).
- Observed failure before the implementation: one physical installation-token
  POST reached recording transport and `NewRuntime` returned nil. This was a
  causal failure, not a compile or fixture failure.

## GREEN

- Green: `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1`
- Commit: `69e48064a` (`fix(engine): admit GitHub App token minting`).
- Observed result: passed. Missing shared coordination returned
  `*coordination.SharedRateLimitUnavailableError` with zero sends; the GitHub
  hook used the engine's declared-route requester.

## Post-green regression

- `d8a4d4353` removed dead compatibility wrappers after the hook capability
  migration; no behavior changed.
- `d1757063e` added the unreachable-coordinator zero-send regression. It
  asserted `SharedRateLimitCoordinatorUnreachable` without passing a secret to
  the coordinator or diagnostic.
- The real-coordinator check passed with one minted token, one budget-exhausted
  child, and exactly one fixture POST.
