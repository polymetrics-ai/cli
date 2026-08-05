# Shopify connector

## Overview

This bundle records Shopify Admin API parity for parent issue #3013 using official Shopify documentation. The current source inventory reviewed on 2026-08-01 contains 1166 `api_surface.json` rows: 287 GraphQL queries, 518 GraphQL mutations, 317 REST non-delete rows, and 44 REST DELETE rows.

The connector implements a fixture-backed `shop` read stream and typed REST path-parameter delete actions that the current engine can express safely (42 executable DELETE actions). All other official rows are present as blocked operation-ledger entries with source URLs. Blocked rows are in scope for future fixed streams, fixed direct reads, fixed binary reads, CDC/changefeed surfaces, or typed reverse-ETL write actions; they are not blanket unsafe exclusions.

## Auth setup

Configure `shop_domain` with the canonical lowercase Shopify Admin API host only, for example `fixture-shop.myshopify.com`; a merchant's custom storefront domain is not an Admin API host. Schemes, paths, ports, trailing dots, and non-`myshopify.com` hosts are refused before credentials are persisted. Provide the Admin API token through the credential store or environment-backed secret input. Do not put token values in prompt text, docs, fixtures, or issue comments. Requests use the `X-Shopify-Access-Token` header.

## Streams notes

- `shop` calls the fixed REST path `/admin/api/latest/shop.json`, emits the `shop` object, and ships a sanitized replay fixture.
- GraphQL query rows remain blocked until each has a fixed reviewed document, typed variables, bounded pagination/output policy, and fixture evidence. No arbitrary GraphQL document or raw variables blob is exposed.
- REST read/direct/binary rows remain blocked unless represented by a fixed stream or future fixed command metadata.

## Write actions & risks

- Implemented REST DELETE actions are `kind: delete`, `body_type: none`, idempotent for `404`, use the official documented path variables (for example `{blog_id}` rather than generic `{id1}`), and declare `confirm: "destructive"`. They still execute only through the existing reverse ETL plan -> preview -> explicit approval -> execute path.
- REST DELETE rows that require query identifiers, such as inventory levels (`inventory_item_id` + `location_id`) and theme assets (`asset[key]`), are blocked with an exact shared write-query foundation dependency instead of being excluded as unsafe.
- GraphQL mutations and REST POST/PUT rows are in scope but blocked until operation-specific typed schemas, redaction, fixtures, and approval text are added.

## Known limits

- Fixture evidence is not certification; `certification.json` intentionally declares no live candidates or write pairings.
- The GitHub issue count tables were preserved by policy addendum and are not treated as implemented counts. This bundle records the current official-source inventory, including DELETE rows, without claiming that blocked rows are implemented.
- Shared foundations still gate provider search/query (#2985), CDC/changefeed truthfulness/state (#2986/#2988), and connector-local write query-parameter support for documented REST DELETE shapes that do not encode all identifiers in the path.
- No live Shopify credentials, provider calls, live writes, certification, or merges were performed for this bundle.
