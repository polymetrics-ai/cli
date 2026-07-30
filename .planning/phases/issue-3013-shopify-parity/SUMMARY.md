# SUMMARY — issue-3013-shopify-parity

Implemented a connector-local Shopify Admin API parity bundle for parent issue #3013 and subissues #3014-#3020.

## Delivered

- Added `internal/connectors/defs/shopify/**` with metadata, auth spec, fixture-backed `shop` stream, typed destructive REST DELETE write actions, operation ledger, CLI surface, docs, certification metadata, schemas, and sanitized fixtures.
- Represented official Shopify Admin API surface from public docs as an operation ledger: 287 GraphQL queries, 518 GraphQL mutations, 264 REST non-delete rows, and 34 REST DELETE rows. Non-executable rows are blocked/planned with source evidence rather than exposed as generic passthrough.
- Covered 33 executable connector rows: fixture-backed `shop` read plus 32 typed destructive delete actions, all with `confirm: "destructive"` and no request body. The two documented REST DELETE shapes that require query/body identifiers (`inventory_levels`, `themes/{id}/assets`) remain blocked/planned with source evidence rather than exposed unsafely.
- Updated connector docs/help/catalog golden artifacts for the catalog count change (`553` -> `554`, declarative bundles `549` -> `550`).
- Appended idempotent captain-policy addendum marker `captain-policy-shopify-destructive-in-scope-r1` to #3013 and #3014-#3020 via `gh-axi`.

## Safety

- No live credentials, provider calls, write execution, certification, VPS/Thaalam work, merges, or default-branch pushes.
- Reverse ETL safety remains the existing plan -> preview -> explicit approval -> execute path; destructive Shopify actions additionally require typed `destructive` confirmation.
- Fixture-only evidence is explicitly not certification.

## Verification

See `VERIFICATION.md`; focused validation, conformance, CLI/golden tests, vet, build, connector-boundary, and diff check passed.
