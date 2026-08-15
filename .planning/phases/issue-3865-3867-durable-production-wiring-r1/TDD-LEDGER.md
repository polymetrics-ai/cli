# TDD LEDGER — #3865/#3867 durable production wiring

## Red

- Pending: durable file-store contracts do not compile because production
  stores and atomic mutation/claim APIs do not exist.
- Pending: production composition tests fail because `app.Open` constructs no
  auth or parking coordinator, resolved runtimes carry no admission owner, and
  ETL dispatch neither parks nor resumes.
- Exact commands and failure output will be appended before production edits.

## Green

- Pending focused, race, child-process, and live PostgreSQL evidence.

## Refactor

- Pending formatting, review findings, derived-artifact regeneration, and
  broader gates.

## GSD lifecycle

- `scripts/gsd prompt discuss-phase 3865-3867-durable-coordination-r1 --auto`
- `scripts/gsd prompt plan-phase 3865-3867-durable-coordination-r1 --tdd --skip-research`
- Execute/verify/code-review commands remain pending.
- Inline/manual fallback: issue-scoped work is not a numbered roadmap phase and
  the canonical contract forbids role spawning.
