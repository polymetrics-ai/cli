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

## Adversarial gap cycle — strengthened red

Read-only review rejected the first characterization as too conclusion-shaped and found shell
sandbox gaps. The uncommitted partial implementation was deleted before this cycle; only tests,
fixtures, and phase artifacts remained when the following failures were captured.

### fixture-replay

- Status: red-confirmed
- Command: `go test ./internal/agentloop/... -count=1`
- Exit: 1
- Expected evidence: the strengthened suite cannot compile because fact-based `Fixture`, `Fact`,
  `LoadFixture`, and `Replay` do not exist.
- Strengthening: neutral resource/owner/before/after facts; observed decision/outcome correctness
  separated from required policy; fact/order/actor/binding mutations for all thirteen incidents;
  distinct synthetic identities for the three correct HALTs; phantom-dispatch and missing-handoff
  reasons; missing-artifact plus later repo-gate failure.
- Security negatives added: oversize file/string/event/count, symlink file/entry, duplicate incident
  IDs, missing required correctness booleans, `.jsonl` pre-open rejection, and non-echoing errors.

### structural-redaction

- Status: red-confirmed
- Command: `go test ./internal/agentloop/... -count=1`
- Exit: 1
- Expected evidence: required presence-aware booleans, structural fact validation, and generic
  non-echoing validation/I/O errors are absent.

### safety-policy

- Status: red-confirmed
- Commands: `go test ./internal/agentloop/... -count=1` and
  `bash scripts/tests/auto-loop-control.sh`
- Exit: 1 for each
- Expected evidence: immutable Go/shell safety policy remains absent; direct enable/open/run/resume
  commands cannot yet be rejected by a policy implementation.

### loopctl-cli

- Status: red-confirmed
- Command: `go test ./cmd/loopctl/... -count=1`
- Exit: 1
- Expected evidence: eight `run` call sites are undefined, including negative enable-like commands
  and a caller-path non-echo test.

### driver-fuse

- Status: red-confirmed
- Command: `bash scripts/tests/auto-loop-control.sh`
- Exit: 1 with 39 expected failures
- Expected evidence: under `env -i`, stub-only PATH, isolated HOME/config, non-writable state, and
  unreadable resume input, both drivers invoke pre-guard tools and fail to deny run/resume,
  enable-like environment, `--enable`, and `--force`; help also invokes tools.

### entrypoint-enumeration

- Status: red-confirmed
- Command: `bash scripts/tests/auto-loop-control.sh`
- Exit: 1
- Expected evidence: independently discovered filename+semantic candidates have no canonical
  helper/inventory/guard parity yet.

### make-gate

- Status: red-confirmed
- Command: `bash scripts/tests/auto-loop-control.sh`
- Exit: 1
- Expected evidence: the phase target and verify integration remain absent.

No partial implementation from the rejected first green attempt was present during this gap-red
capture. Implementation may resume only from this strengthened contract.

## Implementation and adversarial gap reds

The standard-library implementation first reached green, then every blocking review finding was
converted to a focused failing test before its correction:

- cross-identity/ambiguous fact composition and invalid synthetic ID grammar failed before
  correlation and grammar enforcement;
- unknown CLI commands/flags echoed runtime canaries before generic diagnostics;
- leading wrong-scope decoys suppressed valid ordered/special rules before bounded tuple search;
- arbitrary Expected values and incident IDs reached replay output before closed enums and
  incident-to-derived-policy binding;
- historical truth assertions failed for the non-durable turn-23 HALT, fail-open dead workers,
  same-head merge-state race, full S3 wait, and turn-26 terminal ledger divergence;
- final human-ready facts matched an over-broad wait rule before blocked-stage correlation;
- wrong-resource stale-head and ratification mutations remained classified before resource
  identity enforcement.

Each gap command exited 1 for the asserted missing behavior, then the corresponding focused test
and full `internal/agentloop` suite exited 0 after the smallest correction. Tests were not weakened;
obsolete fixture claims were replaced with source-grounded facts.

## Green checkpoint — 2026-07-12

- `go test ./internal/agentloop/... -count=1` -> exit 0.
- `go test -race ./internal/agentloop/... -count=1` -> exit 0.
- `go test ./cmd/loopctl/... -count=1` -> exit 0.
- `bash scripts/tests/auto-loop-control.sh` -> exit 0, `auto-loop-control: ok`.
- `make agent-loop-test` -> exit 0.
- Independent adversarial review -> APPROVE, no remaining P0/P1.

## Refactor and broad verification checkpoint

- `go version && go mod verify` -> Go 1.25.12; all modules verified.
- `bash -n` over both drivers, helper, and harness plus `jq empty` over all fixtures -> exit 0.
- `git diff --check` and issue write-scope audit -> exit 0.
- Uninterrupted `make verify` after the final review fix -> exit 0, including full Go tests,
  build, docs, smoke, lint (0 issues), connectorgen validation (547 connectors, 0 findings), and
  the new Phase 0 target.

All seven behavior tasks and the phase-memory task are complete. No partial or interrupted broad
gate is claimed as verification evidence.
