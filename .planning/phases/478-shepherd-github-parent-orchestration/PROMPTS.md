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
