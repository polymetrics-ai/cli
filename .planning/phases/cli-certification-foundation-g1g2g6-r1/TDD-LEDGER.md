# TDD ledger: connector certification foundation G1/G2/G6

## Planned evidence

| Slice | Red | Green | Refactor / verification |
| --- | --- | --- | --- |
| G1 projection | Existing sweep schema cannot name exact kind/class/action and rejects neither a mismatch nor a delete action. | Classifier derives exact normalized types from operation, intent, capability, and transport descriptors. | Table tests cover happy, bad, edge, database, and delete routes. |
| G2 generated sweep | Existing bytes omit projection fields and can represent no projection as N/A. | Each row has the generated fields and generated artifact checks are current. | Snapshot drift, validation, and GitHub sweep check pass. |
| G6 evidence import | Existing imports can publish a prefix, final files can be read partial, and the live script deletes a valid evidence record after a drift check. | Batch validation publishes all or none via no-replace atomic paths; readers see valid JSON; script generates scoped shard before checking. | Concurrent reader and forced-failure regressions pass. |

## Actual evidence

### 2026-08-19 — planning checkpoint

- Red: pending production tests; the authoritative scout report verified missing sweep projections and direct final-path proof writes.
- Green: pending implementation.
- Manual GSD fallback: lifecycle prompts resolved and executed inline because isolated compatible GSD workers are unavailable and the task contract forbids role spawning.
