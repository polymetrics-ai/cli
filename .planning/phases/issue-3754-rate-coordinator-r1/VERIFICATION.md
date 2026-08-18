# Verification — issue #3754

Status: passed by automated and live-coordinator evidence. Manual GSD fallback is recorded in
`PLAN.md`: phase 3754 is not registered numerically, so the generated workflow prompts were
performed inline without spawning incompatible roles.

- [x] Local scope registry applies a real shared-in-process budget and human/JSON inspection labels it `process-local`.
- [x] `require_shared` engages the external registry when available.
- [x] `require_shared` returns a typed no-fallback `coordinator_not_configured` reason when unavailable.
- [x] Opaque coordination identity is absent from argv, environment, files, logs, receipts, and evidence: API types admit only the #3863 `RateLimitScopeKey`; the key test derives a real projection and rejects subject/binding canaries; CLI/error tests expose no scope; helper processes receive a credential-free allowlist; no persistence, logging, or receipt code changed.
- [x] Two separate processes under one opaque scope obey the external budget.
- [x] Context cancellation, atomicity, server-time TTL/reset, and all supported declared models are covered.
- [x] No connector-specific production literal/branch, no production bundle edit, no parking/resumption, and no generic execution surface.
- [x] Focused package tests, race test, targeted vet/build, formatting, CLI/help/docs/website parity, and individual repository gates pass.
- [x] CI remediation regenerated the sole stale website artifact after the
  rate-limit provenance documentation change; local website lint and typecheck
  pass. The third-party Snyk status is an opaque `1 test has failed` baseline:
  the live integration head and declared comparison base have the identical
  failure, while this branch has no dependency-manifest delta. The repository
  Security workflow and `govulncheck` passed.

## Evidence

- CLI disclosure: `./pm connectors inspect github | rg -A 2 '^RATE LIMIT COORDINATION$'` printed the exact process-local boundary.
- Typed refusal: `TestSharedRateLimitRegistryRefusesWhenCoordinatorIsMissing` and
  `TestRequireSharedRateLimitPolicyRefusesWithoutCoordinator` passed and logged
  `require_shared result=refused reason=coordinator_not_configured`.
- Live shared proof:
  `POLYMETRICS_COORDINATION_INTEGRATION=1 go test -tags=coordinationintegration ./internal/coordination ./internal/connectors/engine -run 'Test(SharedRateLimitCoordinatesSeparateProcesses|SharedRateLimitHonorsEveryDeclaredBudgetModel|RequireSharedRateLimitPolicyUsesAvailableCoordinator)' -count=1 -v`
  passed. The command uses the local Dragonfly service; it proved one grant plus one blocked helper
  process, all four models, and shared requester admission.
- Package checks: `go test -timeout 20m ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine ./internal/cli -count=1`; focused
  `go test -race -timeout 20m ./internal/coordination ./internal/connectors/engine -run 'Test(RateLimit|SharedRateLimit|RequireShared|LocalRateLimit)' -count=1`; targeted `go vet`; `go build ./cmd/pm`; `git diff --check`.
- Repository gates: `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check` all passed.
- CLI parity: `pm help connectors`, bare `pm connectors`, `pm connectors inspect github --help`,
  and both inspect formats passed; updated source help, generated `docs/cli/connectors.md`,
  `docs/migration/conventions.md`, and `website/content/docs/cli-reference.mdx`.
- Issue proof: exact command and verbatim output posted to #3754. No credential was supplied,
  printed, or stored. Provider-specific enforcement remains #3990 because this change deliberately
  does not alter a connector declaration.
- CI repair: `cd website && pnpm run gen:website-data`, `pnpm run lint`, and
  `pnpm run typecheck` passed. The two failed GitHub Actions runs identified only
  `website/lib/docs.generated.ts` as stale.
- Rebase resolution: rebased with preserved merges to the live
  `integration/4015-mvp-flat-r1` head `71da37e`. Retained its approval-carrier
  guard and sync-mode derivation while retaining #3754 registry wiring, then
  passed focused coordination, engine, and CLI tests plus the regenerated-data
  guard.
- Snyk disposition: GitHub exposes only `1 test has failed`; the same status is
  present on `71da37e` and the declared `fbd06e7` base. No dependency manifest
  differs from the live base, and the linked report requires Snyk sign-in, so
  this is recorded as an out-of-range baseline failure rather than masked by an
  unrelated update.

## Correction 3/5 — #4035 (in progress)

- [x] The late `Finish` observation remains readable and affects an admission after three concurrency lease TTLs, while a second admission proves occupancy was released at one TTL.
- [x] Caller cancellation interrupts a stalled UDS exchange, and a syntactically valid response deliberately sent after cancellation cannot be observed as a grant.
- [x] Focused coordinator race/ownership/cleanup and multi-process tiny-budget tests pass.
- [ ] Required `-race` package checks pass; no-mistakes delivery is deferred until firstmate dispatches it with `--skip rebase,pr`.

Evidence: `TestRateBudgetLeaseTTLFreesConcurrencyWithoutDroppingLateObservation` asserts the live admission transition from grant → post-expiry grant → one-minute refusal after the late observation. `TestUnixRateBudgetCoordinatorClientCancellationInterruptsStalledExchange` asserts the cancelled caller returns promptly; `TestUnixRateBudgetCoordinatorClientCancellationWinsResponseRace` asserts a valid ready response released after cancellation is not returned as a grant. `TestUnixRateBudgetCoordinatorMultiProcessTinyBudget` starts one actual UDS owner and eight test subprocesses, observes exactly three grants and five refusals, then asserts its 0700 run directory and 0600 socket are removed on close.

## Correction 5/5 — #4049 (passed local verification)

- [x] A default requester with a config-matching endpoint-sensitive policy fails before a local transport send; a mixed policy fixture also records no rate-budget mutation before the refusal.
- [x] A table test proves all fourteen GitHub WriteHook REST sends call `Runtime.RequesterFor` with an existing declaration, including label, comment/final-state, and PR metadata/reviewer follow-ups.
- [x] The real `create_label` hook under `require_shared` with no coordinator returns an `errors.As`-visible `*coordination.SharedRateLimitUnavailableError` with `coordinator_not_configured` and makes exactly zero local transport sends.
- [x] GitHub `rate_limits.json` is unchanged; `POST /graphql` remains an untouched declaration exclusion.
- [x] Focused engine/GitHub-hook tests, full coordination package matrix, coordination race test, scoped vet/build/format/diff checks, and all individual non-monolithic repository gates passed.

Evidence: `go test -count=1 -timeout 20m ./internal/connectors/engine/... ./internal/connectors/hooks/github/...` passed. `go test -count=1 -timeout 20m ./internal/coordination/...` and `go test -race -count=1 -timeout 20m ./internal/coordination` passed. `go vet ./internal/connectors/engine/... ./internal/connectors/hooks/github/... ./internal/coordination/...`, `go build ./cmd/pm`, and `git diff --check` passed. No provider endpoint, credential, or external coordinator was used.
