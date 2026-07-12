# TDD Ledger

Phase: 325-agentloop-characterization

## Planning checkpoint

- Production edits: none.
- Red evidence: not yet run; no behavior task is marked `red-confirmed`.
- GSD adapter: manual fallback recorded because `scripts/gsd prompt programming-loop ...` is not
  registered; installed helper plus universal runtime loop is active.
- Next action completed: all tests and thirteen sanitized fixture files were added before any
  production implementation.

## Red checkpoint — 2026-07-12

### fixture-replay

- Status: red-confirmed
- Command: `go test ./internal/agentloop/... -count=1`
- Exit: 1
- Expected evidence: package build fails on undefined `LoadFixture`, `Replay`, `LoadFixtures`,
  `ReplayAll`, and related Phase 0 types. All thirteen strict JSON fixture files parse with `jq`.

### structural-redaction

- Status: red-confirmed
- Command: `go test ./internal/agentloop/... -count=1`
- Exit: 1
- Expected evidence: tests referencing `ValidateFixture` and `ValidationError` cannot compile
  because no scanner/validator implementation exists. The sensitive canary is assembled only in
  test memory and is not written to a fixture.

### safety-policy

- Status: red-confirmed
- Commands: `go test ./internal/agentloop/... -count=1` and
  `bash scripts/tests/auto-loop-control.sh`
- Exit: 1 for each
- Expected evidence: Go safety functions are undefined; shell reports the safety helper missing
  and every run/resume path lacks `AUTO_LOOP_DISABLED_PHASE_0`.

### loopctl-cli

- Status: red-confirmed
- Command: `go test ./cmd/loopctl/... -count=1`
- Exit: 1
- Expected evidence: six call sites fail on undefined `run`; no command implementation exists.

### driver-fuse

- Status: red-confirmed
- Command: `bash scripts/tests/auto-loop-control.sh`
- Exit: 1
- Expected evidence: both drivers invoke harmless marker binaries on run and resume; both persist
  state before denial; both treat `--help` as work and invoke the marker. No external tool ran.

### entrypoint-enumeration

- Status: red-confirmed
- Command: `bash scripts/tests/auto-loop-control.sh`
- Exit: 1
- Expected evidence: canonical safety helper/inventory is absent and the wrappers are not marked.

### make-gate

- Status: red-confirmed
- Command: `bash scripts/tests/auto-loop-control.sh`
- Exit: 1
- Expected evidence: `agent-loop-test` is absent and `verify` does not include it.

## Red integrity

- Production files changed before these failures: none.
- External effects: none. Driver characterization ran copies under temporary directories with
  inert local binaries that wrote only a marker; temporary directories were removed.
- Test weakening: none.
- Implementation authorization: the strict TDD gate may pass only after reading these exact task
  IDs and `red-confirmed` statuses.
