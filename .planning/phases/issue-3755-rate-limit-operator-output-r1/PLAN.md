# Phase issue-3755-rate-limit-operator-output-r1 — rate-limit operator output

## Required skills and GSD path

`golang-how-to`; `golang-cli`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-testing`; `golang-context`; and `golang-concurrency` were loaded. `scripts/gsd prompt discuss-phase 3755` and `scripts/gsd prompt plan-phase 3755 --tdd` were generated and applied inline because Pi roles are unavailable and the canonical contract forbids role spawning.

`scripts/gsd prompt execute-phase 3755` was generated and applied inline before the red-test and implementation slices. The same documented no-spawn fallback is used for verify-work and code-review; their evidence is recorded in `VERIFICATION.md` and `REVIEW.md`.

## TDD slices

| Slice | RED proof | GREEN implementation | Guard |
| --- | --- | --- | --- |
| A — bounded secret-free run report | A report cannot represent declared/undeclared state or safely aggregate a selected policy, provider observation, local wait, provider retry wait, and request latency | Add a concurrency-safe summary owned by `RuntimeConfig` with a fixed declaration state and one coalesced row per selected policy | fields accept only IDs, subject kinds, selector dimensions, scalar counters/times, and typed budget facts; never a map/string value from secrets/config/binding/revision |
| B — observation bridge | Resolved policies and requester responses affect enforcement but do not update the report; admission/retry timing is invisible | Wrap existing resolved limiter admission and observer paths; emit exact pacing waits from the injected limiter clock plus bounded request/retry timing callbacks | no change to resolver selectors, limiter reservation, retry count, sleep value, reset handling, or requester payload/header behavior |
| C — sync result and CLI | ETL run output only reports record counts, so neither human nor JSON output can distinguish pacing, provider 429, and latency | Attach one collector to source/destination runtime configs, persist the snapshot on `app.Run`, print a compact ETL rate-limit line, and expose the same object under `run.rate_limit` | no per-request lines, no output filtering/masking/redaction path, and failure summaries remain safe |
| D — declaration/state and documentation | No declaration reports as an absence, not an explicit operator state; docs omit the machine/human contract | Report `undeclared` for absent bundle policy; cover declared observation/429 and update ETL manual/website/generated help surfaces | test-only `rate_limits.json` only; production `defs/defs.go` unchanged |

## Commit checkpoints

1. Commit the discussion, plan, TDD ledger, and verification checklist.
2. Commit a focused red-test checkpoint after the new test contracts fail for missing report/ETL output behavior.
3. Commit the green bounded reporting, CLI/docs, and test-only fixture slice after scoped verification.
4. Commit review fixes only after rerunning the affected tests and checks.

## CLI help/manual/website parity

- No command, flag, namespace behavior, or completion changes: not applicable.
- Verify `pm help etl`, `pm etl`, and `pm etl run --help`; add/update the ETL result/output documentation in `docs/cli/**` and the corresponding website source/generated surface.
- Test both the existing human output and `--json` ETL run payload; document that `run.rate_limit` is a bounded secret-free summary and `undeclared` is not a claim of unlimited access.

## Verification

Run focused connector/core/app/CLI tests, race tests for the report collector, `gofmt`, targeted `go vet` and build, CLI help/manual/website generator or grep checks, connector validation/surface-sync, and each non-monolithic repository gate. Full `go test ./...` and `make verify` remain CI-owned by repository policy.
