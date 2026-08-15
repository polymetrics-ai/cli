# Context — #3859 native database apply strategies

## Task delivery header

- Issue: `Closes #3859 — feat(sync): implement native database apply strategies`
- Base branch: `integration/4015-mvp-flat-r1`
- Delivery: direct PR from `fm/cli-3859-native-apply-strategies-r1` to
  `integration/4015-mvp-flat-r1`; do not merge it.
- Required local proof: scoped unit tests, the explicit PostgreSQL dbtest run,
  and the repository's individual non-suite verification gates.

## Fixed decisions

- The task begins from the rebased integration head. `MappingContractV1`, the
  transactional `DatabaseWriteExecutor`, the PostgreSQL managed-target driver,
  and sealed workset delivery are present. They are consumed, not recreated.
- #3858 owns source-page production and checkpoint persistence. This slice owns
  only the closed registered target apply boundary and must not duplicate a
  source reader, query builder, or source checkpoint path.
- Target deletes are explicit tombstones. A record missing from a page is never
  a deletion signal. History tombstones close the current validity interval;
  they do not physically delete historical state.
- A public PostgreSQL generic-write capability or command is out of scope.
  Existing reverse delivery continues to require its plan, preview, approval,
  and execute boundary.

## Inline GSD fallback

`scripts/gsd doctor`, every required `sources` lookup, and
`go run ./cmd/agentcontractgen check` passed. The generated `discuss-phase`
and `plan-phase --tdd` prompts are executed inline because the canonical
single-worker contract forbids GSD-role spawning. This context and the
discussion, plan, TDD ledger, verification checklist, and review record are
the durable fallback evidence.

## Scope

The implementation may update the shared closed native-apply contract,
database write boundary/tests, PostgreSQL private managed-target driver/tests,
and the PostgreSQL definition only where it truthfully declares an executable
strategy. It must not change #4125, #4136, or #4090; introduce dependencies;
accept raw SQL/HTTP/shell input; add a public write command; or change source
reader behavior owned by #3858.
