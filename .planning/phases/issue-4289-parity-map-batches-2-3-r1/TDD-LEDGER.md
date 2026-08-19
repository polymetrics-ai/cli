# Issue #4289 — TDD Ledger

## Red

On `main`, all nineteen selected bundles lack their connector-local source lock and corrected six-class declaration-disposition ledger. The observable integrity assertion is therefore impossible: a source denominator cannot equal a nonexistent inventory or disposition map.

## Green

For each selected bundle, a source lock pins a credential-free public provider description and the disposition ledger contains exactly one row per documented operation. The local integrity check verifies:

- source inventory count equals `ledger_dispositions` count;
- every row has all required corrected batch-1 fields;
- every source operation has exactly one method/path API-surface binding;
- parity-class totals equal the pinned source denominator;
- `foundation-gap` records include a concrete engine file/line and minimal change; enabled typed `direct_write` rows carry reverse-ETL eligibility metadata using the actual `generic-typed-destination-executor` refusal at `internal/app/issue_label_warehouse_transport.go:85-95` rather than the retired estate-wide source gap;
- unauthored connector work is `declaration-pending`, and elevated scopes do not disable rows.

`connectorgen validate` and `surface-sync --check` remain the production structural green gates. No behavior changes are made, so no Go unit-test red phase is appropriate; the map-integrity assertion is the testable artifact behavior added by this issue.

## Refactor

Keep the generated ledgers connector-local and use the exact corrected batch-1 schema. Do not alter engine code, infer schemas, fabricate transport descriptors, request a credential, or produce a terminal-command/certification claim.
