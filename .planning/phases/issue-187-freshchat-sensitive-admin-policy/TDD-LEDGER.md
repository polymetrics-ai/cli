# TDD Ledger — Issue #187

## GSD note

`scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`), so this slice uses the manual GSD/TDD fallback.

## Red target

- Freshchat write metadata must mark admin, sensitive, and destructive operations with typed confirmation challenges.
- Commandrunner must carry those challenges into connector command write plans.

Expected initial failure: only Freshchat delete actions currently declare `confirm: destructive`; admin/sensitive writes have no confirmation challenge and the schema only permits `destructive`.

## Green target

Extend confirmation vocabulary, add Freshchat confirmations, and update docs/generated data.

## Verification ledger

Pending.
