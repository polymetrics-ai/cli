# PersistIQ artifact materialization pilot - summary

**Date:** 2026-08-08
**Status:** failed closed at materialization; pilot complete, no second connector started

## Executive result

The one-connector pilot fetched and independently parsed PersistIQ's public
OpenAPI document, but the existing batch materializer refused to produce a
bundle. It found executable legacy coverage for `GET /v1/mailboxes` that is not
present in the cited 21-operation artifact. The current source bundle also
contains `GET /v1/activities` and `GET /v1/accounts`, which are absent from the
artifact. Those legacy streams were not silently deleted or rewritten.

This is a failed pilot, not a certification. No PersistIQ credential was read,
requested, printed, or used; no provider API operation was exercised.

## Exact operation map

| Bucket | Count |
|---|---:|
| ETL | 11 |
| direct_read | 1 |
| reverse_etl | 7 |
| direct_write | 2 |
| binary_download | 0 |
| Unclassified | 0 |
| **Total** | **21** |

The map reconciles with the fetched artifact's 12 GET, 6 POST, 1 PATCH, 1
PUT, and 1 DELETE operations. `GET /v1/leads/{id}` is the blocked direct-read
candidate; `POST /v1/leads` and `PUT /v1/webhook_plugin` are direct-write
candidates; the seven record-shaped mutations are reverse ETL.

## Artifact evidence

- URL: `https://persistiq.com/api-docs/v1/swagger.json`
- Verified format: OpenAPI 3.0.1
- Paths: 14
- Operations: 21
- Bytes: 47,796
- SHA-256: `0bf3e1ecbfbf6215360b5bb8f9d4fda816df4e1872470a00b529fb3e8b80946f`
- Fetched: 1; parsed: 1

The artifact and its metadata are in `ARTIFACT-MANIFEST.json` and
`artifacts/persistiq.json`.

## Timed wall-clock results

| Step | Wall-clock | Evidence |
|---|---:|---|
| 1. Identify ledger link | 0.02s | ledger URL lookup |
| 2. Map 21 operations | 0.04s | five-bucket mapping |
| 3. Fetch, digest, parse | 2.75s | one bounded fetch + OpenAPI parse |
| 4. Materialize and static gates | 16.92s | plan 4.87s; materialize 2.43s; validate 1.70s; surface-sync 0.86s; batch gate 0.73s; repository runtime-preflight test 6.33s |
| 5. Report | 0.01s | local evidence collation |
| **Total** | **19.74s** | sum of timed slices |

## Real counts and gates

| Measure | Result |
|---|---:|
| Fetched | 1 |
| Parsed | 1 |
| Materialized | 0 |
| Gated | 0 |
| Reachable generated commands | 0 |
| Failed connector candidates | 1 |

Baseline Red was observed against the real binary: `error: unknown command
"persistiq"`. No generated binary sweep was possible because no destination
bundle was written. The repository-wide runtime-preflight test passed for the
existing embedded definitions, but that is not evidence that a generated
PersistIQ command is reachable.

## Gate disposition

- `connectorgen batch plan`: pass, one PersistIQ candidate / 21 surveyed ops.
- `connectorgen batch materialize`: **fail closed** at coverage on
  `GET /v1/mailboxes`.
- `connectorgen validate`: source-only pass with 0 findings; generated pilot
  candidate did not exist.
- `surface-sync --check`: source-only failure, `runtime endpoint ledger
  drift=true`.
- `connectorgen batch gate`: source-only failure, legacy v0 provenance refusal.
- `TestEveryImplementedCommandPassesRuntimePreflight`: repository test pass;
  not generated PersistIQ evidence.
- Real `pm` reachability: baseline failed; generated commands 0/0 because
  materialization failed.

## Certification statement

Certification is **WITHHELD**. This pilot is implemented only through fetched
artifact evidence and attempted static tooling; it is **not certified** and was
**never exercised against the PersistIQ provider**.

## Handoff / go-no-go input

The captain must decide how to reconcile existing undocumented-but-supported
legacy streams with the strict artifact coverage requirement before any bulk
run. This pilot did not choose, delete, or invent that reconciliation.
