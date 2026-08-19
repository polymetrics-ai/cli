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
- `foundation-gap` records include a concrete engine file/line and minimal change; enabled typed `direct_write` rows carry reverse-ETL eligibility metadata using the actual persisted App/CLI dispatch refusal at `internal/app/transport_dispatch.go:53-67`, while the action ledger separately records the one-action-per-mode selection limit at `internal/connectors/sync_transport.go:388-415`;
- unauthored connector work is `declaration-pending`, and elevated scopes do not disable rows.

`connectorgen validate` and `surface-sync --check` remain the production structural green gates. No behavior changes are made, so no Go unit-test red phase is appropriate; the map-integrity assertion is the testable artifact behavior added by this issue.

## Refactor

Keep the generated ledgers connector-local and use the exact corrected batch-1 schema. Do not alter engine code, infer schemas, fabricate transport descriptors, request a credential, or produce a terminal-command/certification claim. Never promote a rendered index or a self-referential count to complete source coverage.

## Held-PR Repair — Red / Green

**Red:** the PayPal Transaction Search OpenAPI document exposed only two reporting routes, so a lock pinned solely to that file could not represent the provider's complete documented REST surface. The ten batch-3 locks also did not expose a root `counts.total`, preventing the fleet-wide source-accounting check the captain required.

**Green:** `generate-parity-maps.mjs paypal-transaction` now consumes all thirteen official `openapi/*.json` files from PayPal's published specification archive and maps 115 exact operation declarations. `verify-parity-maps.mjs` fails unless both root and REST total counts equal the immutable source-operation inventory. The fresh green run reports `verified 19 connectors / 5127 documented operations`.

## Reconciliation Relaunch — Red / Green / Refactor

**Red:** the source-ledger implementation proves only source accounting and the six original parity classifications. It does not prove that every faithfully representable documented operation has an exact typed direct action, a connector-owned typed-destination binding, a source transport declaration, or an installed-binary command artifact. The new seven-surface ledger assertion must therefore fail against the preserved source-map-only state.

`node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs` → **Red**: exits 1 and writes `SEVEN-SURFACE-LEDGER.json`; for example, Grafana reports 139 missing direct reads, 171 missing direct writes, and four missing ETL bindings. The test derives this from each source-locked ledger, bundle operations, CLI commands, and transport declaration, rather than counting source rows alone.

**Green:** after merging `fm/cli-reverse-etl-destination-r1`, each connector-local definition carries only source-backed executable contracts: exact typed direct-read/write actions, distinct binary contracts where provider evidence represents a transfer rather than REST, ETL source declarations, eligible typed reverse-ETL destinations, and generated operation/CLI surfaces. The seven-surface ledger reports every connector and asserts no documented operation was silently omitted, disabled for privilege/destructiveness, or moved into a generic writer.

**Refactor:** retain the immutable source locks and existing disposition rows as inputs; normalize action and transport metadata mechanically through connector-local generation. Keep REST direct-write, binary transfer, and reverse-ETL destination contracts distinct. Unsupported remains a precise engine incapability with source and refusal evidence, never a proxy for missing authoring.

**Foundation hold:** the initial #4304 commit composes generic typed destinations but does not yet select them in the App/CLI persisted-dispatch path. Declarations can be structurally green, but an installed App/CLI reverse-ETL run is red until the updated foundation branch is merged and that path is exercised. No connector-local substitute is permitted.

## Typed Write Eligibility — Red / Green

**Red:** a source-row-only eligibility assertion leaves five existing typed actions without an exact action disposition because their legacy paths are base-relative or their published source inventory has no exact row. It also makes a one-action destination look like a connector-complete reverse-ETL declaration.

**Green:** `verify-parity-maps.mjs` requires every `writes.json` action to have exactly one action-level eligibility disposition. The 621 target actions are semantically eligible and individually representable by `declarative_api/declarative_typed_destination`; none is excluded for safety, privilege, or destructive behavior. The disposition records whether it has an exact source-row binding and names `declarative-typed-destination-action-multiplicity` when the current one-action-per-mode descriptor cannot select all eligible actions. This is a foundation hold, never a connector-local selector or a completion claim.

## REST Body Input — Red

Provider-backed nested REST bodies cannot be faithfully exposed from the current operation CLI surface: `commandrunner` accepts structured JSON only for fixed GraphQL operations and refuses an exact direct-write `body` mapping. The plan records the bounded declaration-owned REST-body foundation request under `[key=rest-structured-body-cli]`; no raw-body or unvalidated connector workaround is allowed.
