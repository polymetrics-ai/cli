# TDD Ledger — durable parking resume claim race

## Red

- `go test -timeout 20m ./internal/app -run
  '^TestOpenReadsCurrentStateWhileLegacyStateLockIsHeld$' -count=1 -v` failed:
  a second `app.Open` returned `lock state: create lock file: .../state.json.lock:
  file exists`. The reader had not reached `RateParkingCoordinator.Start`, so
  it could not participate in the durable resume claim.
- The pre-existing full process case had measured 6 passes / 1 failure on the
  shared integration base under load. Its retained child-process diagnostic
  reports the failing exit code and output; no retry, skip, quarantine,
  concurrency reduction, or timeout change was used in this repair.

## Green

- `TestOpenReadsCurrentStateWhileLegacyStateLockIsHeld` (**happy path**) proves
  `app.Open` reads the actual current revision while a concurrent legacy writer
  owns `state.json.lock`; state mutations remain covered by
  `TestNewStateStoreMutationsHonorLegacyStateLock`.
- `TestOpenRejectsMalformedStateBeforeCreatingDurableParkingStore` (**bad
  path**) asserts a wrapped `*json.SyntaxError` and proves the refusal happens
  before the durable parking store is created.
- `TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess/edge: concurrent
  durable reopeners share one resume claim` (**edge case**) drives the shipped
  child CLI dispatcher and `app.Open` twice against one parked checkpoint. It
  asserts both helpers succeed, exactly one provider resume send occurs, and
  replay emits no extra send.
- `go test -timeout 20m ./internal/cli -run
  '^TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess$' -count=20 -v`
  passed 20 consecutive runs in `266.034s`. Host load was `{ 24.73 24.67
  21.65 }` before and `{ 44.73 33.22 26.05 }` after the run.
- Focused app tests passed after the correction:
  `TestOpenReadsCurrentStateWhileLegacyStateLockIsHeld`,
  `TestOpenRejectsMalformedStateBeforeCreatingDurableParkingStore`, and
  `TestNewStateStoreMutationsHonorLegacyStateLock`.

## Refactor

- `app.Open` now takes a lock-free coherent snapshot of atomically replaced
  current project state. The legacy `O_EXCL` marker still serializes every
  state mutation. This is a state-transition correction, not a retry: the
  durable store's atomic cross-process claim remains the single-winner gate.
