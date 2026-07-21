# Plan: Issue #471 Pi AgentSession Shepherd

## Objective

Deliver a dependency-free Pi extension that owns a deterministic, resumable interactive
supervision cycle over embedded AgentSession lanes and proves the design through a merge-disabled,
read-only canary against CLI Architecture v2 PR #438.

## Required Workflow And Skills

- GSD command attempted: `scripts/gsd prompt programming-loop run --phase 471-pi-agent-session-shepherd`.
- Adapter result: unavailable on the current `main`; manual-GSD fallback recorded in `CONTEXT.md`.
- Loaded skills: `gsd-programming-loop`, `github-issue-first-delivery`,
  `architecture-patterns`, and `javascript-testing-patterns`.
- Read repository contracts: required skill routing, Pi adapter, issue agent contract, parent
  orchestration contract, universal runtime loop, automated review routing, and Claude review loop.
- No Go implementation is in scope, so Go implementation skills are not required for production
  edits. Root Go gates still run as compatibility checks.

## Architecture

Use ports and adapters so the deterministic core has no Pi SDK dependency:

1. `domain.ts`: run/lane types, invariant validation, stale-evidence rejection, hard-gate
   classification, geometric-mean rating, and restart reconciliation.
2. `arguments.ts`: strict slash-command parser with bounded identifiers and fail-closed flags.
3. `state-store.ts`: atomic mode-0600 JSON persistence outside the repository under the Pi agent
   directory; injectable temporary roots for tests.
4. `runner.ts`: AgentRunner port plus a bounded child registry and concurrency governor.
5. `sdk-runner.ts`: Pi 0.80.6 adapter with exact model/thinking/tool checks, in-memory child
   sessions/settings, recursion prevention, timeouts, and idempotent cleanup.
6. `controller.ts`: read-only scout and validator orchestration, exact run/generation/head/nonce
   binding, deterministic gating, rating, persistence, stop, and resume.
7. `index.ts`: `/pm-shepherd` registration, concise status/help, lifecycle notifications, and
   parent `session_shutdown` cleanup.

## Command Surface

- Bare `/pm-shepherd`, `help`, and `status` never start a model.
- `canary --issue 397 --pr 438 --read-only --backend sdk-inproc --experimental` is the first live
  path and runs independent scout and validator sessions.
- `start` and `resume` require the same explicit backend/experimental acknowledgement. Until a
  mutating path has an isolated worktree and scope contract, omission of `--read-only` fails closed.
- `stop --issue N` aborts only children owned by that exact run.

## TDD Tasks

1. RED — command parsing and registration
   - reject unknown actions, invalid IDs, duplicate flags, control characters, traversal, missing
     experimental acknowledgement, and non-read-only initial runs;
   - prove status/help do not dispatch an agent.
2. RED — deterministic domain
   - prove stale run/generation/lane/head/nonce evidence is rejected;
   - prove hard gates override high diagnostic scores;
   - prove concurrency is bounded at two and mutation at one;
   - prove recovery converts persisted running lanes to interrupted.
3. RED — persistence and controller
   - prove atomic replacement, mode 0600, redacted/bounded summaries, duplicate-run rejection,
     cooperative stop, and resume generation changes.
4. RED — SDK adapter
   - inject a fake SDK and prove exact Pi version/surface, model, thinking, tool allowlist,
     no-extension resources, in-memory state, timeout abort, unsubscribe, and dispose.
5. GREEN — implement the minimum production code to satisfy each failing slice.
6. REFACTOR — centralize invariants, redaction, cleanup, and rendering only after focused tests pass.
7. VERIFY — run focused tests, Pi extension discovery/load smoke, root compatibility gates, and an
   independent exact-head read-only canary against PR #438.

## Verification

- `node --test .pi/extensions/shepherd/*.test.ts`
- strict TypeScript no-emit check when the installed Pi package provides a compiler-compatible
  resolution path; otherwise record the exact unsupported check and rely on Pi load plus Node tests
- `pi --list-extensions`
- project-local Pi extension load smoke without a model call
- `git diff --check`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`

## Commit Checkpoints

1. Plan and RED evidence.
2. GREEN/refactor implementation.
3. Verification and canary evidence.
4. Review fixes, if any.

## Hard Stops

- new dependency or provider/auth change;
- secret access or persistence;
- mutating PR #438, connector calls, reverse ETL, production effects, or broad process control;
- quality-gate reduction;
- merge of this PR or any parent PR into `main`.
