# SUMMARY — issue-3013-shopify-parity

Rebuilt the connector-local Shopify Admin API parity bundle for parent issue #3013 and subissues #3014-#3020 from Shopify's current published reference. The local definition slice is committed-ready but cannot be promoted to full runtime parity until two shared dependencies are resolved.

## Delivered

- Rebuilt `internal/connectors/defs/shopify/**` with current source inventory, auth spec, fixture-backed `shop` stream, operation ledger, static CLI surface, docs, certification metadata, schemas, and connector-local regression tests.
- Represented 1,098 current official rows exactly once: 287 GraphQL queries, 518 GraphQL mutations, and 293 REST operations (152 GET, 73 POST, 35 PUT, 33 DELETE). Every row carries an official citation URL and `2026-08-06` retrieval date in `source_inventory.json`.
- Replaced 42 legacy `writes.json` DELETE actions with 33 current fixed typed `rest_write` declarations and fixtures. The 10 retired owner/global Metafield DELETE routes are removed; the current inventory-level identifier-pair DELETE is now typed. Every typed DELETE has `confirmation.kind: "destructive"`, bounded output, and plan -> preview -> explicit approval -> execute metadata.
- Declared 1,098 individual static `pm shopify <command>` paths with no generic GraphQL, HTTP path, query, or body passthrough. The fixture-backed `shop read` stream remains implemented; all other commands retain truthful planned availability.
- Appended idempotent captain-policy addendum marker `captain-policy-shopify-destructive-in-scope-r1` to #3013 and #3014-#3020 via `gh-axi`.

## Safety

- No live credentials, live provider calls, write execution, certification, VPS/Thaalam work, merges, or default-branch pushes. Public Shopify documentation was fetched only to rebuild the cited inventory.
- Reverse ETL safety remains the existing plan -> preview -> explicit approval -> execute path; destructive Shopify actions additionally require typed `destructive` confirmation.
- Fixture-only evidence is explicitly not certification.

## Shared dependencies

- The no-redaction captain policy requires direct-write `output_policy: "json"`. The shared `internal/connectors/engine/schema/cli_surface.schema.json` does not admit that value, even though the declared REST write executor supports it. The 33 typed deletes therefore remain planned until the schema owner lands that change and surface-sync can derive their executable bindings.
- #3809 owns the unrelated shared icon generator defect. The connector-local work reached the genuine registry boundary when global command-runner preflight panicked for the missing Shopify icon row. No generator retry, registry workaround, or hand edit was made.

## Resume host policy

- Tightened `shop_domain` to canonical lowercase `*.myshopify.com` only, rejecting custom storefront domains, schemes, paths, ports, and trailing dots before credential persistence.
- Added connector-local rejection/acceptance evidence and app-level credential-boundary cases for the non-`myshopify.com` and canonical host paths.
- App-level host-boundary and generated-document verification remain blocked by #3809's missing generated Shopify icon coverage. No generated registry file was hand-edited.

## Verification

See `VERIFICATION.md`; current connector-local tests, full definition validation, surface-sync check, conformance, vet, and diff check passed. Global command-runner preflight is blocked at icon registration, and no-redaction direct-write promotion is blocked by the shared command schema.
