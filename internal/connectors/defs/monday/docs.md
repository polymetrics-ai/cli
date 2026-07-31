# Overview

The Monday connector reads monday.com GraphQL data and exposes connector-owned metadata for the official monday.com GraphQL reference. Existing ETL streams remain `boards`, `items`, `users`, `teams`, and `tags`. Additional documented query operations are represented as fixed, bounded direct-query commands when the current connector runtime can safely execute their scalar inputs; complex query shapes remain planned/blocked with source evidence.

This bundle now records 254 official docs operations from the developer.monday.com sitemap/reference pages: 66 queries and 188 mutations. The parent issue's r3 audit records 292 operations from `monday_graphql_schema_current`; this worker did not fetch `https://api.monday.com/v2/get_schema` because the task forbids live provider calls, so schema-only operations not present in public docs pages remain a recorded source dependency, not fabricated rows.

## Auth setup

Connection fields:

- `api_token` (optional, secret): monday.com personal API token sent verbatim as the `Authorization` header.
- `access_token` (optional, secret): monday.com OAuth access token sent verbatim as the `Authorization` header.
- `api_version` (optional): sent as the `API-Version` header when configured.
- `base_url` (optional): defaults to `https://api.monday.com/v2`; used by fixture replay and proxies.
- `page_size` (optional): defaults to `50` for hook-backed streams.
- `max_pages` (optional): positive integer cap for hook-backed stream pagination; empty, `all`, `unlimited`, or non-positive values mean unbounded up to the hook safety cap.

Never pass token values in chat, CLI arguments, docs, fixtures, or logs. Use environment variables or stdin-backed credential loading.

## Streams notes

The five legacy-parity ETL streams are still handled by the Monday StreamHook because monday.com carries pagination state inside GraphQL POST bodies:

- `boards`, `users`, `teams`, and `tags` use page-number pagination.
- `items` uses `boards { items_page }` followed by `next_items_page` cursor pagination.

Additional docs-sourced query operations are modeled in `operations.json` and `cli_surface.json`. Commands with scalar-only arguments can run as bounded fixed-document direct reads through the existing operation direct-read path. Operations with array/object GraphQL arguments stay planned/blocked until a connector-local or shared typed variable contract can pass those inputs without becoming a raw GraphQL escape hatch.

## Write actions & risks

Monday reverse ETL is enabled only through named GraphQL write actions in `writes.json`; no arbitrary GraphQL document, method, path, or body command is exposed. This bundle declares 102 executable scalar-input mutation actions and keeps 86 complex/binary-input mutations planned/blocked with source evidence.

Every write action uses a fixed GraphQL mutation document, a draft-07 record schema, and the existing reverse ETL safety path: plan -> preview -> explicit approval -> execute. Destructive/admin/delete-class actions (for example `delete_board`) set `confirm: "destructive"` so execution requires typed destructive confirmation in addition to approval.

Representative fixtures under `fixtures/writes/` prove the GraphQL request shape for `update_board` and destructive `delete_board` without contacting monday.com.

## Known limits

- No live monday.com provider calls, credentials, writes, or certification were performed for this parity slice.
- The live `/v2/get_schema` source named by #82 was not fetched under the no-live-provider-call gate. The public docs reference inventory currently yields 254 operations, while #82 preserves the previous r3 count of 292; this count divergence is recorded for firstmate/human reconciliation.
- Complex GraphQL input objects, arrays, and multipart/binary file uploads are not passed through scalar template variables by the existing write/direct-read foundations. Those operations are present as planned/blocked rows instead of an unsafe generic GraphQL or file-upload escape hatch.
- Direct query command output is capped and redacted with `json_redacted`; binary/file-like operations remain planned/blocked unless represented by a bounded fixed operation.
- Fixture evidence is not live certification. Certification remains a separate human-gated lane.
