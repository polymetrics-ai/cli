# TDD Ledger — #3860 polling-watermark truth surfaces

## Slice 1 — preflight-derived surface eligibility

- Red: pending. Add a focused rendered-surface test before production changes. It must observe a blocked/unavailable row when the declaration or registry does not pass preflight, and it must fail if a renderer labels that state `implemented` or calls it CDC.
- Green: pending. The real preflight projection supplies deterministic mode rows. Tests assert the complete visible contract for unsafe cursor, soft-delete-only, identity-mismatch/rebootstrap, and a valid registered declaration.
- Refactor: pending. Preserve one runtime decision source and avoid copied capability strings.

## Slice 2 — native protocol surface and docs parity

- Red: pending. Add an assertion that a database-native PostgreSQL definition has an empty REST endpoint list and that generated/help documentation does not invent an endpoint.
- Green: pending. Update source docs and generated output from a newly built binary; assert hard deletes are not observable, rebootstrap is required, and polling is not change capture.
- Refactor: pending. Review generated diffs line by line and retain only issue-scoped changes.

## Status

Plan created before production edits. Red/green command output and test names will be recorded as each slice is executed.
