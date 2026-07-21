# Prompt and Worker Contracts

Phase: `471-pi-agent-session-shepherd`

## 2026-07-21 autonomous replacement pivot

```text
Treat the standalone Go/tmux Shepherd as abandoned. Build #471 as the authoritative, complete
in-process Pi AgentSession Shepherd. It owns research, GSD parent/sub-issue planning, dependency
scheduling, isolated implementation, RED/GREEN/refactor, verification, review/correction,
sub-PR integration, recovery, durable human decisions, and exact-head parent merge after explicit
human approval. The existing read-only code is only the control-plane seed.
```

## Parent orchestrator instruction

```text
Read #471, PR #472, AGENTS.md, the parent orchestrator contract, and PLAN.md. Reconcile disk, Git,
and GitHub before action. Build the ready queue. Dispatch every dependency-ready issue whose write
scope is disjoint, each in its own branch/worktree. Record spawned or one exact not_spawned reason.
Do not infer completion from agent prose. Keep #472 draft until all child, verification, review,
canary, and exact-head human gates pass.
```

## Mutating worker handoff template

Every in-process implementation/correction session receives all of the following:

- objective and acceptance criteria for exactly one child issue;
- parent issue #471, parent branch, and PR #472;
- owned issue branch, canonical isolated worktree, and PR base;
- exact allowed write scope and forbidden shared files;
- required GSD/TDD workflow, skills, verification commands, and handoff schema;
- exact model route `openai-codex/gpt-5.6-sol`/`high`;
- bounded workspace tools and typed host operations only; and
- hard stops for secrets, authority expansion, dependency/scope changes, destructive actions,
  quality-gate reduction, and default-branch mutation.

Planning, research, issue proposal, verification, review, disposition, and orchestration sessions
use `openai-codex/gpt-5.6-sol`/`xhigh` and cannot mutate outside their explicit role.

## Human decision request template

```text
<!-- pm-shepherd-decision:<request-id> -->
@karthik-sivadas Shepherd requires a human decision.
Target: <issue-or-pr>
Generation: <generation>
Head: <sha-or-not-applicable>
Reason: <bounded evidence-backed reason>
Options: <explicit options>
Reply here with: /shepherd decide <request-id> <option>
```

The broker must create this comment idempotently, accept only the configured human on the bound
target/current generation/head, persist the source URL and actor, consume once, and revalidate
before resuming.

## Historical foundation prompts

The earlier read-only scout/validator and exact-head #438 canary prompts remain historical TDD
evidence in the trace directory. They do not define the replacement's final feature scope.
