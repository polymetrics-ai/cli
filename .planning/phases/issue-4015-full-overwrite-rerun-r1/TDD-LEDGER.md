# TDD Ledger

## Source-semantic checkpoint eligibility

- Red: Pending. Add the six-mode orchestrator boundary test and the second-run real PostgreSQL full-overwrite replacement test before production changes. Record exact assertion failures and counts.
- Green: Pending. Full-refresh modes must omit the prior checkpoint; incremental modes must retain it. The binary full-overwrite rerun must independently prove exact replacement.
- Refactor: Pending. Replace duplicate Arrow-only logic with the one shared rule, then run formatting, focused tests, broad gates, and review.

## Required observable evidence

- First full-overwrite run: exact `records_read`, `records_loaded`, target count, and named sample.
- Changed source: exact source count; deleted ID 1, updated ID 2 named sample, inserted ID 4.
- Second full-overwrite run: exact `records_read`, `records_loaded`, target count, IDs, named sample, and ID 1 absence.
- Incremental replay: exact `records_read=0`, `records_loaded=0`, unchanged target count, and unchanged named sample.

