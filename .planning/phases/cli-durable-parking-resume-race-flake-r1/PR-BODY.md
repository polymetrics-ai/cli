## Intent

Fix the durable parking `resume-race` process flake without masking it. This
PR targets `integration/4015-mvp-flat-r1`; its branch is
`fm/cli-durable-parking-resume-race-flake-r1`.

## Causal race and correction

Two concurrent short-lived CLI processes each call `app.Open`, which starts a
durable parking coordinator and attempts to claim a due parked run. Before
this change, the losing process could fail **before `Claim`**: `app.Open`
called `JSONStore.Load`, which tried to create the legacy `state.json.lock`
`O_EXCL` writer marker while the winning resume process held it for a state
mutation. The loser exited with `state.json.lock: file exists` and never
participated in the durable claim.

`state.json` writers already atomically replace a complete file. `app.Open`
now reads that coherent current snapshot without taking the legacy writer
marker; state mutations remain protected by the existing exclusive marker and
the durable parking store retains its advisory-lock-backed single-winner claim.
This is not a retry, skip, quarantine, timeout increase, or concurrency
reduction.

## Test contract

- Happy: `TestOpenReadsCurrentStateWhileLegacyStateLockIsHeld` proves the real
  app constructor returns the observed state revision while a writer owns the
  legacy marker.
- Bad: `TestOpenRejectsMalformedStateBeforeCreatingDurableParkingStore` asserts
  a wrapped `*json.SyntaxError` and no durable parking-store creation.
- Edge: `TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess/edge:
  concurrent durable reopeners share one resume claim` exercises the shipped
  `cli.Run` → `app.Open` path twice, asserts both children succeed, exactly one
  provider resume send, and no replay send.

## Verification

- Red first: the happy-path app test failed before implementation with
  `state.json.lock: file exists`.
- 20 consecutive full-concurrency process passes:
  `go test -timeout 20m ./internal/cli -run '^TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess$' -count=20 -v`
  — passed in `266.034s`; host load `{ 24.73 24.67 21.65 }` before and
  `{ 44.73 33.22 26.05 }` after.
- `go test -timeout 20m ./internal/app` — pass (`232.795s`).
- `go test -timeout 20m ./internal/cli` — pass (`442.083s`).
- `go vet ./internal/app ./internal/cli`, `go build ./cmd/pm`, and individual
  `tidy-check`, `docs-check`, `smoke-no-build`, `agent-contract-check`,
  `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`,
  `release-workflow-check`, and `lint` — pass.

## Delivery record

- GSD lifecycle: resolved and executed inline via `scripts/gsd prompt` for
  `discuss-phase`, `plan-phase --tdd --skip-research`, `execute-phase`,
  `verify-work`, and `code-review`. Inline/manual fallback applies because this
  single-worker task has no compatible Pi isolated-role runtime and the
  canonical contract prohibits role spawning.
- Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and
  `golang-concurrency`.
- CLI help/manual/website parity: not applicable; no public surface changed.
- Automated review route: `claude_auto` on PR open for the reviewed commit
  range; no Copilot fallback requested. Await and disposition any review input
  before the human-gated merge.
