# GPT-5.5 Implementation Prompt — Issue #389

You are the implementation worker for GitHub issue #389 on branch
`feat/389-autonomous-shepherd`, stacked on parent branch `feat/372-gsd-pi-go-shepherd` and parent
PR #390. Work only in this checkout. Do not commit, push, create/update GitHub objects, read
credentials, or merge anything. The final `main` merge is human-gated.

## Objective

Finish the existing test-first Go implementation so this command is the one issue-scoped entry
point after validated intake:

```text
shepherd supervise --config <absolute-config> --issue <N> --context <validated-context.json>
```

It must continuously choose only the canonical GSD unit and stop only at a typed blocker, human
decision, or final parent-PR merge gate. Preserve the standalone nested Go module and official
local GSD Pi 1.11.0. Do not add dependencies or expand Podman; host/local GSD is the target.

## Mandatory reading

Read completely before editing:

- `AGENTS.md`
- `.planning/phases/issue-389-shepherd-hardening/{PLAN,TDD-LEDGER,VERIFICATION}.md`
- `.gsd/{PROJECT,REQUIREMENTS,PREFERENCES}.md`
- `agent-runtime/shepherd/README.md`
- current diffs under `agent-runtime/shepherd/**`

Load/apply the repo-local Go implementation rules and the Go testing, context, concurrency, error,
safety, observability, design, and interfaces guidance named in the plan. Follow strict RED → GREEN
→ refactor and update the TDD ledger with commands and concise evidence.

## Current interrupted state — preserve and complete it

The current uncommitted changes are intentional, partially implemented RED/GREEN work. Do not
discard or replace them wholesale.

Current failing command:

```text
cd agent-runtime/shepherd && go test ./...
```

Known failures:

1. `cmd/shepherd/main.go` references missing `classifyUnitFailure` and
   `isAutomaticallyRetryable` helpers.
2. `TestUnitAttemptBudgetSurvivesStoreRestart` creates a unit-attempt row before its referenced
   delivery exists, so its fixture must initialize the canonical delivery first.

## Required changes

### 1. Runtime prompt/tool admission

- Keep the new qualified GSD Pi 1.11.0 package + active-cache prompt compatibility logic.
- Enforce `advertised GSD tools ⊆ allowed GSD tools` before Pi starts.
- Fail closed with `runtime_contract_mismatch` for partial, symlinked, unknown, or mismatched
  runtime resource shapes.
- Preserve and extend the focused tests; do not loosen exact version validation.

### 2. One canonical GSD project per main issue

- Finish immutable persistence for issue, parent issue, branch, base, worktree, GSD project root,
  initial SHA, context hash, and GSD version.
- The same issue/exact identity must be idempotently adoptable after restart.
- Two issues must never share one project root/controller identity.
- Finish atomic `.gsd/ISSUE.json` bootstrap and missing issue-local PROJECT/REQUIREMENTS/PREFERENCES
  creation without overwriting an existing differently bound project.
- Keep Shepherd SQLite as controller truth and native GSD SQLite as workflow truth.
- Explicitly set `GSD_PROJECT_ROOT` for local runs.

### 3. Durable attempts, signals, nested-agent lifecycle, and visibility

- Finish durable per `{delivery,generation,unit,head}` attempt reservation and exhaustion across
  database reopen/restart. Default max attempts is 3 and config must remain bounded.
- Add typed failure classes. Automatically retry only reversible runtime/artifact/interruption
  failures while budget remains. Scope, identity, model, stale-head, authority, and contract
  failures must fail closed.
- Finish bounded reconciliation of issue-scoped orphaned subagent-run records after the GSD process
  is no longer live. A `running` child with no live owner becomes `interrupted`; do not touch another
  issue's records.
- Preserve one heartbeat at most every 15 seconds and include only bounded operational child
  status/count/turn metadata—never prompts, output, reasoning, or chain-of-thought.
- Ensure cancellation paths terminate/wait for owned goroutines and do not leave ambiguous
  `running` records.

### 4. Exact completion proof

A successful process exit is insufficient. Before a unit is marked successful require all of:

- canonical GSD query state advanced from the before snapshot, or reached canonical complete;
- expected unit artifact exists and is a bounded regular file for units that require one;
- exact Git head continuity before checkpoint/promotion and a fresh post-checkpoint head;
- no running or unreconciled issue-scoped child;
- observed exact role model and `high` thinking;
- current lease, clean scoped checkpoint, and no write-scope breach.

Represent failures as typed classes such as `runtime_contract_mismatch`, `artifact_missing`,
`false_green`, `interrupted`, `orphaned_subagent`, `stale_head`, `scope_breach`, `model_drift`, and
`retry_exhausted`. Add table-driven regression tests.

Artifact expectations must be narrow and based on the canonical unit, for example milestone
planning requires the native GSD roadmap/projected planning artifact and task execution relies on
canonical query advancement plus its declared validation/checkpoint evidence. Do not invent broad
filesystem scans.

### 5. Functional one-command supervisor

- Add `supervise` to CLI usage and tests.
- Bootstrap or adopt the validated issue once, reconcile state, and loop over `runner.Query()`.
- Dispatch only `snapshot.Next.UnitType` when `Next.Action == "dispatch"`; map
  `discuss-milestone` through the existing targeted `discuss` path.
- Planning/research/discussion/completion/validation/UAT use GPT-5.6 Sol/high. `execute-task` and
  delegated execution use GPT-5.5/high. Continue to verify observed model/thinking after each unit.
- Never use generic `auto`, never dispatch two units concurrently, and never merge the parent PR.
- On retryable failure, continue only when durable budget remains. On exhausted/unsafe/unknown
  failure, emit a bounded typed status and stop at the existing blocked/human-gate boundary.
- On canonical `phase=complete,next.action=stop`, persist/return the final human gate rather than
  performing a merge.
- Make restart idempotent: a completed unit must not be dispatched again merely because the
  supervisor process restarted.
- Add deterministic fake-runner/policy tests that prove canonical command selection, retry stop,
  restart behavior, and final human gate. Prefer a small `internal/supervisor` policy package over
  growing `cmd/shepherd/main.go` further.

## Documentation and evidence

- Update `agent-runtime/shepherd/README.md` and `shepherd.example.json` for local `supervise` usage,
  max attempts, issue-local GSD project identity, role models, heartbeat child fields, and final
  human gate.
- Update phase TDD/verification/run-state artifacts. Keep raw prompts and reasoning out of runtime
  logs and PR summaries.

## Verification

Run and fix, in order:

```bash
cd agent-runtime/shepherd
gofmt -w cmd internal
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/shepherd
make verify
cd ../..
go list ./...
```

Run broader root gates only if shared root Go behavior changes. Do not weaken, delete, or skip tests
to obtain green results. End with a concise file/change summary and exact verification results.
