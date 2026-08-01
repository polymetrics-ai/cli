# PLAN — issue-3013-shopify-parity

## GSD path and fallback

- Required adapter checks: `scripts/gsd doctor` passed on 2026-07-30.
- Required `scripts/gsd prompt programming-loop init --phase issue-3013-shopify-parity --dry-run` was attempted but the repo-local command registry returned `unknown GSD command: programming-loop`.
- Fallback GSD command used: `scripts/gsd prompt quick "Implement documented Shopify connector parity for issue 3013 within connector-local definition bundle, including destructive/delete operations with typed confirmation and plan-preview-approval-execute safety, no live credentials or provider calls"`.
- Manual universal-loop fallback: maintain this PLAN, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json`; record red/green evidence before production edits where possible.

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

## TDD/validation approach

- Red/preflight: prove Shopify bundle is absent or invalid before scaffold; capture `connectorgen validate internal/connectors/defs/shopify` failure.
- Green: generate connector-local bundle and pass focused validation/conformance as far as shared foundations allow.
- Refactor: format/generated JSON deterministic order; keep docs and counts synchronized.

## Safety notes

- No Shopify credentials requested, printed, stored, or summarized.
- No provider API calls except public documentation fetches.
- Destructive/delete operations are not blanket-excluded; they are represented as in-scope but blocked/planned unless backed by typed `confirm: "destructive"` write actions and the existing plan -> preview -> explicit approval -> execute path.
- Fixture-only evidence is not certification.
