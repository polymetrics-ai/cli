# Prompt Trace: #478

## Kickoff snapshot

- Coordinator objective: implement issue #478 in exactly the three assigned production modules,
  matching tests/fixtures, and this issue-local phase directory.
- Model role: `openai-codex/gpt-5.6-sol:high` implementation worker.
- Immutable base: `3addb1f48be1afe8b1e2b59b54247679d7293805` on
  `feat/471-pi-agent-session-shepherd`.
- GSD preflight: `scripts/gsd doctor` passed; `scripts/gsd prompt programming-loop init --phase
  478-shepherd-github-parent-orchestration --dry-run` returned `unknown GSD command:
  programming-loop`.
- Runtime decision: `manual_gsd_fallback` plus `local_critical_path`. Two attempts to dispatch the
  read-only API-recon subtask were rejected by the runtime agent-thread limit.
- Contract resolution: the live #478 issue/coordinator handoff supersedes an older local draft;
  only exact-head `codex_independent` review using `openai-codex/gpt-5.6-sol:xhigh` can satisfy the
  automated review gate. Claude, Copilot, generic Codex, and human records are rejected for that
  gate.
- Human gates: all policy exceptions and the final parent ready/merge decision route through the
  existing broker. This worker will represent but not cross them.

## Stable-head resume

- The parent temporarily paused #478 review work while #475 established a stable predecessor,
  then explicitly resumed this branch without resetting its pushed TDD history.
- Reconciliation found pushed head `db9fbc33` (the adversarial test-only RED checkpoint), with
  prior production GREEN at `90321ffb`; the uncommitted matching correction was preserved.
- Parent review policy: finish and push an exact GREEN candidate, but do not start any reviewer.
  The parent orchestrator owns the later stable-head specialist campaign.
- Correction GREEN: `40ce66d4b5010b92089895a05709687143d15a05`; 27/27 focused tests and
  strict owned TypeScript pass before broader verification.
- Authorized broader verification only: serialized Shepherd TypeScript tests, strict pinned Pi
  0.80.6 TypeScript, pinned offline Pi RPC, and immutable-base/diff/owned-scope checks. No Go,
  connector, certification, runtime-service, or `make` command is permitted.

## Functional review correction kickoff

- Frozen reviewed head: `093b3c90409cedc6b7008b7510f53937eb1ebbc1`; frozen base:
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Review input: `/tmp/478-REVIEW-FUNCTIONAL.md`, 7 critical and 4 warning findings, all accepted
  for one strict correction slice.
- Sequence: artifact-only plan commit, one behavior-level test-only RED commit covering all eleven
  findings with production byte identity proof, coherent GREEN, authorized Shepherd verification,
  push/update PR #487, then parent-owned fresh xhigh review.
- Scope exclusion: no #475/#479/controller/top-level wiring; the session provenance addition is a
  scoped interface and fixture only.
- GSD: `manual_gsd_fallback`; cycle decision `local_critical_path` in the isolated #478 clone.

## Cycle 3 corrected-review prompt

- Inputs: `/tmp/478-REVIEW-CORRECTED-1.md` and `/tmp/478-REVIEW-CORRECTED-2.md`, both read in full.
- Frozen candidate/base: `3f285722a505ea426d53a34f95716781d1aca7c2` /
  `3addb1f48be1afe8b1e2b59b54247679d7293805`.
- Implementation/correction route: Codex `openai-codex/gpt-5.6-sol:high`,
  `local_critical_path`; `xhigh` remains parent-owned planning/review/orchestration only, and no
  Claude/Copilot finding is imported.
- Sequence: artifact-only plan commit; exactly one test/fixture-only RED commit proving all three
  production blobs unchanged; architectural GREEN/refactor; focused/authorized broad verification;
  evidence commit. Network publication remains deferred under the recorded DNS failure.
- Fourteen-invariant batch: canonical persisted plans; mutating-only children; exact outer/nested
  evidence identity, completeness, chronology, and freshness; receipt provenance; durable
  cross-instance mutations; exact ancestry; deterministic same-marker review attempts; symbolic-ref
  rejection; versioned plan-bound CI; monotonic roster CAS; exported attestation API; adversarial
  bounds, partial effects, and secret safety.
- Exclusions: #479 controller, parent planning artifacts, live GitHub, Go/connectors/certification,
  runtime services, `make`, external reviewer, integration, or merge.
