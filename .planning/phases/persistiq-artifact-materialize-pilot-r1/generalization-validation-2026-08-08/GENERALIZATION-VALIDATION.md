# Generator generalization validation

**Date:** 2026-08-08

**Purpose:** Evidence-only validation of PR #3957 before merge. The three
selected candidates were staged outside `internal/connectors/defs`; no
generated connector surface was added to the PR's production bundle corpus.
No credentials were read or used, and no provider operation was exercised.

## Candidate selection

The eligible-pool exclusions were applied first: 28 connectors already live
in `main` and the nine in-flight connectors (`workday-rest`, `jira`,
`help-scout`, `greenhouse`, `chatwoot`, `gmail`, `lever-hiring`,
`zendesk-support`, and `github`) were not touched.

The deliberate shapes were:

| Connector | Shape | Ledger operations | Artifact URL | Result |
|---|---|---:|---|---|
| watchmode | public, read-only, OpenAPI 3 | 23 (23 read / 0 write) | `https://api.watchmode.com/openapi.json` | pass |
| docuseal | public, write-heavy, OpenAPI 3.1 | 23 (7 read / 16 write) | `https://console.docuseal.com/openapi.yml` | **failed** |
| float | public, Swagger 2.0, mixed/write-heavy | 95 (44 read / 51 write) | `https://developer.float.com/swagger-api-v3.yaml` | **failed** |

This validation does **not** establish generator readiness. Two of the three
selected shapes fail before a bundle can be materialized, as detailed below.

## Watchmode — pass, including discrepancy handling

Artifact evidence:

- Verified format: OpenAPI 3.0.3; 23 paths / 23 operations.
- Bytes: 101,353.
- SHA-256: `9e306e252b816d5ec68aa65473eab846e845ffc40e3cdeb4d9da9cadb05a7f48`.
- Fetched: 1; parsed: 1; materialized: 1; gated: 1; dropped: 0.

The complete artifact operation map is in
[`watchmode-operation-mapping.json`](reports/watchmode-operation-mapping.json):

| Mapping bucket | Count |
|---|---:|
| etl | 0 |
| direct_read | 23 |
| reverse_etl | 0 |
| direct_write | 0 |
| binary_download | 0 |
| unclassified | 0 |
| **Mapped artifact operations** | **23** |

Materialization retained 22 source-surface rows absent from the artifact,
with the exact marker `present-in-surface-absent-from-artifact`; the
versioned `/v1/.../` rows are listed in the generated `api_surface.json`.
The generator did not refuse the connector.

| Measure | Count |
|---|---:|
| Materialized API-surface rows | 45 |
| Implemented commands | 13 |
| Named-dependency commands | 32 |
| Flagged discrepancies | 22 |
| Reachable real-binary command paths | 45/45 |
| Implemented commands reachable | 13/13 |
| Not-implemented commands visible/reachable with `--help` | 32/32 |
| Failed command paths | 0 |

All 32 named-dependency commands identify
`engine.direct_read_executor`; none is falsely marked `implemented`.

Static gates:

- `connectorgen validate`: pass, 0 findings.
- `surface-sync --check`: pass, no drift.
- `connectorgen batch gate`: pass, 1 included / 0 dropped; 13 implemented
  commands passed the real runtime preflight.
- `TestEveryImplementedCommandPassesRuntimePreflight`: pass (the existing
  551-bundle corpus test).
- Real `pm` binary: bare namespace succeeded; all 45 command help paths were
  reachable without credentials or provider network access.

Timed slices (wall clock):

| Step | Time |
|---|---:|
| Identify ledger link | 0.04s |
| Fetch and digest | 2.52s |
| Map 23 operations from the materialized artifact inventory | 0.02s |
| Batch plan | 1.78s |
| Materialize and parse | 0.65s |
| Validate | 0.67s |
| Surface-sync derive / check | 0.68s / 0.64s |
| Batch gate and staged runtime preflight | 0.66s |
| Existing-corpus runtime-preflight regression test | 5.28s |
| Build real binary with staged bundle overlay | 9.71s |
| Bare namespace | 2.48s |
| 45-command reachability sweep | 54.70s |
| Report collation | 0.06s |
| **Total observed slices** | **79.89s** |

## DocuSeal — failed before mapping

Artifact evidence:

- Verified format by YAML parse: OpenAPI 3.1.0; 17 paths / 23 path
  operations; 11 top-level webhooks.
- Bytes: 192,929.
- SHA-256: `7ac10d1c39b335bce962b6de277d88aded8ce476518b83835c76ad80157e0e4b`.
- Fetch succeeded: 1.

`batch materialize` dropped the candidate at `artifact_inventory_unknown`:

```text
top-level webhooks (formCompletedWebhook, formDeclinedWebhook,
formStartedWebhook, formViewedWebhook, submissionArchivedWebhook,
submissionCompletedWebhook, submissionCreatedWebhook, submissionExpiredWebhook,
templateArchivedWebhook, templateCreatedWebhook, templateUpdatedWebhook)
cannot be represented as provider request paths
```

This is a generator refusal, not an executor availability decision. Therefore
the emitted counts are: mapped 0 (the 23 path operations were not materialized),
implemented 0, named-dependency 0, flagged-discrepancy 0 (bundle was never
created), reachable 0, failed candidates 1. Static gates and binary
reachability were not applicable because no bundle was emitted.

Timings: identify 0.05s; fetch/digest 2.38s; materializer/parser refusal
0.63s; total observed failure path 3.06s.

## Float — failed before mapping

Artifact evidence:

- Verified root format: Swagger 2.0; 49 path entries. The root has external
  path-item references, so the provider's 95 operations cannot be
  exhaustively enumerated from the fetched root alone.
- Bytes: 8,634.
- SHA-256: `d204eae066136386aea4ea955fb9d0d08ef9ca85eafabc2bb2dcd30b8751211c`.
- Fetch succeeded: 1.

`batch materialize` dropped the candidate at `artifact_inventory_unknown`:

```text
external path-item reference "paths/accounts.yaml#/accounts"
cannot be exhaustively resolved
```

This is also a generator refusal before mapping, not an executor availability
decision. Therefore the emitted counts are: mapped 0 (the ledger's 95
operations were not materialized), implemented 0, named-dependency 0,
flagged-discrepancy 0, reachable 0, failed candidates 1. Static gates and
binary reachability were not applicable because no bundle was emitted.

Timings: identify 0.04s; fetch/digest 0.33s; materializer/parser refusal
0.63s; total observed failure path 1.00s.

## Candidate substitutions and additional failures

- The suggested `web-scrapper` artifact fetched successfully (OpenAPI 3.0.3,
  47,427 bytes, SHA-256
  `0c1d30f72dab7c61544387706f921cb19a4b02fa692a69768d433af9a11f8716`), but
  the ledger marks it `access_model=partner_gated`; the existing batch planner
  correctly refuses it rather than misrepresenting it as public. It was not
  counted as one of the three validation candidates.
- Ding Connect was the first Swagger-2 substitution, but its ledger URL
  returned HTTP 403 on two polite requests (plain and browser User-Agent).
  Float was selected as the eligible public Swagger-2 replacement.

## Conclusion and certification

The watchmode pass proves that a read-only OpenAPI 3 surface and 22
surface/artifact discrepancies are retained and exposed without refusal.
The DocuSeal and Float failures show that the generator does **not** yet
generalize to top-level webhook declarations or externally referenced Swagger
path items. The generator is therefore **not ready** and this evidence must
not be read as certification.

All results are static/documentation-only. No connector is certified; none was
exercised against its provider. The eligible 392 remain untouched and no
generated production connector surface was added to PR #3957.
