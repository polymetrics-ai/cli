# Plan — issue #3754 local and optional shared rate-limit scope registries

Issue: #3754. Parent: #3855. Base/target: `integration/4015-mvp-flat-r1`.

## GSD path

- `scripts/gsd doctor`, `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`, and `go run ./cmd/agentcontractgen check` passed.
- `scripts/gsd prompt discuss-phase 3754 --auto` was generated and executed inline. The
  phase registry has no numerical phase 3754, so the normal Pi phase workflow cannot
  create its numbered artifacts. This directory is the documented manual fallback.
- Generate and execute `plan-phase 3754 --tdd`, `execute-phase 3754`,
  `verify-work 3754`, and scoped `code-review 3754 --files=...` prompts inline. If
  verification finds a gap, generate `plan-phase 3754 --gaps` then execute
  `execute-phase 3754 --gaps-only` before re-verification.

## Required skills loaded

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`,
`golang-context`, `golang-concurrency`, `golang-cli`, and `golang-documentation`.
Runtime planning also used `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.

## TDD slices

| Slice | Red proof | Green implementation | Scope guard |
| --- | --- | --- | --- |
| A — explicit policy | A declaration cannot express `require_shared`; an absent field has no observable process-local provenance | Add closed declaration enum/schema validation and a safe policy-status projection | Default remains local; no production connector declaration changes. |
| B — shared atomic registry | Two independent registry clients can both consume a one-unit scope; unavailable coordinator can be mistaken for a local limiter | Add atomic, server-time/TTL shared admission/observation under `internal/coordination` with typed unavailability/unsupported outcomes | Opaque key only; no durable state, parking, resume, or credential input. |
| C — resolver/wiring | `require_shared` continues through the process-local registry or doesn't reach a configured shared coordinator | Resolve exactly one registry per declared policy; configure optional client invocation-scoped; local status says process-local and shared unavailable fails closed | No global/config inheritance selects shared for a local policy. |
| D — proof/review | No end-to-end test demonstrates separate processes sharing a budget or six-surface absence | Add opt-in real Dragonfly two-process test plus focused fake/typed tests; run binary/provenance and issue-comment proof | No credential, raw environment dump, provider call, or generic execution. |

## Verification plan

- RED then GREEN focused `internal/coordination`, `internal/connectors/connsdk`,
  `internal/connectors/engine`, `internal/config`, and `internal/cli` tests; race run for
  coordination/engine.
- `gofmt`, targeted `go vet`, `go build ./cmd/pm`, `git diff --check`, and the individual
  non-monolithic repository gates named in `AGENTS.md`.
- CLI parity only if a public status/inspect surface changes: `pm help <topic>`, bare
  namespace, exact command help, `docs/cli/**`, website reference, generated docs, and
  tests. Otherwise record each as not applicable.
- Run the two-process test against an explicitly provided local Dragonfly endpoint when
  available. Do not start/stop shared runtime services merely to make the test pass.
- Post the exact safe command and output to issue #3754. If live external coordination
  is unavailable, state that clearly, name the opt-in integration test/follow-up, and
  distinguish it from the unit proof.

## Commit checkpoints

1. Planning/TDD evidence.
2. Focused failing-test (RED) evidence.
3. Green implementation plus local/package verification.
4. Verification and review-artifact checkpoint only.

## CI-repair addendum — 2026-08-14

- The Website Data and Website CI/CD jobs both reproduced a stale generated-data
  RED state after the rate-limit provenance paragraph was added to
  `website/content/docs/cli-reference.mdx`: Actions runs `31809146629` and
  `31809146767` reported only `M website/lib/docs.generated.ts` after
  `pnpm run gen:website-data`.
- GREEN is the repository-prescribed deterministic regeneration of that one
  artifact, followed by the generator check, website lint, and typecheck. It
  changes no coordinator behavior, connector declaration, dependency, or live
  proof comment.
- The GitHub Security workflow and `govulncheck` passed. Snyk subsequently
  reported its opaque `1 test has failed` status, but the current integration
  head and the declared `fbd06e7` comparison base report the same failure and
  this branch has no dependency-manifest delta. The access-controlled report
  exposes no affected in-diff path, so no speculative dependency or security
  change is permitted.
- This CI repair also used `golang-continuous-integration`, `golang-lint`,
  `vercel-react-best-practices`, and `vercel-composition-patterns`. The
  repository routing names `frontend-design` and `web-design-guidelines`, but
  neither is installed in this environment; no React component or design-system
  code is in scope for this generated-data repair.

## Safety and exclusions

- No raw credential, secret-derived equality, scope subject, opaque key, binding, approval
  revision, or provider response is placed in argv, environment, files, logs, receipts,
  tests, issue comments, or planning evidence.
- No generic HTTP, SQL, or shell tool; every connector operation retains the warehouse
  transport boundary.
- No #3865 fence behavior, #3867 parking/resumption, or #3990 GitHub budget declaration.

## Correction 3/5 — #4035 late observations and UDS cancellation

### Task Delivery Header

- Issue: Closes #4035 — fix(coordination): retain late observations and propagate UDS cancellation.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with checks green; the API-reported base is read back after opening.
- Working branch: `fm/cli-4035-late-observation-uds-r1`.
- Task: Preserve valid late completion observations after the short concurrency lease expires while releasing occupancy promptly; bind UDS request/response I/O to caller cancellation and refuse a response that races after cancellation.
- Verification: deterministic coordinator and UDS tests, `go test -race -timeout 20m ./internal/coordination/... ./internal/connectors/engine/...`, the focused multi-process tiny-budget test, and the no-mistakes pipeline after commit.

### Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A late valid `Finish` observation survives beyond the prior retention window | live | A deterministic clock advances three lease TTLs; the expired lease no longer blocks a second admission, then its 429/reset observation makes the next admission return a one-minute wait. |
| Cancellation interrupts a stalled UDS exchange | fake | A local UDS listener is necessary to deterministically hold the response. The client returns `context.Canceled` promptly after request acceptance, proving the blocked read was interrupted. |
| A response released after cancellation is never granted | fake | A local UDS listener reads exactly one request, waits for cancellation, then sends a syntactically valid ready/grant response; the caller still receives `context.Canceled`, not a grant. |
| Shared owner still protects a tiny budget across processes | fake | Test-only helper subprocesses are necessary to prove separate process admission; exactly three grants and five refusals are asserted. |

### GSD/TDD execution record

`scripts/gsd doctor`, all five canonical `sources` resolutions, `go run ./cmd/agentcontractgen check`, and generated prompts for `discuss-phase`, `plan-phase 4035 --tdd`, `execute-phase`, `verify-work`, and `code-review` were run. Issue #4035 is not a numbered roadmap phase and the repository's canonical single-worker contract disallows role spawning, so this append is the inline/manual GSD fallback.

Plan: add the secret-free reservation protocol and a run-local UDS owner/client without changing provider policy, connector declarations, CLI/docs, Dragonfly shared-registry logic, cohort fencing, or parking/resume. The TDD sequence is: write the lease-expiry and cancellation race tests; capture their failing compile/test output; implement minimal in-memory coordinator lifecycle and cancellation-aware UDS exchange; then run the focused race, ownership, cleanup, and multi-process proofs.

Delivery mechanism: when firstmate dispatches no-mistakes, run it with `--skip rebase,pr` because those steps target the repository default `main`. Create this child pull request separately with the explicit base `integration/4015-mvp-flat-r1`, then read the API-reported `.base.ref`; a default-base PR is not acceptable evidence.
