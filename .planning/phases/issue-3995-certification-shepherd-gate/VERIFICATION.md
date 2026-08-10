# VERIFICATION — issue #3995 shared connector-certification Shepherd gate

## Status

Inline/manual GSD verification and code review are complete. The final `no-mistakes` gate and
child-PR publication remain pending.

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
- [x] Focused Go, generator, formatting, static/build, individual repository, and workflow evidence
      gates pass: `go test -timeout 20m ./internal/agentcontract -count=1`,
      `go test -timeout 20m ./cmd/agentcontractgen -count=1`,
      `go test -timeout 20m ./internal/cli -count=1`, `go vet ./...`, `go build ./cmd/pm`,
      `make lint`, `go run ./cmd/agentcontractgen sync`,
      `go run ./cmd/agentcontractgen check`, `make agent-contract-check`,
      `scripts/verify-gsd-workflow origin/feat/3988-github-certification`, and `git diff --check`.
      Earlier individual repository gates also passed: `tidy-check`, `docs-check-no-build`,
      `smoke-no-build`, `connectorgen-validate`, `connectorgen-surface-sync`,
      `connectorgen-certification-matrix`, `connector-canon-check`, `connector-boundary`, and
      `release-workflow-check`.
- [x] Inline GSD verify/gap loop and code review record every finding/disposition. `REVIEW.md`
      records two findings, both fixed in correction round 1 of 5; no actionable review findings
      remain.
- [ ] `no-mistakes` runs without `--yes`; final branch/PR target and parent references are correct.
