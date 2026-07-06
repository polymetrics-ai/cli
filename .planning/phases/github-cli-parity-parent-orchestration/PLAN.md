# Plan: GitHub CLI Parity Parent Orchestration

Parent issue: #44
Parent PR: #49
Parent branch: `feat/44-github-cli-parity`

## Model Policy

- GSD loop model: Codex `gpt-5.5`
- Reasoning effort: `xhigh`
- Service tier: `priority`
- Source of truth: `.planning/config.json`

## Manual GSD Fallback

The `gsd-programming-loop` skill references `scripts/programming-loop.mjs` and
`scripts/tdd-gate.mjs`, but this checkout does not contain those scripts. Use the manual GSD loop:
plan, record TDD evidence, implement green slices, verify locally, update traces, then stop at the
human gate.

## Checklist

- [x] Rebase `feat/44-github-cli-parity` on latest `origin/main`.
- [x] Preserve the #51 parent orchestrator workflow already merged into `main`.
- [x] Keep GitHub CLI surface metadata and website documentation from the parent branch.
- [x] Update GSD model policy to Codex `gpt-5.5` with `xhigh` reasoning.
- [ ] Push rebased parent branch.
- [ ] Resolve current PR #49 CodeRabbit findings with dispositions.
- [ ] Continue sub-issues in dependency order:
  - [ ] #35 help renderer
  - [ ] #37 operation ledger
  - [ ] #36 stream-backed runner
  - [ ] #38 constrained direct reads
  - [ ] #39 declarative GraphQL engine
  - [ ] #40 Projects and Discussions
  - [ ] #41 sensitive/admin policy
  - [ ] #42 cross-connector rollout learnings

## Parallelism

Run subissues in parallel only when write scopes are disjoint and dependencies are satisfied. The
orchestrator owns shared parent artifacts, parent PR body, parent branch pushes, review coverage
records, and final human-readiness.

## Human Gates

- Parent PR merge into `main`.
- New dependencies.
- Auth scope changes or `gh auth refresh`.
- Production deployment.
- Secret access or storage.
- Destructive GitHub actions.
- Generic shell, unrestricted HTTP write, unrestricted SQL write, or unrestricted raw API tooling.
- Reverse ETL execution outside plan, preview, approval, execute.
