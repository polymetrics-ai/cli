# Jira documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`jira`, landing order 25). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json?_v=1.8516.72`
- **Kind**: openapi, OAS `3.0.1`
- **Retrieved**: 2026-08-07, 2445625 bytes
- **Documented operations: 616**
- **By method**: DELETE 89, GET 275, HEAD 0, OPTIONS 0, PATCH 0, POST 134, PUT 118, TRACE 0
- **Read / write split**: 299 read, 317 write
- **Deprecated (still counted)**: 29

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 616 |
| Re-derived | 616 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Fetched with curl. The version-pinned URL (?_v=1.8516.72) returned HTTP 404; re-fetched the identical path with the query string stripped and got HTTP 200 (2,445,625 bytes) -- reporting http_status/bytes for that successful fetch. Document declares `openapi: 3.0.1` (not a `swagger:` key), so artifact_kind is reported as 'openapi' even though Atlassian's own URL/branding calls it 'swagger-v3'. Counted one operation per (METHOD, path) pair under top-level `paths` for METHOD in {get,put,post,delete,patch,head,options,trace}; ignored parameters/summary/servers/$ref siblings. POST reads vs writes: an initial keyword pass over path+operationId+summary (search/query/list/graphql/lookup/autocomplete/suggest/count/preview/match/compare/validate/parse/bulkfetch, all matched with a leading word-boundary regex to avoid substrings like 'count' inside 'account') was then manually checked against every hit's `description` text. 'filter' was deliberately excluded from the keyword list up front because Jira's saved-search object is a stored resource literally named 'Filter' (POST /rest/api/3/filter creates one -- a write, not a read). Manual review caught and corrected 3 more cases the keyword pass alone would have gotten wrong: POST /rest/api/3/issue/properties [bulkSetIssuesPropertiesList] matched 'list' but its description reads 'Sets or updates... on issues' -> moved to write; POST /rest/api/3/expression/analyse [analyseExpression] and POST /rest/api/3/expression/eval [evaluateJiraExpression] are both non-mutating compute/validate endpoints that hadn't matched any read keyword at all -> moved to read. jql/pdcleaner (migrateQueries) and jql/sanitize (sanitiseJqlQueries) were kept as read after confirming via description that both are stateless JQL-text transforms with no persistence, despite mutation-sounding operationIds.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 5** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm jira <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#81** (old generation); children are expected at **#82–#87** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/jira_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^jira_'`) — gong carried two, and a targeted
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
  non-jira path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
