# Live-path defects r1: route admission, bounded windows, credential errors

## Task Delivery Header

- Issues: Refs #4119, #4125, #4169 (one Firstmate-authorized corrective PR)
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: direct PR against `integration/4015-mvp-flat-r1`; after opening,
  read the API-reported base as required by the delivery header.
- Working branch: `fm/cli-live-path-defects-r1`
- Exclusion: do not touch #4158 or its managed PostgreSQL control assertion.
- No secrets, credentialed provider calls, provider writes, generic HTTP-write,
  generic SQL-write, or new dependencies are authorized.

## Lifecycle and skills

- Completed before planning: `scripts/gsd doctor`, `scripts/gsd sources` for
  `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
  `code-review`; `go run ./cmd/agentcontractgen check`.
- Generated prompts are executed inline because this run has no compatible
  isolated GSD worker and the task forbids role spawning. This is the documented
  manual-GSD fallback.
- Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, and `golang-structs-interfaces`.

## Slices and ownership

| Commit | Issue | Production scope | Acceptance proof |
| --- | --- | --- | --- |
| 1 | #4119 | requester/rate-admission route selection and its focused tests | A redirect into a shared route is charged/refused against the destination; local-only redirects still send; the typed refusal keeps its route/reason through formatters. |
| 2 | #4125 | `internal/coordination/shared_rate_limits.go` and focused tests | zero, negative, maximum, one-past-maximum, and duration-overflow inputs have defined typed results before cache or coordinator I/O. |
| 3 | #4169 | CLI/provider error classification and real construction-path tests | provider 401 yields a credential-class error with no secret value, while an internal failure remains `internal_error`; rejection performs no provider write or checkpoint advance. |

## TDD plan

Each behavior begins with separately named tests for the Firstmate test
contract. Red evidence is captured in `TDD-LEDGER.md` before production edits;
green evidence is recorded after implementation.

1. **#4119 route admission.** Add the happy redirect-to-shared-policy test,
   bad non-canonical/unresolved-route refusal-before-send test, and edge tests
   for local-only redirect and base-prefixed canonical route. The tests assert
   charged policy/refusal reason and provider-double send count, not merely an
   error.
2. **#4125 window bounds.** Add happy maximum-accepted-window admission,
   bad negative/one-past-limit typed rejections before cache I/O, and edge
   cases for zero and an integer that would overflow `time.Duration`. Tests
   assert the typed error and the exact cache-call count.
3. **#4169 credential classification.** Add happy provider-401 credential
   classification, bad genuine-internal-failure preservation, and edge tests
   for no credential leakage and no write/checkpoint advancement. Exercise the
   real CLI construction path where the production classification boundary is
   reachable.

## CLI help/manual/website parity

This changes neither command topology, flags, help text, output schema, nor
documented connector surface. The human-facing classification changes an error
category only. Runtime help/manual/website generated artifacts are therefore
not applicable; focused binary/CLI classification tests and existing help
smoke checks are retained in verification.

## Verification and checkpoints

1. Commit and push this plan checkpoint.
2. For each slice: write and run red tests; implement only that slice; run
   focused green tests; commit and push the green slice.
3. Run focused package tests plus `internal/cli`, `go vet`, build, generated
   connector checks where applicable, and the non-full-suite verification gates
   required by AGENTS.md. Do not run `no-mistakes`.
4. Execute `verify-work` and `code-review` inline, record dispositions, push,
   open the direct PR, and API-verify its base branch.
