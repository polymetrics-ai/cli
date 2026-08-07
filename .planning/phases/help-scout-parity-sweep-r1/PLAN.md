# Help Scout documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`help-scout`, landing order 9). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://developer.helpscout.com/mailbox-api/endpoints/`
- **Kind**: html_reference, OAS `n/a (no machine-readable spec found)`
- **Retrieved**: 2026-08-07, 60307 bytes
- **Documented operations: 145**
- **By method**: DELETE 19, GET 79, PATCH 6, POST 21, PUT 20
- **Read / write split**: 79 read, 66 write
- **Deprecated (still counted)**: 0

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 146 |
| Re-derived | 145 |
| Delta | -1 |

**The ledger was correct when taken** and has since drifted; see the note below.

**How it was counted.** The literal given artifact_url (…/mailbox-api/endpoints/) returns HTTP 404 -- it is not a real page, just a section prefix. Searched hard for a machine-readable spec first: (1) checked developer.helpscout.com for any linked .json/.yaml/openapi/swagger file -- none found on the base Mailbox API page or the 404 page; (2) web-searched "Help Scout Mailbox API openapi swagger json" and site-scoped queries -- no results; (3) listed all ~110 repos in the helpscout GitHub org via the GitHub API -- API client libraries (PHP/Java) exist but no OpenAPI/Swagger spec repo. Concluded no public machine-readable spec exists; fell back to the HTML index per the task's instructions and marked artifact_kind html_reference. Fetched https://developer.helpscout.com/mailbox-api/ (HTTP 200, 60,307 bytes, --compressed needed -- the server gzips despite Accept-Encoding not being requested by plain curl) and extracted its left-nav sidebar, which is shared across every docs page and lists every '/mailbox-api/endpoints/**' doc page: 146 unique page URLs (174 raw hrefs before dedup -- the sidebar repeats the 'current page' entry an extra 1-2x for a handful of pages, a rendering artifact, not a count signal). Fetched all 146 endpoint pages individually (curl --compressed, one had a literal space in its URL segment requiring %20 encoding: 'users/conversation reassignment/{get,update}'). Parsed each page's rendered 'Request' code block (<span class="nf">METHOD</span> <span class="nn">path</span>) -- every one of the 146 pages yields exactly 1 method+path line (0 pages with 0 or >1 matches). That gives 146 raw doc-page-to-operation mappings, but per the stated counting policy (one operation = one unique METHOD+path pair), 2 of those 146 pages document the IDENTICAL 'GET /v2/conversations/{id}/threads/{id}/original-source' route -- one page shows the request with 'Accept: application/json', the other 'Accept: message/rfc822' (content negotiation on the same endpoint, documented as two separate walkthroughs). Deduping strictly by (METHOD,path) therefore yields 145 unique operations, not 146. Verified the '-v3' pages (conversations/get-v3, customers/list-v3, threads/list-v3) are genuinely distinct '/v3/...' paths, not further duplicates. Checked all 12 pages that mention the substring 'deprecat' individually: every single one refers to a deprecated FIELD/parameter inside a still-active operation (e.g. the legacy 'organization' string field superseded by 'organizationId', or a legacy tag 'color' field) -- zero operations/endpoints themselves are deprecated, so deprecated_count=0. Read vs write: all 79 GETs are reads; all 66 POST/PUT/PATCH/DELETE operations were checked by name and none are search/list-shaped POSTs (Help Scout puts all its filtering/search on GET query params, e.g. GET /v2/conversations), so post_classified_as_read is empty and all non-GET methods are writes.

## Hazards

- **Path-collapse hazard**: operations share a path and are selected by header/body. Rows in `api_surface.json` must be keyed so each operation stays distinct, or the whole surface collapses into a handful of rows. This has already bitten this programme once (DynamoDB's `X-Amz-Target`).

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 26** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 5** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm help-scout <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#212** (old generation); children are expected at **#213–#218** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/help-scout_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^help-scout_'`) — gong carried two, and a targeted
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
  non-help-scout path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
