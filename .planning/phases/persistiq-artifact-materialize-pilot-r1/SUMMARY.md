# PersistIQ artifact materialization pilot - summary

**Date:** 2026-08-08
**Status:** generator capability and staged multi-source generalization pilots passed static gates; production 392 remains deferred

## Executive result

The captain-policy rerun mapped all 21 documented PersistIQ operations into a
materialized bundle. It kept all three existing source-surface operations that
the fetched artifact does not document and marked each with the exact reason
`present-in-surface-absent-from-artifact`. Three documented operations remain
visible as `not_implemented` commands with machine-checkable named
dependencies; none is falsely marked implemented.

No credential was read, requested, printed, or used. No provider operation was
exercised. Certification is withheld.

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

The complete method/path classification is in
[`operation-mapping.json`](rerun-2026-08-08/operation-mapping.json). The map
reconciled with the fetched artifact's 12 GET, 6 POST, 1 PATCH, 1 PUT, and 1
DELETE operations.

## Artifact evidence

- URL: `https://persistiq.com/api-docs/v1/swagger.json`
- Verified format: OpenAPI 3.0.1
- Paths: 14
- Operations: 21
- Bytes: 47,796
- SHA-256: `0bf3e1ecbfbf6215360b5bb8f9d4fda816df4e1872470a00b529fb3e8b80946f`
- Fetched: 1; parsed: 1

The fresh artifact evidence is in
[`artifact-evidence.txt`](rerun-2026-08-08/artifact-evidence.txt) and
[`persistiq.json`](rerun-2026-08-08/artifacts/persistiq.json).

## Materialization result

| Measure | Count |
|---|---:|
| Artifact operations mapped | 21 |
| Materialized API-surface rows | 24 |
| Implemented commands | 21 |
| Named-dependency commands | 3 |
| Flagged discrepancies | 3 |
| Reachable command help paths | 24/24 |
| Implemented commands reachable | 21/21 |
| Not-implemented commands visibly blocked before network | 3/3 |
| Failed connector candidates | 0 |

Named dependencies:

- `GET /v1/leads/{id}` → `engine.direct_read_executor`
- `POST /v1/leads` → `engine.rest_write_body_envelope`
- `PUT /v1/webhook_plugin` → `review.webhook_url_mutation`

Exact flagged discrepancies:

- `GET /v1/mailboxes`
- `GET /v1/activities`
- `GET /v1/accounts`

## Static and runtime-preflight gates

- `connectorgen validate`: pass, 0 findings.
- `surface-sync --check`: pass, no drift.
- `connectorgen batch gate`: pass, 1 included / 0 dropped; 21 implemented
  commands passed the real runtime preflight.
- `TestEveryImplementedCommandPassesRuntimePreflight`: pass.
- Real built `pm` binary: all 24 generated help paths reachable; the three
  named-dependency commands returned an intentional
  `availability=not_implemented` block and never reached credentials/network.

The generated bundle and gate reports are under
[`rerun-2026-08-08`](rerun-2026-08-08/).

## Multi-source generalization summary

Watchmode (23 read-only OpenAPI operations), DocuSeal (34 OpenAPI 3.1
operations including 11 top-level webhooks), and Float (102 operations after
bounded Swagger-2 external path traversal) all materialized and passed the
combined static gate. Their real binaries reached every generated command path:
45/45, 34/34, and 104/104 respectively, with zero failures. Copper's 77
Postman operations also materialized and passed static gates; its current
legacy native scaffold does not expose an embedded command surface, so the
summary makes no reachability claim for it.

The exact bucket counts, hashes, per-operation provenance, alternatives,
discrepancies, and final wall-clock slices are in
[`generalization-validation-2026-08-08/GENERALIZATION-VALIDATION.md`](generalization-validation-2026-08-08/GENERALIZATION-VALIDATION.md).
All four outputs are staged evidence, not production connector bundles.

## Timed wall-clock results

| Step | Wall-clock | Evidence |
|---|---:|---|
| 1. Identify ledger link | 0.03s | `timing-step1.txt` |
| 2. Map 21 operations | 0.03s | `timing-step2.txt` |
| 3. Fetch, digest, parse | 2.70s | `timing-step3.txt` |
| 4. Materialize, static gates, binary reachability | 50.07s | materialize/gates 13.60s; bare namespace/build/reachability/block checks 36.47s |
| 5. Report collation | 0.09s | `timing-step5-report.txt` |
| **Total** | **52.92s** | sum of timed slices |

Materialization no longer runs the repository-wide preflight per candidate.
The batch design now fetches separately, materializes separately, and gates
once over the staged result; review-sized batch boundaries remain available
for commits and targeted diagnosis.

## Certification statement

Certification is **WITHHELD**. The connector is implemented according to the
static evidence above, **not certified**, and was **never exercised against the
PersistIQ provider**.

## Handoff / go-no-go input

PersistIQ is the only connector materialized as the original end-to-end pilot;
Watchmode, DocuSeal, Float, and Copper are staged generator-validation
evidence. The eligible 392 pool remains untouched in this phase evidence; the
captain's batch-efficiency ruling is recorded for the next authorized phase.
