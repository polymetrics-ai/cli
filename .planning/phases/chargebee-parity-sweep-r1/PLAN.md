# Chargebee documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`chargebee`, landing order 23). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://raw.githubusercontent.com/chargebee/openapi/main/spec/chargebee_api_v2_pc_v2_spec.json`
- **Kind**: openapi, OAS `3.1.0`
- **Retrieved**: 2026-08-07, 14552466 bytes
- **Documented operations: 438**
- **By method**: DELETE 0, GET 135, HEAD 0, OPTIONS 0, PATCH 0, POST 303, PUT 0, TRACE 0
- **Read / write split**: 171 read, 267 write
- **Deprecated (still counted)**: 0

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 438 |
| Re-derived | 438 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Counted every unique METHOD+path under `paths` in the OpenAPI 3.1.0 doc (303 POST, 135 GET, 0 PUT/PATCH/DELETE = 438, exactly matching the ledger); Chargebee mutates exclusively via POST. Webhook EVENTS come from the separate top-level `webhooks` object (227 keys) and are excluded from the total per policy. POST read/write classification used path/operationId patterns (search/query/list/batch-retrieve/bulk-retrieve), the substring 'retriev'/'estimat' in operationId (all 80 hits manually reviewed via description text -- pure fetches or non-persisting previews, zero counter-examples), and a chargebee-specific rule that all 15 `/exports/*` POSTs are bulk data exports of existing records (confirmed via description: 'triggers export of ... data ... CSV files'), distinct from the unrelated `export_payment_source` gateway-migration action which stayed a write.

## Hazards

- **`connectorgen batch materialize` is BLOCKED** (finding F2): this artifact is OAS 3.1 with a non-empty top-level `webhooks` object, which `batchArtifactWebhooksUnknown` rejects fail-closed. Hand-author around the gate; do NOT relax it. Webhooks are deferred to `cli-webhook-surface-sweep-r1`.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 227** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 5** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm chargebee <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#3175** (new generation); children are expected at **#3176–#3182** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/chargebee_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^chargebee_'`) — gong carried two, and a targeted
   `-run` missed the second.
2. **GREEN** — author the bundle to satisfy it.
3. **REFACTOR** — docs, catalogs, operation endpoint ledger resync.
4. Gates, then no-mistakes.

`check_red_observed.py` refuses to let this connector proceed to implementation until the red
failure is real observed output.

## Safety notes

- Do not loosen `connectorgen validate`, the connector boundary gate, `certify`, or
  `TestEveryImplementedCommandPassesRuntimePreflight` to make this pass.
- Nothing is marked `implemented` unless its command genuinely runs; every block names a dependency.
- Run the WHOLE `cmd/connectorgen` package plus `internal/cli`, never just a targeted `-run`.
- Regenerating docs rewrites ~1,034 files of pre-existing `main` drift (finding F4) — revert every
  non-chargebee path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
