# PLAN — issue-3013-shopify-parity

## GSD path and fallback

- Required adapter checks: `scripts/gsd doctor` passed on 2026-07-30.
- Required `scripts/gsd prompt programming-loop init --phase issue-3013-shopify-parity --dry-run` was attempted but the repo-local command registry returned `unknown GSD command: programming-loop`.
- Fallback GSD command used: `scripts/gsd prompt quick "Implement documented Shopify connector parity for issue 3013 within connector-local definition bundle, including destructive/delete operations with typed confirmation and plan-preview-approval-execute safety, no live credentials or provider calls"`.
- Manual universal-loop fallback: maintain this PLAN, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json`; record red/green evidence before production edits where possible.
- Resume (2026-08-06): `scripts/gsd prompt discuss-phase issue-3013-shopify-parity --auto` and `scripts/gsd prompt plan-phase issue-3013-shopify-parity --tdd --skip-research --auto` were generated and executed inline. The lifecycle is inline because all remaining work owns one Shopify bundle and the task's no-delegation rule prevents an isolated mutating worker fan-out.

## Required skills loaded

- `.agents/agentic-delivery/references/required-skills-routing.md`
- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-documentation`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-safety`
- `golang-graphql`
- `golang-lint`
- `no-mistakes`

## Scope

Allowed production scope is connector-local Shopify definition work:

- `internal/connectors/defs/shopify/**`
- connector-owned fixtures and generated/manual connector docs/metadata where required
- GitHub issue-body captain-policy addendum via `gh-axi` for #3013 and #3014-#3020

Do not edit shared runtime, engine, CLI dispatcher, hook/native registries, dependencies, or any other connector. Do not run live Shopify credentials/provider calls or execute writes.

## Slices

1. **Issue policy addendum**
   - Append idempotent captain-policy text to #3013 and #3014-#3020.
   - Preserve existing bodies and count tables; explicitly state destructive/delete operations are included when typed `destructive` confirmation and plan -> preview -> explicit approval -> execute are present.

2. **Official source inventory and ledger generation**
   - Inventory Shopify Admin GraphQL latest full index and Admin REST latest resource pages from official Shopify docs.
   - Generate a complete connector-local operation ledger under `internal/connectors/defs/shopify/api_surface.json` / `operations.json` without fabricating implemented or certified counts.
   - Mark rows that are not executable in this slice as blocked operation-ledger rows with source evidence rather than legacy blanket unsafe exclusions.

3. **Bundle scaffold and conformance shape**
   - Create `metadata.json`, `spec.json`, `streams.json`, `schemas/`, `fixtures/check.json`, `cli_surface.json`, `certification.json`, and `docs.md` sufficient for connector-local validation.
   - Keep runtime behavior safe and non-live: no generic GraphQL/raw HTTP/write passthrough; no credential fixture values; no certification claim.

4. **Verification and evidence**
   - Run focused validation: `go run ./cmd/connectorgen validate internal/connectors/defs/shopify`.
   - Run conformance if validation permits: `go test ./internal/connectors/conformance -run 'TestConformance/shopify' -count=1`.
   - Run `git diff --check` and path-scope guard.
   - Record any shared-foundation blocker exactly if connector-local work cannot claim full implementation.

5. **Parity checkpoint re-audit before review-ready**
   - Re-fetch official Shopify Admin GraphQL full index and Admin REST latest resource pages from the sitemap.
   - Compare fresh operation sets against `api_surface.json`; treat missing/stale rows and non-canonical DELETE path placeholders as findings, not waivers.
   - Update connector-local ledger, typed DELETE write schemas/fixtures, source inventory, docs, and verification evidence with truthful post-fix counts.
   - Keep DELETE/destructive operations in scope: executable only when represented by fixed write actions with `confirm: "destructive"`; query-identifier DELETE shapes remain blocked with the exact shared write-query dependency.

6. **Credential-boundary host restriction (captain decision `admin-api-host-allowlist`)**
   - Start red by adding the narrowly scoped Shopify row to the established credential-acceptance regression table in `internal/app/app_test.go`; it must reject a non-`myshopify.com` host before persistence, name `shop_domain` and `pattern`, and never echo the supplied value.
   - Change only the connector-owned `shop_domain` declaration to `^[a-z0-9][a-z0-9-]*\\.myshopify\\.com$`; do not add a custom-domain allow-list, a scheme/path/port exception, or shared-runtime code.
   - Add a green acceptance test for `fixture-shop.myshopify.com`, then regenerate Shopify manual/skill docs so the restriction is visible to operators.

7. **2026-08-06 published-reference rebuild and typed direct operations**
   - Rebuild the ledger from the current Shopify Admin GraphQL latest full-index Markdown and the 67 Admin REST latest resource-page Markdown artifacts discovered through Shopify's public sitemap; do not reuse the prior aggregate count as source evidence.
   - Record one canonical citation URL and the `2026-08-06` retrieval date for every ledger row in the connector-local source inventory. The rebuilt reference set is 287 GraphQL queries, 518 GraphQL mutations, and 293 REST operations (152 GET, 73 POST, 35 PUT, 33 DELETE): 1,098 rows total. The AccessScope resource page already contains the access-scope endpoint, so it must be represented once rather than double-counted from the usage guide.
   - Preserve the 136 still-current GraphQL mutation rows previously reviewed as destructive as explicit `destructive_action` ledger rows with a per-row `typed_destructive_confirmation` implementation requirement. They remain planned until their individual fixed document/schema contracts exist; they are not generic unsafe exclusions.
   - Replace the legacy reverse-ETL DELETE declarations that overlap the current reference with connector-owned fixed `rest_write` operations and individually named `direct_write` commands. Each must use `mutation_class: "destructive"`, typed `confirmation.kind: "destructive"`, bounded output, and the existing plan -> preview -> explicit approval -> execute path. Keep their commands planned until #3852 admits the required no-redaction `json` output policy in the shared schema.
   - Leave the remaining published operations individually represented by static command declarations with truthful `planned` availability. Do not expose generic GraphQL, HTTP, path, query, or body passthrough.
   - Do not edit or regenerate the shared icon registry. Generated catalog/manual/app validation remains deferred until #3809 repairs `icons-generate`.

8. **Shared-foundation handoff**
   - #3852 owns the shared `cli_surface` schema enum gap: the runtime supports non-redacting direct-write JSON output, but the schema currently rejects its declaration. Do not substitute a redacting policy or edit the shared schema here.
   - #3809 owns the icon-generator/registry repair. Do not regenerate or hand-edit registry output here.

## TDD/validation approach

- Red/preflight: prove Shopify bundle is absent or invalid before scaffold; capture `connectorgen validate internal/connectors/defs/shopify` failure.
- Green: generate connector-local bundle and pass focused validation/conformance as far as shared foundations allow.
- Refactor: format/generated JSON deterministic order; keep docs and counts synchronized.
- Resume host restriction: red credential-boundary test -> spec pattern/description -> green rejection and canonical-host acceptance tests -> generated docs -> focused app, conformance, connector validation, CLI/help, and boundary checks.
- Published-reference rebuild: preserve the red evidence that the former 1,166-row REST/GraphQL claim differs from Shopify's current public Markdown (1,098 rows), then validate the regenerated ledger, source provenance, typed direct-write preflight, and conformance without invoking the icon-dependent app/docs paths.

## Safety notes

- No Shopify credentials requested, printed, stored, or summarized.
- No provider API calls except public documentation fetches.
- Destructive/delete operations are not blanket-excluded; they are represented as in-scope but blocked/planned unless backed by typed `confirm: "destructive"` write actions and the existing plan -> preview -> explicit approval -> execute path.
- Fixture-only evidence is not certification.
