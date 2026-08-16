Refs #4072
Refs #4015

## Intent

Complete only #4072's residual lifecycle requirement. PR #4134's declared
route, shared-admission, no-retry, and Dragonfly behavior is preserved rather
than rebuilt.

## What Changed

- Added the existing secret-free `BudgetCoordinator` as a RuntimeConfig seam
  with an engine default, then consumed it at the narrow declared custom-auth
  request boundary.
- Derived each reservation batch from the declaration-selected policy
  fingerprint, opaque scope, and declared budgets; no raw route, transport,
  request, headers, JWT, private key, or minted token crosses the seam.
- A granted GitHub App installation-token request now has one `Decide`, one
  physical POST, and one `Finish`. A non-grant returns typed
  `RateBudgetRefusalError{Code: reservation_denied}` before I/O, has no lease,
  and deliberately makes zero `Finish` calls.

## Happy, Bad, and Edge Cases

- Happy: `TestGitHubAppAuthBudgetLifecycleGrantFinishesExactlyOnce` counts
  `Decide=1`, `Finish=1`, and `send=1`; it also asserts the completed granted
  opaque lease and an attempted completion observation.
- Bad: `TestGitHubAppAuthBudgetLifecycleRefusalDoesNotFinishOrSend` asserts
  the specific typed budget refusal, `Decide=1`, `Finish=0`, and `send=0`.
- Edge: the completion call uses `context.WithoutCancel` so cancellation that
  arrives after a physical attempt cannot strand a granted in-flight lease.
- Regression: unchanged GitHub auth tests retain missing/unreachable
  `require_shared` typed pre-I/O refusal and the failed-mint single-POST rule.

## TDD and GSD Evidence

- Red commit `694c15d86`: focused counter test failed with `Decide=0` and an
  untyped/non-refusal path, proving the production consumer was absent.
- Green commit `b589b9804`: adapter and counter proof pass.
- Resolved and used inline/manual GSD prompts: `discuss-phase --auto`,
  `plan-phase --tdd --skip-research`, `execute-phase --interactive`,
  `verify-work --auto`, and `code-review`. The named issue residual is not a
  numeric roadmap phase and the lane forbids role spawning; the fallback and
  artifacts are in `.planning/phases/issue-4072-budget-lifecycle-residual-r1/`.
- Skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`,
  `golang-concurrency`, and the five GSD lifecycle skills.

## Testing

- `go test -timeout 20m ./internal/connectors/hooks/github -run '^TestGitHubAppAuthBudgetLifecycle' -count=1`
- `go test -timeout 20m ./internal/connectors/hooks/github -run '^TestGitHubAppAuth' -count=1`
- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go test -timeout 20m ./internal/connectors -count=1`
- `go test -timeout 20m ./internal/connectors/connsdk -count=1`
- `go test -timeout 20m ./internal/cli -count=1`
- `go test -timeout 20m ./cmd/connectorgen -count=1`
- `go test -timeout 20m -race ./internal/connectors/hooks/github -run '^TestGitHubAppAuthBudgetLifecycle' -count=1`
- `go vet ./internal/connectors ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors/hooks/github`
- `go build ./cmd/pm`
- `pnpm --dir website run gen:docs` twice (byte-stable)
- Independent passing gates: `make tidy-check`, `make lint`, `make docs-check`,
  `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, `make release-workflow-check`.

The broad GitHub-auth `-race` run has a disclosed non-gate timing failure in
the unchanged unreachable-coordinator test: its fixed five-second context
expires under race instrumentation. No test was weakened or skipped; the
normal regression suite and focused lifecycle race test pass.

## CLI/help/docs parity

Not applicable: this does not change a CLI surface, help output, command,
flag, manual source, or connector definition. Website docs generation still
ran twice and was byte-stable.

## Safety

No credentials or provider calls were used. Test recorders retain only safe
counts, opaque lifecycle data, and declared policy data.

## Automated Review Coverage

- Target: `integration/4015-mvp-flat-r1` (parent PR #4016 targets `main`).
- Planned primary route: `claude_auto` on this trusted-author, non-draft
  sub-PR; fallback `parent_pr_fallback` only if no sub-PR review record is
  produced. Status: pending until GitHub posts the automatic review.
- No Claude or Copilot findings exist at PR creation; no disposition is due.
