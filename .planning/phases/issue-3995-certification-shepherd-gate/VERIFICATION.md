# VERIFICATION — issue #3995 shared connector-certification Shepherd gate

## Status

Planned. This checklist is updated after each inline/manual GSD stage.

## Required acceptance evidence

- [x] R1 RED fails before evaluator implementation and records the exact expected GitHub ID.
- [x] R1/R2 all-green fixture returns `PROCEED`; each isolated criterion defect names its exact ID.
- [x] Invalid/missing/unknown schema, unknown field, missing evidence pointer/sidecar, and omitted
      adapter gate field fail closed.
- [x] Current #3984 GitHub generated baseline returns `RETRY`, while generic contract check passes.
- [x] Evaluation is read-only: no evidence creation, credential access, provider call, or
      `cmd/connectorgen/certification*.go` edit.
- [x] Four generated harness projections contain equivalent canonical gate I/O schema; sync/check
      pass and deliberately missing/drifted projection tests fail.
- [x] Required transition boundaries reject non-`PROCEED` verdicts without discarding exact IDs.
- [ ] Focused Go, generator, formatting, static/build, individual repository, and workflow evidence
      gates pass.
- [ ] Inline GSD verify/gap loop and code review record every finding/disposition.
- [ ] `no-mistakes` runs without `--yes`; final branch/PR target and parent references are correct.
