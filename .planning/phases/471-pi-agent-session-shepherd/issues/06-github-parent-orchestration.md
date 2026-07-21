# Objective

Implement typed GitHub orchestration for parent issues, dependency-linked sub-issues, draft parent
PRs, stacked sub-PRs, CI/review evidence, correction disposition, and provisional integration.

Parent: #471
Parent PR: #472
Dependencies: #474, #476, and #477 (Wave 3)
Branch: `feat/478-shepherd-github-parent-orchestration`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/github-orchestrator.ts`
- `.pi/extensions/shepherd/review-router.ts`
- `.pi/extensions/shepherd/github-evidence.ts`
- matching tests/fixtures
- this issue's GSD/TDD artifacts

Controller/command integration remains reserved for the integration issue.

## Acceptance criteria

- [ ] Parent objectives become bounded child records with dependencies, scopes, branches, PR bases,
      required skills, verification, human gates, and idempotency markers.
- [ ] Parent branch/draft PR setup, sub-issue creation, sub-PR creation, and roster/status updates are
      retry-safe and reconcile existing GitHub state before mutation.
- [ ] Child PRs target the parent branch and use `Refs`; only the parent PR closes the parent issue.
- [ ] CI, requested changes, unresolved threads, exact reviewed commit ranges, and finding
      dispositions are represented as authoritative evidence.
- [ ] This program's automated quality gate is a controller-owned, in-process independent Codex
      review using `gpt-5.6-sol` with `xhigh`. Coverage binds the exact base/head/range, observed
      stable head, changed paths, findings, and dispositions, and reruns after any head change.
      Claude and Copilot are intentionally not claimed as review coverage for this program.
- [ ] Only green, scoped, reviewed sub-PRs integrate. Missing or stale exact-head Codex coverage is
      recorded as blocked rather than inferred from CI, prose, scores, or another provider.
- [ ] Parent PR never becomes ready or mergeable until every required child and exact-head gate passes.

## TDD and verification

Use fake GitHub state machines with RED idempotency/staleness/review cases first. Required skills:
`github-issue-first-delivery`, `gsd-workstreams`, `javascript-testing-patterns`.

```bash
node --test .pi/extensions/shepherd/github-orchestrator.test.ts \
  .pi/extensions/shepherd/review-router.test.ts \
  .pi/extensions/shepherd/github-evidence.test.ts
git diff --check
```

Human gates: final parent ready/merge decisions and all policy exceptions route through the broker.
