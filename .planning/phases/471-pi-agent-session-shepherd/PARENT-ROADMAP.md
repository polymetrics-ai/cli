# Autonomous In-Process Shepherd Parent Roadmap

## Objective

Replace the abandoned standalone Go/tmux Shepherd program with a first-class Pi extension that
owns the complete issue-to-merge delivery loop. Shepherd must decompose a parent objective into
small issue-scoped workstreams, schedule independent work in parallel and dependent work in order,
run bounded Pi `AgentSession` workers inside the current Pi process, verify and review every result,
correct failures, integrate eligible sub-PRs, and remain active until the parent work is complete or
a genuine human decision is required.

This issue is the authoritative parent for the replacement. The existing in-process read-only
implementation is a control-plane and safety foundation, not the release boundary. The abandoned
Go path is superseded, not completed.

Supersedes #372, #389, and #470.
Consumes the issue-first/GSD contracts established by #371.
First end-to-end consumer: #397 and draft PR #438 (CLI Architecture v2).

## Background and architecture decision

- Use the Pi 0.80.6 public `createAgentSession` API. Do not launch a second `pi` process and do not
  use tmux as the orchestration transport.
- Keep the framework-independent controller behind explicit ports for policy, workers, state,
  workspaces/Git, GitHub, evidence, review, decisions, and time.
- Mutating workers receive one issue, one branch, one isolated worktree, one declared write scope,
  one PR base, and bounded tools. They never share the coordinator checkout.
- Implementation workers use `openai-codex/gpt-5.6-sol` with `high`; planning, research, review,
  validation, and orchestration workers use the same model with `xhigh`. No 5.5 route is allowed.
- Agent prose is untrusted evidence. Git, GitHub, test, CI, review, and persisted state are the
  authoritative sources for transitions.
- Persist bounded redacted state and an append-only decision/audit trail. Never persist prompts,
  chain-of-thought, credentials, secret values, or unrestricted command output.
- The macOS threat boundary is trusted same-user local automation using a private owned state root;
  do not claim protection from a hostile same-UID process without native descriptor-relative I/O.
- Retain the legacy shell loop only as a documented rollback path until the replacement canary and
  cutover pass. The standalone Go program is not a fallback.

## End-to-end state machine

`INTAKE -> RESEARCH -> PARENT_PLAN -> ISSUE_CREATE -> PARENT_SETUP -> SCHEDULE -> EXECUTE -> VERIFY
-> REVIEW -> CORRECT (when needed) -> INTEGRATE -> FINAL_VERIFY -> HUMAN_DECISION -> MERGE -> COMPLETE`

The scheduler derives a dependency DAG and bounded ready queue. Independent, non-colliding issues
may run concurrently; dependency or write-scope collisions serialize automatically. Restarts
reconcile persisted intent with current Git/GitHub truth before any new mutation.

## Parent delivery topology

- Parent issue: #471
- Parent branch: `feat/471-pi-agent-session-shepherd`
- Parent PR: #472 (`feat/471-pi-agent-session-shepherd` -> `main`, draft)
- Sub-PR base: `feat/471-pi-agent-session-shepherd`
- Closing policy: sub-PRs use `Refs #<child>` and `Refs #471`; only the final parent PR closes #471.

Planned dependency waves:

1. Harden the durable control-plane foundation already present on the parent branch.
2. In parallel, implement the dependency policy/reconciler, scoped in-process worker runtime,
   isolated workspace/Git adapter, and durable GitHub human-decision broker.
3. Implement parent/sub-issue/stacked-PR and automated-review orchestration on those ports.
4. Integrate the autonomous parallel scheduler, correction loop, and command surface.
5. Prove crash recovery, auditability, operator UX, and legacy-shell cutover.
6. Run an end-to-end canary against CLI Architecture v2 and finish exact-head parent verification.

## Sub-issue roster and dependency queue

| Wave | Issue | Branch | Depends on | Primary scope | Initial state |
|---|---|---|---|---|---|
| 1 | #473 | `feat/473-shepherd-control-plane-foundation` | none | durable control plane and adversarial hardening | `worker_ready` |
| 2 | #474 | `feat/474-shepherd-dependency-policy` | #473 | pure DAG/policy/reconciler | `planned` |
| 2 | #475 | `feat/475-shepherd-agent-session-runtime` | #473 | scoped in-process worker runtime | `planned` |
| 2 | #476 | `feat/476-shepherd-worktree-git-adapter` | #473 | isolated worktree and typed Git operations | `planned` |
| 2 | #477 | `feat/477-shepherd-github-decision-broker` | #473 | durable authenticated human decisions | `planned` |
| 3 | #478 | `feat/478-shepherd-github-parent-orchestration` | #474, #476, #477 | parent/sub-issue/PR/review orchestration | `planned` |
| 4 | #479 | `feat/479-shepherd-autonomous-controller` | #474-#478 | scheduler/controller/command integration | `planned` |
| 5 | #480 | `feat/480-shepherd-recovery-cutover` | #479 | recovery, auditability, operator UX, cutover | `planned` |
| 6 | #481 | `test/481-shepherd-cli-architecture-canary` | #480 | #397/#438 end-to-end canary and final evidence | `planned` |

Each issue body owns its exact write scope, required skills, verification, and human gates. #474,
#475, #476, and #477 are intentionally disjoint and launch in parallel after #473 is integrated.
#478 waits for the policy, Git, and decision ports; #479 is the deliberate shared wiring point.

## Human-decision contract

Shepherd proceeds without prompting for ordinary, reversible, in-scope delivery actions. It pauses
only for a repository-defined human gate or a decision that changes authority, scope, security,
cost, external production state, or the parent merge.

- Use the parent issue for requirements, scope, authority, or dependency decisions.
- Use the relevant PR for head-specific review, exception, or merge decisions.
- Post exactly one idempotent request containing a durable request ID, permitted options, target
  issue/PR, generation, and exact head SHA when applicable; mention the configured human.
- Accept only an allowlisted human response using
  `/shepherd decide <request-id> <option>` on the bound issue/PR and current generation/head.
- Consume a decision once, persist its actor and source URL, revalidate all affected gates, resume,
  and record the outcome. Stale, edited, duplicate, bot-authored, or ambiguous replies fail closed.
- Shepherd must never infer approval from silence, emoji, review text, an agent score, or CI success.
- A parent PR may merge to `main` only after an explicit fresh human `approve-merge` decision for
  the exact verified head. Shepherd revalidates immediately before merge; otherwise it waits.

## Acceptance criteria

- [ ] `/pm-shepherd start --issue <parent>` can create or reconcile the parent plan, child roster,
      dependency graph, parent branch, and draft parent PR.
- [ ] Ready work is dispatched in parallel only when dependencies, write scopes, worktree isolation,
      review policy, and configured concurrency permit it.
- [ ] Every mutating worker is an in-process Pi `AgentSession` with exact model/thinking routing and
      least-authority, workspace-bounded tools; there is no Pi subprocess or tmux requirement.
- [ ] Every child follows GSD planning and red-green-refactor, opens a scoped sub-PR against the
      parent branch, records exact verification, and receives required automated/human review
      coverage before integration.
- [ ] Failed tests, review findings, stale evidence, conflicts, or CI failures enter a bounded
      correction/retry loop and never become success by score alone.
- [ ] Human gates use the durable authenticated comment protocol, survive restart, wait without
      busy-looping, and resume exactly once after a valid response.
- [ ] State ownership, cancellation, lease rollover, crash recovery, root identity, and exact-head
      evidence are adversarially tested; impossible persisted states fail closed.
- [ ] `status`, `stop`, and `resume` are deterministic and explain the current stage, ready/running/
      blocked lanes, dependency reason, verification/review state, and pending human action.
- [ ] No secret value is requested, printed, summarized, stored, or passed in prompts. GitHub auth
      comes from the existing host environment/keychain and is exposed only to typed host actions.
- [ ] The full local gates, Pi extension smoke, restart tests, GitHub sandbox tests, and a canary
      against #397/PR #438 pass with an auditable trace.
- [ ] The legacy shell Shepherd is marked deprecated only after the canary passes; abandoned Go
      issues/PRs are closed with cross-links and their branches are retained as history.
- [ ] The parent PR remains draft until every required child is integrated, exact-head verification
      and review coverage are clean, and the only remaining gate is explicit human merge approval.

## Required workflow and skills

- Parent issue orchestrator contract and stacked parent/sub-issue workflow.
- Repo-local GSD programming loop for every behavior change, with the manual fallback recorded only
  when the adapter is unavailable.
- `gsd-workstreams`, `gsd-plan-phase`, `github-issue-first-delivery`,
  `architecture-patterns`, and `javascript-testing-patterns`.
- Task-specific Pi/TypeScript, security, CLI help/docs parity, and design skills as routed by
  `.agents/agentic-delivery/references/required-skills-routing.md`.

## Verification

```bash
node --test .pi/extensions/shepherd/*.test.ts
node --test .pi/extensions/shepherd/**/*.test.ts
pi --list-extensions
git diff --check
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Runtime/GitHub integration tests must use isolated fixtures or a designated sandbox/canary target,
must not print tokens, and must not merge a parent PR without the exact human gate above.

## Source links

- Parent orchestrator contract: `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- Parent orchestration loop: `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- Autonomous stage model: `.agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md`
- Universal GSD policy: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- Pi adapter: `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- Worker handoff: `.agents/agentic-delivery/contracts/worker-handoff-template.md`
