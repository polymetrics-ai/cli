# Plan: Fence and Supervise the Existing Shepherd Controller

Issue: #326  
Parent: #323 / PR #324  
Branch: `fix/326-shepherd-supervision`  
Base: `fix/323-auto-loop-hardening` at `aef7fe86`

## Outcome

Harden `scripts/pi-shepherd-loop.sh` in place against concurrent controllers, unbounded
mega-turns, leader-exit orphans, surviving descendants, and non-durable HALT. The launcher remains
the only supervised entrypoint and the Phase 0 migration fuse remains closed in production.

The unpublished `fix/326-controller-fencing` experiment is excluded: it is not wired into the real
launcher, is substantially larger than this issue, and has a known local-expiry race. It remains
parked for the separately approved cleanup process.

## Test-first slices

1. Add a sandbox harness that copies the real launcher, mechanically removes only the marked Phase
   0 guard in the temporary copy, and uses fake local Pi processes. Production has no bypass.
2. RED: prove 32-way duplicate starts, hard-deadline descendant cleanup, leader-exit orphaning,
   durable HALT, HALT resume denial, signal cleanup, fence movement, and restart-persistent turn
   limits fail with the current launcher.
3. Add an inherited nonblocking worktree lock, atomic `CONTROL.json`, exact controller fence,
   persisted limits/turn identity, and bounded role process-group supervision in the existing file.
4. Persist HALT/recovery before teardown and deny unsafe resume before prompt/provider access.
5. Refactor only after the focused suite, Phase 0 harness, repeated concurrency checks, and full
   repository gates pass.

## Boundaries

- Production: `scripts/pi-shepherd-loop.sh` and the `agent-loop-test` Make recipe only.
- Tests: `scripts/tests/pi-shepherd-supervision.sh`.
- Memory: `.planning/phases/326-shepherd-supervision/**`.
- Bash plus Python standard library only; no dependency or alternate executable.
- No live enable route, provider call, connector/GitHub mutation, takeover, signed human resume,
  validator transaction, worker capability broker, merge broker, or parent-to-main merge.

## Checkpoints

1. Plan and RED tests.
2. Minimal GREEN implementation.
3. Review fixes and full verification.
4. Push/open stacked PR only after coherent green evidence.

