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
| `declarative-operation-route-override` | 5 | Open: Help Scout v3 route selection; owned by `cli-operation-route-override-foundation-r1` |
| `declarative-typed-destination-action-specific-source-bindings` | 1 | Open: exact Help Scout `update_customer` source binding |

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
