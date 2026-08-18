# Verification — durable parking resume claim race

## Goal-backward result

**Passed.** A concurrent CLI opener no longer fails on the project-state
writer marker before it reaches the durable rate-parking coordinator. Atomic
state reads are coherent with atomic replacement; state mutations retain the
legacy exclusive marker; durable parking retains its advisory-lock claim as
the single resume winner.

## Production-path evidence

- Red: `TestOpenReadsCurrentStateWhileLegacyStateLockIsHeld` originally failed
  with `state.json.lock: file exists`, establishing that the loser failed before
  `RateParkingCoordinator.Start` and therefore before `Claim`.
- Green: that test now opens the real `App` and asserts the loaded current
  revision while a concurrent writer owns the marker.
- Bad path: `TestOpenRejectsMalformedStateBeforeCreatingDurableParkingStore`
  asserts `*json.SyntaxError` and verifies that no rate-parking store exists.
- Edge path: the real child-process CLI test invokes `cli.Run` and `app.Open`
  twice. Its two helpers both succeed, its provider observes exactly one resume
  send, and its replay check observes no second send.

## Commands and results

- `go test -timeout 20m ./internal/app -run '^(TestOpenReadsCurrentStateWhileLegacyStateLockIsHeld|TestOpenRejectsMalformedStateBeforeCreatingDurableParkingStore|TestNewStateStoreMutationsHonorLegacyStateLock)$' -count=1 -v` — pass.
- `go test -timeout 20m ./internal/cli -run '^TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess$' -count=20 -v` — 20 consecutive passes in `266.034s`; load `{ 24.73 24.67 21.65 }` before and `{ 44.73 33.22 26.05 }` after.
- `go test -timeout 20m ./internal/app` — pass (`232.795s`).
- `go test -timeout 20m ./internal/cli` — pass (`442.083s`).
- `go vet ./internal/app ./internal/cli` and `go build ./cmd/pm` — pass.
- Individual gates `tidy-check`, `docs-check`, `smoke-no-build`,
  `agent-contract-check`, `connectorgen-validate`,
  `connectorgen-surface-sync`, `connector-boundary`,
  `release-workflow-check`, and `lint` — pass.

## GSD lifecycle

Resolved and executed inline: `discuss-phase`, `plan-phase --tdd
--skip-research`, `execute-phase`, `verify-work`, and `code-review` through
`scripts/gsd prompt`. The non-Pi worker has no compatible isolated workflow
runtime and this repository's single-worker contract prohibits role spawning;
the lifecycle, TDD evidence, verification, and review were performed inline.

## CLI parity

Not applicable: no public command, flag, output, help, manual, connector
surface, or website behavior changed.
