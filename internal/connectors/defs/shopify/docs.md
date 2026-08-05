# Shopify connector

## Overview

This bundle records Shopify Admin API parity for parent issue #3013 from Shopify's current published reference, retrieved on 2026-08-06. The source inventory has 1098 `api_surface.json` rows: 287 GraphQL queries, 518 GraphQL mutations, and 293 REST rows (152 GET, 73 POST, 35 PUT, and 33 DELETE). Every ledger row has an official citation URL and retrieval date in `source_inventory.json`.

The connector implements a fixture-backed `shop` read stream. It also declares each current documented REST DELETE as one fixed, typed `rest_write` operation and static `direct_write` command (33 total), including the inventory-level identifier pair. These commands are intentionally marked `planned` until #3852's shared `cli_surface` schema accepts the captain-required non-redacting `json` output policy; no redacting fallback is declared. The 136 currently documented GraphQL mutations that the prior ledger classified as destructive remain explicit `destructive_action` ledger rows and planned static commands: a future implementation must add a fixed document, typed input, bounded non-redacting output, and typed destructive confirmation. All other official rows are present as blocked operation-ledger entries with source URLs and individually named static command declarations. Blocked rows are in scope for future fixed streams, fixed direct reads, fixed binary reads, CDC/changefeed surfaces, or typed reverse-ETL write actions; they are not blanket unsafe exclusions.

## Auth setup

Configure `shop_domain` with the canonical lowercase Shopify Admin API host only, for example `fixture-shop.myshopify.com`; a merchant's custom storefront domain is not an Admin API host. Schemes, paths, ports, trailing dots, and non-`myshopify.com` hosts are refused before credentials are persisted. Provide the Admin API token through the credential store or environment-backed secret input. Do not put token values in prompt text, docs, fixtures, or issue comments. Requests use the `X-Shopify-Access-Token` header.

## Streams notes

- `shop` calls the fixed REST path `/admin/api/latest/shop.json`, emits the `shop` object, and ships a sanitized replay fixture.
- GraphQL query rows remain blocked until each has a fixed reviewed document, typed variables, bounded pagination/output policy, and fixture evidence. No arbitrary GraphQL document or raw variables blob is exposed.
- REST read/direct/binary rows remain blocked unless represented by a fixed stream or future fixed command metadata. Every row nevertheless has a concrete `pm shopify <command>` declaration; there is no generic method/path/query/body command.

## Write actions & risks

- The 33 current documented REST DELETE rows use connector-owned `rest_write` operations with exact paths, typed identifier flags, `mutation_class: "destructive"`, `destructive: true`, `confirmation.kind: "destructive"`, `batchable: false`, and a bounded 1 MiB response. They are designed for the existing plan -> preview -> explicit approval -> execute path.
- The 136 current GraphQL mutations retained from the prior destructive-action ledger are in scope but remain blocked until each has a fixed reviewed document, typed input schema, bounded non-redacting JSON output, and plan -> preview -> explicit approval -> execute with typed destructive confirmation. They are not demoted to a generic or unsafe exclusion.
- The inventory-level DELETE uses its published `inventory_item_id` and `location_id` identifiers as required typed inputs. Retired owner/global Metafield DELETE routes and the old theme-asset query route are absent from Shopify's current published reference and are not retained.
- #3852 owns the shared command-surface output-policy schema that currently excludes the required non-redacting `json` direct-write value. Until that lane repairs the shared schema, these otherwise complete typed delete commands remain `planned`; this bundle does not substitute a redacting policy or a generic HTTP write command.
- The other GraphQL mutations and REST POST/PUT rows are in scope but blocked until operation-specific fixed documents or typed schemas, fixtures, and approval contracts are added.

## Known limits

- Fixture evidence is not certification; `certification.json` intentionally declares no live candidates or write pairings.
- The GitHub issue count tables were preserved by policy addendum and are not treated as implemented counts. This bundle records the current official-source inventory, including DELETE rows, without claiming that blocked rows are implemented.
- Shared foundations still gate the no-redaction command-surface output policy (#3852), generated icon registry coverage (#3809), provider search/query (#2985), and CDC/changefeed truthfulness/state (#2986/#2988). The Shopify bundle itself does not edit those shared components.
- No live Shopify credentials, live API calls, live writes, certification, or merges were performed for this bundle.
