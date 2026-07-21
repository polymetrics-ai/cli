# Plan: Autonomous In-Process Shepherd

## Objective

Deliver the complete issue-to-human-merge-readiness Shepherd described by #471 through draft parent
PR #472. The
implementation must execute as bounded Pi `AgentSession` children inside the active Pi process,
parallelize independent work, serialize dependencies/collisions, and wait for exact authenticated
human decisions without relying on the abandoned Go program or an external shell driver.

## Required workflow and skills

- Parent issue orchestrator contract and stacked parent/sub-issue workflow.
- `gsd-programming-loop` (manual fallback recorded while the adapter command is unavailable).
- `gsd-workstreams`, `gsd-plan-phase`, `github-issue-first-delivery`,
  `architecture-patterns`, and `javascript-testing-patterns`.
- Task-specific security, Git, CLI help/docs parity, and end-to-end testing skills routed by
  `.agents/agentic-delivery/references/required-skills-routing.md`.

Implementation workers run `openai-codex/gpt-5.6-sol`/`high`; planning, research, orchestration,
verification, review, and disposition run `openai-codex/gpt-5.6-sol`/`xhigh`.

## Architecture

The deterministic controller depends on explicit ports:

1. durable state/lease and target-evidence foundation;
2. pure dependency DAG, stage policy, ready queue, retry budget, and reconciler;
3. scoped in-process `AgentSession` role runtime and tool policy;
4. isolated worktree plus typed Git operations;
5. durable GitHub human-decision broker;
6. parent issue/sub-issue/stacked-PR/review orchestration;
7. autonomous controller, v2 state/effect journal, typed parent refresh/rebase, concrete host
   adapters, and `/pm-shepherd` command integration; and
8. crash recovery, bounded audit log, operator UX, reversible cutover preparation, and a canary
   harness whose successful result is required before deprecation activation.

Target stage machine:

```text
INTAKE -> RESEARCH -> PARENT_PLAN -> ISSUE_CREATE -> PARENT_SETUP -> SCHEDULE
       -> EXECUTE -> VERIFY -> REVIEW -> CORRECT -> INTEGRATE
       -> FINAL_VERIFY -> HUMAN_DECISION -> MERGE (observer-only) -> COMPLETE
```

Every stage transition is host-validated and persisted. A wake begins by reconciling with current
Git/GitHub truth. A correction loop is bounded; exhaustion becomes a durable human decision rather
than an infinite retry.

## Dependency waves

| Wave | Issue | Scope | Dependencies | Parallel rule |
|---|---|---|---|---|
| 1 | #473 | control-plane/lease/state/lifecycle hardening | none | critical path now |
| 2 | #474 | dependency policy/reconciler | #473 | parallel with #475-#477 |
| 2 | #475 | in-process role runtime/tool policy | #473 | parallel with #474/#476/#477 |
| 2 | #476 | worktree and typed Git adapter | #473 | parallel with #474/#475/#477 |
| 2 | #477 | GitHub human-decision broker | #473 | parallel with #474-#476 |
| 3 | #478 | parent/sub-issue/PR/review orchestration | #474/#476/#477 | after ports are stable |
| 4 | #479 | shared autonomous controller and command wiring | #474-#478 | deliberate integration point |
| 5 | #480 | recovery, audit, operator UX, reversible legacy-cutover preparation | #479 | sequential safety gate |
| 6 | #481 | #397/#438 canary, post-pass deprecation activation, and final evidence | #480 | final validation and activation |

Each mutating child uses its issue branch/worktree and targets the parent branch. Child PRs use
`Refs #<child>` and `Refs #471`; only #472 closes #471. #474-#477 have disjoint file ownership and
must be dispatched together after #473 is reviewed and integrated.

## TDD execution contract

For each child:

1. plan and seed the verification/TDD ledger before production edits;
2. capture deterministic RED evidence for behavior, authority, race, or failure handling;
3. implement the smallest GREEN slice and commit/push coherent checkpoints;
4. refactor only while focused tests remain green;
5. open/update a child PR against the parent branch;
6. run issue gates, CI, adversarial review, and written disposition;
7. correct accepted findings test-first and re-review the exact head; and
8. integrate only after scope, verification, and review evidence meet the parent contract.

## Human decision and merge contract

- Requirement/scope/authority questions are posted on #471; head-specific review/merge questions
  are posted on #472 or the relevant child PR.
- The comment contains one durable marker, request ID, exact target/generation/head, allowed
  options, and `/shepherd decide <request-id> <option>` syntax.
- Only an allowlisted human answer on the bound target can be consumed, once.
- Shepherd revalidates immediately after the decision. A changed head invalidates approval and
  creates a new request.
- No direct push to `main` is allowed, and no Shepherd port may merge the parent PR. After the fresh
  exact-head human `approve-merge` decision and all repository gates, Shepherd records readiness,
  waits for a human-owned merge, and reaches `COMPLETE` only after authoritative GitHub/default-
  branch observation proves that exact merge.

## Current critical path

#473, #474, #476, and #477 are integrated. #475 is completing one consolidated stable-head
RED/GREEN correction, including workspace-keyed disjoint mutator reservations. #478 is in a fresh
exact-head independent-review campaign after its consolidated correction. Once both are clean and
integrated, execute #479 using the preflight scope and RED matrix recorded in
`traces/479-preflight-interface-audit.md`; do not begin production wiring against the old read-only
v1 DTO or immutable initial-parent-base assumptions.

## Verification

```bash
node --test .pi/extensions/shepherd/*.test.ts
pi --list-extensions
git diff --check
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Additional child-specific integration/fault-injection commands are recorded in issue bodies and
verification artifacts. The overall phase remains unverified until #481 and exact-head parent
review finish.
