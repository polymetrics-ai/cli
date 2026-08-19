# Issue #4289 — TDD Ledger

## Red

On `main`, all nineteen selected bundles lack their connector-local source lock and corrected six-class declaration-disposition ledger. The observable integrity assertion is therefore impossible: a source denominator cannot equal a nonexistent inventory or disposition map.

## Green

For each selected bundle, a source lock pins a credential-free public provider description and the disposition ledger contains exactly one row per documented operation. The local integrity check verifies:

- source inventory count equals `ledger_dispositions` count;
- every row has all required corrected batch-1 fields;
- every source operation has exactly one method/path API-surface binding;
- parity-class totals equal the pinned source denominator;
- every source lock has `counts.total`, per-kind/method counts, and non-self-referential `operations_found` with a coverage-confidence basis; a partial inventory is visible as a hold, not reported as 100% declared;
- `foundation-gap` records include a concrete engine file/line and minimal change; enabled typed `direct_write` rows carry reverse-ETL eligibility metadata using the actual `generic-typed-destination-executor` refusal at `internal/app/issue_label_warehouse_transport.go:85-95` rather than the retired estate-wide source gap;
- unauthored connector work is `declaration-pending`, and elevated scopes do not disable rows.
- every source-traced operation with an explicit shared binary/download/upload runtime blocker is
  emitted into one machine-readable gap row with `enabled: false` and
  `merge_ready_eligible: false`; the grouped fan-out preserves every affected operation rather
  than treating the generic owner as a reason to omit it.

`connectorgen validate` and `surface-sync --check` remain the production structural green gates. No behavior changes are made, so no Go unit-test red phase is appropriate; the map-integrity assertion is the testable artifact behavior added by this issue.

## Refactor

Keep the generated ledgers connector-local and use the exact corrected batch-1 schema. Do not alter engine code, infer schemas, fabricate transport descriptors, request a credential, or produce a terminal-command/certification claim. Never promote a rendered index or a self-referential count to complete source coverage.

## Held-PR Repair — Red / Green

**Red:** the PayPal Transaction Search OpenAPI document exposed only two reporting routes, so a lock pinned solely to that file could not represent the provider's complete documented REST surface. The ten batch-3 locks also did not expose a root `counts.total`, preventing the fleet-wide source-accounting check the captain required.

**Green:** `generate-parity-maps.mjs paypal-transaction` now consumes all thirteen official `openapi/*.json` files from PayPal's published specification archive and maps 115 exact operation declarations. `verify-parity-maps.mjs` fails unless both root and REST total counts equal the immutable source-operation inventory. The fresh green run reports `verified 19 connectors / 5127 documented operations`.

## Captain gap ledger — Red / Green

**Red:** shared binary and multipart foundation blockers appeared only as endpoint text or an
ad-hoc nested eligibility note, so they had no stable cross-connector fan-out, exact source
revision/hash, closure commands, or portfolio-level merge-readiness effect.

**Green:** `generate-missing-foundation-gaps.mjs --check` reconstructs the checked-in JSON ledger
from the provider locks, dispositions, and Gong multipart action bindings. It fails on every drift;
the main map verifier runs this check before it reports success.
