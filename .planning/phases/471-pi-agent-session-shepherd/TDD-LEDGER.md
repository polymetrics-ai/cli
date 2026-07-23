# TDD Ledger

Phase: `471-pi-agent-session-shepherd`

## Autonomous replacement reset

The entries below preserve RED/GREEN/refactor evidence for the earlier read-only control-plane
foundation. They do **not** establish completion of the autonomous Shepherd now required by #471.
The replacement TDD queue was issue-scoped. The current-gate column below records whether each
capability is present in the verified #479 aggregate; it does not claim independent GitHub
issue/PR closure. Historical per-issue artifacts retain their own process and review gates.

| Issue | RED target | Current gate |
|---|---|---|
| #473 | cancellation ordering, epoch cleanup/acquisition race, root identity, state invariants, epoch bounds | integrated through PR #482; issue reopened pending #472→`main` |
| #474 | DAG cycles/readiness/collisions/retry/reconciliation | integrated through PR #483; issue reopened pending #472→`main` |
| #475 | exact role routing, least authority, abort/close, bounded handoffs | integrated through PR #486; stale blocked child review is superseded by #479/#490 and final parent review |
| #476 | worktree identity, safe Git operations, crash/idempotency | integrated through PR #484; issue reopened pending #472→`main` |
| #477 | idempotent comments, actor/head binding, consume-once decisions | integrated through PR #485; issue reopened pending #472→`main` |
| #478 | issue/PR idempotency, review coverage, scoped integration | integrated through PR #487; issue remains open pending #472→`main` |
| #479 | full production trajectories, command UX, exact-effect recovery, and parent-base CAS | 17/17 functional matrix and remote CI integrated through PR #489 at parent `daaa2263` |
| #490 | Pi 0.80.10/event-agnostic completion and bounded workflow-engine developer-tool boundary | integrated through PR #491 at parent `c3f4f683`; child review does not approve the parent seam |
| #480 | fault-injected restart/audit/reversible cutover preparation | worker-ready; behavior RED required before production edits |
| #481 | deterministic + read-only live #397/#438 canary and post-pass deprecation activation | dependency-blocked by #480; behavior RED required before harness/docs edits |

Every child must append its exact failing command/output summary before production changes, then
record GREEN/refactor commits and verification. Historical canary success cannot be reused as a
pass for a mutating autonomous path.

## Reconciliation plan-first checkpoint (2026-07-23)

- GSD path: `scripts/gsd doctor` passed; `scripts/gsd prompt plan-phase 471 --skip-research`
  generated the official Pi planning contract. The already-recorded manual programming-loop fallback
  remains active because the absent command was checked once and not retried.
- Required skills/references: `gsd-core`, parent-orchestrator contracts, Pi/runtime routing, and the
  complete #490 workflow-engine boundary. No Go implementation skill applies before final parent
  verification.
- Shepherd preflight: help was inspected through offline Pi 0.80.10 RPC; read-only status reported
  `No persisted Shepherd run exists for this issue.` This authorizes one `start`, not a duplicate
  run; all later execution uses `resume`.
- GitHub reconciliation: PRs #482-#489 and #491 are merged only into the parent branch. #480 is the
  sole ready implementation and #481 is blocked by it. Stale/deleted remote child branches and
  retained historical worktrees will not be dispatched.
- TDD decision: seed #480/#481 PLAN, TDD ledger, and verification checklists before the ignored
  production plan is started. Each child gets one behavior-level RED→GREEN→refactor cycle and at
  most one comprehensive review/correction round.

## Parent-owned live preflight RED/GREEN

- Required-check RED: the selected tests executed against compiling production and failed because
  an unprotected parent branch queried a protected-only endpoint and GitHub null `app_id` values
  were rejected. GREEN at `95a4bd697` passed 3/3 selected cases, 660/660 affected tests, and live
  read-only policy construction with exact empty parent checks plus protected `main` checks.
- Embedded OAuth behavior RED: `/pm-shepherd canary --issue 397 --pr 438 --read-only --backend
  sdk-inproc --experimental --max-concurrency 2 --timeout-seconds 900` ran from clean persistent
  checkout `/private/tmp/shepherd-481-cli-architecture-canary` at exact head `21d195aff`. Durable
  generation 1 (`run-689079d0-089c-4d72-ab78-3dd1b213923e`) halted: the scout completed exact
  read-only reconciliation, while the concurrent validator failed with `embedded AgentSession
  cannot inherit host-only OAuth for openai-codex`. #438 stayed open/draft/unmerged and unchanged.
- Intended GREEN: one lazily initialized public Pi `ModelRuntime` is shared by all embedded legacy
  and production session adapters for the extension host. Concurrent callers await the same
  initialization; a different agent directory or genuinely unavailable child auth fails closed.
  The runtime reads the normal mode-0600 credential store itself; Shepherd never reads, copies,
  logs, stores, or forwards a credential value. Focused concurrency/auth tests and a new live
  canary generation must pass before dispatch.
- Orchestration decision: `local_critical_path`. A core embedded-session preflight defect blocks the
  controller that would dispatch #480, so the parent owner performs only this bounded TDD
  correction. Workflow-engine run `93478be5-f7af-41d7-abf1-494a67cdaebf` supplied two read-only
  gpt-5.6-sol/high research lanes with no GitHub mutation authority; it is analysis, not Shepherd
  success evidence or human approval.

## Historical read-only foundation ledger

## RED: command parser and registration

- Status: expected failures captured before production code.
- `node --test .pi/extensions/shepherd/arguments.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `arguments.ts`.
- Offline Pi RPC command discovery exited 1 because `.pi/extensions/shepherd/index.ts` did not
  exist. No model, auth, or network call was made.

## RED: deterministic domain

- Status: expected failure captured before production code.
- `node --test .pi/extensions/shepherd/domain.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `domain.ts`.
- Covered contracts: geometric rating, low-score correction, hard-gate precedence, dependency
  readiness, bounded concurrency/one mutator, and crash reconciliation.

## RED: atomic state and controller

- Status: expected failures captured before production code.
- State-store test exited 1 with `ERR_MODULE_NOT_FOUND` for `state-store.ts`.
- Controller test exited 1 with `ERR_MODULE_NOT_FOUND` for `controller.ts`.
- Covered contracts: mode-0600 atomic state, summary redaction/bounds, duplicate ownership,
  concurrent independent lanes, stale evidence halt, fresh resume generation, stop, and shutdown.

## RED: Pi SDK adapter

- Status: expected failure captured before production code.
- `node --test .pi/extensions/shepherd/sdk-runner.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `sdk-runner.ts`.
- Covered contracts: resource isolation, exact model/thinking/tools, in-memory state, recursion and
  persistence rejection, run-bound abort, and unconditional cleanup.

## RED: exact target evidence

- Status: expected failure captured before production code.
- `node --test .pi/extensions/shepherd/target-evidence.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `target-evidence.ts`.
- Covered contracts: argv-only git/GitHub reads, clean tree, open PR, exact branch, and exact head.

## Gate result

The manual strict-TDD gate passes: all behavior tests and the registration smoke failed for the
expected missing-production-code reason. GREEN implementation may now begin. Refactor evidence will
be added only after the focused suite passes.

## GREEN: deterministic core and SDK adapter

- Core worker commit `38dcb435745f333787fff3e3b4ea3dd0d585db1c` passed 14/14 focused
  tests and was integrated as `ee2285b1`.
- SDK worker commit `3583b584c14365b14257cbb27f81ea16d4a08340` passed 3/3 focused
  tests and strict TypeScript, and was integrated as `107e74fb`.
- Subsequent no-tools hardening commit `2ccd6f4c` expanded the SDK suite to 7/7: embedded sessions
  receive no built-in or custom tools, fail closed on malformed/oversized evidence and event
  limits, time out through owned abort, and always clean up.
- State persistence hardening commit `19e6dcaf` passed 6/6 focused tests: disk DTOs are explicitly
  allowlisted, unknown disk fields fail closed, summaries are redacted/bounded/single-line, and
  runtime-only fields are stripped before serialization.

## RED to GREEN: independent review corrections

- Reviewer RED: controller enum transitions could not round-trip through the validating store.
  GREEN: canonical `succeeded|failed|halted|stopped` lane states; real-store regression passes.
- Reviewer RED: persisted `running` state could not resume after a host crash. GREEN: resume first
  reconciles it to `interrupted`, then creates a fresh generation/head/nonce; regression passes.
- Reviewer RED: an in-flight lane could overwrite `stop`. GREEN: controller-owned cancellation is
  checked before and after final persistence; deferred-runner regression leaves disk `stopped`.
- Reviewer RED: child read tools were not repository-confined. GREEN: child sessions receive
  `tools: []`, `customTools: []`, and `noTools: "all"`; requested tools fail before creation.
- Reviewer RED: model-authored success could finish without recapturing the target. GREEN: host
  code recaptures exact local, PR, and check evidence after both lanes and hard-halts on change.
- Extension RED: two issues could launch concurrently and shutdown produced late UI output.
  GREEN: one process-wide active run, shutdown cleanup/waiting, and notification suppression.

## Current focused gate

- `node --test .pi/extensions/shepherd/*.test.ts`: 49/49 pass.
- Strict TypeScript no-emit over all production files including `index.ts`: pass.
- Offline Pi 0.80.6 RPC command registration: pass (`pm-shepherd`, source `extension`).

## RED to GREEN: ownership and shutdown audit

- Audit RED reproduced two controllers sharing one FileStateStore: both completed and dispatched
  four total lanes. GREEN: a root-global mode-0600 O_EXCL lease with PID/token/inode fencing allows
  one owner; live owners fail closed and only same-issue `resume` can recover a dead owner.
- Audit RED stopped during initial target capture: stop returned no-state while lanes later ran.
  GREEN: lifecycle ownership exists before the first await, stop marks it cancelled, waits its
  terminal persistence, and the regression records zero AgentRunner dispatches.
- Audit RED aborted during delayed `createSession`: the eventual child still prompted. GREEN:
  run-level cancellation tombstones and AbortSignal checks reject after creation, clean up, and
  prove zero prompts plus dispose.
- Audit RED hung `abort`/`waitForIdle`: close never settled and dispose was skipped. GREEN: cleanup
  steps and runner close are bounded, unsubscribe/dispose remain unconditional, and both run/close
  hang probes settle with dispose observed.
- Graceful shutdown now cancels owned lifecycles as `interrupted`; unowned stop cannot rewrite disk.

## RED to GREEN: final deep-review remediation

- Exact-head GSD deep review at `7f745427d38995940b8f57517d0241d1e10d3f64` reported ten
  critical findings and six warnings. The report is preserved in `471-REVIEW.md`; no finding was
  waived.
- SDK RED: 10/18 focused cases failed for non-success Pi terminals, post-prompt abort, hanging or
  late setup, end-to-end deadlines, shared close settlement, and Windows path forms. GREEN:
  23/23 pass; only a verified `stop` terminal is parsed, one deadline owns setup through teardown,
  and every late child is cleaned without prompting.
- State/lease RED: eight adversarial cases failed for secret-bearing persistence, crash-partial
  publication, PID reuse, symlink roots/files, trusted-root escape, and contradictory state.
  GREEN: 20/20 state-store tests pass with a CAS-style append-only lease journal, atomic hard-link
  publication, process identity, descriptor-bounded no-follow reads, trusted-root checks, fixed
  persisted summary categories, and run/lane/dependency invariants.
- Controller RED: 6/24 focused cases failed for dropped/mismatched resume PRs, early lease release,
  terminal stop races, shutdown suppression, and lexical worktree identity. GREEN: 30/30 pass with
  effective persisted target identity, canonical worktree caching, structured cancellation/join,
  mandatory leases, terminal linearization, and propagated cleanup failures.
- Integration RED: the root strict-TypeScript command rejected the descriptor buffer type, and an
  added concurrency test reproduced two launches racing during asynchronous worktree resolution.
  GREEN: a typed `Uint8Array<ArrayBuffer>` descriptor buffer and a synchronous launch reservation
  make the complete suite 82/82 green.
- Manual-GSD fallback remained necessary because the repository adapter still did not expose
  `programming-loop`. All three implementation lanes used `gpt-5.6-sol` at high reasoning and
  recorded their RED/GREEN evidence before integration.

## Current focused gate after deep-review fixes

- `node --test .pi/extensions/shepherd/*.test.ts`: 82/82 pass.
- Strict TypeScript no-emit over every production Shepherd module: pass.
- Offline Pi 0.80.6 RPC registration: pass (`pm-shepherd`, source `extension`).
- `git diff --check`: pass.
