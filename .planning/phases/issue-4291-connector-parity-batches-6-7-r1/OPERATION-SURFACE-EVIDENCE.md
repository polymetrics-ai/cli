# Operation-surface evidence — issue #4291

Machine record: `OPERATION-SURFACE-EVIDENCE.json`.

## Increment 1 — stale destination-gap reconciliation

The record contains all 3,932 source-locked provider operations across the twenty assigned
connectors. Every row carries source URL, source-lock `rest.info_version`, SHA-256, canonical
mapping, seven separately named surface cells (direct read/write, ETL, reverse ETL, binary
read/write, and executable CLI), generated-projection cells, certification cells, and merge
readiness. No missing source trace or surface-cell key was observed.

The former `generic-typed-destination-executor` gap was stale after foundation
`609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` for `salesloft`, `copper`, `klaviyo`, `intercom`,
and `freshdesk`. Their 499 direct-write rows have no connector-owned typed action, so their
reverse-ETL status is now explicitly `declaration-pending`, not a foundation gap and not a
safety exclusion. Each connector's destination transport is likewise an authored-declaration
pending item with the exact typed action, source binding, acknowledgement, strategy, fixture,
and conformance work named.

| Open gap id | Rows | Status |
| --- | ---: | --- |
| `declarative-operation-route-override` | 5 | Resolved by `6410fe59c`; the five `mailbox_v3` commands now route through declared operation routes. |
| `declarative-typed-destination-action-specific-source-bindings` | 1 | Resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` plus the connector-owned `customers.id → customerId` binding. |

## Increment 2 — stale destination-gap reconciliation complete

`segment`, `activecampaign`, `iterable`, `square`, and `braintree` supplied the remaining 612
stale generic-destination entries. As in increment 1, none has a connector-owned typed
`writes.json` action, so every direct write is explicitly `declaration-pending` for reverse ETL.
The generic-destination identifier now occurs in no in-scope source ledger or operation-evidence
row. This is not a declaration of deployability: each connector still needs its closed typed action,
source binding, acknowledgement, per-mode strategy, fixture, conformance evidence, generated CLI,
and provider-live certification.

The operation-level fixture/conformance cells are intentionally recorded as `not_recorded` where
the current declaration does not tie evidence to that exact provider operation. They are not
treated as N/A or certification success. Provider-live certification remains pending for every
connector because this lane is credential-free. Therefore the portfolio remains non-merge-ready.

## Increment 3 — Gorgias executor-boundary reconciliation

The prior projection made two false claims: it inferred direct-write classification from POST/PUT
verbs for two documented read/export operations, and it treated the declared-partial `files
download` command as executable. The corrected six source rows are: two direct-write scalar-body
foundation gaps, one recursive-filter direct-read foundation gap, one POST text-export foundation
gap, one PUT direct-read foundation gap, and one `provider-contract-unavailable` binary download.
The latter lacks an exact provider redirect host/final media contract and is not a foundation gap.

The resulting provisional portfolio split is **576 runtime-enabled**, **3,350
connector-declaration-pending**, **5 execution-foundation-blocked**, and **1
provider-contract-unavailable**. Every affected row now has a specific surface classification and
refusal; none is counted as an enabled command.

## Increment 4 — Help Scout foundation reconciliation

The five v3 direct-read rows are now enabled and command-bound through their named `mailbox_v3`
operation routes. The focused route test proves requests use `/v3`, never `/v2/v3`. The exact
`customers(id) → update_customer(customerId)` destination binding is no longer a foundation gap;
it remains correctly marked as connector-owned App/CLI/conformance work rather than falsely
certified. The foundation register now has four open Gorgias executor gaps and no open Help Scout
foundation gap.

## Increment 5 — Outreach reachability and typed-destination contract correction

The 3,350 declaration-pending rows split into **3,254 schema-v3 evidence-reader-gated rows** and
**96 independently convertible Outreach ETL rows**. The latter now have exact API-surface command
bindings and source-disposition evidence. A fresh built binary reached exactly `error: missing
--credential` for all 96 `pm outreach <stream> list` paths; no credential, provider request, or
provider-live certification was used.

The ten branch-only `declarative_typed_destination` declarations are retracted from this readiness
claim. Each selected `full_overwrite`, which the closed adapter rejects, and none supplied the
action-owned per-record batch, private receipt locator, provider-state read-back, or matching
conformance evidence required by [the sync transport contract](../../../docs/sync-transport-definition.md).
Their typed direct-write actions remain enabled; their reverse-ETL cells are accurately
`eligible_declaration_pending`, not foundation gaps. The source transports remain declared.
