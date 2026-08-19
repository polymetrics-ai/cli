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
| `generic-typed-destination-executor` | 612 | Stale only on the five remaining unreconciled ledgers; next independent increment removes it |

The operation-level fixture/conformance cells are intentionally recorded as `not_recorded` where
the current declaration does not tie evidence to that exact provider operation. They are not
treated as N/A or certification success. Provider-live certification remains pending for every
connector because this lane is credential-free. Therefore the portfolio remains non-merge-ready.
