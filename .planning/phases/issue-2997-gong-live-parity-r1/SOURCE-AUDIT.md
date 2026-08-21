# Gong source audit — 2026-08-21 UTC

## Locked official source

- Source: `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=`.
- Retrieved credential-free at `2026-08-21T22:55:13Z`.
- OpenAPI `3.0.1`; info version `V2`; 59 paths; 69 operations.
- Exact artifact: 453,797 bytes; SHA-256
  `294bf80b28e773d66a30bd0a8e76344140b17cf0225803e759d1e112b6b1fa13`.
- Canonical sorted `(method, path, operation_id, deprecated)` inventory fingerprint:
  `591484a79221a3993643898feef20050c57ca89a0acee028134c639e1fb99014`.
  The committed source lock produces the identical fingerprint.
- Method distribution: DELETE 3, GET 29, PATCH 1, POST 28, PUT 8.

## Exact surface inventory

| Provider surface | Count | Declaration-owned mapping |
| --- | ---: | --- |
| ETL | 12 | Exact stream bindings and implemented `etl` commands. |
| Direct read | 30 | Implemented bounded CLI commands; every row has the exact API-surface endpoint. |
| Direct write / reverse ETL command | 27 | One implemented, named reverse-ETL action per write; confirmation and plan → preview → approval → execute remain required. |
| Binary download | 0 | The official document declares no binary response operation. |
| Binary / multipart upload | 3 | `PUT /v2/calls/{id}/media`, `POST /v2/crm/entities`, and `POST /v2/targets/{targetId}/assignments`; named multipart actions with bounded files and approval digest binding. |
| Application command surface | 69 | The source disposition, API surface, stream/write/operation declarations, and CLI each carry an exact binding; all source rows are enabled. |

The primary source-map class remains `direct_write` for the three multipart operations because
they are also fixed write actions; their binary-upload capability is recorded separately above.
No destructive provider operation is omitted: the three DELETE rows are implemented commands
with typed destructive confirmation.

## Output and foundation audit

- Gong direct reads declare no `sensitive_policy.redact_fields`. The `json_redacted` policy
  deep-clones ordinary provider JSON; the only output masking is for concrete configured
  credential values and retains an explicit marker. Connector docs and the focused surface test
  forbid wording or declarations that imply ordinary response-field redaction.
- The merged parent `c3f83cbf6eabbae00219566fb02719ca2d6c480d` contains the published
  structured-body, source-import, and closed multipart/runtime heads. Focused Gong multipart
  conformance passed after that merge. The Batch 2/3 missing-foundation ledger was regenerated:
  it contains zero Gong gap rows and retains 48 unrelated portfolio rows.
- The strict source-importer lock schema is now used by the Gong source lock. Its scoped command
  is deliberately **not** marked green: `connectorgen source-import gong --check` rejects the
  official fixed query-bearing source URL with `artifact URL must not include a query`. Gong has
  no query-free official equivalent (the query-free route returns 404). This is a
  provider-neutral source-import URL-policy dependency, not a connector workaround or a reason
  to omit any Gong operation.
- A fresh initialized project with no credential ran the built `pm` binary against all 69
  implemented Gong command paths: 30 direct reads, 27 reverse-ETL writes, and 12 ETL streams.
  Every command reached `missing --credential`; none returned `unknown command`, a partial-command
  block, or an API-surface preflight block. No request reached Gong.

## Remaining live gate

No approved non-echoing disposable Gong credential reference is available. Live App-path
certification, provider readback, and cleanup remain unrun; no browser session or customer
credential was used as a substitute.
