## Summary

Build the authoritative autonomous Pi `AgentSession` Shepherd replacement for the abandoned
standalone Go/tmux program.

The parent PR will integrate issue-scoped sub-PRs that provide dependency-aware planning and
scheduling, isolated in-process implementation workers, typed Git/GitHub operations, durable human
decision gates, verification/review/correction loops, recovery, and an end-to-end CLI Architecture
v2 canary.

The local parent contains the earlier control-plane and aggregate MVP inputs. The production #479
child remains outside the parent until its publication, CI, review, and integration gates pass.

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

- #479 production matrix: functional implementation available locally; remote CI/review/PR gates pending.
- Autonomous child roster: #473-#481, dependency-linked from #471; lifecycle reconciliation pending.
- #480 and #481: dependency-blocked.
- Final parent verification: not started.
- Human merge gate: pending after all implementation, CI, review, and canary gates.

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
