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
  pass. The third-party Snyk status was pending with no reported in-diff finding;
  the repository Security workflow and `govulncheck` passed.

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
