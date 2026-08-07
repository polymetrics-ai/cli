# Mixpanel documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`mixpanel`, landing order 5). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://docs.mixpanel.com/openapi/{annotations,data-pipelines,experiments,export,feature-flags-management,feature-flags,gdpr,identity,ingestion,lexicon-schemas,query,service-accounts,warehouse-connectors}.openapi.yaml (13 real OpenAPI files, discovered via the Mintlify nav JSON embedded in the recorded overview page)`
- **Kind**: openapi, OAS `mixed: 3.0.2 x9 files, 3.0.3 x1 (query.yaml), 3.1.0 x3 (experiments.yaml, feature-flags-management.yaml, feature-flags.yaml)`
- **Retrieved**: 2026-08-07, 392835 bytes
- **Documented operations: 104**
- **By method**: DELETE 10, GET 41, PATCH 4, POST 44, PUT 5
- **Read / write split**: 44 read, 60 write
- **Deprecated (still counted)**: 0

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 105 |
| Re-derived | 104 |
| Delta | -1 |

**The ledger was correct when taken** and has since drifted; see the note below.

**How it was counted.** Ledger recorded https://docs.mixpanel.com/reference/overview as html_reference. That page is Mintlify-hosted (not ReadMe.com despite ReadMe-origin x-readme-deploy-id extension keys inside the specs, meaning the specs were originally authored on ReadMe and later migrated/mirrored to Mintlify). The page's embedded Next.js nav-tree JSON (docsConfig) references 13 distinct 'openapi/<name>.openapi.yaml' files. All 13 were fetched directly at https://docs.mixpanel.com/openapi/<name>.openapi.yaml and all returned HTTP 200 text/yaml (annotations, data-pipelines, experiments, export, feature-flags-management, feature-flags, gdpr, identity, ingestion, lexicon-schemas, query, service-accounts, warehouse-connectors; combined 392,835 bytes). Parsed all 13 with PyYAML and counted unique (METHOD, path) pairs across the combined set: 104. Raw operationId/page count summed across the 13 files (pre-dedup) = 105, which is an EXACT match to ledger_total. The -1 delta comes from one genuine cross-file path collision: POST /import is documented as two separate operationIds — 'identity-merge' in identity.yaml and 'import-events' in ingestion.yaml — which the counting policy's method+path dedup rule collapses into a single action. No OAS 3.1 top-level 'webhooks' object exists in any of the 13 files (checked explicitly; also zero occurrences of the string 'webhook' anywhere in any file), so blocks_connectorgen_batch is false despite 3 of the 13 files being OAS 3.1.0.

## Hazards

- **Path-collapse hazard**: operations share a path and are selected by header/body. Rows in `api_surface.json` must be keyed so each operation stays distinct, or the whole surface collapses into a handful of rows. This has already bitten this programme once (DynamoDB's `X-Amz-Target`).

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 0** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm mixpanel <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#3158** (new generation); children are expected at **#3159–#3165** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/mixpanel_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^mixpanel_'`) — gong carried two, and a targeted
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
  non-mixpanel path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
