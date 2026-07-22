# Phase 479: production Shepherd matrix

## Objective

Deliver one in-process `/pm-shepherd` production path that owns reviewed schema-2 intake,
dependency-aware parallel child work, isolated worktrees, fixed bounded verification, commit/push
and stacked PR publication, independent exact-head review/correction, child integration, crash-safe
resume, and an exact-head parent human wait. Shepherd must never merge the parent PR into the
default branch and must complete only after observing the human-owned merge.

The hardened schema-1 read-only canary remains a separate explicit command. Production plan input
is schema version 2; the persisted production runtime DTO intentionally remains schema version 1
with kind `production_autonomous`. “Schema-v2 status/state” is not an accurate description.

## Delivery and process record

- Active issue: #479; parent issue: #471; parent branch:
  `feat/471-pi-agent-session-shepherd`; implementation branch:
  `feat/479-shepherd-production-matrix`.
- Never push or merge `main`. Child integration targets only the non-default parent branch. The
  final parent merge remains a human gate.
- Implementation followed the repository's documented manual-GSD fallback. `scripts/gsd doctor`
  passed and listed 69 commands, but `scripts/gsd prompt programming-loop ...` was unavailable when
  implementation began. This exception does not waive TDD; the evidence gap is recorded honestly
  in `TDD-LEDGER.md`.
- Documentation used the available adapter path:
  `scripts/gsd prompt docs-update "Update only .pi/README.md and
  .planning/phases/479-shepherd-production-matrix artifacts for schema-v2 production Shepherd;
  verify every claim against code and tests; preserve all other documentation."`
- Required skills recorded for implementation/review:
  `gsd-programming-loop`, `gsd-workstreams`, `github-issue-first-delivery`,
  `architecture-patterns`, and `javascript-testing-patterns`. Documentation loaded
  `golang-documentation`, `golang-cli`, and `golang-security` as routed for README/command/auth and
  filesystem behavior.
- CLI help/manual/website parity: `/pm-shepherd` is a Pi extension command, not a Go `pm` command.
  `.pi/README.md`, bare/help/status behavior, extension registration, and tests are applicable;
  `docs/cli/**`, website pages, and generated `pm` manual artifacts are not applicable to this
  change.

## Frozen runtime contract

```text
Pi /pm-shepherd
  -> production extension composition
     -> canonical repository/worktree and immutable plan binding
     -> ProductionRepositoryPlanIntake (schema 2 + canonical digest)
     -> ProductionShepherdController (one generation + CAS state)
        -> ProductionEffectRecoveryBarrier before scheduling
        -> deterministic scheduler (dependencies, capacity, canonical scope collisions)
        -> ProductionChildPipeline
           -> ProductionWorkspaceLifecycle + typed GitAdapter
           -> embedded implementation/correction AgentSession roles
           -> BoundedVerificationRunner (configured executable + argv, never a shell)
           -> typed commit/push and authoritative timeout reconciliation
           -> GitHub child issue/PR, exact-head review, dispositions, integration
        -> ProductionParentFinalizer (receipts, CI, clean exact-head review, draft -> ready)
        -> ProductionParentGateAdapter (exact human request + authoritative merge observation)
```

The model receives bounded scoped workspace capabilities, not generic shell, Git, GitHub, or
merge-main tools. External effects are typed, cancellable, idempotency-keyed, journaled, and
reconciled against authoritative Git/GitHub state. The parent workflow must already have created the
marker-bound parent PR; absence or ambiguity fails closed.

## Workstream ownership used

| Lane | Exclusive ownership | Contract consumed |
|---|---|---|
| A: durable autonomy | production contract, state, effect journal, recovery, scheduler | frozen production DTOs and CAS fences |
| B: workspace/verification | bounded verification and production workspace/Git lifecycle | state/effect ports and existing typed Git/workspace adapters |
| C: GitHub/human/review | bounded GitHub transport, evidence, review, decisions, parent lifecycle | orchestration and durable authority ports |
| D: composition/docs | child pipeline, controller/runtime/index/extension integration and phase docs | exported A–C ports |

## Acceptance matrix

The authoritative wording is the issue #479 preflight matrix. Each final disposition must cite a
machine test or a precise open mechanism; passing a helper unit test alone is not enough when the
production composition does not use that helper.

| # | Required behavior | Primary final-head evidence source | Disposition |
|---:|---|---|---|
| 1 | Intake → parallel children → PR → review/correction → integration → exact parent wait | production runtime/controller/child-pipeline/parent-lifecycle trajectory tests | Pending final-head freeze |
| 2 | Dependencies, canonical collisions, cap, deterministic idle | `production-scheduler.test.ts`; controller trajectory | Pending final-head freeze |
| 3 | Disjoint mutators coexist; collisions denied; own-lease release | workspace lifecycle and AgentSession lease tests | Pending final-head freeze |
| 4 | Parent movement refresh/rebase/reclaim then reverify/rereview | controller stale-parent and real workspace refresh tests | Pending final-head freeze |
| 5 | Per-stage budgets, durable exhaustion wait, no prose success | production state, controller, and child-pipeline tests | Pending final-head freeze |
| 6 | Crash-safe prepare/publish/observe/consume/apply/persist without duplicates | effect journal, recovery, runtime composition, pipeline, and parent-gate tests | Pending final-head freeze |
| 7 | Human replies fail closed unless exact, current, unique, unedited, allowlisted | GitHub decision broker and production human-gate tests | Pending final-head freeze |
| 8 | Stop at every stage aborts and joins before stopped state/lease release | production controller, parent lifecycle, extension, and runtime cancellation tests | Pending final-head freeze |
| 9 | Stop/shutdown race, stale generation, sibling abort, persistence failure | controller/state/AgentSession/extension race tests | Pending final-head freeze |
| 10 | Timeout-after-publication reconciliation for commit/push/PR/integration | workspace lifecycle, child pipeline, GitHub orchestrator tests | Pending final-head freeze |
| 11 | Stable resume ownership and typed durable refresh/reclaim | production state/workspace/controller/finalizer tests | Pending final-head freeze |
| 12 | Findings/dispositions and fresh clean exact-head rereview after movement | production review adapter, pipeline, and controller tests | Pending final-head freeze |
| 13 | Dirty/scope/wrong coordinates/draft/untrusted CI/prose fail closed | workspace handoff, child pipeline, GitHub evidence, parent finalizer tests | Pending final-head freeze |
| 14 | Parent head invalidates gate; no main merge; completion only after observed merge | parent lifecycle, controller, pipeline, and extension tests | Pending final-head freeze |
| 15 | Hostile shapes/paths/controls and bounded argv/output/timeout/cancel | contract/intake/verification/tool/Git/GitHub tests | Pending final-head freeze |
| 16 | Bare/help/invalid/status and unresolved-initialization stop | argument and extension tests plus offline RPC | Pending final-head freeze |
| 17 | Top-level read-only rejected; internal read-only never integrates | contract/intake/orchestrator/AgentSession/tool-policy tests | Pending final-head freeze |

The final status and exact open rows are frozen in `VERIFICATION.md` and projected into
`RUN-STATE.json`; this plan must not be interpreted as a claim that pending rows passed.

## Checkpoints

1. [x] Plan, contracts, intake, scheduler, durable state/effect/recovery primitives.
2. [x] Isolated workspace lifecycle, bounded verification, typed Git and GitHub adapters.
3. [x] Controller, child pipeline, review/correction, integration, and parent lifecycle candidate.
4. [ ] Final production runtime/index composition at one stable head.
5. [ ] Focused matrix, full Shepherd suite, strict TypeScript, offline Pi RPC, docs/diff/scope gates.
6. [ ] One consolidated stable-head blocker review and one correction pass when required.
7. [ ] Commit the exact evidence artifacts; leave parent/default-branch integration human-gated.

## Proportional verification commands

```bash
node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts
node /Users/karthiksivadas/.npm/_npx/a322a253dbd59f36/node_modules/typescript/lib/tsc.js \
  --noEmit --strict --target ES2024 --module NodeNext --moduleResolution NodeNext \
  --allowImportingTsExtensions --skipLibCheck \
  --baseUrl /Users/karthiksivadas/.nvm/versions/node/v24.13.1/lib/node_modules \
  --typeRoots /Users/karthiksivadas/.nvm/versions/node/v24.13.1/lib/node_modules/@earendil-works/pi-coding-agent/node_modules/@types \
  .pi/extensions/shepherd/*.ts
printf '{"id":"commands","type":"get_commands"}\n' | \
  PI_OFFLINE=1 PI_CODING_AGENT_DIR=/tmp/pm-shepherd-rpc \
  /Users/karthiksivadas/.nvm/versions/node/v24.13.1/bin/pi \
  --mode rpc --no-session --approve --no-extensions --no-skills \
  --no-prompt-templates --no-context-files -e .pi/extensions/shepherd/index.ts
git diff --check
```

Do not run broad Go/connector certification in this TypeScript/Pi issue slice. Live GitHub checks
are optional and use only an explicitly designated sandbox with healthy ambient `gh` authentication;
tokens are never printed, stored, or passed in prompts.
