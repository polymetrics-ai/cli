## Intent

Refs #4119, #4125, and #4169. This Firstmate-authorized corrective PR targets
`integration/4015-mvp-flat-r1`; it does not merge or target `main`.

## What Changed

- Adds redirect destination-route regression coverage. The integration base
  already performs destination admission at the requester send boundary, so no
  duplicate transport implementation was added.
- Rejects non-positive and too-large shared rate-limit windows with typed errors
  before duration/TTL conversion or coordinator I/O.
- Surfaces provider-verified HTTP 401 as safe `auth/credential_error`, while
  true internal errors remain `internal/internal_error`.

## TDD and Test Contract

- #4119: happy local redirect, bad shared-destination typed refusal with zero
  destination sends, and base-prefixed edge coverage.
- #4125: happy maximum duration/TTL, bad negative/one-past pre-I/O refusals,
  and zero/duration-overflow edge cases.
- #4169: fresh-binary happy path proves the typed envelope plus one read, zero
  writes, and unchanged checkpoint; bad internal-error preservation; edge
  credential-redaction proof.
- Red/green evidence is recorded in
  `.planning/phases/cli-live-path-defects-r1/TDD-LEDGER.md`.

## Lifecycle, Skills, and Verification

- Inline/manual GSD fallback completed: `discuss-phase` → `plan-phase --tdd`
  → `execute-phase` → `verify-work` → `code-review`; generated prompts and
  the fallback reason are recorded in the phase artifacts.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, and `golang-structs-interfaces`.
- Passed focused engine/coordination/connsdk/CLI tests and the fresh-binary
  provider-401 proof, plus `go vet ./...`, `go build ./cmd/pm`, connector
  validate/surface-sync, tidy, lint, docs, smoke, agent contract, connector
  boundary, release workflow, and diff checks.
- CLI help/manual/website parity is not applicable: no command, flag, help
  topic, output schema, manual, website, or generated help artifact changed.
  `pm connectors`, `pm help github`, and `pm github issues list --help` were
  smoke-checked.

## Safety and Review

- No provider credentials or live provider calls were used. Error formatting
  retains no raw provider URL/body or credential value. No provider write,
  generic write tool, or reverse execution was introduced.
- The stale `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip`
  flow-control fixture was resolved upstream by merged PR #4174. After
  confirming the updated base contains `ec1f200c9`, this branch's fresh-binary
  round trip passes locally. See `VERIFICATION.md`.
- `TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess` is an existing
  concurrent-resume proof, outside this PR's behavior changes. At the same
  dispatch base it produced 6 passes and 1 failure across directly observed
  runs while the 1-minute host load was 11.6–16.0 on 12 CPUs. The preserved
  assertion now reports the failed child exit code and sanitized CLI output;
  it neither changes timing nor the underlying durable-parking behavior.
- Inline code review found no actionable issues. Claude automatic review is
  expected to trigger on this PR; no Copilot fallback has been requested.

## Checkpoints

1. `de48c216a` — GSD/TDD plan
2. `2b43c5c9e` — #4119 destination admission proof
3. `f61b0819d` — #4125 bounded shared window fix
4. `cab821f5e` — #4169 typed credential classification
5. `329699f2a` — durable concurrent-resume failure diagnostics
6. `aa3704271` — merged integration base containing #4174
