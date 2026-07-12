# TDD Ledger: Shepherd Supervision

## Workflow preflight — 2026-07-12

- Issue #326 was narrowed before edits to the existing Shepherd launcher lifecycle.
- `github-issue-first-delivery` and `gsd-programming-loop` are active. The repository GSD adapter
  still omits `programming-loop`; the installed helper plus universal/manual TDD loop is the
  recorded fallback.
- Parent branch `fix/323-auto-loop-hardening` and draft PR #324 exist; the branch starts from the
  verified alignment commit `aef7fe86`.
- The task is shell/control-plane work; no Go or UI implementation skill is required for this slice.
- Production edits: none at this checkpoint.

## RED matrix

| Scenario | Current unsafe behavior | Required result |
|---|---|---|
| 32 concurrent starts | multiple providers may launch | one winner; 31 `CONTROLLER_HELD` |
| hard deadline + grandchild | child liveness can extend turn | TERM/KILL entire exact group; no validator |
| leader exits, child lives | root exit can look complete | durable orphan HALT; no checkpoint/validator |
| validator HALT | process-only exit | atomic HALT before quiescent exit |
| resume after HALT | prompt is read and loop restarts | reject before prompt/provider; bytes unchanged |
| signal during turn | descendants may survive | drain exact group; `recovery_required` |
| fence moves | old controller keeps mutating | drain and fail `CONTROL_FENCE_MISMATCH` |
| pause/resume at cap | turn counter resets | persisted monotonic ordinal and cap |

RED command/output will be appended after the tests exist and before production changes.

## RED checkpoint — 2026-07-12

- Added `scripts/tests/pi-shepherd-supervision.sh`. It copies the exact launcher body into a
  temporary fixture and mechanically removes only the Phase 0 guard; production has no enable flag
  or environment bypass. Fake Pi processes, synthetic local state, and test-owned PIDs are used.
- `bash -n scripts/tests/pi-shepherd-supervision.sh` -> pass.
- `bash scripts/tests/pi-shepherd-supervision.sh` -> expected exit 1 with 11 failures:
  - all 32 concurrent controllers launched an orchestrator; no contender received the required
    controller-held result;
  - hard deadline exited 3, started a validator, and wrote no halted/quiescent control state;
  - leader exit with a live descendant was accepted, validator started, and the child survived;
  - validator HALT produced no durable control latch;
  - signal termination left its descendant alive and wrote no recovery state;
  - the turn cap/ordinal was not persisted for resume.
- Production files changed before RED: none.
- Expected-failure integrity: the harness cleaned every recorded test child on exit; no external
  provider, GitHub mutation, credential, or non-test process was touched.

## RED harness audit

- An independent read-only audit confirmed each baseline failure maps to a missing launcher
  behavior, then identified potential false-green/test-safety gaps before implementation.
- The harness now removes exactly one sentinel-delimited Phase 0 block, uses a clean environment,
  verifies nonce-bound PID identity before cleanup, preserves same-user bystanders outside the
  role group, measures hard-deadline elapsed time, requires exact contention exits/codes, renames
  rather than chmods the resume prompt, and adds the previously missing fence-movement case.
- Audited RED: `bash scripts/tests/pi-shepherd-supervision.sh` exits 1 with 13 expected assertions,
  including 32/32 controller launches, natural deadline completion, live orphan, missing durable
  HALT/recovery, absent active fence, and reset turn cap. Bystanders survive, and only nonce-bound
  test processes are eligible for cleanup.

## GREEN implementation — 2026-07-12

- The existing `scripts/pi-shepherd-loop.sh` now retains one nonblocking `flock` across its lock
  acquisition re-exec and registered descendants. Thirty-two concurrent starts produce one winner
  and 31 typed `CONTROLLER_HELD` exits without launching a second provider.
- One atomic, size-bounded, no-symlink `CONTROL.json` stores the exact controller epoch fence,
  lifecycle phase, lease, active turn/role PID+PGID, independent role session IDs, durable limits,
  monotonic turn ordinal, and restart-persistent revert/no-verdict/active-time counters.
- Orchestrator and validator roles launch in distinct process groups behind a ready handshake. The
  controller applies one persisted hard deadline, bounded TERM/KILL drainage, heartbeat/fence
  checks, leader-exit orphan detection, and quiescence proof before moving to the next role.
- HALT/recovery is latched before teardown. Failed persistence or failed drainage cannot be
  finalized as halted/quiescent, and the last process handles remain available when quiescence is
  uncertain. Fresh start and resume both reject halted/recovery/dirty active state before prompt or
  provider access.
- Only the Shepherd validator defaults to `openai-codex/gpt-5.6-sol --thinking high`; the
  orchestrator remains `openai-codex/gpt-5.5` and worker configuration is unchanged. Default model
  availability is checked after fenced authority and startup-safe traps but before mutable work.
- The production Phase 0 first-action fuse remains first, closed, and has no enable bypass. Tests
  remove only its unique sentinel-delimited block from an isolated temporary copy.

## Independent review corrections — 2026-07-12

- An adversarial implementation review found six P1 lifecycle gaps: signal windows before full
  traps and before role binding, false quiescence after failed teardown, ignored HALT-latch
  failure, incomplete paused-state validation, and counters that reset on resume. Each received a
  production fix plus a deterministic regression test before final verification.
- The final 19-scenario suite additionally covers bounded TERM during validator-model discovery,
  launch-before-bind, bind-before-authorization, deadline-before-authorization, a shared remaining
  orchestrator/validator deadline, descendant-only inherited-lock fencing after controller and role
  leader SIGKILL, failed HALT persistence, and fence movement after durable role binding.
- State tests separately reject each dirty paused invariant, malformed/negative/boolean persisted
  numerics, dangling and live control symlinks, and hardlinks without rewriting bytes or launching
  work. Result tests retire stale verdicts and require Shepherd PROCEED before `human_gate`/`done`,
  while `budget` and `blocked` remain unconditional safety stops.
- Final test review found that lock-acquisition re-exec could re-export an internally generated
  validator default and misclassify it as a caller override. The default is now deliberately
  non-exported across re-exec, so only genuinely caller-supplied `VALIDATOR_ARGS` skips the exact
  default-model preflight.
- Same-epoch rollback detection is intentionally not claimed by this controller-epoch fence. A
  stale same-fence snapshot can only be eliminated by the per-transition predecessor/version
  contract owned by #327; it remains an enable blocker behind the closed Phase 0 fuse.
- Authenticated recovery/takeover after controller SIGKILL and OS containment of
  setsid/double-fork escapees are intentionally not claimed here. They remain dependency-ordered
  work for #339 and #342; the closed Phase 0 fuse prevents live autonomous use in the interim.

## Focused GREEN evidence — 2026-07-12

- `bash -n scripts/pi-shepherd-loop.sh scripts/tests/pi-shepherd-supervision.sh` -> pass.
- `shellcheck --severity=warning scripts/pi-shepherd-loop.sh scripts/tests/pi-shepherd-supervision.sh` -> pass.
- `bash scripts/tests/pi-shepherd-supervision.sh` -> `pi-shepherd-supervision: ok`.
- `bash scripts/tests/auto-loop-control.sh` -> `auto-loop-control: ok`.
- `make agent-loop-test` -> pass, including Go unit/race/CLI gates and both shell harnesses.
- Two independent final read-only reviews found no remaining in-scope P0/P1 blocker. They classify
  same-epoch rollback (#327) and daemon-escape containment (#342) as later enable blockers, not as
  claims or regressions of this closed-fuse #326 slice.
- Final `make verify` -> pass: formatting/tidy, vet, all Go tests, build, connector-doc validation,
  smoke flow, lint, 547 connector definitions, Go race/control gates, Phase 0 controls, and the
  19-scenario Shepherd supervision harness.

## Post-PR stability correction — 2026-07-12

- Repeated local execution exposed scheduler-dependent failures that single-pass CI did not catch.
  The strongest preserved RED fixture reached `recovery_required` with
  `CONTROLLER_UNCLEAN_EXIT`, an allocated turn, and no validator event. A separate run failed the
  persisted turn-cap assertion at repetition 16, and the original SIGKILL setup failed at
  repetition 38.
- Root cause: role authorization used shell noclobber redirection. That creates the authorization
  path before `printf` fills it, so the inert child could read an empty token and exit 125 before
  executing the provider. `kill -0` also classified zombie leaders as running, and the first reaping
  correction could wait for the same leader twice.
- Production GREEN: explicit zombie detection reaps once; an empty or uncertain `ps` result remains
  supervised until conclusive evidence or the persisted deadline. The subsequent exact-head review
  replaced the intermediate filesystem authorization design as described below.
- Oracle GREEN: the SIGKILL child verifies that its inherited descriptor is the canonical lock and
  does not alter its lock state. The test observes durable role binding before child/lock readiness,
  proves each killed identity disappears, and polls for lock release under a bound instead of
  sleeping. Turn-cap/HALT failures capture one atomic control snapshot plus bounded diagnostics.
- Cumulative correction evidence: combined instant-exit/HALT/ratification matrix 30/30;
  turn-cap matrix 50/50; descendant-held-lock matrix 50/50; two complete 19-scenario suites;
  Phase 0 harness; `make agent-loop-test`; and final `make verify` all pass. Tests use only fake
  providers and test-owned processes.

## Exact-head P1 review corrections — 2026-07-12

- Independent local review blocked commit `508a4554` on four P1 lifecycle/oracle findings and one
  P1 evidence-hygiene finding: a mismatched PID could be signalled, file authorization could be
  preseeded or consumed before publication completed, the SIGKILL oracle could acquire the lock it
  claimed to observe, and unknown filters could silently execute zero tests.
- RED deterministically reproduced all control failures: preseed escaped to a paused completed turn,
  controller death never reached a safe pre-GO barrier, and simulated PGID movement killed the
  now-untrusted role PID. Invalid, duplicate, empty, and unknown filters are separately rejected.
- The filesystem authorization path was removed. An inert role now re-derives its exact durable
  fence/turn/role/PID/PGID binding, writes a readiness marker, and kernel-stops itself. Only after the
  controller bind call has returned from directory fsync and the exact PID is observed stopped does
  the controller send `CONT`; the role revalidates parent liveness, deadline, and binding after
  resume before `exec`.
- A PID/PGID mismatch now preserves recovery evidence without signalling either untrusted identity.
  The inherited-lock oracle only validates the descriptor/inode; rejection by a fresh controller
  after the original controller and role leader are proven gone establishes that the untouched
  descendant retained the lock.
- GREEN evidence on the corrected design: kernel-GO adversarial matrix 20/20, failed-HALT/instant
  validator 50/50, shared validator deadline 10/10, two complete 22-scenario suites, Phase 0
  controls, `make agent-loop-test`, and final `make verify`. All provider processes are synthetic
  and test-owned.
