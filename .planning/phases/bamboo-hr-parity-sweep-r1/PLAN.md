# Bamboo Hr documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`bamboo-hr`, landing order 18). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://dash.readme.com/api/v1/api-registry/fdj6s2umsezc7dn (bamboohr-api-1.yaml, OpenAPI 3.1.0) — the live spec backing documentation.bamboohr.com/reference/*`
- **Kind**: openapi, OAS `3.1.0`
- **Retrieved**: 2026-08-07, 992202 bytes
- **Documented operations: 311**
- **By method**: DELETE 31, GET 148, PATCH 10, POST 77, PUT 45
- **Read / write split**: 154 read, 157 write
- **Deprecated (still counted)**: 17

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 311 |
| Re-derived | 311 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Ledger recorded https://documentation.bamboohr.com/reference/getting-started as html_reference. That ReadMe.com project's embedded metadata shows the current STABLE version ('1', version_clean 1.0.0, is_stable:true) has TWO registered OpenAPI files: 'bamboohr-api.json' (uuid 99skfm3ywo65g) and 'bamboohr-api-1.yaml' (uuid fdj6s2umsezc7dn). Fetched both via https://dash.readme.com/api/v1/api-registry/<uuid> (both HTTP 200). They are NOT the same spec: bamboohr-api.json is OpenAPI 3.0.1 with 137 paths / 173 operationIds, path prefix '/{companyDomain}/v1/...', description referencing the legacy 'bamboohr.com/api/documentation' domain. bamboohr-api-1.yaml is OpenAPI 3.1.0 with 230 paths / 311 unique (METHOD,path) operations (zero internal duplicates), path prefix '/api/v1/...', description referencing 'documentation.bamboohr.com' (the exact recorded domain). To determine which is actually live, fetched the site's sitemap.xml (330 total /reference/* URLs) and cross-checked slugs against both specs' operationIds: 306/311 of bamboohr-api-1.yaml's operationIds match live sitemap slugs directly (the other 5 are duplicate-titled operations that ReadMe disambiguates with numeric suffixes, e.g. get-goals-aggregate-v11/v12), while ZERO of bamboohr-api.json's 124 exclusive operationIds appear anywhere in the live sitemap. This proves bamboohr-api.json is an orphaned/unused legacy registry entry and bamboohr-api-1.yaml is the sole live spec. Counted unique (METHOD,path) pairs from bamboohr-api-1.yaml alone: 311 — an exact match to ledger_total.

## Hazards

- **`connectorgen batch materialize` is BLOCKED** (finding F2): this artifact is OAS 3.1 with a non-empty top-level `webhooks` object, which `batchArtifactWebhooksUnknown` rejects fail-closed. Hand-author around the gate; do NOT relax it. Webhooks are deferred to `cli-webhook-surface-sweep-r1`.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 6** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 8** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm bamboo-hr <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#3215** (new generation); children are expected at **#3216–#3222** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/bamboo-hr_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^bamboo-hr_'`) — gong carried two, and a targeted
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
  non-bamboo-hr path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
