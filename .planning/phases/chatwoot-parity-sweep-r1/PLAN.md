# Chatwoot documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`chatwoot`, landing order 8). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json`
- **Kind**: openapi, OAS `3.1.0`
- **Retrieved**: 2026-08-07, 488785 bytes
- **Documented operations: 148**
- **By method**: DELETE 18, GET 64, PATCH 22, POST 42, PUT 2
- **Read / write split**: 64 read, 84 write
- **Deprecated (still counted)**: 0

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 146 |
| Re-derived | 148 |
| Delta | 2 |

**Finding: the ledger is stale.** The live artifact disagrees; see the note below.

**How it was counted.** Counted from https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json (HTTP 200, 488785 bytes), fetched 2026-08-07. The recorded ledger URL (github.com/.../swagger/tag_groups) is a directory listing, not a fetchable artifact; the repo's own build bundles the tag_groups/*.yml fragments into this single swagger.json, which I used directly (spec_fragmented=false from my retrieval perspective -- I did not need to assemble fragments myself). Parsed with python3 json module. Counted every paths.<path>.<method> entry where method is one of get/put/post/delete/patch/head/options/trace (92 paths, 148 operations). openapi=3.1.0, info.version=1.1.0. No top-level 'webhooks' OAS object present, so oas_31_webhooks_block=false and blocks_connectorgen_batch=false. Webhook subscription CRUD lives at /api/v1/accounts/{account_id}/webhooks (4 ops, included in total per policy). The 10 webhook event names were pulled from the enum on components.schemas.webhook_create_update_payload.properties.subscriptions.items (not counted in operations_total per policy).

## Hazards

- **Path-collapse hazard**: operations share a path and are selected by header/body. Rows in `api_surface.json` must be keyed so each operation stays distinct, or the whole surface collapses into a handful of rows. This has already bitten this programme once (DynamoDB's `X-Amz-Target`).

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 10** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 4** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm chatwoot <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#148** (old generation); children are expected at **#149–#154** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/chatwoot_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^chatwoot_'`) — gong carried two, and a targeted
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
  non-chatwoot path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
