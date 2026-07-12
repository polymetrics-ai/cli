# Test Plan: Phase 0

## Fixture and replay tests

- Table-drive all thirteen fixture files and expected violation/decision/exit triples.
- Assert fixtures use complete synthetic bindings at fixture and event level, contiguous sequences,
  stable run/generation/controller identity, and neutral typed facts with complete
  resource/owner/before/after values. Multi-turn incidents retain distinct turn/attempt bindings.
- Assert observed decisions/outcomes are separate from required policy and preserve correct
  HALT/RETRY records rather than forcing observed != required.
- Mutate artifact/head/owner/actor/order/binding/HALT/worker facts and prove replay changes or
  rejects the incident instead of echoing an expected label.
- Assert dead-worker evidence distinguishes phantom dispatch from executed-without-handoff and
  false-green evidence includes a missing required artifact plus later repo-gate failure.
- Assert directory replay is filename-sorted and repeated marshaling is byte-identical.
- Reject unknown fields, trailing JSON, incomplete/mismatched/non-synthetic identity, non-monotonic
  sequence, unknown event kind, `.jsonl`, oversize input, raw command/prompt/session path, and a
  sensitive string assembled only in test memory.

## Safety and CLI tests

- Assert closed status is independent of environment and has no mutating method or command.
- Assert only the two marked driver paths are tracked and untracked guard requests fail closed.
- Assert loopctl help/bare behavior, status, entrypoints, guard-driver, replay, malformed fixture,
  unknown-command behavior, negative enable/open/run/resume commands, and non-echoing errors.

## Shell characterization

- Copy drivers and safety helper into a temporary fake repo.
- Use harmless stub executables that only write a marker.
- Red baseline: current run/resume reaches the stubs and writes loop state.
- Green requirement: run and resume exit 78, marker is absent, and no loop state directory is
  created; help exits 0 and also creates no state.
- Discover wrapper candidates independently from filename and semantic signals; require every
  candidate to be inventoried and guarded.
- Run drivers under `env -i`, a stub-only PATH, isolated config/home, unwritable state, unreadable
  resume input, and hostile enable-like environment/flag canaries. No external tool may run.

## Broader gates

- Package tests under race detection.
- `make agent-loop-test` aggregates the exact Phase 0 gates.
- `make verify` remains authoritative; no partial run may set `verificationPassed=true`.
