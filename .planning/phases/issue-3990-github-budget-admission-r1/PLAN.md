# Issue #3990: GitHub shared REST and GraphQL budget admission

## Task Delivery Header

- Issue: Refs #3990 — GitHub Certification: enforce REST and GraphQL budgets across the whole run
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green; read the base back through `gh api /repos/polymetrics-ai/cli/pulls/<n> --jq .base.ref` after opening.
- Working branch: `fm/cli-3990-github-budget-admission-r1`
- Task: Add GitHub GraphQL primary/cost policy admission and response observation, require shared coordination for the GitHub certification run, emit bounded provider-budget execution evidence, and prove cross-process enforcement without transmitting or retaining credential values.
- Verification: Focused engine, coordination, and certification tests; the GitHub multi-process tiny-budget proof; definition validation and surface sync; no-mistakes pipeline through green CI.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| GitHub GraphQL query and mutation traffic is admitted before its physical `POST /graphql` send. | live | A provider double records zero requests when admission refuses and one request only after a permitted GraphQL send. |
| GraphQL `cost`, `remaining`, and `resetAt` tighten the declared budget after a response. | live | A response whose cost exceeds the provisional reservation causes the next same-scope request to wait/refuse; a separate scope remains sendable. |
| GitHub certification selects `require_shared` and fails closed if the coordinator is unavailable. | live | The certification preflight returns a typed unavailable error and the provider double records zero sends. |
| REST primary, REST secondary, search/resource, and GraphQL families remain separate policy/budget contracts. | live | Tests select each declared family and assert its own policy ID, scope projection, and resource observation rather than accepting a generic `core` counter. |
| A tiny shared budget enforces across processes. | live | Two independently started worker processes share one opaque scope: the second reports a refusal/wait and its provider-double send counter remains zero after the first consumes capacity. |
| Certification writes structured attempt, wait/reset, and not-sent/deadline evidence without raw scopes or provider secrets. | live | A deterministic certification run serializes the named event kinds, excludes raw scope material, and terminates before a deadline-exceeding wait. |
| Real-binary GitHub sweep leaves its cleanup ledger complete. | fake | Requires real GitHub credentials and provider authorization, neither supplied by this task. The committed proof records the exact authorized command and its required observable outputs for the delivery operator. |

## Foundation Check

| Need | Existing proof | Decision |
| --- | --- | --- |
| Opaque shared rate-limit registry | #3754 merged as #4122; `engine.ConfigureSharedRateLimitRegistry` and `coordination.SharedRateLimiter` are present. | Reuse; do not duplicate. |
| Secret-free scope identity | #3863 is landed; `connectors.CoordinationIdentity.RateScopeKey` derives opaque keys from declared non-secret subjects. | Reuse; tests must not expose raw subjects. |
| GraphQL result parsing | `graphQLOperationResponse` already parses `rateLimit { cost, remaining, resetAt }`. | Extend the admitted requester observation seam rather than add a second parser. |
| Certification executable boundary | `internal/connectors/certify` executes through the real binary/CLI harness and has an append-only cleanup ledger. | Extend its GitHub profile/run evidence; no generic HTTP or raw GraphQL escape hatch. |

## Scope and ownership

- Target connector: `github` only.
- Expected production paths: `internal/connectors/defs/github/rate_limits.json`, `internal/connectors/{connsdk,engine,certify}/**`, and any directly necessary shared coordination seam already supplied by #3754.
- Explicit exclusion: do not change `internal/coordination/shared_rate_limits.go` duration conversion (`#4125`), auth/hooks/spec work owned by #4072, credential behavior, or unrelated connector declarations.
- The certification runner must use provider-declared `account`, `installation`, `endpoint/repository`, and `ip` scope subjects passed only through `CoordinationIdentity`; it must never key or log raw credential material.

## TDD slices

1. **GraphQL declaration and requester observation.** Red tests prove GraphQL currently has no applicable policy and a returned rate-limit body does not affect the next send. Green: declare policy families and provide the typed GraphQL observation path using actual `cost`, `remaining`, and `resetAt`.
2. **Shared certification admission and events.** Red tests prove a GitHub certification-selected requester silently remains local or can send after a missing shared coordinator. Green: select `require_shared`, emit structured `attempt`, `wait`/`reset`, and `not_sent` events, and cut off waits beyond the run deadline.
3. **Process boundary proof.** Red integration test starts independent workers with an isolated registry and demonstrates the second would send; green test uses the shared coordinator and proves second-worker zero sends under a deliberately tiny budget.
4. **Certification/review evidence.** Extend per-resource observation coverage and the binary-sweep ledger contract. The live credentialed sweep remains an operator-authorized evidence command, not an automated credentialed test.

## Lifecycle record

- `scripts/gsd doctor`, `sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` completed before planning.
- Generated GSD prompts were executed inline: this runtime has no compatible isolated GSD worktree agent and the task contract prohibits role spawning. This is a manual-GSD fallback, not a lifecycle waiver.
- Loaded skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-graphql`, and `no-mistakes`.

## Commit checkpoints

1. Plan and red test evidence.
2. GraphQL admission/observation green slice.
3. Shared-certification/process-proof green slice.
4. Verification and review fixes.
