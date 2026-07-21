# Objective

Run and document an end-to-end autonomous Shepherd canary against CLI Architecture v2 (#397 and
draft PR #438), proving dependency-aware execution through final human readiness without bypassing
the consumer program's existing gates.

Parent: #471
Parent PR: #472
Dependency: #480 (Wave 6)
Branch: `test/481-shepherd-cli-architecture-canary`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- Shepherd integration tests/fixtures and canary harness
- `.planning/phases/471-pi-agent-session-shepherd/**` final traces/evidence
- parent PR/issue status records owned by the orchestrator
- `.pi/README.md` and Shepherd-specific `.agents/agentic-delivery/**` documentation only for the
  parent-owned post-canary deprecation activation after every canary gate passes

CLI Architecture v2 code, issue roster, parent branch, and PR #438 may be mutated only through its
own parent-orchestrator contract and explicit existing human gates. The canary must not merge #438.

## Acceptance criteria

- [ ] A clean restartable run reconciles #397/#438, rebuilds its dependency queue, and reports exact
      implemented/remaining/blocking state before any mutation.
- [ ] At least two genuinely independent eligible lanes execute concurrently in isolated worktrees,
      while dependency/collision lanes serialize with recorded reasons.
- [ ] Each exercised child shows GSD planning, RED/GREEN/refactor evidence, local/CI verification,
      scoped sub-PR handling, and review/correction/integration evidence.
- [ ] A synthetic or real designated human gate creates one correct GitHub request, waits across
      restart, consumes one allowlisted response, and resumes.
- [ ] No credential appears in prompt, state, logs, comments, screenshots, or test output.
- [ ] #438 remains draft/unmerged unless its own exact human gate is separately satisfied.
- [ ] A failed or interrupted canary leaves legacy-shell rollback documentation unchanged; only a
      fully passing canary permits the parent-owned deprecation activation commit.
- [ ] After a fully passing canary, the parent orchestrator applies the pre-reviewed deprecation
      documentation/status delta, reruns affected gates, and records its exact commit as evidence.
- [ ] Parent #472 full gates and independent exact-head review pass; #472 reaches
      `ready_for_human`, not self-approved.

## TDD and verification

First build deterministic sandbox fixtures; run the live canary only after unit/integration gates.
Required skills: `gsd-verify-work`, `gsd-code-review`, `e2e-testing-patterns`, parent orchestration.

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

Human gates: exact parent merge approval remains required. No live connector credential test is
needed to prove Shepherd orchestration; if separately requested, use host environment/keychain only.
