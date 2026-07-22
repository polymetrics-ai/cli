# Issue #479 Shepherd MVP plan

## Outcome

Deliver one usable in-process autonomous Shepherd vertical slice on top of the exact green #475 and
#478 checkpoints. `/pm-shepherd start` must schedule independent child work in parallel, honor
dependencies and write-scope collisions, persist operator-visible state, support status/stop/resume,
and stop at a durable human parent-merge gate. It must never merge to `main`.

Base: `c958d7e608eac4afedceea953408a8fd4cfcb566`

Integrated dependencies:

- #475 `4272f05784e08f6cbc6961a600d302f4077eff41`
- #478 `4b8caaf2d0492e4ab10185f366cbb98c33fc21df`

## Bounded delivery contract

1. Add one fake-port behavior trajectory before production code.
2. Implement the smallest v2 state, store, scheduler, and injected lifecycle ports needed by it.
3. Route autonomous start/resume through the new controller; retain the v1 read-only canary.
4. Verify focused behavior, strict TypeScript, offline Pi registration, and a deterministic local
   canary.
5. Run exactly one blocker-only review. Fix only defects that prevent the trajectory, command load,
   stop/resume, durable human wait, or no-main-merge boundary.

Deferred to #480/backlog: exhaustive crash-boundary journaling, multi-review quorum, exhaustive CAS
and hostile-object hardening, stale-sibling rebase/reclaim, and every edge permutation in the old
17-part preflight matrix. Deferred work cannot reopen this MVP after its acceptance checks pass.

## Runtime and skills

The repo-local GSD adapter passed `scripts/gsd doctor` but does not expose `programming-loop`
(`unknown GSD command`), so this issue uses the documented manual-GSD fallback. Skills/references:
`gsd-programming-loop`, `gsd-workstreams`, `architecture-patterns`,
`javascript-testing-patterns`, required-skills routing, parent-orchestrator contracts, Pi adapter,
runtime integration, CLI help/docs parity, and the universal runtime loop.

Implementation uses `gpt-5.6-sol/high`; orchestration and the single review use
`gpt-5.6-sol/xhigh`. The planning cycle used three read-only mapping agents and one owning
integrator, with no overlapping edits.

## Acceptance checks

- [x] Autonomous start needs only `--issue`; canary still requires explicit read-only acknowledgement.
- [x] Two independent non-colliding child jobs overlap at concurrency two; a dependent child waits.
- [x] Each child passes through execute, verify, review, and integrate ports.
- [x] State is persisted after meaningful transitions and rendered by status.
- [x] Stop joins accepted work and persists stopped; resume continues unfinished work.
- [x] Completion of child work becomes `waiting_human` at `parent_merge`.
- [x] The controller exposes no parent-to-main merge operation.

Pi slash-command help and `.pi/README.md` are applicable. Go `pm help`, generated manuals, and the
website are not applicable because this is a project-local Pi extension, not a Go CLI command.
