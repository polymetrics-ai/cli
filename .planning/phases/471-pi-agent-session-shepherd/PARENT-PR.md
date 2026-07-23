## Summary

Build the authoritative autonomous Pi `AgentSession` Shepherd replacement for the abandoned
standalone Go/tmux program.

The parent PR will integrate issue-scoped sub-PRs that provide dependency-aware planning and
scheduling, isolated in-process implementation workers, typed Git/GitHub operations, durable human
decision gates, verification/review/correction loops, recovery, and an end-to-end CLI Architecture
v2 canary.

The parent contains integrated PRs #482-#489 and #491. The exact reconciled parent after #491 is
`c3f4f683e60ac52bcedae04b2e9448e4523b5234`; local, remote, and GitHub heads matched. Integration
invalidated stale parent review evidence. #480 is worker-ready and #481 is dependency-blocked.

Refs #471
Supersedes #372
Supersedes #389
Supersedes #470

## Parent delivery contract

- Parent branch: `feat/471-pi-agent-session-shepherd`
- Child PRs target this branch and use `Refs`, not closing keywords.
- Independent child issues may run in parallel only with disjoint scopes and isolated worktrees.
- Every behavior change follows GSD red-green-refactor and records exact verification/review proof.
- The PR stays draft through child integration and final verification.
- Merge to `main` requires an explicit fresh human approval for the exact verified head.

## Current state

- #473-#479 and #490: provisionally integrated into the non-default parent. Their issues stay open
  until this PR lands on `main`; prematurely closed #473/#474/#476/#477 were reopened.
- #479: PR #489 merged as parent `daaa2263`; hosted checks passed and the complete 17-row matrix is
  preserved.
- #490: PR #491 merged as parent `c3f4f683`; Pi 0.80.10, strict family/provenance/RPC checks, and the
  bounded workflow-engine developer-tool boundary are integrated.
- #480: worker-ready through the single durable `/pm-shepherd` plan.
- #481: dependency-blocked by #480; its read-only live #397/#438 canary must not merge or mutate
  draft PR #438.
- Final parent verification/review: pending after #481.
- Review override: one Codex 5.6-sol xhigh parent round replaces Claude/Copilot for this program but
  is not human approval.
- Human merge gate: pending after all exact-head gates.

## Verification

```bash
node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts
pi --list-extensions
git diff --check
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```
