---
phase: issue-4072-github-app-auth-admission-r1
plan: "01"
type: tdd
status: complete
base: integration/4015-mvp-flat-r1
requirements:
  - ISSUE-4072
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
  - internal/connectors/engine/auth.go
  - internal/connectors/engine/hooks.go
  - internal/connectors/engine/read.go
  - internal/connectors/hooks/github/hooks.go
  - internal/connectors/hooks/github/hooks_test.go
  - internal/connectors/hooks/github/github_app_rate_admission_integration_test.go
---

# TDD plan: GitHub App token admission

## Task Delivery Header

- Issue: Refs #4072 — gate GitHub App token minting through shared rate admission.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green; the pull-request base will be read back through `gh-axi api`.
- Working branch: `fm/cli-4072-github-app-rate-admission-r1`
- Task: Ensure the GitHub App installation-token POST enters the shared rate-admission path, rejects unavailable required coordination before sending, and preserves #3754 resolved physical-path matching and typed errors.
- Verification: Focused and full engine/GitHub tests, race detector, vet, build, required repository gates, a real two-process Dragonfly budget proof, no-mistakes pipeline, and GitHub API PR-base read-back.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| GitHub App minting shares the admission path | live | Two independently launched processes contend for one real Dragonfly budget; exactly one token mints, one waits to deadline, and exactly one fixture token POST occurs. |
| Unavailable required coordination cannot send a token | fake | A recording transport is required to safely prove zero provider sends without real GitHub credentials; it records zero sends while the typed unavailable error is recoverable for both missing and refused local coordinator endpoints. |
| The PR targets the integration branch | live | After opening, `gh-axi api /repos/polymetrics-ai/cli/pulls/<n> --jq .base.ref` must return exactly `integration/4015-mvp-flat-r1`. |

## Objective

Route GitHub App's installation-token POST through the landed #3754 requester
admission boundary. A `require_shared` policy must reject a missing or
unreachable coordinator before a token request reaches transport.

## Design

1. Construct the base requester and `rateLimitResolver` before selecting
   custom authentication.
2. Give only `DeclaredRouteAuthHook` an engine-owned `DoJSON` capability. It
   resolves the declared route to a configured requester and sends the hook's
   escaped physical path through `Requester.Do`.
3. Have GitHub App authentication use that capability for
   `POST /app/installations/{installation_id}/access_tokens`; remove its direct
   HTTP call.
4. Preserve #3754's actual-path matching, typed unavailable errors, and the
   accepted pre-redirect boundary.

## TDD acceptance

- Red: a real GitHub bundle with `app-installation=require_shared` makes a raw
  mint reach recording transport before refusal on the recovered implementation.
- Green: missing and unreachable coordinators return
  `*coordination.SharedRateLimitUnavailableError` and record zero token sends.
- Live: two separate child test processes share a real Dragonfly fixed-window
  budget of one; exactly one mints, one waits to context deadline, and the
  fixture sees exactly one escaped token route POST.
- Regression: ordinary engine and GitHub hook suites, race detector, vet,
  generators, docs, build, and boundary checks remain green.

## Execution record

- Red test commit: `51c2de835`.
- Production green commit: `69e48064a`.
- Post-green cleanup: `d8a4d4353`.
- Unreachable-coordinator regression: `d1757063e`.
- Verification and review evidence is in `TDD-LEDGER.md`, `VERIFICATION.md`,
  `UAT.md`, and `REVIEW.md`.

## Manual lifecycle fallback

`gsd-sdk` cannot resolve this non-numeric issue phase and the canonical
delivery contract forbids GSD role spawning in this lane. The lifecycle was
executed inline; Firstmate owns no-mistakes, remote PR, and CI completion.
