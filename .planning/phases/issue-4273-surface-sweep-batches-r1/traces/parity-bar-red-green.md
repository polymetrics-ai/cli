# Captain parity-bar gap closure — issue #4273

## Task Delivery Header

- Issue: Refs #4273 — surface-sweep batch 1 declarative coverage
- Base branch: `chore/4277-connector-surface-sweep`
- Merges into: `chore/4277-connector-surface-sweep` → `main`
- Delivery: PR #4275 is open against the stated parent branch and awaits Firstmate integration.
- Working branch: `fm/cli-surface-sweep-batches-r1`
- Task: Amend the durable sweep evidence after the captain raised parity to more than 90% of documented provider operations per connector.
- Verification: JSON parse/invariant checks; `connectorgen validate`; `surface-sync --check`; connector boundary; targeted runtime-preflight test; and the repository verification gates.

## GSD gap-plan fallback

Resolved `scripts/gsd sources plan-phase`, `execute-phase`, `verify-work`, and
`code-review`, then generated `scripts/gsd prompt plan-phase 4273 --gaps --tdd`.
The official workflow's role execution cannot run in this Pi adapter because the
canonical contract forbids role spawning. The authorised inline/manual fallback
is therefore recorded in the phase plan, TDD ledger, verification, and review.
No production Go code, foundation schema, scraper, or replacement pipeline was
introduced.

## RED

The prior 552-row progress ledger had a state and parity-class fields but no
per-connector documented-operation percentage. It could truthfully say eight
connectors were `gated` while leaving the remaining 282 materialized provider
operations and all twelve dropped source-ledger rows outside an exhaustive,
fixed-vocabulary rejection record.

## GREEN

- Progress schema v2 gives all 552 connector rows `coverage_percentage` plus
  structured coverage state. The 20 batch-1 candidates measure 99 delivered of
  780 documented operations; no candidate meets the >90% target.
- The batch-1 rejection list records 681 undelivered operations: 282 exact
  provider method/path rows and 12 exact source-ledger rows where materialize
  could not inventory individual operations. Each carries evidence,
  recoverability, recovery, and one allowed reason.
- Foundation-gap rejection rows have matching G12/G14/G15/G17/G18/G19 entries;
  G13 is resolved by main commit `31bfe62eba72ea906a4fa152027db6af1f77908b`.
- The parent and batch-1 branch were rebased onto that DELETE action-kind fix
  before this evidence update.
