# Test Plan: Phase 0

## Fixture and replay tests

- Table-drive all thirteen fixture files and expected violation/decision/exit triples.
- Assert fixtures use complete equal bindings at fixture and event level, synthetic identities,
  contiguous sequences, known event kinds, and legacy fail-open versus required fail-closed
  expectations.
- Assert replay is derived from event order and rejects an intentionally mismatched expectation.
- Assert directory replay is filename-sorted and repeated marshaling is byte-identical.
- Reject unknown fields, trailing JSON, incomplete/mismatched/non-synthetic identity, non-monotonic
  sequence, unknown event kind, `.jsonl`, oversize input, raw command/prompt/session path, and a
  sensitive string assembled only in test memory.

## Safety and CLI tests

- Assert closed status is independent of environment and has no mutating method or command.
- Assert only the two marked driver paths are tracked and untracked guard requests fail closed.
- Assert loopctl help/bare behavior, status, entrypoints, guard-driver, replay, malformed fixture,
  and unknown-command exit/stdout/stderr behavior.

## Shell characterization

- Copy drivers and safety helper into a temporary fake repo.
- Use harmless stub executables that only write a marker.
- Red baseline: current run/resume reaches the stubs and writes loop state.
- Green requirement: run and resume exit 78, marker is absent, and no loop state directory is
  created; help exits 0 and also creates no state.
- Compare wrappers marked `AUTO_LOOP_RUN_ENTRYPOINT` to safety inventory so a future wrapper cannot
  silently bypass enumeration.

## Broader gates

- Package tests under race detection.
- `make agent-loop-test` aggregates the exact Phase 0 gates.
- `make verify` remains authoritative; no partial run may set `verificationPassed=true`.
