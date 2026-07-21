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

For #471, automated quality review specifically means an independent controller-owned Codex
`gpt-5.6-sol`/`xhigh` AgentSession bound to the exact base/head/range. Claude and Copilot are
intentionally skipped for this program and must not be claimed in evidence. Any head change
invalidates the record and requires a fresh stable-head review; parent merge remains a separate
exact-head human decision.

## Shepherd sub-worker verification boundary

Every Shepherd implementation, correction, and independent-review prompt must keep local
verification proportional to that TypeScript lane:

1. run the child issue's focused RED/GREEN tests;
2. run the complete `.pi/extensions/shepherd/*.test.ts` suite;
3. run strict no-emit TypeScript against the repository-pinned Pi `0.80.6` declarations;
4. run the Shepherd extension through offline Pi RPC; and
5. run `git diff --check` plus changed-path/write-scope checks.

Do **not** run `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, connector build/certification,
or `make verify` in a Shepherd sub-worker. Those repository-wide gates run once on the integrated
parent head and independently in GitHub CI. A broad gate already started by a child is recorded as
`superseded/cancelled_by_parent_policy`; it is neither a functional pass nor failure. This boundary
does not weaken the final parent gate and must be copied verbatim into every dispatched child or
correction prompt.

## Stable-head review convergence campaign

Security-critical Shepherd changes, changes containing two independent state machines, or a
candidate above 500 changed production lines use a stable-head review campaign instead of one
generalist first-blocker pass:

1. freeze one exact base/head range and do not dispatch a correction after the first BLOCKED result;
2. reserve review capacity by pausing unrelated implementation lanes when necessary;
3. dispatch independent `openai-codex/gpt-5.6-sol`/`xhigh` specialists for lifecycle/resource
   ownership, security/parser/consumer behavior, and complexity/API/test quality as applicable;
4. require every specialist to continue after P0/P1 findings, inspect every relevant production
   function, and return both all findings and a completed invariant coverage ledger;
5. let the parent synthesize and disposition every result only after all specialists finish against
   the same frozen head;
6. commit one behavior-level test-only RED checkpoint covering the complete synthesized finding
   set before any production correction; and
7. rerun the full stable-head campaign after a correction changes a state machine, parser, authority
   boundary, or other architecture—not merely the latest delta.

A missing-module, compile-only, or file-level failure does not establish RED for a broad feature.
Create the smallest compiling/throwing scaffold needed for every planned behavior test to execute
and fail for its intended assertion. Lifecycle work must define a phase × trigger × hook-outcome
matrix with timer/handle accounting. Redaction work must define syntax × key × nesting × consumer ×
preservation cases plus deterministic work metrics. Keep code review and planning/evidence review
as separate specialist responsibilities so repetitive artifact diffs do not consume security
review context.

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
