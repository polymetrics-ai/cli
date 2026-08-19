# Connector surface sweep progress

This is declaration and local runtime-preflight evidence, not provider-live certification.

| Measure | Total |
| --- | ---: |
| Connector rows | 552 |
| Pending | 532 |
| Gated | 8 |
| Skipped | 12 |
| Failed / materialized / validated intermediate states | 0 / 0 / 0 |
| Direct-read declarations | 23 |
| Direct-write declarations | 6 |
| ETL transport declarations | 2 |
| Reverse-ETL transport declarations | 2 |
| Binary declarations | 14 |
| sync_transport.json files | 2 |

## Batch 1 result

A 20-candidate materialization group produced 8 gated survivors and 12 skipped candidates. The survivors have 99 real commandrunner-preflight checks across 381 declared provider operations. All eight lack certification.json, so no certification is claimed.

Skipped connectors are retained with their exact materializer stage/reason in progress.json. They are not silently retried.

## Resume pointer

Batch 2 should plan a fresh 20-connector candidate group from the recorded external provider-artifact ledger. Do not add transport or direct-read claims while G12–G16 remain open.
