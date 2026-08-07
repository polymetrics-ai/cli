# GSD lifecycle trace — cli-top50-sweep-resume4-r1 (zendesk-support)

Branch `fm/cli-top50-sweep-resume2-r1`. Program `cli-top50-fixed-schema-sweep-r1`, landing order #3
under the captain's largest-first reversal, behind github (1220) and workday-rest (907).

Continues `gsd-top50-sweep-resume3-r1.md`. The delivery contract in `AGENTS.md` outranks the
dispatch brief where they differ (captain, 2026-08-07).

## Branch-name correction

The dispatch brief named `fm/cli-top50-sweep-resume3-r1` and said to check it out. **That branch
does not exist** — it has never been pushed. `PROGRESS.md`'s handoff block, which the brief itself
points at as authoritative, names `fm/cli-top50-sweep-resume2-r1` and explains why: history is
continuous across `…-consolidated` → `…-continue-r1` → `…-resume2-r1`, and only the name changed
because each dispatch scaffold names its own branch. Work continued on `resume2-r1`, which carries
the whole sweep. Reported to firstmate rather than left implicit.

## Command provenance

Resolved with `./scripts/gsd sources <command>` (a **Node** script — `bash scripts/gsd` dies with a
syntax error). `plan-phase`, `execute-phase` and `verify-work` all resolve to:

```
.gsd/commands.json
.gsd/upstream.lock.json
.gsd/official-docs/COMMANDS.md
```

Gate: `GSD_BASE_REF=origin/main ./scripts/verify-gsd-workflow` → **exit 0**.

## Lifecycle

| Phase | Where it landed |
| --- | --- |
| `discuss-phase` | `PLAN.md` — baseline analysis, the four judgements, the paging judgement |
| `plan-phase --tdd` | `PLAN.md` slices/hazards, `RUN-STATE.json`, `DERIVED-OPERATIONS.json` at 625 |
| `execute-phase` | `TDD-LEDGER.md` cycles 1–3: red captured → reads → writes |
| `verify-work` | `VERIFICATION.md` — every gate run, real output quoted, golden baseline measured |
| `code-review` | pending; runs under `/no-mistakes` per the firstmate contract |

## Inline execution, recorded as required

Executed **inline** rather than by spawning isolated agents. The canonical parent-worker contract
forbids spawning an orchestrator, shepherd, planner, reviewer, verifier, GSD role, or extra worker
for this job. Recorded here per `AGENTS.md`.

## Required skills

`golang-how-to` routing → `golang-testing` (red-first surface test and its tightening),
`golang-cli` (command surface, flag shapes, group help, bare-namespace behaviour),
`golang-security` (no credential became a flag: `validate_token` stays blocked and
`set_user_password`'s required field is redacted rather than bound),
`golang-structs-interfaces` (the `covered_by` schema and why an operation-backed `direct_write` has
no representation in it).

## What this phase changed about the sweep's shared knowledge

- **Finding 39** — the ledger CAN reconcile. zendesk-support is the first connector in the sweep
  with a zero delta, because the provider publishes a machine-readable spec keyed one operation per
  (method, path) rather than documentation pages. Evidence about this artifact, not the next.
- **Finding 40** — a blanket-blocked inventory is its own failure mode, and it is invisible to a
  count. This connector's 631 rows were complete and correctly counted while 509 of them were
  unreachable. Parity is reachability, not inventory.
- **Finding 41** — `covered_by.write` must name a writes.json ACTION; an operation-backed
  `direct_write` has no `covered_by` representation at all. Discovered only because the write pass
  was validated; it decides the whole shape of a mutation slice.
- **Finding 42** — `covered_by.writes` (plural) used for its stated purpose for the first time, on
  two real union endpoints. github's foundation fix generalised.
- **Finding 43** — a `oneOf` is not always a union. Distinguish label-only arms, type variants, and
  genuinely distinct contracts mechanically; conflating them either ships a lie or blocks a
  reachable operation.
- **Finding 44** — guard the BUILT artifact, not its inputs. An input-shaped empty-contract guard
  passed three operations whose declared body schema had no properties.
- **Finding 45** — the binary trap can run backwards. A non-JSON success response on a PUT is a
  representation, not a download.
