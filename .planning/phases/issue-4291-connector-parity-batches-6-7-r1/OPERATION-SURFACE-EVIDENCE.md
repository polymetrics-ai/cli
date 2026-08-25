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

## Increment 6 — schema-v3 declaration-first ETL cohort

Twenty-six previously declaration-pending source operations are now executable ETL commands: eleven
ActiveCampaign streams, five Freshdesk streams, three Iterable streams, three Segment streams, and four
Square streams. Each row keeps the source operation identity, provider route and citation, canonical
command path, ETL lane, and `streams.json` execution component. The installed binary stopped at the
credential preflight for all 26 paths in a clean local project; provider-live certification remains
pending. Their persisted source-transport cells remain `declaration-pending` because none has invented
connector-owned transport/conformance evidence.

This advances the prior 3,254 schema-v3-reader-gated rows by 26 without changing a lock: **3,228** remain
blocked from generated validation by the shared evidence reader, while the 26 independently declarable
stream commands are source-backed and binary-reachable. `api_surface.json` served only as an exact
checked-in projection consistency check, not as a provider-boundary substitute.

## Increment 7 — schema-v3 declaration-first ETL cohort 2

Forty more source-backed ETL commands are installed for Braintree (10), Close (14), Intercom (5),
Salesloft (5), and Klaviyo (6). Their clean-project binary result is consistently the credential
preflight, and their generated website catalog projections are current. The 3,932-row split is now
**738 enabled**, **3,188 declaration-pending**, **5 foundation-gap**, and **1 provider-contract-
unavailable**.

Copper's five apparent ETL streams are not in this increment: the checked-in execution component is a
`__legacy_hook` path but the exact source operation is a POST search request. A GET CLI declaration would
misrepresent method and body semantics, so the deferred row must name the connector-owned POST-search
stream declaration instead. That is ordinary declaration work, not a reason to omit the provider API or
mislabel an engine gap.

## Increment 8 — schema-v3 declaration-first ETL cohort 3

Braze, Chargebee, and Customer.io add 69 exact source-backed ETL commands (21, 32, and 16 respectively).
All retain operation identity, citation, canonical route, command binding, and `streams.json` component;
all stop at credential preflight in a clean local project. Generated website catalog artifacts were
regenerated with the source declarations.

The current portfolio split is **807 enabled**, **3,119 declaration-pending**, **5 foundation-gap**, and
**1 provider-contract-unavailable**. The pending state is not a safety, privilege, destructive, or live-
certification exclusion. It is either connector declaration work with its exact execution component named,
or the common schema-v3 source-projection reader prerequisite for global generated checks.
