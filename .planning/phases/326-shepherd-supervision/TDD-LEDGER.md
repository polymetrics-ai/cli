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
- The filesystem authorization path was removed. The first correction made the role re-derive its
  exact durable fence/turn/role/PID/PGID binding and kernel-stop itself until the controller sent
  `CONT` after bind returned from directory fsync. This revision was intentionally held from parent
  integration until an exact-head adversarial review completed.
- A PID/PGID mismatch now preserves recovery evidence without signalling either untrusted identity.
  The inherited-lock oracle only validates the descriptor/inode; rejection by a fresh controller
  after the original controller and role leader are proven gone establishes that the untouched
  descendant retained the lock.
- GREEN evidence on that intermediate design: kernel-GO adversarial matrix 20/20,
  failed-HALT/instant validator 50/50, shared validator deadline 10/10, two complete 22-scenario
  suites, Phase 0 controls, `make agent-loop-test`, and final `make verify`. All provider processes
  were synthetic and test-owned.

## Final private-capability correction — 2026-07-12

- Exact-head review of `ad8406a8` found two further P1s despite the green intermediate suite. Any
  same-UID process could send `SIGCONT` to the stopped role, and each newly spawned role calculated
  its deadline from a fresh full `TURN_TIMEOUT_SECONDS`; a late validator could therefore outlive
  the one persisted turn deadline. The reviewer dynamically reproduced both before integration.
- RED now sends an external `SIGCONT` in the durable-bind window and then kills the controller. It
  requires that the provider never starts and that the inert role disappears. The existing delayed
  validator case separately proves that orchestrator and validator share one bounded turn budget.
- GREEN replaces signals and published authorization files with a mode-0600 FIFO that the
  controller opens and unlinks before spawning the role. The inherited descriptor is a private,
  one-use kernel capability. The role requires exact canonical binding plus the bounded `GO\n`
  token, then revalidates parent liveness, shared deadline, and binding immediately before `exec`.
  A preseeded path can only make startup fail closed.
- `start_role_group` passes `TURN_DEADLINE_MONO - SECONDS`, never a fresh role timeout. The
  controller publishes GO only after `control_state bind` has returned from file and directory
  fsync and a final fence/deadline check succeeds.
- Final-capability evidence before exact-head review: focused authorization/deadline set passes;
  the post-bind pre-GO oracle sends the formerly exploitable external `SIGCONT`; ten parallel
  repetitions of the four-risk matrix pass 10/10; two complete 22-scenario suites pass; Bash
  syntax, ShellCheck warning, Phase 0, Go unit/race, `make agent-loop-test`, and complete
  `make verify` gates pass. The production Phase 0 fuse remains closed.

## Deep exact-head RED correction — 2026-07-13

- A post-PR session/oracle audit found additional false-green classes after the earlier green
  suite: the controller accepted incomplete or low-score verdicts, trusted a valid verdict from a
  nonzero validator, let the mutable trace script replace the verdict before parsing, published a
  checkpoint despite invalid `RUN.json` or a missing HEAD, and reused a prior run's global
  checkpoint namespace.
- RED: the strict verdict matrix initially produced 15 expected failures across malformed control
  state, aliased state roots, re-anchored deadlines, unbounded readiness/assert helpers, static FIFO
  injection, uncertain terminal commits, and trace deadline escape. After those lifecycle fixes,
  the dedicated verdict/checkpoint RED still failed four assertions: nonzero validator released,
  trace overwrite released, invalid RUN checkpointed, and missing HEAD checkpointed. A separate
  fresh-run RED accepted/restored the prior run's ordinal-only checkpoint.
- Oracle review strengthened the fixtures before GREEN: delayed-begin teardown registers and
  verifies a TERM-resistant descendant; FIFO flood readiness is published only after successful
  writes reach `EAGAIN`; malformed REVERT cases are seeded with a real checkpoint; RETRY must be
  logged as RETRY; checkpoint restore must emit the exact run-bound cleanup record; and symlink and
  hardlink bundles must remain untrusted and unmodified.

## Deep exact-head GREEN correction — 2026-07-13

- Verdict parsing is one bounded, anchored snapshot with duplicate/nonfinite rejection, the exact
  six-field schema, finite 1–5 scores, `PROCEED >= 4`, RETRY 2–<4, fail-closed REVERT <2, nonempty
  control-character-free reason/correction text, null-field semantics, and an exact current-run
  checkpoint target. A nonzero validator cannot contribute a verdict.
- Trace distillation now runs before the Shepherd. Nothing mutable runs between validator
  quiescence and the private snapshot, so the validator judges the full turn and trace cannot
  rewrite its result afterward.
- Checkpoints are namespaced by the current `run_id`. Publication validates bounded strict
  `RUN.json` plus a canonical Git HEAD, creates private no-alias bundle files, fsyncs them, and only
  then atomically publishes that run's `LAST_GOOD`. Restore consumes the snapshotted target,
  revalidates the entire no-symlink/no-hardlink bundle, atomically restores RUN and writes a
  run-bound cleanup record before the turn can complete. Failure enters durable recovery without
  publishing success.
- Signal/HALT handling has a reentrancy latch and ignores repeated termination signals while the
  emergency transition is in flight. The full-gate recipe refuses filtered test execution.
- Focused GREEN: the 12-function verdict/checkpoint/deadline/oracle matrix passes, spanning 23
  isolated fixtures and all new RED cases. Bash syntax, ShellCheck warning severity, and scoped
  diff checks pass. The complete 44-function gate is the next delivery checkpoint.

## Explicit enable blockers retained

- Providers still inherit the worktree lock open-file description and can release it. The trusted
  guardian/launch broker in #342 must be the sole lock owner before Phase 0 can open.
- Startup initialization, process probes, and teardown grace are not yet one absolute monotonic
  end-to-end deadline. #342 owns the bounded guardian/probe/containment closure.
- The shell snapshot does not claim the canonical validation/event/checkpoint transaction or
  durable correction replay. #327/#335 own current-turn validation binding, append-only CAS state,
  and RETRY/REVERT transaction semantics. The Phase 0 first-action fuse remains closed until these
  dependencies and later canary gates pass.

## Final delivery verification — 2026-07-13

- Five consecutive repetitions of the deterministic expired-bind oracle pass after it was changed
  to prove entry into the inert bind-delay window before asserting deadline behavior.
- `make agent-loop-test` passes in 9m09s: Go unit/race/CLI controller tests, Phase 0 controls, and
  all 44 Shepherd supervision functions. The Make target sets `SHEPHERD_REQUIRE_FULL=1`, so a test
  filter cannot produce this green result.
- `make verify` passes in 16m58s: format/tidy, vet, all Go tests, build, connector documentation,
  smoke, lint, 547 connector definitions, race/control gates, Phase 0 controls, and the full
  44-function Shepherd suite.
- Three independent exact-head read-only reviews report no remaining in-scope P0/P1. They
  independently verified process authorization/oracles, verdict/checkpoint durability, and test
  discrimination. The #327/#335/#342 deferrals above remain production-enable blockers rather than
  claims of this closed-fuse issue.

## CI timing-oracle correction — 2026-07-13

- GitHub Verify passed every Go, connector, build, documentation, smoke, lint, and control gate,
  then exposed one runner-speed-dependent assertion in the validator shared-deadline test. The
  controller had no corresponding failure; the oracle compared elapsed time after child startup to
  a fixed 7.5-second threshold instead of the persisted turn deadline.
- The fixture now deliberately consumes most of a 15-second turn before launching a 30-second
  TERM-resistant validator child, snapshots `active_turn.deadline_at`, and requires teardown within
  the actual remaining persisted budget plus five bounded cleanup seconds. It also rejects a
  fixture that did not consume enough budget to distinguish a wrongly refreshed role timeout.
- Corrected focused evidence: 10/10 consecutive repetitions pass with no natural child completion.
  The complete unfiltered `make agent-loop-test` gate also passes in 9m38s. A new CI Verify run is
  required before delivery is green again.

- The next CI run passed the corrected deadline case, then exposed that the unrelated cross-run
  fixture discarded its seed run's status and used the generic 30-second scheduling budget. The
  harness default is now 60 seconds for non-deadline fixtures (deadline tests keep explicit narrow
  limits), and the seed must prove exit 0, released control state, and a published marker while
  retaining bounded diagnostics on failure.
- Corrected cross-run evidence: 10/10 focused repetitions pass, followed by the complete unfiltered
  `make agent-loop-test` gate in 9m44s with all 44 Shepherd functions green.
