# Quickbooks documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`quickbooks`, landing order 29). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `same base path, expanded to all 74 sibling pages under .../api/accounting/all-entities/<entity-slug> (the page's own left-nav enumerates them; see derivation_note)`
- **Kind**: html_reference
- **Retrieved**: 2026-08-07, 1806305 bytes
- **Documented operations: 198**
- **By method**: n/a
- **Read / write split**: 122 read, 76 write
- **Deprecated (still counted)**: unknown

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | unknown |
| Re-derived | 198 |
| Delta | unknown |

**Undetermined** — a defensible count could not be established from a current artifact.

**How it was counted.** Enumerated the full 'All entities' left-nav from the given page - 74 entities total (confirmed via the site's own client-side React state, not a static HTML list; the raw HTML is a ~1.2MB webpack/SPA shell with zero server-rendered entity content, so every page required JS execution via a headless browser). The 74 span both true CRUD-ish accounting-data entities (45 of 74, e.g. Invoice, Customer, Bill) and read-only Reports (29 of 74, e.g. BalanceSheet, ProfitAndLoss, AgedReceivables - each exposing exactly one GET /v3/company/{realmId}/reports/{ReportName} endpoint, except TrialBalance which has 2 locale-specific report names: TrialBalance and TrialBalanceFR). No official OpenAPI/Swagger spec exists for this API (checked: Intuit's own developer-community forum states this explicitly; only unofficial/community-maintained specs exist on GitHub Gists, which were deliberately NOT used as a source of truth). The surface is explicitly NOT a uniform per-entity operation matrix - per-entity unique-operation count ranges from 1 (reports, Batch, ChangeDataCapture, Entitlements, TaxService) to 8 (Invoice, SalesReceipt, which add Void + two Send variants on top of the base Create/Query/Read/Delete/PDF set) - so a single multiplication would misrepresent the surface; the total below is a literal per-entity sum, not an assumed constant. Method: for each of the 74 entities, rendered its own /all-entities/<slug> page, read its own left-nav sub-items (that entity's own documented-operation-section list, e.g. Account: Create/Query/Read/Full-update), extracted the literal 'Request URL: METHOD /path' code sample under each section, and reduced to unique (METHOD, path) pairs PER ENTITY before summing across all 74 (summing per-entity, without a further cross-entity dedup pass, is valid because the entity name is embedded in essentially every path, e.g. /v3/company/{realmId}/invoice vs .../customer - the one deliberate exception is the entity-agnostic Query endpoint GET /v3/company/{realmId}/query, which is genuinely identical text across all 38 entities that document it; I counted it once per entity anyway since Intuit documents 'Query an X' as that entity's own operation section with an entity-specific example, and collapsing it to a single API-wide row would understate what is actually and separately documented on each of those 38 pages - this is the inverse judgment call from the Create/Update collapse above, and is called out explicitly here rather than silently applied). A first automated pass (74 pages via a background sub-agent sweep, 0 page-load failures) found 185 unique pairs, but a manual audit caught systematic extraction-window bugs on 8 entities (Estimate, Payment, SalesReceipt, Invoice, CreditMemo, PurchaseOrder, RefundReceipt, TrialBalance) where a parenthetical explainer before a code sample (e.g. '(Specifying an explicit email address)') pushed the real Request URL past the capture window, either dropping it entirely or concatenating two adjacent examples into one unparseable string. I revisited all 8 directly with a wider, targeted capture and hand-verified the true structure by reading the raw rendered text myself (see surprises for exactly what each correction was). Final verified count: 74/74 entities processed with 0 errors, 0 unparsed entries, 0 zero-operation entities. operations_total = 198 = sum of each entity's own unique (method,path) pairs (29 report entities contribute 30 of these: TrialBalance=2, every other report=1; the remaining 45 entities contribute 168). Method split: GET=122 (operations_read), POST=76 (operations_write - QuickBooks' REST API never uses PUT/PATCH/DELETE HTTP verbs; every write is POST, with intent distinguished by body content or a '?operation=' query flag).

## Hazards

- **Path-collapse hazard**: Not GraphQL-style (each entity genuinely has multiple distinct paths), but a real, narrower collapse exists: for most entities, the 'Create X' / 'Full update X' / 'Sparse update X' documentation sections all resolve to the SAME 'POST /v3/company/{realmId}/{entity}' method+path pair, differentiated only by request-body content (Id+SyncToken present = update; sparse:true = partial update) - e.g. SalesReceipt documents 3 separate sections that all collapse to one pair. Delete/Void variants are differentiated only by a '?operation=' query-string flag on the same base path (e.g. POST .../invoice vs POST .../invoice?operation=delete vs POST .../invoice?operation=void); I treated the query string as part of the distinguishing path for counting (defensible since Intuit documents each as a separately-headed, separately-'Try it'-able operation), but a stricter path-only definition would collapse these further still. api_surface rows must key on the documented operation NAME/intent (Create/Read/Update/Delete/Void/Send/Query), not on path alone, or up to 3 distinct write operations per entity silently merge into one row. Rows in `api_surface.json` must be keyed so each operation stays distinct, or the whole surface collapses into a handful of rows. This has already bitten this programme once (DynamoDB's `X-Amz-Target`).

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: unknown** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 0** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm quickbooks <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#3094** (new generation); children are expected at **#3095–#3101** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/quickbooks_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^quickbooks_'`) — gong carried two, and a targeted
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
  non-quickbooks path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
